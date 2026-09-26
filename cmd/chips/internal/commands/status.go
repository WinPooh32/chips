package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/WinPooh32/chips/cmd/chips/internal/lib/issues"
)

func newStatusCmd() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = "status <id> <open|in_progress|done>"
	cmd.Short = "Move an issue to a new status"
	cmd.Args = cobra.ExactArgs(twoArgs)
	cmd.RunE = func(c *cobra.Command, args []string) error {
		note, err := c.Flags().GetString("note")
		if err != nil {
			return fmt.Errorf("read --note flag: %w", err)
		}

		to, err := parseStatus(args[1])
		if err != nil {
			return err
		}

		store := issues.NewStore(storeDir())
		if err := store.Status(args[0], to, note); err != nil {
			return fmt.Errorf("move issue: %w", err)
		}

		fmt.Fprintf(c.OutOrStdout(), "moved %s to %s\n", args[0], to)

		return nil
	}
	cmd.Flags().String("note", "", "note appended to the issue body")

	return cmd
}

// parseStatus converts a status name into a status.
func parseStatus(v string) (issues.Status, error) {
	switch issues.Status(v) {
	case issues.Open, issues.InProgress, issues.Done:
		return issues.Status(v), nil
	case issues.Blocked:
		return "", fmt.Errorf("invalid status %q", v)
	default:
		return "", fmt.Errorf("invalid status %q", v)
	}
}
