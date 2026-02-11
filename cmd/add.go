package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <title> [description]",
	Short: "Add a new ticket",
	Long: `Add a new ticket with the given title.

Description can be provided as:
  1. A positional argument (for simple text)
  2. Via stdin (for multi-line content with special characters)

When stdin is not a terminal, the description is read from stdin.
This allows heredocs and pipes for rich content:

  todo add 'Fix auth' <<'EOF'
  The ` + "`" + `handleAuth` + "`" + ` function needs refactoring.

  ` + "```" + `go
  func handleAuth() {
      // fix this
  }
  ` + "```" + `
  EOF

  echo "Simple description" | todo add 'Fix bug'

A ticket without a description gets state "new".
A ticket with a description gets state "refined".`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		var description string

		if len(args) > 1 {
			description = args[1]
		} else {
			// Check if stdin has data (not a terminal)
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}
				description = strings.TrimRight(string(data), "\n")
			}
		}

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		ticket, err := tickets.Add(dir, title, description)
		if err != nil {
			return err
		}

		fmt.Printf("Added %s %s %s\n", cliStateIcon(ticket.State), cliID(ticket.ID), ticket.Title)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
