package stats

import (
	"sort"
	"time"

	"github.com/GerardoFC8/claumeter/internal/pricing"
	"github.com/GerardoFC8/claumeter/internal/usage"
)

type Totals struct {
	InputTokens           int
	OutputTokens          int
	CacheCreationTokens   int // total = 5m + 1h (kept for backwards compat)
	CacheCreation5mTokens int
	CacheCreation1hTokens int
	CacheReadTokens       int
	Turns                 int
	Prompts               int
	Cost                  float64 // USD, summed per-event
}

func (t Totals) TotalInput() int { return t.InputTokens + t.CacheCreationTokens + t.CacheReadTokens }
func (t Totals) GrandTotal() int { return t.TotalInput() + t.OutputTokens }
func (t *Totals) addEvent(e usage.Event) {
	t.InputTokens += e.InputTokens
	t.OutputTokens += e.OutputTokens
	t.CacheCreationTokens += e.CacheCreationTokens
	t.CacheReadTokens += e.CacheReadTokens
	// Split 5m/1h: if JSONL lacks breakdown, attribute the whole creation to 1h
	// (upper-bound assumption, matches costForEvent).
	if e.CacheCreation5mTokens+e.CacheCreation1hTokens == 0 {
		t.CacheCreation1hTokens += e.CacheCreationTokens
	} else {
		t.CacheCreation5mTokens += e.CacheCreation5mTokens
		t.CacheCreation1hTokens += e.CacheCreation1hTokens
	}
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

// CostBucket is a per-kind cost row decorated with percent-of-total.
type CostBucket struct {
	Kind   string  // "input" | "cache_write_5m" | "cache_write_1h" | "cache_read" | "output"
	Tokens int
	Rate   float64
	Cost   float64
	Pct    float64 // 0..100, share of total cost (the parent ModelBreakdown.TotalCost or overall)
}

// ModelBreakdown holds the 5-bucket breakdown for a single model.
type ModelBreakdown struct {
	Model     string
	Buckets   []CostBucket
	TotalCost float64
	Pct       float64 // share of grand total across all models
}

// CostBreakdownReport aggregates per-model breakdowns plus an overall row.
type CostBreakdownReport struct {
	Overall   ModelBreakdown // Model = "overall"
	ByModel   []ModelBreakdown
	TotalCost float64
}

// BuildCostBreakdown derives the 5-bucket breakdown per model (sorted by cost
// descending) plus an "overall" aggregate.
func BuildCostBreakdown(r Report) CostBreakdownReport {
	out := CostBreakdownReport{TotalCost: r.Overall.Cost}

	toUsage := func(t Totals) pricing.Usage {
		return pricing.Usage{
			InputTokens:           t.InputTokens,
			CacheCreation5mTokens: t.CacheCreation5mTokens,
			CacheCreation1hTokens: t.CacheCreation1hTokens,
			CacheReadTokens:       t.CacheReadTokens,
			OutputTokens:          t.OutputTokens,
		}
	}

	buildOne := func(model string, t Totals, parentCost float64) ModelBreakdown {
		raw := pricing.Breakdown(model, toUsage(t))
		buckets := make([]CostBucket, 0, len(raw))
		var total float64
		for _, b := range raw {
			total += b.Cost
		}
		for _, b := range raw {
			pct := 0.0
			if total > 0 {
				pct = b.Cost / total * 100
			}
			buckets = append(buckets, CostBucket{
				Kind:   b.Kind,
				Tokens: b.Tokens,
				Rate:   b.Rate,
				Cost:   b.Cost,
				Pct:    pct,
			})
		}
		share := 0.0
		if parentCost > 0 {
			share = total / parentCost * 100
		}
		return ModelBreakdown{
			Model:     model,
			Buckets:   buckets,
			TotalCost: total,
			Pct:       share,
		}
	}

	for _, m := range r.ByModel {
		out.ByModel = append(out.ByModel, buildOne(m.Model, m.Totals, r.Overall.Cost))
	}
	sort.Slice(out.ByModel, func(i, j int) bool {
		return out.ByModel[i].TotalCost > out.ByModel[j].TotalCost
	})

	// Overall = sum of each bucket kind across all per-model breakdowns.
	kinds := []string{"input", "cache_write_5m", "cache_write_1h", "cache_read", "output"}
	idx := map[string]int{}
	for i, k := range kinds {
		idx[k] = i
	}
	agg := make([]CostBucket, len(kinds))
	for i, k := range kinds {
		agg[i].Kind = k
	}
	var overallCost float64
	for _, mb := range out.ByModel {
		for _, b := range mb.Buckets {
			i := idx[b.Kind]
			agg[i].Tokens += b.Tokens
			agg[i].Cost += b.Cost
			overallCost += b.Cost
		}
	}
	for i := range agg {
		if overallCost > 0 {
			agg[i].Pct = agg[i].Cost / overallCost * 100
		}
	}
	out.Overall = ModelBreakdown{
		Model:     "overall",
		Buckets:   agg,
		TotalCost: overallCost,
		Pct:       100,
	}
	return out
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
	Overall        Totals
	ByDay          []DayStat
	ByModel        []ModelStat
	BySession      []SessionStat
	ByProject      []ProjectStat
	Tools          ToolStats
	Models         []string // stable list of model names seen (sorted by total usage desc)
	DateRange      [2]time.Time
	PromptsByHour  [24]int // prompts per hour-of-day aggregated across all filtered data
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
		r.PromptsByHour[p.Timestamp.Local().Hour()]++
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
