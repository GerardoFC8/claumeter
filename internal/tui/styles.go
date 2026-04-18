package tui

import "github.com/charmbracelet/lipgloss"

// currentTheme is the active theme. All package-level style vars below are
// reassigned by applyTheme() whenever the user cycles themes.
// Initialised to dark (the original palette) to preserve existing behaviour.
var currentTheme *Theme = &themeDark

// Package-level color vars — referenced directly by overview.go and search.go.
// These are reassigned by applyTheme(); do NOT treat them as constants.
var (
	colorPrimary  = lipgloss.Color("#7C3AED")
	colorAccent   = lipgloss.Color("#22D3EE")
	colorMuted    = lipgloss.Color("#6B7280")
	colorGood     = lipgloss.Color("#10B981")
	colorWarn     = lipgloss.Color("#F59E0B")
	colorFg       = lipgloss.Color("#E5E7EB")
	colorFgDim    = lipgloss.Color("#9CA3AF")
	colorSelected = lipgloss.Color("#1F2937")
)

// Package-level derived style vars — call sites in all other tui files use
// these names directly; they are reassigned by applyTheme().
var (
	titleStyle     = themeDark.Title
	tabStyle       = themeDark.Tab
	tabActiveStyle = themeDark.TabActive
	headerBarStyle = themeDark.HeaderBar
	footerStyle    = themeDark.Footer
	sectionStyle   = themeDark.Section
	cardStyle      = themeDark.Card
	cardLabelStyle = themeDark.CardLabel
	cardValueStyle = themeDark.CardValue
	accentStyle    = themeDark.Accent_
	goodStyle      = themeDark.Good_
	warnStyle      = themeDark.Warn_
)

// applyTheme switches currentTheme and reassigns every package-level style var.
// Must be called before the next View() render.
func applyTheme(t *Theme) {
	currentTheme = t

	colorPrimary  = t.Primary
	colorAccent   = t.Accent
	colorMuted    = t.Muted
	colorGood     = t.Good
	colorWarn     = t.Warn
	colorFg       = t.Fg
	colorFgDim    = t.FgDim
	colorSelected = t.Selected

	titleStyle     = t.Title
	tabStyle       = t.Tab
	tabActiveStyle = t.TabActive
	headerBarStyle = t.HeaderBar
	footerStyle    = t.Footer
	sectionStyle   = t.Section
	cardStyle      = t.Card
	cardLabelStyle = t.CardLabel
	cardValueStyle = t.CardValue
	accentStyle    = t.Accent_
	goodStyle      = t.Good_
	warnStyle      = t.Warn_
}
