package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/GerardoFC8/claumeter/internal/export"
	"github.com/GerardoFC8/claumeter/internal/stats"
	"github.com/GerardoFC8/claumeter/internal/usage"
)

func runExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	defaultRoot, _ := usage.DefaultProjectsDir()
	root := fs.String("root", defaultRoot, "directory with Claude Code JSONL transcripts")
	format := fs.String("format", "json", "output format: json, csv, markdown (md)")
	rangeArg := fs.String("range", "all", "range: all, today, yesterday, last-7d, last-30d, last-90d, this-week, this-month, or YYYY-MM-DD[:YYYY-MM-DD]")
	outFile := fs.String("o", "", "output file (defaults to stdout)")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	label, from, to, preset, custom, err := resolveRange(*rangeArg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}

	data, err := usage.ParseAll(*root, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	var filtered usage.Data
	if custom {
		filtered = stats.ApplyRange(data, from, to)
	} else {
		filtered = preset.Apply(data)
	}
	r := stats.Build(filtered)

	var w io.Writer = os.Stdout
	if *outFile != "" {
		f, err := os.Create(*outFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}

	switch strings.ToLower(*format) {
	case "json":
		err = export.ToJSON(w, label, from, to, r)
	case "csv":
		err = export.ToCSV(w, r)
	case "markdown", "md":
		err = export.ToMarkdown(w, label, r)
	default:
		fmt.Fprintf(os.Stderr, "unknown format %q (want json, csv, markdown)\n", *format)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// resolveRange maps a --range argument to either a FilterPreset (custom=false)
// or an explicit [from, to) window (custom=true). It also returns a human label.
func resolveRange(arg string) (label string, from, to time.Time, preset stats.FilterPreset, custom bool, err error) {
	arg = strings.TrimSpace(strings.ToLower(arg))
	switch arg {
	case "", "all", "all-time":
		return "All time", time.Time{}, time.Time{}, stats.FilterAll, false, nil
	case "today":
		preset = stats.FilterToday
	case "yesterday":
		preset = stats.FilterYesterday
	case "last-7d", "last7d", "week7", "7d":
		preset = stats.FilterLast7Days
	case "last-30d", "last30d", "30d":
		preset = stats.FilterLast30Days
	case "last-90d", "last90d", "90d":
		preset = stats.FilterLast90Days
	case "this-week", "week":
		preset = stats.FilterThisWeek
	case "this-month", "month":
		preset = stats.FilterThisMonth
	default:
		f, t, perr := stats.ParseRange(arg, time.Local)
		if perr != nil {
			return "", time.Time{}, time.Time{}, 0, false, perr
		}
		label = fmt.Sprintf("%s → %s", f.Format("2006-01-02"), t.AddDate(0, 0, -1).Format("2006-01-02"))
		return label, f, t, 0, true, nil
	}
	from, to = preset.Range(time.Now())
	return preset.Label(), from, to, preset, false, nil
}
