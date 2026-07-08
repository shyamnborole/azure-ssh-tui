package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorAzureBlue = lipgloss.Color("#0078D4")
	ColorSuccess   = lipgloss.Color("#107C10")
	ColorWarning   = lipgloss.Color("#D83B01")
	ColorError     = lipgloss.Color("#A4262C")
	ColorText      = lipgloss.Color("#F3F2F1")
	ColorSubText   = lipgloss.Color("#A19F9D")
	ColorHighlight = lipgloss.Color("#323130")

	// Base Styles
	BaseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorAzureBlue).
			Padding(0, 1).
			Bold(true)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorAzureBlue)
)
