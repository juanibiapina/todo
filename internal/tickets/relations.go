package tickets

import (
	"fmt"
	"strings"
)

// TicketRelations holds computed relationships for a ticket.
type TicketRelations struct {
	// ParentTicket is the resolved parent ticket (nil if no parent or not found).
	ParentTicket *Ticket
	// Blockers are unclosed tickets this ticket depends on.
	Blockers []*Ticket
	// Blocking are tickets that depend on this ticket (and this ticket is unclosed).
	Blocking []*Ticket
	// Children are tickets whose parent is this ticket.
	Children []*Ticket
	// Linked are resolved tickets from this ticket's links.
	Linked []*Ticket
}

// ComputeRelations computes all relationships for a given ticket
// using the full list of all tickets.
func ComputeRelations(ticket *Ticket, allTickets []*Ticket) *TicketRelations {
	// Build ID -> Ticket map
	ticketMap := make(map[string]*Ticket)
	for _, t := range allTickets {
		ticketMap[t.ID] = t
	}

	rel := &TicketRelations{}

	// Resolve parent with title
	if ticket.Parent != "" {
		if parent, ok := ticketMap[ticket.Parent]; ok {
			rel.ParentTicket = parent
		}
	}

	// Blockers: unclosed deps of this ticket
	for _, depID := range ticket.Deps {
		if dep, ok := ticketMap[depID]; ok {
			if dep.Status != "closed" {
				rel.Blockers = append(rel.Blockers, dep)
			}
		}
	}

	// Blocking: tickets that have this ticket as a dep, where this ticket is unclosed
	if ticket.Status != "closed" {
		for _, t := range allTickets {
			if t.ID == ticket.ID {
				continue
			}
			for _, depID := range t.Deps {
				if depID == ticket.ID {
					rel.Blocking = append(rel.Blocking, t)
					break
				}
			}
		}
	}

	// Children: tickets whose parent is this ticket
	for _, t := range allTickets {
		if t.Parent == ticket.ID {
			rel.Children = append(rel.Children, t)
		}
	}

	// Linked: resolve link IDs to tickets
	for _, linkID := range ticket.Links {
		if linked, ok := ticketMap[linkID]; ok {
			rel.Linked = append(rel.Linked, linked)
		}
	}

	return rel
}

// FormatRelations returns a formatted string of relationship sections.
// Only non-empty sections are included. Returns empty string if no relations.
func FormatRelations(rel *TicketRelations) string {
	var b strings.Builder

	if rel.ParentTicket != nil {
		// This is handled by enhancing the parent line in FullString output,
		// so we don't add a separate section here.
	}

	if len(rel.Blockers) > 0 {
		b.WriteString("\n## Blockers\n")
		for _, t := range rel.Blockers {
			b.WriteString(formatRelationLine(t))
			b.WriteString("\n")
		}
	}

	if len(rel.Blocking) > 0 {
		b.WriteString("\n## Blocking\n")
		for _, t := range rel.Blocking {
			b.WriteString(formatRelationLine(t))
			b.WriteString("\n")
		}
	}

	if len(rel.Children) > 0 {
		b.WriteString("\n## Children\n")
		for _, t := range rel.Children {
			b.WriteString(formatRelationLine(t))
			b.WriteString("\n")
		}
	}

	if len(rel.Linked) > 0 {
		b.WriteString("\n## Linked\n")
		for _, t := range rel.Linked {
			b.WriteString(formatRelationLine(t))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// formatRelationLine formats a single ticket reference for a relation section.
// Format: "- id [status] Title" or "- id Title" when status is empty.
func formatRelationLine(t *Ticket) string {
	if t.Status != "" {
		return fmt.Sprintf("- %s [%s] %s", t.ID, t.Status, t.Title)
	}
	return fmt.Sprintf("- %s %s", t.ID, t.Title)
}

// FormatParentLine returns an enhanced parent line with title.
// Format: "parent: id (Title)" or empty string if no parent ticket.
func FormatParentLine(rel *TicketRelations) string {
	if rel.ParentTicket == nil {
		return ""
	}
	return fmt.Sprintf("%s (%s)", rel.ParentTicket.ID, rel.ParentTicket.Title)
}
