import QtQuick
import Quickshell
import Quickshell.Io
import qs.Common
import qs.Widgets
import qs.Modules.Plugins

PluginComponent {
    id: root

    property real cost: 0
    property int prompts: 0
    property int turns: 0
    property real tokens: 0
    property string topModel: ""
    property var byModel: []
    property string loadError: ""
    property bool loaded: false

    function refresh() {
        if (claumeterProcess.running) {
            return
        }
        claumeterProcess.running = true
    }

    function shortModel(name) {
        if (!name) return ""
        return name.replace(/^claude-/, "")
    }

    function compactTokens(n) {
        if (n >= 1e9) return (n / 1e9).toFixed(2) + "B"
        if (n >= 1e6) return (n / 1e6).toFixed(2) + "M"
        if (n >= 1e3) return (n / 1e3).toFixed(1) + "K"
        return String(Math.round(n))
    }

    Process {
        id: claumeterProcess

        running: false
        command: ["sh", "-c", "export PATH=\"$HOME/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/bin:/opt/homebrew/bin:/usr/local/bin:$PATH\"; claumeter today --format=json"]

        stdout: StdioCollector {
            onStreamFinished: {
                try {
                    const data = JSON.parse(text)
                    root.cost = data.cost_usd || 0
                    root.prompts = data.prompts || 0
                    root.turns = data.turns || 0
                    root.tokens = data.tokens || 0
                    root.topModel = data.top_model || ""
                    root.byModel = data.by_model || []
                    root.loaded = true
                    root.loadError = ""
                } catch (e) {
                    root.loadError = "parse error"
                    root.loaded = true
                }
            }
        }

        stderr: StdioCollector {
            onStreamFinished: {
                if (text && text.length > 0) {
                    root.loadError = "claumeter: " + text.trim()
                    root.loaded = true
                }
            }
        }
    }

    Timer {
        interval: 30000
        running: true
        repeat: true
        triggeredOnStart: true
        onTriggered: root.refresh()
    }

    horizontalBarPill: Component {
        Row {
            spacing: Theme.spacingXS

            DankIcon {
                name: "monitoring"
                color: root.loadError ? Theme.error : Theme.primary
                size: Theme.iconSize - 4
                anchors.verticalCenter: parent.verticalCenter
            }

            StyledText {
                text: root.loadError
                      ? "—"
                      : "$" + root.cost.toFixed(2)
                color: root.loadError ? Theme.error : Theme.surfaceText
                font.pixelSize: Theme.fontSizeMedium
                anchors.verticalCenter: parent.verticalCenter
            }
        }
    }

    verticalBarPill: Component {
        Column {
            spacing: Theme.spacingXS

            DankIcon {
                name: "monitoring"
                color: root.loadError ? Theme.error : Theme.primary
                size: Theme.iconSize - 4
                anchors.horizontalCenter: parent.horizontalCenter
            }

            StyledText {
                text: root.loadError ? "—" : "$" + Math.round(root.cost)
                color: root.loadError ? Theme.error : Theme.surfaceText
                font.pixelSize: Theme.fontSizeSmall
                anchors.horizontalCenter: parent.horizontalCenter
            }
        }
    }

    popoutWidth: 380
    popoutHeight: 320

    popoutContent: Component {
        PopoutComponent {
            headerText: "claumeter · today"
            detailsText: root.topModel ? "Top: " + root.shortModel(root.topModel) : ""
            showCloseButton: true

            Column {
                width: parent.width
                spacing: Theme.spacingM

                Row {
                    width: parent.width
                    spacing: Theme.spacingL

                    Column {
                        spacing: 2
                        StyledText {
                            text: "Cost"
                            color: Theme.surfaceVariantText
                            font.pixelSize: Theme.fontSizeSmall
                        }
                        StyledText {
                            text: "$" + root.cost.toFixed(2)
                            color: Theme.primary
                            font.pixelSize: Theme.fontSizeLarge
                            font.bold: true
                        }
                    }

                    Column {
                        spacing: 2
                        StyledText {
                            text: "Prompts"
                            color: Theme.surfaceVariantText
                            font.pixelSize: Theme.fontSizeSmall
                        }
                        StyledText {
                            text: String(root.prompts)
                            color: Theme.surfaceText
                            font.pixelSize: Theme.fontSizeLarge
                            font.bold: true
                        }
                    }

                    Column {
                        spacing: 2
                        StyledText {
                            text: "Turns"
                            color: Theme.surfaceVariantText
                            font.pixelSize: Theme.fontSizeSmall
                        }
                        StyledText {
                            text: String(root.turns)
                            color: Theme.surfaceText
                            font.pixelSize: Theme.fontSizeLarge
                            font.bold: true
                        }
                    }

                    Column {
                        spacing: 2
                        StyledText {
                            text: "Tokens"
                            color: Theme.surfaceVariantText
                            font.pixelSize: Theme.fontSizeSmall
                        }
                        StyledText {
                            text: root.compactTokens(root.tokens)
                            color: Theme.surfaceText
                            font.pixelSize: Theme.fontSizeLarge
                            font.bold: true
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
                    model: root.byModel

                    Row {
                        width: parent.width
                        spacing: Theme.spacingM

                        StyledText {
                            text: root.shortModel(modelData.model)
                            color: Theme.surfaceText
                            width: parent.width * 0.55
                            elide: Text.ElideRight
                            anchors.verticalCenter: parent.verticalCenter
                        }
                        StyledText {
                            text: String(modelData.turns) + " turns"
                            color: Theme.surfaceVariantText
                            font.pixelSize: Theme.fontSizeSmall
                            width: parent.width * 0.2
                            anchors.verticalCenter: parent.verticalCenter
                        }
                        StyledText {
                            text: "$" + modelData.cost_usd.toFixed(2)
                            color: Theme.primary
                            font.bold: true
                            anchors.verticalCenter: parent.verticalCenter
                        }
                    }
                }

                Item {
                    width: parent.width
                    height: Theme.spacingM
                }

                StyledText {
                    visible: root.loadError.length > 0
                    text: root.loadError
                    color: Theme.error
                    font.pixelSize: Theme.fontSizeSmall
                    wrapMode: Text.WordWrap
                    width: parent.width
                }
            }
        }
    }
}
