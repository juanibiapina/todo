package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/juanibiapina/todo/internal/config"
	"github.com/juanibiapina/todo/internal/store"
	"github.com/spf13/cobra"
)

var uncheckCmd = &cobra.Command{
	Use:   "uncheck <id>",
	Short: "Uncheck a todo item",
	Long:  `Mark a todo item as unchecked.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid item ID: %s", args[0])
		}

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		s, err := store.Open(config.DefaultFilePath())
		if err != nil {
			return err
		}
		defer s.Close()

		if err := s.Uncheck(dir, id); err != nil {
			return err
		}

		item, err := s.Get(dir, id)
		if err != nil {
			return err
		}

		fmt.Printf("%d [ ] %s\n", item.ID, item.Text)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uncheckCmd)
}
