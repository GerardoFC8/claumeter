import QtQuick
import Quickshell.Io
import qs.Common

// Central state owner. All tabs and the pill read from this Item.
// Exposes:
//   richMode        - bool: daemon reachable
//   currentRange    - string: "today" | "last-7d" | "last-30d" | "all"
//   rangeLabel      - string: human label for current range
//   todayData       - object: compact /today payload
// TODO(v0.8.0): add quotaData property sourced from GET /quota on the daemon.
//   Shape: { plan, configured, limit_messages, window_seconds, used_in_window, used_pct, reset_in_seconds }
//   Poll every 60s; display a quota pill in the Quickshell widget when configured=true.
//   statsData       - object: full /stats payload (null when range == today)
//   loadError       - string: last error message (empty = ok)
Item {
    id: root

    // ------------------------------------------------------------------ state
    property bool richMode: false
    property string currentRange: "today"
    property string rangeLabel: "Today"

    property var todayData: null
    property var statsData: null
    property string loadError: ""

    // convenience aliases used by the pill
    readonly property real cost:    todayData ? (todayData.cost_usd || 0) : 0
    readonly property bool loaded:  todayData !== null

    // ------------------------------------------------------------------ internals
    property int _healthFailures: 0
    readonly property int _healthThreshold: 2

    // ------------------------------------------------------------------ helpers
    function shortModel(name) {
        if (!name) return ""
        return name.replace(/^claude-/, "")
    }

    function compactTokens(n) {
        if (n >= 1e9) return (n / 1e9).toFixed(2) + "B"
        if (n >= 1e6) return (n / 1e6).toFixed(2) + "M"
        if (n >= 1e3) return (n / 1e3).toFixed(1) + "K"
        return String(Math.round(n))
    }

    function elideMiddlePath(p) {
        if (!p || p.length <= 32) return p || ""
        const parts = p.split("/").filter(s => s.length > 0)
        if (parts.length <= 2) return p
        return "/" + parts[0] + "/.../" + parts[parts.length - 1]
    }

    // ------------------------------------------------------------------ health probe
    function probeHealth() {
        const xhr = new XMLHttpRequest()
        xhr.open("GET", "http://127.0.0.1:7777/healthz", true)
        xhr.timeout = 3000
        xhr.onreadystatechange = function () {
            if (xhr.readyState !== XMLHttpRequest.DONE) return
            if (xhr.status === 200) {
                _healthFailures = 0
                if (!root.richMode) {
                    root.richMode = true
                    // immediate data fetch in rich mode: pill + full stats
                    fetchTodayRich()
                    fetchStats(currentRange)
                }
            } else {
                _handleHealthFailure()
            }
        }
        xhr.onerror   = function () { _handleHealthFailure() }
        xhr.ontimeout = function () { _handleHealthFailure() }
        xhr.send()
    }

    function _handleHealthFailure() {
        _healthFailures++
        if (_healthFailures >= _healthThreshold) {
            root.richMode = false
        }
    }

    // ------------------------------------------------------------------ rich mode fetch: /today
    function fetchTodayRich() {
        const xhr = new XMLHttpRequest()
        xhr.open("GET", "http://127.0.0.1:7777/today", true)
        xhr.timeout = 5000
        xhr.onreadystatechange = function () {
            if (xhr.readyState !== XMLHttpRequest.DONE) return
            if (xhr.status === 200) {
                try {
                    root.todayData = JSON.parse(xhr.responseText)
                    root.loadError = ""
                } catch (e) {
                    root.loadError = "parse error: /today"
                }
            } else {
                root.loadError = "HTTP " + xhr.status + " from /today"
            }
        }
        xhr.onerror   = function () { root.loadError = "network error: /today" }
        xhr.ontimeout = function () { root.loadError = "timeout: /today" }
        xhr.send()
    }

    // ------------------------------------------------------------------ fetch stats (rich + degraded)
    readonly property var _validRanges: ["today", "last-7d", "last-30d", "all"]

    function fetchStats(range) {
        // NOTE: we fetch /stats for EVERY range, including "today". The
        // compact /today payload drives the rapid 3s pill poll (lightweight),
        // but Activity and Sessions tabs need the full by_day / by_session
        // structure, which only /stats returns.
        if (_validRanges.indexOf(range) < 0) return
        const labels = { "today": "Today", "last-7d": "Last 7 days", "last-30d": "Last 30 days", "all": "All time" }
        root.currentRange = range
        root.rangeLabel = labels[range] || range

        if (!root.richMode) {
            fetchStatsDegraded(range)
            return
        }

        const xhr = new XMLHttpRequest()
        xhr.open("GET", "http://127.0.0.1:7777/stats?range=" + range, true)
        xhr.timeout = 8000
        xhr.onreadystatechange = function () {
            if (xhr.readyState !== XMLHttpRequest.DONE) return
            if (xhr.status === 200) {
                try {
                    root.statsData = JSON.parse(xhr.responseText)
                    root.loadError = ""
                } catch (e) {
                    root.loadError = "parse error: /stats"
                }
            } else {
                root.loadError = "HTTP " + xhr.status + " from /stats"
            }
        }
        xhr.onerror   = function () { root.loadError = "network error: /stats" }
        xhr.ontimeout = function () { root.loadError = "timeout: /stats" }
        xhr.send()
    }

    // Degraded mode stats: shells out to `claumeter export --format=json --range=X`
    // which returns the same shape as the /stats endpoint. Runs for every
    // range (including "today") so Activity/Sessions tabs have by_day /
    // by_session data; OverviewTab still uses the lighter todayData for
    // the pill numbers when range === "today".
    function fetchStatsDegraded(range) {
        statsProcess.rangeArg = range
        if (!statsProcess.running) {
            statsProcess.running = true
        }
    }

    // ------------------------------------------------------------------ degraded mode: process
    Process {
        id: claumeterProcess
        running: false
        command: ["sh", "-c", "export PATH=\"$HOME/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/bin:/opt/homebrew/bin:/usr/local/bin:$PATH\"; claumeter today --format=json"]

        stdout: StdioCollector {
            onStreamFinished: {
                try {
                    const data = JSON.parse(text)
                    root.todayData = data
                    root.loadError = ""
                } catch (e) {
                    root.loadError = "parse error"
                }
            }
        }

        stderr: StdioCollector {
            onStreamFinished: {
                if (text && text.length > 0) {
                    root.loadError = "claumeter: " + text.trim()
                }
            }
        }
    }

    Process {
        id: statsProcess
        running: false
        property string rangeArg: "last-7d"
        command: ["sh", "-c", "export PATH=\"$HOME/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/bin:/opt/homebrew/bin:/usr/local/bin:$PATH\"; claumeter export --format=json --range=" + rangeArg]

        stdout: StdioCollector {
            onStreamFinished: {
                try {
                    root.statsData = JSON.parse(text)
                    root.loadError = ""
                } catch (e) {
                    root.loadError = "parse error: stats"
                }
            }
        }

        stderr: StdioCollector {
            onStreamFinished: {
                if (text && text.length > 0) {
                    root.loadError = "claumeter stats: " + text.trim()
                }
            }
        }
    }

    function refreshDegraded() {
        if (!claumeterProcess.running) {
            claumeterProcess.running = true
        }
        if (root.currentRange !== "today" && !statsProcess.running) {
            statsProcess.rangeArg = root.currentRange
            statsProcess.running = true
        }
    }

    // ------------------------------------------------------------------ timers
    // Health probe every 10s
    Timer {
        id: healthTimer
        interval: 10000
        running: true
        repeat: true
        triggeredOnStart: true
        onTriggered: root.probeHealth()
    }

    // Rich mode: poll /today every 3s
    Timer {
        id: richPollTimer
        interval: 3000
        running: root.richMode
        repeat: true
        onTriggered: root.fetchTodayRich()
    }

    // Degraded mode: poll process every 30s
    Timer {
        id: degradedTimer
        interval: 30000
        running: !root.richMode
        repeat: true
        triggeredOnStart: true
        onTriggered: root.refreshDegraded()
    }

    Component.onCompleted: {
        // Prime statsData for Activity/Sessions tabs even before the user
        // touches a range chip. Only matters in degraded mode — if the daemon
        // is up, probeHealth will overwrite via /stats shortly.
        fetchStatsDegraded(root.currentRange)
    }
}
