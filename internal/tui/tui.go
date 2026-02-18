package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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

// View mode
type viewMode int

const (
	viewAll viewMode = iota
	viewReady
	viewBlocked
	viewClosed
)

// tickMsg refreshes ticket data from disk
type tickMsg time.Time

// Model is the main TUI model
type Model struct {
	dir        string
	items      []*tickets.Ticket
	allTickets []*tickets.Ticket
	scroll     ScrollState
	view       viewMode

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

	// Cached rendered markdown for the detail panel
	cachedDetailID    string // ticket ID of cached render
	cachedDetailDesc  string // raw description that was rendered
	cachedDetailWidth int    // width used for rendering
	cachedRendered    string // glamour-rendered output
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
		allItems, _ := tickets.List(m.dir)
		return ticketsLoadedMsg{allTickets: allItems}
	}
}

type ticketsLoadedMsg struct {
	allTickets []*tickets.Ticket
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
		m.allTickets = msg.allTickets
		m.applyView()
		m.updateDetailContent()

	case actionDoneMsg:
		m.message = msg.message
		m.isError = msg.isError
		m.messageTime = time.Now()
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

func (m *Model) applyView() {
	switch m.view {
	case viewReady:
		m.items = m.filterReady()
	case viewBlocked:
		m.items = m.filterBlocked()
	case viewClosed:
		m.items = m.filterClosed()
	default: // viewAll — open/in_progress (not closed)
		var items []*tickets.Ticket
		for _, t := range m.allTickets {
			if t.Status != "closed" {
				items = append(items, t)
			}
		}
		m.items = items
	}
	m.scroll.ClampToCount(len(m.items))
}

func (m *Model) filterReady() []*tickets.Ticket {
	statusMap := make(map[string]string)
	for _, t := range m.allTickets {
		statusMap[t.ID] = t.Status
	}

	var ready []*tickets.Ticket
	for _, t := range m.allTickets {
		if t.Status == "closed" {
			continue
		}
		allDepsDone := true
		for _, depID := range t.Deps {
			depStatus, exists := statusMap[depID]
			if exists && depStatus != "closed" {
				allDepsDone = false
				break
			}
		}
		if !allDepsDone {
			continue
		}
		ready = append(ready, t)
	}

	sort.Slice(ready, func(i, j int) bool {
		if ready[i].Priority != ready[j].Priority {
			return ready[i].Priority < ready[j].Priority
		}
		return ready[i].ID < ready[j].ID
	})
	return ready
}

func (m *Model) filterBlocked() []*tickets.Ticket {
	statusMap := make(map[string]string)
	for _, t := range m.allTickets {
		statusMap[t.ID] = t.Status
	}

	var blocked []*tickets.Ticket
	for _, t := range m.allTickets {
		if t.Status == "closed" {
			continue
		}
		hasUnclosed := false
		for _, depID := range t.Deps {
			depStatus, exists := statusMap[depID]
			if exists && depStatus != "closed" {
				hasUnclosed = true
				break
			}
		}
		if !hasUnclosed {
			continue
		}
		blocked = append(blocked, t)
	}

	sort.Slice(blocked, func(i, j int) bool {
		if blocked[i].Priority != blocked[j].Priority {
			return blocked[i].Priority < blocked[j].Priority
		}
		return blocked[i].ID < blocked[j].ID
	})
	return blocked
}

func (m *Model) filterClosed() []*tickets.Ticket {
	type closedTicket struct {
		ticket *tickets.Ticket
		mtime  int64
	}

	var closed []closedTicket
	for _, t := range m.allTickets {
		if t.Status != "closed" {
			continue
		}
		path := filepath.Join(tickets.DirPath(m.dir), t.ID+".md")
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		closed = append(closed, closedTicket{ticket: t, mtime: info.ModTime().UnixNano()})
	}

	sort.Slice(closed, func(i, j int) bool {
		return closed[i].mtime > closed[j].mtime
	})

	items := make([]*tickets.Ticket, len(closed))
	for i, ct := range closed {
		items[i] = ct.ticket
	}
	return items
}

func (m *Model) renderMarkdown(text string, width int) string {
	if width <= 0 {
		width = 80
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return ansi.Wrap(text, width, "")
	}
	rendered, err := r.Render(text)
	if err != nil {
		return ansi.Wrap(text, width, "")
	}
	return strings.TrimRight(rendered, "\n")
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

	// ID
	b.WriteString(metaLabelStyle.Render("ID: "))
	b.WriteString(ticketIDStyle.Render(t.ID))
	b.WriteString("\n")

	// Status — always shown, default to "open"
	status := t.Status
	if status == "" {
		status = "open"
	}
	b.WriteString(metaLabelStyle.Render("Status: "))
	b.WriteString(metaValueStyle.Render(status))
	b.WriteString("\n")

	// Type
	if t.Type != "" {
		b.WriteString(metaLabelStyle.Render("Type: "))
		b.WriteString(metaValueStyle.Render(t.Type))
		b.WriteString("\n")
	}

	// Priority — always shown as P<n>
	b.WriteString(metaLabelStyle.Render("Priority: "))
	b.WriteString(metaValueStyle.Render(fmt.Sprintf("P%d", t.Priority)))
	b.WriteString("\n")

	// Assignee
	if t.Assignee != "" {
		b.WriteString(metaLabelStyle.Render("Assignee: "))
		b.WriteString(metaValueStyle.Render(t.Assignee))
		b.WriteString("\n")
	}

	// Created
	if t.Created != "" {
		b.WriteString(metaLabelStyle.Render("Created: "))
		b.WriteString(metaValueStyle.Render(t.Created))
		b.WriteString("\n")
	}

	// Parent — enhanced with resolved title if available
	if t.Parent != "" {
		b.WriteString(metaLabelStyle.Render("Parent: "))
		rel := tickets.ComputeRelations(t, m.allTickets)
		if rel.ParentTicket != nil {
			b.WriteString(ticketIDStyle.Render(rel.ParentTicket.ID))
			b.WriteString(metaValueStyle.Render(" ("+rel.ParentTicket.Title+")"))
		} else {
			b.WriteString(metaValueStyle.Render(t.Parent))
		}
		b.WriteString("\n")
	}

	// ExternalRef
	if t.ExternalRef != "" {
		b.WriteString(metaLabelStyle.Render("Ref: "))
		b.WriteString(metaValueStyle.Render(t.ExternalRef))
		b.WriteString("\n")
	}

	// Tags
	if len(t.Tags) > 0 {
		b.WriteString(metaLabelStyle.Render("Tags: "))
		b.WriteString(metaValueStyle.Render(strings.Join(t.Tags, ", ")))
		b.WriteString("\n")
	}

	// Deps
	if len(t.Deps) > 0 {
		b.WriteString(metaLabelStyle.Render("Deps: "))
		b.WriteString(metaValueStyle.Render(strings.Join(t.Deps, ", ")))
		b.WriteString("\n")
	}

	// Links
	if len(t.Links) > 0 {
		b.WriteString(metaLabelStyle.Render("Links: "))
		b.WriteString(metaValueStyle.Render(strings.Join(t.Links, ", ")))
		b.WriteString("\n")
	}

	descWidth := m.detailView.Width

	// Design section
	if t.Design != "" {
		b.WriteString("\n")
		b.WriteString(sectionHeadingStyle.Render("Design"))
		b.WriteString("\n")
		b.WriteString(m.renderMarkdown(t.Design, descWidth))
		b.WriteString("\n")
	}

	// Acceptance section
	if t.Acceptance != "" {
		b.WriteString("\n")
		b.WriteString(sectionHeadingStyle.Render("Acceptance"))
		b.WriteString("\n")
		b.WriteString(m.renderMarkdown(t.Acceptance, descWidth))
		b.WriteString("\n")
	}

	// Description section
	if t.Description != "" {
		b.WriteString("\n")
		b.WriteString(sectionHeadingStyle.Render("Description"))
		b.WriteString("\n")

		// Use cached render if ticket/description/width haven't changed
		if t.ID == m.cachedDetailID && t.Description == m.cachedDetailDesc && descWidth == m.cachedDetailWidth {
			b.WriteString(m.cachedRendered)
		} else {
			rendered := m.renderMarkdown(t.Description, descWidth)
			m.cachedDetailID = t.ID
			m.cachedDetailDesc = t.Description
			m.cachedDetailWidth = descWidth
			m.cachedRendered = rendered
			b.WriteString(rendered)
		}
	}

	// Computed relationships
	rel := tickets.ComputeRelations(t, m.allTickets)
	m.renderRelationSection(&b, "Blockers", rel.Blockers)
	m.renderRelationSection(&b, "Blocking", rel.Blocking)
	m.renderRelationSection(&b, "Children", rel.Children)
	m.renderRelationSection(&b, "Linked", rel.Linked)

	m.detailView.SetContent(b.String())
	m.detailView.GotoTop()
}

func (m *Model) renderRelationSection(b *strings.Builder, heading string, items []*tickets.Ticket) {
	if len(items) == 0 {
		return
	}
	b.WriteString("\n")
	b.WriteString(sectionHeadingStyle.Render(heading))
	b.WriteString("\n")
	for _, t := range items {
		b.WriteString(m.renderRelationLine(t))
		b.WriteString("\n")
	}
}

func (m *Model) renderRelationLine(t *tickets.Ticket) string {
	status := t.Status
	if status == "" {
		status = "open"
	}
	return "  " + ticketIDStyle.Render(t.ID) + " " +
		mutedStyle.Render("["+status+"]") + " " +
		metaValueStyle.Render(t.Title)
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

	case "d", "c":
		if len(m.items) > 0 {
			return m, m.closeTicket(m.items[m.scroll.Cursor].ID)
		}

	case "s":
		if len(m.items) > 0 {
			return m, m.startTicket(m.items[m.scroll.Cursor].ID)
		}

	case "r":
		if len(m.items) > 0 {
			return m, m.reopenTicket(m.items[m.scroll.Cursor].ID)
		}

	case " ":
		if len(m.items) > 0 {
			return m, m.copyTicket(m.items[m.scroll.Cursor])
		}

	case "1":
		m.view = viewAll
		m.applyView()
		m.scroll.Reset()
		m.updateDetailContent()
	case "2":
		m.view = viewReady
		m.applyView()
		m.scroll.Reset()
		m.updateDetailContent()
	case "3":
		m.view = viewBlocked
		m.applyView()
		m.scroll.Reset()
		m.updateDetailContent()
	case "4":
		m.view = viewClosed
		m.applyView()
		m.scroll.Reset()
		m.updateDetailContent()
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
		t, err := tickets.Add(m.dir, &tickets.Ticket{Title: title})
		if err != nil {
			return actionDoneMsg{message: fmt.Sprintf("Error: %v", err), isError: true}
		}
		return actionDoneMsg{message: fmt.Sprintf("Added %s: %s", t.ID, t.Title)}
	}
}

func (m Model) startTicket(id string) tea.Cmd {
	return func() tea.Msg {
		title, err := tickets.SetStatus(m.dir, id, "in_progress")
		if err != nil {
			return actionDoneMsg{message: fmt.Sprintf("Error: %v", err), isError: true}
		}
		return actionDoneMsg{message: fmt.Sprintf("Started: %s", title)}
	}
}

func (m Model) closeTicket(id string) tea.Cmd {
	return func() tea.Msg {
		title, err := tickets.SetStatus(m.dir, id, "closed")
		if err != nil {
			return actionDoneMsg{message: fmt.Sprintf("Error: %v", err), isError: true}
		}
		return actionDoneMsg{message: fmt.Sprintf("Closed: %s", title)}
	}
}

func (m Model) reopenTicket(id string) tea.Cmd {
	return func() tea.Msg {
		title, err := tickets.SetStatus(m.dir, id, "open")
		if err != nil {
			return actionDoneMsg{message: fmt.Sprintf("Error: %v", err), isError: true}
		}
		return actionDoneMsg{message: fmt.Sprintf("Reopened: %s", title)}
	}
}



func (m Model) copyTicket(t *tickets.Ticket) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(t.FullString())
		if err != nil {
			return actionDoneMsg{message: fmt.Sprintf("Error: %v", err), isError: true}
		}
		return actionDoneMsg{message: fmt.Sprintf("Copied %s to clipboard", t.ID)}
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
	var listTitle string
	switch m.view {
	case viewReady:
		listTitle = "Tickets [Ready]"
	case viewBlocked:
		listTitle = "Tickets [Blocked]"
	case viewClosed:
		listTitle = "Tickets [Closed]"
	default:
		listTitle = "Tickets [All]"
	}
	listContent := m.renderTicketList(leftW - 4)
	listPanel := m.renderPanel(1, listTitle, listContent, leftW, totalH, m.activePanel == panelList)

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

func (m Model) priorityBadge(priority int, selected bool) string {
	text := fmt.Sprintf("[P%d]", priority)
	var style lipgloss.Style
	switch {
	case priority <= 1:
		style = priorityHighStyle
	case priority == 2:
		style = priorityMedStyle
	default:
		style = priorityLowStyle
	}
	if selected {
		style = style.Background(selectionBg)
	}
	return style.Render(text)
}

func (m Model) statusBadge(status string, selected bool) string {
	if status == "" {
		status = "open"
	}
	text := fmt.Sprintf("[%s]", status)
	var style lipgloss.Style
	switch status {
	case "in_progress":
		style = statusActiveStyle
	case "closed":
		style = statusClosedStyle
	default:
		style = statusDefaultStyle
	}
	if selected {
		style = style.Background(selectionBg)
	}
	return style.Render(text)
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

		// ID
		var id string
		if isSelected {
			id = ticketIDSelStyle.Render(t.ID)
		} else {
			id = ticketIDStyle.Render(t.ID)
		}

		// Priority and status badges
		prioBadge := m.priorityBadge(t.Priority, isSelected)
		statBadge := m.statusBadge(t.Status, isSelected)

		// Title (truncated)
		// Layout: SP + ID(3) + SP + [P<n>](4) + [status](2+len) + SP + Title
		status := t.Status
		if status == "" {
			status = "open"
		}
		prefixW := 1 + 3 + 1 + 4 + (2 + len(status)) + 1
		maxTitleLen := width - prefixW
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
			line = sp + id + sp + prioBadge + statBadge + sp + title
			padding := width - lipgloss.Width(line)
			if padding > 0 {
				line = line + selectedBgStyle.Render(strings.Repeat(" ", padding))
			}
		} else {
			line = " " + id + " " + prioBadge + statBadge + " " + title
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
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
				m.renderKey("1-4", "views"),
				m.renderKey("a", "add"),
				m.renderKey("s", "start"),
				m.renderKey("c", "close"),
				m.renderKey("r", "reopen"),
				m.renderKey("space", "copy"),
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
		"  " + m.renderKey("s", "start (in_progress)"),
		"  " + m.renderKey("c/d", "close"),
		"  " + m.renderKey("r", "reopen"),
		"  " + m.renderKey("space", "copy ticket to clipboard"),
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
	p := tea.NewProgram(New(dir), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
