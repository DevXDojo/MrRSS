import AppKit
import SwiftUI
import XCTest
@testable import MrRSS

@MainActor
final class SidebarBackgroundMenuTests: XCTestCase {
    private var defaults: UserDefaults!
    private var suiteName: String!

    override func setUp() {
        super.setUp()
        suiteName = "MrRSSBackgroundMenuTests-\(UUID().uuidString)"
        defaults = UserDefaults(suiteName: suiteName)
    }

    override func tearDown() {
        defaults.removePersistentDomain(forName: suiteName)
        defaults = nil
        suiteName = nil
        super.tearDown()
    }

    func testRightClickBelowTheRowsOffersToCreateAFolder() async throws {
        let (window, hostingView) = try await makeSidebar()
        let tableView = try XCTUnwrap(firstDescendant(of: NSTableView.self, in: hostingView))

        let lastRow = tableView.rect(ofRow: max(0, tableView.numberOfRows - 1))
        let emptyPoint = NSPoint(x: 100, y: lastRow.maxY + 60)
        XCTAssertEqual(tableView.row(at: emptyPoint), -1, "The probe should land below the last row.")

        let titles = try XCTUnwrap(menuTitles(at: emptyPoint, in: tableView, window: window))
        XCTAssertTrue(
            titles.contains("New Folder…"),
            "The empty area of the sidebar should offer to create a folder, got \(titles)."
        )
    }

    func testRightClickOnAFeedStillOffersTheFeedActions() async throws {
        let (window, hostingView) = try await makeSidebar()
        let tableView = try XCTUnwrap(firstDescendant(of: NSTableView.self, in: hostingView))

        let feedRow = tableView.rect(ofRow: tableView.numberOfRows - 1)
        let point = NSPoint(x: 100, y: feedRow.midY)
        XCTAssertNotEqual(tableView.row(at: point), -1)

        let titles = try XCTUnwrap(menuTitles(at: point, in: tableView, window: window))
        XCTAssertTrue(
            titles.contains("Delete Feed"),
            "A row keeps its own menu instead of the background one, got \(titles)."
        )
        XCTAssertFalse(titles.contains("New Folder…"))
    }

    private func makeSidebar() async throws -> (NSWindow, NSHostingView<SidebarView>) {
        let client = DelayedAPIClient()
        client.defaultFeeds = [
            Feed(id: 1, url: "https://a.example.com/feed", title: "Feed A", category: ""),
            Feed(id: 2, url: "https://b.example.com/feed", title: "Feed B", category: "")
        ]
        let viewModel = AppViewModel(api: client, autoLoad: false, defaults: defaults)
        viewModel.refreshFeeds()
        try await Task.sleep(for: .milliseconds(250))

        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 260, height: 600),
            styleMask: [.titled, .resizable],
            backing: .buffered,
            defer: false
        )
        let hostingView = NSHostingView(rootView: SidebarView(viewModel: viewModel))
        window.contentView = hostingView
        window.makeKeyAndOrderFront(nil)
        hostingView.layoutSubtreeIfNeeded()
        try await Task.sleep(for: .milliseconds(500))
        return (window, hostingView)
    }

    /// Mirrors how AppKit resolves a context menu: it hit tests the click and
    /// then asks each ancestor in turn until one answers.
    private func menuTitles(at point: NSPoint, in tableView: NSTableView, window: NSWindow) -> [String]? {
        let windowPoint = tableView.convert(point, to: nil)
        guard let event = NSEvent.mouseEvent(
            with: .rightMouseDown,
            location: windowPoint,
            modifierFlags: [],
            timestamp: 0,
            windowNumber: window.windowNumber,
            context: nil,
            eventNumber: 0,
            clickCount: 1,
            pressure: 1
        ) else {
            return nil
        }

        var candidate = window.contentView?.hitTest(windowPoint)
        while let view = candidate {
            if let menu = view.menu(for: event) {
                return menu.items.map(\.title)
            }
            candidate = view.superview
        }
        return nil
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
