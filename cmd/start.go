package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <id>",
	Short: "Start working on a ticket (set status to in_progress)",
	Long:  `Set a ticket's status to in_progress.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		title, err := tickets.SetStatus(dir, id, "in_progress")
		if err != nil {
			return err
		}

		fmt.Printf("Started ticket: %s\n", title)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
