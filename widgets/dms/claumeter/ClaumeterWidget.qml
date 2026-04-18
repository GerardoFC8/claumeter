import QtQuick
import Quickshell
import Quickshell.Io
import qs.Common
import qs.Widgets
import qs.Modules.Plugins

PluginComponent {
    id: root

    // ------------------------------------------------------------------ shared state
    // NOTE: id must NOT be "data" — Item.data is the reserved default
    // property holding child items, and it shadows the id binding.
    ClaumeterData {
        id: store
    }

    // ------------------------------------------------------------------ bar pills
    horizontalBarPill: Component {
        Row {
            spacing: Theme.spacingXS

            DankIcon {
                name: "monitoring"
                color: store.loadError ? Theme.error : Theme.primary
                size: Theme.iconSize - 4
                anchors.verticalCenter: parent.verticalCenter
            }

            StyledText {
                text: store.loadError
                      ? "\u2014"
                      : "$" + store.cost.toFixed(2)
                color: store.loadError ? Theme.error : Theme.surfaceText
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
                color: store.loadError ? Theme.error : Theme.primary
                size: Theme.iconSize - 4
                anchors.horizontalCenter: parent.horizontalCenter
            }

            StyledText {
                text: store.loadError ? "\u2014" : "$" + Math.round(store.cost)
                color: store.loadError ? Theme.error : Theme.surfaceText
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
            dataRef: store
        }
    }
}
