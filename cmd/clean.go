package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/config"
	"github.com/juanibiapina/todo/internal/store"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Delete all checked items",
	Long:  `Remove all checked todo items for the current directory.`,
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

		n, err := s.Clean(dir)
		if err != nil {
			return err
		}

		fmt.Printf("Deleted %d item(s)\n", n)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
