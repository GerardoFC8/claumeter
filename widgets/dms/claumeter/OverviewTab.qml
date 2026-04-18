import QtQuick
import qs.Common
import qs.Widgets

Item {
    id: root

    property var dataRef: null
    property bool richMode: false
    property bool degradedMode: !richMode

    readonly property real   resCost:    _resolve("cost_usd",  0)
    readonly property int    resPrompts: _resolve("prompts",   0)
    readonly property int    resTurns:   _resolve("turns",     0)
    readonly property real   resTokens:  _resolveTokens()
    readonly property var    resByModel: _resolveByModel()

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

    Column {
        anchors.fill: parent
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
