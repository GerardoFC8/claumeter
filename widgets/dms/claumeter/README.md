# claumeter — DankMaterialShell plugin

Native [DMS](https://danklinux.com/) / Quickshell plugin for niri/Hyprland/sway users. Shows today's Claude Code cost in the DankBar with a popout for the full breakdown.

![Bar pill: "$199.82", popout shows Cost / Prompts / Turns / Tokens plus per-model breakdown]

## Requirements

- [DankMaterialShell](https://github.com/AvengeMedia/DankMaterialShell) **≥ 1.2.0**
- `claumeter` binary on `PATH` (`brew tap GerardoFC8/tap && brew install claumeter`)

## Install

From this repo:

```bash
mkdir -p ~/.config/DankMaterialShell/plugins/claumeter
cp widgets/dms/claumeter/plugin.json \
   widgets/dms/claumeter/ClaumeterWidget.qml \
   ~/.config/DankMaterialShell/plugins/claumeter/
```

Or pull directly from GitHub without cloning the repo:

```bash
mkdir -p ~/.config/DankMaterialShell/plugins/claumeter
curl -sL -o ~/.config/DankMaterialShell/plugins/claumeter/plugin.json \
  https://raw.githubusercontent.com/GerardoFC8/claumeter/main/widgets/dms/claumeter/plugin.json
curl -sL -o ~/.config/DankMaterialShell/plugins/claumeter/ClaumeterWidget.qml \
  https://raw.githubusercontent.com/GerardoFC8/claumeter/main/widgets/dms/claumeter/ClaumeterWidget.qml
```

Then enable the plugin in DMS:

1. Open the DMS settings panel (right-click the bar → Settings, or run `dms settings`).
2. Go to the **Plugins** tab.
3. Enable **claumeter**.
4. Go to **DankBar** → drag the `claumeter` widget into the section you want (left / center / right).

The pill shows `$XX.XX` for today. Click it for a popout with prompts, turns, tokens, and a per-model breakdown. The widget refreshes every 30 seconds.

## Daemon mode (optional, real-time)

By default the widget spawns `claumeter today --format=json` every 30s. If you run `claumeter serve` at startup, swap the `command` in `ClaumeterWidget.qml` to:

```qml
command: ["sh", "-c",
  "curl -s http://127.0.0.1:7777/today"]
```

Now responses are sub-millisecond and you can drop the interval to 2–3 seconds without taxing anything.

For true push updates (no polling at all) the widget would need to subscribe to `/live` (SSE) via a small shim. Left as a follow-up; for now the curl-based poll is plenty fast.

## Notes

- The plugin uses DMS theme tokens (`Theme.primary`, `Theme.surfaceText`, ...) so it inherits your color scheme automatically.
- If the widget shows `—` instead of a dollar amount, check that `claumeter` is on `PATH` for the DMS process. Open the popout to see the error message.
- QML apis used: `Process` / `StdioCollector` from `Quickshell.Io`, `Timer` from QtQuick, `PluginComponent` / `DankIcon` / `StyledText` / `PopoutComponent` from DMS.
