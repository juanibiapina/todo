package tickets

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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

// findTicketFile finds a ticket file by ID.
// It tries an exact match first, then falls back to partial (substring) matching.
// Returns an error if zero or multiple tickets match a partial ID.
func findTicketFile(dir, id string) (string, error) {
	// Try exact match first
	path := ticketFilePath(dir, id)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	// Fall back to partial matching: scan all .md files
	ticketsDir := DirPath(dir)
	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("ticket not found: %s", id)
		}
		return "", err
	}

	var matches []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		ticketID := strings.TrimSuffix(entry.Name(), ".md")
		if strings.Contains(ticketID, id) {
			matches = append(matches, filepath.Join(ticketsDir, entry.Name()))
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("ticket not found: %s", id)
	case 1:
		return matches[0], nil
	default:
		// Extract IDs from matched paths for the error message
		var ids []string
		for _, m := range matches {
			ids = append(ids, strings.TrimSuffix(filepath.Base(m), ".md"))
		}
		sort.Strings(ids)
		return "", fmt.Errorf("ambiguous ticket ID %q: matches %s", id, strings.Join(ids, ", "))
	}
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
		Design:      fm.Design,
		Acceptance:  fm.Acceptance,
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
// The caller provides a pre-populated Ticket; Add generates the ID and writes to disk.
func Add(dir string, t *Ticket) (*Ticket, error) {
	if err := EnsureDir(dir); err != nil {
		return nil, err
	}

	// Validate parent exists if set
	if t.Parent != "" {
		if _, err := findTicketFile(dir, t.Parent); err != nil {
			return nil, fmt.Errorf("parent ticket not found: %s", t.Parent)
		}
	}

	// Get existing IDs to avoid collision
	tickets, err := List(dir)
	if err != nil {
		return nil, err
	}

	t.ID = generateUniqueID(existingIDs(tickets))

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

// Done marks a ticket as closed by setting its status.
func Done(dir string, id string) (string, error) {
	path, err := findTicketFile(dir, id)
	if err != nil {
		return "", err
	}

	t, err := parseFile(path)
	if err != nil {
		return "", err
	}

	t.Status = "closed"

	if err := writeFile(dir, t); err != nil {
		return "", err
	}

	return t.Title, nil
}

// validStatuses lists the allowed ticket statuses.
var validStatuses = map[string]bool{
	"open":        true,
	"in_progress": true,
	"closed":      true,
}

// SetStatus changes a ticket's status after validation.
func SetStatus(dir string, id string, status string) (string, error) {
	if !validStatuses[status] {
		return "", fmt.Errorf("invalid status: %q (valid: open, in_progress, closed)", status)
	}

	path, err := findTicketFile(dir, id)
	if err != nil {
		return "", err
	}

	t, err := parseFile(path)
	if err != nil {
		return "", err
	}

	t.Status = status

	if err := writeFile(dir, t); err != nil {
		return "", err
	}

	return t.Title, nil
}

// AddDep adds a dependency from one ticket to another.
// Both tickets must exist. The operation is idempotent.
func AddDep(dir string, id string, depID string) error {
	// Resolve and validate the ticket
	path, err := findTicketFile(dir, id)
	if err != nil {
		return err
	}

	// Resolve and validate the dependency ticket
	depPath, err := findTicketFile(dir, depID)
	if err != nil {
		return err
	}

	t, err := parseFile(path)
	if err != nil {
		return err
	}

	// Extract the resolved (full) dep ID from the path
	resolvedDepID := strings.TrimSuffix(filepath.Base(depPath), ".md")

	// Check if already present (idempotent)
	for _, d := range t.Deps {
		if d == resolvedDepID {
			return nil
		}
	}

	t.Deps = append(t.Deps, resolvedDepID)

	return writeFile(dir, t)
}

// RemoveDep removes a dependency from a ticket.
// Both tickets must exist. The operation is idempotent.
func RemoveDep(dir string, id string, depID string) error {
	// Resolve and validate the ticket
	path, err := findTicketFile(dir, id)
	if err != nil {
		return err
	}

	// Resolve and validate the dependency ticket
	depPath, err := findTicketFile(dir, depID)
	if err != nil {
		return err
	}

	t, err := parseFile(path)
	if err != nil {
		return err
	}

	// Extract the resolved (full) dep ID from the path
	resolvedDepID := strings.TrimSuffix(filepath.Base(depPath), ".md")

	// Remove if present (idempotent — no error if not found)
	var newDeps []string
	for _, d := range t.Deps {
		if d != resolvedDepID {
			newDeps = append(newDeps, d)
		}
	}
	t.Deps = newDeps

	return writeFile(dir, t)
}

// AddLink creates bidirectional links between all provided ticket IDs.
// All tickets must exist. The operation is idempotent.
// For 3+ IDs, all pairs are linked.
func AddLink(dir string, ids []string) error {
	// Resolve all IDs and load tickets
	type resolved struct {
		id   string
		path string
	}
	var tickets []resolved

	for _, id := range ids {
		path, err := findTicketFile(dir, id)
		if err != nil {
			return err
		}
		fullID := strings.TrimSuffix(filepath.Base(path), ".md")
		tickets = append(tickets, resolved{id: fullID, path: path})
	}

	// For each ticket, add all other tickets as links
	for i, ticket := range tickets {
		t, err := parseFile(ticket.path)
		if err != nil {
			return err
		}

		modified := false
		for j, other := range tickets {
			if i == j {
				continue
			}
			// Check if already present (idempotent)
			found := false
			for _, l := range t.Links {
				if l == other.id {
					found = true
					break
				}
			}
			if !found {
				t.Links = append(t.Links, other.id)
				modified = true
			}
		}

		if modified {
			if err := writeFile(dir, t); err != nil {
				return err
			}
		}
	}

	return nil
}

// RemoveLink removes a bidirectional link between two tickets.
// Both tickets must exist. The operation is idempotent.
func RemoveLink(dir string, id string, targetID string) error {
	// Resolve both IDs
	path, err := findTicketFile(dir, id)
	if err != nil {
		return err
	}
	targetPath, err := findTicketFile(dir, targetID)
	if err != nil {
		return err
	}

	resolvedID := strings.TrimSuffix(filepath.Base(path), ".md")
	resolvedTargetID := strings.TrimSuffix(filepath.Base(targetPath), ".md")

	// Remove targetID from id's links
	t, err := parseFile(path)
	if err != nil {
		return err
	}
	var newLinks []string
	for _, l := range t.Links {
		if l != resolvedTargetID {
			newLinks = append(newLinks, l)
		}
	}
	t.Links = newLinks
	if err := writeFile(dir, t); err != nil {
		return err
	}

	// Remove id from targetID's links
	t2, err := parseFile(targetPath)
	if err != nil {
		return err
	}
	var newLinks2 []string
	for _, l := range t2.Links {
		if l != resolvedID {
			newLinks2 = append(newLinks2, l)
		}
	}
	t2.Links = newLinks2
	if err := writeFile(dir, t2); err != nil {
		return err
	}

	return nil
}

// AddNote appends a timestamped note to a ticket's description under a ## Notes section.
func AddNote(dir string, id string, text string) (string, error) {
	path, err := findTicketFile(dir, id)
	if err != nil {
		return "", err
	}

	t, err := parseFile(path)
	if err != nil {
		return "", err
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04 UTC")
	noteEntry := fmt.Sprintf("**%s**\n\n%s", timestamp, text)

	if strings.Contains(t.Description, "## Notes") {
		// Append to existing Notes section
		t.Description = t.Description + "\n\n" + noteEntry
	} else {
		// Add Notes section
		if t.Description != "" {
			t.Description = t.Description + "\n\n## Notes\n\n" + noteEntry
		} else {
			t.Description = "## Notes\n\n" + noteEntry
		}
	}

	if err := os.WriteFile(path, []byte(t.FullString()), 0644); err != nil {
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
