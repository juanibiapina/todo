package tickets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

func TestAddAndList(t *testing.T) {
	dir := tempDir(t)

	ticket, err := Add(dir, "Fix login bug", "")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if ticket.Title != "Fix login bug" {
		t.Errorf("title = %q, want %q", ticket.Title, "Fix login bug")
	}
	if len(ticket.ID) != 3 {
		t.Errorf("id length = %d, want 3", len(ticket.ID))
	}

	tickets, err := List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tickets) != 1 {
		t.Fatalf("len = %d, want 1", len(tickets))
	}
	if tickets[0].Title != "Fix login bug" {
		t.Errorf("title = %q", tickets[0].Title)
	}
}

func TestAddWithDescription(t *testing.T) {
	dir := tempDir(t)

	ticket, err := Add(dir, "Refactor auth", "Move auth to middleware layer.")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if ticket.Description != "Move auth to middleware layer." {
		t.Errorf("description = %q", ticket.Description)
	}

	// Round-trip
	loaded, err := Show(dir, ticket.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if loaded.Description != "Move auth to middleware layer." {
		t.Errorf("loaded description = %q", loaded.Description)
	}
}

func TestAddWithMultilineBacktickDescription(t *testing.T) {
	dir := tempDir(t)

	desc := "Fix the `handleAuth` function.\n\n```go\nfunc handleAuth() {\n  // TODO\n}\n```\n\nThis needs $variables and `backticks`."
	// Use actual newlines
	desc = strings.ReplaceAll(desc, "\\n", "\n")

	ticket, err := Add(dir, "Backtick test", desc)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	loaded, err := Show(dir, ticket.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if loaded.Description != desc {
		t.Errorf("description mismatch.\ngot:  %q\nwant: %q", loaded.Description, desc)
	}
}

func TestShowByID(t *testing.T) {
	dir := tempDir(t)

	ticket, err := Add(dir, "Test ticket", "")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	byID, err := Show(dir, ticket.ID)
	if err != nil {
		t.Fatalf("Show by ID: %v", err)
	}
	if byID.Title != "Test ticket" {
		t.Errorf("byID title = %q", byID.Title)
	}
}

func TestDone(t *testing.T) {
	dir := tempDir(t)

	ticket, err := Add(dir, "To remove", "")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	title, err := Done(dir, ticket.ID)
	if err != nil {
		t.Fatalf("Done: %v", err)
	}
	if title != "To remove" {
		t.Errorf("title = %q", title)
	}

	tickets, err := List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tickets) != 0 {
		t.Errorf("len = %d, want 0", len(tickets))
	}

	// Verify file is actually deleted
	path := filepath.Join(DirPath(dir), ticket.ID+".md")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("file still exists: %s", path)
	}
}

func TestSetDescription(t *testing.T) {
	dir := tempDir(t)

	ticket, err := Add(dir, "Desc test", "")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	_, err = SetDescription(dir, ticket.ID, "New description here.")
	if err != nil {
		t.Fatalf("SetDescription: %v", err)
	}

	loaded, err := Show(dir, ticket.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if loaded.Description != "New description here." {
		t.Errorf("description = %q", loaded.Description)
	}
}

func TestSetDescriptionWithBackticks(t *testing.T) {
	dir := tempDir(t)

	ticket, err := Add(dir, "Backtick desc", "")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	desc := "Use `fmt.Println` for output.\n\n```go\nfmt.Println(\"hello\")\n```"
	desc = strings.ReplaceAll(desc, "\\n", "\n")

	_, err = SetDescription(dir, ticket.ID, desc)
	if err != nil {
		t.Fatalf("SetDescription: %v", err)
	}

	loaded, err := Show(dir, ticket.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if loaded.Description != desc {
		t.Errorf("description mismatch.\ngot:  %q\nwant: %q", loaded.Description, desc)
	}
}

func TestMultipleTickets(t *testing.T) {
	dir := tempDir(t)

	Add(dir, "First", "")
	second, err := Add(dir, "Second", "With description")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	Add(dir, "Third", "")

	tickets, err := List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tickets) != 3 {
		t.Fatalf("len = %d, want 3", len(tickets))
	}

	// Remove middle one by ID
	Done(dir, second.ID)

	tickets, err = List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tickets) != 2 {
		t.Fatalf("len = %d, want 2", len(tickets))
	}
}

func TestNotFoundErrors(t *testing.T) {
	dir := tempDir(t)

	// Ensure directory exists
	EnsureDir(dir)

	_, err := Show(dir, "nonexistent")
	if err == nil {
		t.Error("Show: expected error")
	}

	_, err = Done(dir, "nonexistent")
	if err == nil {
		t.Error("Done: expected error")
	}

	_, err = SetDescription(dir, "nonexistent", "desc")
	if err == nil {
		t.Error("SetDescription: expected error")
	}
}

func TestListEmptyDir(t *testing.T) {
	dir := tempDir(t)

	tickets, err := List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tickets) != 0 {
		t.Errorf("len = %d, want 0", len(tickets))
	}
}

func TestFileFormat(t *testing.T) {
	dir := tempDir(t)

	ticket, err := Add(dir, "My Ticket", "Some description")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Check file exists with <id>.md naming
	expectedFilename := ticket.ID + ".md"
	path := filepath.Join(DirPath(dir), expectedFilename)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)

	// Verify YAML-frontmatter-first format
	if !strings.HasPrefix(content, "---\n") {
		t.Error("should start with YAML frontmatter delimiter")
	}
	if !strings.Contains(content, "id: "+ticket.ID) {
		t.Error("missing id in frontmatter")
	}
	if !strings.Contains(content, "# My Ticket") {
		t.Error("missing title heading after frontmatter")
	}
	if !strings.Contains(content, "Some description") {
		t.Error("missing description")
	}

	// Verify order: frontmatter before title
	fmEnd := strings.Index(content[4:], "\n---\n")
	titleIdx := strings.Index(content, "# My Ticket")
	if fmEnd < 0 || titleIdx < 0 || titleIdx < fmEnd {
		t.Error("title should come after frontmatter closing delimiter")
	}
}

func TestTicketFileName(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"abc", "abc.md"},
		{"XYZ", "XYZ.md"},
		{"123", "123.md"},
	}

	for _, tt := range tests {
		got := ticketFileName(tt.id)
		if got != tt.want {
			t.Errorf("ticketFileName(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestEnsureDir(t *testing.T) {
	dir := tempDir(t)

	ticketsDir := DirPath(dir)

	// Directory shouldn't exist yet
	if _, err := os.Stat(ticketsDir); !os.IsNotExist(err) {
		t.Error("tickets dir should not exist initially")
	}

	// Ensure creates it
	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}

	// Now it should exist
	info, err := os.Stat(ticketsDir)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if !info.IsDir() {
		t.Error("tickets should be a directory")
	}

	// Calling again should be fine
	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir (second call): %v", err)
	}
}

func TestParseFileRoundTripAllFields(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	original := &Ticket{
		Title:       "Full ticket",
		ID:          "ful",
		Description: "A complete ticket.",
		Status:      "open",
		Type:        "feature",
		Priority:    3,
		Assignee:    "alice",
		Created:     "2026-01-15",
		Parent:      "prt",
		ExternalRef: "JIRA-123",
		Deps:        []string{"dep1", "dep2"},
		Links:       []string{"lnk1"},
		Tags:        []string{"backend", "urgent"},
	}

	if err := writeFile(dir, original); err != nil {
		t.Fatalf("writeFile: %v", err)
	}

	loaded, err := Show(dir, "ful")
	if err != nil {
		t.Fatalf("Show: %v", err)
	}

	if loaded.Title != original.Title {
		t.Errorf("Title = %q, want %q", loaded.Title, original.Title)
	}
	if loaded.ID != original.ID {
		t.Errorf("ID = %q, want %q", loaded.ID, original.ID)
	}
	if loaded.Description != original.Description {
		t.Errorf("Description = %q, want %q", loaded.Description, original.Description)
	}
	if loaded.Status != original.Status {
		t.Errorf("Status = %q, want %q", loaded.Status, original.Status)
	}
	if loaded.Type != original.Type {
		t.Errorf("Type = %q, want %q", loaded.Type, original.Type)
	}
	if loaded.Priority != original.Priority {
		t.Errorf("Priority = %d, want %d", loaded.Priority, original.Priority)
	}
	if loaded.Assignee != original.Assignee {
		t.Errorf("Assignee = %q, want %q", loaded.Assignee, original.Assignee)
	}
	if loaded.Created != original.Created {
		t.Errorf("Created = %q, want %q", loaded.Created, original.Created)
	}
	if loaded.Parent != original.Parent {
		t.Errorf("Parent = %q, want %q", loaded.Parent, original.Parent)
	}
	if loaded.ExternalRef != original.ExternalRef {
		t.Errorf("ExternalRef = %q, want %q", loaded.ExternalRef, original.ExternalRef)
	}
	if len(loaded.Deps) != len(original.Deps) {
		t.Errorf("Deps len = %d, want %d", len(loaded.Deps), len(original.Deps))
	}
	if len(loaded.Links) != len(original.Links) {
		t.Errorf("Links len = %d, want %d", len(loaded.Links), len(original.Links))
	}
	if len(loaded.Tags) != len(original.Tags) {
		t.Errorf("Tags len = %d, want %d", len(loaded.Tags), len(original.Tags))
	}
}

func TestParseFileDescriptionWithDashes(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	original := &Ticket{
		Title:       "Dash test",
		ID:          "dsh",
		Description: "Some text\n---\nMore text after dashes",
	}

	if err := writeFile(dir, original); err != nil {
		t.Fatalf("writeFile: %v", err)
	}

	loaded, err := Show(dir, "dsh")
	if err != nil {
		t.Fatalf("Show: %v", err)
	}

	if loaded.Description != original.Description {
		t.Errorf("Description = %q, want %q", loaded.Description, original.Description)
	}
}
