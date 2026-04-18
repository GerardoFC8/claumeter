package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"

	"github.com/GerardoFC8/claumeter/internal/stats"
)

// newTurnsTable builds the turns table for session drill-down.
func newTurnsTable(sd stats.SessionDetail, width int) table.Model {
	// Fixed columns with known widths.
	const (
		wTime   = 10 // "HH:MM:SS" + padding
		wModel  = 20 // short model name
		wIn     = 9  // input tokens
		wOut    = 9  // output tokens
		wCost   = 8  // cost
		wBorder = 2  // padding/border
	)
	fixedW := wTime + wModel + wIn + wOut + wCost + wBorder
	wTools := width - fixedW - 4 // 4 = table outer padding estimate
	if wTools < 12 {
		wTools = 12
	}

	cols := []table.Column{
		{Title: "Time", Width: wTime},
		{Title: "Model", Width: wModel},
		{Title: "Input", Width: wIn},
		{Title: "Output", Width: wOut},
		{Title: "Cost", Width: wCost},
		{Title: "Tools", Width: wTools},
	}

	rows := make([]table.Row, 0, len(sd.Turns))
	for _, t := range sd.Turns {
		toolStr := strings.Join(t.Tools, ", ")
		if len(toolStr) > wTools-1 {
			toolStr = toolStr[:wTools-4] + "..."
		}
		rows = append(rows, table.Row{
			t.Timestamp.Local().Format("15:04:05"),
			shortModel(t.Model),
			compactNumber(t.Totals.TotalInput()),
			compactNumber(t.Totals.OutputTokens),
			formatCost(t.Totals.Cost),
			toolStr,
		})
	}

	return makeTable(cols, rows)
}

func (m *Model) buildTables() {
	m.buildTablesWithQuery("")
}

func (m *Model) buildTablesWithQuery(query string) {
	m.tblActivity = newActivityTable(m.report, query)
	m.tblSess = newSessionsTable(m.report, query)
	m.tblProj = newProjectsTable(m.report, query)
}

// newActivityTable: matrix Day × Model with totals row at bottom.
func newActivityTable(r stats.Report, query string) table.Model {
	cols := []table.Column{
		{Title: "Day", Width: 12},
		{Title: "Prompts", Width: 8},
		{Title: "Turns", Width: 7},
	}
	for _, m := range r.Models {
		cols = append(cols, table.Column{Title: shortModel(m), Width: 11})
	}
	cols = append(cols,
		table.Column{Title: "Tokens", Width: 10},
		table.Column{Title: "Cost", Width: 9},
	)

	filteredDays := filterActivityRows(r.ByDay, query)

	rows := make([]table.Row, 0, len(filteredDays)+2)
	for _, d := range filteredDays {
		row := table.Row{
			d.Day,
			humanNumber(d.Totals.Prompts),
			humanNumber(d.Totals.Turns),
		}
		for _, model := range r.Models {
			t := d.ByModel[model]
			if t.GrandTotal() == 0 {
				row = append(row, "—")
			} else {
				row = append(row, compactNumber(t.GrandTotal()))
			}
		}
		row = append(row, compactNumber(d.Totals.GrandTotal()), formatCost(d.Totals.Cost))
		rows = append(rows, row)
	}

	if len(filteredDays) > 0 {
		sep := make(table.Row, len(cols))
		for i, c := range cols {
			sep[i] = strings.Repeat("─", c.Width)
		}
		rows = append(rows, sep)

		totalRow := table.Row{
			"▸ TOTAL",
			humanNumber(r.Overall.Prompts),
			humanNumber(r.Overall.Turns),
		}
		for _, model := range r.Models {
			var modelTotal int
			for _, mm := range r.ByModel {
				if mm.Model == model {
					modelTotal = mm.Totals.GrandTotal()
					break
				}
			}
			totalRow = append(totalRow, compactNumber(modelTotal))
		}
		totalRow = append(totalRow,
			compactNumber(r.Overall.GrandTotal()),
			formatCost(r.Overall.Cost),
		)
		rows = append(rows, totalRow)
	}

	return makeTable(cols, rows)
}

func newSessionsTable(r stats.Report, query string) table.Model {
	cols := []table.Column{
		{Title: "Session", Width: 10},
		{Title: "Started", Width: 16},
		{Title: "Duration", Width: 9},
		{Title: "Prompts", Width: 8},
		{Title: "Turns", Width: 7},
		{Title: "Project", Width: 26},
		{Title: "Tokens", Width: 11},
		{Title: "Cost", Width: 9},
	}
	filteredSessions := filterSessionRows(r.BySession, query)
	rows := make([]table.Row, 0, len(filteredSessions))
	for _, s := range filteredSessions {
		dur := s.LastSeen.Sub(s.FirstSeen)
		rows = append(rows, table.Row{
			shortSession(s.SessionID),
			s.FirstSeen.Local().Format("2006-01-02 15:04"),
			formatDuration(dur),
			humanNumber(s.Totals.Prompts),
			humanNumber(s.Totals.Turns),
			truncate(shortenPath(s.Cwd), 26),
			compactNumber(s.Totals.GrandTotal()),
			formatCost(s.Totals.Cost),
		})
	}
	return makeTable(cols, rows)
}

func newProjectsTable(r stats.Report, query string) table.Model {
	cols := []table.Column{
		{Title: "Project", Width: 42},
		{Title: "Prompts", Width: 8},
		{Title: "Turns", Width: 8},
		{Title: "Input", Width: 11},
		{Title: "Cache rd.", Width: 11},
		{Title: "Output", Width: 10},
		{Title: "Tokens", Width: 11},
		{Title: "Cost", Width: 10},
	}
	filteredProjects := filterProjectRows(r.ByProject, query)
	rows := make([]table.Row, 0, len(filteredProjects))
	for _, p := range filteredProjects {
		rows = append(rows, table.Row{
			truncate(p.Cwd, 42),
			humanNumber(p.Totals.Prompts),
			humanNumber(p.Totals.Turns),
			compactNumber(p.Totals.InputTokens),
			compactNumber(p.Totals.CacheReadTokens),
			compactNumber(p.Totals.OutputTokens),
			compactNumber(p.Totals.GrandTotal()),
			formatCost(p.Totals.Cost),
		})
	}
	return makeTable(cols, rows)
}

func makeTable(cols []table.Column, rows []table.Row) table.Model {
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(15),
	)
	t.SetStyles(focusedTableStyles())
	return t
}
