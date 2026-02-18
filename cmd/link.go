package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link <id> <id> [id...]",
	Short: "Create bidirectional links between tickets",
	Long:  `Create bidirectional links between two or more tickets. All tickets are linked to each other. The operation is idempotent.`,
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		err = tickets.AddLink(dir, args)
		if err != nil {
			return err
		}

		fmt.Printf("Linked tickets: %s\n", strings.Join(args, ", "))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(linkCmd)
}
