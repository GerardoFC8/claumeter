package stats

// Delta holds the A value, B value, and derived diff metrics between two ranges
// for a single metric. When A is zero, Pct is 0.0 (division-by-zero guard).
type Delta struct {
	A     float64 // value for range A
	B     float64 // value for range B
	Delta float64 // B - A (absolute; negative means B < A)
	Pct   float64 // (B - A) / A * 100; 0.0 when A == 0
	Gt    bool    // true when B > A strictly
}

func newDelta(a, b float64) Delta {
	d := Delta{
		A:     a,
		B:     b,
		Delta: b - a,
		Gt:    b > a,
	}
	if a != 0 {
		d.Pct = (b - a) / a * 100
	}
	return d
}

// Comparison holds per-metric Deltas comparing two Totals snapshots.
type Comparison struct {
	Prompts             Delta
	Turns               Delta
	InputTokens         Delta
	CacheCreationTokens Delta
	CacheReadTokens     Delta
	OutputTokens        Delta
	TotalTokens         Delta
	CostUSD             Delta
}

// Compare builds a Comparison from two Totals (a = baseline, b = current).
func Compare(a, b Totals) Comparison {
	return Comparison{
		Prompts:             newDelta(float64(a.Prompts), float64(b.Prompts)),
		Turns:               newDelta(float64(a.Turns), float64(b.Turns)),
		InputTokens:         newDelta(float64(a.InputTokens), float64(b.InputTokens)),
		CacheCreationTokens: newDelta(float64(a.CacheCreationTokens), float64(b.CacheCreationTokens)),
		CacheReadTokens:     newDelta(float64(a.CacheReadTokens), float64(b.CacheReadTokens)),
		OutputTokens:        newDelta(float64(a.OutputTokens), float64(b.OutputTokens)),
		TotalTokens:         newDelta(float64(a.GrandTotal()), float64(b.GrandTotal())),
		CostUSD:             newDelta(a.Cost, b.Cost),
	}
}
