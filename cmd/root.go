package cmd

import (
	"os"

	"github.com/juanibiapina/todo/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "todo",
	Short: "Local ticket tracking in markdown",
	Long: `A CLI for managing tickets stored as markdown in the current directory.

Tickets are stored in a TODO.md file. Each ticket has a title, a 3-character
ID, and an optional description.

Descriptions can be passed via stdin to support multi-line content with backticks,
code blocks, and any special characters without shell escaping issues.`,
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
