package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/juanibiapina/todo/internal/config"
	"github.com/juanibiapina/todo/internal/store"
	"github.com/spf13/cobra"
)

var addSection bool

var addCmd = &cobra.Command{
	Use:   "add <text>...",
	Short: "Add a new todo item",
	Long:  `Add a new unchecked todo item or section divider with the given text.`,
	Args: cobra.MatchAll(func(cmd *cobra.Command, args []string) error {
		if addSection {
			return nil
		}
		if len(args) < 1 {
			return fmt.Errorf("requires at least 1 arg(s), only received 0")
		}
		return nil
	}),
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

		if addSection {
			title := strings.TrimSpace(strings.Join(args, " "))
			_, err = s.AddSection(dir, title)
			if err != nil {
				return err
			}
			return nil
		}

		text := strings.Join(args, " ")
		item, err := s.Add(dir, text)
		if err != nil {
			return err
		}

		fmt.Printf("%d %s\n", item.ID, item.Text)
		return nil
	},
}

func init() {
	addCmd.Flags().BoolVarP(&addSection, "section", "s", false, "Add a section divider instead of a todo item")
	rootCmd.AddCommand(addCmd)
}
