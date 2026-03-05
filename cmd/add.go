package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/juanibiapina/todo/internal/store"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <text>...",
	Short: "Add a new todo item",
	Long:  `Add a new unchecked todo item with the given text.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		s, err := store.Open(store.DefaultDBPath())
		if err != nil {
			return err
		}
		defer s.Close()

		item, err := s.Add(dir, text)
		if err != nil {
			return err
		}

		fmt.Printf("%d %s\n", item.ID, item.Text)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
