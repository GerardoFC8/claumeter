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

- **Six-tab TUI** — Overview, Activity, Sessions, Projects, Tools, Compare.
- **Activity matrix** — day × model token breakdown with a TOTAL footer row.
- **Cost estimation** — USD per day, model, session, and project using current Anthropic pricing (Opus, Sonnet, Haiku 4.x).
- **Prompts vs Turns** — correctly separates human messages from assistant API completions. Claude Code is agentic, so the ratio is usually 20–30×.
- **Session drill-down** — `enter` on the Sessions tab opens a turn-by-turn detail view with per-call tokens, cost, and tools.
- **Comparative mode** — the Compare tab diffs two date ranges side-by-side with delta and %. Same logic also available via `claumeter compare` and the `/compare` daemon endpoint.
- **Quota awareness** — plan-aware rate-limit estimation from a local 5-hour rolling window. Header badge shows `X/Y msgs · 5h window (Z%) · resets Nh Mm`. Cycle plans with `Q`.
- **Cost projection** — projected monthly cost based on the last 7 days of actual spend.
- **Cache Hit Rate** — share of input tokens served from Anthropic's prompt cache.
- **Hourly activity heatmap** — sparkline showing which hours of the day you use Claude Code the most.
- **Top expensive sessions** — top 3 costliest sessions surfaced in Overview at a glance.
- **Tools visibility** — see which built-ins (`Read`, `Bash`, `Edit`…), skills, MCP servers, and sub-agent types drove your spend.
- **Global date filters** — cycle `All` / `Today` / `Yesterday` / `Last 7 / 30 / 90 days` / `This week` / `This month` with `f` / `F`.
- **Keyboard search** — `/` filters Activity / Sessions / Projects / Tools live.
- **Three themes** — dark, light, high-contrast. Cycle with `t`, persisted to config.
- **Responsive layout** — 3 breakpoints adapt the UI to split terminals and narrow laptop screens.
- **Help overlay** — press `?` anytime for full keybindings + glossary (Prompt, Turn, agentic ratio, cache tiers, MCP).
- **First-run onboarding** — welcome screen summarizes the 6 tabs; shown once, then silent.
- **Compact subcommands** — `claumeter today`, `week`, `range`, `compare`, `quota` for shell prompts and scripts. JSON output for tooling.
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

Just run `claumeter` — it reads `~/.claude/projects/**/*.jsonl` and opens the TUI. On first run a welcome screen describes the tabs.

### Keybindings

| Key | Action |
|---|---|
| `1`–`6` / `tab` / `shift+tab` / `h` / `l` | Switch tab (Overview / Activity / Sessions / Projects / Tools / Compare) |
| `f` / `F` | Cycle date filter forward / backward |
| `/` | Search (applies to Activity / Sessions / Projects / Tools) |
| `enter` | Drill into session detail (Sessions tab) |
| `a` / `A` | Cycle range A forward / backward (Compare tab) |
| `b` / `B` | Cycle range B forward / backward (Compare tab) |
| `t` | Cycle theme (dark → light → high-contrast) |
| `Q` | Cycle Claude plan (pro → max-5x → max-20x → unset) for quota estimates |
| `?` | Toggle help overlay (keybindings + glossary) |
| `j` / `k` | Row down / up |
| `g` / `G` | Jump to top / bottom |
| `ctrl+d` / `ctrl+u` | Half page down / up |
| `ctrl+f` / `ctrl+b` | Full page down / up |
| `esc` | Close overlay / exit detail / clear search |
| `q` / `ctrl+c` | Quit |

### Subcommands

```bash
claumeter                                    # interactive TUI (default)
claumeter today [--json]                     # compact one-liner for today
claumeter week  [--json]                     # this-week summary
claumeter range <from[:to]> [--json]         # custom date range (YYYY-MM-DD)
claumeter compare <a> <b> [--json]           # side-by-side comparison of two ranges
claumeter quota  [--plan PLAN] [--json]      # estimated 5h rate-limit status
claumeter export --format=<fmt> [--range R] [-o file]
claumeter serve  [--port N] [--host H] [--token T]
claumeter config [get|set|path]
claumeter version
claumeter help
```

`--format` accepts `json`, `csv`, or `markdown`. `--range` accepts a preset (`today`, `yesterday`, `last-7d`, `last-30d`, `last-90d`, `this-week`, `this-month`, `all`) or a raw `YYYY-MM-DD[:YYYY-MM-DD]`. `--plan` for quota accepts `pro`, `max-5x`, `max-20x`.

### Examples

```bash
# quick "what did I spend today" for your shell prompt
claumeter today
# → Today: 72 prompts · 1,212 turns · 169.22M tokens · $174.23 (claude-opus-4-7)

# did I burn more this week than last?
claumeter compare last-7d this-week

# am I about to hit my 5h quota?
claumeter quota --plan max-5x

# last 7 days as CSV for a spreadsheet
claumeter export --format=csv --range last-7d -o last-week.csv

# markdown report to paste into Slack or docs
claumeter export --format=markdown --range this-month

# pipe JSON to jq for arbitrary queries
claumeter export --format=json --range last-30d | jq '.by_model[] | {model, cost_usd}'

# custom date range
claumeter range 2026-04-01:2026-04-17 --json
```

## Configuration

Config lives at `~/.config/claumeter/config.toml` (or `$XDG_CONFIG_HOME/claumeter/config.toml`). Managed via `claumeter config`:

```bash
claumeter config path           # print the resolved config path
claumeter config show           # dump current config as TOML
claumeter config get theme      # read a single key
claumeter config set theme dark # dark | light | high-contrast
claumeter config set plan pro   # "" | pro | max-5x | max-20x
```

Theme, plan, onboarding state, and per-tab hint counters persist automatically when you change them from the TUI.

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
| `GET /compare?a=RANGE_A&b=RANGE_B` | Side-by-side delta between two ranges |
| `GET /quota?plan=PLAN` | Estimated 5h rate-limit status for the given plan |
| `GET /live` | Server-Sent Events — initial snapshot then updates whenever the JSONL changes |

The daemon watches `~/.claude/projects/**` with fsnotify so `/today` and `/live` stay fresh without polling.

For remote exposure (e.g. a home-lab dashboard) pass `--host 0.0.0.0 --token <secret>`. Without a token claumeter refuses to bind a non-loopback address.

## Widgets

Ready-made status-bar and prompt integrations live under [`widgets/`](./widgets/). Current bundle:

- [DMS](./widgets/dms/claumeter/) — native DankMaterialShell / Quickshell plugin (Linux/Wayland, rich popout)
- [Waybar](./widgets/waybar/) — niri / sway / Hyprland / river
- [Starship](./widgets/starship/) — shell prompt segment
- [Tmux](./widgets/tmux/) — status-bar segment

Each widget can run standalone (`claumeter today` polled on an interval) or point at `claumeter serve` for sub-millisecond responses and real-time push via SSE. See [`docs/WIDGETS.md`](./docs/WIDGETS.md) for the full catalog and chooser table.

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
- [x] Widget bundle for Waybar, starship, tmux, and a rich DMS popout.
- [x] Session drill-down (turn-by-turn detail).
- [x] Comparative date ranges (CLI + daemon + TUI tab).
- [x] Plan-aware rate-limit awareness (local 5h window).
- [x] Help overlay, first-run onboarding, responsive layout.
- [ ] Per-subagent cost attribution (blocked upstream on Claude Code's JSONL schema).
- [ ] Budget alerts with desktop notifications.
- [ ] More widgets: Eww, polybar, sketchybar.
- [ ] Team aggregation server (opt-in, self-hosted) and Prometheus exporter.

## Related projects

- [ccusage](https://github.com/ryoppippi/ccusage) — mature Node.js CLI, great for scripting.
- [Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) — Python real-time monitor with limit predictions.

`claumeter` differentiates on interactivity, data correctness (Prompts vs Turns, tool attribution), and the widget ecosystem.

## License

[MIT](LICENSE) © Gerardo Franco
