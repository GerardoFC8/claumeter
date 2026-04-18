package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderWelcomeOverlay(width, height int) string {
	if width == 0 || height == 0 {
		return ""
	}
	lines := []string{
		accentStyle.Render("Welcome to claumeter"),
		"",
		cardLabelStyle.Render("Your Claude Code usage dashboard — six tabs, zero guesswork."),
		"",
		cardValueStyle.Render("1  Overview") + "   cost, tokens, top models and projects at a glance",
		cardValueStyle.Render("2  Activity") + "   day-by-day token breakdown and hourly usage patterns",
		cardValueStyle.Render("3  Sessions") + "   per-conversation detail  (press enter to drill in)",
		cardValueStyle.Render("4  Projects") + "   spend grouped by project folder",
		cardValueStyle.Render("5  Tools   ") + "   which tools, MCP servers, skills and sub-agents you used",
		cardValueStyle.Render("6  Compare ") + "   diff two date ranges side-by-side",
		"",
		"",
		cardLabelStyle.Render("Press ? anytime for full help  ·  Press any key to continue  ·  ctrl+c to quit"),
	}

	content := strings.Join(lines, "\n")

	boxW := width - 12
	if boxW < 60 {
		boxW = 60
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(2, 4).
		Width(boxW).
		Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
