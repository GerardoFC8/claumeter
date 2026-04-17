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
- **Cost estimation** — USD per day, model, session, and project using current Anthropic pricing (Opus, Sonnet, Haiku 4.x).
- **Prompts vs Turns** — correctly separates human messages from assistant API completions. Claude Code is agentic, so the ratio is usually 20–30×.
- **Tools visibility** — see which built-ins (`Read`, `Bash`, `Edit`…), skills, MCP servers, and sub-agent types drove your spend.
- **Global date filters** — cycle `All` / `Today` / `Yesterday` / `Last 7 / 30 / 90 days` / `This week` / `This month` with `f` / `F`.
- **Compact subcommands** — `claumeter today`, `week`, `range` for shell prompts and scripts. JSON output for tooling.
- **Exports** — JSON, CSV, and Markdown dumps ready to paste into docs, pipe into `jq`, or load in a spreadsheet.
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

### Subcommands

```bash
claumeter                                    # interactive TUI (default)
claumeter today [--json]                     # compact one-liner for today
claumeter week  [--json]                     # this-week summary
claumeter range <from[:to]> [--json]         # custom date range (YYYY-MM-DD)
claumeter export --format=<fmt> [--range R] [-o file]
claumeter version
claumeter help
```

`--format` accepts `json`, `csv`, or `markdown`. `--range` accepts a preset (`today`, `yesterday`, `last-7d`, `last-30d`, `last-90d`, `this-week`, `this-month`, `all`) or a raw `YYYY-MM-DD[:YYYY-MM-DD]`.

### Examples

```bash
# quick "what did I spend today" for your shell prompt
claumeter today
# → Today: 72 prompts · 1,212 turns · 169.22M tokens · $174.23 (claude-opus-4-7)

# last 7 days as CSV for a spreadsheet
claumeter export --format=csv --range last-7d -o last-week.csv

# markdown report to paste into Slack or docs
claumeter export --format=markdown --range this-month

# pipe JSON to jq for arbitrary queries
claumeter export --format=json --range last-30d | jq '.by_model[] | {model, cost_usd}'

# custom date range
claumeter range 2026-04-01:2026-04-17 --json
```

## Daemon mode

Start a local HTTP server that exposes usage over JSON + SSE:

```bash
claumeter serve --port 7777
```

Endpoints:

| Endpoint | What it returns |
|---|---|
| `GET /healthz` | Liveness + parsed event count |
| `GET /today` | Compact JSON summary of today |
| `GET /stats?range=last-7d` | Full report for a preset range |
| `GET /range?from=YYYY-MM-DD&to=YYYY-MM-DD` | Full report for a custom range |
| `GET /session/{id}` | Single session detail (first 8 chars of the UUID work) |
| `GET /live` | Server-Sent Events — initial snapshot then updates whenever the JSONL changes |

The daemon watches `~/.claude/projects/**` with fsnotify so `/today` and `/live` stay fresh without polling.

For remote exposure (e.g. a home-lab dashboard) pass `--host 0.0.0.0 --token <secret>`. Without a token claumeter refuses to bind a non-loopback address.

## Widgets

Ready-made status-bar and prompt integrations live under [`widgets/`](./widgets/). Current bundle:

- [Waybar](./widgets/waybar/) — niri / sway / Hyprland / river
- [Starship](./widgets/starship/) — shell prompt segment
- [Tmux](./widgets/tmux/) — status-bar segment

Each widget can run standalone (`claumeter today --format=waybar` polled on an interval) or point at `claumeter serve` for sub-millisecond responses and real-time push via SSE.

## How it works

Claude Code stores every session as JSONL at `~/.claude/projects/<encoded-cwd>/<session-uuid>.jsonl`. Each assistant event has a `message.usage` block with input / output / cache tokens plus a model identifier.

`claumeter` streams those files with a concurrent worker pool, extracts the usage, and aggregates by day, model, session, project, and tool.

The parser distinguishes:

- **Events** — every `type:"assistant"` message, including sub-agent (sidechain) turns that still cost tokens.
- **Prompts** — `type:"user"` messages with real text content. Excludes tool results, meta injections (attachment reminders), and sub-agent task briefings.
- **Tool uses** — `tool_use` items classified as built-in, `Skill`, `mcp__*`, or `Agent`.

## Roadmap

- [x] Interactive TUI with filters, activity matrix, and tool visibility.
- [x] Cost estimation with a versioned pricing table.
- [x] `claumeter today` / `week` / `range` compact subcommands for scripting and shell prompts.
- [x] JSON / CSV / Markdown export.
- [x] Daemon mode with HTTP API (`/stats`, `/today`, `/live`) and file-watch live tail.
- [x] Widget bundle for Waybar, starship, tmux.
- [ ] More widgets: Eww, polybar, sketchybar.
- [ ] Budget alerts with desktop notifications.
- [ ] Per-subagent drill-down and comparative date ranges.

## Related projects

- [ccusage](https://github.com/ryoppippi/ccusage) — mature Node.js CLI, great for scripting.
- [Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) — Python real-time monitor with limit predictions.

`claumeter` differentiates on interactivity, data correctness (Prompts vs Turns, tool attribution), and the upcoming widget ecosystem.

## License

[MIT](LICENSE) © Gerardo Franco
