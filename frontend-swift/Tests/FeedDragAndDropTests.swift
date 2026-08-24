import AppKit
import SwiftUI
import UniformTypeIdentifiers
import XCTest
@testable import MrRSS

@MainActor
final class FeedDragAndDropTests: XCTestCase {
    private var defaults: UserDefaults!
    private var suiteName: String!

    override func setUp() {
        super.setUp()
        suiteName = "MrRSSDragTests-\(UUID().uuidString)"
        defaults = UserDefaults(suiteName: suiteName)
    }

    override func tearDown() {
        defaults.removePersistentDomain(forName: suiteName)
        defaults = nil
        suiteName = nil
        super.tearDown()
    }

    func testTheDraggedPayloadUsesTheApplicationsOwnType() {
        let provider = NSItemProvider()
        provider.register(FeedTransfer(feedID: 1))

        XCTAssertEqual(UTType.mrrssFeed.identifier, "com.devxdojo.mrrss.feed")
        XCTAssertEqual(
            provider.registeredTypeIdentifiers,
            [UTType.mrrssFeed.identifier],
            "A subscription travels under an identifier of our own, not as generic text."
        )
    }

    func testTheDraggedPayloadSurvivesTheRoundTrip() async throws {
        let provider = NSItemProvider()
        provider.register(FeedTransfer(feedID: 42))

        XCTAssertTrue(provider.registeredTypeIdentifiers.contains(UTType.mrrssFeed.identifier))

        let restored: FeedTransfer = try await withCheckedThrowingContinuation { continuation in
            _ = provider.loadTransferable(type: FeedTransfer.self) { result in
                continuation.resume(with: result)
            }
        }

        XCTAssertEqual(restored, FeedTransfer(feedID: 42))
    }

    func testDroppingOnAFolderMovesEveryFeedInTheDrag() async throws {
        let client = DelayedAPIClient()
        let viewModel = try await makeViewModel(client)

        await viewModel.moveFeeds(ids: [1, 3], toFolder: "Tech")

        XCTAssertEqual(client.categoryMutations.map(\.id), [1, 3])
        XCTAssertEqual(Set(client.categoryMutations.map(\.category)), ["Tech"])
        XCTAssertEqual(viewModel.feeds(inFolder: "Tech").map(\.id), [1, 2, 3])
    }

    func testDroppingAFeedThatNoLongerExistsIsIgnored() async throws {
        let client = DelayedAPIClient()
        let viewModel = try await makeViewModel(client)

        await viewModel.moveFeeds(ids: [99, 1], toFolder: "Tech")

        XCTAssertEqual(client.categoryMutations.map(\.id), [1])
        XCTAssertNil(viewModel.errorMessage)
    }

    func testTheSidebarAcceptsSubscriptionDrags() async throws {
        let client = DelayedAPIClient()
        let viewModel = try await makeViewModel(client)

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

        // Without a drop destination the sidebar registers only SwiftUI's own
        // row reordering type.
        let registered = registeredDraggedTypes(in: hostingView)
        XCTAssertTrue(
            registered.contains(NSPasteboard.PasteboardType(UTType.data.identifier)),
            "A folder row should register as a drop target, got \(registered)."
        )
    }

    private func makeViewModel(_ client: DelayedAPIClient) async throws -> AppViewModel {
        client.defaultFeeds = [
            Feed(id: 1, url: "https://a.example.com/feed", title: "Feed A", category: ""),
            Feed(id: 2, url: "https://b.example.com/feed", title: "Feed B", category: "Tech"),
            Feed(id: 3, url: "https://c.example.com/feed", title: "Feed C", category: "")
        ]
        let viewModel = AppViewModel(api: client, autoLoad: false, defaults: defaults)
        viewModel.refreshFeeds()
        try await Task.sleep(for: .milliseconds(250))
        return viewModel
    }

    private func registeredDraggedTypes(in root: NSView) -> Set<NSPasteboard.PasteboardType> {
        var types = Set(root.registeredDraggedTypes)
        for subview in root.subviews {
            types.formUnion(registeredDraggedTypes(in: subview))
        }
        return types
    }
}
