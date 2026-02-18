package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var depTreeCmd = &cobra.Command{
	Use:   "tree <id>",
	Short: "Show dependency tree for a ticket",
	Long:  `Display a tree of dependencies for a ticket using box-drawing characters.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		full, _ := cmd.Flags().GetBool("full")

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		output, err := tickets.DepTree(dir, id, full)
		if err != nil {
			return err
		}

		fmt.Println(output)

		return nil
	},
}

func init() {
	depTreeCmd.Flags().Bool("full", false, "Show full tree without deduplication")
	depCmd.AddCommand(depTreeCmd)
}
