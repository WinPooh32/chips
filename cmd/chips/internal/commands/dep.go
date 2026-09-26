package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/WinPooh32/chips/cmd/chips/internal/lib/issues"
)

func newDepCmd() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = "dep"
	cmd.Short = "Manage issue dependencies"
	cmd.AddCommand(newDepAddCmd(), newDepRmCmd())

	return cmd
}

func newDepAddCmd() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = "add <id> <dep-id>"
	cmd.Short = "Add a dependency to an issue"
	cmd.Args = cobra.ExactArgs(twoArgs)
	cmd.RunE = func(c *cobra.Command, args []string) error {
		store := issues.NewStore(storeDir())
		if err := store.DepAdd(args[0], args[1]); err != nil {
			return fmt.Errorf("add dependency: %w", err)
		}

		fmt.Fprintf(c.OutOrStdout(), "%s is blocked by %s\n", args[0], args[1])

		return nil
	}

	return cmd
}

func newDepRmCmd() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = "rm <id> <dep-id>"
	cmd.Short = "Remove a dependency from an issue"
	cmd.Args = cobra.ExactArgs(twoArgs)
	cmd.RunE = func(c *cobra.Command, args []string) error {
		store := issues.NewStore(storeDir())
		if err := store.DepRemove(args[0], args[1]); err != nil {
			return fmt.Errorf("remove dependency: %w", err)
		}

		fmt.Fprintf(c.OutOrStdout(), "%s is no longer blocked by %s\n", args[0], args[1])

		return nil
	}

	return cmd
}
