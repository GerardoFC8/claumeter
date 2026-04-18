import QtQuick
import qs.Common
import qs.Widgets
import qs.Modules.Plugins

// Rich popout: header, filter chips, tab bar, tab content area.
// Receives dataRef (ClaumeterData) from parent via property.
PopoutComponent {
    id: popout

    property var dataRef: null

    headerText: "claumeter \u00b7 " + (dataRef ? dataRef.rangeLabel : "Today")
    showCloseButton: true

    // ------------------------------------------------------------------ filter chips
    readonly property var ranges: [
        { key: "today",    label: "Today"  },
        { key: "last-7d",  label: "7d"     },
        { key: "last-30d", label: "30d"    },
        { key: "all",      label: "All"    }
    ]

    Column {
        width: parent.width
        spacing: Theme.spacingS

        // Chip row
        Row {
            width: parent.width
            spacing: Theme.spacingXS

            Repeater {
                model: popout.ranges

                Rectangle {
                    id: chip
                    readonly property bool isActive: dataRef ? dataRef.currentRange === modelData.key : modelData.key === "today"
                    height: 28
                    width: chipLabel.implicitWidth + Theme.spacingM * 2
                    radius: Theme.cornerRadius
                    color: chip.isActive ? Theme.primary : Theme.surfaceContainerHigh

                    StyledText {
                        id: chipLabel
                        anchors.centerIn: parent
                        text: modelData.label
                        color: chip.isActive ? (Theme.onPrimary !== undefined ? Theme.onPrimary : "#ffffff") : Theme.surfaceText
                        font.pixelSize: Theme.fontSizeSmall
                    }

                    MouseArea {
                        anchors.fill: parent
                        cursorShape: Qt.PointingHandCursor
                        onClicked: {
                            if (dataRef) dataRef.fetchStats(modelData.key)
                        }
                    }
                }
            }
        }

        // Tab bar
        Row {
            width: parent.width
            spacing: Theme.spacingXS

            Repeater {
                model: dataRef && dataRef.richMode
                       ? [{ key: 0, label: "Overview" }, { key: 1, label: "Activity" }, { key: 2, label: "Sessions" }]
                       : [{ key: 0, label: "Overview" }]

                Rectangle {
                    id: tabBtn
                    readonly property bool isActive: tabLoader.currentTab === modelData.key
                    height: 26
                    width: tabBtnLabel.implicitWidth + Theme.spacingM * 2
                    radius: Theme.cornerRadius
                    color: tabBtn.isActive ? Qt.rgba(Theme.primary.r, Theme.primary.g, Theme.primary.b, 0.15) : "transparent"
                    border.width: tabBtn.isActive ? 1 : 0
                    border.color: Theme.primary

                    StyledText {
                        id: tabBtnLabel
                        anchors.centerIn: parent
                        text: modelData.label
                        color: tabBtn.isActive ? Theme.primary : Theme.surfaceVariantText
                        font.pixelSize: Theme.fontSizeSmall
                    }

                    MouseArea {
                        anchors.fill: parent
                        cursorShape: Qt.PointingHandCursor
                        onClicked: tabLoader.currentTab = modelData.key
                    }
                }
            }
        }

        // Tab content
        Item {
            width: parent.width
            // height fills remaining popout space: total - header - details - chips - tabs - spacing
            height: {
                const used = popout.headerHeight
                           + popout.detailsHeight
                           + 28   // chip row
                           + 26   // tab bar
                           + Theme.spacingS * 3
                           + Theme.spacingM * 2  // PopoutComponent internal padding estimate
                return Math.max(120, 460 - used)
            }

            property int currentTab: 0
            id: tabLoader

            Loader {
                anchors.fill: parent
                sourceComponent: {
                    if (tabLoader.currentTab === 1) return activityComp
                    if (tabLoader.currentTab === 2) return sessionsComp
                    return overviewComp
                }
            }
        }
    }

    // ------------------------------------------------------------------ tab components
    Component {
        id: overviewComp
        OverviewTab {
            dataRef: popout.dataRef
            richMode: popout.dataRef ? popout.dataRef.richMode : false
        }
    }

    Component {
        id: activityComp
        ActivityTab {
            dataRef: popout.dataRef
        }
    }

    Component {
        id: sessionsComp
        SessionsTab {
            dataRef: popout.dataRef
        }
    }
}
