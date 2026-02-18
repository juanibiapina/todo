package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var addNoteCmd = &cobra.Command{
	Use:   "add-note <id> [text]",
	Short: "Append a timestamped note to a ticket",
	Long: `Append a timestamped note to a ticket's description under a ## Notes section.

Note text can be provided as:
  1. A positional argument (for simple text)
  2. Via stdin (for multi-line content)

Each note is prefixed with a UTC timestamp. The ## Notes header is added
automatically on the first note and reused for subsequent notes.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]
		var text string

		if len(args) > 1 {
			text = args[1]
		} else {
			// Check if stdin has data (not a terminal)
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}
				text = strings.TrimRight(string(data), "\n")
			}
		}

		if text == "" {
			return fmt.Errorf("no note text provided (pass as argument or via stdin)")
		}

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		title, err := tickets.AddNote(dir, ref, text)
		if err != nil {
			return err
		}

		fmt.Printf("Added note to: %s\n", title)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addNoteCmd)
}
