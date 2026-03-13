package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/store"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List todo items",
	Long:  `List all todo items for the current directory.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		s, err := store.Open(store.DefaultDBPath())
		if err != nil {
			return err
		}
		defer s.Close()

		if err := s.CleanIfNewDay(dir); err != nil {
			return err
		}

		items, err := s.List(dir)
		if err != nil {
			return err
		}

		for _, item := range items {
			if item.IsSection {
				fmt.Println("──────────")
				continue
			}
			check := " "
			if item.Checked {
				check = "x"
			}
			fmt.Printf("%d [%s] %s\n", item.ID, check, item.Text)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
