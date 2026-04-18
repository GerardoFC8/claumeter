import QtQuick
import QtQuick.Controls
import qs.Common
import qs.Widgets

// Tools tab — 2×2 grid: Built-in / MCP servers / Skills / Sub-agents.
// Source: statsData.tools (ToolsDTO from export.go).
// Fields: builtins, mcps, mcp_servers, skills, sub_agents — each [{name,count}].
Item {
    id: root

    property var dataRef: null

    readonly property var toolsData: {
        if (!dataRef || !dataRef.statsData || !dataRef.statsData.tools) return null
        return dataRef.statsData.tools
    }

    function top5(arr) {
        if (!arr || arr.length === 0) return []
        const s = arr.slice()
        s.sort(function(a, b) { return (b.count || 0) - (a.count || 0) })
        return s.slice(0, 5)
    }

    // Category definitions — label + field key in ToolsDTO
    readonly property var categories: [
        { label: "Built-in tools", key: "builtins"    },
        { label: "MCP servers",    key: "mcp_servers" },
        { label: "Skills",         key: "skills"      },
        { label: "Sub-agents",     key: "sub_agents"  }
    ]

    StyledText {
        id: totalLine
        visible: toolsData !== null
        text: toolsData ? "Total tool calls: " + (toolsData.total || 0) : ""
        color: Theme.surfaceVariantText
        font.pixelSize: Theme.fontSizeSmall
        anchors { top: parent.top; left: parent.left; right: parent.right }
        height: visible ? implicitHeight + Theme.spacingXS : 0
    }

    Grid {
        anchors {
            top: totalLine.bottom
            topMargin: Theme.spacingS
            left: parent.left
            right: parent.right
            bottom: parent.bottom
        }
        columns: 2
        columnSpacing: Theme.spacingS
        rowSpacing: Theme.spacingS

        Repeater {
            model: root.categories

            Rectangle {
                width: (root.width - Theme.spacingS) / 2
                height: (root.height - totalLine.height - Theme.spacingS * 3) / 2
                radius: Theme.cornerRadius
                color: Theme.surfaceContainerHigh

                Column {
                    anchors {
                        fill: parent
                        margins: Theme.spacingM
                    }
                    spacing: 4

                    StyledText {
                        text: modelData.label
                        color: Theme.primary
                        font.pixelSize: Theme.fontSizeSmall
                        font.bold: true
                        width: parent.width
                    }

                    // Top 5 entries
                    Repeater {
                        model: {
                            if (!root.toolsData) return []
                            return root.top5(root.toolsData[modelData.key] || [])
                        }

                        Row {
                            width: parent.width
                            spacing: 4

                            StyledText {
                                text: modelData.name || ""
                                color: Theme.surfaceText
                                font.pixelSize: Theme.fontSizeSmall
                                elide: Text.ElideRight
                                width: parent.width * 0.72
                            }
                            StyledText {
                                text: "\u00b7 " + (modelData.count || 0)
                                color: Theme.surfaceVariantText
                                font.pixelSize: Theme.fontSizeSmall
                                horizontalAlignment: Text.AlignRight
                                width: parent.width * 0.26
                            }
                        }
                    }

                    StyledText {
                        visible: !root.toolsData || !(root.toolsData[modelData.key] || []).length
                        text: root.toolsData ? "No data" : "Loading…"
                        color: Theme.surfaceVariantText
                        font.pixelSize: Theme.fontSizeSmall
                    }
                }
            }
        }
    }
}
