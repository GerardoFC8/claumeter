# claumeter widgets

Bar / prompt integrations for claumeter. Pick the one that matches your stack.

## Quick chooser

| Widget | Platform | Runs on | Visual | Path |
|--------|----------|---------|--------|------|
| DMS | Linux (Wayland) | DankMaterialShell | Rich (tabs + popout) | `widgets/dms/claumeter/` |
| Waybar | Linux (Wayland) | niri · sway · Hyprland · river | Pill + tooltip | `widgets/waybar/` |
| Starship | Cross-platform | Any shell with Starship | Inline segment | `widgets/starship/` |
| Tmux | Cross-platform | tmux | Status-line segment | `widgets/tmux/` |

**Rich** means a clickable popout with tabs, per-model breakdown, and filter chips.
**Pill** means a compact `$XX.XX` token in your bar or prompt line.

---

## Two modes

Every widget supports two operating modes. The choice is independent per widget.

### Subprocess polling (default)

The widget runs `claumeter today --format=<fmt>` on an interval (typically every 15–30 seconds). No daemon required — works anywhere `claumeter` is on `PATH`.

Each poll spawns a short-lived process that re-parses `~/.claude/projects/**/*.jsonl`. On typical machines this completes in well under a second even on large transcript sets, so the overhead is not significant at 30 s intervals.

This is the right starting point. Install the widget, confirm it works, then consider the daemon if you want finer-grained refresh.

### Daemon-backed (recommended for active users)

Start the daemon at login:

```bash
# Homebrew (macOS and Linux)
brew services start claumeter

# Or manually
claumeter serve --port 7777
```

The daemon watches `~/.claude/projects/**` with fsnotify and keeps an in-memory snapshot. Widgets that detect the HTTP endpoint at `http://127.0.0.1:7777` switch from subprocess spawning to a local HTTP query — responses are sub-millisecond. You can safely drop poll intervals to 2–5 seconds without measurable overhead.

The daemon also exposes `/live` (Server-Sent Events) for true push updates. The DMS widget documents how to wire SSE; the other widgets note the daemon curl variant in their own READMEs.

**When to use it**: if you run multiple widgets simultaneously, want real-time accuracy, or notice the subprocess polls adding up on a slow filesystem.

---

## Widgets

### DMS (DankMaterialShell)

Native Quickshell QML plugin for DankMaterialShell on niri, Hyprland, or sway. Renders a `$XX.XX` pill in the DankBar; clicking opens a popout with prompts, turns, tokens, and a per-model cost breakdown. Inherits your DMS color scheme automatically via theme tokens. Refreshes every 30 seconds in subprocess mode; daemon mode drops latency to sub-millisecond and allows 2–3 s intervals.

[Install](../widgets/dms/claumeter/README.md)

### Waybar

Custom Waybar module. Shows today's cost as a pill with a tooltip. Works on any wlroots-based compositor. The `on-click` action launches the full claumeter TUI in your terminal of choice. Daemon variant swaps the `exec` command to a `curl` call against `/today`.

[Install](../widgets/waybar/README.md)

### Starship

Custom command segment for the [Starship](https://starship.rs) cross-shell prompt. Outputs `$199.82 · 77p` (cost + prompt count) inline. Because Starship rerenders on every prompt, the daemon variant (`curl` against `/today`) is especially worthwhile for heavy users who run commands frequently.

[Install](../widgets/starship/README.md)

### Tmux

Status-right segment for tmux. Re-renders on the tmux `status-interval` (default 15 s). The daemon variant replaces the subprocess call with a `curl | jq` one-liner — zero extra process spawning on each render tick.

[Install](../widgets/tmux/README.md)

---

## Planned

The following widgets are under consideration for a future release. No install instructions yet.

- **Eww** (Linux) — rich GTK-backed widget for setups that use eww bars rather than Waybar.
- **polybar** (Linux, X11) — module for X11 bar users who have not moved to Wayland.
- **sketchybar** (macOS) — item plugin for macOS menu-bar replacement users.

---

## Troubleshooting

If a widget stays blank or shows `$0.00`, see [widgets/dms/claumeter/README.md — Troubleshooting](../widgets/dms/claumeter/README.md#troubleshooting). The documented failure modes (PATH not inherited by the compositor session, Qt QML cache stale after edits) apply across all widgets: the subprocess cannot find `claumeter` if the session manager did not inherit your shell's `PATH`. The fix — prefixing Homebrew paths in the command or symlinking the binary to `/usr/local/bin` — is the same regardless of which widget you use.

---

## Contributing

To add a new bar or prompt integration: create a sub-directory under `widgets/` named after the bar or tool, add a `README.md` that covers requirements, install steps, and a config snippet that either polls `claumeter today --format=<fmt>` on an interval or queries `http://127.0.0.1:7777/today` when the daemon is running. Keep the README self-contained — users will land there from package managers and search engines without context from this catalog. Open a PR; the maintainer will add an entry to this file and to the root `README.md`.
