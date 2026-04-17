package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/GerardoFC8/claumeter/internal/tui"
	"github.com/GerardoFC8/claumeter/internal/usage"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	defaultRoot, err := usage.DefaultProjectsDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	root := flag.String("root", defaultRoot, "directory with Claude Code JSONL transcripts")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("claumeter %s (commit %s, built %s)\n", version, commit, date)
		return
	}

	if _, err := os.Stat(*root); err != nil {
		fmt.Fprintf(os.Stderr, "cannot access %s: %v\n", *root, err)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.New(*root), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
