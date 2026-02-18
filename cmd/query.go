package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

// queryTicket is a JSON-serializable representation of a ticket.
// Slices are always present as arrays (never null).
type queryTicket struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Status      string   `json:"status,omitempty"`
	Type        string   `json:"type,omitempty"`
	Priority    int      `json:"priority"`
	Assignee    string   `json:"assignee,omitempty"`
	Created     string   `json:"created,omitempty"`
	Parent      string   `json:"parent,omitempty"`
	ExternalRef string   `json:"external_ref,omitempty"`
	Design      string   `json:"design,omitempty"`
	Acceptance  string   `json:"acceptance,omitempty"`
	Description string   `json:"description,omitempty"`
	Deps        []string `json:"deps"`
	Links       []string `json:"links"`
	Tags        []string `json:"tags"`
}

func toQueryTicket(t *tickets.Ticket) queryTicket {
	q := queryTicket{
		ID:          t.ID,
		Title:       t.Title,
		Status:      t.Status,
		Type:        t.Type,
		Priority:    t.Priority,
		Assignee:    t.Assignee,
		Created:     t.Created,
		Parent:      t.Parent,
		ExternalRef: t.ExternalRef,
		Design:      t.Design,
		Acceptance:  t.Acceptance,
		Description: t.Description,
		Deps:        t.Deps,
		Links:       t.Links,
		Tags:        t.Tags,
	}

	// Ensure slices are never null in JSON output
	if q.Deps == nil {
		q.Deps = []string{}
	}
	if q.Links == nil {
		q.Links = []string{}
	}
	if q.Tags == nil {
		q.Tags = []string{}
	}

	return q
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Output tickets as JSONL",
	Long:  `Output all tickets as JSON Lines (one JSON object per line) with all frontmatter fields. Supports filtering by status, type, assignee, and tag.`,
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
		typeFilter, _ := cmd.Flags().GetString("type")
		assigneeFilter, _ := cmd.Flags().GetString("assignee")
		tagFilter, _ := cmd.Flags().GetString("tag")

		for _, t := range allItems {
			// Apply --status filter
			if statusFilter != "" {
				if statusFilter == "open" {
					if t.Status != "" && t.Status != "open" {
						continue
					}
				} else if t.Status != statusFilter {
					continue
				}
			}

			// Apply --type filter
			if typeFilter != "" && t.Type != typeFilter {
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

			q := toQueryTicket(t)
			jsonBytes, err := json.Marshal(q)
			if err != nil {
				return fmt.Errorf("failed to marshal ticket %s: %w", t.ID, err)
			}
			fmt.Println(string(jsonBytes))
		}

		return nil
	},
}

func init() {
	queryCmd.Flags().String("status", "", "Filter by status (open, in_progress, closed)")
	queryCmd.Flags().String("type", "", "Filter by type (bug, feature, task, epic, chore)")
	queryCmd.Flags().StringP("assignee", "a", "", "Filter by assignee")
	queryCmd.Flags().StringP("tag", "T", "", "Filter by tag")
	rootCmd.AddCommand(queryCmd)
}
