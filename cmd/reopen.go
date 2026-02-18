package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var reopenCmd = &cobra.Command{
	Use:   "reopen <id>",
	Short: "Reopen a ticket (set status to open)",
	Long:  `Set a ticket's status back to open.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		title, err := tickets.SetStatus(dir, id, "open")
		if err != nil {
			return err
		}

		fmt.Printf("Reopened ticket: %s\n", title)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(reopenCmd)
}
