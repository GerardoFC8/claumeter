package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/GerardoFC8/claumeter/internal/config"
	"github.com/GerardoFC8/claumeter/internal/quota"
	"github.com/GerardoFC8/claumeter/internal/usage"
	"time"
)

const quotaUsage = `usage: claumeter quota [--plan pro|max-5x|max-20x] [--json] [--root PATH]

Show estimated Claude Code rate-limit status based on the local 5-hour
message window. Limits are unofficial estimates — Anthropic does not
publish quota APIs for Claude Code.

Flags:
  --plan PLAN   Plan to evaluate (pro, max-5x, max-20x). Defaults to config.
  --json        Emit JSON instead of a human-readable table.
  --root PATH   Claude Code projects root. Defaults to ~/.claude/projects.

Env overrides (no recompile required when Anthropic changes limits):
  CLAUMETER_QUOTA_PRO_MESSAGES=50 claumeter quota --plan pro
  CLAUMETER_QUOTA_MAX5X_MESSAGES=250
  CLAUMETER_QUOTA_MAX20X_MESSAGES=1000
`

func runQuota(args []string) int {
	fs := flag.NewFlagSet("quota", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprint(os.Stderr, quotaUsage) }

	defaultRoot, _ := usage.DefaultProjectsDir()
	root := fs.String("root", defaultRoot, "directory with Claude Code JSONL transcripts")
	planFlag := fs.String("plan", "", "plan: pro | max-5x | max-20x")
	asJSON := fs.Bool("json", false, "emit JSON")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	plan := *planFlag
	if plan == "" {
		cfg, err := config.Load()
		if err == nil {
			plan = cfg.Plan
		}
	}

	if plan == "" {
		fmt.Print(quotaNoPlanMessage())
		return 0
	}

	data, err := usage.ParseAll(*root, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading usage data:", err)
		return 2
	}

	status := quota.Compute(data, plan, time.Now())

	if !status.Configured {
		fmt.Fprintf(os.Stderr, "unknown plan %q\n\n", plan)
		fmt.Print(quotaNoPlanMessage())
		return 1
	}

	if *asJSON {
		out, err := json.MarshalIndent(map[string]any{
			"plan":             status.Plan,
			"configured":       status.Configured,
			"limit_messages":   status.Limit.MessagesPerWindow,
			"window_seconds":   int(status.Window.Seconds()),
			"used_in_window":   status.UsedInWindow,
			"used_pct":         status.UsedPct,
			"reset_in_seconds": int(status.ResetIn.Seconds()),
			"at":               status.At,
			"description":      status.Limit.Description,
		}, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error marshalling JSON:", err)
			return 2
		}
		fmt.Println(string(out))
		return 0
	}

	printQuotaTable(status)
	return 0
}

func printQuotaTable(s quota.Status) {
	fmt.Printf("Plan:      %s\n", s.Limit.Description)
	fmt.Printf("Window:    %s\n", formatQuotaDuration(s.Window))
	fmt.Printf("Used:      %d / %d  (%.1f%%)\n",
		s.UsedInWindow, s.Limit.MessagesPerWindow, s.UsedPct)
	if s.ResetIn > 0 {
		fmt.Printf("Resets in: %s\n", formatQuotaDuration(s.ResetIn))
	} else {
		fmt.Printf("Resets in: n/a (no messages in window)\n")
	}
}

// formatQuotaDuration renders a duration as "Xh Ym" (e.g. "3h 12m" or "5h0m").
func formatQuotaDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := int(math.Floor(d.Hours()))
	m := int(d.Minutes()) - h*60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}

func quotaNoPlanMessage() string {
	msg := "No plan configured. Set one with:\n"
	msg += "  claumeter config set plan max-5x\n\n"
	msg += "Or pass --plan pro|max-5x|max-20x explicitly.\n"
	msg += "Available plans:\n"
	for _, name := range quota.Plans() {
		pl, _ := quota.LookupPlan(name)
		msg += fmt.Sprintf("  - %-10s %s\n", name, pl.Description)
	}
	msg += "\nNote: these are unofficial estimates. Override via env:\n"
	msg += "  CLAUMETER_QUOTA_PRO_MESSAGES=50 claumeter quota --plan pro\n"
	return msg
}
