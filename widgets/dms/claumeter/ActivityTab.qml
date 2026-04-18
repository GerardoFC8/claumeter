import QtQuick
import QtQuick.Controls
import qs.Common
import qs.Widgets

// Activity tab: days sorted DESC by cost, scrollable.
Item {
    id: root

    property var dataRef: null

    // by_day sorted DESC by cost
    readonly property var days: {
        if (!dataRef || !dataRef.statsData || !dataRef.statsData.by_day) return []
        // by_day is [{day, ...totals flattened, by_model: {model: totals}}].
        // Go's embedded TotalsDTO flattens turns / cost_usd at the top level.
        const arr = dataRef.statsData.by_day.slice()
        arr.sort(function(a, b) {
            return (b.cost_usd || 0) - (a.cost_usd || 0)
        })
        return arr
    }

    Flickable {
        anchors.fill: parent
        contentHeight: dayColumn.implicitHeight
        clip: true
        flickableDirection: Flickable.VerticalFlick

        ScrollBar.vertical: ScrollBar { policy: ScrollBar.AsNeeded }

        Column {
            id: dayColumn
            width: root.width
            spacing: Theme.spacingS

            Repeater {
                model: root.days

                Item {
                    width: root.width
                    height: dayRow.implicitHeight + Theme.spacingXS * 2

                    Column {
                        id: dayRow
                        width: parent.width - Theme.spacingL
                        anchors.verticalCenter: parent.verticalCenter
                        spacing: 2

                        Row {
                            width: parent.width

                            StyledText {
                                text: modelData.day || ""
                                color: Theme.surfaceText
                                font.pixelSize: Theme.fontSizeMedium
                                font.bold: true
                                width: parent.width * 0.6
                                elide: Text.ElideRight
                            }

                            StyledText {
                                text: "$" + (modelData.cost_usd || 0).toFixed(2)
                                color: Theme.primary
                                font.bold: true
                                font.pixelSize: Theme.fontSizeMedium
                                horizontalAlignment: Text.AlignRight
                                width: parent.width * 0.4
                            }
                        }

                        // Top 3 models
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

                    Rectangle {
                        anchors.bottom: parent.bottom
                        width: parent.width
                        height: 1
                        color: Theme.outlineVariant
                        opacity: 0.5
                    }
                }
            }

            // Empty state
            StyledText {
                visible: root.days.length === 0
                width: root.width
                text: dataRef && dataRef.statsData ? "No activity data for this range." : "Loading activity…"
                color: Theme.surfaceVariantText
                font.pixelSize: Theme.fontSizeMedium
                wrapMode: Text.WordWrap
                horizontalAlignment: Text.AlignHCenter
            }
        }
    }
}
