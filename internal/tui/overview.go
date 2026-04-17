package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/GerardoFC8/claumeter/internal/stats"
)

func renderOverview(r stats.Report, width int) string {
	if r.Overall.Turns == 0 && r.Overall.Prompts == 0 {
		return sectionStyle.Render("No data found in ~/.claude/projects/")
	}

	ratio := ""
	if r.Overall.Prompts > 0 {
		ratio = fmt.Sprintf("%.1f×", float64(r.Overall.Turns)/float64(r.Overall.Prompts))
	}

	cards := []string{
		card("Total cost", goodStyle.Render(formatCost(r.Overall.Cost)), "USD, estimated"),
		card("Tokens", accentStyle.Render(compactNumber(r.Overall.GrandTotal())), humanNumber(r.Overall.GrandTotal())),
		card("Prompts", cardValueStyle.Render(humanNumber(r.Overall.Prompts)), "human msgs"),
		card("Turns", cardValueStyle.Render(humanNumber(r.Overall.Turns)), "API completions"),
		card("Turns/Prompt", accentStyle.Render(ratio), "agentic ratio"),
		card("Sessions", cardValueStyle.Render(humanNumber(len(r.BySession))), ""),
		card("Days", cardValueStyle.Render(humanNumber(len(r.ByDay))), ""),
	}
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, cards...)

	breakdown := lipgloss.JoinVertical(lipgloss.Left,
		accentStyle.Render("Token breakdown"),
		fmt.Sprintf("  Input (fresh)         %s", humanNumber(r.Overall.InputTokens)),
		fmt.Sprintf("  Cache creation (1h)   %s", humanNumber(r.Overall.CacheCreationTokens)),
		fmt.Sprintf("  Cache read            %s", humanNumber(r.Overall.CacheReadTokens)),
		fmt.Sprintf("  Output                %s", humanNumber(r.Overall.OutputTokens)),
	)

	dateRange := ""
	if !r.DateRange[0].IsZero() {
		dateRange = fmt.Sprintf("  %s  →  %s  (%s)",
			r.DateRange[0].Local().Format("2006-01-02 15:04"),
			r.DateRange[1].Local().Format("2006-01-02 15:04"),
			formatDuration(r.DateRange[1].Sub(r.DateRange[0])),
		)
	}
	rangeBlock := lipgloss.JoinVertical(lipgloss.Left,
		accentStyle.Render("Date range"),
		dateRange,
	)

	topModels := topNModels("Top models (by cost)", r.ByModel, 5)
	topDays := topNDays("Top days (by cost)", r.ByDay, 5)
	topProjects := topNProjects("Top projects (by cost)", r.ByProject, 5)

	col1 := lipgloss.JoinVertical(lipgloss.Left, breakdown, "", rangeBlock)
	col2 := lipgloss.JoinVertical(lipgloss.Left, topModels, "", topDays)
	col3 := topProjects

	bottom := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(width/3).Render(col1),
		lipgloss.NewStyle().Width(width/3).Render(col2),
		lipgloss.NewStyle().Width(width/3).Render(col3),
	)

	return sectionStyle.Render(lipgloss.JoinVertical(lipgloss.Left, row1, "", bottom))
}

func card(label, primary, secondary string) string {
	content := cardLabelStyle.Render(label) + "\n" + primary
	if secondary != "" {
		content += "\n" + cardLabelStyle.Render(secondary)
	}
	return cardStyle.Render(content)
}

func topNModels(title string, models []stats.ModelStat, n int) string {
	if len(models) == 0 {
		return accentStyle.Render(title) + "\n  (empty)"
	}
	sorted := make([]stats.ModelStat, len(models))
	copy(sorted, models)
	sortByCost := func(i, j int) bool { return sorted[i].Totals.Cost > sorted[j].Totals.Cost }
	bubbleSort(len(sorted), sortByCost, func(i, j int) { sorted[i], sorted[j] = sorted[j], sorted[i] })
	if n > len(sorted) {
		n = len(sorted)
	}
	max := sorted[0].Totals.Cost
	var b strings.Builder
	b.WriteString(accentStyle.Render(title))
	b.WriteString("\n")
	for i := 0; i < n; i++ {
		m := sorted[i]
		b.WriteString(fmt.Sprintf("  %-22s %s %s\n",
			truncate(shortModel(m.Model), 22),
			barFloat(m.Totals.Cost, max, 10),
			formatCost(m.Totals.Cost),
		))
	}
	return b.String()
}

func topNDays(title string, days []stats.DayStat, n int) string {
	if len(days) == 0 {
		return accentStyle.Render(title) + "\n  (empty)"
	}
	sorted := make([]stats.DayStat, len(days))
	copy(sorted, days)
	sortByCost := func(i, j int) bool { return sorted[i].Totals.Cost > sorted[j].Totals.Cost }
	bubbleSort(len(sorted), sortByCost, func(i, j int) { sorted[i], sorted[j] = sorted[j], sorted[i] })
	if n > len(sorted) {
		n = len(sorted)
	}
	max := sorted[0].Totals.Cost
	var b strings.Builder
	b.WriteString(accentStyle.Render(title))
	b.WriteString("\n")
	for i := 0; i < n; i++ {
		d := sorted[i]
		b.WriteString(fmt.Sprintf("  %-12s %s %s\n",
			d.Day,
			barFloat(d.Totals.Cost, max, 10),
			formatCost(d.Totals.Cost),
		))
	}
	return b.String()
}

func topNProjects(title string, projects []stats.ProjectStat, n int) string {
	if len(projects) == 0 {
		return accentStyle.Render(title) + "\n  (empty)"
	}
	sorted := make([]stats.ProjectStat, len(projects))
	copy(sorted, projects)
	sortByCost := func(i, j int) bool { return sorted[i].Totals.Cost > sorted[j].Totals.Cost }
	bubbleSort(len(sorted), sortByCost, func(i, j int) { sorted[i], sorted[j] = sorted[j], sorted[i] })
	if n > len(sorted) {
		n = len(sorted)
	}
	max := sorted[0].Totals.Cost
	var b strings.Builder
	b.WriteString(accentStyle.Render(title))
	b.WriteString("\n")
	for i := 0; i < n; i++ {
		p := sorted[i]
		b.WriteString(fmt.Sprintf("  %-26s %s %s\n",
			truncate(shortenPath(p.Cwd), 26),
			barFloat(p.Totals.Cost, max, 10),
			formatCost(p.Totals.Cost),
		))
	}
	return b.String()
}

func bubbleSort(n int, less func(i, j int) bool, swap func(i, j int)) {
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if less(j, i) {
				swap(i, j)
			}
		}
	}
}

func bar(v, max, width int) string {
	if max == 0 {
		return strings.Repeat(" ", width)
	}
	filled := (v * width) / max
	if filled > width {
		filled = width
	}
	return lipgloss.NewStyle().Foreground(colorAccent).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(colorMuted).Render(strings.Repeat("░", width-filled))
}

func barFloat(v, max float64, width int) string {
	if max <= 0 {
		return strings.Repeat(" ", width)
	}
	filled := int(float64(width) * v / max)
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return lipgloss.NewStyle().Foreground(colorAccent).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(colorMuted).Render(strings.Repeat("░", width-filled))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:max]
	}
	return s[:max-1] + "…"
}
