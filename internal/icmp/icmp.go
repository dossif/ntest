// Package icmp implements the ICMP ping test: it sends an echo request to
// the target host on every tick of an interval timer and logs the RTT (or
// the error) until the context is cancelled. In graph mode it redraws a
// live ASCII RTT chart instead of logging a line per tick.
package icmp

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/digineo/go-ping"
	"github.com/dossif/ntest/internal/dns"
	"github.com/guptarohit/asciigraph"
	log "github.com/sirupsen/logrus"
)

// graphWindow is how many of the most recent pings are kept on screen in
// graph mode — a fixed size keeps the chart a reasonable, fairly constant
// width instead of growing forever across a long-running test.
const graphWindow = 60

// Test holds everything needed to run a repeating ICMP ping against Host.
type Test struct {
	Api      *ping.Pinger  // raw-socket pinger bound to the local bind address
	Host     string        // target host, resolved fresh before every ping
	Ns       string        // DNS server used to resolve Host, "" for the OS resolver
	Timeout  time.Duration // per-ping deadline
	Interval time.Duration // delay between pings
	Warn     time.Duration // RTT above this logs at Warn instead of Info (or colors the graph point)
	Graph    bool          // redraw a live ASCII RTT chart instead of logging a line per tick
}

// NewTest resolves bind and builds a raw-socket pinger. Host is not
// resolved here: resolution happens on every tick in Execute, so an
// initially unresolvable host (e.g. DNS not yet propagated) doesn't stop
// the test from starting — it just logs a resolution error each tick until
// the host resolves.
func NewTest(bind string, host string, timeout time.Duration, interval time.Duration, warn time.Duration, ns string, graph bool) (*Test, error) {
	bindIp, err := dns.ResolveAddr(bind, "")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve bind: %w", err)
	}
	api, err := ping.New(bindIp.String(), "")
	if err != nil {
		return nil, fmt.Errorf("failed to create new icmp pinger: %w", err)
	}
	return &Test{
		Api:      api,
		Host:     host,
		Ns:       ns,
		Timeout:  timeout,
		Interval: interval,
		Warn:     warn,
		Graph:    graph,
	}, nil
}

// Execute runs the ping loop until ctx is cancelled. It always returns nil
// on cancellation; resolution and ping errors are logged (or shown on the
// graph), not returned, so neither stops the monitoring loop.
func (t *Test) Execute(ctx context.Context) error {
	if t.Graph {
		return t.executeGraph(ctx)
	}

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
			}
			// Resolved before every ping (not once at startup): the target
			// may still be propagating in DNS, or its address may change
			// between ticks, so a stale or failed resolution shouldn't be
			// cached across attempts.
			ip, err := dns.ResolveAddr(t.Host, t.Ns)
			if err != nil {
				log.WithFields(lg).Errorf("failed to resolve host: %v", err)
				continue
			}
			// Show "host[ip]" only when the host was a name that got
			// resolved, otherwise just the bare IP.
			if t.Host != ip.IP.String() {
				lg["dest"] = fmt.Sprintf("%v[%v]", t.Host, ip.IP.String())
			}
			rtt, err := t.Api.Ping(ip, t.Timeout)
			if err != nil {
				log.WithFields(lg).Errorf("icmp error: %v", err)
			} else if rtt > t.Warn {
				log.WithFields(lg).Warnf("icmp rtt %v: warn threshold %v exceed", rtt.Round(time.Millisecond), t.Warn)
			} else {
				log.WithFields(lg).Infof("icmp rtt %v", rtt.Round(time.Millisecond))
			}
		}
	}
}

// executeGraph is the --graph variant of Execute: instead of one log line
// per tick, it keeps a sliding window of the last graphWindow RTTs (in ms,
// NaN for a failed resolution or ping — asciigraph renders a NaN as a gap
// rather than a zero, so a lost ping doesn't read as a 0ms reply) and
// redraws the whole chart in place on every tick.
func (t *Test) executeGraph(ctx context.Context) error {
	ticker := time.NewTicker(t.Interval)
	defer ticker.Stop()

	rtts := make([]float64, 0, graphWindow)
	var attempts, fails int
	var lastErr string

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			attempts++
			ip, err := dns.ResolveAddr(t.Host, t.Ns)
			if err != nil {
				fails++
				lastErr = fmt.Sprintf("failed to resolve host: %v", err)
				rtts = pushWindow(rtts, math.NaN(), graphWindow)
				t.redraw(rtts, attempts, fails, lastErr)
				continue
			}
			rtt, err := t.Api.Ping(ip, t.Timeout)
			if err != nil {
				fails++
				lastErr = fmt.Sprintf("icmp error: %v", err)
				rtts = pushWindow(rtts, math.NaN(), graphWindow)
			} else {
				lastErr = ""
				rtts = pushWindow(rtts, float64(rtt)/float64(time.Millisecond), graphWindow)
			}
			t.redraw(rtts, attempts, fails, lastErr)
		}
	}
}

// ANSI colors for the pass/fail status banner — kept local rather than
// pulled from asciigraph's AnsiColor, since that type is only accepted by
// asciigraph's own Option functions, not usable for plain fmt output.
const (
	ansiRed   = "\033[1;31m"
	ansiGreen = "\033[1;32m"
	ansiReset = "\033[0m"
)

// redraw clears the screen and prints the current chart plus a pass/fail
// status banner and a stats footer. Bypasses logrus entirely — this is a
// full-screen redraw, not a log line, so it's written straight to stdout.
//
// A gap in the chart line (a NaN point, see executeGraph) already marks a
// failed attempt, but it's easy to miss at a glance — especially at the
// right edge, where a gap can look identical to "no data yet". The banner
// below makes the current status ("✓ ok" / "✗ ping failed: <reason>")
// impossible to miss, and the "fails" counter in the stats line gives the
// running total across the whole test, not just what fits in the window.
func (t *Test) redraw(rtts []float64, attempts int, fails int, lastErr string) {
	warnMs := float64(t.Warn) / float64(time.Millisecond)
	graph := asciigraph.Plot(rtts,
		asciigraph.Height(10),
		asciigraph.Caption(fmt.Sprintf("RTT (ms) — %s", t.Host)),
		asciigraph.ColorAbove(asciigraph.Yellow, warnMs),
	)

	var status string
	if lastErr != "" {
		status = fmt.Sprintf("%s✗ ping failed: %s%s", ansiRed, lastErr, ansiReset)
	} else {
		status = fmt.Sprintf("%s✓ ok%s", ansiGreen, ansiReset)
	}

	minRTT, maxRTT, avg, last, ok := rttStats(rtts)
	var stats string
	if ok {
		stats = fmt.Sprintf("last: %.0fms  min: %.0fms  max: %.0fms  avg: %.0fms  attempts: %d  fails: %d", last, minRTT, maxRTT, avg, attempts, fails)
	} else {
		stats = fmt.Sprintf("waiting for first reply...  attempts: %d  fails: %d", attempts, fails)
	}

	// "\033[H\033[2J" moves the cursor home and clears the screen — no
	// terminal raw mode, no resize handling, just clear-and-reprint on
	// every tick.
	fmt.Print("\033[H\033[2J")
	fmt.Println(graph)
	fmt.Println()
	fmt.Println(status)
	fmt.Println(stats)
}

// pushWindow appends v to s, dropping from the front once s exceeds max
// entries, so the chart shows a fixed-size sliding window of the most
// recent attempts instead of growing forever.
func pushWindow(s []float64, v float64, max int) []float64 {
	s = append(s, v)
	if len(s) > max {
		s = s[len(s)-max:]
	}
	return s
}

// rttStats returns the last, min, max and average of the non-NaN values in
// rtts. ok is false if rtts contains no successful reply yet.
func rttStats(rtts []float64) (min, max, avg, last float64, ok bool) {
	min = math.Inf(1)
	max = math.Inf(-1)
	var sum float64
	var n int
	for _, v := range rtts {
		if math.IsNaN(v) {
			continue
		}
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
		n++
		last = v
		ok = true
	}
	if n > 0 {
		avg = sum / float64(n)
	}
	return min, max, avg, last, ok
}
