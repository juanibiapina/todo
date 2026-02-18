package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <id> <status>",
	Short: "Set the status of a ticket",
	Long:  `Set the status of a ticket. Valid statuses: open, in_progress, closed.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		status := args[1]

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		title, err := tickets.SetStatus(dir, id, status)
		if err != nil {
			return err
		}

		fmt.Printf("Status of %s set to %s\n", title, status)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
