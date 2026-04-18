package tui

import "github.com/charmbracelet/lipgloss"

// Theme holds the color palette and all precomputed derived styles for one theme.
type Theme struct {
	Name string

	// Base colors
	Primary  lipgloss.Color
	Accent   lipgloss.Color
	Muted    lipgloss.Color
	Good     lipgloss.Color
	Warn     lipgloss.Color
	Fg       lipgloss.Color
	FgDim    lipgloss.Color
	Selected lipgloss.Color

	// Derived styles (precomputed from the colors above, same shape as styles.go)
	Title     lipgloss.Style
	Tab       lipgloss.Style
	TabActive lipgloss.Style
	HeaderBar lipgloss.Style
	Footer    lipgloss.Style
	Section   lipgloss.Style
	Card      lipgloss.Style
	CardLabel lipgloss.Style
	CardValue lipgloss.Style
	Accent_   lipgloss.Style
	Good_     lipgloss.Style
	Warn_     lipgloss.Style
	Error_    lipgloss.Style
}

func buildTheme(
	name string,
	primary, accent, muted, good, warn, fg, fgDim, selected lipgloss.Color,
) Theme {
	tab := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(fgDim)

	return Theme{
		Name:     name,
		Primary:  primary,
		Accent:   accent,
		Muted:    muted,
		Good:     good,
		Warn:     warn,
		Fg:       fg,
		FgDim:    fgDim,
		Selected: selected,

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(primary).
			Padding(0, 1),

		Tab: tab,

		TabActive: tab.
			Foreground(accent).
			Bold(true).
			Underline(true),

		HeaderBar: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder(), false, false, true, false).
			BorderForeground(muted),

		Footer: lipgloss.NewStyle().
			Foreground(muted).
			Padding(0, 1),

		Section: lipgloss.NewStyle().
			Padding(1, 2),

		Card: lipgloss.NewStyle().
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(muted).
			Margin(0, 1, 0, 0),

		CardLabel: lipgloss.NewStyle().Foreground(fgDim),
		CardValue: lipgloss.NewStyle().Foreground(fg).Bold(true),

		Accent_: lipgloss.NewStyle().Foreground(accent).Bold(true),
		Good_:   lipgloss.NewStyle().Foreground(good).Bold(true),
		Warn_:   lipgloss.NewStyle().Foreground(warn).Bold(true),
		Error_:  lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true),
	}
}

var (
	// themeDark preserves the original hard-coded palette exactly.
	themeDark = buildTheme(
		"dark",
		"#7C3AED", // Primary
		"#22D3EE", // Accent
		"#6B7280", // Muted
		"#10B981", // Good
		"#F59E0B", // Warn
		"#E5E7EB", // Fg
		"#9CA3AF", // FgDim
		"#1F2937", // Selected
	)

	// themeLight — darker saturated foregrounds, readable on light terminal configs.
	themeLight = buildTheme(
		"light",
		"#6D28D9", // Primary
		"#0E7490", // Accent
		"#4B5563", // Muted
		"#047857", // Good
		"#B45309", // Warn
		"#111827", // Fg
		"#4B5563", // FgDim
		"#E5E7EB", // Selected
	)

	// themeHigh — monochrome + single accent, aims for WCAG AAA on dark terminals.
	themeHigh = buildTheme(
		"high-contrast",
		"#FFFFFF", // Primary
		"#FFFF00", // Accent
		"#D1D5DB", // Muted
		"#00FF00", // Good
		"#FFA500", // Warn
		"#FFFFFF", // Fg
		"#E5E7EB", // FgDim
		"#FFFFFF", // Selected
	)
)

// themeByName resolves a theme by its name string.
// Falls back to dark for unrecognised names.
func themeByName(name string) *Theme {
	switch name {
	case "light":
		return &themeLight
	case "high-contrast":
		return &themeHigh
	default:
		return &themeDark
	}
}

// allThemes returns the full ordered list of themes used for cycling.
func allThemes() []*Theme {
	return []*Theme{&themeDark, &themeLight, &themeHigh}
}
