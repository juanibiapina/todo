package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/juanibiapina/todo/internal/tickets"
)

// CLI colors matching TUI style (ANSI 0-15)
var (
	cliMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	cliCyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	cliMagenta = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
)

func cliStateIcon(s tickets.State) string {
	icon := tickets.StateIcon(s)
	switch s {
	case tickets.StateNew:
		return cliMuted.Render(icon)
	case tickets.StateRefined:
		return cliCyan.Render(icon)
	default:
		return cliMuted.Render(icon)
	}
}

func cliID(id string) string {
	return cliMagenta.Render(id)
}

func formatTicketLine(t *tickets.Ticket) string {
	return fmt.Sprintf("%s %s %s", cliStateIcon(t.State), cliID(t.ID), t.Title)
}
