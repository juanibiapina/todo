package cmd

import (
	"fmt"
	"os"

	"github.com/juanibiapina/todo/internal/tickets"
	"github.com/spf13/cobra"
)

var depCycleCmd = &cobra.Command{
	Use:   "cycle",
	Short: "Detect dependency cycles among open tickets",
	Long:  `Detect dependency cycles using DFS-based analysis on open (non-closed) tickets. Outputs normalized cycles with member details.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		output, err := tickets.DepCycles(dir)
		if err != nil {
			return err
		}

		if output != "" {
			fmt.Println(output)
		}

		return nil
	},
}

func init() {
	depCmd.AddCommand(depCycleCmd)
}
