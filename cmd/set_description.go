package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var setDescriptionCmd = &cobra.Command{
	Use:   "set-description <id> [description]",
	Short: "Set or replace a ticket's description",
	Long: `Set or replace the description of a ticket.

Description can be provided as:
  1. A positional argument (for simple text)
  2. Via stdin (for multi-line content with special characters)

When stdin is not a terminal, the description is read from stdin.
This allows heredocs and pipes for rich content:

  todo set-description aBc <<'EOF'
  The ` + "`" + `handleAuth` + "`" + ` function needs work.

  ` + "```" + `go
  func handleAuth() {
      // new implementation
  }
  ` + "```" + `
  EOF`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]
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

		if description == "" {
			return fmt.Errorf("no description provided (pass as argument or via stdin)")
		}

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		title, err := tickets.SetDescription(dir, ref, description)
		if err != nil {
			return err
		}

		fmt.Printf("Updated description: %s\n", title)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setDescriptionCmd)
}
