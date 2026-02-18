package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [title] [description]",
	Short: "Add a new ticket",
	Long: `Add a new ticket with the given title.

Title defaults to "Untitled" if not provided.

Description can be provided as:
  1. The -d/--description flag
  2. A positional argument (for simple text)
  3. Via stdin (for multi-line content with special characters)

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

  echo "Simple description" | todo add 'Fix bug'`,
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Title: first arg or "Untitled"
		title := "Untitled"
		if len(args) > 0 {
			title = args[0]
		}

		// Description priority: -d flag > 2nd positional arg > stdin
		descFlag, _ := cmd.Flags().GetString("description")
		var description string

		if descFlag != "" {
			description = descFlag
		} else if len(args) > 1 {
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

		// Get flag values
		ticketType, _ := cmd.Flags().GetString("type")
		priority, _ := cmd.Flags().GetInt("priority")
		assignee, _ := cmd.Flags().GetString("assignee")
		externalRef, _ := cmd.Flags().GetString("external-ref")
		parent, _ := cmd.Flags().GetString("parent")
		design, _ := cmd.Flags().GetString("design")
		acceptance, _ := cmd.Flags().GetString("acceptance")
		tagsStr, _ := cmd.Flags().GetString("tags")

		// Validate type
		validTypes := map[string]bool{
			"bug": true, "feature": true, "task": true, "epic": true, "chore": true,
		}
		if !validTypes[ticketType] {
			return fmt.Errorf("invalid type %q: must be one of bug, feature, task, epic, chore", ticketType)
		}

		// Validate priority
		if priority < 0 || priority > 4 {
			return fmt.Errorf("invalid priority %d: must be between 0 and 4", priority)
		}

		// Default assignee to git user.name if not set
		if !cmd.Flags().Changed("assignee") {
			gitName, err := exec.Command("git", "config", "user.name").Output()
			if err == nil {
				assignee = strings.TrimSpace(string(gitName))
			}
		}

		// Parse tags
		var tags []string
		if tagsStr != "" {
			for _, tag := range strings.Split(tagsStr, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tags = append(tags, tag)
				}
			}
		}

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		t := &tickets.Ticket{
			Title:       title,
			Description: description,
			Type:        ticketType,
			Priority:    priority,
			Assignee:    assignee,
			ExternalRef: externalRef,
			Parent:      parent,
			Design:      design,
			Acceptance:  acceptance,
			Tags:        tags,
		}

		ticket, err := tickets.Add(dir, t)
		if err != nil {
			return err
		}

		fmt.Printf("Added %s %s\n", cliID(ticket.ID), ticket.Title)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringP("description", "d", "", "Ticket description")
	addCmd.Flags().StringP("type", "t", "task", "Ticket type (bug/feature/task/epic/chore)")
	addCmd.Flags().IntP("priority", "p", 2, "Priority (0-4)")
	addCmd.Flags().StringP("assignee", "a", "", "Assignee (defaults to git user.name)")
	addCmd.Flags().String("external-ref", "", "External reference (e.g. JIRA-123)")
	addCmd.Flags().String("parent", "", "Parent ticket ID (must exist)")
	addCmd.Flags().String("design", "", "Design notes")
	addCmd.Flags().String("acceptance", "", "Acceptance criteria")
	addCmd.Flags().String("tags", "", "Comma-separated tags")
}
