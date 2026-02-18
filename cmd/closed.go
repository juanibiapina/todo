package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var closedCmd = &cobra.Command{
	Use:   "closed",
	Short: "Show recently closed tickets",
	Long:  `Show closed tickets sorted by file modification time (most recent first), with an optional limit.`,
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

		assigneeFilter, _ := cmd.Flags().GetString("assignee")
		tagFilter, _ := cmd.Flags().GetString("tag")
		limit, _ := cmd.Flags().GetInt("limit")

		type closedTicket struct {
			ticket *tickets.Ticket
			mtime  int64
		}

		var closed []closedTicket
		for _, t := range allItems {
			// Only closed tickets
			if t.Status != "closed" {
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

			// Stat the file for mtime
			path := filepath.Join(tickets.DirPath(dir), t.ID+".md")
			info, err := os.Stat(path)
			if err != nil {
				continue // skip if stat fails
			}

			closed = append(closed, closedTicket{ticket: t, mtime: info.ModTime().UnixNano()})
		}

		// Sort by mtime descending (most recently modified first)
		sort.Slice(closed, func(i, j int) bool {
			return closed[i].mtime > closed[j].mtime
		})

		// Apply limit
		if limit > 0 && len(closed) > limit {
			closed = closed[:limit]
		}

		for _, ct := range closed {
			fmt.Println(formatClosedLine(ct.ticket))
		}

		return nil
	},
}

func init() {
	closedCmd.Flags().IntP("limit", "n", 20, "Maximum number of tickets to show")
	closedCmd.Flags().StringP("assignee", "a", "", "Filter by assignee")
	closedCmd.Flags().StringP("tag", "T", "", "Filter by tag")
	rootCmd.AddCommand(closedCmd)
}
