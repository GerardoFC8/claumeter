// Package export serializes a stats.Report as JSON, CSV, or Markdown for
// consumption by scripts, dashboards, or human readers.
package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/GerardoFC8/claumeter/internal/stats"
)

const Schema = "v1"

// Payload is the canonical shape returned by ToJSON. Field names are stable.
type Payload struct {
	Schema    string         `json:"schema"`
	Generated time.Time      `json:"generated"`
	Range     RangeDTO       `json:"range"`
	Overall   TotalsDTO      `json:"overall"`
	ByDay     []DayDTO       `json:"by_day"`
	ByModel   []ModelDTO     `json:"by_model"`
	BySession []SessionDTO   `json:"by_session"`
	ByProject []ProjectDTO   `json:"by_project"`
	Tools     ToolsDTO       `json:"tools"`
}

type RangeDTO struct {
	Label string    `json:"label"`
	From  time.Time `json:"from"`
	To    time.Time `json:"to"`
}

type TotalsDTO struct {
	Prompts             int     `json:"prompts"`
	Turns               int     `json:"turns"`
	InputTokens         int     `json:"input_tokens"`
	CacheCreationTokens int     `json:"cache_creation_tokens"`
	CacheReadTokens     int     `json:"cache_read_tokens"`
	OutputTokens        int     `json:"output_tokens"`
	TotalTokens         int     `json:"total_tokens"`
	CostUSD             float64 `json:"cost_usd"`
}

type DayDTO struct {
	Day string    `json:"day"`
	TotalsDTO
	ByModel map[string]TotalsDTO `json:"by_model,omitempty"`
}

type ModelDTO struct {
	Model string `json:"model"`
	TotalsDTO
}

type SessionDTO struct {
	SessionID string    `json:"session_id"`
	Cwd       string    `json:"cwd"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	Models    []string  `json:"models"`
	TotalsDTO
}

type ProjectDTO struct {
	Cwd string `json:"cwd"`
	TotalsDTO
}

type ToolsDTO struct {
	Total    int           `json:"total"`
	Builtins []ToolEntryDTO `json:"builtins"`
	MCPs     []ToolEntryDTO `json:"mcps"`
	Servers  []ToolEntryDTO `json:"mcp_servers"`
	Skills   []ToolEntryDTO `json:"skills"`
	Agents   []ToolEntryDTO `json:"sub_agents"`
}

type ToolEntryDTO struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func buildPayload(label string, from, to time.Time, r stats.Report) Payload {
	p := Payload{
		Schema:    Schema,
		Generated: time.Now().UTC(),
		Range:     RangeDTO{Label: label, From: from, To: to},
		Overall:   totalsDTO(r.Overall),
		Tools:     toolsDTO(r.Tools),
	}
	for _, d := range r.ByDay {
		entry := DayDTO{Day: d.Day, TotalsDTO: totalsDTO(d.Totals)}
		if len(d.ByModel) > 0 {
			entry.ByModel = make(map[string]TotalsDTO, len(d.ByModel))
			for m, t := range d.ByModel {
				entry.ByModel[m] = totalsDTO(t)
			}
		}
		p.ByDay = append(p.ByDay, entry)
	}
	for _, m := range r.ByModel {
		p.ByModel = append(p.ByModel, ModelDTO{Model: m.Model, TotalsDTO: totalsDTO(m.Totals)})
	}
	for _, s := range r.BySession {
		models := make([]string, 0, len(s.Models))
		for m := range s.Models {
			models = append(models, m)
		}
		p.BySession = append(p.BySession, SessionDTO{
			SessionID: s.SessionID,
			Cwd:       s.Cwd,
			FirstSeen: s.FirstSeen,
			LastSeen:  s.LastSeen,
			Models:    models,
			TotalsDTO: totalsDTO(s.Totals),
		})
	}
	for _, pr := range r.ByProject {
		p.ByProject = append(p.ByProject, ProjectDTO{Cwd: pr.Cwd, TotalsDTO: totalsDTO(pr.Totals)})
	}
	return p
}

func totalsDTO(t stats.Totals) TotalsDTO {
	return TotalsDTO{
		Prompts:             t.Prompts,
		Turns:               t.Turns,
		InputTokens:         t.InputTokens,
		CacheCreationTokens: t.CacheCreationTokens,
		CacheReadTokens:     t.CacheReadTokens,
		OutputTokens:        t.OutputTokens,
		TotalTokens:         t.GrandTotal(),
		CostUSD:             round2(t.Cost),
	}
}

func toolsDTO(t stats.ToolStats) ToolsDTO {
	return ToolsDTO{
		Total:    t.Total,
		Builtins: toolEntries(t.Builtins),
		MCPs:     toolEntries(t.MCPs),
		Servers:  toolEntries(t.Servers),
		Skills:   toolEntries(t.Skills),
		Agents:   toolEntries(t.Agents),
	}
}

func toolEntries(in []stats.ToolEntry) []ToolEntryDTO {
	out := make([]ToolEntryDTO, 0, len(in))
	for _, e := range in {
		out = append(out, ToolEntryDTO{Name: e.Name, Count: e.Count})
	}
	return out
}

// ToJSON writes a full JSON dump.
func ToJSON(w io.Writer, label string, from, to time.Time, r stats.Report) error {
	p := buildPayload(label, from, to, r)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}

// ToCSV writes per-day rows as CSV.
func ToCSV(w io.Writer, r stats.Report) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	header := []string{
		"day", "prompts", "turns",
		"input_tokens", "cache_creation_tokens", "cache_read_tokens", "output_tokens",
		"total_tokens", "cost_usd",
	}
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, d := range r.ByDay {
		row := []string{
			d.Day,
			strconv.Itoa(d.Totals.Prompts),
			strconv.Itoa(d.Totals.Turns),
			strconv.Itoa(d.Totals.InputTokens),
			strconv.Itoa(d.Totals.CacheCreationTokens),
			strconv.Itoa(d.Totals.CacheReadTokens),
			strconv.Itoa(d.Totals.OutputTokens),
			strconv.Itoa(d.Totals.GrandTotal()),
			strconv.FormatFloat(round2(d.Totals.Cost), 'f', 2, 64),
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// ToMarkdown writes a human-readable markdown report with overall + by-day + by-model.
func ToMarkdown(w io.Writer, label string, r stats.Report) error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# claumeter usage report\n\n")
	fmt.Fprintf(&b, "_Range: %s · %d prompts · %d turns · %s tokens · $%.2f_\n\n",
		label, r.Overall.Prompts, r.Overall.Turns, humanInt(r.Overall.GrandTotal()), r.Overall.Cost)

	fmt.Fprintf(&b, "## By day\n\n")
	fmt.Fprintf(&b, "| Day | Prompts | Turns | Tokens | Cost |\n")
	fmt.Fprintf(&b, "|---|---:|---:|---:|---:|\n")
	for _, d := range r.ByDay {
		fmt.Fprintf(&b, "| %s | %d | %d | %s | $%.2f |\n",
			d.Day, d.Totals.Prompts, d.Totals.Turns,
			humanInt(d.Totals.GrandTotal()), d.Totals.Cost,
		)
	}

	fmt.Fprintf(&b, "\n## By model\n\n")
	fmt.Fprintf(&b, "| Model | Turns | Tokens | Cost |\n")
	fmt.Fprintf(&b, "|---|---:|---:|---:|\n")
	for _, m := range r.ByModel {
		fmt.Fprintf(&b, "| %s | %d | %s | $%.2f |\n",
			m.Model, m.Totals.Turns,
			humanInt(m.Totals.GrandTotal()), m.Totals.Cost,
		)
	}

	fmt.Fprintf(&b, "\n## By project\n\n")
	fmt.Fprintf(&b, "| Project | Prompts | Turns | Tokens | Cost |\n")
	fmt.Fprintf(&b, "|---|---:|---:|---:|---:|\n")
	for _, p := range r.ByProject {
		fmt.Fprintf(&b, "| %s | %d | %d | %s | $%.2f |\n",
			p.Cwd, p.Totals.Prompts, p.Totals.Turns,
			humanInt(p.Totals.GrandTotal()), p.Totals.Cost,
		)
	}

	_, err := w.Write(b.Bytes())
	return err
}

// SessionDetailDTO is the JSON shape returned by GET /session/:id.
type SessionDetailDTO struct {
	SessionID string           `json:"session_id"`
	Cwd       string           `json:"cwd"`
	FirstSeen time.Time        `json:"first_seen"`
	LastSeen  time.Time        `json:"last_seen"`
	Models    []string         `json:"models"`
	TotalsDTO                  // flattened at top level
	Turns     []SessionTurnDTO `json:"turns"`
}

// SessionTurnDTO is one assistant turn inside a session.
type SessionTurnDTO struct {
	Timestamp time.Time `json:"timestamp"`
	Model     string    `json:"model"`
	TotalsDTO            // flattened
	Tools     []string  `json:"tools,omitempty"`
}

// NewSessionDetail converts a stats.SessionDetail to its DTO form.
func NewSessionDetail(sd stats.SessionDetail) SessionDetailDTO {
	turns := make([]SessionTurnDTO, 0, len(sd.Turns))
	for _, t := range sd.Turns {
		turns = append(turns, SessionTurnDTO{
			Timestamp: t.Timestamp,
			Model:     t.Model,
			TotalsDTO: totalsDTO(t.Totals),
			Tools:     t.Tools,
		})
	}
	return SessionDetailDTO{
		SessionID: sd.SessionID,
		Cwd:       sd.Cwd,
		FirstSeen: sd.FirstSeen,
		LastSeen:  sd.LastSeen,
		Models:    sd.Models,
		TotalsDTO: totalsDTO(sd.Totals),
		Turns:     turns,
	}
}

// --- Comparison DTOs ---

// ComparisonSideDTO describes one side (A or B) of a comparison.
type ComparisonSideDTO struct {
	Label string    `json:"label"`
	From  time.Time `json:"from,omitempty"`
	To    time.Time `json:"to,omitempty"`
	TotalsDTO
}

// DeltaDTO is the wire shape for a single metric comparison.
type DeltaDTO struct {
	A     float64 `json:"a"`
	B     float64 `json:"b"`
	Delta float64 `json:"delta"`
	Pct   float64 `json:"pct"`
	Gt    bool    `json:"gt"`
}

// ComparisonDeltasDTO holds one DeltaDTO per tracked metric.
type ComparisonDeltasDTO struct {
	Prompts             DeltaDTO `json:"prompts"`
	Turns               DeltaDTO `json:"turns"`
	InputTokens         DeltaDTO `json:"input_tokens"`
	CacheCreationTokens DeltaDTO `json:"cache_creation_tokens"`
	CacheReadTokens     DeltaDTO `json:"cache_read_tokens"`
	OutputTokens        DeltaDTO `json:"output_tokens"`
	TotalTokens         DeltaDTO `json:"total_tokens"`
	CostUSD             DeltaDTO `json:"cost_usd"`
}

// ComparisonPayload is the top-level JSON shape returned by NewComparison and
// the GET /compare endpoint.
type ComparisonPayload struct {
	Schema    string              `json:"schema"`
	Generated time.Time           `json:"generated"`
	A         ComparisonSideDTO   `json:"a"`
	B         ComparisonSideDTO   `json:"b"`
	Deltas    ComparisonDeltasDTO `json:"deltas"`
}

func deltaDTO(d stats.Delta) DeltaDTO {
	return DeltaDTO{
		A:     d.A,
		B:     d.B,
		Delta: d.Delta,
		Pct:   d.Pct,
		Gt:    d.Gt,
	}
}

// NewComparison constructs a ComparisonPayload from two resolved range sides and
// the Comparison produced by stats.Compare.
func NewComparison(
	aLabel string, aFrom, aTo time.Time, aTotals stats.Totals,
	bLabel string, bFrom, bTo time.Time, bTotals stats.Totals,
	cmp stats.Comparison,
) ComparisonPayload {
	return ComparisonPayload{
		Schema:    Schema,
		Generated: time.Now().UTC(),
		A: ComparisonSideDTO{
			Label:     aLabel,
			From:      aFrom,
			To:        aTo,
			TotalsDTO: totalsDTO(aTotals),
		},
		B: ComparisonSideDTO{
			Label:     bLabel,
			From:      bFrom,
			To:        bTo,
			TotalsDTO: totalsDTO(bTotals),
		},
		Deltas: ComparisonDeltasDTO{
			Prompts:             deltaDTO(cmp.Prompts),
			Turns:               deltaDTO(cmp.Turns),
			InputTokens:         deltaDTO(cmp.InputTokens),
			CacheCreationTokens: deltaDTO(cmp.CacheCreationTokens),
			CacheReadTokens:     deltaDTO(cmp.CacheReadTokens),
			OutputTokens:        deltaDTO(cmp.OutputTokens),
			TotalTokens:         deltaDTO(cmp.TotalTokens),
			CostUSD:             deltaDTO(cmp.CostUSD),
		},
	}
}

func round2(f float64) float64 { return float64(int64(f*100+0.5)) / 100 }

func humanInt(n int) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.2fB", float64(n)/1_000_000_000)
	case n >= 1_000_000:
		return fmt.Sprintf("%.2fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return strconv.Itoa(n)
	}
}
