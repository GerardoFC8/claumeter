package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/GerardoFC8/claumeter/internal/stats"
)

// searchState holds the textinput and active flag for the search bar.
type searchState struct {
	input  textinput.Model
	active bool
}

func newSearchState() searchState {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Prompt = "search: "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(colorMuted)
	ti.TextStyle = lipgloss.NewStyle().Foreground(colorFg)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(colorAccent)
	ti.CharLimit = 120
	return searchState{input: ti}
}

// query returns the current search string, lowercased for comparison.
func (s searchState) query() string {
	return strings.ToLower(s.input.Value())
}

// renderBar renders the search bar line (no trailing newline).
// Call this only when active == true.
func (s searchState) renderBar(width int) string {
	bar := lipgloss.NewStyle().
		Foreground(colorMuted).
		Width(width).
		Render(s.input.View())
	return bar
}

// applyFilter returns only the rows whose field values (returned by fields())
// contain query as a case-insensitive substring. Empty query returns all rows.
func applyFilter(query string, rows [][]string, fields func(row []string) []string) [][]string {
	if query == "" {
		return rows
	}
	out := make([][]string, 0, len(rows))
	for _, r := range rows {
		for _, f := range fields(r) {
			if strings.Contains(strings.ToLower(f), query) {
				out = append(out, r)
				break
			}
		}
	}
	return out
}

// filterActivityRows filters DayStat slice by query (day string + model names).
func filterActivityRows(days []stats.DayStat, query string) []stats.DayStat {
	if query == "" {
		return days
	}
	out := make([]stats.DayStat, 0, len(days))
	for _, d := range days {
		if strings.Contains(strings.ToLower(d.Day), query) {
			out = append(out, d)
			continue
		}
		for model := range d.ByModel {
			if strings.Contains(strings.ToLower(model), query) {
				out = append(out, d)
				break
			}
		}
	}
	return out
}

// filterSessionRows filters SessionStat slice by query (session_id, cwd, model names).
func filterSessionRows(sessions []stats.SessionStat, query string) []stats.SessionStat {
	if query == "" {
		return sessions
	}
	out := make([]stats.SessionStat, 0, len(sessions))
	for _, s := range sessions {
		if strings.Contains(strings.ToLower(s.SessionID), query) ||
			strings.Contains(strings.ToLower(s.Cwd), query) {
			out = append(out, s)
			continue
		}
		for model := range s.Models {
			if strings.Contains(strings.ToLower(model), query) {
				out = append(out, s)
				break
			}
		}
	}
	return out
}

// filterProjectRows filters ProjectStat slice by query (cwd).
func filterProjectRows(projects []stats.ProjectStat, query string) []stats.ProjectStat {
	if query == "" {
		return projects
	}
	out := make([]stats.ProjectStat, 0, len(projects))
	for _, p := range projects {
		if strings.Contains(strings.ToLower(p.Cwd), query) {
			out = append(out, p)
		}
	}
	return out
}

// filterToolEntries filters a []ToolEntry slice by query (name).
func filterToolEntries(entries []stats.ToolEntry, query string) []stats.ToolEntry {
	if query == "" {
		return entries
	}
	out := make([]stats.ToolEntry, 0, len(entries))
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Name), query) {
			out = append(out, e)
		}
	}
	return out
}

// filterToolStats applies the query to all four ToolStats slices.
func filterToolStats(ts stats.ToolStats, query string) stats.ToolStats {
	if query == "" {
		return ts
	}
	return stats.ToolStats{
		Total:    ts.Total, // keep original total; we're filtering display only
		Builtins: filterToolEntries(ts.Builtins, query),
		MCPs:     filterToolEntries(ts.MCPs, query),
		Servers:  filterToolEntries(ts.Servers, query),
		Skills:   filterToolEntries(ts.Skills, query),
		Agents:   filterToolEntries(ts.Agents, query),
	}
}
