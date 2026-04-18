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

// splitFlagArgs separates positional arguments from flag tokens so flags
// can appear anywhere on the command line (not just before positionals,
// which is the stdlib flag package default). Flags that take a value
// MUST be declared in valueFlags so we know to consume the next token.
func splitFlagArgs(args []string) (positional, flagTokens []string) {
	valueFlags := map[string]bool{"--root": true, "-root": true}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if len(a) > 0 && a[0] == '-' {
			flagTokens = append(flagTokens, a)
			if !strings.Contains(a, "=") && valueFlags[a] && i+1 < len(args) {
				flagTokens = append(flagTokens, args[i+1])
				i++
			}
			continue
		}
		positional = append(positional, a)
	}
	return
}

const compareUsage = `usage: claumeter compare <a> <b> [--root PATH] [--json]

Arguments:
  a, b   A date range. Accepts presets: today, yesterday, last-7d, last-30d,
         last-90d, this-week, this-month, all
         or a raw range:
           YYYY-MM-DD              (single day)
           YYYY-MM-DD:YYYY-MM-DD   (inclusive)

Flags:
  --root PATH   Claude Code projects root. Defaults to ~/.claude/projects.
  --json        Emit ComparisonPayload JSON instead of a human-readable table.
`

// resolveCompareRange tries a preset name first, then a raw date range. Returns the
// human label, the [from, to) window, and the filtered data.
func resolveCompareRange(arg string, data usage.Data) (label string, from, to time.Time, filtered usage.Data, err error) {
	if p, ok := stats.ResolvePreset(arg); ok {
		filtered = p.Apply(data)
		if p != stats.FilterAll {
			from, to = p.Range(time.Now())
		}
		return p.Label(), from, to, filtered, nil
	}
	from, to, err = stats.ParseRange(arg, time.Local)
	if err != nil {
		return "", time.Time{}, time.Time{}, usage.Data{},
			fmt.Errorf("unknown range %q — use a preset (today, last-7d, this-week, ...) or YYYY-MM-DD[:YYYY-MM-DD]: %w", arg, err)
	}
	filtered = stats.ApplyRange(data, from, to)
	label = fmt.Sprintf("%s → %s", from.Format("2006-01-02"), to.AddDate(0, 0, -1).Format("2006-01-02"))
	return label, from, to, filtered, nil
}

func runCompare(args []string) int {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	defaultRoot, _ := usage.DefaultProjectsDir()
	root := fs.String("root", defaultRoot, "directory with Claude Code JSONL transcripts")
	asJSON := fs.Bool("json", false, "emit ComparisonPayload JSON")

	// ContinueOnError lets us print our own usage message.
	fs.Usage = func() { fmt.Fprint(os.Stderr, compareUsage) }

	// Split positional args from flag tokens so users can write flags in any
	// position (e.g. `compare today yesterday --json` instead of being forced
	// to put --json first). The stdlib flag package stops at the first non-
	// flag arg, which is a sharp edge for a CLI with required positionals.
	positional, flagTokens := splitFlagArgs(args)
	if err := fs.Parse(flagTokens); err != nil {
		return 1
	}

	if len(positional) < 2 {
		fmt.Fprint(os.Stderr, compareUsage)
		return 1
	}
	argA := positional[0]
	argB := positional[1]

	data, err := usage.ParseAll(*root, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading usage data:", err)
		return 2
	}

	aLabel, aFrom, aTo, aFiltered, err := resolveCompareRange(argA, data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error resolving range A:", err)
		return 2
	}
	bLabel, bFrom, bTo, bFiltered, err := resolveCompareRange(argB, data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error resolving range B:", err)
		return 2
	}

	aReport := stats.Build(aFiltered)
	bReport := stats.Build(bFiltered)
	cmp := stats.Compare(aReport.Overall, bReport.Overall)

	if *asJSON {
		payload := export.NewComparison(
			aLabel, aFrom, aTo, aReport.Overall,
			bLabel, bFrom, bTo, bReport.Overall,
			cmp,
		)
		out, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error marshalling JSON:", err)
			return 2
		}
		fmt.Println(string(out))
		return 0
	}

	printCompareTable(aLabel, aFrom, aTo, bLabel, bFrom, bTo, cmp)
	return 0
}

// printCompareTable prints an aligned two-column comparison table.
func printCompareTable(aLabel string, aFrom, aTo time.Time, bLabel string, bFrom, bTo time.Time, cmp stats.Comparison) {
	fmtDate := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02")
	}
	aRange := aLabel
	if !aFrom.IsZero() {
		aRange = fmt.Sprintf("%s (%s → %s)", aLabel, fmtDate(aFrom), fmtDate(aTo.AddDate(0, 0, -1)))
	}
	bRange := bLabel
	if !bFrom.IsZero() {
		bRange = fmt.Sprintf("%s (%s → %s)", bLabel, fmtDate(bFrom), fmtDate(bTo.AddDate(0, 0, -1)))
	}

	fmt.Printf("claumeter compare — A: %s\n", aRange)
	fmt.Printf("                    B: %s\n\n", bRange)

	const hdrFmt = "%-24s %14s %14s %14s %8s\n"
	const rowFmt = "%-24s %14s %14s %14s %7.1f%%\n"
	fmt.Printf(hdrFmt, "", "A", "B", "Delta", "%")
	fmt.Printf(hdrFmt, "------------------------", "--------------", "--------------", "--------------", "-------")

	type row struct {
		label string
		d     stats.Delta
		money bool
	}
	rows := []row{
		{"Prompts", cmp.Prompts, false},
		{"Turns", cmp.Turns, false},
		{"Input tokens", cmp.InputTokens, false},
		{"Cache creation tokens", cmp.CacheCreationTokens, false},
		{"Cache read tokens", cmp.CacheReadTokens, false},
		{"Output tokens", cmp.OutputTokens, false},
		{"Tokens (total)", cmp.TotalTokens, false},
		{"Cost (USD)", cmp.CostUSD, true},
	}

	for _, r := range rows {
		var aStr, bStr, dStr string
		if r.money {
			aStr = fmt.Sprintf("$%.2f", r.d.A)
			bStr = fmt.Sprintf("$%.2f", r.d.B)
			dStr = fmt.Sprintf("%+.2f", r.d.Delta) // explicit sign
			if r.d.Delta < 0 {
				dStr = fmt.Sprintf("-$%.2f", -r.d.Delta)
			} else {
				dStr = fmt.Sprintf("+$%.2f", r.d.Delta)
			}
		} else {
			aStr = compactInt(int(r.d.A))
			bStr = compactInt(int(r.d.B))
			delta := int(r.d.Delta)
			if delta >= 0 {
				dStr = "+" + compactInt(delta)
			} else {
				dStr = "-" + compactInt(-delta)
			}
		}
		fmt.Printf(rowFmt, r.label, aStr, bStr, dStr, r.d.Pct)
	}
}
