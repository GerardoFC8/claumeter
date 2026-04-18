package stats

import (
	"testing"
	"time"

	"github.com/GerardoFC8/claumeter/internal/usage"
)

func makeEvent(sessionID, model string, ts time.Time, input, output int) usage.Event {
	return usage.Event{
		SessionID:    sessionID,
		Model:        model,
		Timestamp:    ts,
		Cwd:          "/tmp/proj",
		InputTokens:  input,
		OutputTokens: output,
	}
}

func TestBuildSessionDetail_FullMatch(t *testing.T) {
	t0 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Minute)
	data := usage.Data{
		Events: []usage.Event{
			makeEvent("abc123", "claude-sonnet-4-5", t0, 100, 50),
			makeEvent("abc123", "claude-sonnet-4-5", t1, 200, 80),
			makeEvent("other-session", "claude-haiku-3", t0, 10, 5),
		},
	}

	detail, ok := BuildSessionDetail(data, "abc123")
	if !ok {
		t.Fatal("expected match, got false")
	}
	if detail.SessionID != "abc123" {
		t.Errorf("SessionID = %q, want %q", detail.SessionID, "abc123")
	}
	if len(detail.Turns) != 2 {
		t.Fatalf("len(Turns) = %d, want 2", len(detail.Turns))
	}
	// Verify chronological order.
	if !detail.Turns[0].Timestamp.Equal(t0) {
		t.Errorf("Turns[0].Timestamp = %v, want %v", detail.Turns[0].Timestamp, t0)
	}
	if !detail.Turns[1].Timestamp.Equal(t1) {
		t.Errorf("Turns[1].Timestamp = %v, want %v", detail.Turns[1].Timestamp, t1)
	}
	// Totals sum check: sum of turn Totals == session Totals (tokens).
	sumInput := detail.Turns[0].Totals.InputTokens + detail.Turns[1].Totals.InputTokens
	if sumInput != detail.Totals.InputTokens {
		t.Errorf("sumInput = %d, detail.Totals.InputTokens = %d", sumInput, detail.Totals.InputTokens)
	}
	sumOutput := detail.Turns[0].Totals.OutputTokens + detail.Turns[1].Totals.OutputTokens
	if sumOutput != detail.Totals.OutputTokens {
		t.Errorf("sumOutput = %d, detail.Totals.OutputTokens = %d", sumOutput, detail.Totals.OutputTokens)
	}
}

func TestBuildSessionDetail_PrefixMatch(t *testing.T) {
	t0 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	fullID := "abcdefghijklmnop"
	data := usage.Data{
		Events: []usage.Event{
			makeEvent(fullID, "claude-sonnet-4-5", t0, 100, 50),
		},
	}

	// Match by 8-char prefix.
	detail, ok := BuildSessionDetail(data, "abcdefgh")
	if !ok {
		t.Fatal("expected prefix match, got false")
	}
	if detail.SessionID != fullID {
		t.Errorf("SessionID = %q, want %q", detail.SessionID, fullID)
	}
}

func TestBuildSessionDetail_NoMatch(t *testing.T) {
	t0 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	data := usage.Data{
		Events: []usage.Event{
			makeEvent("abc123", "claude-sonnet-4-5", t0, 100, 50),
		},
	}

	_, ok := BuildSessionDetail(data, "zzz")
	if ok {
		t.Error("expected no match, got true")
	}
}

func TestBuildSessionDetail_ChronologicalTurns(t *testing.T) {
	// Events inserted in reverse order — output should be sorted ASC.
	t0 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Minute)
	t2 := t0.Add(2 * time.Minute)
	data := usage.Data{
		Events: []usage.Event{
			makeEvent("sess1", "claude-sonnet-4-5", t2, 300, 100),
			makeEvent("sess1", "claude-sonnet-4-5", t0, 100, 50),
			makeEvent("sess1", "claude-sonnet-4-5", t1, 200, 80),
		},
	}

	detail, ok := BuildSessionDetail(data, "sess1")
	if !ok {
		t.Fatal("expected match")
	}
	if len(detail.Turns) != 3 {
		t.Fatalf("len(Turns) = %d, want 3", len(detail.Turns))
	}
	for i, want := range []time.Time{t0, t1, t2} {
		if !detail.Turns[i].Timestamp.Equal(want) {
			t.Errorf("Turns[%d].Timestamp = %v, want %v", i, detail.Turns[i].Timestamp, want)
		}
	}
}

func TestBuildSessionDetail_TotalsSumEqualsTurnSum(t *testing.T) {
	t0 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	events := []usage.Event{
		makeEvent("s1", "claude-sonnet-4-5", t0, 100, 50),
		makeEvent("s1", "claude-sonnet-4-5", t0.Add(time.Minute), 200, 80),
		makeEvent("s1", "claude-sonnet-4-5", t0.Add(2*time.Minute), 300, 120),
	}
	data := usage.Data{Events: events}

	detail, ok := BuildSessionDetail(data, "s1")
	if !ok {
		t.Fatal("expected match")
	}

	var sumIn, sumOut int
	var sumCost float64
	for _, turn := range detail.Turns {
		sumIn += turn.Totals.InputTokens
		sumOut += turn.Totals.OutputTokens
		sumCost += turn.Totals.Cost
	}

	if sumIn != detail.Totals.InputTokens {
		t.Errorf("sum of turn input tokens (%d) != session Totals.InputTokens (%d)", sumIn, detail.Totals.InputTokens)
	}
	if sumOut != detail.Totals.OutputTokens {
		t.Errorf("sum of turn output tokens (%d) != session Totals.OutputTokens (%d)", sumOut, detail.Totals.OutputTokens)
	}
	// Cost: allow tiny float64 rounding — use a tolerance.
	const eps = 1e-9
	diff := sumCost - detail.Totals.Cost
	if diff < -eps || diff > eps {
		t.Errorf("sum of turn costs (%v) != session Totals.Cost (%v)", sumCost, detail.Totals.Cost)
	}
}

func TestBuildSessionDetail_ModelsAlphabetical(t *testing.T) {
	t0 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	data := usage.Data{
		Events: []usage.Event{
			makeEvent("sess", "claude-sonnet-4-5", t0, 100, 50),
			makeEvent("sess", "claude-haiku-3-5", t0.Add(time.Minute), 50, 20),
			makeEvent("sess", "claude-opus-4", t0.Add(2*time.Minute), 200, 100),
		},
	}

	detail, ok := BuildSessionDetail(data, "sess")
	if !ok {
		t.Fatal("expected match")
	}
	if len(detail.Models) != 3 {
		t.Fatalf("len(Models) = %d, want 3", len(detail.Models))
	}
	for i := 1; i < len(detail.Models); i++ {
		if detail.Models[i] < detail.Models[i-1] {
			t.Errorf("Models not sorted: %v[%d]=%q > %v[%d]=%q",
				detail.Models, i-1, detail.Models[i-1],
				detail.Models, i, detail.Models[i])
		}
	}
}
