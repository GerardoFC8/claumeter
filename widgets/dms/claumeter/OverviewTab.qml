import QtQuick
import qs.Common
import qs.Widgets

// Overview tab: big stat numbers + by_model list.
// Receives data via properties set by ClaumeterPopout.
Item {
    id: root

    property var dataRef: null      // ClaumeterData instance
    property bool richMode: false
    property bool degradedMode: !richMode

    // Resolved values: unify compact and full payload shapes
    readonly property real   resCost:    _resolve("cost_usd",  0)
    readonly property int    resPrompts: _resolve("prompts",   0)
    readonly property int    resTurns:   _resolve("turns",     0)
    readonly property real   resTokens:  _resolveTokens()
    readonly property var    resByModel: _resolveByModel()

    function _resolve(field, fallback) {
        if (!dataRef) return fallback
        const range = dataRef.currentRange
        if (range === "today" || !dataRef.statsData) {
            // compact payload
            return dataRef.todayData ? (dataRef.todayData[field] || fallback) : fallback
        }
        // full payload
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
        // full payload: by_model is [{model, totals}]
        if (!dataRef.statsData.by_model) return []
        return dataRef.statsData.by_model.map(function(m) {
            return {
                model: m.model,
                turns: m.totals ? m.totals.turns : 0,
                cost_usd: m.totals ? m.totals.cost_usd : 0
            }
        })
    }

    Column {
        anchors.fill: parent
        spacing: Theme.spacingM

        // Big stat row
        Row {
            width: parent.width
            spacing: Theme.spacingL

            Repeater {
                model: [
                    { label: "Cost",    value: "$" + root.resCost.toFixed(2),                color: Theme.primary },
                    { label: "Prompts", value: String(root.resPrompts),                       color: Theme.onSurface !== undefined ? Theme.onSurface : Theme.surfaceText },
                    { label: "Turns",   value: String(root.resTurns),                         color: Theme.onSurface !== undefined ? Theme.onSurface : Theme.surfaceText },
                    { label: "Tokens",  value: dataRef ? dataRef.compactTokens(root.resTokens) : "0", color: Theme.onSurface !== undefined ? Theme.onSurface : Theme.surfaceText }
                ]

                Column {
                    spacing: 2
                    StyledText {
                        text: modelData.label
                        color: Theme.surfaceVariantText
                        font.pixelSize: Theme.fontSizeSmall
                    }
                    StyledText {
                        text: modelData.value
                        color: modelData.color
                        font.pixelSize: Theme.fontSizeLarge
                        font.bold: true
                    }
                }
            }
        }

        Rectangle {
            width: parent.width
            height: 1
            color: Theme.outlineVariant
        }

        StyledText {
            text: "By model"
            color: Theme.surfaceVariantText
            font.pixelSize: Theme.fontSizeSmall
        }

        Repeater {
            model: root.resByModel

            Row {
                width: root.width
                spacing: Theme.spacingM

                StyledText {
                    text: dataRef ? dataRef.shortModel(modelData.model) : modelData.model
                    color: Theme.surfaceText
                    width: root.width * 0.55
                    elide: Text.ElideRight
                    anchors.verticalCenter: parent.verticalCenter
                    font.pixelSize: Theme.fontSizeMedium
                }
                StyledText {
                    text: String(modelData.turns) + " turns"
                    color: Theme.surfaceVariantText
                    font.pixelSize: Theme.fontSizeSmall
                    width: root.width * 0.2
                    anchors.verticalCenter: parent.verticalCenter
                }
                StyledText {
                    text: "$" + (modelData.cost_usd || 0).toFixed(2)
                    color: Theme.primary
                    font.bold: true
                    font.pixelSize: Theme.fontSizeMedium
                    anchors.verticalCenter: parent.verticalCenter
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
                    left: parent.left
                    right: parent.right
                    verticalCenter: parent.verticalCenter
                    margins: Theme.spacingS
                }
                text: "Daemon not running. Run `brew services start claumeter` for live data and more tabs."
                color: Theme.surfaceVariantText
                font.pixelSize: Theme.fontSizeSmall
                wrapMode: Text.WordWrap
            }
        }

        // Error line
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
