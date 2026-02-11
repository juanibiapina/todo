package tickets

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const defaultFileName = "TODO.md"

// FilePath returns the path to the tickets file in the given directory.
func FilePath(dir string) string {
	return dir + "/" + defaultFileName
}

// parse reads a tickets file and returns the header line and all tickets.
func parse(path string) (string, []*Ticket, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, nil
		}
		return "", nil, err
	}

	var tickets []*Ticket
	var current *Ticket
	var header string
	inFrontMatter := false
	fmCount := 0
	pastFM := false
	var descLines []string

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "## ") {
			// Finish previous ticket
			if current != nil {
				if pastFM && len(descLines) > 0 {
					current.Description = strings.TrimRight(strings.Join(descLines, "\n"), "\n")
				}
				tickets = append(tickets, current)
			}

			current = &Ticket{
				Title: strings.TrimPrefix(line, "## "),
			}
			inFrontMatter = false
			fmCount = 0
			pastFM = false
			descLines = nil
			continue
		}

		if strings.HasPrefix(line, "# ") && current == nil {
			header = line
			continue
		}

		if current == nil {
			continue
		}

		if !pastFM && line == "---" {
			fmCount++
			if fmCount == 1 {
				inFrontMatter = true
			} else if fmCount == 2 {
				inFrontMatter = false
				pastFM = true
			}
			continue
		}

		if inFrontMatter {
			if strings.HasPrefix(line, "id: ") {
				current.ID = strings.TrimPrefix(line, "id: ")
			}
			// Ignore state: lines for backwards compatibility
			continue
		}

		if pastFM {
			descLines = append(descLines, line)
		}
	}

	// Finish last ticket
	if current != nil {
		if pastFM && len(descLines) > 0 {
			current.Description = strings.TrimRight(strings.Join(descLines, "\n"), "\n")
		}
		tickets = append(tickets, current)
	}

	return header, tickets, scanner.Err()
}

// write serializes the header and tickets back to the file.
func write(path string, header string, tickets []*Ticket) error {
	var b strings.Builder

	if header != "" {
		b.WriteString(header)
		b.WriteString("\n")
	}

	for _, t := range tickets {
		b.WriteString("\n")
		b.WriteString(t.FullString())
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}

// EnsureFile creates the tickets file with a header if it doesn't exist.
func EnsureFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, []byte("# TODO\n"), 0644)
}

// existingIDs collects all IDs from a ticket list.
func existingIDs(tickets []*Ticket) map[string]bool {
	ids := make(map[string]bool)
	for _, t := range tickets {
		if t.ID != "" {
			ids[t.ID] = true
		}
	}
	return ids
}

// findTicket resolves a ticket ID to a ticket.
func findTicket(tickets []*Ticket, id string) *Ticket {
	for _, t := range tickets {
		if t.ID == id {
			return t
		}
	}
	return nil
}

// List returns all tickets from the file in the given directory.
func List(dir string) ([]*Ticket, error) {
	path := FilePath(dir)
	_, tickets, err := parse(path)
	return tickets, err
}

// Add creates a new ticket and returns it.
func Add(dir string, title string, description string) (*Ticket, error) {
	path := FilePath(dir)
	if err := EnsureFile(path); err != nil {
		return nil, err
	}

	header, tickets, err := parse(path)
	if err != nil {
		return nil, err
	}

	t := &Ticket{
		Title:       title,
		ID:          generateUniqueID(existingIDs(tickets)),
		Description: description,
	}

	tickets = append(tickets, t)

	if err := write(path, header, tickets); err != nil {
		return nil, err
	}

	return t, nil
}

// Show returns a ticket by reference (ID or title).
func Show(dir string, ref string) (*Ticket, error) {
	path := FilePath(dir)
	_, tickets, err := parse(path)
	if err != nil {
		return nil, err
	}

	t := findTicket(tickets, ref)
	if t == nil {
		return nil, fmt.Errorf("ticket not found: %s", ref)
	}

	return t, nil
}

// Done removes a ticket by reference.
func Done(dir string, ref string) (string, error) {
	path := FilePath(dir)
	header, tickets, err := parse(path)
	if err != nil {
		return "", err
	}

	var remaining []*Ticket
	var title string
	for _, t := range tickets {
		if t.ID == ref {
			title = t.Title
			continue
		}
		remaining = append(remaining, t)
	}

	if title == "" {
		return "", fmt.Errorf("ticket not found: %s", ref)
	}

	if err := write(path, header, remaining); err != nil {
		return "", err
	}

	return title, nil
}

// findTicketIndex resolves a ticket ID to an index in the ticket slice.
func findTicketIndex(tickets []*Ticket, id string) int {
	for i, t := range tickets {
		if t.ID == id {
			return i
		}
	}
	return -1
}

// MoveUp swaps a ticket with the one above it.
func MoveUp(dir string, ref string) (string, error) {
	path := FilePath(dir)
	header, tickets, err := parse(path)
	if err != nil {
		return "", err
	}

	idx := findTicketIndex(tickets, ref)
	if idx < 0 {
		return "", fmt.Errorf("ticket not found: %s", ref)
	}
	if idx == 0 {
		return tickets[idx].Title, nil // already at top
	}

	tickets[idx], tickets[idx-1] = tickets[idx-1], tickets[idx]

	if err := write(path, header, tickets); err != nil {
		return "", err
	}

	return tickets[idx-1].Title, nil
}

// MoveDown swaps a ticket with the one below it.
func MoveDown(dir string, ref string) (string, error) {
	path := FilePath(dir)
	header, tickets, err := parse(path)
	if err != nil {
		return "", err
	}

	idx := findTicketIndex(tickets, ref)
	if idx < 0 {
		return "", fmt.Errorf("ticket not found: %s", ref)
	}
	if idx == len(tickets)-1 {
		return tickets[idx].Title, nil // already at bottom
	}

	tickets[idx], tickets[idx+1] = tickets[idx+1], tickets[idx]

	if err := write(path, header, tickets); err != nil {
		return "", err
	}

	return tickets[idx+1].Title, nil
}

// SetDescription sets or replaces a ticket's description.
func SetDescription(dir string, ref string, description string) (string, error) {
	path := FilePath(dir)
	header, tickets, err := parse(path)
	if err != nil {
		return "", err
	}

	t := findTicket(tickets, ref)
	if t == nil {
		return "", fmt.Errorf("ticket not found: %s", ref)
	}

	t.Description = description

	if err := write(path, header, tickets); err != nil {
		return "", err
	}

	return t.Title, nil
}
