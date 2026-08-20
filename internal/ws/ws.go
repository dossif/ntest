// Package ws implements the WebSocket test: on every tick of an interval
// timer it opens a WebSocket connection (full HTTP Upgrade handshake), sends
// a ping frame and waits for the pong, then closes — logging success or
// failure until the context is cancelled. A successful run confirms the
// WebSocket protocol actually works end to end, not just that the TCP port
// is open — reverse proxies, load balancers and WAFs commonly pass plain
// HTTP/TCP through while breaking the `Upgrade: websocket` handshake.
package ws

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/coder/websocket"
	"github.com/dossif/ntest/internal/dns"
	log "github.com/sirupsen/logrus"
)

// Test holds everything needed to run a repeating WebSocket handshake+ping
// against Url.
type Test struct {
	Url      *url.URL      // parsed --host, scheme defaulted to ws if absent
	Host     string        // Url's bare hostname, resolved fresh before every attempt
	Port     string        // target port, fixed at construction (from Url, not DNS)
	Ns       string        // DNS server used to resolve Host, "" for the OS resolver
	Timeout  time.Duration // deadline covering the handshake and the ping/pong
	Interval time.Duration // delay between attempts
	Domain   string        // if set, overrides the Host header sent in the handshake

	// httpClient performs the handshake; its Transport dials resolvedAddr
	// (see NewTest) rather than letting net/http re-resolve the hostname.
	httpClient *http.Client
	// resolvedAddr is the "ip:port" the transport's DialContext connects
	// to. Execute sets it right before every Dial call, from that tick's
	// fresh DNS resolution; DialContext (invoked synchronously inside
	// Dial) just reads it back. Single-goroutine, sequential ticks — no
	// synchronization needed (same pattern as internal/http).
	resolvedAddr string
}

// NewTest resolves bind and builds the *http.Client used for the WebSocket
// handshake — see the DialContext comment below for why. Host is not
// resolved here: resolution happens on every tick in Execute, so an
// initially unresolvable host (e.g. DNS not yet propagated) doesn't stop
// the test from starting — it just logs a resolution error each tick until
// the host resolves.
func NewTest(bind string, host string, timeout time.Duration, interval time.Duration, domain string, ns string) (*Test, error) {
	bindIp, err := dns.ResolveAddr(bind, "")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve bind: %w", err)
	}
	wsUrl, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host: %w", err)
	}
	if wsUrl.Scheme == "" {
		wsUrl.Scheme = "ws"
	}
	// A bare host with no scheme, e.g. "example.com" (the whole string lands
	// in Path, Host stays empty) or "example.com:8080" (url.Parse reads
	// "example.com" as the scheme and "8080" as opaque data), both leave
	// Hostname() empty — catch that here with an actionable error instead of
	// a confusing DNS resolution failure for an empty hostname.
	if wsUrl.Hostname() == "" {
		return nil, fmt.Errorf("invalid host %q: missing hostname (did you forget the scheme? e.g. ws://%s)", host, host)
	}
	port := wsUrl.Port()
	if port == "" {
		if wsUrl.Scheme == "wss" {
			port = "443"
		} else {
			port = "80"
		}
	}
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: bindIp.IP, Port: 0, Zone: ""},
	}
	test := &Test{
		Url:      wsUrl,
		Host:     wsUrl.Hostname(),
		Port:     port,
		Ns:       ns,
		Timeout:  timeout,
		Interval: interval,
		Domain:   domain,
	}
	transport := &http.Transport{
		// Ignore the "addr" net/http would normally pass here (the URL's
		// host:port, which it resolves itself via the OS resolver) and
		// always dial test.resolvedAddr instead, set by Execute from that
		// tick's fresh resolution — same trick as internal/http, and for
		// the same reason: without it, --dns would only affect the "dest"
		// log field, not the actual connection.
		// TLS (SNI, certificate hostname) is unaffected — net/http derives
		// that from the request URL/Host, not from the dialed address.
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, test.resolvedAddr)
		},
		// A new connection (and therefore a fresh DialContext call reading
		// the just-resolved resolvedAddr) is wanted on every tick anyway —
		// each tick opens and closes its own WebSocket connection, there is
		// nothing to keep alive between ticks.
		DisableKeepAlives: true,
	}
	test.httpClient = &http.Client{Transport: transport}
	return test, nil
}

// Execute runs the handshake+ping loop until ctx is cancelled, logging one
// line per tick. It always returns nil on cancellation; resolution,
// handshake and ping errors are logged, not returned, so none of them stop
// the monitoring loop.
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
				"seq":  seq,
				"dest": t.Url.String(),
			}
			// Resolved before every attempt (not once at startup): the
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

			// One deadline covers both the handshake and the ping/pong
			// round trip — same per-tick pattern as internal/tcp and
			// internal/http.
			rCtx, cancel := context.WithTimeout(ctx, t.Timeout)
			opts := &websocket.DialOptions{HTTPClient: t.httpClient}
			if t.Domain != "" {
				opts.Host = t.Domain
			}
			conn, resp, err := websocket.Dial(rCtx, t.Url.String(), opts)
			if err != nil {
				cancel()
				// Suppress the error log when the dial failed because the
				// outer context was cancelled (e.g. SIGINT arrived mid-
				// handshake) — that's a normal shutdown, not a test
				// failure.
				if ctx.Err() == nil {
					log.WithFields(lg).Errorf("websocket handshake error: %v", err)
				}
				continue
			}
			// Ping requires a concurrent reader to observe the pong;
			// CloseRead starts one that discards everything until the
			// connection closes, which is exactly what a bare
			// handshake+ping probe wants (see package coder/websocket's
			// docs for this exact pattern).
			pingCtx := conn.CloseRead(rCtx)
			pingErr := conn.Ping(pingCtx)
			// Close immediately rather than deferring: a deferred close
			// inside this loop would only run when Execute itself
			// returns, leaking one open connection (and its CloseRead
			// goroutine) per tick until then — the same mistake already
			// fixed once in internal/tcp.
			_ = conn.CloseNow()
			cancel()
			if pingErr != nil {
				if ctx.Err() == nil {
					log.WithFields(lg).Errorf("websocket ping error: %v", pingErr)
				}
				continue
			}
			log.WithFields(lg).Infof("websocket ok: %v", resp.Status)
		}
	}
}
