import AppKit
import SwiftUI
import XCTest
@testable import MrRSS

@MainActor
final class SidebarSelectionTests: XCTestCase {
    func testSidebarMarksTheCurrentItemAsSelected() async throws {
        let client = DelayedAPIClient()
        client.defaultFeeds = [
            Feed(id: 1, url: "https://example.com/1", title: "阮一峰的网络日志", category: ""),
            Feed(id: 2, url: "https://example.com/2", title: "V2EX", category: "")
        ]
        let viewModel = AppViewModel(api: client, autoLoad: false)
        viewModel.refreshFeeds()
        try await Task.sleep(for: .milliseconds(250))
        XCTAssertEqual(viewModel.feeds.count, 2)

        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 260, height: 600),
            styleMask: [.titled, .closable, .resizable],
            backing: .buffered,
            defer: false
        )
        let hostingView = NSHostingView(rootView: SidebarView(viewModel: viewModel))
        window.contentView = hostingView
        window.makeKeyAndOrderFront(nil)
        hostingView.layoutSubtreeIfNeeded()
        try await Task.sleep(for: .milliseconds(500))

        let tableView = try XCTUnwrap(
            firstDescendant(of: NSTableView.self, in: hostingView),
            "The sidebar should be backed by a table view."
        )

        // The default selection is the All Articles filter.
        let defaultSelection = tableView.selectedRowIndexes
        XCTAssertFalse(
            defaultSelection.isEmpty,
            "Clicking a sidebar row must leave the row visibly selected."
        )

        viewModel.selection = .feed(2)
        try await Task.sleep(for: .milliseconds(400))
        hostingView.layoutSubtreeIfNeeded()

        XCTAssertFalse(tableView.selectedRowIndexes.isEmpty)
        XCTAssertNotEqual(
            tableView.selectedRowIndexes,
            defaultSelection,
            "Selecting another sidebar item must move the selection."
        )
    }

    private func firstDescendant<View: NSView>(of type: View.Type, in root: NSView) -> View? {
        if let match = root as? View {
            return match
        }
        for subview in root.subviews {
            if let match = firstDescendant(of: type, in: subview) {
                return match
            }
        }
        return nil
    }
}
