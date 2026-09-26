package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/WinPooh32/chips/cmd/chips/internal/lib/issues"
)

func newReadyCmd() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = readyState
	cmd.Short = "List open issues that are ready to start"
	cmd.RunE = func(c *cobra.Command, _ []string) error {
		store := issues.NewStore(storeDir())

		ready, warns, err := store.Ready()
		if err != nil {
			return fmt.Errorf("list ready issues: %w", err)
		}

		for _, w := range warns {
			fmt.Fprintln(c.ErrOrStderr(), w)
		}

		for _, iss := range ready {
			fmt.Fprintf(c.OutOrStdout(), "%s  %s  %s\n", iss.ID, iss.Type, iss.Title)
		}

		return nil
	}

	return cmd
}
