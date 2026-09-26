package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/WinPooh32/chips/cmd/chips/internal/lib/issues"
)

func newShowCmd() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = "show <id>"
	cmd.Short = "Show an issue"
	cmd.Args = cobra.ExactArgs(1)
	cmd.RunE = func(c *cobra.Command, args []string) error {
		store := issues.NewStore(storeDir())

		iss, st, body, err := store.Read(args[0])
		if err != nil {
			return fmt.Errorf("read issue: %w", err)
		}

		out := c.OutOrStdout()
		fmt.Fprintf(out, "id: %s\n", iss.ID)
		fmt.Fprintf(out, "title: %s\n", iss.Title)
		fmt.Fprintf(out, "type: %s\n", iss.Type)
		fmt.Fprintf(out, "created-at: %s\n", iss.CreatedAt.Format(time.RFC3339))

		if iss.Parent != "" {
			fmt.Fprintf(out, "parent: %s\n", iss.Parent)
		}

		if len(iss.BlockedBy) > 0 {
			fmt.Fprintf(out, "blocked-by: %s\n", strings.Join(iss.BlockedBy, ", "))
		}

		if iss.ClaimedAt != nil {
			fmt.Fprintf(out, "claimed-at: %s\n", iss.ClaimedAt.Format(time.RFC3339))
		}

		fmt.Fprintf(out, "state: %s\n", stateOf(iss, st))

		if body != "" {
			fmt.Fprintln(out, body)
		}

		return nil
	}

	return cmd
}

// stateOf computes the readiness line shown by show.
func stateOf(iss *issues.Issue, st issues.Status) string {
	switch st {
	case issues.Blocked:
		return "blocked on: " + strings.Join(iss.BlockedBy, ", ")
	case issues.InProgress:
		return "in progress"
	case issues.Done:
		return "done"
	case issues.Open:
		return readyState
	default:
		return readyState
	}
}
