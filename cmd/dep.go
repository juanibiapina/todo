package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var depCmd = &cobra.Command{
	Use:   "dep <id> <dep-id>",
	Short: "Add a dependency to a ticket",
	Long:  `Add a dependency from one ticket to another. Both tickets must exist. The operation is idempotent.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		depID := args[1]

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		err = tickets.AddDep(dir, id, depID)
		if err != nil {
			return err
		}

		fmt.Printf("Added dependency %s to %s\n", depID, id)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(depCmd)
}
