package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/GerardoFC8/claumeter/internal/stats"
)

func renderTools(ts stats.ToolStats, width, height int, filterLabel string) string {
	if ts.Total == 0 {
		msg := warnStyle.Render("No tool usage in range "+filterLabel+".") + "\n" +
			cardLabelStyle.Render("Press f to change range or F to cycle back.")
		return sectionStyle.Render(msg)
	}

	header := fmt.Sprintf("%s tool invocations", accentStyle.Render(humanNumber(ts.Total)))

	// 2x2 grid
	panelWidth := width/2 - 2
	if panelWidth < 20 {
		panelWidth = 20
	}
	panelHeight := (height - 10) / 2
	if panelHeight < 6 {
		panelHeight = 6
	}
	itemsPerPanel := panelHeight - 2

	topLeft := toolPanel("Built-in tools", ts.Builtins, itemsPerPanel, panelWidth)
	topRight := toolPanel("MCP — by server", ts.Servers, itemsPerPanel, panelWidth)
	botLeft := toolPanel("Skills", ts.Skills, itemsPerPanel, panelWidth)
	botRight := toolPanel("Sub-agents", ts.Agents, itemsPerPanel, panelWidth)

	row1 := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(panelWidth).Render(topLeft),
		lipgloss.NewStyle().Width(panelWidth).Render(topRight),
	)
	row2 := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(panelWidth).Render(botLeft),
		lipgloss.NewStyle().Width(panelWidth).Render(botRight),
	)

	body := lipgloss.JoinVertical(lipgloss.Left, header, "", row1, row2)
	return sectionStyle.Render(body)
}

func toolPanel(title string, entries []stats.ToolEntry, limit, width int) string {
	if len(entries) == 0 {
		return accentStyle.Render(title) + "\n  " + cardLabelStyle.Render("(none)")
	}

	var b strings.Builder
	b.WriteString(accentStyle.Render(title))
	b.WriteString("\n")

	max := entries[0].Count
	if limit > len(entries) {
		limit = len(entries)
	}
	nameW := width - 20
	if nameW < 10 {
		nameW = 10
	}

	for i := 0; i < limit; i++ {
		e := entries[i]
		b.WriteString(fmt.Sprintf("  %-*s %s %s\n",
			nameW, truncate(e.Name, nameW),
			bar(e.Count, max, 8),
			humanNumber(e.Count),
		))
	}
	if limit < len(entries) {
		remaining := 0
		for _, e := range entries[limit:] {
			remaining += e.Count
		}
		b.WriteString(cardLabelStyle.Render(fmt.Sprintf("  … %d more (%s total)", len(entries)-limit, humanNumber(remaining))))
	}
	return b.String()
}
