import AppKit
import XCTest
@testable import MrRSS

/// SwiftUI's `help(_:)` never reaches `NSToolbarItem`, so the tooltips are
/// filled in from the labels. These check that rule.
final class ToolbarTooltipsTests: XCTestCase {
    override func setUp() {
        super.setUp()
        Localization.shared.setLanguage(.english)
    }

    private func item(identifier: String, label: String) -> NSToolbarItem {
        let item = NSToolbarItem(itemIdentifier: .init(identifier))
        item.label = label
        return item
    }

    func testAnItemIsDescribedByItsLabel() {
        let toolbarItem = item(identifier: "A1B2", label: "Mark All as Read")

        XCTAssertEqual(ToolbarTooltips.tooltip(for: toolbarItem), "Mark All as Read")
    }

    func testSpacersAndSeparatorsAreLeftAlone() {
        XCTAssertNil(ToolbarTooltips.tooltip(for: item(identifier: "NSToolbarFlexibleSpaceItem", label: "")))
        XCTAssertNil(ToolbarTooltips.tooltip(for: item(identifier: "A1B2", label: "   ")))
    }

    func testTheSidebarToggleIsDescribedInTheReadersLanguage() {
        let toggle = item(
            identifier: "com.apple.SwiftUI.navigationSplitView.toggleSidebar",
            label: "Hide Sidebar"
        )

        XCTAssertEqual(ToolbarTooltips.tooltip(for: toggle), t("shortcut.toggle.sidebar"))

        Localization.shared.setLanguage(.chineseSimplified)
        XCTAssertEqual(ToolbarTooltips.tooltip(for: toggle), "切换侧边栏")
        Localization.shared.setLanguage(.english)
    }

    func testOtherSystemItemsKeepTheirOwnBehaviour() {
        let separator = item(identifier: "com.apple.SwiftUI.splitViewSeparator-0", label: "Separator")

        XCTAssertNil(ToolbarTooltips.tooltip(for: separator))
    }

    func testApplyingRewritesAStaleTooltip() {
        let toolbar = NSToolbar(identifier: "test")
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 400, height: 300),
            styleMask: [.titled],
            backing: .buffered,
            defer: false
        )
        window.toolbar = toolbar

        // A control that names its own state relabels itself as the state
        // changes, and the tooltip has to follow.
        let toolbarItem = item(identifier: "A1B2", label: "Mark as Read")
        toolbarItem.toolTip = "Mark as Read"
        toolbar.insertItem(withItemIdentifier: toolbarItem.itemIdentifier, at: 0)

        toolbarItem.label = "Mark as Unread"
        ToolbarTooltips.apply(in: window)

        XCTAssertEqual(ToolbarTooltips.tooltip(for: toolbarItem), "Mark as Unread")
    }
}
