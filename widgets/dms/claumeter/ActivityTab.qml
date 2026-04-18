import QtQuick
import QtQuick.Controls
import qs.Common
import qs.Widgets

Item {
    id: root

    property var dataRef: null

    readonly property var days: {
        if (!dataRef || !dataRef.statsData || !dataRef.statsData.by_day) return []
        const arr = dataRef.statsData.by_day.slice()
        arr.sort(function(a, b) { return (b.cost_usd || 0) - (a.cost_usd || 0) })
        return arr
    }

    // Section header
    Row {
        id: headerRow
        width: parent.width
        height: 24

        StyledText {
            text: "Day"
            color: Theme.primary
            font.pixelSize: Theme.fontSizeSmall
            font.bold: true
            width: parent.width * 0.55
        }
        StyledText {
            text: "Prompts"
            color: Theme.primary
            font.pixelSize: Theme.fontSizeSmall
            font.bold: true
            width: parent.width * 0.2
            horizontalAlignment: Text.AlignRight
        }
        StyledText {
            text: "Cost"
            color: Theme.primary
            font.pixelSize: Theme.fontSizeSmall
            font.bold: true
            width: parent.width * 0.22
            horizontalAlignment: Text.AlignRight
        }
    }

    Flickable {
        anchors {
            top: headerRow.bottom
            topMargin: Theme.spacingXS
            left: parent.left
            right: parent.right
            bottom: parent.bottom
        }
        contentHeight: dayColumn.implicitHeight
        clip: true
        flickableDirection: Flickable.VerticalFlick

        ScrollBar.vertical: ScrollBar { policy: ScrollBar.AsNeeded }

        Column {
            id: dayColumn
            width: root.width
            spacing: 0

            Repeater {
                model: root.days

                Rectangle {
                    width: root.width
                    height: dayRow.implicitHeight + Theme.spacingS * 2
                    color: index % 2 === 0 ? "transparent" : Qt.rgba(Theme.surfaceText.r, Theme.surfaceText.g, Theme.surfaceText.b, 0.04)

                    Column {
                        id: dayRow
                        width: parent.width
                        anchors.verticalCenter: parent.verticalCenter
                        spacing: 3

                        Row {
                            width: parent.width

                            StyledText {
                                text: modelData.day || ""
                                color: Theme.surfaceText
                                font.pixelSize: Theme.fontSizeMedium
                                font.bold: true
                                width: parent.width * 0.55
                                elide: Text.ElideRight
                            }
                            StyledText {
                                text: String(modelData.prompts || 0)
                                color: Theme.surfaceVariantText
                                font.pixelSize: Theme.fontSizeMedium
                                horizontalAlignment: Text.AlignRight
                                width: parent.width * 0.2
                            }
                            StyledText {
                                text: "$" + (modelData.cost_usd || 0).toFixed(2)
                                color: Theme.primary
                                font.bold: true
                                font.pixelSize: Theme.fontSizeMedium
                                horizontalAlignment: Text.AlignRight
                                width: parent.width * 0.22
                            }
                        }

                        Row {
                            width: parent.width
                            spacing: Theme.spacingXS

                            Repeater {
                                model: {
                                    if (!modelData.by_model) return []
                                    const entries = Object.keys(modelData.by_model).map(function(k) {
                                        return { model: k, cost_usd: modelData.by_model[k].cost_usd || 0 }
                                    })
                                    entries.sort(function(a, b) { return b.cost_usd - a.cost_usd })
                                    return entries.slice(0, 3)
                                }

                                StyledText {
                                    text: (dataRef ? dataRef.shortModel(modelData.model) : modelData.model) + " $" + modelData.cost_usd.toFixed(2)
                                    color: Theme.surfaceVariantText
                                    font.pixelSize: Theme.fontSizeSmall
                                }
                            }
                        }
                    }
                }
            }

            StyledText {
                visible: root.days.length === 0
                width: root.width
                text: dataRef && dataRef.statsData ? "No activity data for this range." : "Loading activity…"
                color: Theme.surfaceVariantText
                font.pixelSize: Theme.fontSizeMedium
                wrapMode: Text.WordWrap
                horizontalAlignment: Text.AlignHCenter
                topPadding: Theme.spacingL
            }
        }
    }
}
