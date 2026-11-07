package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/juanibiapina/todo/internal/store"
)

func renderScreen(t *testing.T, width, height int, items []string) []string {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "todo.md"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	dir := "/tmp/project"
	for _, text := range items {
		if _, err := s.Add(dir, text); err != nil {
			t.Fatalf("Add: %v", err)
		}
	}

	var m tea.Model = New(s, dir)
	m, _ = m.Update(m.Init()())
	m, _ = m.Update(tea.WindowSizeMsg{Width: width, Height: height})

	return strings.Split(ansi.Strip(m.View()), "\n")
}

func TestViewKeepsMarginAroundContent(t *testing.T) {
	const width, height = 40, 10
	items := []string{"a long item that wraps past the right edge of the screen"}
	for i := range 20 {
		items = append(items, fmt.Sprintf("item %d", i))
	}

	lines := renderScreen(t, width, height, items)

	if len(lines) > height {
		t.Fatalf("got %d lines, want at most %d:\n%s", len(lines), height, strings.Join(lines, "\n"))
	}
	if strings.TrimSpace(lines[0]) != "" {
		t.Errorf("first line = %q, want blank", lines[0])
	}
	if !strings.HasPrefix(lines[1], "  Todo") {
		t.Errorf("title line = %q, want it to start with two spaces and Todo", lines[1])
	}
	for _, line := range lines {
		if strings.Contains(line, "j/k: move") {
			continue
		}
		if w := ansi.StringWidth(strings.TrimRight(line, " ")); w > width-framePaddingX {
			t.Errorf("line %q is %d wide, want at most %d", line, w, width-framePaddingX)
		}
	}
}
