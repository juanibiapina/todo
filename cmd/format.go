package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/juanibiapina/todo/internal/tickets"
)

// CLI colors matching TUI style (ANSI 0-15)
var (
	cliMagenta = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
)

func cliID(id string) string {
	return cliMagenta.Render(id)
}

func formatTicketLine(t *tickets.Ticket) string {
	return fmt.Sprintf("%s %s", cliID(t.ID), t.Title)
}
