package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/GerardoFC8/claumeter/internal/stats"
	"github.com/GerardoFC8/claumeter/internal/usage"
)

type compactPayload struct {
	Range    string  `json:"range"`
	From     string  `json:"from,omitempty"`
	To       string  `json:"to,omitempty"`
	Prompts  int     `json:"prompts"`
	Turns    int     `json:"turns"`
	Tokens   int     `json:"tokens"`
	Cost     float64 `json:"cost_usd"`
	TopModel string  `json:"top_model,omitempty"`
	ByModel  []byModelEntry `json:"by_model,omitempty"`
}

type byModelEntry struct {
	Model  string  `json:"model"`
	Turns  int     `json:"turns"`
	Tokens int     `json:"tokens"`
	Cost   float64 `json:"cost_usd"`
}

// runCompact handles `today` and `week`.
func runCompact(cmd, _ string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	defaultRoot, _ := usage.DefaultProjectsDir()
	root := fs.String("root", defaultRoot, "directory with Claude Code JSONL transcripts")
	asJSON := fs.Bool("json", false, "JSON output")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	var preset stats.FilterPreset
	switch cmd {
	case "today":
		preset = stats.FilterToday
	case "week":
		preset = stats.FilterThisWeek
	default:
		fmt.Fprintf(os.Stderr, "runCompact: unexpected command %q\n", cmd)
		os.Exit(1)
	}
	data, err := usage.ParseAll(*root, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	filtered := preset.Apply(data)
	r := stats.Build(filtered)
	from, to := preset.Range(time.Now())
	writeCompact(preset.Label(), from, to, r, *asJSON)
}

// runRange handles `range 2026-04-01:2026-04-17`.
func runRange(args []string) {
	fs := flag.NewFlagSet("range", flag.ExitOnError)
	defaultRoot, _ := usage.DefaultProjectsDir()
	root := fs.String("root", defaultRoot, "directory with Claude Code JSONL transcripts")
	asJSON := fs.Bool("json", false, "JSON output")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: claumeter range <from>[:<to>] [--json]")
		os.Exit(2)
	}
	from, to, err := stats.ParseRange(fs.Arg(0), time.Local)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	data, err := usage.ParseAll(*root, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	filtered := stats.ApplyRange(data, from, to)
	r := stats.Build(filtered)
	label := fmt.Sprintf("%s → %s", from.Format("2006-01-02"), to.AddDate(0, 0, -1).Format("2006-01-02"))
	writeCompact(label, from, to, r, *asJSON)
}

func writeCompact(label string, from, to time.Time, r stats.Report, asJSON bool) {
	topModel := ""
	if len(r.ByModel) > 0 {
		topModel = r.ByModel[0].Model
	}
	if asJSON {
		payload := compactPayload{
			Range:    label,
			Prompts:  r.Overall.Prompts,
			Turns:    r.Overall.Turns,
			Tokens:   r.Overall.GrandTotal(),
			Cost:     round2(r.Overall.Cost),
			TopModel: topModel,
		}
		if !from.IsZero() {
			payload.From = from.Format(time.RFC3339)
		}
		if !to.IsZero() {
			payload.To = to.Format(time.RFC3339)
		}
		for _, m := range r.ByModel {
			payload.ByModel = append(payload.ByModel, byModelEntry{
				Model:  m.Model,
				Turns:  m.Totals.Turns,
				Tokens: m.Totals.GrandTotal(),
				Cost:   round2(m.Totals.Cost),
			})
		}
		out, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Println(string(out))
		return
	}
	line := fmt.Sprintf("%s: %d prompts · %d turns · %s tokens · $%.2f",
		label,
		r.Overall.Prompts,
		r.Overall.Turns,
		compactInt(r.Overall.GrandTotal()),
		r.Overall.Cost,
	)
	if topModel != "" {
		line += " (" + topModel + ")"
	}
	fmt.Println(line)
}

func round2(f float64) float64 {
	return float64(int64(f*100+0.5)) / 100
}

func compactInt(n int) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.2fB", float64(n)/1_000_000_000)
	case n >= 1_000_000:
		return fmt.Sprintf("%.2fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
