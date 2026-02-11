package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a ticket as done (remove it)",
	Long:  `Remove a ticket from the file, marking it as complete.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		title, err := tickets.Done(dir, ref)
		if err != nil {
			return err
		}

		fmt.Printf("Completed ticket: %s\n", title)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(doneCmd)
}
