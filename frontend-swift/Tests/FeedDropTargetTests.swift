import AppKit
import SwiftUI
import UniformTypeIdentifiers
import XCTest
@testable import MrRSS

/// Enough of a dragging session to drive a real drop through AppKit. The drop
/// path cannot be reached from SwiftUI alone, and it is exactly where a
/// mismatch between the dragged type and the target's registered types goes
/// unnoticed: the drag still starts, and nothing accepts it.
private final class StubDraggingInfo: NSObject, NSDraggingInfo {
    var draggingDestinationWindow: NSWindow?
    var draggingSourceOperationMask: NSDragOperation = .move
    var draggingLocation: NSPoint = .zero
    var draggedImageLocation: NSPoint = .zero
    var draggedImage: NSImage?
    var draggingPasteboard: NSPasteboard
    var draggingSource: Any?
    var draggingSequenceNumber: Int = 1
    var draggingFormation: NSDraggingFormation = .default
    var animatesToDestination: Bool = false
    var numberOfValidItemsForDrop: Int = 1
    var springLoadingHighlight: NSSpringLoadingHighlight = .none
    var items: [NSDraggingItem] = []

    init(pasteboard: NSPasteboard) {
        draggingPasteboard = pasteboard
        super.init()
    }

    func slideDraggedImage(to screenPoint: NSPoint) {}

    override func namesOfPromisedFilesDropped(atDestination dropDestination: URL) -> [String]? {
        nil
    }

    func enumerateDraggingItems(
        options enumOpts: NSDraggingItemEnumerationOptions,
        for view: NSView?,
        classes classArray: [AnyClass],
        searchOptions: [NSPasteboard.ReadingOptionKey: Any],
        using block: (NSDraggingItem, Int, UnsafeMutablePointer<ObjCBool>) -> Void
    ) {
        var stop: ObjCBool = false
        for (index, item) in items.enumerated() {
            block(item, index, &stop)
            if stop.boolValue { return }
        }
    }

    func resetSpringLoading() {}
}

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
        let item = NSPasteboardItem()
        item.setData(try JSONEncoder().encode(transfer), forType: .init(UTType.mrrssFeed.identifier))

        let pasteboard = NSPasteboard(name: .init("MrRSSDropTargetTests-\(UUID().uuidString)"))
        pasteboard.clearContents()
        pasteboard.writeObjects([item])

        let info = StubDraggingInfo(pasteboard: pasteboard)
        info.draggingDestinationWindow = target.window
        info.draggingLocation = target.convert(NSPoint(x: 4, y: 4), to: nil)
        info.items = [NSDraggingItem(pasteboardWriter: item)]

        XCTAssertNotEqual(
            target.draggingEntered(info), [],
            "The target refused the drag, so the pointer would show no drop cursor."
        )
        XCTAssertTrue(target.performDragOperation(info))
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
        var found: [NSView] = []
        if !view.registeredDraggedTypes.isEmpty, !(view is NSTableView) {
            found.append(view)
        }
        for subview in view.subviews {
            found.append(contentsOf: dropDestinations(in: subview))
        }
        return found
    }
}
