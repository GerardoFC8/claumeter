// Package quota estimates rate-limit status for a given Claude Code plan.
// Anthropic does not publish quota APIs for Claude Code; all limits here
// are unofficial estimates derived from community reports (early 2026).
// Users can tune individual plan limits via environment variables without
// recompiling.
package quota

import (
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GerardoFC8/claumeter/internal/usage"
)

// PlanLimit describes an approximate rate-limit for a Claude Code plan.
type PlanLimit struct {
	Plan               string
	MessagesPerWindow  int
	Window             time.Duration
	Description        string // e.g. "Claude Pro (~45 msgs / 5h, estimated)"
}

// plans holds the default known plan limits.
// Keys are canonical lower-case plan names.
var plans = map[string]PlanLimit{
	"pro": {
		Plan:              "pro",
		MessagesPerWindow: 45,
		Window:            5 * time.Hour,
		Description:       "Claude Pro (~45 msgs / 5h, estimated)",
	},
	"max-5x": {
		Plan:              "max-5x",
		MessagesPerWindow: 225,
		Window:            5 * time.Hour,
		Description:       "Claude Max 5x (~225 msgs / 5h, estimated)",
	},
	"max-20x": {
		Plan:              "max-20x",
		MessagesPerWindow: 900,
		Window:            5 * time.Hour,
		Description:       "Claude Max 20x (~900 msgs / 5h, estimated)",
	},
}

// envOverrideKeys maps canonical plan keys to the env var that overrides
// MessagesPerWindow.
var envOverrideKeys = map[string]string{
	"pro":     "CLAUMETER_QUOTA_PRO_MESSAGES",
	"max-5x":  "CLAUMETER_QUOTA_MAX5X_MESSAGES",
	"max-20x": "CLAUMETER_QUOTA_MAX20X_MESSAGES",
}

// Plans returns the sorted list of known plan names.
func Plans() []string {
	keys := make([]string, 0, len(plans))
	for k := range plans {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// LookupPlan returns the PlanLimit for the given plan name (case-insensitive).
// Environment variable overrides are applied to MessagesPerWindow when set.
// Returns false when the name is unrecognised.
func LookupPlan(name string) (PlanLimit, bool) {
	key := strings.ToLower(strings.TrimSpace(name))
	pl, ok := plans[key]
	if !ok {
		return PlanLimit{}, false
	}

	// Apply env override when present and parseable.
	if envKey, exists := envOverrideKeys[key]; exists {
		if raw := os.Getenv(envKey); raw != "" {
			if n, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil && n > 0 {
				pl.MessagesPerWindow = n
			}
		}
	}

	return pl, true
}

// Status is the quota snapshot at a given moment.
type Status struct {
	Plan            string        // the plan name (empty means "unknown"/unset)
	Configured      bool          // true when Plan is recognised; false = no display
	Limit           PlanLimit     // full plan limit spec; zero-valued when !Configured
	UsedInWindow    int           // prompts inside [now-Window, now]
	UsedPct         float64       // UsedInWindow / Limit.MessagesPerWindow * 100, capped at 100
	ResetIn         time.Duration // time until the oldest in-window prompt ages out; 0 if UsedInWindow == 0
	Window          time.Duration // echoed from Limit.Window
	At              time.Time     // snapshot timestamp
}

// Compute returns the quota Status for the given plan name at time now.
// When plan is "" or unrecognised, returns Status{Configured: false, At: now}.
// Uses data.Prompts as the countable unit (user-facing messages).
func Compute(data usage.Data, plan string, now time.Time) Status {
	limit, ok := LookupPlan(plan)
	if !ok {
		return Status{Plan: plan, At: now}
	}

	windowStart := now.Add(-limit.Window)

	var used int
	var oldest time.Time

	for _, p := range data.Prompts {
		ts := p.Timestamp
		// inclusive: ts >= windowStart && ts <= now
		if (ts.Equal(windowStart) || ts.After(windowStart)) && !ts.After(now) {
			used++
			if oldest.IsZero() || ts.Before(oldest) {
				oldest = ts
			}
		}
	}

	pct := 0.0
	if limit.MessagesPerWindow > 0 {
		pct = float64(used) / float64(limit.MessagesPerWindow) * 100
		if pct > 100 {
			pct = 100
		}
	}

	var resetIn time.Duration
	if used > 0 && !oldest.IsZero() {
		resetIn = oldest.Add(limit.Window).Sub(now)
		if resetIn < 0 {
			resetIn = 0
		}
	}

	return Status{
		Plan:         plan,
		Configured:   true,
		Limit:        limit,
		UsedInWindow: used,
		UsedPct:      pct,
		ResetIn:      resetIn,
		Window:       limit.Window,
		At:           now,
	}
}
