# claumeter — Tmux status segment

Shows your Claude Code spend in the tmux status bar.

## Install

1. Make sure `claumeter` is on your `PATH`.
2. Append the snippet from [`claumeter.conf`](./claumeter.conf) to your `~/.tmux.conf`.
3. Reload tmux (`tmux source-file ~/.tmux.conf` or `Ctrl-b : source ~/.tmux.conf`).

## Output

Your tmux status right-side will look like:

```
$199.82 · 77p | 14:32
```

## Performance tip

Tmux re-renders the status bar every `status-interval` seconds (default 15). Running `claumeter today` that often re-parses the JSONL files each time. Fast but not free.

For a daemon-backed zero-cost variant, start `claumeter serve` at login and query the HTTP API:

```tmux
set -g status-right "#(curl -s http://127.0.0.1:7777/today | jq -r '\"$\\(.cost_usd) · \\(.prompts)p\"') | %H:%M "
```
