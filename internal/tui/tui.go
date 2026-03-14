package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/juanibiapina/todo/internal/store"
)

var (
	checkedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Strikethrough(true)
	uncheckedStyle = lipgloss.NewStyle()
	cursorStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	helpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	sectionStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
)

type mode int

const (
	modeNormal mode = iota
	modeAdd
	modeEdit
)

type Model struct {
	store        *store.Store
	directory    string
	items        []*store.Item
	cursor       int
	mode         mode
	input        textinput.Model
	width        int
	height       int
	scrollOffset int
	err          error
}

func New(s *store.Store, directory string) Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = "New item..."
	ti.CharLimit = 256

	return Model{
		store:     s,
		directory: directory,
		input:     ti,
	}
}

// headerLines is the number of lines used by the fixed header ("Todo\n\n").
const headerLines = 2

// footerLines is the number of lines used by the fixed footer ("\n" + help text + "\n").
const footerLines = 2

// itemLineCount returns the number of terminal lines an item at index i
// occupies, including the add-mode input line if applicable.
func (m Model) itemLineCount(i int) int {
	item := m.items[i]

	lines := 1
	if item.IsSection {
		// Sections are always one line.
	} else if m.mode == modeEdit && i == m.cursor {
		// Editing replaces the item with a single-line input.
	} else if m.width > 0 {
		// Calculate wrapped line count.
		prefixWidth := 6 // "  " (cursor) + "[ ] " (check + space)
		if prefixWidth < m.width {
			availableWidth := m.width - prefixWidth
			wrapped := lipgloss.NewStyle().Width(availableWidth).Render(item.Text)
			lines = len(strings.Split(wrapped, "\n"))
		}
	}

	// Add mode inserts an input line after the cursor item.
	if m.mode == modeAdd && i == m.cursor {
		lines++
	}

	return lines
}

// ensureCursorVisible adjusts scrollOffset so the cursor item is fully
// within the visible area of the viewport.
func ensureCursorVisible(m *Model) {
	if m.height == 0 || len(m.items) == 0 || m.cursor >= len(m.items) {
		return
	}

	availableHeight := m.height - headerLines - footerLines
	if availableHeight <= 0 {
		return
	}

	// Calculate line range of the cursor item.
	cursorStart := 0
	for i := 0; i < m.cursor; i++ {
		cursorStart += m.itemLineCount(i)
	}
	cursorEnd := cursorStart + m.itemLineCount(m.cursor)

	// Scroll up if cursor is above viewport.
	if cursorStart < m.scrollOffset {
		m.scrollOffset = cursorStart
	}
	// Scroll down if cursor is below viewport.
	if cursorEnd > m.scrollOffset+availableHeight {
		m.scrollOffset = cursorEnd - availableHeight
	}

	// Clamp so we don't leave blank space at the bottom when the viewport
	// grows (e.g. terminal resize) or items are deleted.
	totalLines := cursorEnd // lines up to and including cursor
	for i := m.cursor + 1; i < len(m.items); i++ {
		totalLines += m.itemLineCount(i)
	}
	if maxOffset := max(0, totalLines-availableHeight); m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadItems
}

type itemsLoadedMsg []*store.Item
type errMsg error

func (m Model) loadItems() tea.Msg {
	items, err := m.store.List(m.directory)
	if err != nil {
		return errMsg(err)
	}
	return itemsLoadedMsg(items)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// 2 for indent prefix, 4 for "[ ] "
		m.input.Width = max(0, m.width-6)
		ensureCursorVisible(&m)
		return m, nil

	case itemsLoadedMsg:
		m.items = msg
		if m.cursor >= len(m.items) {
			m.cursor = max(0, len(m.items)-1)
		}
		ensureCursorVisible(&m)
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil

	case tea.KeyMsg:
		var cmd tea.Cmd
		if m.mode == modeAdd || m.mode == modeEdit {
			m, cmd = m.updateInput(msg)
		} else {
			m, cmd = m.updateNormal(msg)
		}
		ensureCursorVisible(&m)
		return m, cmd
	}

	return m, nil
}

func (m Model) updateNormal(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}

	case "J":
		if m.cursor < len(m.items)-1 {
			item := m.items[m.cursor]
			nextItem := m.items[m.cursor+1]
			m.store.Swap(m.directory, item.ID, nextItem.ID)
			m.cursor++
			return m, m.loadItems
		}

	case "K":
		if m.cursor > 0 {
			item := m.items[m.cursor]
			prevItem := m.items[m.cursor-1]
			m.store.Swap(m.directory, item.ID, prevItem.ID)
			m.cursor--
			return m, m.loadItems
		}

	case "enter":
		if len(m.items) > 0 {
			item := m.items[m.cursor]
			if !item.IsSection {
				m.store.Toggle(m.directory, item.ID)
				return m, m.loadItems
			}
		}

	case " ":
		if len(m.items) > 0 {
			item := m.items[m.cursor]
			if !item.IsSection {
				m.store.Toggle(m.directory, item.ID)
				return m, m.loadItems
			}
		}

	case "a":
		m.mode = modeAdd
		m.input.Reset()
		m.input.Placeholder = "New item..."
		m.input.Focus()
		return m, m.input.Cursor.BlinkCmd()

	case "s":
		if len(m.items) > 0 {
			m.store.InsertSectionAfter(m.directory, m.items[m.cursor].ID)
			m.cursor++
		} else {
			m.store.AddSection(m.directory)
		}
		return m, m.loadItems

	case "e":
		if len(m.items) > 0 && !m.items[m.cursor].IsSection {
			m.mode = modeEdit
			m.input.Reset()
			m.input.Placeholder = "Edit item..."
			m.input.SetValue(m.items[m.cursor].Text)
			m.input.CursorEnd()
			m.input.Focus()
			return m, m.input.Cursor.BlinkCmd()
		}

	case "d":
		if len(m.items) > 0 {
			item := m.items[m.cursor]
			m.store.Delete(m.directory, item.ID)
			return m, m.loadItems
		}

	case "c":
		if len(m.items) > 0 {
			clipboard.WriteAll(m.items[m.cursor].Text)
		}

	case "C":
		m.store.Clean(m.directory)
		return m, m.loadItems
	}

	return m, nil
}

func (m Model) updateInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		text := strings.TrimSpace(m.input.Value())
		if m.mode == modeEdit {
			if text != "" && len(m.items) > 0 {
				m.store.Edit(m.directory, m.items[m.cursor].ID, text)
			}
		} else {
			if text != "" {
				if len(m.items) > 0 {
					m.store.InsertItemAfter(m.directory, text, m.items[m.cursor].ID)
					m.cursor++
				} else {
					m.store.Add(m.directory, text)
				}
			}
		}
		m.mode = modeNormal
		m.input.Blur()
		return m, m.loadItems

	case "esc":
		m.mode = modeNormal
		m.input.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	// Render all item lines into a slice.
	var itemLines []string

	if len(m.items) == 0 {
		if m.mode == modeAdd {
			itemLines = append(itemLines, cursorStyle.Render("> ")+"[ ] "+m.input.View())
		} else {
			itemLines = append(itemLines, "  No items. Press 'a' to add one.")
		}
	}

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor && m.mode != modeAdd {
			cursor = cursorStyle.Render("> ")
		}

		if item.IsSection {
			rule := "──────────"
			if m.width > 0 {
				ruleWidth := max(0, m.width-lipgloss.Width(cursor))
				rule = strings.Repeat("─", ruleWidth)
			}
			itemLines = append(itemLines, cursor+sectionStyle.Render(rule))
		} else if m.mode == modeEdit && i == m.cursor {
			check := "[ ]"
			if item.Checked {
				check = "[x]"
			}
			itemLines = append(itemLines, cursorStyle.Render("> ")+check+" "+m.input.View())
		} else {
			check := "[ ]"
			textStyle := uncheckedStyle
			if item.Checked {
				check = "[x]"
				textStyle = checkedStyle
			}

			prefix := fmt.Sprintf("%s%s ", cursor, check)
			prefixWidth := lipgloss.Width(prefix)

			if m.width > 0 && prefixWidth < m.width {
				availableWidth := m.width - prefixWidth
				wrapped := lipgloss.NewStyle().Width(availableWidth).Render(item.Text)
				lines := strings.Split(wrapped, "\n")
				indent := strings.Repeat(" ", prefixWidth)
				for j, line := range lines {
					styledLine := textStyle.Render(line)
					if j == 0 {
						itemLines = append(itemLines, prefix+styledLine)
					} else {
						itemLines = append(itemLines, indent+styledLine)
					}
				}
			} else {
				text := textStyle.Render(item.Text)
				itemLines = append(itemLines, fmt.Sprintf("%s%s %s", cursor, check, text))
			}
		}

		if m.mode == modeAdd && i == m.cursor {
			itemLines = append(itemLines, cursorStyle.Render("> ")+"[ ] "+m.input.View())
		}
	}

	// Clip to viewport.
	availableHeight := len(itemLines)
	if m.height > 0 {
		availableHeight = max(1, m.height-headerLines-footerLines)
	}

	visibleStart := m.scrollOffset
	if visibleStart > len(itemLines) {
		visibleStart = max(0, len(itemLines)-availableHeight)
	}
	visibleEnd := min(visibleStart+availableHeight, len(itemLines))
	visibleLines := itemLines[visibleStart:visibleEnd]

	// Build output: header + visible lines + footer.
	var b strings.Builder
	b.WriteString(titleStyle.Render("Todo"))
	b.WriteString("\n\n")
	for _, line := range visibleLines {
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")
	if m.mode == modeAdd || m.mode == modeEdit {
		b.WriteString(helpStyle.Render("  enter: save • esc: cancel"))
	} else {
		b.WriteString(helpStyle.Render("  j/k: move • J/K: reorder • space/enter: toggle • a: add • s: section • e: edit • d: delete • c: copy • C: clean • esc/q: quit"))
	}
	b.WriteString("\n")

	return b.String()
}
