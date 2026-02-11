package tickets

import "fmt"

// Ticket represents a single ticket.
type Ticket struct {
	Title       string
	ID          string
	Description string
}

// String returns a formatted single-line representation: "ID Title"
func (t *Ticket) String() string {
	return fmt.Sprintf("%s %s", t.ID, t.Title)
}

// FullString returns the full markdown representation of a ticket.
func (t *Ticket) FullString() string {
	s := fmt.Sprintf("## %s\n---\nid: %s\n---\n", t.Title, t.ID)
	if t.Description != "" {
		s += t.Description + "\n"
	}
	return s
}
