package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/config"
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

		s, err := store.Open(config.DefaultFilePath())
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
			active := ""
			if item.IsActive {
				active = "  (active)"
			}
			fmt.Printf("%d [%s] %s%s\n", item.ID, check, item.Text, active)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
