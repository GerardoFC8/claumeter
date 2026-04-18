import QtQuick
import QtQuick.Controls
import qs.Common
import qs.Widgets

Item {
    id: root

    property var dataRef: null

    readonly property var sessions: {
        if (!dataRef || !dataRef.statsData || !dataRef.statsData.by_session) return []
        const arr = dataRef.statsData.by_session.slice()
        arr.sort(function(a, b) { return (b.cost_usd || 0) - (a.cost_usd || 0) })
        return arr
    }

    // Section header
    Row {
        id: headerRow
        width: parent.width
        height: 24

        StyledText {
            text: "Session"
            color: Theme.primary
            font.pixelSize: Theme.fontSizeSmall
            font.bold: true
            width: parent.width * 0.45
        }
        StyledText {
            text: "Turns"
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
            width: parent.width * 0.32
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
        contentHeight: sessionColumn.implicitHeight
        clip: true
        flickableDirection: Flickable.VerticalFlick

        ScrollBar.vertical: ScrollBar { policy: ScrollBar.AsNeeded }

        Column {
            id: sessionColumn
            width: root.width
            spacing: 0

            Repeater {
                model: root.sessions

                Rectangle {
                    width: root.width
                    height: sessRow.implicitHeight + Theme.spacingS * 2
                    color: index % 2 === 0 ? "transparent" : Qt.rgba(Theme.surfaceText.r, Theme.surfaceText.g, Theme.surfaceText.b, 0.04)

                    Column {
                        id: sessRow
                        width: parent.width
                        anchors.verticalCenter: parent.verticalCenter
                        spacing: 3

                        Row {
                            width: parent.width

                            StyledText {
                                text: (modelData.session_id || "").substring(0, 8)
                                color: Theme.surfaceText
                                font.pixelSize: Theme.fontSizeMedium
                                font.bold: true
                                font.family: "monospace"
                                width: parent.width * 0.45
                                elide: Text.ElideRight
                            }
                            StyledText {
                                text: String(modelData.turns || 0)
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
                                width: parent.width * 0.32
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
                }
            }

            StyledText {
                visible: root.sessions.length === 0
                width: root.width
                text: dataRef && dataRef.statsData ? "No session data for this range." : "Loading sessions…"
                color: Theme.surfaceVariantText
                font.pixelSize: Theme.fontSizeMedium
                wrapMode: Text.WordWrap
                horizontalAlignment: Text.AlignHCenter
                topPadding: Theme.spacingL
            }
        }
    }
}
