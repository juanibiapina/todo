package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var moveUpCmd = &cobra.Command{
	Use:   "move-up <id>",
	Short: "Move a ticket up (swap with previous)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		title, err := tickets.MoveUp(dir, args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Moved up: %s\n", title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(moveUpCmd)
}
