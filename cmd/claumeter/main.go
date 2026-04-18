package main

import (
	"fmt"
	"os"
	"strings"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const helpText = `claumeter — interactive TUI for Claude Code token usage

USAGE:
  claumeter [flags]                       Launch the interactive TUI (default).
  claumeter today   [--json] [--root P]   Compact summary for today.
  claumeter week    [--json] [--root P]   Compact summary for this week.
  claumeter range <from[:to]> [--json]    Compact summary for a date range.
                                          Dates are YYYY-MM-DD in local time.
  claumeter export --format=<fmt> [...]   Dump full report. Formats: json, csv, markdown.
  claumeter serve [--port N] [--token T]  HTTP daemon exposing /today /stats /range /session.
  claumeter config <verb> [key] [value]   Manage the user config file (TOML).
  claumeter version                       Print version and exit.
  claumeter help                          Show this help.

EXAMPLES:
  claumeter today
  claumeter week --json
  claumeter range 2026-04-01:2026-04-17
  claumeter export --format=json -o usage.json
  claumeter export --format=csv --range last-7d
  claumeter serve --port 7777
  claumeter config show
  claumeter config set theme light
  claumeter config get daemon_port
  claumeter --root /other/path          # TUI pointing at a different root
`

func main() {
	if len(os.Args) >= 2 {
		arg := os.Args[1]
		if !strings.HasPrefix(arg, "-") {
			switch arg {
			case "today", "week":
				runCompact(arg, "", os.Args[2:])
				return
			case "range":
				runRange(os.Args[2:])
				return
			case "export":
				runExport(os.Args[2:])
				return
			case "serve":
				runServe(os.Args[2:])
				return
			case "config":
				os.Exit(runConfig(os.Args[2:]))
			case "version":
				printVersion()
				return
			case "help":
				fmt.Print(helpText)
				return
			default:
				fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", arg, helpText)
				os.Exit(2)
			}
		}
	}
	runTUI(os.Args[1:])
}

func printVersion() {
	fmt.Printf("claumeter %s (commit %s, built %s)\n", version, commit, date)
}
