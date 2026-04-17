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
		card("Total tokens", accentStyle.Render(compactNumber(r.Overall.GrandTotal())), humanNumber(r.Overall.GrandTotal())),
		card("Prompts", goodStyle.Render(humanNumber(r.Overall.Prompts)), "human msgs"),
		card("Turns", cardValueStyle.Render(humanNumber(r.Overall.Turns)), "API completions"),
		card("Turns/Prompt", accentStyle.Render(ratio), "agentic ratio"),
		card("Sessions", cardValueStyle.Render(humanNumber(len(r.BySession))), ""),
		card("Projects", cardValueStyle.Render(humanNumber(len(r.ByProject))), ""),
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

	topModels := topN("Top models", r.ByModel, 5, func(i int) (string, int) {
		s := r.ByModel[i]
		return shortModel(s.Model), s.Totals.GrandTotal()
	})

	topDays := topNSorted("Top days (by tokens)", r.ByDay, 5)
	topProjects := topNProjects("Top projects", r.ByProject, 5)

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

func topN[T any](title string, items []T, n int, extract func(i int) (string, int)) string {
	if len(items) == 0 {
		return accentStyle.Render(title) + "\n  (empty)"
	}
	if n > len(items) {
		n = len(items)
	}
	var b strings.Builder
	b.WriteString(accentStyle.Render(title))
	b.WriteString("\n")
	max := 0
	for i := 0; i < n; i++ {
		_, v := extract(i)
		if v > max {
			max = v
		}
	}
	for i := 0; i < n; i++ {
		name, v := extract(i)
		b.WriteString(fmt.Sprintf("  %-22s %s %s\n", truncate(name, 22), bar(v, max, 10), compactNumber(v)))
	}
	return b.String()
}

func topNSorted(title string, days []stats.DayStat, n int) string {
	if len(days) == 0 {
		return accentStyle.Render(title) + "\n  (empty)"
	}
	sorted := make([]stats.DayStat, len(days))
	copy(sorted, days)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Totals.GrandTotal() > sorted[i].Totals.GrandTotal() {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	if n > len(sorted) {
		n = len(sorted)
	}
	var b strings.Builder
	b.WriteString(accentStyle.Render(title))
	b.WriteString("\n")
	max := sorted[0].Totals.GrandTotal()
	for i := 0; i < n; i++ {
		s := sorted[i]
		b.WriteString(fmt.Sprintf("  %-12s %s %s\n", s.Day, bar(s.Totals.GrandTotal(), max, 10), compactNumber(s.Totals.GrandTotal())))
	}
	return b.String()
}

func topNProjects(title string, projects []stats.ProjectStat, n int) string {
	if len(projects) == 0 {
		return accentStyle.Render(title) + "\n  (empty)"
	}
	if n > len(projects) {
		n = len(projects)
	}
	var b strings.Builder
	b.WriteString(accentStyle.Render(title))
	b.WriteString("\n")
	max := projects[0].Totals.GrandTotal()
	for i := 0; i < n; i++ {
		p := projects[i]
		b.WriteString(fmt.Sprintf("  %-26s %s %s\n", truncate(shortenPath(p.Cwd), 26), bar(p.Totals.GrandTotal(), max, 10), compactNumber(p.Totals.GrandTotal())))
	}
	return b.String()
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

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:max]
	}
	return s[:max-1] + "…"
}
