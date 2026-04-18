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
        const arr = dataRef.statsData.by_session.slice()
        arr.sort(function(a, b) {
            const ca = a.totals ? a.totals.cost_usd : 0
            const cb = b.totals ? b.totals.cost_usd : 0
            return cb - ca
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
                                text: "$" + ((modelData.totals && modelData.totals.cost_usd) ? modelData.totals.cost_usd.toFixed(2) : "0.00")
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
                text: dataRef && dataRef.statsData ? "No session data." : "Select a range other than Today for session data."
                color: Theme.surfaceVariantText
                font.pixelSize: Theme.fontSizeMedium
                wrapMode: Text.WordWrap
                horizontalAlignment: Text.AlignHCenter
            }
        }
    }
}
