package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/juanibiapina/todo/internal/tickets"
)

// Panel focus
type panel int

const (
	panelList panel = iota
	panelDetail
)

// Modal mode
type modalMode int

const (
	modalNone modalMode = iota
	modalAdd
	modalHelp
)

// tickMsg refreshes ticket data from disk
type tickMsg time.Time

// Model is the main TUI model
type Model struct {
	dir    string
	items  []*tickets.Ticket
	scroll ScrollState

	activePanel panel
	modal       modalMode
	width       int
	height      int
	ready       bool

	message     string
	messageTime time.Time
	isError     bool

	textInput  textinput.Model
	detailView viewport.Model
}

// New creates a new TUI model for the given directory.
func New(dir string) Model {
	ti := textinput.New()
	ti.Placeholder = "Ticket title..."
	ti.CharLimit = 256
	ti.Width = 50

	return Model{
		dir:         dir,
		activePanel: panelList,
		modal:       modalNone,
		textInput:   ti,
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) loadTickets() tea.Cmd {
	return func() tea.Msg {
		items, _ := tickets.List(m.dir)
		return ticketsLoadedMsg{items: items}
	}
}

type ticketsLoadedMsg struct {
	items []*tickets.Ticket
}

type actionDoneMsg struct {
	message string
	isError bool
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadTickets(), tickCmd())
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Calculate layout
		listH := m.height - 2 // header + status bar
		m.scroll.VisibleRows = listH - 2
		if m.scroll.VisibleRows < 1 {
			m.scroll.VisibleRows = 1
		}

		detailW := m.width - m.listPanelWidth() - 4
		detailH := listH - 3
		if detailW < 10 {
			detailW = 10
		}
		if detailH < 1 {
			detailH = 1
		}
		m.detailView = viewport.New(detailW, detailH)

	case tickMsg:
		return m, tea.Batch(m.loadTickets(), tickCmd())

	case ticketsLoadedMsg:
		m.items = msg.items
		m.scroll.ClampToCount(len(m.items))
		m.updateDetailContent()

	case actionDoneMsg:
		m.message = msg.message
		m.isError = msg.isError
		m.messageTime = time.Now()
		return m, m.loadTickets()

	case moveResultMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("Error: %v", msg.err)
			m.isError = true
			m.messageTime = time.Now()
		} else {
			// Move cursor to follow the ticket
			if msg.direction == "up" && m.scroll.Cursor > 0 {
				m.scroll.Up()
			} else if msg.direction == "down" && m.scroll.Cursor < len(m.items)-1 {
				m.scroll.Down(len(m.items))
			}
		}
		return m, m.loadTickets()

	case tea.KeyMsg:
		if time.Since(m.messageTime) > 3*time.Second {
			m.message = ""
		}

		if m.modal != modalNone {
			return m.updateModal(msg)
		}
		return m.updateMain(msg)
	}

	return m, nil
}

func (m *Model) updateDetailContent() {
	if len(m.items) == 0 || m.scroll.Cursor >= len(m.items) {
		m.detailView.SetContent(mutedStyle.Render("No ticket selected"))
		return
	}

	t := m.items[m.scroll.Cursor]
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	b.WriteString(titleStyle.Render(t.Title))
	b.WriteString("\n\n")

	// Metadata
	b.WriteString(mutedStyle.Render("ID:    "))
	b.WriteString(ticketIDStyle.Render(t.ID))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("State: "))
	b.WriteString(m.stateStyled(t.State, false))
	b.WriteString("\n")

	if t.Description != "" {
		b.WriteString("\n")
		// Wrap description to fit panel width
		descWidth := m.detailView.Width
		if descWidth > 0 {
			b.WriteString(ansi.Wrap(t.Description, descWidth, ""))
		} else {
			b.WriteString(t.Description)
		}
	}

	m.detailView.SetContent(b.String())
	m.detailView.GotoTop()
}

func (m Model) updateModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.modal {
	case modalAdd:
		switch msg.String() {
		case "esc":
			m.modal = modalNone
			return m, nil
		case "enter":
			title := m.textInput.Value()
			if title != "" {
				m.modal = modalNone
				return m, m.addTicket(title)
			}
		case "ctrl+c":
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd

	case modalHelp:
		switch msg.String() {
		case "esc", "?", "q":
			m.modal = modalNone
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		return m, tea.Quit

	case "tab":
		if m.activePanel == panelList {
			m.activePanel = panelDetail
		} else {
			m.activePanel = panelList
		}

	case "?":
		m.modal = modalHelp
	}

	switch m.activePanel {
	case panelList:
		return m.updateListPanel(msg)
	case panelDetail:
		return m.updateDetailPanel(msg)
	}
	return m, nil
}

func (m Model) updateListPanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.scroll.Up() {
			m.updateDetailContent()
		}
	case "down", "j":
		if m.scroll.Down(len(m.items)) {
			m.updateDetailContent()
		}
	case "g":
		m.scroll.First()
		m.updateDetailContent()
	case "G":
		m.scroll.Last(len(m.items))
		m.updateDetailContent()

	case "a":
		m.modal = modalAdd
		m.textInput.Reset()
		m.textInput.Focus()
		return m, textinput.Blink

	case "d":
		if len(m.items) > 0 {
			return m, m.deleteTicket(m.items[m.scroll.Cursor].ID)
		}

	case "s":
		if len(m.items) > 0 {
			return m, m.cycleState(m.items[m.scroll.Cursor].ID)
		}
	case "S":
		if len(m.items) > 0 {
			return m, m.cycleStateBack(m.items[m.scroll.Cursor].ID)
		}

	case "K":
		if len(m.items) > 0 {
			return m, m.moveUp(m.items[m.scroll.Cursor].ID)
		}
	case "J":
		if len(m.items) > 0 {
			return m, m.moveDown(m.items[m.scroll.Cursor].ID)
		}
	}

	return m, nil
}

func (m Model) updateDetailPanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "up", "k":
		m.detailView.LineUp(1)
	case "down", "j":
		m.detailView.LineDown(1)
	case "g":
		m.detailView.GotoTop()
	case "G":
		m.detailView.GotoBottom()
	case "ctrl+u":
		m.detailView.HalfViewUp()
	case "ctrl+d":
		m.detailView.HalfViewDown()
	default:
		m.detailView, cmd = m.detailView.Update(msg)
	}
	return m, cmd
}

// Actions

func (m Model) addTicket(title string) tea.Cmd {
	return func() tea.Msg {
		t, err := tickets.Add(m.dir, title, "")
		if err != nil {
			return actionDoneMsg{message: fmt.Sprintf("Error: %v", err), isError: true}
		}
		return actionDoneMsg{message: fmt.Sprintf("Added %s: %s", t.ID, t.Title)}
	}
}

func (m Model) deleteTicket(ref string) tea.Cmd {
	return func() tea.Msg {
		title, err := tickets.Done(m.dir, ref)
		if err != nil {
			return actionDoneMsg{message: fmt.Sprintf("Error: %v", err), isError: true}
		}
		return actionDoneMsg{message: fmt.Sprintf("Completed: %s", title)}
	}
}

func (m Model) cycleState(ref string) tea.Cmd {
	return func() tea.Msg {
		title, state, err := tickets.CycleState(m.dir, ref)
		if err != nil {
			return actionDoneMsg{message: fmt.Sprintf("Error: %v", err), isError: true}
		}
		return actionDoneMsg{message: fmt.Sprintf("%s → %s", title, state)}
	}
}

func (m Model) cycleStateBack(ref string) tea.Cmd {
	return func() tea.Msg {
		title, state, err := tickets.CycleStateBack(m.dir, ref)
		if err != nil {
			return actionDoneMsg{message: fmt.Sprintf("Error: %v", err), isError: true}
		}
		return actionDoneMsg{message: fmt.Sprintf("%s → %s", title, state)}
	}
}

type moveResultMsg struct {
	direction string // "up" or "down"
	err       error
}

func (m Model) moveUp(ref string) tea.Cmd {
	return func() tea.Msg {
		_, err := tickets.MoveUp(m.dir, ref)
		return moveResultMsg{direction: "up", err: err}
	}
}

func (m Model) moveDown(ref string) tea.Cmd {
	return func() tea.Msg {
		_, err := tickets.MoveDown(m.dir, ref)
		return moveResultMsg{direction: "down", err: err}
	}
}

// View renders the UI.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	var s strings.Builder
	s.WriteString(m.renderPanels())
	s.WriteString("\n")
	s.WriteString(m.renderStatusBar())

	if m.modal != modalNone {
		return m.renderModal(s.String())
	}

	return s.String()
}

func (m Model) listPanelWidth() int {
	w := m.width * 40 / 100
	if w < 35 {
		w = 35
	}
	if w > 60 {
		w = 60
	}
	return w
}

func (m Model) renderPanels() string {
	leftW := m.listPanelWidth()
	rightW := m.width - leftW
	totalH := m.height - 1

	if totalH < 4 {
		totalH = 4
	}

	// Left: list panel
	listContent := m.renderTicketList(leftW - 4)
	listPanel := m.renderPanel(1, "Tickets", listContent, leftW, totalH, m.activePanel == panelList)

	// Right: detail panel
	detailTitle := "Details"
	if len(m.items) > 0 && m.scroll.Cursor < len(m.items) {
		detailTitle = fmt.Sprintf("Details: %s", m.items[m.scroll.Cursor].ID)
	}
	m.detailView.Width = rightW - 4
	m.detailView.Height = totalH - 3
	detailContent := m.detailView.View()
	detailPanel := m.renderPanel(2, detailTitle, detailContent, rightW, totalH, m.activePanel == panelDetail)

	return lipgloss.JoinHorizontal(lipgloss.Top, listPanel, detailPanel)
}

func (m Model) renderTicketList(width int) string {
	if len(m.items) == 0 {
		return mutedStyle.Render("No tickets. Press 'a' to add one.")
	}

	var lines []string
	start, end := m.scroll.VisibleRange(len(m.items))

	for i := start; i < end; i++ {
		t := m.items[i]
		isSelected := i == m.scroll.Cursor

		// State icon
		icon := m.stateStyled(t.State, isSelected)

		// ID
		var id string
		if isSelected {
			id = ticketIDSelStyle.Render(t.ID)
		} else {
			id = ticketIDStyle.Render(t.ID)
		}

		// Title (truncated)
		maxTitleLen := width - 10 // icon(2) + space + id(3) + space + margin
		if maxTitleLen < 5 {
			maxTitleLen = 5
		}
		titleStr := t.Title
		if len(titleStr) > maxTitleLen {
			titleStr = titleStr[:maxTitleLen-1] + "…"
		}

		var title string
		if isSelected {
			title = ticketTitleSelStyle.Render(titleStr)
		} else {
			title = ticketTitleStyle.Render(titleStr)
		}

		var line string
		if isSelected {
			sp := selectedBgStyle.Render(" ")
			line = sp + icon + sp + id + sp + title
			padding := width - lipgloss.Width(line)
			if padding > 0 {
				line = line + selectedBgStyle.Render(strings.Repeat(" ", padding))
			}
		} else {
			line = fmt.Sprintf(" %s %s %s", icon, id, title)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m Model) stateStyled(state tickets.State, selected bool) string {
	icon := tickets.StateIcon(state)
	if selected {
		switch state {
		case tickets.StateNew:
			return ticketNewSelStyle.Render(icon)
		case tickets.StateRefined:
			return ticketRefSelStyle.Render(icon)
		case tickets.StatePlanned:
			return ticketPlanSelStyle.Render(icon)
		default:
			return ticketNewSelStyle.Render(icon)
		}
	}
	switch state {
	case tickets.StateNew:
		return ticketNewStyle.Render(icon)
	case tickets.StateRefined:
		return ticketRefStyle.Render(icon)
	case tickets.StatePlanned:
		return ticketPlanStyle.Render(icon)
	default:
		return ticketNewStyle.Render(icon)
	}
}

func (m Model) renderPanel(num int, title, content string, width, height int, active bool) string {
	borderColor := colorBlue
	titleFg := colorBlue
	if active {
		borderColor = primaryColor
		titleFg = primaryColor
	}

	tl, tr, bl, br := "╭", "╮", "╰", "╯"
	h, v := "─", "│"

	numText := fmt.Sprintf("[%d]", num)
	styledNum := lipgloss.NewStyle().Foreground(titleFg).Bold(active).Render(numText)
	styledTitle := lipgloss.NewStyle().Foreground(titleFg).Bold(active).Render(title)
	styledDash := lipgloss.NewStyle().Foreground(borderColor).Render(h)

	numWidth := lipgloss.Width(numText)
	titleWidth := lipgloss.Width(title)

	topBorderRight := width - 2 - numWidth - 1 - titleWidth - 1
	if topBorderRight < 0 {
		topBorderRight = 0
	}
	topLine := lipgloss.NewStyle().Foreground(borderColor).Render(tl+h) +
		styledNum + styledDash + styledTitle +
		lipgloss.NewStyle().Foreground(borderColor).Render(strings.Repeat(h, topBorderRight)+tr)

	bottomLine := lipgloss.NewStyle().Foreground(borderColor).Render(bl + strings.Repeat(h, width-2) + br)
	vBorder := lipgloss.NewStyle().Foreground(borderColor).Render(v)

	contentWidth := width - 4
	contentHeight := height - 2

	contentLines := strings.Split(content, "\n")
	var paddedLines []string
	for i := 0; i < contentHeight; i++ {
		var line string
		if i < len(contentLines) {
			line = contentLines[i]
		}
		line = FitToWidth(line, contentWidth)
		paddedLines = append(paddedLines, vBorder+" "+line+" "+vBorder)
	}

	return topLine + "\n" + strings.Join(paddedLines, "\n") + "\n" + bottomLine
}

func (m Model) renderStatusBar() string {
	var content string

	if m.message != "" && time.Since(m.messageTime) < 3*time.Second {
		var styledMsg string
		if m.isError {
			styledMsg = errorStyle.Render(m.message)
		} else {
			styledMsg = successStyle.Render(m.message)
		}
		msgWidth := lipgloss.Width(styledMsg)
		gap := m.width - msgWidth - 2
		if gap < 0 {
			gap = 0
		}
		content = " " + styledMsg + strings.Repeat(" ", gap) + " "
	} else {
		var parts []string
		switch m.activePanel {
		case panelList:
			parts = append(parts,
				m.renderKey("↑↓", "navigate"),
				m.renderKey("a", "add"),
				m.renderKey("s/S", "state"),
				m.renderKey("d", "done"),
				m.renderKey("K/J", "reorder"),
				m.renderKey("tab", "detail"),
			)
		case panelDetail:
			parts = append(parts,
				m.renderKey("↑↓", "scroll"),
				m.renderKey("g/G", "top/bottom"),
				m.renderKey("tab", "list"),
			)
		}
		parts = append(parts, m.renderKey("?", "help"), m.renderKey("esc/q", "quit"))

		leftSide := strings.Join(parts, " ")
		leftWidth := lipgloss.Width(leftSide)
		gap := m.width - leftWidth - 2
		if gap < 0 {
			gap = 0
		}
		content = " " + leftSide + strings.Repeat(" ", gap) + " "
	}

	return statusBarStyle.Render(content)
}

func (m Model) renderKey(key, desc string) string {
	return helpKeyStyle.Render(key) + " " + helpDescStyle.Render(desc)
}

func (m Model) renderModal(background string) string {
	var content string

	switch m.modal {
	case modalAdd:
		content = m.renderAddModal()
	case modalHelp:
		content = m.renderHelpModal()
	}

	modalWidth := lipgloss.Width(content)
	modalHeight := lipgloss.Height(content)
	x := (m.width - modalWidth) / 2
	y := (m.height - modalHeight) / 2

	return placeOverlay(x, y, content, background)
}

func placeOverlay(x, y int, fg, bg string) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	for i, fgLine := range fgLines {
		bgY := y + i
		if bgY < 0 || bgY >= len(bgLines) {
			continue
		}

		bgLine := bgLines[bgY]
		bgLineWidth := ansi.StringWidth(bgLine)

		var newLine strings.Builder

		if x > 0 {
			left := ansi.Truncate(bgLine, x, "")
			newLine.WriteString(left)
			leftWidth := ansi.StringWidth(left)
			if leftWidth < x {
				newLine.WriteString(strings.Repeat(" ", x-leftWidth))
			}
		}

		newLine.WriteString(fgLine)
		fgLineWidth := ansi.StringWidth(fgLine)

		rightStart := x + fgLineWidth
		if rightStart < bgLineWidth {
			right := truncateLeft(bgLine, rightStart)
			newLine.WriteString(right)
		}

		bgLines[bgY] = newLine.String()
	}

	return strings.Join(bgLines, "\n")
}

func truncateLeft(s string, n int) string {
	if n <= 0 {
		return s
	}

	var result strings.Builder
	width := 0
	inEscape := false
	escapeSeq := strings.Builder{}

	for _, r := range s {
		if inEscape {
			escapeSeq.WriteRune(r)
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
				if width >= n {
					result.WriteString(escapeSeq.String())
				}
				escapeSeq.Reset()
			}
			continue
		}

		if r == '\x1b' {
			inEscape = true
			escapeSeq.WriteRune(r)
			continue
		}

		charWidth := 1
		if r > 127 {
			charWidth = ansi.StringWidth(string(r))
		}

		if width >= n {
			result.WriteRune(r)
		}
		width += charWidth
	}

	return result.String()
}

func (m Model) renderAddModal() string {
	title := dialogTitleStyle.Render("Add Ticket")
	input := m.textInput.View()
	help := helpDescStyle.Render("enter: add • esc: cancel")

	content := title + "\n\n" + input + "\n\n" + help
	return dialogStyle.Render(content)
}

func (m Model) renderHelpModal() string {
	title := dialogTitleStyle.Render("Keyboard Shortcuts")

	sections := []string{
		helpKeyStyle.Render("Ticket List"),
		"  " + m.renderKey("↑/k ↓/j", "move cursor"),
		"  " + m.renderKey("g/G", "first/last"),
		"  " + m.renderKey("a", "add ticket"),
		"  " + m.renderKey("d", "mark done (remove)"),
		"  " + m.renderKey("s/S", "cycle state forward/back"),
		"  " + m.renderKey("K/J", "reorder up/down"),
		"",
		helpKeyStyle.Render("Detail Panel"),
		"  " + m.renderKey("↑/k ↓/j", "scroll"),
		"  " + m.renderKey("g/G", "top/bottom"),
		"  " + m.renderKey("ctrl+u/d", "half page"),
		"",
		helpKeyStyle.Render("General"),
		"  " + m.renderKey("tab", "switch panel"),
		"  " + m.renderKey("?", "this help"),
		"  " + m.renderKey("esc/q", "quit"),
	}

	help := helpDescStyle.Render("\npress esc or ? to close")
	content := title + "\n\n" + strings.Join(sections, "\n") + help

	return dialogStyle.Width(45).Render(content)
}

// Start launches the TUI.
func Start(dir string) error {
	// Ensure tickets file exists
	if err := tickets.EnsureFile(tickets.FilePath(dir)); err != nil {
		return err
	}

	p := tea.NewProgram(New(dir), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
