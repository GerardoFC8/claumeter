# claumeter — Starship prompt segment

Shows your current Claude Code spend in your [Starship](https://starship.rs) prompt.

## Install

1. Make sure `claumeter` is on your `PATH`.
2. Paste the snippet from [`starship.toml`](./starship.toml) into your `~/.config/starship.toml`.
3. Pick the variant you want — inline on the right prompt, a separate line, etc.

## Output

```
$199.82 · 77p
```

Cost for today and number of real human prompts. Compact on purpose — the full breakdown lives one keystroke away in `claumeter`.

## Performance note

Starship runs every custom command on every prompt render. Parsing 300 MB of JSONL every single time would be brutal. Two options:

1. **Rely on Claude Code's update cadence** — the default interval (no caching) is actually fine because most users run commands at human speed (well below 10 Hz).
2. **Point at the daemon** instead — start `claumeter serve` at login and query `http://127.0.0.1:7777/today` with a tiny `curl | jq` wrapper. Sub-millisecond response.

```toml
[custom.claumeter]
command = "curl -s http://127.0.0.1:7777/today | jq -r '\"$\\(.cost_usd) · \\(.prompts)p\"'"
when = "curl -s --max-time 1 http://127.0.0.1:7777/healthz"
format = "[$output]($style) "
style = "bold yellow"
```
