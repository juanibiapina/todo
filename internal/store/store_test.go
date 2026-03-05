package store

import (
	"os"
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestAdd(t *testing.T) {
	s := openTestStore(t)

	item, err := s.Add("/tmp/project", "Buy groceries")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	if item.ID <= 0 {
		t.Errorf("expected positive ID, got %d", item.ID)
	}
	if item.Text != "Buy groceries" {
		t.Errorf("expected text 'Buy groceries', got %q", item.Text)
	}
	if item.Directory != "/tmp/project" {
		t.Errorf("expected directory '/tmp/project', got %q", item.Directory)
	}
	if item.Checked {
		t.Error("expected unchecked")
	}
}

func TestAddAutoIncrements(t *testing.T) {
	s := openTestStore(t)

	i1, _ := s.Add("/d", "First")
	i2, _ := s.Add("/d", "Second")
	i3, _ := s.Add("/d", "Third")

	if i2.ID <= i1.ID || i3.ID <= i2.ID {
		t.Errorf("expected ascending IDs, got %d, %d, %d", i1.ID, i2.ID, i3.ID)
	}
}

func TestListEmpty(t *testing.T) {
	s := openTestStore(t)

	items, err := s.List("/tmp/project")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestListReturnsItemsForDirectory(t *testing.T) {
	s := openTestStore(t)

	s.Add("/project-a", "Task A1")
	s.Add("/project-a", "Task A2")
	s.Add("/project-b", "Task B1")

	items, err := s.List("/project-a")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Text != "Task A1" || items[1].Text != "Task A2" {
		t.Errorf("unexpected items: %v, %v", items[0].Text, items[1].Text)
	}
}

func TestListOrderedByID(t *testing.T) {
	s := openTestStore(t)

	s.Add("/d", "First")
	s.Add("/d", "Second")
	s.Add("/d", "Third")

	items, _ := s.List("/d")
	for i := 1; i < len(items); i++ {
		if items[i].ID <= items[i-1].ID {
			t.Errorf("items not ordered: id %d after %d", items[i].ID, items[i-1].ID)
		}
	}
}

func TestListIsolatesByDirectory(t *testing.T) {
	s := openTestStore(t)

	s.Add("/dir-a", "A item")
	s.Add("/dir-b", "B item")

	items, _ := s.List("/dir-a")
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Text != "A item" {
		t.Errorf("expected 'A item', got %q", items[0].Text)
	}
}

func TestCheck(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Task")

	err := s.Check("/d", item.ID)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	items, _ := s.List("/d")
	if !items[0].Checked {
		t.Error("expected item to be checked")
	}
}

func TestCheckIdempotent(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Task")
	s.Check("/d", item.ID)
	err := s.Check("/d", item.ID)
	if err != nil {
		t.Fatalf("second Check should succeed: %v", err)
	}
}

func TestCheckNotFound(t *testing.T) {
	s := openTestStore(t)

	err := s.Check("/d", 999)
	if err == nil {
		t.Fatal("expected error for non-existent item")
	}
}

func TestCheckWrongDirectory(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/dir-a", "Task")

	err := s.Check("/dir-b", item.ID)
	if err == nil {
		t.Fatal("expected error for wrong directory")
	}
}

func TestUncheck(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Task")
	s.Check("/d", item.ID)

	err := s.Uncheck("/d", item.ID)
	if err != nil {
		t.Fatalf("Uncheck: %v", err)
	}

	items, _ := s.List("/d")
	if items[0].Checked {
		t.Error("expected item to be unchecked")
	}
}

func TestUncheckNotFound(t *testing.T) {
	s := openTestStore(t)

	err := s.Uncheck("/d", 999)
	if err == nil {
		t.Fatal("expected error for non-existent item")
	}
}

func TestToggle(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Task")

	// First toggle: unchecked -> checked
	newState, err := s.Toggle("/d", item.ID)
	if err != nil {
		t.Fatalf("Toggle: %v", err)
	}
	if !newState {
		t.Error("expected checked after first toggle")
	}

	// Second toggle: checked -> unchecked
	newState, err = s.Toggle("/d", item.ID)
	if err != nil {
		t.Fatalf("Toggle: %v", err)
	}
	if newState {
		t.Error("expected unchecked after second toggle")
	}
}

func TestGet(t *testing.T) {
	s := openTestStore(t)

	added, _ := s.Add("/d", "My task")

	item, err := s.Get("/d", added.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if item.Text != "My task" {
		t.Errorf("expected 'My task', got %q", item.Text)
	}
}

func TestGetNotFound(t *testing.T) {
	s := openTestStore(t)

	_, err := s.Get("/d", 999)
	if err == nil {
		t.Fatal("expected error for non-existent item")
	}
}

func TestClean(t *testing.T) {
	s := openTestStore(t)

	i1, _ := s.Add("/d", "Keep this")
	i2, _ := s.Add("/d", "Remove this")
	i3, _ := s.Add("/d", "Also remove")

	s.Check("/d", i2.ID)
	s.Check("/d", i3.ID)

	n, err := s.Clean("/d")
	if err != nil {
		t.Fatalf("Clean: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 deleted, got %d", n)
	}

	items, _ := s.List("/d")
	if len(items) != 1 {
		t.Fatalf("expected 1 item remaining, got %d", len(items))
	}
	if items[0].ID != i1.ID {
		t.Errorf("expected remaining item ID %d, got %d", i1.ID, items[0].ID)
	}
}

func TestCleanNoCheckedItems(t *testing.T) {
	s := openTestStore(t)

	s.Add("/d", "Unchecked")

	n, err := s.Clean("/d")
	if err != nil {
		t.Fatalf("Clean: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 deleted, got %d", n)
	}
}

func TestCleanIsolatesByDirectory(t *testing.T) {
	s := openTestStore(t)

	i1, _ := s.Add("/dir-a", "A checked")
	s.Check("/dir-a", i1.ID)

	s.Add("/dir-b", "B unchecked")
	i3, _ := s.Add("/dir-b", "B checked")
	s.Check("/dir-b", i3.ID)

	// Clean only dir-b
	n, _ := s.Clean("/dir-b")
	if n != 1 {
		t.Errorf("expected 1 deleted from dir-b, got %d", n)
	}

	// dir-a still has its checked item
	items, _ := s.List("/dir-a")
	if len(items) != 1 {
		t.Errorf("expected 1 item in dir-a, got %d", len(items))
	}
}

func TestDefaultDBPath(t *testing.T) {
	// Test TODO_DB override
	t.Setenv("TODO_DB", "/custom/path.db")
	if got := DefaultDBPath(); got != "/custom/path.db" {
		t.Errorf("expected /custom/path.db, got %q", got)
	}

	// Test XDG_DATA_HOME
	t.Setenv("TODO_DB", "")
	t.Setenv("XDG_DATA_HOME", "/xdg/data")
	if got := DefaultDBPath(); got != "/xdg/data/todo/todo.db" {
		t.Errorf("expected /xdg/data/todo/todo.db, got %q", got)
	}

	// Test default (home-based)
	t.Setenv("XDG_DATA_HOME", "")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".local", "share", "todo", "todo.db")
	if got := DefaultDBPath(); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestOpenCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	dbPath := filepath.Join(dir, "test.db")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected database file to exist")
	}
}
