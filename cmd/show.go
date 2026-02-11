package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <title|id>",
	Short: "Show a ticket's full details",
	Long: `Show the full details of a ticket, including its front matter and description.

The ticket can be referenced by its 3-character ID or its title.`,
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
