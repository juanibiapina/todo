package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

// DefaultFilePath returns the path to the todo Markdown file.
//
// Resolution order:
//  1. TODO_FILE environment variable
//  2. "file" key in config file at <xdg.ConfigHome>/todo/config.toml
//  3. Default: <xdg.DataHome>/todo/todo.md
func DefaultFilePath() string {
	if v := os.Getenv("TODO_FILE"); v != "" {
		return expandPath(v)
	}

	configPath := filepath.Join(xdg.ConfigHome, "todo", "config.toml")
	if v := readConfigFileKey(configPath); v != "" {
		return expandPath(v)
	}

	return filepath.Join(xdg.DataHome, "todo", "todo.md")
}

// readConfigFileKey reads the "file" key from a simple TOML config file.
func readConfigFileKey(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "file") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"")
			return value
		}
	}
	return ""
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
