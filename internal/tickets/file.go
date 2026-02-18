package tickets

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultDirName = "docs/tickets"

// DirPath returns the path to the tickets directory in the given directory.
func DirPath(dir string) string {
	return filepath.Join(dir, defaultDirName)
}

// ticketFileName generates the filename for a ticket.
func ticketFileName(id string) string {
	return fmt.Sprintf("%s.md", id)
}

// ticketFilePath returns the full path for a ticket file.
func ticketFilePath(dir, id string) string {
	return filepath.Join(DirPath(dir), ticketFileName(id))
}

// findTicketFile finds a ticket file by ID using exact match.
func findTicketFile(dir, id string) (string, error) {
	path := ticketFilePath(dir, id)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("ticket not found: %s", id)
		}
		return "", err
	}
	return path, nil
}

// parseFile reads a single ticket file and returns the ticket.
func parseFile(path string) (*Ticket, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)

	// Expect YAML frontmatter opening delimiter
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("missing frontmatter opening delimiter")
	}

	// Find the closing --- delimiter
	rest := content[4:] // skip opening "---\n"
	closingIdx := strings.Index(rest, "\n---\n")
	if closingIdx < 0 {
		return nil, fmt.Errorf("missing frontmatter closing delimiter")
	}

	yamlContent := rest[:closingIdx]
	afterFrontmatter := rest[closingIdx+5:] // skip "\n---\n"

	// Parse YAML frontmatter
	var fm frontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, fmt.Errorf("invalid frontmatter YAML: %w", err)
	}

	// Parse title line (expect "# Title\n")
	var title string
	var description string

	if strings.HasPrefix(afterFrontmatter, "# ") {
		nlIdx := strings.Index(afterFrontmatter, "\n")
		if nlIdx >= 0 {
			title = afterFrontmatter[2:nlIdx]
			descPart := afterFrontmatter[nlIdx+1:]
			description = strings.TrimRight(descPart, "\n")
		} else {
			title = afterFrontmatter[2:]
		}
	}

	ticket := &Ticket{
		Title:       title,
		ID:          fm.ID,
		Description: description,
		Status:      fm.Status,
		Type:        fm.Type,
		Priority:    fm.Priority,
		Assignee:    fm.Assignee,
		Created:     fm.Created,
		Parent:      fm.Parent,
		ExternalRef: fm.ExternalRef,
		Deps:        fm.Deps,
		Links:       fm.Links,
		Tags:        fm.Tags,
	}

	return ticket, nil
}

// writeFile writes a ticket to its file.
func writeFile(dir string, t *Ticket) error {
	path := ticketFilePath(dir, t.ID)
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

	t.Description = description

	// Overwrite in place — filename doesn't depend on title
	if err := os.WriteFile(path, []byte(t.FullString()), 0644); err != nil {
		return "", err
	}

	return t.Title, nil
}
