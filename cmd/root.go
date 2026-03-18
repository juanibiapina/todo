package cmd

import (
	"os"

	"github.com/juanibiapina/todo/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "todo",
	Short: "A simple per-directory todo list",
	Long:  `A CLI for managing simple todo items stored in a Markdown file, scoped per directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default command: list
		return listCmd.RunE(cmd, args)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version.Version
	rootCmd.SilenceUsage = true
}
