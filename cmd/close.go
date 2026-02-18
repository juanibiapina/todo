package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var closeCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close a ticket (set status to closed)",
	Long:  `Set a ticket's status to closed.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		title, err := tickets.SetStatus(dir, id, "closed")
		if err != nil {
			return err
		}

		fmt.Printf("Closed ticket: %s\n", title)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(closeCmd)
}
