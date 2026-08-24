import AppKit
import SwiftUI
import UniformTypeIdentifiers
import XCTest
@testable import MrRSS

@MainActor
final class FeedDropTargetTests: XCTestCase {
    private var defaults: UserDefaults!
    private var suiteName: String!

    override func setUp() {
        super.setUp()
        suiteName = "MrRSSDropTargetTests-\(UUID().uuidString)"
        defaults = UserDefaults(suiteName: suiteName)
    }

    override func tearDown() {
        defaults.removePersistentDomain(forName: suiteName)
        defaults = nil
        suiteName = nil
        super.tearDown()
    }

    func testDroppingOnAFolderFilesTheSubscription() async throws {
        let (client, hostingView) = try await makeSidebar()
        let target = try XCTUnwrap(folderDestination(in: hostingView))

        try await drop(FeedTransfer(feedID: 1), on: target)

        XCTAssertEqual(client.categoryMutations.map(\.id), [1])
        XCTAssertEqual(client.categoryMutations.map(\.category), ["Tech"])
    }

    func testDroppingOnTheSectionHeaderTakesTheSubscriptionOutOfItsFolder() async throws {
        let (client, hostingView) = try await makeSidebar()
        let target = try XCTUnwrap(headerDestination(in: hostingView))

        try await drop(FeedTransfer(feedID: 2), on: target)

        XCTAssertEqual(client.categoryMutations.map(\.id), [2])
        XCTAssertEqual(client.categoryMutations.map(\.category), [""])
    }

    func testTheDraggedTypeMatchesWhatTheTargetsAccept() throws {
        // A type declared with UTType(exportedAs:) carries no conformances
        // unless the bundle declares it too, and then no target accepts it.
        XCTAssertTrue(UTType.mrrssFeed.conforms(to: .data))
        XCTAssertTrue(UTType.mrrssFeed.conforms(to: .item))
        XCTAssertFalse(UTType.mrrssFeed.conforms(to: .plainText))
        XCTAssertNotEqual(UTType.mrrssFeed, .data)
    }

    private func drop(_ transfer: FeedTransfer, on target: NSView) async throws {
        let result = try SidebarDropProbe.drop(transfer, on: target, at: NSPoint(x: 4, y: 4))

        XCTAssertNotEqual(
            result.entered, [],
            "The target refused the drag, so the pointer would show no drop cursor."
        )
        XCTAssertTrue(result.performed)
        try await Task.sleep(for: .milliseconds(600))
    }

    private func makeSidebar() async throws -> (DelayedAPIClient, NSHostingView<SidebarView>) {
        let client = DelayedAPIClient()
        client.defaultFeeds = [
            Feed(id: 1, url: "https://a.example.com/feed", title: "Feed A", category: ""),
            Feed(id: 2, url: "https://b.example.com/feed", title: "Feed B", category: "Tech")
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
        try await Task.sleep(for: .milliseconds(600))
        return (client, hostingView)
    }

    /// A folder covers most of the row; the section header covers its label
    /// only, which is much narrower.
    private func folderDestination(in view: NSView) -> NSView? {
        dropDestinations(in: view).first { $0.bounds.width > 100 }
    }

    private func headerDestination(in view: NSView) -> NSView? {
        dropDestinations(in: view).first { $0.bounds.width <= 100 }
    }

    private func dropDestinations(in view: NSView) -> [NSView] {
        SidebarDropProbe.destinations(in: view)
    }
}
