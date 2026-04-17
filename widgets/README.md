# claumeter widgets

Ready-to-paste integrations for your status bar and prompt.

| Widget | Path | Compositor / Shell |
|---|---|---|
| [DankMaterialShell](./dms/claumeter/) | `dms/claumeter/` | niri / Hyprland / sway users on DMS (native plugin) |
| [Waybar](./waybar/) | `waybar/config.jsonc` + `style.css` | niri · sway · Hyprland · river |
| [Starship](./starship/) | `starship/starship.toml` | any shell that runs Starship |
| [Tmux](./tmux/) | `tmux/claumeter.conf` | any Tmux setup |

More widgets (Eww, polybar, sketchybar) coming in a follow-up release.

## Two modes

Each widget can run in one of two modes:

1. **Direct (default)** — the widget calls `claumeter today --format=...` on an interval. Simple, no setup, reparses JSONL every poll (fast — under a second even on 300 MB).
2. **Daemon-backed** — start `claumeter serve` at login, point the widget at `http://127.0.0.1:7777/today`. Sub-millisecond, and the daemon watches files so updates are real-time via `/live` (SSE).

If you're heavy on Claude Code or want multi-widget consistency, go daemon mode.
