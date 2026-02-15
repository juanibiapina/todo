package tickets

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const defaultDirName = "docs/tickets"

// DirPath returns the path to the tickets directory in the given directory.
func DirPath(dir string) string {
	return filepath.Join(dir, defaultDirName)
}

// slugify converts a title to a URL-safe slug.
func slugify(title string) string {
	// Lowercase
	s := strings.ToLower(title)

	// Replace non-alphanumeric characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")

	// Trim leading/trailing hyphens
	s = strings.Trim(s, "-")

	// Truncate to reasonable length
	if len(s) > 50 {
		s = s[:50]
		// Don't end with a hyphen
		s = strings.TrimRight(s, "-")
	}

	return s
}

// ticketFileName generates the filename for a ticket.
func ticketFileName(id, title string) string {
	return fmt.Sprintf("%s-%s.md", id, slugify(title))
}

// ticketFilePath returns the full path for a ticket file.
func ticketFilePath(dir, id, title string) string {
	return filepath.Join(DirPath(dir), ticketFileName(id, title))
}

// findTicketFile finds a ticket file by ID using glob pattern.
func findTicketFile(dir, id string) (string, error) {
	pattern := filepath.Join(DirPath(dir), id+"-*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("ticket not found: %s", id)
	}
	return matches[0], nil
}

// parseFile reads a single ticket file and returns the ticket.
func parseFile(path string) (*Ticket, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ticket *Ticket
	inFrontMatter := false
	fmCount := 0
	pastFM := false
	var descLines []string

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()

		// Title line
		if strings.HasPrefix(line, "# ") && ticket == nil {
			ticket = &Ticket{
				Title: strings.TrimPrefix(line, "# "),
			}
			continue
		}

		if ticket == nil {
			continue
		}

		// Front matter delimiters
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

		// Front matter content
		if inFrontMatter {
			if strings.HasPrefix(line, "id: ") {
				ticket.ID = strings.TrimPrefix(line, "id: ")
			}
			continue
		}

		// Description content
		if pastFM {
			descLines = append(descLines, line)
		}
	}

	if ticket != nil && pastFM && len(descLines) > 0 {
		ticket.Description = strings.TrimRight(strings.Join(descLines, "\n"), "\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ticket, nil
}

// writeFile writes a ticket to its file.
func writeFile(dir string, t *Ticket) error {
	path := ticketFilePath(dir, t.ID, t.Title)
	return os.WriteFile(path, []byte(t.FullString()), 0644)
}

// EnsureDir creates the tickets directory if it doesn't exist.
func EnsureDir(dir string) error {
	ticketsDir := DirPath(dir)
	return os.MkdirAll(ticketsDir, 0755)
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

// List returns all tickets from the tickets directory.
func List(dir string) ([]*Ticket, error) {
	ticketsDir := DirPath(dir)

	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var tickets []*Ticket
	var filenames []string

	// Collect filenames for sorting
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		filenames = append(filenames, entry.Name())
	}

	// Sort alphabetically for consistent ordering
	sort.Strings(filenames)

	// Parse each file
	for _, filename := range filenames {
		path := filepath.Join(ticketsDir, filename)
		t, err := parseFile(path)
		if err != nil {
			continue // Skip files that can't be parsed
		}
		if t != nil {
			tickets = append(tickets, t)
		}
	}

	return tickets, nil
}

// Add creates a new ticket and returns it.
func Add(dir string, title string, description string) (*Ticket, error) {
	if err := EnsureDir(dir); err != nil {
		return nil, err
	}

	// Get existing IDs to avoid collision
	tickets, err := List(dir)
	if err != nil {
		return nil, err
	}

	t := &Ticket{
		Title:       title,
		ID:          generateUniqueID(existingIDs(tickets)),
		Description: description,
	}

	if err := writeFile(dir, t); err != nil {
		return nil, err
	}

	return t, nil
}

// Show returns a ticket by ID.
func Show(dir string, id string) (*Ticket, error) {
	path, err := findTicketFile(dir, id)
	if err != nil {
		return nil, err
	}

	return parseFile(path)
}

// Done removes a ticket by ID.
func Done(dir string, id string) (string, error) {
	path, err := findTicketFile(dir, id)
	if err != nil {
		return "", err
	}

	t, err := parseFile(path)
	if err != nil {
		return "", err
	}

	if err := os.Remove(path); err != nil {
		return "", err
	}

	return t.Title, nil
}

// SetDescription sets or replaces a ticket's description.
func SetDescription(dir string, id string, description string) (string, error) {
	path, err := findTicketFile(dir, id)
	if err != nil {
		return "", err
	}

	t, err := parseFile(path)
	if err != nil {
		return "", err
	}

	// Remove old file (in case title changed affecting filename)
	if err := os.Remove(path); err != nil {
		return "", err
	}

	t.Description = description

	if err := writeFile(dir, t); err != nil {
		return "", err
	}

	return t.Title, nil
}
