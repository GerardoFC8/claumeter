package quota

import (
	"testing"
	"time"

	"github.com/GerardoFC8/claumeter/internal/usage"
)

func makePrompts(base time.Time, offsets ...time.Duration) []usage.Prompt {
	ps := make([]usage.Prompt, 0, len(offsets))
	for _, off := range offsets {
		ps = append(ps, usage.Prompt{Timestamp: base.Add(off)})
	}
	return ps
}

func TestUnknownPlan(t *testing.T) {
	now := time.Now()
	s := Compute(usage.Data{}, "enterprise", now)
	if s.Configured {
		t.Fatal("expected Configured=false for unknown plan")
	}
	if s.Plan != "enterprise" {
		t.Fatalf("expected Plan=enterprise, got %q", s.Plan)
	}
}

func TestEmptyPlan(t *testing.T) {
	now := time.Now()
	s := Compute(usage.Data{}, "", now)
	if s.Configured {
		t.Fatal("expected Configured=false for empty plan")
	}
}

func TestKnownPlanNoPrompts(t *testing.T) {
	now := time.Now()
	s := Compute(usage.Data{}, "pro", now)
	if !s.Configured {
		t.Fatal("expected Configured=true for pro")
	}
	if s.UsedInWindow != 0 {
		t.Fatalf("expected UsedInWindow=0, got %d", s.UsedInWindow)
	}
	if s.ResetIn != 0 {
		t.Fatalf("expected ResetIn=0, got %v", s.ResetIn)
	}
	if s.UsedPct != 0 {
		t.Fatalf("expected UsedPct=0, got %f", s.UsedPct)
	}
}

func TestOnlyInWindowCounted(t *testing.T) {
	now := time.Now()
	// 3 prompts inside the 5h window, 2 outside.
	prompts := makePrompts(now,
		-1*time.Hour,   // in
		-2*time.Hour,   // in
		-4*time.Hour,   // in (just inside)
		-6*time.Hour,   // out (beyond 5h)
		-10*time.Hour,  // out
	)
	data := usage.Data{Prompts: prompts}
	s := Compute(data, "pro", now)
	if s.UsedInWindow != 3 {
		t.Fatalf("expected 3 in-window prompts, got %d", s.UsedInWindow)
	}
}

func TestUsedPctCapsAt100(t *testing.T) {
	now := time.Now()
	// pro limit is 45; send 60 prompts in window.
	prompts := make([]usage.Prompt, 60)
	for i := range prompts {
		prompts[i] = usage.Prompt{Timestamp: now.Add(-time.Duration(i+1) * time.Minute)}
	}
	s := Compute(usage.Data{Prompts: prompts}, "pro", now)
	if s.UsedPct != 100 {
		t.Fatalf("expected UsedPct=100 (capped), got %f", s.UsedPct)
	}
}

func TestUsedPctValue(t *testing.T) {
	now := time.Now()
	// max-5x limit = 225; use 45 → 20%
	prompts := make([]usage.Prompt, 45)
	for i := range prompts {
		prompts[i] = usage.Prompt{Timestamp: now.Add(-time.Duration(i+1) * time.Minute)}
	}
	s := Compute(usage.Data{Prompts: prompts}, "max-5x", now)
	want := 45.0 / 225.0 * 100
	if s.UsedPct != want {
		t.Fatalf("expected UsedPct=%f, got %f", want, s.UsedPct)
	}
}

func TestResetInComputedFromOldest(t *testing.T) {
	now := time.Now()
	// oldest prompt is 3h ago; with 5h window, resets in 2h.
	prompts := makePrompts(now,
		-3*time.Hour, // oldest
		-1*time.Hour,
		-30*time.Minute,
	)
	s := Compute(usage.Data{Prompts: prompts}, "pro", now)
	if s.UsedInWindow != 3 {
		t.Fatalf("expected 3 prompts, got %d", s.UsedInWindow)
	}
	// Should be ~2h (within a small delta for test timing).
	expectedResetIn := 2 * time.Hour
	delta := s.ResetIn - expectedResetIn
	if delta < 0 {
		delta = -delta
	}
	if delta > 2*time.Second {
		t.Fatalf("expected ResetIn ~2h, got %v (delta %v)", s.ResetIn, delta)
	}
}

func TestWindowBoundaryInclusive(t *testing.T) {
	now := time.Now()
	// A prompt exactly at the window boundary (now - 5h) must be counted.
	windowStart := now.Add(-5 * time.Hour)
	prompts := []usage.Prompt{
		{Timestamp: windowStart},                  // exactly at boundary — must be counted
		{Timestamp: windowStart.Add(-time.Second)}, // just outside — must NOT be counted
	}
	s := Compute(usage.Data{Prompts: prompts}, "pro", now)
	if s.UsedInWindow != 1 {
		t.Fatalf("expected 1 prompt (boundary inclusive), got %d", s.UsedInWindow)
	}
}

func TestEnvOverride(t *testing.T) {
	t.Setenv("CLAUMETER_QUOTA_PRO_MESSAGES", "100")
	pl, ok := LookupPlan("pro")
	if !ok {
		t.Fatal("expected pro to be found")
	}
	if pl.MessagesPerWindow != 100 {
		t.Fatalf("expected MessagesPerWindow=100 via env override, got %d", pl.MessagesPerWindow)
	}
}

func TestEnvOverrideMax5x(t *testing.T) {
	t.Setenv("CLAUMETER_QUOTA_MAX5X_MESSAGES", "300")
	pl, ok := LookupPlan("max-5x")
	if !ok {
		t.Fatal("expected max-5x to be found")
	}
	if pl.MessagesPerWindow != 300 {
		t.Fatalf("expected MessagesPerWindow=300 via env override, got %d", pl.MessagesPerWindow)
	}
}

func TestLookupPlanCaseInsensitive(t *testing.T) {
	cases := []string{"Pro", "PRO", "pro", "MAX-5X", "Max-5x", "max-5x"}
	for _, c := range cases {
		pl, ok := LookupPlan(c)
		if !ok {
			t.Errorf("LookupPlan(%q): expected ok=true", c)
			continue
		}
		if pl.Plan == "" {
			t.Errorf("LookupPlan(%q): got empty Plan field", c)
		}
	}
}

func TestPlansReturnsSorted(t *testing.T) {
	ps := Plans()
	if len(ps) == 0 {
		t.Fatal("Plans() returned empty slice")
	}
	for i := 1; i < len(ps); i++ {
		if ps[i] < ps[i-1] {
			t.Fatalf("Plans() not sorted: %v[%d]=%q < %v[%d]=%q", ps, i, ps[i], ps, i-1, ps[i-1])
		}
	}
}
