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

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Fix login bug", "fix-login-bug"},
		{"Hello World!", "hello-world"},
		{"  spaces  ", "spaces"},
		{"UPPERCASE", "uppercase"},
		{"special@#$chars", "special-chars"},
		{"multiple---hyphens", "multiple-hyphens"},
		{"", ""},
	}

	for _, tt := range tests {
		got := slugify(tt.input)
		if got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
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
	pattern := filepath.Join(DirPath(dir), ticket.ID+"-*.md")
	matches, _ := filepath.Glob(pattern)
	if len(matches) != 0 {
		t.Errorf("file still exists: %v", matches)
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

	// Check file exists in tickets directory with correct name pattern
	expectedFilename := ticket.ID + "-my-ticket.md"
	path := filepath.Join(DirPath(dir), expectedFilename)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "# My Ticket\n") {
		t.Error("missing or wrong title heading")
	}
	if !strings.Contains(content, "id: "+ticket.ID) {
		t.Error("missing id")
	}
	if !strings.Contains(content, "Some description") {
		t.Error("missing description")
	}
}

func TestTicketFileName(t *testing.T) {
	tests := []struct {
		id    string
		title string
		want  string
	}{
		{"abc", "Fix login bug", "abc-fix-login-bug.md"},
		{"XYZ", "Hello World!", "XYZ-hello-world.md"},
		{"123", "Test", "123-test.md"},
	}

	for _, tt := range tests {
		got := ticketFileName(tt.id, tt.title)
		if got != tt.want {
			t.Errorf("ticketFileName(%q, %q) = %q, want %q", tt.id, tt.title, got, tt.want)
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
