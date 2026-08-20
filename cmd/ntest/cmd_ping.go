package main

import (
	"context"
	"time"

	"github.com/dossif/ntest/internal/icmp"
	"github.com/dossif/ntest/internal/signal"
	"github.com/spf13/cobra"
)

// pingCmd implements `ntest ping <host>`: repeated ICMP echo requests to
// host, logged on every tick until the process is interrupted.
var pingCmd = &cobra.Command{
	Use:   "ping <host>",
	Short: "ICMP ping test",
	Args:  cobra.ExactArgs(1),
	RunE:  runPing,
}

// init registers the ping subcommand and its flags on the root command.
func init() {
	f := pingCmd.Flags()
	f.String("bind", "0.0.0.0", "bind address")
	f.Duration("timeout", 3*time.Second, "request timeout (e.g. 500ms, 3s, 1m)")
	f.Duration("interval", time.Second, "interval between pings (e.g. 500ms, 3s, 1m)")
	f.Duration("warn", 100*time.Millisecond, "RTT warn threshold (e.g. 500ms, 3s, 1m)")
	f.String("dns", "", "DNS server for host resolution (e.g. 8.8.8.8)")
	f.Bool("graph", false, "show a live ASCII RTT chart instead of per-attempt log lines")
	rootCmd.AddCommand(pingCmd)
}

// runPing reads the target host from args and the ping flags, wires up
// signal-based cancellation and runs the ICMP ping loop until ctx is
// cancelled. The implementation lives in internal/icmp — "ping" is just the
// user-facing command name.
func runPing(cmd *cobra.Command, args []string) error {
	host := args[0]
	f := cmd.Flags()
	bind, _ := f.GetString("bind")
	timeout, _ := f.GetDuration("timeout")
	interval, _ := f.GetDuration("interval")
	warn, _ := f.GetDuration("warn")
	dns, _ := f.GetString("dns")
	graph, _ := f.GetBool("graph")

	if err := validateTimeout(timeout); err != nil {
		return err
	}
	if err := validateInterval(interval); err != nil {
		return err
	}
	if err := validateWarn(warn); err != nil {
		return err
	}
	if err := validateDNS(dns); err != nil {
		return err
	}

	t, err := icmp.NewTest(bind, host, timeout, interval, warn, dns, graph)
	if err != nil {
		return err
	}

	ctx := signal.ContextWithSignal(context.Background())
	return t.Execute(ctx)
}
