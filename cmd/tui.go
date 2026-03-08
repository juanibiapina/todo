package cmd

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/juanibiapina/todo/internal/store"
	"github.com/juanibiapina/todo/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive todo list",
	Long:  `Open an interactive terminal UI for managing todo items.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		s, err := store.Open(store.DefaultDBPath())
		if err != nil {
			return err
		}
		defer s.Close()

		if err := s.CleanIfNewDay(dir); err != nil {
			return err
		}

		m := tui.New(s, dir)
		p := tea.NewProgram(m, tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
