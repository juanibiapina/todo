package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTodoFileEnvOverride(t *testing.T) {
	t.Setenv("TODO_FILE", "/custom/path/todo.md")
	got := DefaultFilePath()
	if got != "/custom/path/todo.md" {
		t.Errorf("expected /custom/path/todo.md, got %q", got)
	}
}

func TestTodoFileEnvExpandsEnvVars(t *testing.T) {
	t.Setenv("MY_DIR", "/my/custom/dir")
	t.Setenv("TODO_FILE", "$MY_DIR/todo.md")
	got := DefaultFilePath()
	if got != "/my/custom/dir/todo.md" {
		t.Errorf("expected /my/custom/dir/todo.md, got %q", got)
	}
}

func TestTodoFileEnvExpandsTilde(t *testing.T) {
	t.Setenv("TODO_FILE", "~/todo.md")
	got := DefaultFilePath()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "todo.md")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestDefaultPath(t *testing.T) {
	t.Setenv("TODO_FILE", "")
	got := DefaultFilePath()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".local", "share", "todo", "todo.md")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
