package commands

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/WinPooh32/chips/cmd/chips/internal/lib/issues"
)

func newCloseCmd() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = "close <id>..."
	cmd.Short = "Close one or more issues"
	cmd.Args = cobra.MinimumNArgs(1)
	cmd.RunE = func(c *cobra.Command, args []string) error {
		reason, err := c.Flags().GetString("reason")
		if err != nil {
			return fmt.Errorf("read --reason flag: %w", err)
		}

		if reason == "" {
			return errors.New("--reason is required")
		}

		store := issues.NewStore(storeDir())
		if err := store.Close(args, reason); err != nil {
			return fmt.Errorf("close issues: %w", err)
		}

		for _, id := range args {
			fmt.Fprintf(c.OutOrStdout(), "closed %s\n", id)
		}

		return nil
	}
	cmd.Flags().String("reason", "", "why the issue is closed")

	return cmd
}
