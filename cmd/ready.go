package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var readyCmd = &cobra.Command{
	Use:   "ready",
	Short: "Show tickets ready to work on",
	Long:  `Show open/in_progress tickets with all deps closed or no deps, sorted by priority then ID.`,
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

		var ready []*tickets.Ticket
		for _, t := range allItems {
			// Only open/in_progress tickets (not closed)
			if t.Status == "closed" {
				continue
			}

			// Check all deps are closed or missing (non-blocking)
			allDepsDone := true
			for _, depID := range t.Deps {
				depStatus, exists := statusMap[depID]
				if exists && depStatus != "closed" {
					allDepsDone = false
					break
				}
				// Missing deps treated as non-blocking
			}
			if !allDepsDone {
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

			ready = append(ready, t)
		}

		// Sort by priority ascending, then by ID ascending
		sort.Slice(ready, func(i, j int) bool {
			if ready[i].Priority != ready[j].Priority {
				return ready[i].Priority < ready[j].Priority
			}
			return ready[i].ID < ready[j].ID
		})

		for _, t := range ready {
			fmt.Println(formatReadyLine(t))
		}

		return nil
	},
}

func init() {
	readyCmd.Flags().StringP("assignee", "a", "", "Filter by assignee")
	readyCmd.Flags().StringP("tag", "T", "", "Filter by tag")
	rootCmd.AddCommand(readyCmd)
}
