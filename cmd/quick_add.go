package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var quickAddCmd = &cobra.Command{
	Use:   "quick-add",
	Short: "Quick add a ticket via interactive prompt",
	Long: `Opens a one-line prompt to quickly add a ticket.

Designed for tmux popup shortcuts — opens a gum input,
adds the ticket, and exits immediately.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		gum := exec.Command("gum", "input", "--placeholder", "Add a ticket...", "--width", "0")
		gum.Stdin = os.Stdin
		gum.Stderr = os.Stderr
		out, err := gum.Output()
		if err != nil {
			return nil // user cancelled
		}

		title := strings.TrimSpace(string(out))
		if title == "" {
			return nil
		}

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		ticket, err := tickets.Add(dir, title, "")
		if err != nil {
			return err
		}

		fmt.Printf("Added %s: %s\n", ticket.ID, ticket.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(quickAddCmd)
}
