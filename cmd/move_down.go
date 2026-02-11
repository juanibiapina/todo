package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var moveDownCmd = &cobra.Command{
	Use:   "move-down <title|id>",
	Short: "Move a ticket down (swap with next)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		title, err := tickets.MoveDown(dir, args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Moved down: %s\n", title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(moveDownCmd)
}
