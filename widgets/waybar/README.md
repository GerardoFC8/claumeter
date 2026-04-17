# claumeter — Waybar module

Drop-in custom module for [Waybar](https://github.com/Alexays/Waybar). Works on any wlroots-based Wayland compositor: **niri**, **sway**, **Hyprland**, **river**, etc.

## Install

1. Make sure `claumeter` is on your `PATH` (`brew install claumeter` does that automatically).
2. Copy the snippet from [`config.jsonc`](./config.jsonc) into your Waybar config (usually `~/.config/waybar/config.jsonc`).
3. Add `"custom/claumeter"` to your `modules-right` (or wherever you want it).
4. Optionally copy the style rules from [`style.css`](./style.css) into `~/.config/waybar/style.css`.
5. Reload Waybar (`killall -SIGUSR2 waybar` or restart).

## Poll interval

The default `interval: 30` polls `claumeter today --format=waybar` every 30 seconds. Each call reparses `~/.claude/projects/**` — fast (well under a second) but if you want true real-time updates, point the module at the daemon instead:

```jsonc
"custom/claumeter": {
  "exec": "curl -s http://127.0.0.1:7777/today | jq -r '\"\\(.text // \"$\" + (.cost_usd|tostring))\"'",
  "interval": 5,
  "return-type": "json"
}
```

Or start `claumeter serve` and consume `/live` via SSE from a small wrapper script.

## Click to launch TUI

`on-click` opens the TUI for the full breakdown. Adjust the terminal to your setup:

```jsonc
"on-click": "alacritty -e claumeter"      // alacritty
"on-click": "kitty -e claumeter"          // kitty
"on-click": "foot -e claumeter"           // foot
"on-click": "wezterm start claumeter"     // wezterm
```
