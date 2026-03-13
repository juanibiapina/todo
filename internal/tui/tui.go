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
	store     *store.Store
	directory string
	items     []*store.Item
	cursor    int
	mode      mode
	input     textinput.Model
	width     int
	height    int
	err       error
}

func New(s *store.Store, directory string) Model {
	ti := textinput.New()
	ti.Placeholder = "New item..."
	ti.CharLimit = 256

	return Model{
		store:     s,
		directory: directory,
		input:     ti,
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
		// 2 for "  " prefix, 2 for "> " prompt
		m.input.Width = max(0, m.width-4)
		return m, nil

	case itemsLoadedMsg:
		m.items = msg
		if m.cursor >= len(m.items) {
			m.cursor = max(0, len(m.items)-1)
		}
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil

	case tea.KeyMsg:
		if m.mode == modeAdd || m.mode == modeEdit {
			return m.updateInput(msg)
		}
		return m.updateNormal(msg)
	}

	return m, nil
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
			clipboard.WriteAll(m.items[m.cursor].Text)
		}

	case "a":
		m.mode = modeAdd
		m.input.Reset()
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
			m.input.SetValue(m.items[m.cursor].Text)
			m.input.CursorEnd()
			m.input.Focus()
			return m, m.input.Cursor.BlinkCmd()
		}

	case "d":
		if len(m.items) > 0 {
			item := m.items[m.cursor]
			if item.IsSection {
				m.store.Delete(m.directory, item.ID)
			} else if item.Checked {
				m.store.Delete(m.directory, item.ID)
			} else {
				m.store.Check(m.directory, item.ID)
			}
			return m, m.loadItems
		}

	case "c":
		m.store.Clean(m.directory)
		return m, m.loadItems
	}

	return m, nil
}

func (m Model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

	var b strings.Builder

	b.WriteString(titleStyle.Render("Todo"))
	b.WriteString("\n\n")

	if len(m.items) == 0 && m.mode != modeAdd {
		b.WriteString("  No items. Press 'a' to add one.\n")
	}

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}

		if item.IsSection {
			rule := "──────────"
			if m.width > 0 {
				ruleWidth := max(0, m.width-lipgloss.Width(cursor))
				rule = strings.Repeat("─", ruleWidth)
			}
			b.WriteString(cursor + sectionStyle.Render(rule) + "\n")
			continue
		}

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
					b.WriteString(prefix + styledLine + "\n")
				} else {
					b.WriteString(indent + styledLine + "\n")
				}
			}
		} else {
			text := textStyle.Render(item.Text)
			b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, check, text))
		}
	}

	if m.mode == modeAdd || m.mode == modeEdit {
		b.WriteString("\n")
		b.WriteString("  " + m.input.View())
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.mode == modeAdd || m.mode == modeEdit {
		b.WriteString(helpStyle.Render("  enter: save • esc: cancel"))
	} else {
		b.WriteString(helpStyle.Render("  j/k: move • J/K: reorder • enter: toggle • space: copy • a: add • s: section • e: edit • d: done/delete • c: clean • esc/q: quit"))
	}
	b.WriteString("\n")

	return b.String()
}
