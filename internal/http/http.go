// Package http implements the HTTP request test: it sends a request to the
// target URL on every tick of an interval timer and logs the status (or the
// error) until the context is cancelled.
package http

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/dossif/ntest/internal/dns"
	log "github.com/sirupsen/logrus"
)

// Test holds everything needed to run a repeating HTTP request against Url.
type Test struct {
	Api      *http.Client
	Url      *url.URL      // parsed --host, scheme defaulted to http if absent
	Host     string        // Url's bare hostname, resolved fresh before every request
	Port     string        // target port, fixed at construction (from Url, not DNS)
	Ns       string        // DNS server used to resolve Host, "" for the OS resolver
	Timeout  time.Duration // per-request deadline
	Interval time.Duration // delay between requests
	Domain   string        // if set, overrides the Host header
	Method   string        // HTTP method
	Body     string        // request body

	// resolvedAddr is the "ip:port" the transport's DialContext connects
	// to. Execute sets it right before every t.Api.Do call, from that
	// tick's fresh DNS resolution; DialContext (invoked synchronously
	// inside Do) just reads it back. Single-goroutine, sequential ticks —
	// no synchronization needed.
	resolvedAddr string
}

// NewTest resolves bind and builds an *http.Client whose transport always
// dials the address Execute last resolved — see the DialContext comment
// below for why. Host is not resolved here: resolution happens on every
// tick in Execute, so an initially unresolvable host (e.g. DNS not yet
// propagated) doesn't stop the test from starting — it just logs a
// resolution error each tick until the host resolves.
func NewTest(bind string, host string, timeout time.Duration, interval time.Duration, method string, domain string, body string, ns string) (*Test, error) {
	bindIp, err := dns.ResolveAddr(bind, "")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve bind: %w", err)
	}
	hostUrl, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host: %w", err)
	}
	if hostUrl.Scheme == "" {
		hostUrl.Scheme = "http"
	}
	// A bare host with no scheme, e.g. "example.com" (the whole string lands
	// in Path, Host stays empty) or "example.com:8080" (url.Parse reads
	// "example.com" as the scheme and "8080" as opaque data), both leave
	// Hostname() empty — catch that here with an actionable error instead of
	// a confusing DNS resolution failure for an empty hostname.
	if hostUrl.Hostname() == "" {
		return nil, fmt.Errorf("invalid host %q: missing hostname (did you forget the scheme? e.g. http://%s)", host, host)
	}
	port := hostUrl.Port()
	if port == "" {
		if hostUrl.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: bindIp.IP, Port: 0, Zone: ""},
	}
	test := &Test{
		Url:      hostUrl,
		Host:     hostUrl.Hostname(),
		Port:     port,
		Ns:       ns,
		Timeout:  timeout,
		Interval: interval,
		Domain:   domain,
		Method:   method,
		Body:     body,
	}
	transport := &http.Transport{
		// Ignore the "addr" net/http would normally pass here (the URL's
		// host:port, which it resolves itself via the OS resolver) and
		// always dial test.resolvedAddr instead, set by Execute from that
		// tick's fresh resolution. Without this override, --dns would only
		// affect the "dest" log field: the actual TCP connection would
		// still go through Go's default resolver and completely ignore the
		// configured DNS server.
		// TLS (SNI, certificate hostname) is unaffected — net/http derives
		// that from the request URL/Host, not from the dialed address.
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, test.resolvedAddr)
		},
		// DialContext only runs on a connection-pool miss: a kept-alive
		// connection from an earlier tick is otherwise reused as-is, which
		// would silently keep talking to whatever IP that connection was
		// opened to — defeating the fresh resolution done every tick above
		// the moment DNS returns a different address (failover, DNS round-
		// robin, a record added after the process started). Disabling
		// keep-alives forces a new dial, and therefore a new DialContext
		// call reading the just-resolved resolvedAddr, on every request.
		DisableKeepAlives: true,
	}
	// No client-level Timeout here: the per-request deadline is enforced by
	// the context.WithTimeout(ctx, t.Timeout) applied to each request in
	// Execute, driven entirely by --timeout. A second, hardcoded client
	// timeout would silently cap --timeout below whatever value it was set
	// to, with nothing in the logs to explain the early failures.
	test.Api = &http.Client{Transport: transport}
	return test, nil
}

// Execute runs the request loop until ctx is cancelled, logging one line
// per tick. It always returns nil on cancellation; resolution errors,
// request errors and 4xx/5xx responses are logged, not returned, so none
// of them stop the monitoring loop.
func (t *Test) Execute(ctx context.Context) error {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})

	ticker := time.NewTicker(t.Interval)
	defer ticker.Stop()
	var seq int
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			seq++
			lg := log.Fields{
				"seq":    seq,
				"dest":   t.Url.String(),
				"method": t.Method,
			}
			// Resolved before every request (not once at startup): the
			// target may still be propagating in DNS, or its address may
			// change between ticks, so a stale or failed resolution
			// shouldn't be cached across attempts.
			ip, err := dns.ResolveAddr(t.Host, t.Ns)
			if err != nil {
				log.WithFields(lg).Errorf("failed to resolve host: %v", err)
				continue
			}
			t.resolvedAddr = net.JoinHostPort(ip.IP.String(), t.Port)
			lg["dest"] = fmt.Sprintf("%v (%v)", t.Url, ip.IP)

			req, err := http.NewRequest(t.Method, t.Url.String(), bytes.NewReader([]byte(t.Body)))
			if err != nil {
				log.WithFields(lg).Errorf("failed to create http request: %v", err)
				continue
			}
			// Override the Host header (SNI/cert hostname are unaffected,
			// they come from the request URL, not from this header).
			if t.Domain != "" {
				req.Host = t.Domain
			}
			// A fresh timeout context per tick, cancelled right after use —
			// see the equivalent comment in internal/tcp/tcp.go.
			rCtx, cancel := context.WithTimeout(ctx, t.Timeout)
			req = req.WithContext(rCtx)
			resp, err := t.Api.Do(req)
			cancel()
			if err != nil {
				// Suppress the error log on shutdown (ctx cancelled mid-
				// request) — that's expected, not a test failure.
				if ctx.Err() == nil {
					log.WithFields(lg).Errorf("http error: %v", err)
				}
				continue
			}
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				log.WithFields(lg).Warnf("http %v error", resp.Status)
			} else if resp.StatusCode >= 500 && resp.StatusCode < 600 {
				log.WithFields(lg).Errorf("http %v error", resp.Status)
			} else {
				log.WithFields(lg).Infof("http ok: %v", resp.Status)
			}
			// Keep-alives are disabled (see DisableKeepAlives above), so
			// there's no pooled connection to return by draining first — a
			// plain Close just releases this request's socket.
			_ = resp.Body.Close()
		}
	}
}
