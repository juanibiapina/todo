package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var setStateCmd = &cobra.Command{
	Use:   "set-state <title|id> <state>",
	Short: "Change a ticket's state",
	Long: `Change the state of a ticket.

Valid states: new, refined, planned

The ticket can be referenced by its 3-character ID or its title.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]
		state := tickets.State(args[1])

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		title, err := tickets.SetState(dir, ref, state)
		if err != nil {
			return err
		}

		fmt.Printf("Set state to %s: %s\n", state, title)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setStateCmd)
}
