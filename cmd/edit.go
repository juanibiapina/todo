package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/juanibiapina/todo/internal/store"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <id> <text>...",
	Short: "Edit a todo item",
	Long:  `Replace the text of a todo item.`,
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid item ID: %s", args[0])
		}

		text := strings.Join(args[1:], " ")

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		s, err := store.Open(store.DefaultDBPath())
		if err != nil {
			return err
		}
		defer s.Close()

		if err := s.Edit(dir, id, text); err != nil {
			return err
		}

		item, err := s.Get(dir, id)
		if err != nil {
			return err
		}

		check := " "
		if item.Checked {
			check = "x"
		}
		fmt.Printf("%d [%s] %s\n", item.ID, check, item.Text)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
