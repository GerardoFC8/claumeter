import QtQuick
import QtQuick.Controls
import qs.Common
import qs.Widgets

Item {
    id: root

    property var dataRef: null
    property bool richMode: false
    property bool degradedMode: !richMode

    readonly property real   resCost:      _resolve("cost_usd",  0)
    readonly property int    resPrompts:   _resolve("prompts",   0)
    readonly property int    resTurns:     _resolve("turns",     0)
    readonly property real   resTokens:    _resolveTokens()
    readonly property var    resByModel:   _resolveByModel()
    readonly property var    resBreakdown: _resolveBreakdown()

    function _resolve(field, fallback) {
        if (!dataRef) return fallback
        const range = dataRef.currentRange
        if (range === "today" || !dataRef.statsData) {
            return dataRef.todayData ? (dataRef.todayData[field] || fallback) : fallback
        }
        return dataRef.statsData.overall ? (dataRef.statsData.overall[field] || fallback) : fallback
    }

    function _resolveTokens() {
        if (!dataRef) return 0
        const range = dataRef.currentRange
        if (range === "today" || !dataRef.statsData) {
            return dataRef.todayData ? (dataRef.todayData.tokens || 0) : 0
        }
        return dataRef.statsData.overall ? (dataRef.statsData.overall.total_tokens || 0) : 0
    }

    function _resolveByModel() {
        if (!dataRef) return []
        const range = dataRef.currentRange
        if (range === "today" || !dataRef.statsData) {
            return dataRef.todayData ? (dataRef.todayData.by_model || []) : []
        }
        if (!dataRef.statsData.by_model) return []
        return dataRef.statsData.by_model.map(function(m) {
            return { model: m.model, turns: m.turns || 0, cost_usd: m.cost_usd || 0 }
        })
    }

    function _resolveBreakdown() {
        if (!dataRef || !dataRef.statsData || !dataRef.statsData.cost_breakdown) return null
        return dataRef.statsData.cost_breakdown
    }

    function _bucketLabel(kind) {
        switch (kind) {
            case "input":          return "Input"
            case "cache_write_5m": return "Cache W (5m)"
            case "cache_write_1h": return "Cache W (1h)"
            case "cache_read":     return "Cache Read"
            case "output":         return "Output"
        }
        return kind
    }

    function _fmtTokens(n) {
        if (!n) return "0"
        if (n >= 1e9) return (n / 1e9).toFixed(2) + "B"
        if (n >= 1e6) return (n / 1e6).toFixed(2) + "M"
        if (n >= 1e3) return (n / 1e3).toFixed(1) + "K"
        return String(n)
    }

    Flickable {
        anchors.fill: parent
        contentHeight: contentCol.implicitHeight
        clip: true
        flickableDirection: Flickable.VerticalFlick

        ScrollBar.vertical: ScrollBar { policy: ScrollBar.AsNeeded }

    Column {
        id: contentCol
        width: root.width
        spacing: Theme.spacingM

        // 4 stat cards
        Row {
            width: parent.width
            spacing: Theme.spacingS

            Repeater {
                model: [
                    { label: "Cost",    value: "$" + root.resCost.toFixed(2),                                      accent: true  },
                    { label: "Prompts", value: String(root.resPrompts),                                             accent: false },
                    { label: "Turns",   value: String(root.resTurns),                                              accent: false },
                    { label: "Tokens",  value: dataRef ? dataRef.compactTokens(root.resTokens) : "0",              accent: false }
                ]

                Rectangle {
                    width: (root.width - Theme.spacingS * 3) / 4
                    height: cardCol.implicitHeight + Theme.spacingM * 2
                    radius: Theme.cornerRadius
                    color: Theme.surfaceContainerHigh

                    Column {
                        id: cardCol
                        anchors {
                            left: parent.left
                            right: parent.right
                            verticalCenter: parent.verticalCenter
                            margins: Theme.spacingM
                        }
                        spacing: 4

                        StyledText {
                            text: modelData.label
                            color: Theme.surfaceVariantText
                            font.pixelSize: Theme.fontSizeSmall
                        }
                        StyledText {
                            text: modelData.value
                            color: modelData.accent ? Theme.primary : Theme.surfaceText
                            font.pixelSize: Theme.fontSizeLarge
                            font.bold: true
                        }
                    }
                }
            }
        }

        // By model section header
        StyledText {
            text: "By model"
            color: Theme.primary
            font.pixelSize: Theme.fontSizeSmall
            font.bold: true
        }

        // By model rows
        Column {
            width: parent.width
            spacing: 0

            Repeater {
                model: root.resByModel

                Rectangle {
                    width: root.width
                    height: modelRow.implicitHeight + Theme.spacingS
                    color: index % 2 === 0 ? "transparent" : Qt.rgba(Theme.surfaceText.r, Theme.surfaceText.g, Theme.surfaceText.b, 0.04)

                    Row {
                        id: modelRow
                        width: parent.width
                        anchors.verticalCenter: parent.verticalCenter
                        spacing: Theme.spacingS

                        StyledText {
                            text: dataRef ? dataRef.shortModel(modelData.model) : modelData.model
                            color: Theme.surfaceText
                            width: root.width * 0.55
                            elide: Text.ElideRight
                            font.pixelSize: Theme.fontSizeMedium
                        }
                        StyledText {
                            text: String(modelData.turns) + " turns"
                            color: Theme.surfaceVariantText
                            font.pixelSize: Theme.fontSizeSmall
                            width: root.width * 0.25
                        }
                        StyledText {
                            text: "$" + (modelData.cost_usd || 0).toFixed(2)
                            color: Theme.primary
                            font.bold: true
                            font.pixelSize: Theme.fontSizeMedium
                            horizontalAlignment: Text.AlignRight
                            width: root.width * 0.18
                        }
                    }
                }
            }
        }

        // Cost breakdown header (overall + per-model, 5 buckets each)
        StyledText {
            visible: root.resBreakdown !== null
            text: "Cost breakdown (Input / Cache / Output)"
            color: Theme.primary
            font.pixelSize: Theme.fontSizeSmall
            font.bold: true
        }

        // Breakdown list: overall first, then each model.
        Column {
            visible: root.resBreakdown !== null
            width: parent.width
            spacing: Theme.spacingS

            Repeater {
                model: {
                    if (!root.resBreakdown) return []
                    const rows = []
                    if (root.resBreakdown.overall) {
                        rows.push({
                            model: "overall",
                            total_cost_usd: root.resBreakdown.overall.total_cost_usd || 0,
                            pct_of_grand_total: 100,
                            buckets: root.resBreakdown.overall.buckets || []
                        })
                    }
                    const by = root.resBreakdown.by_model || []
                    for (let i = 0; i < by.length; i++) {
                        if ((by[i].total_cost_usd || 0) <= 0) continue
                        rows.push(by[i])
                    }
                    return rows
                }

                Rectangle {
                    width: root.width
                    height: bkCol.implicitHeight + Theme.spacingS * 2
                    radius: Theme.cornerRadius
                    color: index === 0
                        ? Qt.rgba(Theme.primary.r, Theme.primary.g, Theme.primary.b, 0.08)
                        : Theme.surfaceContainerHigh

                    Column {
                        id: bkCol
                        anchors {
                            left: parent.left; right: parent.right
                            top: parent.top
                            margins: Theme.spacingS
                        }
                        spacing: 2

                        // Header: model · total $ · % of grand total
                        Row {
                            width: parent.width
                            spacing: Theme.spacingS

                            StyledText {
                                text: modelData.model === "overall"
                                    ? "Overall"
                                    : (dataRef ? dataRef.shortModel(modelData.model) : modelData.model)
                                color: Theme.surfaceText
                                font.bold: true
                                font.pixelSize: Theme.fontSizeMedium
                                width: parent.width * 0.5
                                elide: Text.ElideRight
                            }
                            StyledText {
                                text: "$" + (modelData.total_cost_usd || 0).toFixed(2)
                                color: Theme.primary
                                font.bold: true
                                font.pixelSize: Theme.fontSizeMedium
                                width: parent.width * 0.2
                                horizontalAlignment: Text.AlignRight
                            }
                            StyledText {
                                text: (modelData.pct_of_grand_total || 0).toFixed(1) + "% of total"
                                color: Theme.surfaceVariantText
                                font.pixelSize: Theme.fontSizeSmall
                                width: parent.width * 0.28
                                horizontalAlignment: Text.AlignRight
                            }
                        }

                        // Bucket rows
                        Repeater {
                            model: (modelData.buckets || []).filter(function(b) {
                                return (b.tokens || 0) > 0 || (b.cost_usd || 0) > 0
                            })

                            Row {
                                width: bkCol.width
                                spacing: Theme.spacingS

                                StyledText {
                                    text: root._bucketLabel(modelData.kind)
                                    color: Theme.surfaceVariantText
                                    font.pixelSize: Theme.fontSizeSmall
                                    width: parent.width * 0.22
                                }
                                StyledText {
                                    text: root._fmtTokens(modelData.tokens || 0)
                                    color: Theme.surfaceText
                                    font.pixelSize: Theme.fontSizeSmall
                                    width: parent.width * 0.22
                                    horizontalAlignment: Text.AlignRight
                                }
                                StyledText {
                                    text: "$" + (modelData.rate_usd_per_mtok || 0).toFixed(2) + "/M"
                                    color: Theme.surfaceVariantText
                                    font.pixelSize: Theme.fontSizeSmall
                                    width: parent.width * 0.18
                                    horizontalAlignment: Text.AlignRight
                                }
                                StyledText {
                                    text: "$" + (modelData.cost_usd || 0).toFixed(2)
                                    color: Theme.surfaceText
                                    font.pixelSize: Theme.fontSizeSmall
                                    width: parent.width * 0.18
                                    horizontalAlignment: Text.AlignRight
                                }
                                StyledText {
                                    text: (modelData.pct_of_parent || 0).toFixed(1) + "%"
                                    color: Theme.primary
                                    font.pixelSize: Theme.fontSizeSmall
                                    width: parent.width * 0.18
                                    horizontalAlignment: Text.AlignRight
                                }
                            }
                        }
                    }
                }
            }
        }

        // Degraded mode banner
        Rectangle {
            visible: root.degradedMode
            width: parent.width
            height: bannerText.implicitHeight + Theme.spacingS * 2
            color: Qt.rgba(Theme.primary.r, Theme.primary.g, Theme.primary.b, 0.1)
            radius: Theme.cornerRadius

            StyledText {
                id: bannerText
                anchors {
                    left: parent.left; right: parent.right
                    verticalCenter: parent.verticalCenter
                    margins: Theme.spacingS
                }
                text: "Daemon not running. Run `brew services start claumeter` for live data and more tabs."
                color: Theme.surfaceVariantText
                font.pixelSize: Theme.fontSizeSmall
                wrapMode: Text.WordWrap
            }
        }

        StyledText {
            visible: dataRef ? dataRef.loadError.length > 0 : false
            text: dataRef ? dataRef.loadError : ""
            color: Theme.error
            font.pixelSize: Theme.fontSizeSmall
            wrapMode: Text.WordWrap
            width: parent.width
        }
    }
    }
}
