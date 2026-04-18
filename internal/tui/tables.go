package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"

	"github.com/GerardoFC8/claumeter/internal/stats"
)

// columnDef describes a table column with a minimum acceptable width and a
// drop priority (lower number = drop first when space is tight).
type columnDef struct {
	title    string
	width    int
	minWidth int
	priority int // higher = keep longer; 1 = drop first
}

// fitColumns selects which columns to show given availableWidth.
// It drops columns in ascending priority order until the rest fit.
// If after dropping all optional columns the remainders still exceed
// availableWidth, the flexible column (largest width, highest priority)
// gets truncated to whatever remains.
// Returns table.Column slice ready for use.
func fitColumns(defs []columnDef, availableWidth int) []table.Column {
	const minTerminalCols = 60
	if availableWidth < minTerminalCols {
		cols := make([]table.Column, 0, 2)
		for i, d := range defs {
			if i >= 2 {
				break
			}
			w := availableWidth / 2
			if w < d.minWidth {
				w = d.minWidth
			}
			cols = append(cols, table.Column{Title: d.title, Width: w})
		}
		return cols
	}

	active := make([]bool, len(defs))
	for i := range active {
		active[i] = true
	}

	const cellPadPerCol = 2
	const sectionPadOuter = 4
	total := func() int {
		s := sectionPadOuter
		for i, d := range defs {
			if active[i] {
				s += d.width + cellPadPerCol
			}
		}
		return s
	}

	for total() > availableWidth {
		minPrio := -1
		minIdx := -1
		for i, d := range defs {
			if !active[i] {
				continue
			}
			if minPrio < 0 || d.priority < minPrio {
				minPrio = d.priority
				minIdx = i
			}
		}
		if minIdx < 0 {
			break
		}
		active[minIdx] = false
	}

	cols := make([]table.Column, 0, len(defs))
	for i, d := range defs {
		if active[i] {
			cols = append(cols, table.Column{Title: d.title, Width: d.width})
		}
	}
	return cols
}

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
		if len([]rune(toolStr)) > wTools-1 {
			toolStr = truncate(toolStr, wTools-1)
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
	m.tblActivity = newActivityTable(m.report, query, m.width)
	m.tblSess = newSessionsTable(m.report, query, m.width)
	m.tblProj = newProjectsTable(m.report, query, m.width)
}

// activityColDefs returns column definitions for the Activity table.
// Model token columns are inserted after Turns with a shared priority of 3.
// Priority: Date(7) > Total Tokens(6) > Prompts(5) > Turns(4) > model cols(3) > Cost is always kept last (but highest prio).
func activityColDefs(models []string) []columnDef {
	defs := []columnDef{
		{title: "Day", width: 12, minWidth: 10, priority: 7},
		{title: "Prompts", width: 8, minWidth: 6, priority: 5},
		{title: "Turns", width: 7, minWidth: 6, priority: 4},
	}
	for _, m := range models {
		defs = append(defs, columnDef{title: shortModel(m), width: 11, minWidth: 7, priority: 3})
	}
	defs = append(defs,
		columnDef{title: "Total Tokens", width: 13, minWidth: 8, priority: 6},
		columnDef{title: "Cost", width: 9, minWidth: 7, priority: 8},
	)
	return defs
}

var sessColDefs = []columnDef{
	{title: "Session", width: 10, minWidth: 8, priority: 8},
	{title: "Cost", width: 9, minWidth: 7, priority: 7},
	{title: "Prompts (msgs)", width: 15, minWidth: 8, priority: 6},
	{title: "Turns (API)", width: 12, minWidth: 7, priority: 5},
	{title: "Duration", width: 9, minWidth: 7, priority: 4},
	{title: "Started", width: 16, minWidth: 10, priority: 3},
	{title: "Project", width: 26, minWidth: 10, priority: 2},
	{title: "Total Tokens", width: 13, minWidth: 8, priority: 1},
}

var projColDefs = []columnDef{
	{title: "Project", width: 36, minWidth: 14, priority: 8},
	{title: "Cost", width: 10, minWidth: 7, priority: 7},
	{title: "Total Tokens", width: 13, minWidth: 8, priority: 6},
	{title: "Prompts (msgs)", width: 15, minWidth: 8, priority: 5},
	{title: "Turns (API)", width: 12, minWidth: 7, priority: 4},
	{title: "Input (tok)", width: 12, minWidth: 7, priority: 3},
	{title: "Cache Read (tok)", width: 17, minWidth: 8, priority: 2},
	{title: "Output (tok)", width: 13, minWidth: 7, priority: 1},
}

// newActivityTable: matrix Day × Model with totals row at bottom.
func newActivityTable(r stats.Report, query string, availWidth int) table.Model {
	defs := activityColDefs(r.Models)
	cols := fitColumns(defs, availWidth)

	visibleTitles := make(map[string]bool, len(cols))
	for _, c := range cols {
		visibleTitles[c.Title] = true
	}

	filteredDays := filterActivityRows(r.ByDay, query)

	rows := make([]table.Row, 0, len(filteredDays)+2)
	for _, d := range filteredDays {
		row := buildActivityRow(d, r.Models, r.Overall, defs, visibleTitles, false)
		rows = append(rows, row)
	}

	if len(filteredDays) > 0 {
		sep := make(table.Row, len(cols))
		for i, c := range cols {
			sep[i] = strings.Repeat("─", c.Width)
		}
		rows = append(rows, sep)

		totalDay := stats.DayStat{Day: "▸ TOTAL"}
		totalDay.Totals = r.Overall
		totalDay.ByModel = make(map[string]stats.Totals)
		for _, mm := range r.ByModel {
			totalDay.ByModel[mm.Model] = mm.Totals
		}
		rows = append(rows, buildActivityRow(totalDay, r.Models, r.Overall, defs, visibleTitles, true))
	}

	return makeTable(cols, rows)
}

func buildActivityRow(d stats.DayStat, models []string, overall stats.Totals, defs []columnDef, visible map[string]bool, isTotal bool) table.Row {
	dayLabel := d.Day
	if isTotal {
		dayLabel = "▸ TOTAL"
	}
	row := table.Row{}
	for _, def := range defs {
		if !visible[def.title] {
			continue
		}
		switch def.title {
		case "Day":
			row = append(row, dayLabel)
		case "Prompts":
			if isTotal {
				row = append(row, humanNumber(overall.Prompts))
			} else {
				row = append(row, humanNumber(d.Totals.Prompts))
			}
		case "Turns":
			if isTotal {
				row = append(row, humanNumber(overall.Turns))
			} else {
				row = append(row, humanNumber(d.Totals.Turns))
			}
		case "Total Tokens":
			if isTotal {
				row = append(row, compactNumber(overall.GrandTotal()))
			} else {
				row = append(row, compactNumber(d.Totals.GrandTotal()))
			}
		case "Cost":
			if isTotal {
				row = append(row, formatCost(overall.Cost))
			} else {
				row = append(row, formatCost(d.Totals.Cost))
			}
		default:
			found := false
			for _, m := range models {
				if shortModel(m) == def.title {
					found = true
					t := d.ByModel[m]
					if t.GrandTotal() == 0 {
						row = append(row, "—")
					} else {
						row = append(row, compactNumber(t.GrandTotal()))
					}
					break
				}
			}
			if !found {
				row = append(row, "")
			}
		}
	}
	return row
}

func newSessionsTable(r stats.Report, query string, availWidth int) table.Model {
	cols := fitColumns(sessColDefs, availWidth)

	visibleTitles := make(map[string]bool, len(cols))
	for _, c := range cols {
		visibleTitles[c.Title] = true
	}

	filteredSessions := filterSessionRows(r.BySession, query)
	rows := make([]table.Row, 0, len(filteredSessions))
	for _, s := range filteredSessions {
		dur := s.LastSeen.Sub(s.FirstSeen)
		row := table.Row{}
		for _, def := range sessColDefs {
			if !visibleTitles[def.title] {
				continue
			}
			switch def.title {
			case "Session":
				row = append(row, shortSession(s.SessionID))
			case "Cost":
				row = append(row, formatCost(s.Totals.Cost))
			case "Prompts (msgs)":
				row = append(row, humanNumber(s.Totals.Prompts))
			case "Turns (API)":
				row = append(row, humanNumber(s.Totals.Turns))
			case "Duration":
				row = append(row, formatDuration(dur))
			case "Started":
				row = append(row, s.FirstSeen.Local().Format("2006-01-02 15:04"))
			case "Project":
				row = append(row, truncate(shortenPath(s.Cwd), 26))
			case "Total Tokens":
				row = append(row, compactNumber(s.Totals.GrandTotal()))
			}
		}
		rows = append(rows, row)
	}
	return makeTable(cols, rows)
}

func newProjectsTable(r stats.Report, query string, availWidth int) table.Model {
	cols := fitColumns(projColDefs, availWidth)

	visibleTitles := make(map[string]bool, len(cols))
	for _, c := range cols {
		visibleTitles[c.Title] = true
	}

	filteredProjects := filterProjectRows(r.ByProject, query)
	rows := make([]table.Row, 0, len(filteredProjects))
	for _, p := range filteredProjects {
		row := table.Row{}
		for _, def := range projColDefs {
			if !visibleTitles[def.title] {
				continue
			}
			switch def.title {
			case "Project":
				row = append(row, truncate(p.Cwd, 36))
			case "Cost":
				row = append(row, formatCost(p.Totals.Cost))
			case "Total Tokens":
				row = append(row, compactNumber(p.Totals.GrandTotal()))
			case "Prompts (msgs)":
				row = append(row, humanNumber(p.Totals.Prompts))
			case "Turns (API)":
				row = append(row, humanNumber(p.Totals.Turns))
			case "Input (tok)":
				row = append(row, compactNumber(p.Totals.InputTokens))
			case "Cache Read (tok)":
				row = append(row, compactNumber(p.Totals.CacheReadTokens))
			case "Output (tok)":
				row = append(row, compactNumber(p.Totals.OutputTokens))
			}
		}
		rows = append(rows, row)
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
