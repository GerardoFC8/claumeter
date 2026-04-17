# claumeter

> Interactive TUI for analyzing Claude Code token usage.

[![Release](https://img.shields.io/github/v/release/GerardoFC8/claumeter)](https://github.com/GerardoFC8/claumeter/releases/latest)
[![CI](https://github.com/GerardoFC8/claumeter/actions/workflows/ci.yml/badge.svg)](https://github.com/GerardoFC8/claumeter/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/GerardoFC8/claumeter)](https://goreportcard.com/report/github.com/GerardoFC8/claumeter)

`claumeter` is a Go TUI that parses your local Claude Code JSONL transcripts and surfaces what matters: tokens by day, model, session, and project, plus a full breakdown of the tools, skills, MCP servers, and sub-agents you invoked.

```
Day         Prompts  Turns   opus-4-6  opus-4-7  sonnet-4-6  haiku   Total
2026-04-17      45    580         —    320.5M       8.1M      2.1M   330.7M
2026-04-16      89   1,692    31.28M   345.94M     11.38M     9.41M  398.01M
2026-04-15     106   1,660   162.45M        —      5.80M     17.89M  186.15M
──────────  ──────  ─────   ────────  ────────  ──────────  ───────  ──────
▸ TOTAL        240  3,932   193.73M   666.44M     25.29M    29.41M   914.86M
```

## Features

- **Activity matrix** — day × model token breakdown with a TOTAL footer row.
- **Prompts vs Turns** — correctly separates human messages from assistant API completions. Claude Code is agentic, so the ratio is usually 20–30×.
- **Tools visibility** — see which built-ins (`Read`, `Bash`, `Edit`…), skills, MCP servers, and sub-agent types drove your spend.
- **Global date filters** — cycle `All` / `Today` / `Yesterday` / `Last 7 / 30 / 90 days` / `This week` / `This month` with `f` / `F`.
- **Vim navigation** — `j/k/g/G/ctrl+d/u/b/f` inside tables, `h/l` between tabs.
- **Streaming JSONL parser** with a `NumCPU` worker pool — parses 300 MB / 600 files in under a second.

## Install

### Homebrew (macOS and Linux)

```bash
brew tap GerardoFC8/tap
brew install claumeter
```

### `go install`

```bash
go install github.com/GerardoFC8/claumeter/cmd/claumeter@latest
```

### Prebuilt binaries

Grab the tarball for your platform from the [latest release](https://github.com/GerardoFC8/claumeter/releases/latest): `linux_x86_64`, `linux_arm64`, `darwin_x86_64`, `darwin_arm64`.

## Usage

Just run `claumeter` — it reads `~/.claude/projects/**/*.jsonl` and opens the TUI.

### Keybindings

| Key | Action |
|---|---|
| `1`–`5` / `tab` / `h` / `l` | Switch tab (Overview / Activity / Sessions / Projects / Tools) |
| `f` / `F` | Cycle date filter forward / backward |
| `j` / `k` | Row down / up |
| `g` / `G` | Jump to top / bottom |
| `ctrl+d` / `ctrl+u` | Half page down / up |
| `ctrl+f` / `ctrl+b` | Full page down / up |
| `q` / `esc` | Quit |

### Flags

```
claumeter --root <path>    # default: ~/.claude/projects/
claumeter --version
```

## How it works

Claude Code stores every session as JSONL at `~/.claude/projects/<encoded-cwd>/<session-uuid>.jsonl`. Each assistant event has a `message.usage` block with input / output / cache tokens plus a model identifier.

`claumeter` streams those files with a concurrent worker pool, extracts the usage, and aggregates by day, model, session, project, and tool.

The parser distinguishes:

- **Events** — every `type:"assistant"` message, including sub-agent (sidechain) turns that still cost tokens.
- **Prompts** — `type:"user"` messages with real text content. Excludes tool results, meta injections (attachment reminders), and sub-agent task briefings.
- **Tool uses** — `tool_use` items classified as built-in, `Skill`, `mcp__*`, or `Agent`.

## Roadmap

- [x] Interactive TUI with filters, activity matrix, and tool visibility.
- [ ] Cost estimation with a versioned pricing table.
- [ ] `claumeter today` / `week` compact subcommands for scripting and shell prompts.
- [ ] JSON / CSV / Markdown export.
- [ ] Daemon mode with HTTP API (`/stats`, `/today`, `/live`) and file-watch live tail.
- [ ] Widget bundle for Waybar, Eww, polybar, tmux, sketchybar, starship.
- [ ] Per-subagent drill-down and comparative date ranges.

## Related projects

- [ccusage](https://github.com/ryoppippi/ccusage) — mature Node.js CLI, great for scripting.
- [Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) — Python real-time monitor with limit predictions.

`claumeter` differentiates on interactivity, data correctness (Prompts vs Turns, tool attribution), and the upcoming widget ecosystem.

## License

[MIT](LICENSE) © Gerardo Franco
