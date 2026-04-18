package stats

import (
	"sort"
	"strings"
	"time"

	"github.com/GerardoFC8/claumeter/internal/usage"
)

// SessionTurn captures one assistant turn inside a session.
type SessionTurn struct {
	Timestamp time.Time
	Model     string
	// Totals is the usage of this single Event (not cumulative).
	Totals Totals
	// Tools contains the tool_use items the assistant emitted in this turn.
	// Empty if the turn had no tool calls.
	Tools []string
}

// SessionDetail is the per-session drill-down report.
type SessionDetail struct {
	SessionID string
	Cwd       string
	FirstSeen time.Time
	LastSeen  time.Time
	Models    []string       // sorted, unique
	Totals    Totals         // session totals
	Turns     []SessionTurn  // chronological, oldest first
}

// toolLabel produces the short display label for a ToolUse.
func toolLabel(u usage.ToolUse) string {
	switch u.Kind {
	case usage.ToolBuiltin:
		return u.Name
	case usage.ToolSkill:
		target := u.Target
		if target == "" {
			target = "(unknown)"
		}
		return "Skill/" + target
	case usage.ToolMCP:
		// Pass through the full raw tool name (e.g. "mcp__server__tool").
		if u.Name != "" {
			return u.Name
		}
		return "mcp__" + u.MCPServer + "__" + u.Target
	case usage.ToolAgent:
		target := u.Target
		if target == "" {
			target = "(unknown)"
		}
		return "Agent/" + target
	}
	return u.Name
}

// BuildSessionDetail returns the detail for a specific session.
// If sessionID has length >= 8 it also matches by prefix (so short ids work).
// Returns (SessionDetail{}, false) when no session matches.
func BuildSessionDetail(data usage.Data, sessionID string) (SessionDetail, bool) {
	matchID := func(id string) bool {
		if id == sessionID {
			return true
		}
		if len(sessionID) >= 8 && strings.HasPrefix(id, sessionID) {
			return true
		}
		return false
	}

	// Collect matching events.
	var events []usage.Event
	for _, e := range data.Events {
		if matchID(e.SessionID) {
			events = append(events, e)
		}
	}
	if len(events) == 0 {
		return SessionDetail{}, false
	}

	// Sort events chronologically.
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	// Collect matching tool uses, grouped by truncated-to-second timestamp for
	// pairing with events.
	type toolKey struct {
		sessionID string
		ts        time.Time // truncated to second
	}
	toolsByKey := map[time.Time][]string{}
	for _, u := range data.ToolUses {
		if !matchID(u.SessionID) {
			continue
		}
		tsSec := u.Timestamp.Truncate(time.Second)
		toolsByKey[tsSec] = append(toolsByKey[tsSec], toolLabel(u))
	}

	// Determine the canonical session ID (may have been matched by prefix).
	canonicalID := events[0].SessionID
	cwd := events[0].Cwd

	// Build turns.
	turns := make([]SessionTurn, 0, len(events))
	modelsSet := map[string]struct{}{}
	var sessTotal Totals

	for _, e := range events {
		if e.SessionID != canonicalID {
			// Only one session per call (prefix may be ambiguous, but we committed
			// to the first match).
			continue
		}
		modelsSet[e.Model] = struct{}{}

		t := Totals{}
		t.InputTokens = e.InputTokens
		t.OutputTokens = e.OutputTokens
		t.CacheCreationTokens = e.CacheCreationTokens
		t.CacheReadTokens = e.CacheReadTokens
		t.Turns = 1
		t.Cost = costForEvent(e)

		sessTotal.InputTokens += t.InputTokens
		sessTotal.OutputTokens += t.OutputTokens
		sessTotal.CacheCreationTokens += t.CacheCreationTokens
		sessTotal.CacheReadTokens += t.CacheReadTokens
		sessTotal.Turns++
		sessTotal.Cost += t.Cost

		tsSec := e.Timestamp.Truncate(time.Second)
		tools := toolsByKey[tsSec]
		// De-duplicate tools within a turn while preserving order.
		var deduped []string
		seen := map[string]struct{}{}
		for _, lbl := range tools {
			if _, ok := seen[lbl]; !ok {
				seen[lbl] = struct{}{}
				deduped = append(deduped, lbl)
			}
		}

		turns = append(turns, SessionTurn{
			Timestamp: e.Timestamp,
			Model:     e.Model,
			Totals:    t,
			Tools:     deduped,
		})
	}

	// Also count prompts for session totals.
	for _, p := range data.Prompts {
		if matchID(p.SessionID) && p.SessionID == canonicalID {
			sessTotal.Prompts++
		}
	}

	// Build sorted models slice.
	models := make([]string, 0, len(modelsSet))
	for m := range modelsSet {
		models = append(models, m)
	}
	sort.Strings(models)

	// Determine first/last seen.
	firstSeen := events[0].Timestamp
	lastSeen := events[len(events)-1].Timestamp

	// Check prompts for earlier firstSeen / later lastSeen.
	for _, p := range data.Prompts {
		if !matchID(p.SessionID) || p.SessionID != canonicalID {
			continue
		}
		if p.Timestamp.Before(firstSeen) {
			firstSeen = p.Timestamp
		}
		if p.Timestamp.After(lastSeen) {
			lastSeen = p.Timestamp
		}
	}

	return SessionDetail{
		SessionID: canonicalID,
		Cwd:       cwd,
		FirstSeen: firstSeen,
		LastSeen:  lastSeen,
		Models:    models,
		Totals:    sessTotal,
		Turns:     turns,
	}, true
}
