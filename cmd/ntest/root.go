// Command ntest is a CLI tool for continuous ICMP, TCP, HTTP and WebSocket
// reachability testing. This file defines the Cobra root command and the
// process entry point; one subcommand per test type (icmp, tcp, http, ws)
// lives in its own cmd_*.go file in this same package, registered via
// init().
package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// appVersion is overridden at build time via:
//
//	-ldflags "-X main.appVersion=<version>"
//
// "main", not a module-relative import path — this is the actual main
// package being built (cmd/ declares "package main"), and the linker
// addresses it as literally "main" regardless of which directory it lives
// in. See the Makefile, which derives the version from `git describe --tags`.
var appVersion = "0.0.0"

var rootCmd = &cobra.Command{
	Use:     "ntest",
	Short:   fmt.Sprintf("Network testing tool v%s", appVersion),
	Version: appVersion,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
