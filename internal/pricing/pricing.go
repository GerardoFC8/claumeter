// Package pricing applies Anthropic's published per-million-token rates to
// parsed usage events and returns costs in USD. Rates are snapshots of
// https://platform.claude.com/docs/en/about-claude/pricing as of 2026-04.
package pricing

import (
	"regexp"
)

// Rate holds USD per 1M tokens for a model, across all billing categories.
type Rate struct {
	Input        float64
	CacheWrite5m float64
	CacheWrite1h float64
	CacheRead    float64
	Output       float64
}

// rates: canonical model IDs mapped to per-million-token USD rates.
var rates = map[string]Rate{
	// Opus family — 4.5, 4.6, 4.7 share pricing.
	"claude-opus-4-7": {Input: 5, CacheWrite5m: 6.25, CacheWrite1h: 10, CacheRead: 0.50, Output: 25},
	"claude-opus-4-6": {Input: 5, CacheWrite5m: 6.25, CacheWrite1h: 10, CacheRead: 0.50, Output: 25},
	"claude-opus-4-5": {Input: 5, CacheWrite5m: 6.25, CacheWrite1h: 10, CacheRead: 0.50, Output: 25},
	"claude-opus-4-1": {Input: 15, CacheWrite5m: 18.75, CacheWrite1h: 30, CacheRead: 1.50, Output: 75},
	"claude-opus-4":   {Input: 15, CacheWrite5m: 18.75, CacheWrite1h: 30, CacheRead: 1.50, Output: 75},

	// Sonnet family — 4, 4.5, 4.6 share pricing.
	"claude-sonnet-4-6": {Input: 3, CacheWrite5m: 3.75, CacheWrite1h: 6, CacheRead: 0.30, Output: 15},
	"claude-sonnet-4-5": {Input: 3, CacheWrite5m: 3.75, CacheWrite1h: 6, CacheRead: 0.30, Output: 15},
	"claude-sonnet-4":   {Input: 3, CacheWrite5m: 3.75, CacheWrite1h: 6, CacheRead: 0.30, Output: 15},

	// Haiku.
	"claude-haiku-4-5": {Input: 1, CacheWrite5m: 1.25, CacheWrite1h: 2, CacheRead: 0.10, Output: 5},
	"claude-haiku-3-5": {Input: 0.80, CacheWrite5m: 1, CacheWrite1h: 1.6, CacheRead: 0.08, Output: 4},
	"claude-haiku-3":   {Input: 0.25, CacheWrite5m: 0.30, CacheWrite1h: 0.50, CacheRead: 0.03, Output: 1.25},
}

var dateSuffix = regexp.MustCompile(`-\d{8}$`)

// RateFor resolves a model ID to its Rate, tolerating dated variants such as
// "claude-haiku-4-5-20251001" by stripping a trailing -YYYYMMDD suffix.
func RateFor(model string) (Rate, bool) {
	if r, ok := rates[model]; ok {
		return r, true
	}
	stripped := dateSuffix.ReplaceAllString(model, "")
	if r, ok := rates[stripped]; ok {
		return r, true
	}
	return Rate{}, false
}

// Usage is the per-request token breakdown used to compute cost.
type Usage struct {
	InputTokens           int
	CacheCreation5mTokens int
	CacheCreation1hTokens int
	CacheReadTokens       int
	OutputTokens          int
}

// Cost returns USD for a given model + usage. Unknown models return 0.
func Cost(model string, u Usage) float64 {
	r, ok := RateFor(model)
	if !ok {
		return 0
	}
	const million = 1_000_000.0
	return (float64(u.InputTokens)*r.Input+
		float64(u.CacheCreation5mTokens)*r.CacheWrite5m+
		float64(u.CacheCreation1hTokens)*r.CacheWrite1h+
		float64(u.CacheReadTokens)*r.CacheRead+
		float64(u.OutputTokens)*r.Output) / million
}

// Known reports whether a model is priced in the table.
func Known(model string) bool {
	_, ok := RateFor(model)
	return ok
}
