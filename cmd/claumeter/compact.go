package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
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
	format := fs.String("format", "plain", "output: plain, json, waybar, prompt")
	asJSON := fs.Bool("json", false, "shorthand for --format=json")
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
	writeCompact(preset.Label(), from, to, r, resolveFormat(*format, *asJSON))
}

// runRange handles `range 2026-04-01:2026-04-17`.
func runRange(args []string) {
	fs := flag.NewFlagSet("range", flag.ExitOnError)
	defaultRoot, _ := usage.DefaultProjectsDir()
	root := fs.String("root", defaultRoot, "directory with Claude Code JSONL transcripts")
	format := fs.String("format", "plain", "output: plain, json, waybar, prompt")
	asJSON := fs.Bool("json", false, "shorthand for --format=json")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: claumeter range <from>[:<to>] [--format=fmt]")
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
	writeCompact(label, from, to, r, resolveFormat(*format, *asJSON))
}

func resolveFormat(format string, asJSON bool) string {
	if asJSON {
		return "json"
	}
	return strings.ToLower(format)
}

func writeCompact(label string, from, to time.Time, r stats.Report, format string) {
	payload := export.NewCompact(label, from, to, r)
	switch format {
	case "json":
		out, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Println(string(out))
	case "waybar":
		out, _ := json.Marshal(waybarPayload(payload))
		fmt.Println(string(out))
	case "prompt":
		fmt.Println(promptLine(payload))
	case "", "plain":
		fmt.Println(plainLine(payload))
	default:
		fmt.Fprintf(os.Stderr, "unknown format %q (want plain, json, waybar, prompt)\n", format)
		os.Exit(2)
	}
}

func plainLine(p export.CompactPayload) string {
	line := fmt.Sprintf("%s: %d prompts · %d turns · %s tokens · $%.2f",
		p.Range, p.Prompts, p.Turns, compactInt(p.Tokens), p.Cost,
	)
	if p.TopModel != "" {
		line += " (" + p.TopModel + ")"
	}
	return line
}

func promptLine(p export.CompactPayload) string {
	if p.Tokens == 0 {
		return "$0"
	}
	return fmt.Sprintf("$%.2f · %dp", p.Cost, p.Prompts)
}

type waybarOut struct {
	Text    string `json:"text"`
	Tooltip string `json:"tooltip"`
	Class   string `json:"class"`
	Alt     string `json:"alt,omitempty"`
}

func waybarPayload(p export.CompactPayload) waybarOut {
	text := fmt.Sprintf("$%.2f", p.Cost)
	tooltip := fmt.Sprintf("%s\n%d prompts · %d turns · %s tokens",
		p.Range, p.Prompts, p.Turns, compactInt(p.Tokens),
	)
	if p.TopModel != "" {
		tooltip += "\nTop: " + p.TopModel
	}
	for _, m := range p.ByModel {
		tooltip += fmt.Sprintf("\n  %s — $%.2f", m.Model, m.Cost)
	}
	return waybarOut{Text: text, Tooltip: tooltip, Class: "normal"}
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
