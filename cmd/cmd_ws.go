package cmd

import (
	"context"
	"time"

	"github.com/dossif/ntest/internal/signal"
	"github.com/dossif/ntest/internal/ws"
	"github.com/spf13/cobra"
)

// wsCmd implements `ntest ws <host>`: a repeated WebSocket handshake+ping
// against host (a full URI), logged on every tick until the process is
// interrupted.
var wsCmd = &cobra.Command{
	Use:   "ws <host>",
	Short: "WebSocket handshake and ping test",
	Args:  cobra.ExactArgs(1),
	RunE:  runWs,
}

// init registers the ws subcommand and its flags on the root command.
func init() {
	f := wsCmd.Flags()
	f.String("bind", "0.0.0.0", "bind address")
	f.Duration("timeout", 3*time.Second, "handshake + ping/pong timeout (e.g. 500ms, 3s, 1m)")
	f.Duration("interval", time.Second, "interval between attempts (e.g. 500ms, 3s, 1m)")
	f.String("domain", "", "Host header override")
	f.String("dns", "", "DNS server for host resolution (e.g. 8.8.8.8)")
	rootCmd.AddCommand(wsCmd)
}

// runWs reads the target host (a full URI) from args and the ws flags,
// wires up signal-based cancellation and runs the handshake+ping loop until
// ctx is cancelled.
func runWs(cmd *cobra.Command, args []string) error {
	host := args[0]
	f := cmd.Flags()
	bind, _ := f.GetString("bind")
	timeout, _ := f.GetDuration("timeout")
	interval, _ := f.GetDuration("interval")
	domain, _ := f.GetString("domain")
	ns, _ := f.GetString("dns")

	if err := validateTimeout(timeout); err != nil {
		return err
	}
	if err := validateInterval(interval); err != nil {
		return err
	}
	if err := validateDomain(domain); err != nil {
		return err
	}
	if err := validateDNS(ns); err != nil {
		return err
	}

	t, err := ws.NewTest(bind, host, timeout, interval, domain, ns)
	if err != nil {
		return err
	}

	ctx := signal.ContextWithSignal(context.Background())
	return t.Execute(ctx)
}
