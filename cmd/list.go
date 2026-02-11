package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tickets",
	Long: `List all tickets with their state icon, ID, and title.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		items, err := tickets.List(dir)
		if err != nil {
			return err
		}

		if len(items) == 0 {
			fmt.Println("No tickets")
			return nil
		}

		for _, t := range items {
			fmt.Println(formatTicketLine(t))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
