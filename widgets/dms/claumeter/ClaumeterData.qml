import QtQuick
import Quickshell.Io
import qs.Common

Item {
    id: root

    // ------------------------------------------------------------------ state
    property bool richMode: false
    property string currentRange: "today"
    property string rangeLabel: "Today"

    property var todayData: null
    property var statsData: null
    property var quotaData: null
    property string loadError: ""

    readonly property real cost:   todayData ? (todayData.cost_usd || 0) : 0
    readonly property bool loaded: todayData !== null

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

    function formatResetIn(seconds) {
        if (!seconds || seconds <= 0) return "now"
        const h = Math.floor(seconds / 3600)
        const m = Math.floor((seconds % 3600) / 60)
        if (h > 0 && m > 0) return h + "h " + m + "m"
        if (h > 0) return h + "h"
        return m + "m"
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
                    fetchTodayRich()
                    fetchStats(currentRange)
                    fetchQuota()
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

    // ------------------------------------------------------------------ fetch /today
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

    // ------------------------------------------------------------------ fetch /quota
    function fetchQuota() {
        if (!root.richMode) return
        const xhr = new XMLHttpRequest()
        xhr.open("GET", "http://127.0.0.1:7777/quota", true)
        xhr.timeout = 5000
        xhr.onreadystatechange = function () {
            if (xhr.readyState !== XMLHttpRequest.DONE) return
            if (xhr.status === 200) {
                try {
                    root.quotaData = JSON.parse(xhr.responseText)
                } catch (e) {
                    root.quotaData = null
                }
            } else {
                root.quotaData = null
            }
        }
        xhr.onerror   = function () { root.quotaData = null }
        xhr.ontimeout = function () { root.quotaData = null }
        xhr.send()
    }

    // ------------------------------------------------------------------ fetch /stats
    readonly property var _validRanges: ["today", "last-7d", "last-30d", "all"]

    function fetchStats(range) {
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

    function fetchStatsDegraded(range) {
        statsProcess.rangeArg = range
        if (!statsProcess.running) {
            statsProcess.running = true
        }
    }

    // ------------------------------------------------------------------ degraded mode processes
    Process {
        id: claumeterProcess
        running: false
        command: ["sh", "-c", "export PATH=\"$HOME/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/bin:/opt/homebrew/bin:/usr/local/bin:$PATH\"; claumeter today --format=json"]

        stdout: StdioCollector {
            onStreamFinished: {
                try {
                    root.todayData = JSON.parse(text)
                    root.loadError = ""
                } catch (e) {
                    root.loadError = "parse error"
                }
            }
        }

        stderr: StdioCollector {
            onStreamFinished: {
                if (text && text.length > 0) root.loadError = "claumeter: " + text.trim()
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
                if (text && text.length > 0) root.loadError = "claumeter stats: " + text.trim()
            }
        }
    }

    function refreshDegraded() {
        if (!claumeterProcess.running) claumeterProcess.running = true
        if (root.currentRange !== "today" && !statsProcess.running) {
            statsProcess.rangeArg = root.currentRange
            statsProcess.running = true
        }
    }

    // ------------------------------------------------------------------ timers
    Timer {
        id: healthTimer
        interval: 10000
        running: true
        repeat: true
        triggeredOnStart: true
        onTriggered: root.probeHealth()
    }

    Timer {
        id: richPollTimer
        interval: 3000
        running: root.richMode
        repeat: true
        onTriggered: root.fetchTodayRich()
    }

    Timer {
        id: quotaTimer
        interval: 60000
        running: root.richMode
        repeat: true
        onTriggered: root.fetchQuota()
    }

    Timer {
        id: degradedTimer
        interval: 30000
        running: !root.richMode
        repeat: true
        triggeredOnStart: true
        onTriggered: root.refreshDegraded()
    }

    Component.onCompleted: {
        fetchStatsDegraded(root.currentRange)
    }
}
