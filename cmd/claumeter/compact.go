package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/GerardoFC8/claumeter/internal/export"
	"github.com/GerardoFC8/claumeter/internal/stats"
	"github.com/GerardoFC8/claumeter/internal/usage"
)

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
	payload := export.NewCompact(label, from, to, r)
	if asJSON {
		out, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Println(string(out))
		return
	}
	line := fmt.Sprintf("%s: %d prompts · %d turns · %s tokens · $%.2f",
		payload.Range, payload.Prompts, payload.Turns,
		compactInt(payload.Tokens), payload.Cost,
	)
	if payload.TopModel != "" {
		line += " (" + payload.TopModel + ")"
	}
	fmt.Println(line)
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
