package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	filePath := filepath.Join(t.TempDir(), "test.md")
	s, err := Open(filePath)
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

func TestAddSequentialIDs(t *testing.T) {
	s := openTestStore(t)

	i1, _ := s.Add("/d", "First")
	i2, _ := s.Add("/d", "Second")
	i3, _ := s.Add("/d", "Third")

	if i1.ID != 1 || i2.ID != 2 || i3.ID != 3 {
		t.Errorf("expected IDs 1, 2, 3, got %d, %d, %d", i1.ID, i2.ID, i3.ID)
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

func TestListAllEmpty(t *testing.T) {
	s := openTestStore(t)

	items, err := s.ListAll()
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestListAllReturnsItemsFromAllDirectories(t *testing.T) {
	s := openTestStore(t)

	s.Add("/dir-a", "Task A1")
	s.Add("/dir-a", "Task A2")
	s.Add("/dir-b", "Task B1")

	items, err := s.ListAll()
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
}

func TestListAllPreservesDirectoryField(t *testing.T) {
	s := openTestStore(t)

	s.Add("/dir-a", "Task A")
	s.Add("/dir-b", "Task B")

	items, err := s.ListAll()
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}

	dirs := map[string]bool{}
	for _, item := range items {
		dirs[item.Directory] = true
	}
	if !dirs["/dir-a"] {
		t.Error("expected /dir-a in results")
	}
	if !dirs["/dir-b"] {
		t.Error("expected /dir-b in results")
	}
}

func TestListAllActiveItemsFirstWithinDirectory(t *testing.T) {
	s := openTestStore(t)

	s.Add("/d", "Normal")
	active, _ := s.Add("/d", "Active")
	s.ToggleActive("/d", active.ID)

	items, err := s.ListAll()
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if !items[0].IsActive {
		t.Errorf("expected active item first, got %q", items[0].Text)
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

func TestEdit(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Old text")

	err := s.Edit("/d", item.ID, "New text")
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}

	got, _ := s.Get("/d", item.ID)
	if got.Text != "New text" {
		t.Errorf("expected 'New text', got %q", got.Text)
	}
}

func TestEditNotFound(t *testing.T) {
	s := openTestStore(t)

	err := s.Edit("/d", 999, "Whatever")
	if err == nil {
		t.Fatal("expected error for non-existent item")
	}
}

func TestEditWrongDirectory(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/dir-a", "Task")

	err := s.Edit("/dir-b", item.ID, "Nope")
	if err == nil {
		t.Fatal("expected error for wrong directory")
	}
}

func TestEditPreservesCheckedState(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Task")
	s.Check("/d", item.ID)

	s.Edit("/d", item.ID, "Updated task")

	got, _ := s.Get("/d", item.ID)
	if !got.Checked {
		t.Error("expected item to remain checked after edit")
	}
	if got.Text != "Updated task" {
		t.Errorf("expected 'Updated task', got %q", got.Text)
	}
}

func TestDelete(t *testing.T) {
	s := openTestStore(t)

	s.Add("/d", "Keep")
	i2, _ := s.Add("/d", "Delete me")

	err := s.Delete("/d", i2.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, _ := s.List("/d")
	if len(items) != 1 {
		t.Fatalf("expected 1 item remaining, got %d", len(items))
	}
	if items[0].Text != "Keep" {
		t.Errorf("expected remaining item 'Keep', got %q", items[0].Text)
	}
}

func TestDeleteNotFound(t *testing.T) {
	s := openTestStore(t)

	err := s.Delete("/d", 999)
	if err == nil {
		t.Fatal("expected error for non-existent item")
	}
}

func TestDeleteWrongDirectory(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/dir-a", "Task")

	err := s.Delete("/dir-b", item.ID)
	if err == nil {
		t.Fatal("expected error for wrong directory")
	}
}

func TestClean(t *testing.T) {
	s := openTestStore(t)

	s.Add("/d", "Keep this")
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
	if items[0].Text != "Keep this" {
		t.Errorf("expected remaining item 'Keep this', got %q", items[0].Text)
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

func TestSwap(t *testing.T) {
	s := openTestStore(t)

	s.Add("/d", "First")
	s.Add("/d", "Second")
	s.Add("/d", "Third")

	items, _ := s.List("/d")
	// Swap first and second
	err := s.Swap("/d", items[0].ID, items[1].ID)
	if err != nil {
		t.Fatalf("Swap: %v", err)
	}

	items, _ = s.List("/d")
	if items[0].Text != "Second" {
		t.Errorf("expected first item 'Second', got %q", items[0].Text)
	}
	if items[1].Text != "First" {
		t.Errorf("expected second item 'First', got %q", items[1].Text)
	}
	if items[2].Text != "Third" {
		t.Errorf("expected third item 'Third', got %q", items[2].Text)
	}
}

func TestSwapNotFound(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Task")

	err := s.Swap("/d", item.ID, 999)
	if err == nil {
		t.Fatal("expected error for non-existent item")
	}
}

func TestSwapWrongDirectory(t *testing.T) {
	s := openTestStore(t)

	s.Add("/dir-a", "A")
	s.Add("/dir-b", "B")

	// /dir-a has only 1 item, so position 2 is out of range
	err := s.Swap("/dir-a", 1, 2)
	if err == nil {
		t.Fatal("expected error for out-of-range item")
	}
}

func TestListOrderedByPositionAfterSwap(t *testing.T) {
	s := openTestStore(t)

	s.Add("/d", "First")
	s.Add("/d", "Second")
	s.Add("/d", "Third")

	// Move third to first position by swapping down twice.
	// Reload between swaps since IDs are position-based.
	items, _ := s.List("/d")
	s.Swap("/d", items[2].ID, items[1].ID) // Swap positions 3 and 2

	items, _ = s.List("/d")
	s.Swap("/d", items[1].ID, items[0].ID) // Swap positions 2 and 1

	items, _ = s.List("/d")
	if items[0].Text != "Third" {
		t.Errorf("expected first 'Third', got %q", items[0].Text)
	}
	if items[1].Text != "First" {
		t.Errorf("expected second 'First', got %q", items[1].Text)
	}
	if items[2].Text != "Second" {
		t.Errorf("expected third 'Second', got %q", items[2].Text)
	}
}

func TestOpenCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	filePath := filepath.Join(dir, "test.md")

	s, err := Open(filePath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("expected directory to exist")
	}
}

func makeTime(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 12, 0, 0, 0, time.UTC)
}

func TestAddSection(t *testing.T) {
	s := openTestStore(t)

	item, err := s.AddSection("/d")
	if err != nil {
		t.Fatalf("AddSection: %v", err)
	}
	if !item.IsSection {
		t.Error("expected IsSection to be true")
	}
	if item.Text != "" {
		t.Errorf("expected empty text, got %q", item.Text)
	}
	if item.ID <= 0 {
		t.Errorf("expected positive ID, got %d", item.ID)
	}
}

func TestListReturnsSections(t *testing.T) {
	s := openTestStore(t)

	s.Add("/d", "Task 1")
	s.AddSection("/d")
	s.Add("/d", "Task 2")

	items, err := s.List("/d")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].IsSection {
		t.Error("expected first item to not be a section")
	}
	if !items[1].IsSection {
		t.Error("expected second item to be a section")
	}
	if items[2].IsSection {
		t.Error("expected third item to not be a section")
	}
}

func TestCheckRejectsSection(t *testing.T) {
	s := openTestStore(t)

	section, _ := s.AddSection("/d")
	err := s.Check("/d", section.ID)
	if err == nil {
		t.Fatal("expected error when checking a section")
	}
	if !strings.Contains(err.Error(), "is a section") {
		t.Errorf("expected 'is a section' error, got: %v", err)
	}
}

func TestUncheckRejectsSection(t *testing.T) {
	s := openTestStore(t)

	section, _ := s.AddSection("/d")
	err := s.Uncheck("/d", section.ID)
	if err == nil {
		t.Fatal("expected error when unchecking a section")
	}
	if !strings.Contains(err.Error(), "is a section") {
		t.Errorf("expected 'is a section' error, got: %v", err)
	}
}

func TestEditRejectsSection(t *testing.T) {
	s := openTestStore(t)

	section, _ := s.AddSection("/d")
	err := s.Edit("/d", section.ID, "text")
	if err == nil {
		t.Fatal("expected error when editing a section")
	}
	if !strings.Contains(err.Error(), "is a section") {
		t.Errorf("expected 'is a section' error, got: %v", err)
	}
}

func TestToggleRejectsSection(t *testing.T) {
	s := openTestStore(t)

	section, _ := s.AddSection("/d")
	_, err := s.Toggle("/d", section.ID)
	if err == nil {
		t.Fatal("expected error when toggling a section")
	}
	if !strings.Contains(err.Error(), "is a section") {
		t.Errorf("expected 'is a section' error, got: %v", err)
	}
}

func TestInsertItemAfter(t *testing.T) {
	s := openTestStore(t)

	s.Add("/d", "First")
	second, _ := s.Add("/d", "Second")
	s.Add("/d", "Third")

	inserted, err := s.InsertItemAfter("/d", "Between", second.ID)
	if err != nil {
		t.Fatalf("InsertItemAfter: %v", err)
	}
	if inserted.Text != "Between" {
		t.Errorf("expected text 'Between', got %q", inserted.Text)
	}
	if inserted.IsSection {
		t.Error("expected IsSection to be false")
	}

	items, _ := s.List("/d")
	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}
	texts := make([]string, len(items))
	for i, item := range items {
		texts[i] = item.Text
	}
	expected := []string{"First", "Second", "Between", "Third"}
	for i, exp := range expected {
		if texts[i] != exp {
			t.Errorf("position %d: expected %q, got %q", i, exp, texts[i])
		}
	}
}

func TestInsertSectionAfter(t *testing.T) {
	s := openTestStore(t)

	first, _ := s.Add("/d", "First")
	s.Add("/d", "Second")

	section, err := s.InsertSectionAfter("/d", first.ID)
	if err != nil {
		t.Fatalf("InsertSectionAfter: %v", err)
	}
	if !section.IsSection {
		t.Error("expected IsSection to be true")
	}

	items, _ := s.List("/d")
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if !items[1].IsSection {
		t.Errorf("expected section at position 1, got isSection=%v", items[1].IsSection)
	}
}

func TestCleanPreservesSections(t *testing.T) {
	s := openTestStore(t)

	s.AddSection("/d")
	item, _ := s.Add("/d", "Task")
	s.Check("/d", item.ID)

	n, err := s.Clean("/d")
	if err != nil {
		t.Fatalf("Clean: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 deleted, got %d", n)
	}

	items, _ := s.List("/d")
	if len(items) != 1 {
		t.Fatalf("expected 1 item remaining, got %d", len(items))
	}
	if !items[0].IsSection {
		t.Error("expected remaining item to be a section")
	}
}

func TestSwapWithSection(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Task")
	section, _ := s.AddSection("/d")

	err := s.Swap("/d", item.ID, section.ID)
	if err != nil {
		t.Fatalf("Swap: %v", err)
	}

	items, _ := s.List("/d")
	if !items[0].IsSection {
		t.Error("expected section first")
	}
	if items[1].Text != "Task" {
		t.Errorf("expected 'Task' second, got %q", items[1].Text)
	}
}

func TestToggleActive(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Task")

	// First toggle: inactive -> active
	newState, err := s.ToggleActive("/d", item.ID)
	if err != nil {
		t.Fatalf("ToggleActive: %v", err)
	}
	if !newState {
		t.Error("expected active after first toggle")
	}

	// Second toggle: active -> inactive
	newState, err = s.ToggleActive("/d", item.ID)
	if err != nil {
		t.Fatalf("ToggleActive: %v", err)
	}
	if newState {
		t.Error("expected inactive after second toggle")
	}
}

func TestToggleActiveRejectsSection(t *testing.T) {
	s := openTestStore(t)

	section, _ := s.AddSection("/d")
	_, err := s.ToggleActive("/d", section.ID)
	if err == nil {
		t.Fatal("expected error when toggling active on a section")
	}
	if !strings.Contains(err.Error(), "is a section") {
		t.Errorf("expected 'is a section' error, got: %v", err)
	}
}

func TestToggleActiveWrongDirectory(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/dir-a", "Task")

	_, err := s.ToggleActive("/dir-b", item.ID)
	if err == nil {
		t.Fatal("expected error for wrong directory")
	}
}

func TestToggleActiveNotFound(t *testing.T) {
	s := openTestStore(t)

	_, err := s.ToggleActive("/d", 999)
	if err == nil {
		t.Fatal("expected error for non-existent item")
	}
}

func TestListActiveItemsFirst(t *testing.T) {
	s := openTestStore(t)

	s.Add("/d", "First")
	s.Add("/d", "Second")
	third, _ := s.Add("/d", "Third")

	s.ToggleActive("/d", third.ID)

	items, _ := s.List("/d")
	if items[0].Text != "Third" {
		t.Errorf("expected active item 'Third' first, got %q", items[0].Text)
	}
	if items[1].Text != "First" {
		t.Errorf("expected 'First' second, got %q", items[1].Text)
	}
	if items[2].Text != "Second" {
		t.Errorf("expected 'Second' third, got %q", items[2].Text)
	}
}

func TestListActivePreservesRelativeOrder(t *testing.T) {
	s := openTestStore(t)

	first, _ := s.Add("/d", "First")
	s.Add("/d", "Second")
	third, _ := s.Add("/d", "Third")

	s.ToggleActive("/d", first.ID)
	s.ToggleActive("/d", third.ID)

	items, _ := s.List("/d")
	if items[0].Text != "First" {
		t.Errorf("expected 'First' first (active, lower position), got %q", items[0].Text)
	}
	if items[1].Text != "Third" {
		t.Errorf("expected 'Third' second (active, higher position), got %q", items[1].Text)
	}
	if items[2].Text != "Second" {
		t.Errorf("expected 'Second' third (not active), got %q", items[2].Text)
	}
}

func TestListMultipleActive(t *testing.T) {
	s := openTestStore(t)

	first, _ := s.Add("/d", "First")
	second, _ := s.Add("/d", "Second")
	s.Add("/d", "Third")

	s.ToggleActive("/d", first.ID)
	s.ToggleActive("/d", second.ID)

	items, _ := s.List("/d")
	if !items[0].IsActive || !items[1].IsActive {
		t.Error("expected first two items to be active")
	}
	if items[2].IsActive {
		t.Error("expected third item to not be active")
	}
}

func TestGetIncludesIsActive(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Task")

	got, _ := s.Get("/d", item.ID)
	if got.IsActive {
		t.Error("expected IsActive to be false initially")
	}

	s.ToggleActive("/d", item.ID)

	got, _ = s.Get("/d", item.ID)
	if !got.IsActive {
		t.Error("expected IsActive to be true after toggle")
	}
}

func TestCheckClearsActive(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Task")
	s.ToggleActive("/d", item.ID)

	s.Check("/d", item.ID)

	got, _ := s.Get("/d", item.ID)
	if !got.Checked {
		t.Error("expected item to be checked")
	}
	if got.IsActive {
		t.Error("expected IsActive to be false after check")
	}
}

func TestToggleToDoneClearsActive(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Task")
	s.ToggleActive("/d", item.ID)

	newState, err := s.Toggle("/d", item.ID)
	if err != nil {
		t.Fatalf("Toggle: %v", err)
	}
	if !newState {
		t.Error("expected checked after toggle")
	}

	got, _ := s.Get("/d", item.ID)
	if got.IsActive {
		t.Error("expected IsActive to be false after toggling to done")
	}
}

func TestUncheckDoesNotAffectActive(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Task")
	s.ToggleActive("/d", item.ID)
	s.Check("/d", item.ID)

	// Item is now checked and inactive (check cleared active).
	// Uncheck it and verify active stays false.
	s.Uncheck("/d", item.ID)

	got, _ := s.Get("/d", item.ID)
	if got.Checked {
		t.Error("expected item to be unchecked")
	}
	if got.IsActive {
		t.Error("expected IsActive to remain false after uncheck")
	}
}

func TestCleanIfNewDayFirstRun(t *testing.T) {
	s := openTestStore(t)
	s.now = func() time.Time { return makeTime(2025, 3, 8) }

	item, _ := s.Add("/d", "Task")
	s.Check("/d", item.ID)

	err := s.CleanIfNewDay("/d")
	if err != nil {
		t.Fatalf("CleanIfNewDay: %v", err)
	}

	// First run should NOT clean — checked item remains
	items, _ := s.List("/d")
	if len(items) != 1 {
		t.Fatalf("expected 1 item (first run should not clean), got %d", len(items))
	}
	if !items[0].Checked {
		t.Error("expected item to still be checked")
	}
}

func TestCleanIfNewDaySameDay(t *testing.T) {
	s := openTestStore(t)
	s.now = func() time.Time { return makeTime(2025, 3, 8) }

	item, _ := s.Add("/d", "Task")
	s.Check("/d", item.ID)

	// First call seeds the date
	s.CleanIfNewDay("/d")

	// Same day — should not clean
	err := s.CleanIfNewDay("/d")
	if err != nil {
		t.Fatalf("CleanIfNewDay: %v", err)
	}

	items, _ := s.List("/d")
	if len(items) != 1 {
		t.Fatalf("expected 1 item (same day should not clean), got %d", len(items))
	}
}

func TestCleanIfNewDayNewDay(t *testing.T) {
	s := openTestStore(t)
	day1 := makeTime(2025, 3, 8)
	day2 := makeTime(2025, 3, 9)
	s.now = func() time.Time { return day1 }

	// Add items and check some
	s.Add("/d", "Keep this")
	i2, _ := s.Add("/d", "Remove this")
	s.Check("/d", i2.ID)

	// Day 1: seed the date
	s.CleanIfNewDay("/d")

	// Advance to day 2
	s.now = func() time.Time { return day2 }

	err := s.CleanIfNewDay("/d")
	if err != nil {
		t.Fatalf("CleanIfNewDay: %v", err)
	}

	// Checked items should be gone
	items, _ := s.List("/d")
	if len(items) != 1 {
		t.Fatalf("expected 1 item after new-day clean, got %d", len(items))
	}
	if items[0].Text != "Keep this" {
		t.Errorf("expected remaining item 'Keep this', got %q", items[0].Text)
	}
}

func TestCleanIfNewDayDirectoryIsolation(t *testing.T) {
	s := openTestStore(t)
	day1 := makeTime(2025, 3, 8)
	day2 := makeTime(2025, 3, 9)
	s.now = func() time.Time { return day1 }

	// Add checked items in both
	ia, _ := s.Add("/dir-a", "A checked")
	s.Check("/dir-a", ia.ID)
	ib, _ := s.Add("/dir-b", "B checked")
	s.Check("/dir-b", ib.ID)

	// Seed both directories on day 1
	s.CleanIfNewDay("/dir-a")
	s.CleanIfNewDay("/dir-b")

	// Advance to day 2, clean only dir-a
	s.now = func() time.Time { return day2 }
	s.CleanIfNewDay("/dir-a")

	// dir-a should be cleaned
	itemsA, _ := s.List("/dir-a")
	if len(itemsA) != 0 {
		t.Errorf("expected 0 items in dir-a, got %d", len(itemsA))
	}

	// dir-b should still have its checked item
	itemsB, _ := s.List("/dir-b")
	if len(itemsB) != 1 {
		t.Errorf("expected 1 item in dir-b, got %d", len(itemsB))
	}
}

func TestCleanIfNewDayUpdatesDate(t *testing.T) {
	s := openTestStore(t)
	day1 := makeTime(2025, 3, 8)
	day2 := makeTime(2025, 3, 9)
	s.now = func() time.Time { return day1 }

	// Add and check an item on day 1
	i1, _ := s.Add("/d", "Day 1 task")
	s.Check("/d", i1.ID)

	// Seed the date on day 1
	s.CleanIfNewDay("/d")

	// Day 2: clean happens
	s.now = func() time.Time { return day2 }
	s.CleanIfNewDay("/d")

	// Add and check a new item on day 2
	i2, _ := s.Add("/d", "Day 2 task")
	s.Check("/d", i2.ID)

	// Same day 2 call again: should NOT clean
	s.CleanIfNewDay("/d")

	items, _ := s.List("/d")
	if len(items) != 1 {
		t.Fatalf("expected 1 item (same day should not re-clean), got %d", len(items))
	}
	if items[0].Text != "Day 2 task" {
		t.Errorf("expected remaining item 'Day 2 task', got %q", items[0].Text)
	}
}

func TestDeleteLastItemRemovesSection(t *testing.T) {
	s := openTestStore(t)

	item, _ := s.Add("/d", "Only item")
	s.Delete("/d", item.ID)

	// Directory should be gone from the file
	content, _ := os.ReadFile(s.filePath)
	if strings.Contains(string(content), "## /d") {
		t.Error("expected directory section to be removed from file")
	}

	// List should return empty
	items, _ := s.List("/d")
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestCleanRemovesEmptySection(t *testing.T) {
	s := openTestStore(t)

	i1, _ := s.Add("/d", "Task 1")
	i2, _ := s.Add("/d", "Task 2")
	s.Check("/d", i1.ID)
	s.Check("/d", i2.ID)

	s.Clean("/d")

	// Directory should be gone from the file
	content, _ := os.ReadFile(s.filePath)
	if strings.Contains(string(content), "## /d") {
		t.Error("expected directory section to be removed from file")
	}
}

func TestCleanIfNewDayNoItemsNoSection(t *testing.T) {
	s := openTestStore(t)
	s.now = func() time.Time { return makeTime(2025, 3, 8) }

	// CleanIfNewDay on a directory with no items should not create a section
	err := s.CleanIfNewDay("/d")
	if err != nil {
		t.Fatalf("CleanIfNewDay: %v", err)
	}

	// File should not exist (no writes happened)
	if _, err := os.Stat(s.filePath); !os.IsNotExist(err) {
		t.Error("expected no file to be created")
	}
}

func TestCleanIfNewDayRemovesEmptySection(t *testing.T) {
	s := openTestStore(t)
	day1 := makeTime(2025, 3, 8)
	day2 := makeTime(2025, 3, 9)
	s.now = func() time.Time { return day1 }

	// Add a checked item and seed the date
	item, _ := s.Add("/d", "Remove me")
	s.Check("/d", item.ID)
	s.CleanIfNewDay("/d")

	// Day 2: clean removes the only item
	s.now = func() time.Time { return day2 }
	s.CleanIfNewDay("/d")

	// Directory should be gone from the file
	content, _ := os.ReadFile(s.filePath)
	if strings.Contains(string(content), "## /d") {
		t.Error("expected directory section to be removed from file")
	}
}
