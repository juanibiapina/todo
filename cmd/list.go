package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tickets",
	Long:  `List all tickets with their ID and title.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		allItems, err := tickets.List(dir)
		if err != nil {
			return err
		}

		statusFilter, _ := cmd.Flags().GetString("status")
		assigneeFilter, _ := cmd.Flags().GetString("assignee")
		tagFilter, _ := cmd.Flags().GetString("tag")

		var items []*tickets.Ticket
		for _, t := range allItems {
			// Default behavior: hide closed tickets unless --status is specified
			if statusFilter == "" && t.Status == "closed" {
				continue
			}

			// Apply --status filter
			// "open" matches both explicit "open" and empty status (default)
			if statusFilter != "" {
				if statusFilter == "open" {
					if t.Status != "" && t.Status != "open" {
						continue
					}
				} else if t.Status != statusFilter {
					continue
				}
			}

			// Apply --assignee filter
			if assigneeFilter != "" && t.Assignee != assigneeFilter {
				continue
			}

			// Apply --tag filter
			if tagFilter != "" {
				found := false
				for _, tag := range t.Tags {
					if tag == tagFilter {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			items = append(items, t)
		}

		for _, t := range items {
			fmt.Println(formatTicketLine(t))
		}

		return nil
	},
}

func init() {
	listCmd.Flags().String("status", "", "Filter by status (open, in_progress, closed)")
	listCmd.Flags().StringP("assignee", "a", "", "Filter by assignee")
	listCmd.Flags().StringP("tag", "T", "", "Filter by tag")
	rootCmd.AddCommand(listCmd)
}
