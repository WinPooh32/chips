package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/WinPooh32/chips/cmd/chips/internal/lib/issues"
)

func newClaimCmd() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = "claim <id>"
	cmd.Short = "Claim an open issue"
	cmd.Args = cobra.ExactArgs(1)
	cmd.RunE = func(c *cobra.Command, args []string) error {
		store := issues.NewStore(storeDir())
		if err := store.Claim(args[0]); err != nil {
			return fmt.Errorf("claim issue: %w", err)
		}

		fmt.Fprintf(c.OutOrStdout(), "claimed %s\n", args[0])

		return nil
	}

	return cmd
}
