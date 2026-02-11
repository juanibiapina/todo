package tui

import "github.com/charmbracelet/lipgloss"

// Terminal theme colors (ANSI 0-15) — adapts to user's terminal scheme
var (
	colorBlack       = lipgloss.Color("0")
	colorRed         = lipgloss.Color("1")
	colorGreen       = lipgloss.Color("2")
	colorYellow      = lipgloss.Color("3")
	colorBlue        = lipgloss.Color("4")
	colorMagenta     = lipgloss.Color("5")
	colorCyan        = lipgloss.Color("6")
	colorWhite       = lipgloss.Color("7")
	colorBrightBlack = lipgloss.Color("8")
	colorBrightWhite = lipgloss.Color("15")

	// Semantic aliases
	primaryColor = colorYellow
	successColor = colorGreen
	dangerColor  = colorRed
	mutedColor   = colorBrightBlack
	fgColor      = colorWhite

	// Selection background
	selectionBg = colorBrightBlack

	// Status bar
	statusBarStyle = lipgloss.NewStyle().Foreground(fgColor)

	// Help keys
	helpKeyStyle  = lipgloss.NewStyle().Foreground(primaryColor)
	helpDescStyle = lipgloss.NewStyle().Foreground(fgColor)

	// Messages
	errorStyle   = lipgloss.NewStyle().Foreground(dangerColor).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(successColor).Bold(true)
	mutedStyle   = lipgloss.NewStyle().Foreground(mutedColor)

	// Dialog / modal
	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	dialogTitleStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	// Ticket list — normal
	ticketNewStyle = lipgloss.NewStyle().Foreground(mutedColor)
	ticketRefStyle = lipgloss.NewStyle().Foreground(colorCyan)
	ticketPlanStyle = lipgloss.NewStyle().Foreground(successColor)
	ticketIDStyle  = lipgloss.NewStyle().Foreground(colorMagenta)
	ticketTitleStyle = lipgloss.NewStyle()

	// Ticket list — selected (with background)
	selectedBgStyle         = lipgloss.NewStyle().Background(selectionBg)
	ticketNewSelStyle       = lipgloss.NewStyle().Foreground(mutedColor).Background(selectionBg)
	ticketRefSelStyle       = lipgloss.NewStyle().Foreground(colorCyan).Background(selectionBg)
	ticketPlanSelStyle      = lipgloss.NewStyle().Foreground(successColor).Background(selectionBg)
	ticketIDSelStyle        = lipgloss.NewStyle().Foreground(colorMagenta).Background(selectionBg)
	ticketTitleSelStyle     = lipgloss.NewStyle().Background(selectionBg)
)
