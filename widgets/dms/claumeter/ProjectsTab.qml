import QtQuick
import QtQuick.Controls
import qs.Common
import qs.Widgets

Item {
    id: root

    property var dataRef: null

    readonly property var projects: {
        if (!dataRef || !dataRef.statsData || !dataRef.statsData.by_project) return []
        const arr = dataRef.statsData.by_project.slice()
        arr.sort(function(a, b) { return (b.cost_usd || 0) - (a.cost_usd || 0) })
        return arr.slice(0, 10)
    }

    // Section header
    Row {
        id: headerRow
        width: parent.width
        height: 24

        StyledText {
            text: "Project"
            color: Theme.primary
            font.pixelSize: Theme.fontSizeSmall
            font.bold: true
            width: parent.width * 0.5
        }
        StyledText {
            text: "Prompts"
            color: Theme.primary
            font.pixelSize: Theme.fontSizeSmall
            font.bold: true
            width: parent.width * 0.18
            horizontalAlignment: Text.AlignRight
        }
        StyledText {
            text: "Turns"
            color: Theme.primary
            font.pixelSize: Theme.fontSizeSmall
            font.bold: true
            width: parent.width * 0.15
            horizontalAlignment: Text.AlignRight
        }
        StyledText {
            text: "Cost"
            color: Theme.primary
            font.pixelSize: Theme.fontSizeSmall
            font.bold: true
            width: parent.width * 0.15
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
        contentHeight: projColumn.implicitHeight
        clip: true
        flickableDirection: Flickable.VerticalFlick

        ScrollBar.vertical: ScrollBar { policy: ScrollBar.AsNeeded }

        Column {
            id: projColumn
            width: root.width
            spacing: 0

            Repeater {
                model: root.projects

                Rectangle {
                    width: root.width
                    height: projRow.implicitHeight + Theme.spacingS * 2
                    color: index % 2 === 0 ? "transparent" : Qt.rgba(Theme.surfaceText.r, Theme.surfaceText.g, Theme.surfaceText.b, 0.04)

                    Row {
                        id: projRow
                        width: parent.width
                        anchors.verticalCenter: parent.verticalCenter

                        StyledText {
                            text: dataRef ? dataRef.elideMiddlePath(modelData.cwd || "") : (modelData.cwd || "")
                            color: Theme.surfaceText
                            font.pixelSize: Theme.fontSizeMedium
                            elide: Text.ElideMiddle
                            width: parent.width * 0.5
                        }
                        StyledText {
                            text: String(modelData.prompts || 0)
                            color: Theme.surfaceVariantText
                            font.pixelSize: Theme.fontSizeMedium
                            horizontalAlignment: Text.AlignRight
                            width: parent.width * 0.18
                        }
                        StyledText {
                            text: String(modelData.turns || 0)
                            color: Theme.surfaceVariantText
                            font.pixelSize: Theme.fontSizeMedium
                            horizontalAlignment: Text.AlignRight
                            width: parent.width * 0.15
                        }
                        StyledText {
                            text: "$" + (modelData.cost_usd || 0).toFixed(2)
                            color: Theme.primary
                            font.bold: true
                            font.pixelSize: Theme.fontSizeMedium
                            horizontalAlignment: Text.AlignRight
                            width: parent.width * 0.15
                        }
                    }
                }
            }

            StyledText {
                visible: root.projects.length === 0
                width: root.width
                text: dataRef && dataRef.statsData ? "No project data for this range." : "Loading…"
                color: Theme.surfaceVariantText
                font.pixelSize: Theme.fontSizeMedium
                wrapMode: Text.WordWrap
                horizontalAlignment: Text.AlignHCenter
                topPadding: Theme.spacingL
            }
        }
    }
}
