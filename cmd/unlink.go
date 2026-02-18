package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var unlinkCmd = &cobra.Command{
	Use:   "unlink <id> <target-id>",
	Short: "Remove a bidirectional link between tickets",
	Long:  `Remove a bidirectional link between two tickets. The link is removed from both sides. The operation is idempotent.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		targetID := args[1]

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		err = tickets.RemoveLink(dir, id, targetID)
		if err != nil {
			return err
		}

		fmt.Printf("Unlinked %s and %s\n", id, targetID)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(unlinkCmd)
}
