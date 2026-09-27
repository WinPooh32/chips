// Package commands builds the chips cobra command tree on top of the issues
// store.
package commands

import (
	"os"

	"github.com/spf13/cobra"
)

const (
	storeRoot    = ".chips"
	storeRootEnv = "CHIPS_ROOT"
	twoArgs      = 2
	readyState   = "ready"
)

// Version is stamped at release build time via -ldflags -X.
//
//nolint:gochecknoglobals // must be a package-level var for -ldflags -X stamping
var Version = "dev"

// storeDir returns the store root: $CHIPS_ROOT when set, else .chips.
func storeDir() string {
	if v := os.Getenv(storeRootEnv); v != "" {
		return v
	}

	return storeRoot
}

// NewRoot builds the chips root command.
func NewRoot() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = "chips"
	cmd.Short = "chips is a local task tracker"
	cmd.Version = Version
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.AddCommand(
		newCreateCmd(),
		newReadyCmd(),
		newShowCmd(),
		newClaimCmd(),
		newStatusCmd(),
		newDepCmd(),
		newCloseCmd(),
	)

	return cmd
}
