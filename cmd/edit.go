package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Open a ticket in your editor",
	Long:  `Open a ticket file in $EDITOR (default vi). If stdout is not a TTY, prints the file path instead.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		// Resolve the ticket (validates existence and handles partial IDs)
		ticket, err := tickets.Show(dir, ref)
		if err != nil {
			return err
		}

		ticketPath := filepath.Join(tickets.DirPath(dir), ticket.ID+".md")

		// If stdout is not a TTY, print the file path
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			fmt.Println(ticketPath)
			return nil
		}

		// Launch editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		editorCmd := exec.Command(editor, ticketPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		return editorCmd.Run()
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
