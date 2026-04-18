import QtQuick
import qs.Common
import qs.Widgets
import qs.Modules.Plugins

PopoutComponent {
    id: popout

    property var dataRef: null

    headerText: "claumeter \u00b7 " + (dataRef ? dataRef.rangeLabel : "Today")
    showCloseButton: true

    // ------------------------------------------------------------------ range chips
    readonly property var ranges: [
        { key: "today",    label: "Today" },
        { key: "last-7d",  label: "7d"    },
        { key: "last-30d", label: "30d"   },
        { key: "all",      label: "All"   }
    ]

    // ------------------------------------------------------------------ tabs
    readonly property var tabs: [
        { key: 0, label: "Overview"  },
        { key: 1, label: "Activity"  },
        { key: 2, label: "Sessions"  },
        { key: 3, label: "Projects"  },
        { key: 4, label: "Tools"     }
    ]

    Column {
        width: parent.width
        spacing: Theme.spacingS

        // Range chips + plan tag
        Item {
            width: parent.width
            height: 28

            Row {
                id: chipRow
                anchors.left: parent.left
                anchors.verticalCenter: parent.verticalCenter
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
                            onClicked: { if (dataRef) dataRef.fetchStats(modelData.key) }
                        }
                    }
                }
            }

            Rectangle {
                id: planBadge
                visible: dataRef && dataRef.richMode && dataRef.quotaData && dataRef.quotaData.configured
                anchors.right: parent.right
                anchors.verticalCenter: parent.verticalCenter
                height: planBadgeLabel.implicitHeight + 6
                width: planBadgeLabel.implicitWidth + 14
                radius: Theme.cornerRadius
                color: Qt.rgba(Theme.primary.r, Theme.primary.g, Theme.primary.b, 0.18)

                StyledText {
                    id: planBadgeLabel
                    anchors.centerIn: parent
                    text: dataRef && dataRef.quotaData ? (dataRef.quotaData.plan || "") : ""
                    color: Theme.primary
                    font.pixelSize: Theme.fontSizeSmall
                    font.bold: true
                }
            }
        }

        // Tab strip
        Row {
            width: parent.width
            spacing: Theme.spacingXS

            Repeater {
                model: popout.tabs

                Item {
                    id: tabBtn
                    readonly property bool isActive: tabLoader.currentTab === modelData.key
                    height: 32
                    width: tabBtnLabel.implicitWidth + Theme.spacingM * 2

                    StyledText {
                        id: tabBtnLabel
                        anchors.centerIn: parent
                        text: modelData.label
                        color: tabBtn.isActive ? Theme.primary : Theme.surfaceVariantText
                        font.pixelSize: Theme.fontSizeMedium
                        font.bold: tabBtn.isActive
                    }

                    Rectangle {
                        anchors.bottom: parent.bottom
                        width: parent.width
                        height: 2
                        radius: 1
                        color: Theme.primary
                        visible: tabBtn.isActive
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
            id: tabLoader
            width: parent.width
            height: {
                const used = (popout.headerHeight || 40)
                           + (popout.detailsHeight || 0)
                           + 28   // chip row (includes plan badge)
                           + 32   // tab strip
                           + Theme.spacingS * 4
                           + Theme.spacingM * 2
                return Math.max(240, 540 - used)
            }
            property int currentTab: 0

            Loader {
                anchors.fill: parent
                sourceComponent: {
                    switch (tabLoader.currentTab) {
                        case 1: return activityComp
                        case 2: return sessionsComp
                        case 3: return projectsComp
                        case 4: return toolsComp
                        default: return overviewComp
                    }
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
        ActivityTab { dataRef: popout.dataRef }
    }

    Component {
        id: sessionsComp
        SessionsTab { dataRef: popout.dataRef }
    }

    Component {
        id: projectsComp
        ProjectsTab { dataRef: popout.dataRef }
    }

    Component {
        id: toolsComp
        ToolsTab { dataRef: popout.dataRef }
    }
}
