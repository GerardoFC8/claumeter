package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/GerardoFC8/claumeter/internal/stats"
)

const (
	compactWidth = 100
	wideWidth    = 140
	narrowWidth  = 70
)

func renderOverview(r stats.Report, width int) string {
	if r.Overall.Turns == 0 && r.Overall.Prompts == 0 {
		return sectionStyle.Render(
			warnStyle.Render("No JSONL data in ~/.claude/projects/") + "\n\n" +
				cardLabelStyle.Render("Run Claude Code to generate usage data, then reopen claumeter."),
		)
	}

	compact := width < compactWidth

	ratio := ""
	if r.Overall.Prompts > 0 {
		ratio = fmt.Sprintf("%.1f×", float64(r.Overall.Turns)/float64(r.Overall.Prompts))
	}

	allCards := []string{
		card("Total cost", goodStyle.Render(formatCost(r.Overall.Cost)), "USD, estimated"),
		card("Tokens", accentStyle.Render(compactNumber(r.Overall.GrandTotal())), humanNumber(r.Overall.GrandTotal())),
		card("Prompts", cardValueStyle.Render(humanNumber(r.Overall.Prompts)), "human msgs"),
		card("Turns", cardValueStyle.Render(humanNumber(r.Overall.Turns)), "API completions"),
		card("Turns/Prompt", accentStyle.Render(ratio), "agentic ratio"),
		card("Sessions", cardValueStyle.Render(humanNumber(len(r.BySession))), ""),
		card("Days", cardValueStyle.Render(humanNumber(len(r.ByDay))), ""),
	}

	var row1 string
	switch {
	case compact:
		row1 = renderCompactMetrics(r, ratio)
	case width < wideWidth:
		row1 = lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Top, allCards[:4]...),
			lipgloss.JoinHorizontal(lipgloss.Top, allCards[4:]...),
		)
	default:
		row1 = lipgloss.JoinHorizontal(lipgloss.Top, allCards...)
	}

	breakdown := lipgloss.JoinVertical(lipgloss.Left,
		accentStyle.Render("Token breakdown"),
		fmt.Sprintf("  Input (fresh)         %s", humanNumber(r.Overall.InputTokens)),
		fmt.Sprintf("  Cache write (5m)      %s", humanNumber(r.Overall.CacheCreation5mTokens)),
		fmt.Sprintf("  Cache write (1h)      %s", humanNumber(r.Overall.CacheCreation1hTokens)),
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

	cacheHit := renderCacheHitRate(r)
	costProjection := renderCostProjection(r)
	topSessions := renderTopSessions(r)
	heatmap := renderHourlyHeatmap(r)

	costBreakdown := renderCostBreakdown(r, width)

	col1 := lipgloss.JoinVertical(lipgloss.Left, breakdown, "", rangeBlock, "", cacheHit, "", costProjection)
	col2 := lipgloss.JoinVertical(lipgloss.Left, topModels, "", topDays)
	col3 := lipgloss.JoinVertical(lipgloss.Left, topProjects, "", topSessions)

	var bottom string
	switch {
	case width < narrowWidth:
		bottom = lipgloss.JoinVertical(lipgloss.Left,
			col1, "", col2, "", col3,
		)
	case compact:
		half := width / 2
		bottom = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(half).Render(
				lipgloss.JoinVertical(lipgloss.Left, col1, "", col2),
			),
			lipgloss.NewStyle().Width(width-half).Render(col3),
		)
	default:
		bottom = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(width/3).Render(col1),
			lipgloss.NewStyle().Width(width/3).Render(col2),
			lipgloss.NewStyle().Width(width/3).Render(col3),
		)
	}

	return sectionStyle.Render(lipgloss.JoinVertical(lipgloss.Left, row1, "", bottom, "", costBreakdown, "", heatmap))
}

// renderCostBreakdown shows the 5-bucket cost split (input / cache write 5m /
// cache write 1h / cache read / output) per model, with rate + cost + % share.
func renderCostBreakdown(r stats.Report, width int) string {
	title := accentStyle.Render("Cost Breakdown (Input / Cache / Output)")
	if r.Overall.Cost == 0 || len(r.ByModel) == 0 {
		return title + "\n  " + cardLabelStyle.Render("(no priced usage in range)")
	}
	cb := stats.BuildCostBreakdown(r)

	bucketLabel := map[string]string{
		"input":          "Input",
		"cache_write_5m": "Cache W (5m)",
		"cache_write_1h": "Cache W (1h)",
		"cache_read":     "Cache Read",
		"output":         "Output",
	}

	// Header: Bucket | Tokens | Rate | Cost | % of model
	header := fmt.Sprintf("  %-14s %12s %10s %10s %7s",
		cardLabelStyle.Render("Bucket"),
		cardLabelStyle.Render("Tokens"),
		cardLabelStyle.Render("Rate $/M"),
		cardLabelStyle.Render("Cost"),
		cardLabelStyle.Render("% of ∑"),
	)

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n")

	renderModel := func(m stats.ModelBreakdown, pctOfGrand float64) {
		modelHdr := fmt.Sprintf("  %s  %s  %s",
			cardValueStyle.Render(shortModel(m.Model)),
			goodStyle.Render(formatCost(m.TotalCost)),
			cardLabelStyle.Render(fmt.Sprintf("(%.1f%% of grand total)", pctOfGrand)),
		)
		b.WriteString(modelHdr)
		b.WriteString("\n")
		b.WriteString(header)
		b.WriteString("\n")
		for _, bk := range m.Buckets {
			if bk.Tokens == 0 && bk.Cost == 0 {
				continue
			}
			b.WriteString(fmt.Sprintf("  %-14s %12s %10s %10s %6.1f%%\n",
				bucketLabel[bk.Kind],
				humanNumber(bk.Tokens),
				fmt.Sprintf("$%.2f", bk.Rate),
				formatCost(bk.Cost),
				bk.Pct,
			))
		}
	}

	renderModel(cb.Overall, 100)
	for _, m := range cb.ByModel {
		if m.TotalCost == 0 {
			continue
		}
		b.WriteString("\n")
		renderModel(m, m.Pct)
	}
	return b.String()
}

func renderCompactMetrics(r stats.Report, ratio string) string {
	const keyW = 16
	kv := func(key, val string) string {
		padded := fmt.Sprintf("%-*s", keyW, key+":")
		return cardLabelStyle.Render(padded) + val
	}

	turnsVal := cardValueStyle.Render(humanNumber(r.Overall.Turns))
	if ratio != "" {
		turnsVal += "  " + cardLabelStyle.Render("("+ratio+" ratio)")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		kv("Total cost", goodStyle.Render(formatCost(r.Overall.Cost))+" "+cardLabelStyle.Render("USD")),
		kv("Tokens", accentStyle.Render(compactNumber(r.Overall.GrandTotal()))+" "+cardLabelStyle.Render(humanNumber(r.Overall.GrandTotal()))),
		kv("Prompts", cardValueStyle.Render(humanNumber(r.Overall.Prompts))),
		kv("Turns", turnsVal),
		kv("Sessions", cardValueStyle.Render(humanNumber(len(r.BySession)))),
		kv("Days", cardValueStyle.Render(humanNumber(len(r.ByDay)))),
	)
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
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Totals.Cost > sorted[j].Totals.Cost })
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
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Totals.Cost > sorted[j].Totals.Cost })
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
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Totals.Cost > sorted[j].Totals.Cost })
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
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}

func renderCacheHitRate(r stats.Report) string {
	t := r.Overall
	denom := t.CacheReadTokens + t.InputTokens
	if denom == 0 {
		return accentStyle.Render("Cache Hit Rate") + "\n  —"
	}
	pct := float64(t.CacheReadTokens) / float64(denom) * 100
	return lipgloss.JoinVertical(lipgloss.Left,
		accentStyle.Render("Cache Hit Rate"),
		fmt.Sprintf("  %s%.1f%%%s", "", pct, ""),
		cardLabelStyle.Render("  higher = more savings"),
	)
}

func renderCostProjection(r stats.Report) string {
	title := accentStyle.Render("Projected Monthly Cost")
	if len(r.ByDay) < 3 {
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			"  —",
			cardLabelStyle.Render("  need 3+ days of data"),
		)
	}
	// Use up to 7 most recent days.
	days := r.ByDay
	if len(days) > 7 {
		days = days[:7]
	}
	var total float64
	for _, d := range days {
		total += d.Totals.Cost
	}
	avg := total / float64(len(days))
	projected := avg * 30
	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"  "+goodStyle.Render(formatCost(projected)),
		cardLabelStyle.Render(fmt.Sprintf("  based on last %d days avg", len(days))),
	)
}

func renderTopSessions(r stats.Report) string {
	title := accentStyle.Render("Most Expensive Sessions")
	if len(r.BySession) == 0 {
		return title + "\n  (none)"
	}
	sorted := make([]stats.SessionStat, len(r.BySession))
	copy(sorted, r.BySession)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Totals.Cost > sorted[j].Totals.Cost
	})
	n := 3
	if n > len(sorted) {
		n = len(sorted)
	}
	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n")
	for i := 0; i < n; i++ {
		s := sorted[i]
		dur := s.LastSeen.Sub(s.FirstSeen)
		b.WriteString(fmt.Sprintf("  %s  %s  %s\n",
			cardValueStyle.Render(shortSession(s.SessionID)),
			cardLabelStyle.Render(formatDuration(dur)),
			goodStyle.Render(formatCost(s.Totals.Cost)),
		))
	}
	b.WriteString(cardLabelStyle.Render("  press 3 to see all sessions"))
	return b.String()
}

var sparkBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func renderHourlyHeatmap(r stats.Report) string {
	title := accentStyle.Render("Hourly Activity (prompts per hour, all days in range)")
	hours := r.PromptsByHour
	maxVal := 0
	for _, v := range hours {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		return title + "\n  " + cardLabelStyle.Render("(no prompt data)")
	}

	var bars strings.Builder
	bars.WriteString("  ")
	for h := 0; h < 24; h++ {
		v := hours[h]
		idx := 0
		if maxVal > 0 {
			idx = int(float64(v) / float64(maxVal) * float64(len(sparkBlocks)-1))
		}
		if v == 0 {
			bars.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render(string(sparkBlocks[0])))
		} else {
			bars.WriteString(lipgloss.NewStyle().Foreground(colorAccent).Render(string(sparkBlocks[idx])))
		}
	}

	labels := cardLabelStyle.Render("  0    6    12   18  23")
	return lipgloss.JoinVertical(lipgloss.Left, title, bars.String(), labels)
}
