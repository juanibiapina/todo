package tickets

import "fmt"

// State represents a ticket's workflow state.
type State string

const (
	StateNew     State = "new"
	StateRefined State = "refined"
)

// ValidStates contains all valid ticket states.
var ValidStates = []State{StateNew, StateRefined}

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

// NextState returns the next state in the workflow.
func NextState(s State) State {
	switch s {
	case StateNew:
		return StateRefined
	default:
		return s
	}
}

// PrevState returns the previous state in the workflow.
func PrevState(s State) State {
	switch s {
	case StateRefined:
		return StateNew
	default:
		return s
	}
}

// StateIcon returns a nerd font icon for a state.
func StateIcon(s State) string {
	switch s {
	case StateNew:
		return "\uf10c" // nf-fa-circle_o
	case StateRefined:
		return "\uf111" // nf-fa-circle
	default:
		return "\uf10c"
	}
}
