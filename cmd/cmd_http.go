package cmd

import (
	"context"
	"time"

	"github.com/dossif/ntest/internal/http"
	"github.com/dossif/ntest/internal/signal"
	"github.com/spf13/cobra"
)

// httpCmd implements `ntest http <host>`: repeated HTTP requests to host (a
// full URI), logged on every tick until the process is interrupted.
var httpCmd = &cobra.Command{
	Use:   "http <host>",
	Short: "HTTP request test",
	Args:  cobra.ExactArgs(1),
	RunE:  runHttp,
}

// init registers the http subcommand and its flags on the root command.
func init() {
	f := httpCmd.Flags()
	f.String("bind", "0.0.0.0", "bind address")
	f.Duration("timeout", 3*time.Second, "request timeout (e.g. 500ms, 3s, 1m)")
	f.Duration("interval", time.Second, "interval between requests (e.g. 500ms, 3s, 1m)")
	f.String("method", "GET", "HTTP method (uppercase)")
	f.String("domain", "", "Host header override")
	f.String("body", "", "request body")
	f.String("dns", "", "DNS server for host resolution (e.g. 8.8.8.8)")
	rootCmd.AddCommand(httpCmd)
}

// runHttp reads the target host (a full URI) from args and the http flags,
// wires up signal-based cancellation and runs the request loop until ctx is
// cancelled.
func runHttp(cmd *cobra.Command, args []string) error {
	host := args[0]
	f := cmd.Flags()
	bind, _ := f.GetString("bind")
	timeout, _ := f.GetDuration("timeout")
	interval, _ := f.GetDuration("interval")
	method, _ := f.GetString("method")
	domain, _ := f.GetString("domain")
	body, _ := f.GetString("body")
	ns, _ := f.GetString("dns")

	if err := validateTimeout(timeout); err != nil {
		return err
	}
	if err := validateInterval(interval); err != nil {
		return err
	}
	if err := validateMethod(method); err != nil {
		return err
	}
	if err := validateDomain(domain); err != nil {
		return err
	}
	if err := validateDNS(ns); err != nil {
		return err
	}

	t, err := http.NewTest(bind, host, timeout, interval, method, domain, body, ns)
	if err != nil {
		return err
	}

	ctx := signal.ContextWithSignal(context.Background())
	return t.Execute(ctx)
}
