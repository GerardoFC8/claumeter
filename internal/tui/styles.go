package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary  = lipgloss.Color("#7C3AED")
	colorAccent   = lipgloss.Color("#22D3EE")
	colorMuted    = lipgloss.Color("#6B7280")
	colorGood     = lipgloss.Color("#10B981")
	colorWarn     = lipgloss.Color("#F59E0B")
	colorFg       = lipgloss.Color("#E5E7EB")
	colorFgDim    = lipgloss.Color("#9CA3AF")
	colorSelected = lipgloss.Color("#1F2937")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Padding(0, 1)

	tabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorFgDim)

	tabActiveStyle = tabStyle.
			Foreground(colorAccent).
			Bold(true).
			Underline(true)

	headerBarStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder(), false, false, true, false).
			BorderForeground(colorMuted)

	footerStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1)

	sectionStyle = lipgloss.NewStyle().
			Padding(1, 2)

	cardStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Margin(0, 1, 0, 0)

	cardLabelStyle = lipgloss.NewStyle().Foreground(colorFgDim)
	cardValueStyle = lipgloss.NewStyle().Foreground(colorFg).Bold(true)

	accentStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	goodStyle   = lipgloss.NewStyle().Foreground(colorGood).Bold(true)
	warnStyle   = lipgloss.NewStyle().Foreground(colorWarn).Bold(true)
)
