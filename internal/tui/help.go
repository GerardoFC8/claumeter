package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderHelpOverlay(width, height int) string {
	if width == 0 || height == 0 {
		return ""
	}
	sections := []string{
		accentStyle.Render("Keybindings"),
		"",
		cardValueStyle.Render("Navigation") + "  tab/shift+tab · h/l — prev/next tab  · 1–6 jump to tab · ←→ scroll table",
		cardValueStyle.Render("Filter/Search") + "  f/F cycle range · / search · enter confirm · esc cancel · ctrl+u clear",
		cardValueStyle.Render("Theme & Plan") + "  t cycle theme · Q cycle plan",
		cardValueStyle.Render("Sessions") + "  enter drill-down · esc/backspace back",
		cardValueStyle.Render("Compare") + "  a/A cycle range A · b/B cycle range B",
		cardValueStyle.Render("General") + "  j/k ↑↓ · g/G top/bot · q/ctrl+c quit · ? toggle help",
		"",
		accentStyle.Render("Glossary"),
		"",
		cardValueStyle.Render("Prompt") + "        one human message you sent to Claude",
		cardValueStyle.Render("Turn") + "          one API completion (often many per prompt in agentic loops)",
		cardValueStyle.Render("Agentic ratio") + "  turns ÷ prompts — API calls per message",
		cardValueStyle.Render("Cache read") + "    tokens reused from Anthropic's prompt cache (~10% price)",
		cardValueStyle.Render("Cache create") + "  tokens written to 1-hour cache tier",
		cardValueStyle.Render("Input tokens") + "  fresh tokens sent to Claude (full price)",
		cardValueStyle.Render("Output tokens") + " tokens Claude generated (full price)",
		cardValueStyle.Render("MCP") + "           Model Context Protocol — external tool servers",
		cardValueStyle.Render("Session") + "       one continuous Claude Code conversation",
		"",
		cardLabelStyle.Render("Press ? or esc to close"),
	}

	content := strings.Join(sections, "\n")

	boxW := width - 8
	if boxW < 60 {
		boxW = 60
	}
	boxH := height - 4
	if boxH < 10 {
		boxH = 10
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 3).
		Width(boxW).
		Height(boxH).
		Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
