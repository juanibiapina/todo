package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// FitToWidth ensures a string is exactly the specified visual width.
func FitToWidth(s string, width int) string {
	currentWidth := lipgloss.Width(s)
	if currentWidth > width {
		return ansi.Truncate(s, width, "")
	}
	if currentWidth < width {
		return s + strings.Repeat(" ", width-currentWidth)
	}
	return s
}
