package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/juanibiapina/todo/internal/config"
	"github.com/juanibiapina/todo/internal/render"
	"github.com/juanibiapina/todo/internal/store"
	"github.com/spf13/cobra"
)

var listAll bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List todo items",
	Long:  `List all todo items for the current directory, or all directories with --all.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.Open(config.DefaultFilePath())
		if err != nil {
			return err
		}
		defer s.Close()

		if listAll {
			items, err := s.ListAll()
			if err != nil {
				return err
			}

			currentDir := ""
			for _, item := range items {
				if item.Directory != currentDir {
					if currentDir != "" {
						fmt.Println()
					}
					fmt.Printf("%s\n", item.Directory)
					currentDir = item.Directory
				}
				if item.IsSection {
					fmt.Println("  ──────────")
					continue
				}
				check := " "
				if item.Checked {
					check = "x"
				}
				active := ""
				if item.IsActive {
					active = "  (active)"
				}
				fmt.Printf("  %d [%s] %s%s\n", item.ID, check, item.Text, active)
			}
			return nil
		}

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		if err := s.CleanIfNewDay(dir); err != nil {
			return err
		}

		items, err := s.List(dir)
		if err != nil {
			return err
		}

		for _, item := range items {
			if item.IsSection {
				fmt.Println(render.SectionLine(item.Text, 30))
				continue
			}
			check := " "
			if item.Checked {
				check = "x"
			}
			active := ""
			if item.IsActive {
				active = "  (active)"
			}
			indent := strings.Repeat("  ", item.IndentLevel)
			fmt.Printf("%d %s[%s] %s%s\n", item.ID, indent, check, item.Text, active)
		}

		return nil
	},
}

func init() {
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "List items from all directories")
	rootCmd.AddCommand(listCmd)
}
