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
  - Left panel: Ticket list with IDs
  - Right panel: Selected ticket details and description

KEYBINDINGS:

  Ticket List:
    ↑/k ↓/j   Move cursor
    g/G       First/last ticket
    a         Add new ticket
    d         Mark done (remove)
    K/J       Reorder up/down

  Detail Panel:
    ↑/k ↓/j   Scroll description
    g/G       Top/bottom

  General:
    tab       Switch panels
    ?         Show help
    q         Quit`,
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
