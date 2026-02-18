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

	ticket, err := Add(dir, &Ticket{Title: "Fix login bug"})
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

	ticket, err := Add(dir, &Ticket{Title: "Refactor auth", Description: "Move auth to middleware layer."})
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

	ticket, err := Add(dir, &Ticket{Title: "Backtick test", Description: desc})
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

	ticket, err := Add(dir, &Ticket{Title: "Test ticket"})
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

	ticket, err := Add(dir, &Ticket{Title: "To close"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	title, err := Done(dir, ticket.ID)
	if err != nil {
		t.Fatalf("Done: %v", err)
	}
	if title != "To close" {
		t.Errorf("title = %q", title)
	}

	// File should still exist
	path := filepath.Join(DirPath(dir), ticket.ID+".md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file should still exist: %v", err)
	}

	// Ticket should have status=closed
	loaded, err := Show(dir, ticket.ID)
	if err != nil {
		t.Fatalf("Show after Done: %v", err)
	}
	if loaded.Status != "closed" {
		t.Errorf("status = %q, want %q", loaded.Status, "closed")
	}

	// List still returns all tickets (including closed)
	tickets, err := List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tickets) != 1 {
		t.Errorf("len = %d, want 1", len(tickets))
	}
}

func TestSetDescription(t *testing.T) {
	dir := tempDir(t)

	ticket, err := Add(dir, &Ticket{Title: "Desc test"})
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

	ticket, err := Add(dir, &Ticket{Title: "Backtick desc"})
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

	Add(dir, &Ticket{Title: "First"})
	second, err := Add(dir, &Ticket{Title: "Second", Description: "With description"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	Add(dir, &Ticket{Title: "Third"})

	tickets, err := List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tickets) != 3 {
		t.Fatalf("len = %d, want 3", len(tickets))
	}

	// Close middle one by ID
	Done(dir, second.ID)

	// List still returns all 3 (including closed)
	tickets, err = List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tickets) != 3 {
		t.Fatalf("len = %d, want 3", len(tickets))
	}

	// Verify the closed ticket has status=closed
	closed, err := Show(dir, second.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if closed.Status != "closed" {
		t.Errorf("status = %q, want %q", closed.Status, "closed")
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

	ticket, err := Add(dir, &Ticket{Title: "My Ticket", Description: "Some description"})
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
		Design:      "Use microservices",
		Acceptance:  "All tests pass",
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
	if loaded.Design != original.Design {
		t.Errorf("Design = %q, want %q", loaded.Design, original.Design)
	}
	if loaded.Acceptance != original.Acceptance {
		t.Errorf("Acceptance = %q, want %q", loaded.Acceptance, original.Acceptance)
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

func TestAddWithParentValidation(t *testing.T) {
	dir := tempDir(t)

	// Adding with a non-existent parent should fail
	_, err := Add(dir, &Ticket{Title: "Child", Parent: "nonexistent"})
	if err == nil {
		t.Error("Add with non-existent parent should fail")
	}
	if err != nil && !strings.Contains(err.Error(), "parent ticket not found") {
		t.Errorf("unexpected error: %v", err)
	}

	// Create a parent ticket
	parent, err := Add(dir, &Ticket{Title: "Parent ticket"})
	if err != nil {
		t.Fatalf("Add parent: %v", err)
	}

	// Adding with an existing parent should succeed
	child, err := Add(dir, &Ticket{Title: "Child ticket", Parent: parent.ID})
	if err != nil {
		t.Fatalf("Add child: %v", err)
	}
	if child.Parent != parent.ID {
		t.Errorf("Parent = %q, want %q", child.Parent, parent.ID)
	}

	// Verify round-trip
	loaded, err := Show(dir, child.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if loaded.Parent != parent.ID {
		t.Errorf("loaded Parent = %q, want %q", loaded.Parent, parent.ID)
	}
}

func TestSetStatus(t *testing.T) {
	dir := tempDir(t)

	ticket, err := Add(dir, &Ticket{Title: "Status test"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Set to in_progress
	title, err := SetStatus(dir, ticket.ID, "in_progress")
	if err != nil {
		t.Fatalf("SetStatus: %v", err)
	}
	if title != "Status test" {
		t.Errorf("title = %q, want %q", title, "Status test")
	}

	loaded, err := Show(dir, ticket.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if loaded.Status != "in_progress" {
		t.Errorf("status = %q, want %q", loaded.Status, "in_progress")
	}

	// Set to closed
	_, err = SetStatus(dir, ticket.ID, "closed")
	if err != nil {
		t.Fatalf("SetStatus closed: %v", err)
	}
	loaded, err = Show(dir, ticket.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if loaded.Status != "closed" {
		t.Errorf("status = %q, want %q", loaded.Status, "closed")
	}

	// Set to open
	_, err = SetStatus(dir, ticket.ID, "open")
	if err != nil {
		t.Fatalf("SetStatus open: %v", err)
	}
	loaded, err = Show(dir, ticket.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if loaded.Status != "open" {
		t.Errorf("status = %q, want %q", loaded.Status, "open")
	}
}

func TestSetStatusInvalid(t *testing.T) {
	dir := tempDir(t)

	ticket, err := Add(dir, &Ticket{Title: "Invalid status test"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	_, err = SetStatus(dir, ticket.ID, "invalid")
	if err == nil {
		t.Error("SetStatus with invalid status should fail")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetStatusNotFound(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	_, err := SetStatus(dir, "zzz", "open")
	if err == nil {
		t.Error("SetStatus with non-existent ID should fail")
	}
	if err != nil && !strings.Contains(err.Error(), "ticket not found") {
		t.Errorf("unexpected error: %v", err)
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

func TestAddDep(t *testing.T) {
	dir := tempDir(t)

	a, err := Add(dir, &Ticket{Title: "Ticket A"})
	if err != nil {
		t.Fatalf("Add A: %v", err)
	}
	b, err := Add(dir, &Ticket{Title: "Ticket B"})
	if err != nil {
		t.Fatalf("Add B: %v", err)
	}

	err = AddDep(dir, a.ID, b.ID)
	if err != nil {
		t.Fatalf("AddDep: %v", err)
	}

	loaded, err := Show(dir, a.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if len(loaded.Deps) != 1 {
		t.Fatalf("Deps len = %d, want 1", len(loaded.Deps))
	}
	if loaded.Deps[0] != b.ID {
		t.Errorf("Deps[0] = %q, want %q", loaded.Deps[0], b.ID)
	}
}

func TestAddDepIdempotent(t *testing.T) {
	dir := tempDir(t)

	a, err := Add(dir, &Ticket{Title: "Ticket A"})
	if err != nil {
		t.Fatalf("Add A: %v", err)
	}
	b, err := Add(dir, &Ticket{Title: "Ticket B"})
	if err != nil {
		t.Fatalf("Add B: %v", err)
	}

	// Add twice — should be idempotent
	AddDep(dir, a.ID, b.ID)
	err = AddDep(dir, a.ID, b.ID)
	if err != nil {
		t.Fatalf("AddDep (second): %v", err)
	}

	loaded, err := Show(dir, a.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if len(loaded.Deps) != 1 {
		t.Errorf("Deps len = %d, want 1 (should not duplicate)", len(loaded.Deps))
	}
}

func TestAddDepTicketNotFound(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	b, err := Add(dir, &Ticket{Title: "Ticket B"})
	if err != nil {
		t.Fatalf("Add B: %v", err)
	}

	err = AddDep(dir, "zzz", b.ID)
	if err == nil {
		t.Error("AddDep with non-existent ticket should fail")
	}
	if err != nil && !strings.Contains(err.Error(), "ticket not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddDepDepNotFound(t *testing.T) {
	dir := tempDir(t)

	a, err := Add(dir, &Ticket{Title: "Ticket A"})
	if err != nil {
		t.Fatalf("Add A: %v", err)
	}

	err = AddDep(dir, a.ID, "zzz")
	if err == nil {
		t.Error("AddDep with non-existent dep should fail")
	}
	if err != nil && !strings.Contains(err.Error(), "ticket not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRemoveDep(t *testing.T) {
	dir := tempDir(t)

	a, err := Add(dir, &Ticket{Title: "Ticket A"})
	if err != nil {
		t.Fatalf("Add A: %v", err)
	}
	b, err := Add(dir, &Ticket{Title: "Ticket B"})
	if err != nil {
		t.Fatalf("Add B: %v", err)
	}

	AddDep(dir, a.ID, b.ID)

	err = RemoveDep(dir, a.ID, b.ID)
	if err != nil {
		t.Fatalf("RemoveDep: %v", err)
	}

	loaded, err := Show(dir, a.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if len(loaded.Deps) != 0 {
		t.Errorf("Deps len = %d, want 0", len(loaded.Deps))
	}
}

func TestRemoveDepNotPresent(t *testing.T) {
	dir := tempDir(t)

	a, err := Add(dir, &Ticket{Title: "Ticket A"})
	if err != nil {
		t.Fatalf("Add A: %v", err)
	}
	b, err := Add(dir, &Ticket{Title: "Ticket B"})
	if err != nil {
		t.Fatalf("Add B: %v", err)
	}

	// Remove dep that was never added — should succeed (idempotent)
	err = RemoveDep(dir, a.ID, b.ID)
	if err != nil {
		t.Fatalf("RemoveDep (not present): %v", err)
	}
}

func TestRemoveDepTicketNotFound(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	b, err := Add(dir, &Ticket{Title: "Ticket B"})
	if err != nil {
		t.Fatalf("Add B: %v", err)
	}

	err = RemoveDep(dir, "zzz", b.ID)
	if err == nil {
		t.Error("RemoveDep with non-existent ticket should fail")
	}
	if err != nil && !strings.Contains(err.Error(), "ticket not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddLink(t *testing.T) {
	dir := tempDir(t)

	a, err := Add(dir, &Ticket{Title: "Ticket A"})
	if err != nil {
		t.Fatalf("Add A: %v", err)
	}
	b, err := Add(dir, &Ticket{Title: "Ticket B"})
	if err != nil {
		t.Fatalf("Add B: %v", err)
	}

	err = AddLink(dir, []string{a.ID, b.ID})
	if err != nil {
		t.Fatalf("AddLink: %v", err)
	}

	// Both tickets should have each other in links
	loadedA, err := Show(dir, a.ID)
	if err != nil {
		t.Fatalf("Show A: %v", err)
	}
	if len(loadedA.Links) != 1 || loadedA.Links[0] != b.ID {
		t.Errorf("A.Links = %v, want [%s]", loadedA.Links, b.ID)
	}

	loadedB, err := Show(dir, b.ID)
	if err != nil {
		t.Fatalf("Show B: %v", err)
	}
	if len(loadedB.Links) != 1 || loadedB.Links[0] != a.ID {
		t.Errorf("B.Links = %v, want [%s]", loadedB.Links, a.ID)
	}
}

func TestAddLinkThreeTickets(t *testing.T) {
	dir := tempDir(t)

	a, err := Add(dir, &Ticket{Title: "Ticket A"})
	if err != nil {
		t.Fatalf("Add A: %v", err)
	}
	b, err := Add(dir, &Ticket{Title: "Ticket B"})
	if err != nil {
		t.Fatalf("Add B: %v", err)
	}
	c, err := Add(dir, &Ticket{Title: "Ticket C"})
	if err != nil {
		t.Fatalf("Add C: %v", err)
	}

	err = AddLink(dir, []string{a.ID, b.ID, c.ID})
	if err != nil {
		t.Fatalf("AddLink: %v", err)
	}

	// A should link to B and C
	loadedA, err := Show(dir, a.ID)
	if err != nil {
		t.Fatalf("Show A: %v", err)
	}
	if len(loadedA.Links) != 2 {
		t.Errorf("A.Links len = %d, want 2", len(loadedA.Links))
	}

	// B should link to A and C
	loadedB, err := Show(dir, b.ID)
	if err != nil {
		t.Fatalf("Show B: %v", err)
	}
	if len(loadedB.Links) != 2 {
		t.Errorf("B.Links len = %d, want 2", len(loadedB.Links))
	}

	// C should link to A and B
	loadedC, err := Show(dir, c.ID)
	if err != nil {
		t.Fatalf("Show C: %v", err)
	}
	if len(loadedC.Links) != 2 {
		t.Errorf("C.Links len = %d, want 2", len(loadedC.Links))
	}
}

func TestAddLinkIdempotent(t *testing.T) {
	dir := tempDir(t)

	a, err := Add(dir, &Ticket{Title: "Ticket A"})
	if err != nil {
		t.Fatalf("Add A: %v", err)
	}
	b, err := Add(dir, &Ticket{Title: "Ticket B"})
	if err != nil {
		t.Fatalf("Add B: %v", err)
	}

	// Link twice — should be idempotent
	AddLink(dir, []string{a.ID, b.ID})
	err = AddLink(dir, []string{a.ID, b.ID})
	if err != nil {
		t.Fatalf("AddLink (second): %v", err)
	}

	loadedA, err := Show(dir, a.ID)
	if err != nil {
		t.Fatalf("Show A: %v", err)
	}
	if len(loadedA.Links) != 1 {
		t.Errorf("A.Links len = %d, want 1 (should not duplicate)", len(loadedA.Links))
	}

	loadedB, err := Show(dir, b.ID)
	if err != nil {
		t.Fatalf("Show B: %v", err)
	}
	if len(loadedB.Links) != 1 {
		t.Errorf("B.Links len = %d, want 1 (should not duplicate)", len(loadedB.Links))
	}
}

func TestAddLinkTicketNotFound(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	a, err := Add(dir, &Ticket{Title: "Ticket A"})
	if err != nil {
		t.Fatalf("Add A: %v", err)
	}

	err = AddLink(dir, []string{a.ID, "zzz"})
	if err == nil {
		t.Error("AddLink with non-existent ticket should fail")
	}
	if err != nil && !strings.Contains(err.Error(), "ticket not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRemoveLink(t *testing.T) {
	dir := tempDir(t)

	a, err := Add(dir, &Ticket{Title: "Ticket A"})
	if err != nil {
		t.Fatalf("Add A: %v", err)
	}
	b, err := Add(dir, &Ticket{Title: "Ticket B"})
	if err != nil {
		t.Fatalf("Add B: %v", err)
	}

	AddLink(dir, []string{a.ID, b.ID})

	err = RemoveLink(dir, a.ID, b.ID)
	if err != nil {
		t.Fatalf("RemoveLink: %v", err)
	}

	// Both should have empty links
	loadedA, err := Show(dir, a.ID)
	if err != nil {
		t.Fatalf("Show A: %v", err)
	}
	if len(loadedA.Links) != 0 {
		t.Errorf("A.Links len = %d, want 0", len(loadedA.Links))
	}

	loadedB, err := Show(dir, b.ID)
	if err != nil {
		t.Fatalf("Show B: %v", err)
	}
	if len(loadedB.Links) != 0 {
		t.Errorf("B.Links len = %d, want 0", len(loadedB.Links))
	}
}

func TestRemoveLinkIdempotent(t *testing.T) {
	dir := tempDir(t)

	a, err := Add(dir, &Ticket{Title: "Ticket A"})
	if err != nil {
		t.Fatalf("Add A: %v", err)
	}
	b, err := Add(dir, &Ticket{Title: "Ticket B"})
	if err != nil {
		t.Fatalf("Add B: %v", err)
	}

	// Remove link that was never added — should succeed (idempotent)
	err = RemoveLink(dir, a.ID, b.ID)
	if err != nil {
		t.Fatalf("RemoveLink (not present): %v", err)
	}
}

func TestRemoveLinkTicketNotFound(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	a, err := Add(dir, &Ticket{Title: "Ticket A"})
	if err != nil {
		t.Fatalf("Add A: %v", err)
	}

	err = RemoveLink(dir, a.ID, "zzz")
	if err == nil {
		t.Error("RemoveLink with non-existent ticket should fail")
	}
	if err != nil && !strings.Contains(err.Error(), "ticket not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFindTicketFilePartialPrefix(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aBc", Title: "Prefix test"})

	// "aB" is a prefix of "aBc" — should match
	loaded, err := Show(dir, "aB")
	if err != nil {
		t.Fatalf("Show with prefix: %v", err)
	}
	if loaded.ID != "aBc" {
		t.Errorf("ID = %q, want %q", loaded.ID, "aBc")
	}
}

func TestFindTicketFilePartialSuffix(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aBc", Title: "Suffix test"})

	// "Bc" is a suffix of "aBc" — should match
	loaded, err := Show(dir, "Bc")
	if err != nil {
		t.Fatalf("Show with suffix: %v", err)
	}
	if loaded.ID != "aBc" {
		t.Errorf("ID = %q, want %q", loaded.ID, "aBc")
	}
}

func TestFindTicketFilePartialSubstring(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "xYz", Title: "Substring test"})

	// "Y" is a substring of "xYz" — should match
	loaded, err := Show(dir, "Y")
	if err != nil {
		t.Fatalf("Show with substring: %v", err)
	}
	if loaded.ID != "xYz" {
		t.Errorf("ID = %q, want %q", loaded.ID, "xYz")
	}
}

func TestFindTicketFileAmbiguous(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aXb", Title: "First"})
	writeFile(dir, &Ticket{ID: "cXd", Title: "Second"})

	// "X" matches both "aXb" and "cXd" — should error
	_, err := Show(dir, "X")
	if err == nil {
		t.Fatal("Show with ambiguous partial should fail")
	}
	if !strings.Contains(err.Error(), "ambiguous ticket ID") {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "aXb") || !strings.Contains(err.Error(), "cXd") {
		t.Errorf("error should list matching IDs: %v", err)
	}
}

func TestFindTicketFileExactPrecedence(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	// Create ticket "ab" and ticket "abc"
	// Searching for "ab" should exact-match "ab", not partially match both
	writeFile(dir, &Ticket{ID: "ab", Title: "Exact"})
	writeFile(dir, &Ticket{ID: "abc", Title: "Longer"})

	loaded, err := Show(dir, "ab")
	if err != nil {
		t.Fatalf("Show with exact match: %v", err)
	}
	if loaded.ID != "ab" {
		t.Errorf("ID = %q, want %q (exact match should take precedence)", loaded.ID, "ab")
	}
}

func TestFindTicketFileNotFound(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aBc", Title: "Exists"})

	// "ZZZ" matches nothing
	_, err := Show(dir, "ZZZ")
	if err == nil {
		t.Fatal("Show with non-matching partial should fail")
	}
	if !strings.Contains(err.Error(), "ticket not found") {
		t.Errorf("unexpected error: %v", err)
	}
}
