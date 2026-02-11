package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a ticket's full details",
	Long:  `Show the full details of a ticket, including its front matter and description.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		ticket, err := tickets.Show(dir, ref)
		if err != nil {
			return err
		}

		fmt.Print(ticket.FullString())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
