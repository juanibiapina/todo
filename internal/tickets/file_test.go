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
	if ticket.State != StateNew {
		t.Errorf("state = %q, want %q", ticket.State, StateNew)
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
	if ticket.State != StateRefined {
		t.Errorf("state = %q, want %q", ticket.State, StateRefined)
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

func TestShowByIDAndTitle(t *testing.T) {
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

	byTitle, err := Show(dir, "Test ticket")
	if err != nil {
		t.Fatalf("Show by title: %v", err)
	}
	if byTitle.ID != ticket.ID {
		t.Errorf("byTitle ID = %q, want %q", byTitle.ID, ticket.ID)
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
}

func TestSetState(t *testing.T) {
	dir := tempDir(t)

	ticket, err := Add(dir, "State test", "")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	_, err = SetState(dir, ticket.ID, StatePlanned)
	if err != nil {
		t.Fatalf("SetState: %v", err)
	}

	loaded, err := Show(dir, ticket.ID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if loaded.State != StatePlanned {
		t.Errorf("state = %q, want %q", loaded.State, StatePlanned)
	}
}

func TestSetStateInvalid(t *testing.T) {
	dir := tempDir(t)

	Add(dir, "State test", "")

	_, err := SetState(dir, "State test", State("invalid"))
	if err == nil {
		t.Fatal("expected error for invalid state")
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
	// Should auto-promote to refined
	if loaded.State != StateRefined {
		t.Errorf("state = %q, want refined", loaded.State)
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
	Add(dir, "Second", "With description")
	Add(dir, "Third", "")

	tickets, err := List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tickets) != 3 {
		t.Fatalf("len = %d, want 3", len(tickets))
	}

	// Remove middle one
	Done(dir, "Second")

	tickets, err = List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tickets) != 2 {
		t.Fatalf("len = %d, want 2", len(tickets))
	}
	if tickets[0].Title != "First" || tickets[1].Title != "Third" {
		t.Errorf("remaining = %q, %q", tickets[0].Title, tickets[1].Title)
	}
}

func TestNotFoundErrors(t *testing.T) {
	dir := tempDir(t)

	_, err := Show(dir, "nonexistent")
	if err == nil {
		t.Error("Show: expected error")
	}

	_, err = Done(dir, "nonexistent")
	if err == nil {
		t.Error("Done: expected error")
	}

	_, err = SetState(dir, "nonexistent", StateNew)
	if err == nil {
		t.Error("SetState: expected error")
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

	Add(dir, "My Ticket", "Some description")

	data, err := os.ReadFile(filepath.Join(dir, ".tickets.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# Tickets") {
		t.Error("missing header")
	}
	if !strings.Contains(content, "## My Ticket") {
		t.Error("missing ticket heading")
	}
	if !strings.Contains(content, "state: refined") {
		t.Error("missing state")
	}
	if !strings.Contains(content, "Some description") {
		t.Error("missing description")
	}
}
