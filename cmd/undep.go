package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var undepCmd = &cobra.Command{
	Use:   "undep <id> <dep-id>",
	Short: "Remove a dependency from a ticket",
	Long:  `Remove a dependency from a ticket. Both tickets must exist. The operation is idempotent.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		depID := args[1]

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		err = tickets.RemoveDep(dir, id, depID)
		if err != nil {
			return err
		}

		fmt.Printf("Removed dependency %s from %s\n", depID, id)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(undepCmd)
}
