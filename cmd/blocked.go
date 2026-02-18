package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var blockedCmd = &cobra.Command{
	Use:   "blocked",
	Short: "Show tickets blocked by unclosed dependencies",
	Long:  `Show open/in_progress tickets that have at least one unclosed dependency, sorted by priority then ID.`,
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

		// Build a map of ticket ID -> status for dep checking
		statusMap := make(map[string]string)
		for _, t := range allItems {
			statusMap[t.ID] = t.Status
		}

		assigneeFilter, _ := cmd.Flags().GetString("assignee")
		tagFilter, _ := cmd.Flags().GetString("tag")

		type blockedTicket struct {
			ticket           *tickets.Ticket
			unclosedBlockers []string
		}

		var blocked []blockedTicket
		for _, t := range allItems {
			// Only open/in_progress tickets (not closed)
			if t.Status == "closed" {
				continue
			}

			// Find unclosed deps
			var unclosed []string
			for _, depID := range t.Deps {
				depStatus, exists := statusMap[depID]
				if exists && depStatus != "closed" {
					unclosed = append(unclosed, depID)
				}
				// Missing deps treated as non-blocking
			}

			// Must have at least one unclosed dep to be "blocked"
			if len(unclosed) == 0 {
				continue
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

			blocked = append(blocked, blockedTicket{ticket: t, unclosedBlockers: unclosed})
		}

		// Sort by priority ascending, then by ID ascending
		sort.Slice(blocked, func(i, j int) bool {
			if blocked[i].ticket.Priority != blocked[j].ticket.Priority {
				return blocked[i].ticket.Priority < blocked[j].ticket.Priority
			}
			return blocked[i].ticket.ID < blocked[j].ticket.ID
		})

		for _, bt := range blocked {
			fmt.Println(formatBlockedLine(bt.ticket, bt.unclosedBlockers))
		}

		return nil
	},
}

func init() {
	blockedCmd.Flags().StringP("assignee", "a", "", "Filter by assignee")
	blockedCmd.Flags().StringP("tag", "T", "", "Filter by tag")
	rootCmd.AddCommand(blockedCmd)
}
