import QtQuick
import Quickshell
import Quickshell.Io
import qs.Common
import qs.Widgets
import qs.Modules.Plugins

PluginComponent {
    id: root

    // ------------------------------------------------------------------ shared state
    ClaumeterData {
        id: data
    }

    // ------------------------------------------------------------------ bar pills
    horizontalBarPill: Component {
        Row {
            spacing: Theme.spacingXS

            DankIcon {
                name: "monitoring"
                color: data.loadError ? Theme.error : Theme.primary
                size: Theme.iconSize - 4
                anchors.verticalCenter: parent.verticalCenter
            }

            StyledText {
                text: data.loadError
                      ? "\u2014"
                      : "$" + data.cost.toFixed(2)
                color: data.loadError ? Theme.error : Theme.surfaceText
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
                color: data.loadError ? Theme.error : Theme.primary
                size: Theme.iconSize - 4
                anchors.horizontalCenter: parent.horizontalCenter
            }

            StyledText {
                text: data.loadError ? "\u2014" : "$" + Math.round(data.cost)
                color: data.loadError ? Theme.error : Theme.surfaceText
                font.pixelSize: Theme.fontSizeSmall
                anchors.horizontalCenter: parent.horizontalCenter
            }
        }
    }

    // ------------------------------------------------------------------ popout
    popoutWidth: 560
    popoutHeight: 460

    popoutContent: Component {
        ClaumeterPopout {
            dataRef: data
        }
    }
}
