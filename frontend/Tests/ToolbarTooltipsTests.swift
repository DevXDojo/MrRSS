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

    @MainActor
    private func makeToolbar(label: String) -> (NSWindow, NSToolbar, NSToolbarItem, TestToolbarDelegate) {
        // A toolbar restores the arrangement saved under its identifier, so each
        // test gets its own to keep them independent.
        let delegate = TestToolbarDelegate(label: label)
        let toolbar = NSToolbar(identifier: "tooltips-\(UUID().uuidString)")
        toolbar.delegate = delegate
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 400, height: 300),
            styleMask: [.titled],
            backing: .buffered,
            defer: false
        )
        window.toolbar = toolbar
        toolbar.insertItem(withItemIdentifier: TestToolbarDelegate.identifier, at: 0)
        return (window, toolbar, toolbar.items[0], delegate)
    }

    @MainActor
    func testTheTooltipFollowsALaterLabelChange() {
        let (window, _, item, _) = makeToolbar(label: "Placeholder")
        let binder = ToolbarTooltipBinder()
        binder.bind(window)

        // SwiftUI rearranges toolbar items in place, so an item can briefly
        // carry a neighbour's label. The description has to follow the label
        // when it settles, without the binder being asked again.
        item.label = "Refresh Feeds"
        XCTAssertEqual(item.toolTip, "Refresh Feeds")

        item.label = "Refresh"
        XCTAssertEqual(item.toolTip, "Refresh", "a corrected label corrects the description")
    }

    @MainActor
    func testAControlThatNamesItsOwnStateIsRedescribed() {
        let (window, _, item, _) = makeToolbar(label: "Mark as Read")
        let binder = ToolbarTooltipBinder()
        binder.bind(window)

        XCTAssertEqual(item.toolTip, "Mark as Read")

        item.label = "Mark as Unread"
        XCTAssertEqual(item.toolTip, "Mark as Unread")
    }

    @MainActor
    func testAnItemTheToolbarNoLongerHoldsIsForgotten() {
        let (window, toolbar, item, _) = makeToolbar(label: "Refresh")
        let binder = ToolbarTooltipBinder()
        binder.bind(window)

        toolbar.removeItem(at: 0)
        binder.bind(window)

        // Nothing should still be writing to an item the toolbar has dropped.
        item.label = "Gone"
        XCTAssertNotEqual(item.toolTip, "Gone")
    }
}

/// Vends one item, which is what `NSToolbar` needs before it will insert any.
private final class TestToolbarDelegate: NSObject, NSToolbarDelegate {
    static let identifier = NSToolbarItem.Identifier("MrRSSTestItem")

    private let label: String

    init(label: String) {
        self.label = label
    }

    func toolbar(
        _ toolbar: NSToolbar,
        itemForItemIdentifier itemIdentifier: NSToolbarItem.Identifier,
        willBeInsertedIntoToolbar flag: Bool
    ) -> NSToolbarItem? {
        let item = NSToolbarItem(itemIdentifier: itemIdentifier)
        item.label = label
        return item
    }

    func toolbarAllowedItemIdentifiers(_ toolbar: NSToolbar) -> [NSToolbarItem.Identifier] {
        [TestToolbarDelegate.identifier]
    }

    /// Empty, so the test decides when the item is inserted.
    func toolbarDefaultItemIdentifiers(_ toolbar: NSToolbar) -> [NSToolbarItem.Identifier] {
        []
    }
}
