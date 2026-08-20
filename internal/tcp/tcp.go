// Package tcp implements the TCP connect test: it dials Host:Port on every
// tick of an interval timer and logs success or failure until the context
// is cancelled.
package tcp

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/dossif/ntest/internal/dns"
	log "github.com/sirupsen/logrus"
)

// Test holds everything needed to run a repeating TCP dial against
// Host:Port.
type Test struct {
	Api      *net.Dialer   // dialer bound to the local bind address
	Host     string        // target host, resolved fresh before every dial
	Port     int           // target TCP port
	Timeout  time.Duration // per-dial deadline
	Interval time.Duration // delay between dial attempts
	Ns       string        // DNS server used to resolve Host, "" for the OS resolver
}

// NewTest resolves bind and builds a dialer whose local address is pinned
// to bind. Host is not resolved here: resolution happens on every tick in
// Execute, so an initially unresolvable host (e.g. DNS not yet propagated)
// doesn't stop the test from starting — it just logs a resolution error
// each tick until the host resolves.
func NewTest(bind string, host string, port int, timeout time.Duration, interval time.Duration, ns string) (*Test, error) {
	bindIp, err := dns.ResolveAddr(bind, "")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve bind: %w", err)
	}
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: bindIp.IP, Port: 0, Zone: ""},
	}
	return &Test{
		Api:      dialer,
		Host:     host,
		Port:     port,
		Timeout:  timeout,
		Interval: interval,
		Ns:       ns,
	}, nil
}

// Execute runs the dial loop until ctx is cancelled, logging one line per
// tick. It always returns nil on cancellation; resolution and dial errors
// are logged, not returned, so neither stops the monitoring loop.
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
				"dest": t.Host,
				"port": t.Port,
			}
			// Resolved before every dial (not once at startup): the target
			// may still be propagating in DNS, or its address may change
			// between ticks, so a stale or failed resolution shouldn't be
			// cached across attempts.
			ip, err := dns.ResolveAddr(t.Host, t.Ns)
			if err != nil {
				log.WithFields(lg).Errorf("failed to resolve host: %v", err)
				continue
			}
			lg["dest"] = fmt.Sprintf("%v (%v)", t.Host, ip.IP)
			addr := net.JoinHostPort(ip.String(), strconv.Itoa(t.Port))
			// A fresh timeout context is created and cancelled on every
			// tick (never reused across iterations, never deferred to the
			// end of Execute), otherwise the first timeout would expire
			// this context for good and every later dial would fail
			// immediately.
			rCtx, cancel := context.WithTimeout(ctx, t.Timeout)
			conn, err := t.Api.DialContext(rCtx, "tcp", addr)
			cancel()
			if err != nil {
				// Suppress the error log when the dial failed because the
				// outer context was cancelled (e.g. SIGINT arrived mid-dial)
				// — that's a normal shutdown, not a test failure.
				if ctx.Err() == nil {
					log.WithFields(lg).Errorf("tcp error: %v", err)
				}
				continue
			}
			log.WithFields(lg).Infof("tcp ok")
			// Close immediately rather than deferring: a deferred close
			// inside this loop would only run when Execute itself returns,
			// leaking one open socket per successful tick until then.
			_ = conn.Close()
		}
	}
}
