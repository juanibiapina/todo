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

func formatReadyLine(t *tickets.Ticket) string {
	var b strings.Builder

	b.WriteString(cliID(t.ID))
	b.WriteString(fmt.Sprintf(" [P%d]", t.Priority))

	status := t.Status
	if status == "" {
		status = "open"
	}
	b.WriteString(fmt.Sprintf("[%s]", status))

	b.WriteString(" - ")
	b.WriteString(t.Title)

	return b.String()
}

func formatBlockedLine(t *tickets.Ticket, unclosedBlockers []string) string {
	var b strings.Builder

	b.WriteString(cliID(t.ID))
	b.WriteString(fmt.Sprintf(" [P%d]", t.Priority))

	status := t.Status
	if status == "" {
		status = "open"
	}
	b.WriteString(fmt.Sprintf("[%s]", status))

	b.WriteString(" - ")
	b.WriteString(t.Title)

	if len(unclosedBlockers) > 0 {
		b.WriteString(" <- [")
		b.WriteString(strings.Join(unclosedBlockers, ", "))
		b.WriteString("]")
	}

	return b.String()
}

func formatClosedLine(t *tickets.Ticket) string {
	var b strings.Builder

	b.WriteString(cliID(t.ID))
	b.WriteString(" - ")
	b.WriteString(t.Title)

	return b.String()
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
