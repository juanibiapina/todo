package config

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultFilePath returns the path to the todo Markdown file.
//
// Resolution order:
//  1. TODO_FILE environment variable (supports ~ and $VAR expansion)
//  2. Default: ~/.local/share/todo/todo.md
func DefaultFilePath() string {
	if v := os.Getenv("TODO_FILE"); v != "" {
		return expandPath(v)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "todo.md"
	}
	return filepath.Join(home, ".local", "share", "todo", "todo.md")
}

// expandPath expands environment variables and a leading ~ in a path.
func expandPath(path string) string {
	path = os.ExpandEnv(path)
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	return path
}
