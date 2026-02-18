package cmd

import (
	"fmt"
	"strings"

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
	var b strings.Builder

	b.WriteString(cliID(t.ID))

	if t.Status != "" {
		b.WriteString(fmt.Sprintf(" [%s]", t.Status))
	}

	b.WriteString(" - ")
	b.WriteString(t.Title)

	if len(t.Deps) > 0 {
		b.WriteString(" <- [")
		b.WriteString(strings.Join(t.Deps, ", "))
		b.WriteString("]")
	}

	return b.String()
}
