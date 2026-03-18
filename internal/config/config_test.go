package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
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

func TestConfigFileKey(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "todo", "config.toml")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	os.WriteFile(configPath, []byte(`file = "/my/todo.md"`+"\n"), 0644)

	t.Setenv("TODO_FILE", "")
	t.Setenv("XDG_CONFIG_HOME", dir)
	// Force xdg to re-evaluate by setting the env before calling
	xdg.Reload()
	defer xdg.Reload()

	got := DefaultFilePath()
	if got != "/my/todo.md" {
		t.Errorf("expected /my/todo.md, got %q", got)
	}
}

func TestConfigFileExpandsEnvVars(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "todo", "config.toml")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	os.WriteFile(configPath, []byte(`file = "$MY_DATA/todo.md"`+"\n"), 0644)

	t.Setenv("TODO_FILE", "")
	t.Setenv("MY_DATA", "/expanded/data")
	t.Setenv("XDG_CONFIG_HOME", dir)
	xdg.Reload()
	defer xdg.Reload()

	got := DefaultFilePath()
	if got != "/expanded/data/todo.md" {
		t.Errorf("expected /expanded/data/todo.md, got %q", got)
	}
}

func TestConfigFileExpandsTilde(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "todo", "config.toml")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	os.WriteFile(configPath, []byte(`file = "~/todo.md"`+"\n"), 0644)

	t.Setenv("TODO_FILE", "")
	t.Setenv("XDG_CONFIG_HOME", dir)
	xdg.Reload()
	defer xdg.Reload()

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "todo.md")
	got := DefaultFilePath()
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestConfigFileIgnoresComments(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "todo", "config.toml")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	os.WriteFile(configPath, []byte("# file = \"/wrong/path.md\"\nfile = \"/right/path.md\"\n"), 0644)

	t.Setenv("TODO_FILE", "")
	t.Setenv("XDG_CONFIG_HOME", dir)
	xdg.Reload()
	defer xdg.Reload()

	got := DefaultFilePath()
	if got != "/right/path.md" {
		t.Errorf("expected /right/path.md, got %q", got)
	}
}

func TestMissingConfigFileFallsToDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TODO_FILE", "")
	t.Setenv("XDG_CONFIG_HOME", dir) // no config.toml here
	t.Setenv("XDG_DATA_HOME", "/test/data")
	xdg.Reload()
	defer xdg.Reload()

	got := DefaultFilePath()
	expected := "/test/data/todo/todo.md"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestDefaultUsesXDGDataHome(t *testing.T) {
	t.Setenv("TODO_FILE", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // no config file
	t.Setenv("XDG_DATA_HOME", "/xdg/data")
	xdg.Reload()
	defer xdg.Reload()

	got := DefaultFilePath()
	expected := "/xdg/data/todo/todo.md"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
