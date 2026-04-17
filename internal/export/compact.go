package export

import (
	"time"

	"github.com/GerardoFC8/claumeter/internal/stats"
)

// CompactPayload is the shape returned by `claumeter today` / `week` / `range`
// and the `/today` HTTP endpoint. Keep field names stable — widgets depend on
// them.
type CompactPayload struct {
	Range    string           `json:"range"`
	From     string           `json:"from,omitempty"`
	To       string           `json:"to,omitempty"`
	Prompts  int              `json:"prompts"`
	Turns    int              `json:"turns"`
	Tokens   int              `json:"tokens"`
	Cost     float64          `json:"cost_usd"`
	TopModel string           `json:"top_model,omitempty"`
	ByModel  []ByModelCompact `json:"by_model,omitempty"`
}

type ByModelCompact struct {
	Model  string  `json:"model"`
	Turns  int     `json:"turns"`
	Tokens int     `json:"tokens"`
	Cost   float64 `json:"cost_usd"`
}

// NewCompact builds a CompactPayload for the given filtered report. `label` is
// the human-readable range name (e.g. "Today", "This week", "2026-04-01 → 2026-04-17").
// `from` / `to` are the [from, to) window; zero values are omitted from the JSON.
func NewCompact(label string, from, to time.Time, r stats.Report) CompactPayload {
	p := CompactPayload{
		Range:   label,
		Prompts: r.Overall.Prompts,
		Turns:   r.Overall.Turns,
		Tokens:  r.Overall.GrandTotal(),
		Cost:    round2(r.Overall.Cost),
	}
	if !from.IsZero() {
		p.From = from.Format(time.RFC3339)
	}
	if !to.IsZero() {
		p.To = to.Format(time.RFC3339)
	}
	if len(r.ByModel) > 0 {
		p.TopModel = r.ByModel[0].Model
		for _, m := range r.ByModel {
			p.ByModel = append(p.ByModel, ByModelCompact{
				Model:  m.Model,
				Turns:  m.Totals.Turns,
				Tokens: m.Totals.GrandTotal(),
				Cost:   round2(m.Totals.Cost),
			})
		}
	}
	return p
}
