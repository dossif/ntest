package cmd

import (
	"context"
	"time"

	"github.com/dossif/ntest/internal/signal"
	"github.com/dossif/ntest/internal/tcp"
	"github.com/spf13/cobra"
)

// tcpCmd implements `ntest tcp <host>`: repeated TCP dial attempts to
// host:--port, logged on every tick until the process is interrupted.
var tcpCmd = &cobra.Command{
	Use:   "tcp <host>",
	Short: "TCP connection test",
	Args:  cobra.ExactArgs(1),
	RunE:  runTcp,
}

// init registers the tcp subcommand and its flags on the root command.
func init() {
	f := tcpCmd.Flags()
	f.Int("port", 80, "target port")
	f.String("bind", "0.0.0.0", "bind address")
	f.Duration("timeout", 3*time.Second, "dial timeout (e.g. 500ms, 3s, 1m)")
	f.Duration("interval", time.Second, "interval between dial attempts (e.g. 500ms, 3s, 1m)")
	f.String("dns", "", "DNS server for host resolution (e.g. 8.8.8.8)")
	rootCmd.AddCommand(tcpCmd)
}

// runTcp reads the target host from args and the tcp flags, wires up
// signal-based cancellation and runs the dial loop until ctx is cancelled.
func runTcp(cmd *cobra.Command, args []string) error {
	host := args[0]
	f := cmd.Flags()
	port, _ := f.GetInt("port")
	bind, _ := f.GetString("bind")
	timeout, _ := f.GetDuration("timeout")
	interval, _ := f.GetDuration("interval")
	ns, _ := f.GetString("dns")

	if err := validatePort(port); err != nil {
		return err
	}
	if err := validateTimeout(timeout); err != nil {
		return err
	}
	if err := validateInterval(interval); err != nil {
		return err
	}
	if err := validateDNS(ns); err != nil {
		return err
	}

	t, err := tcp.NewTest(bind, host, port, timeout, interval, ns)
	if err != nil {
		return err
	}

	ctx := signal.ContextWithSignal(context.Background())
	return t.Execute(ctx)
}
