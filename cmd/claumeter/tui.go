package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/GerardoFC8/claumeter/internal/config"
	"github.com/GerardoFC8/claumeter/internal/tui"
	"github.com/GerardoFC8/claumeter/internal/usage"
)

func runTUI(args []string) {
	fs := flag.NewFlagSet("claumeter", flag.ExitOnError)
	defaultRoot, err := usage.DefaultProjectsDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	root := fs.String("root", defaultRoot, "directory with Claude Code JSONL transcripts")
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if *showVersion {
		printVersion()
		return
	}
	if _, err := os.Stat(*root); err != nil {
		fmt.Fprintf(os.Stderr, "cannot access %s: %v\n", *root, err)
		os.Exit(1)
	}

	// Load persisted theme; fall back to defaults on error.
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "claumeter: config load error (using defaults):", err)
		cfg = config.Defaults()
	}

	p := tea.NewProgram(tui.NewWithTheme(*root, cfg.Theme), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
