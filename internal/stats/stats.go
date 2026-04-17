package stats

import (
	"sort"
	"time"

	"github.com/GerardoFC8/claumeter/internal/pricing"
	"github.com/GerardoFC8/claumeter/internal/usage"
)

type Totals struct {
	InputTokens         int
	OutputTokens        int
	CacheCreationTokens int
	CacheReadTokens     int
	Turns               int
	Prompts             int
	Cost                float64 // USD, summed per-event
}

func (t Totals) TotalInput() int { return t.InputTokens + t.CacheCreationTokens + t.CacheReadTokens }
func (t Totals) GrandTotal() int { return t.TotalInput() + t.OutputTokens }
func (t *Totals) addEvent(e usage.Event) {
	t.InputTokens += e.InputTokens
	t.OutputTokens += e.OutputTokens
	t.CacheCreationTokens += e.CacheCreationTokens
	t.CacheReadTokens += e.CacheReadTokens
	t.Turns++
	t.Cost += costForEvent(e)
}
func (t *Totals) addPrompt() { t.Prompts++ }

// costForEvent: if the JSONL lacks the 5m/1h breakdown, treat all
// cache-creation tokens as 1h (upper bound, the dominant pattern in Claude
// Code sessions).
func costForEvent(e usage.Event) float64 {
	u := pricing.Usage{
		InputTokens:     e.InputTokens,
		CacheReadTokens: e.CacheReadTokens,
		OutputTokens:    e.OutputTokens,
	}
	if e.CacheCreation5mTokens+e.CacheCreation1hTokens == 0 {
		u.CacheCreation1hTokens = e.CacheCreationTokens
	} else {
		u.CacheCreation5mTokens = e.CacheCreation5mTokens
		u.CacheCreation1hTokens = e.CacheCreation1hTokens
	}
	return pricing.Cost(e.Model, u)
}

type DayStat struct {
	Day     string
	Totals  Totals
	ByModel map[string]Totals
}

type ModelStat struct {
	Model  string
	Totals Totals
}

type SessionStat struct {
	SessionID string
	Cwd       string
	FirstSeen time.Time
	LastSeen  time.Time
	Models    map[string]struct{}
	Totals    Totals
}

type ProjectStat struct {
	Cwd    string
	Totals Totals
}

type ToolEntry struct {
	Name  string
	Count int
}

type ToolStats struct {
	Builtins []ToolEntry
	MCPs     []ToolEntry // aggregated per tool name ("server/tool")
	Servers  []ToolEntry // aggregated per MCP server
	Skills   []ToolEntry
	Agents   []ToolEntry
	Total    int
}

type Report struct {
	Overall   Totals
	ByDay     []DayStat
	ByModel   []ModelStat
	BySession []SessionStat
	ByProject []ProjectStat
	Tools     ToolStats
	Models    []string // stable list of model names seen (sorted by total usage desc)
	DateRange [2]time.Time
}

func Build(data usage.Data) Report {
	r := Report{}
	if len(data.Events) == 0 && len(data.Prompts) == 0 && len(data.ToolUses) == 0 {
		return r
	}

	byDay := map[string]*DayStat{}
	byModel := map[string]*Totals{}
	bySession := map[string]*SessionStat{}
	byProject := map[string]*Totals{}

	var minT, maxT time.Time
	setRange := func(t time.Time) {
		if minT.IsZero() || t.Before(minT) {
			minT = t
		}
		if maxT.IsZero() || t.After(maxT) {
			maxT = t
		}
	}

	ensureSession := func(id, cwd string, ts time.Time) *SessionStat {
		s, ok := bySession[id]
		if !ok {
			s = &SessionStat{
				SessionID: id,
				Cwd:       cwd,
				FirstSeen: ts,
				LastSeen:  ts,
				Models:    map[string]struct{}{},
			}
			bySession[id] = s
			return s
		}
		if ts.Before(s.FirstSeen) {
			s.FirstSeen = ts
		}
		if ts.After(s.LastSeen) {
			s.LastSeen = ts
		}
		return s
	}

	ensureDay := func(d string) *DayStat {
		if ds, ok := byDay[d]; ok {
			return ds
		}
		ds := &DayStat{Day: d, ByModel: map[string]Totals{}}
		byDay[d] = ds
		return ds
	}

	for _, e := range data.Events {
		r.Overall.addEvent(e)

		ds := ensureDay(e.Day())
		ds.Totals.addEvent(e)
		mt := ds.ByModel[e.Model]
		mt.addEvent(e)
		ds.ByModel[e.Model] = mt

		if _, ok := byModel[e.Model]; !ok {
			byModel[e.Model] = &Totals{}
		}
		byModel[e.Model].addEvent(e)

		if _, ok := byProject[e.Cwd]; !ok {
			byProject[e.Cwd] = &Totals{}
		}
		byProject[e.Cwd].addEvent(e)

		s := ensureSession(e.SessionID, e.Cwd, e.Timestamp)
		s.Models[e.Model] = struct{}{}
		s.Totals.addEvent(e)

		setRange(e.Timestamp)
	}

	for _, p := range data.Prompts {
		r.Overall.addPrompt()

		ds := ensureDay(p.Day())
		ds.Totals.addPrompt()

		if _, ok := byProject[p.Cwd]; !ok {
			byProject[p.Cwd] = &Totals{}
		}
		byProject[p.Cwd].addPrompt()

		s := ensureSession(p.SessionID, p.Cwd, p.Timestamp)
		s.Totals.addPrompt()

		setRange(p.Timestamp)
	}

	r.Tools = buildToolStats(data.ToolUses)

	r.DateRange = [2]time.Time{minT, maxT}

	for _, ds := range byDay {
		r.ByDay = append(r.ByDay, *ds)
	}
	sort.Slice(r.ByDay, func(i, j int) bool { return r.ByDay[i].Day > r.ByDay[j].Day })

	for m, t := range byModel {
		r.ByModel = append(r.ByModel, ModelStat{Model: m, Totals: *t})
	}
	sort.Slice(r.ByModel, func(i, j int) bool {
		return r.ByModel[i].Totals.GrandTotal() > r.ByModel[j].Totals.GrandTotal()
	})
	r.Models = make([]string, 0, len(r.ByModel))
	for _, mm := range r.ByModel {
		if mm.Totals.GrandTotal() == 0 {
			continue
		}
		r.Models = append(r.Models, mm.Model)
	}

	for _, s := range bySession {
		r.BySession = append(r.BySession, *s)
	}
	sort.Slice(r.BySession, func(i, j int) bool {
		return r.BySession[i].LastSeen.After(r.BySession[j].LastSeen)
	})

	for p, t := range byProject {
		r.ByProject = append(r.ByProject, ProjectStat{Cwd: p, Totals: *t})
	}
	sort.Slice(r.ByProject, func(i, j int) bool {
		return r.ByProject[i].Totals.GrandTotal() > r.ByProject[j].Totals.GrandTotal()
	})

	return r
}

func buildToolStats(uses []usage.ToolUse) ToolStats {
	ts := ToolStats{Total: len(uses)}
	builtins := map[string]int{}
	mcpTools := map[string]int{}
	servers := map[string]int{}
	skills := map[string]int{}
	agents := map[string]int{}

	for _, u := range uses {
		switch u.Kind {
		case usage.ToolBuiltin:
			builtins[u.Name]++
		case usage.ToolMCP:
			mcpTools[u.MCPServer+"/"+u.Target]++
			servers[u.MCPServer]++
		case usage.ToolSkill:
			label := u.Target
			if label == "" {
				label = "(unknown)"
			}
			skills[label]++
		case usage.ToolAgent:
			label := u.Target
			if label == "" {
				label = "(unknown)"
			}
			agents[label]++
		}
	}

	ts.Builtins = sortEntries(builtins)
	ts.MCPs = sortEntries(mcpTools)
	ts.Servers = sortEntries(servers)
	ts.Skills = sortEntries(skills)
	ts.Agents = sortEntries(agents)
	return ts
}

func sortEntries(m map[string]int) []ToolEntry {
	out := make([]ToolEntry, 0, len(m))
	for k, v := range m {
		out = append(out, ToolEntry{Name: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	return out
}
