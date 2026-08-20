// Package cmd defines the ntest CLI: the Cobra root command and one
// subcommand per test type (icmp, tcp, http), registered via init().
package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// appVersion is overridden at build time via:
//
//	-ldflags "-X github.com/dossif/ntest/cmd.appVersion=<version>"
//
// see build.sh. Keep the package path in sync if this file ever moves.
var appVersion = "0.0.0"

var rootCmd = &cobra.Command{
	Use:     "ntest",
	Short:   fmt.Sprintf("Network testing tool v%s", appVersion),
	Version: appVersion,
}

// Execute runs the root command; called once from main().
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
