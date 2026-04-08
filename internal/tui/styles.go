package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	accentColor  = lipgloss.Color("205") // pink
	dimColor     = lipgloss.Color("240") // gray
	successColor = lipgloss.Color("82")  // green
	warnColor    = lipgloss.Color("214") // orange

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dimStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	tagStyle = lipgloss.NewStyle().
			Foreground(warnColor)

	previewTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(successColor).
				MarginBottom(1)

	previewStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimColor).
			Padding(1, 2)

	searchStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	statusStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			MarginTop(1)

	warnStyle = lipgloss.NewStyle().
			Foreground(warnColor).
			Bold(true)
)
