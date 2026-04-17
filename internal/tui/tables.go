package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"

	"github.com/GerardoFC8/claumeter/internal/stats"
)

func (m *Model) buildTables() {
	m.tblActivity = newActivityTable(m.report)
	m.tblSess = newSessionsTable(m.report)
	m.tblProj = newProjectsTable(m.report)
}

// newActivityTable: matrix Day × Model with totals row at bottom.
func newActivityTable(r stats.Report) table.Model {
	cols := []table.Column{
		{Title: "Day", Width: 12},
		{Title: "Prompts", Width: 8},
		{Title: "Turns", Width: 7},
	}
	for _, m := range r.Models {
		cols = append(cols, table.Column{Title: shortModel(m), Width: 11})
	}
	cols = append(cols, table.Column{Title: "Total", Width: 12})

	rows := make([]table.Row, 0, len(r.ByDay)+2)
	for _, d := range r.ByDay {
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
		row = append(row, compactNumber(d.Totals.GrandTotal()))
		rows = append(rows, row)
	}

	if len(r.ByDay) > 0 {
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
		totalRow = append(totalRow, compactNumber(r.Overall.GrandTotal()))
		rows = append(rows, totalRow)
	}

	return makeTable(cols, rows)
}

func newSessionsTable(r stats.Report) table.Model {
	cols := []table.Column{
		{Title: "Session", Width: 10},
		{Title: "Started", Width: 16},
		{Title: "Duration", Width: 9},
		{Title: "Prompts", Width: 8},
		{Title: "Turns", Width: 7},
		{Title: "Project", Width: 28},
		{Title: "Models", Width: 7},
		{Title: "Total tok.", Width: 13},
	}
	rows := make([]table.Row, 0, len(r.BySession))
	for _, s := range r.BySession {
		dur := s.LastSeen.Sub(s.FirstSeen)
		rows = append(rows, table.Row{
			shortSession(s.SessionID),
			s.FirstSeen.Local().Format("2006-01-02 15:04"),
			formatDuration(dur),
			humanNumber(s.Totals.Prompts),
			humanNumber(s.Totals.Turns),
			truncate(shortenPath(s.Cwd), 28),
			humanNumber(len(s.Models)),
			humanNumber(s.Totals.GrandTotal()),
		})
	}
	return makeTable(cols, rows)
}

func newProjectsTable(r stats.Report) table.Model {
	cols := []table.Column{
		{Title: "Project", Width: 46},
		{Title: "Prompts", Width: 8},
		{Title: "Turns", Width: 8},
		{Title: "Input", Width: 12},
		{Title: "Cache rd.", Width: 13},
		{Title: "Output", Width: 12},
		{Title: "Total", Width: 14},
	}
	rows := make([]table.Row, 0, len(r.ByProject))
	for _, p := range r.ByProject {
		rows = append(rows, table.Row{
			truncate(p.Cwd, 46),
			humanNumber(p.Totals.Prompts),
			humanNumber(p.Totals.Turns),
			humanNumber(p.Totals.InputTokens),
			humanNumber(p.Totals.CacheReadTokens),
			humanNumber(p.Totals.OutputTokens),
			humanNumber(p.Totals.GrandTotal()),
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
