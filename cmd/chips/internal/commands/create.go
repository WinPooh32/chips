package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/WinPooh32/chips/cmd/chips/internal/lib/issues"
)

func newCreateCmd() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = "create <title>"
	cmd.Short = "Create a new issue"
	cmd.Args = cobra.ExactArgs(1)
	cmd.RunE = func(c *cobra.Command, args []string) error {
		typ, err := c.Flags().GetString("type")
		if err != nil {
			return fmt.Errorf("read --type flag: %w", err)
		}

		desc, err := c.Flags().GetString("desc")
		if err != nil {
			return fmt.Errorf("read --desc flag: %w", err)
		}

		parent, err := c.Flags().GetString("parent")
		if err != nil {
			return fmt.Errorf("read --parent flag: %w", err)
		}

		store := issues.NewStore(storeDir())

		iss, err := store.Create(args[0], typ, desc, parent)
		if err != nil {
			return fmt.Errorf("create issue: %w", err)
		}

		fmt.Fprintf(c.OutOrStdout(), "created %s %s\n", iss.ID, iss.Title)

		return nil
	}
	cmd.Flags().String("type", "", "issue type: bug, task, feature, or epic")
	cmd.Flags().String("desc", "", "description body")
	cmd.Flags().String("parent", "", "parent issue id")

	return cmd
}
