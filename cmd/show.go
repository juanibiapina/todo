package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a ticket's full details",
	Long:  `Show the full details of a ticket, including its front matter and description.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		ticket, err := tickets.Show(dir, ref)
		if err != nil {
			return err
		}

		// Load all tickets for computing relations
		allTickets, err := tickets.List(dir)
		if err != nil {
			return err
		}

		rel := tickets.ComputeRelations(ticket, allTickets)

		// Build output
		var b strings.Builder

		// If parent has a resolved title, enhance the frontmatter parent line
		output := ticket.FullString()
		if parentLine := tickets.FormatParentLine(rel); parentLine != "" {
			output = strings.Replace(output, "parent: "+ticket.Parent, "parent: "+parentLine, 1)
		}
		b.WriteString(output)

		// Append computed relation sections
		b.WriteString(tickets.FormatRelations(rel))

		result := b.String()

		// Pipe through pager if TODO_PAGER is set and stdout is a TTY
		pager := os.Getenv("TODO_PAGER")
		if pager != "" && term.IsTerminal(int(os.Stdout.Fd())) {
			return runPager(pager, result)
		}

		fmt.Print(result)
		return nil
	},
}

// runPager pipes content through the specified pager command.
func runPager(pager string, content string) error {
	pagerCmd := exec.Command("sh", "-c", pager)
	pagerCmd.Stdout = os.Stdout
	pagerCmd.Stderr = os.Stderr

	stdin, err := pagerCmd.StdinPipe()
	if err != nil {
		// Fall back to direct output
		fmt.Print(content)
		return nil
	}

	if err := pagerCmd.Start(); err != nil {
		// Fall back to direct output
		fmt.Print(content)
		return nil
	}

	io.WriteString(stdin, content)
	stdin.Close()

	return pagerCmd.Wait()
}

func init() {
	rootCmd.AddCommand(showCmd)
}
