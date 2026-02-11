package tickets

import "fmt"

// State represents a ticket's workflow state.
type State string

const (
	StateNew     State = "new"
	StateRefined State = "refined"
	StatePlanned State = "planned"
)

// ValidStates contains all valid ticket states.
var ValidStates = []State{StateNew, StateRefined, StatePlanned}

// IsValid checks if a state is valid.
func (s State) IsValid() bool {
	for _, v := range ValidStates {
		if s == v {
			return true
		}
	}
	return false
}

// Ticket represents a single ticket.
type Ticket struct {
	Title       string
	ID          string
	State       State
	Description string
}

// String returns a formatted single-line representation: "ID [state] Title"
func (t *Ticket) String() string {
	return fmt.Sprintf("%s [%-8s] %s", t.ID, t.State, t.Title)
}

// FullString returns the full markdown representation of a ticket.
func (t *Ticket) FullString() string {
	s := fmt.Sprintf("## %s\n---\nid: %s\nstate: %s\n---\n", t.Title, t.ID, t.State)
	if t.Description != "" {
		s += t.Description + "\n"
	}
	return s
}
