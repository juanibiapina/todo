package cmd

import (
	"os"

	"github.com/juanibiapina/todo/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	Long: `Launch a full-screen terminal interface for managing tickets.

The TUI provides a split-panel layout:
  - Left panel: Ticket list with priority/status badges
  - Right panel: Full ticket metadata, relationships, and description

KEYBINDINGS:

  Ticket List:
    ↑/k ↓/j   Move cursor
    g/G        First/last ticket
    a          Add new ticket (with defaults: type=task, priority=2)
    s          Start ticket (set status to in_progress)
    c/d        Close ticket (set status to closed)
    r          Reopen ticket (set status to open)
    e          Edit ticket in $EDITOR
    n          Add note to ticket
    space      Copy ticket to clipboard

  Views:
    1          All open tickets
    2          Ready tickets (all deps closed)
    3          Blocked tickets (has unclosed deps)
    4          Closed tickets (sorted by last modified)

  Detail Panel:
    ↑/k ↓/j   Scroll content
    g/G        Top/bottom
    ctrl+u/d   Half page up/down

  General:
    tab        Switch panels
    ?          Show help
    esc/q      Quit`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		return tui.Start(dir)
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
