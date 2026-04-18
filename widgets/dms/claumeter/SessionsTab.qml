import QtQuick
import QtQuick.Controls
import qs.Common
import qs.Widgets

// Sessions tab: sessions sorted DESC by cost, scrollable.
Item {
    id: root

    property var dataRef: null

    readonly property var sessions: {
        if (!dataRef || !dataRef.statsData || !dataRef.statsData.by_session) return []
        // by_session is [{session_id, cwd, ...totals flattened}]. Go's
        // embedded TotalsDTO flattens turns / cost_usd at the top level.
        const arr = dataRef.statsData.by_session.slice()
        arr.sort(function(a, b) {
            return (b.cost_usd || 0) - (a.cost_usd || 0)
        })
        return arr
    }

    Flickable {
        anchors.fill: parent
        contentHeight: sessionColumn.implicitHeight
        clip: true
        flickableDirection: Flickable.VerticalFlick

        ScrollBar.vertical: ScrollBar { policy: ScrollBar.AsNeeded }

        Column {
            id: sessionColumn
            width: root.width
            spacing: Theme.spacingS

            Repeater {
                model: root.sessions

                Item {
                    width: root.width
                    height: sessRow.implicitHeight + Theme.spacingXS * 2

                    Column {
                        id: sessRow
                        width: parent.width - Theme.spacingL
                        anchors.verticalCenter: parent.verticalCenter
                        spacing: 2

                        Row {
                            width: parent.width

                            StyledText {
                                text: (modelData.session_id || "").substring(0, 8)
                                color: Theme.surfaceText
                                font.pixelSize: Theme.fontSizeMedium
                                font.bold: true
                                font.family: "monospace"
                                width: parent.width * 0.55
                                elide: Text.ElideRight
                            }

                            StyledText {
                                text: "$" + (modelData.cost_usd || 0).toFixed(2)
                                color: Theme.primary
                                font.bold: true
                                font.pixelSize: Theme.fontSizeMedium
                                horizontalAlignment: Text.AlignRight
                                width: parent.width * 0.45
                            }
                        }

                        StyledText {
                            text: dataRef ? dataRef.elideMiddlePath(modelData.cwd || "") : (modelData.cwd || "")
                            color: Theme.surfaceVariantText
                            font.pixelSize: Theme.fontSizeSmall
                            elide: Text.ElideMiddle
                            width: parent.width
                        }

                        StyledText {
                            text: (modelData.models || []).map(function(m) {
                                return dataRef ? dataRef.shortModel(m) : m
                            }).join(", ")
                            color: Theme.surfaceVariantText
                            font.pixelSize: Theme.fontSizeSmall
                            elide: Text.ElideRight
                            width: parent.width
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
                visible: root.sessions.length === 0
                width: root.width
                text: dataRef && dataRef.statsData ? "No session data for this range." : "Loading sessions…"
                color: Theme.surfaceVariantText
                font.pixelSize: Theme.fontSizeMedium
                wrapMode: Text.WordWrap
                horizontalAlignment: Text.AlignHCenter
            }
        }
    }
}
