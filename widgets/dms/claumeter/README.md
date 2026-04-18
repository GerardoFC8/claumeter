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
command: ["sh", "-c", "curl -s http://127.0.0.1:7777/today"]
```

Now responses are sub-millisecond and you can drop the interval to 2–3 seconds without taxing anything.

For true push updates (no polling at all) the widget would need to subscribe to `/live` (SSE) via a small shim. Left as a follow-up; for now the curl-based poll is plenty fast.

## Troubleshooting

### Widget shows `$0.00` / `—` silently

The `qs` process is started by your session manager (niri / systemd), which inherits a minimal `PATH` (`/usr/bin`, `/bin`, `/sbin`, ...). Brew-installed binaries living in `/home/linuxbrew/.linuxbrew/bin` or `/opt/homebrew/bin` are NOT on that `PATH`, even though they work from your interactive shell.

The default `command` already prefixes the common Homebrew locations (`$HOME/.linuxbrew/bin`, `/home/linuxbrew/.linuxbrew/bin`, `/opt/homebrew/bin`, `/usr/local/bin`) before invoking `claumeter`, so `brew install claumeter` users should be covered out of the box on both Linux and macOS. If you installed `claumeter` elsewhere, either:

- Add its directory to the shell `export PATH=` line in `ClaumeterWidget.qml`, or
- Symlink the binary into `/usr/local/bin/` (already on the systemd default path).

Open the popout — the bottom error line will show exactly what the shell saw (e.g. `claumeter: command not found`).

### Changes to the QML don't take effect after reload

DMS's `dms ipc plugins reload <id>` re-parses the manifest but **Qt's QML engine caches parsed components in memory**, and the cache-bust flag used by `loadPlugin` isn't reliable for `file://` URLs on every Qt build. If you edited `ClaumeterWidget.qml` and the widget still behaves like the old version, fully restart DMS:

```bash
systemctl --user restart dms   # if you run DMS as a user service
# or, if you launch it from niri / sway / Hyprland directly:
pkill -f 'qs -p /usr/share/quickshell/dms/quickshell'
# then your compositor's autostart will bring it back up
```

After the restart: Settings → Plugins → **Scan** → toggle claumeter ON.

### `Failed to enable plugin: claumeter` toast

Almost always a QML parse error. Tail the live DMS log:

```bash
tail -f /run/user/$(id -u)/quickshell/by-pid/$(pgrep -o -f 'qs -p /usr/share/quickshell/dms')/log.log
```

The entry `PluginService: component error claumeter <path>:<line> <message>` tells you exactly which property / import is broken.

## Notes

- The plugin uses DMS theme tokens (`Theme.primary`, `Theme.surfaceText`, ...) so it inherits your color scheme automatically.
- QML apis used: `Process` / `StdioCollector` from `Quickshell.Io`, `Timer` from QtQuick, `PluginComponent` / `DankIcon` / `StyledText` / `PopoutComponent` from DMS.
- `DankIcon` exposes `size` (px), not `font.pixelSize` — if you customize the pill, remember the difference between `DankIcon` and `StyledText`.
