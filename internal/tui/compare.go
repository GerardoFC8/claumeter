package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/GerardoFC8/claumeter/internal/stats"
)

// renderCompare builds the Compare tab view. It derives both range reports
// lazily from m.allData each render — no rebuild pipeline changes needed.
func (m Model) renderCompare() string {
	now := time.Now()

	// --- header line: labels ---
	aLabel := cardLabelStyle.Render("A:") + " " + accentStyle.Render(m.cmpA.Label())
	bLabel := cardLabelStyle.Render("B:") + " " + accentStyle.Render(m.cmpB.Label())
	labelLine := aLabel + "    " + bLabel

	// --- second line: date windows ---
	aRange := formatRangeWindow(m.cmpA, now)
	bRange := formatRangeWindow(m.cmpB, now)
	windowLine := cardLabelStyle.Render("A: "+aRange+"    B: "+bRange)

	// --- compute ---
	aFiltered := m.cmpA.Apply(m.allData)
	bFiltered := m.cmpB.Apply(m.allData)
	aReport := stats.Build(aFiltered)
	bReport := stats.Build(bFiltered)
	cmp := stats.Compare(aReport.Overall, bReport.Overall)

	// --- table ---
	table := renderCompareTable(cmp, m.width)

	// --- hint footer ---
	hint := cardLabelStyle.Render("a/A = cycle A range  ·  b/B = cycle B range")

	body := strings.Join([]string{
		labelLine,
		windowLine,
		"",
		table,
		"",
		hint,
	}, "\n")

	return sectionStyle.Render(body)
}

func formatRangeWindow(p stats.FilterPreset, now time.Time) string {
	if p == stats.FilterAll {
		return "all time"
	}
	from, to := p.Range(now)
	if from.IsZero() {
		return "all time"
	}
	// to is exclusive; display the last included day.
	return from.Format("2006-01-02") + " → " + to.AddDate(0, 0, -1).Format("2006-01-02")
}

// renderCompareTable builds the aligned metrics table. Column widths adapt
// to m.width: the Metric column shrinks to ensure all numeric columns fit
// at a minimum terminal width of 80 cols.
func renderCompareTable(cmp stats.Comparison, width int) string {
	// Fixed widths for numeric columns (A, B, Delta, %).
	const (
		wA     = 12
		wB     = 12
		wDelta = 14
		wPct   = 8
		wGap   = 2 // space between columns
	)
	fixedW := wA + wB + wDelta + wPct + 4*wGap

	// Metric column: fill remaining space, minimum 12.
	metricW := width - fixedW - 4 // 4 for sectionStyle padding estimate
	if metricW < 12 {
		metricW = 12
	}
	// Cap at 20 so numbers dominate.
	if metricW > 20 {
		metricW = 20
	}

	type row struct {
		metric string
		d      stats.Delta
		isCost bool
	}
	rows := []row{
		{"Prompts", cmp.Prompts, false},
		{"Turns", cmp.Turns, false},
		{"Input tokens", cmp.InputTokens, false},
		{"Cache creation", cmp.CacheCreationTokens, false},
		{"Cache read", cmp.CacheReadTokens, false},
		{"Output tokens", cmp.OutputTokens, false},
		{"Tokens (total)", cmp.TotalTokens, false},
		{"Cost (USD)", cmp.CostUSD, true},
	}

	// --- header ---
	var sb strings.Builder
	header := fmt.Sprintf("%-*s  %*s  %*s  %*s  %*s",
		metricW, "Metric",
		wA, "A",
		wB, "B",
		wDelta, "Delta",
		wPct, "%",
	)
	sb.WriteString(cardLabelStyle.Render(header))
	sb.WriteString("\n")
	sb.WriteString(cardLabelStyle.Render(strings.Repeat("─", metricW+fixedW)))
	sb.WriteString("\n")

	for _, r := range rows {
		metric := truncate(r.metric, metricW)

		var aStr, bStr, deltaStr, pctStr string
		if r.isCost {
			aStr = formatCost(r.d.A)
			bStr = formatCost(r.d.B)
			deltaStr = formatDeltaCost(r.d.Delta)
			pctStr = formatPct(r.d.Pct)
		} else {
			aStr = compactNumber(int(r.d.A))
			bStr = compactNumber(int(r.d.B))
			deltaStr = formatDeltaInt(int(r.d.Delta))
			pctStr = formatPct(r.d.Pct)
		}

		metricCol := cardValueStyle.Render(fmt.Sprintf("%-*s", metricW, metric))
		aCol := cardLabelStyle.Render(fmt.Sprintf("%*s", wA, aStr))
		bCol := cardLabelStyle.Render(fmt.Sprintf("%*s", wB, bStr))
		deltaCol := deltaStyle(r.d.Delta).Render(fmt.Sprintf("%*s", wDelta, deltaStr))
		pctCol := deltaStyle(r.d.Delta).Render(fmt.Sprintf("%*s", wPct, pctStr))

		line := fmt.Sprintf("%s  %s  %s  %s  %s",
			metricCol, aCol, bCol, deltaCol, pctCol)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}

// deltaStyle returns goodStyle for negative delta (less spend/tokens = better)
// and warnStyle for positive delta (more spend/tokens). Zero gets muted.
func deltaStyle(delta float64) lipgloss.Style {
	switch {
	case delta < 0:
		return goodStyle
	case delta > 0:
		return warnStyle
	default:
		return cardLabelStyle
	}
}

func formatDeltaInt(n int) string {
	if n > 0 {
		return "+" + compactNumber(n)
	}
	return compactNumber(n)
}

func formatDeltaCost(c float64) string {
	if c == 0 {
		return "—"
	}
	// Use absolute value for formatting, then prefix sign manually.
	abs := c
	if abs < 0 {
		abs = -abs
	}
	var formatted string
	switch {
	case abs < 0.01:
		formatted = "<$0.01"
	case abs < 10:
		formatted = fmt.Sprintf("$%.2f", abs)
	case abs < 10000:
		formatted = fmt.Sprintf("$%.0f", abs)
	default:
		formatted = fmt.Sprintf("$%.1fK", abs/1000)
	}
	if c > 0 {
		return "+" + formatted
	}
	return "-" + formatted
}

func formatPct(p float64) string {
	if p == 0 {
		return "—"
	}
	if p > 0 {
		return fmt.Sprintf("+%.1f%%", p)
	}
	return fmt.Sprintf("%.1f%%", p)
}
