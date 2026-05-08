// Package render contains shared formatting helpers for displaying todo
// items across the CLI list output and the TUI.
package render

import (
	"strings"
	"unicode/utf8"
)

// SectionParts returns the three pieces of a section divider line so callers
// can style each piece independently (e.g. the title in a different colour
// from the dashes).
//
//   - lead: the leading dashes (and trailing space, when titled)
//   - mid:  the title text (empty when there is no title)
//   - tail: the trailing dashes (and leading space, when titled)
//
// Concatenating lead+mid+tail equals SectionLine(title, width).
//
// If the title is wider than the available width, the trailing piece is
// empty and the line may exceed `width`.
func SectionParts(title string, width int) (lead, mid, tail string) {
	if width < 0 {
		width = 0
	}
	if title == "" {
		return strings.Repeat("─", width), "", ""
	}
	lead = "── "
	mid = title
	prefix := lead + mid + " "
	prefixWidth := utf8.RuneCountInString(prefix)
	if prefixWidth >= width {
		tail = " "
		return lead, mid, tail
	}
	tail = " " + strings.Repeat("─", width-prefixWidth)
	return lead, mid, tail
}

// SectionLine formats a section divider for display at the given terminal
// width. With an empty title it returns a row of box-drawing dashes. With a
// title it returns "── <title> " followed by dashes filling the rest of the
// width. If the title is wider than the available space, the trailing
// dashes are omitted and the line may exceed `width`.
func SectionLine(title string, width int) string {
	lead, mid, tail := SectionParts(title, width)
	return lead + mid + tail
}
