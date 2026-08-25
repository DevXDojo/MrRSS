import AppKit
import SwiftUI
import UniformTypeIdentifiers
import XCTest
@testable import MrRSS

@MainActor
final class SidebarOutlineTests: XCTestCase {
    private var selection: SidebarItem?
    private var placed: [(id: Int, folder: String, index: Int)] = []
    private var moved: [(feed: Int, folder: String?)] = []
    private var renamed: [String] = []
    private var deletedFolders: [String] = []
    private var deletedFeeds: [Int] = []
    private var newFolderRequests = 0

    private let feeds = [
        Feed(id: 1, url: "https://a.example.com/feed", title: "A", category: "", position: 0),
        Feed(id: 2, url: "https://b.example.com/feed", title: "B", category: "", position: 1),
        Feed(id: 3, url: "https://c.example.com/feed", title: "C", category: "Tech", position: 0),
        Feed(id: 4, url: "https://d.example.com/feed", title: "D", category: "Tech", position: 1)
    ]
    private let folders = ["Tech"]

    // MARK: - Contents

    func testTheOutlineListsLibraryThenFoldersThenLooseSubscriptions() throws {
        let (outlineView, _) = makeOutline()

        XCTAssertEqual(titles(of: outlineView), [
            "Library", "All Articles", "Unread", "Favorites", "Read Later",
            "Feeds", "Tech", "A", "B"
        ])
    }

    func testAFolderRevealsItsSubscriptionsWhenExpanded() throws {
        let (outlineView, _) = makeOutline()
        let folder = try XCTUnwrap(node(named: "Tech", in: outlineView))

        outlineView.expandItem(folder)

        XCTAssertEqual(titles(of: outlineView), [
            "Library", "All Articles", "Unread", "Favorites", "Read Later",
            "Feeds", "Tech", "C", "D", "A", "B"
        ])
    }

    func testHeadingsCarryNoSelection() throws {
        let (outlineView, coordinator) = makeOutline()
        let heading = try XCTUnwrap(node(named: "Feeds", in: outlineView))
        let feed = try XCTUnwrap(node(named: "A", in: outlineView))

        XCTAssertFalse(coordinator.outlineView(outlineView, shouldSelectItem: heading))
        XCTAssertTrue(coordinator.outlineView(outlineView, shouldSelectItem: feed))
        XCTAssertTrue(coordinator.outlineView(outlineView, isGroupItem: heading))
    }

    func testChoosingARowReportsTheSelection() throws {
        let (outlineView, coordinator) = makeOutline()
        let feed = try XCTUnwrap(node(named: "B", in: outlineView))

        outlineView.selectRowIndexes(IndexSet(integer: outlineView.row(forItem: feed)), byExtendingSelection: false)
        coordinator.outlineViewSelectionDidChange(Notification(name: NSOutlineView.selectionDidChangeNotification))

        XCTAssertEqual(selection, .feed(2))
    }

    func testAFolderCarriesTheUnreadCountOfEverythingInside() throws {
        let (outlineView, _) = makeOutline(counts: UnreadCounts(total: 9, feedCounts: [3: 4, 4: 5]))
        let folder = try XCTUnwrap(node(named: "Tech", in: outlineView) as SidebarNode?)

        XCTAssertEqual(folder.badge, 9)
    }

    // MARK: - Dragging

    func testASubscriptionTravelsUnderTheApplicationsOwnType() throws {
        let (outlineView, coordinator) = makeOutline()
        let feed = try XCTUnwrap(node(named: "A", in: outlineView))

        let writer = try XCTUnwrap(coordinator.outlineView(outlineView, pasteboardWriterForItem: feed))

        XCTAssertEqual(writer.writableTypes(for: NSPasteboard.general), [.init(UTType.mrrssFeed.identifier)])
    }

    func testAHeadingOffersNothingToDrag() throws {
        let (outlineView, coordinator) = makeOutline()
        let heading = try XCTUnwrap(node(named: "Library", in: outlineView))

        XCTAssertNil(coordinator.outlineView(outlineView, pasteboardWriterForItem: heading))
    }

    func testADragFromSomewhereElseIsRefused() throws {
        let (outlineView, coordinator) = makeOutline()
        let folder = try XCTUnwrap(node(named: "Tech", in: outlineView))
        let info = draggingInfo(with: Data("not ours".utf8), type: .string)

        let operation = coordinator.outlineView(
            outlineView,
            validateDrop: info,
            proposedItem: folder,
            proposedChildIndex: NSOutlineViewDropOnItemIndex
        )

        XCTAssertEqual(operation, [])
    }

    func testAFolderTakesADropOnItself() throws {
        let (outlineView, coordinator) = makeOutline()
        let folder = try XCTUnwrap(node(named: "Tech", in: outlineView))

        let operation = coordinator.outlineView(
            outlineView,
            validateDrop: draggingInfo(for: 1),
            proposedItem: folder,
            proposedChildIndex: NSOutlineViewDropOnItemIndex
        )

        XCTAssertEqual(operation, .move)
    }

    func testTheLibraryRefusesSubscriptions() throws {
        let (outlineView, coordinator) = makeOutline()
        let library = try XCTUnwrap(node(named: "Library", in: outlineView))

        let operation = coordinator.outlineView(
            outlineView,
            validateDrop: draggingInfo(for: 1),
            proposedItem: library,
            proposedChildIndex: 0
        )

        XCTAssertEqual(operation, [])
    }

    func testDroppingOnAFolderFilesTheSubscriptionAtTheEnd() throws {
        let (outlineView, coordinator) = makeOutline()
        let folder = try XCTUnwrap(node(named: "Tech", in: outlineView))

        let accepted = coordinator.outlineView(
            outlineView,
            acceptDrop: draggingInfo(for: 1),
            item: folder,
            childIndex: NSOutlineViewDropOnItemIndex
        )

        XCTAssertTrue(accepted)
        XCTAssertEqual(placed.count, 1)
        XCTAssertEqual(placed.first?.folder, "Tech")
        XCTAssertEqual(placed.first?.index, 2)
    }

    func testDroppingBetweenTwoFoldersLandsOutsideThemBoth() throws {
        let (outlineView, coordinator) = makeOutline()
        let heading = try XCTUnwrap(node(named: "Feeds", in: outlineView))

        // The heading lists its one folder first, so index 1 is the first
        // subscription that sits outside any folder.
        _ = coordinator.outlineView(outlineView, acceptDrop: draggingInfo(for: 3), item: heading, childIndex: 1)

        XCTAssertEqual(placed.first?.folder, "")
        XCTAssertEqual(placed.first?.index, 0)
    }

    func testMovingDownInsideOneListLeavesAPlaceBehind() throws {
        let (outlineView, coordinator) = makeOutline()
        let heading = try XCTUnwrap(node(named: "Feeds", in: outlineView))

        // A is first of the loose subscriptions; dropping it after B means the
        // row it vacates no longer counts.
        let destination = coordinator.destination(for: heading as? SidebarNode, childIndex: 3, feedID: 1)

        XCTAssertEqual(destination?.folder, "")
        XCTAssertEqual(destination?.index, 1)
    }

    func testMovingUpInsideOneListKeepsTheIndex() throws {
        let (outlineView, coordinator) = makeOutline()
        let heading = try XCTUnwrap(node(named: "Feeds", in: outlineView))

        let destination = coordinator.destination(for: heading as? SidebarNode, childIndex: 1, feedID: 2)

        XCTAssertEqual(destination?.index, 0)
    }

    // MARK: - Menu

    func testTheEmptyAreaOffersToCreateAFolder() {
        let (_, coordinator) = makeOutline()

        let titles = coordinator.menuItems(for: nil).map(\.title)

        XCTAssertEqual(titles, ["New Folder…"])
    }

    func testAFolderOffersToRenameAndDeleteItself() throws {
        let (outlineView, coordinator) = makeOutline()
        let folder = try XCTUnwrap(node(named: "Tech", in: outlineView))

        let items = coordinator.menuItems(for: folder)

        XCTAssertEqual(items.map(\.title), ["Rename Folder…", "Delete Folder"])
        items[1].target?.perform(items[1].action, with: items[1])
        XCTAssertEqual(deletedFolders, ["Tech"])
    }

    func testASubscriptionOffersItsOwnActions() throws {
        let (outlineView, coordinator) = makeOutline()
        let feed = try XCTUnwrap(node(named: "A", in: outlineView))

        let items = coordinator.menuItems(for: feed)

        XCTAssertEqual(items.map(\.title), ["Move to Folder", "", "Delete Feed"])
        XCTAssertEqual(items[0].submenu?.items.map(\.title), ["Tech", "", "New Folder…"])
        items[2].target?.perform(items[2].action, with: items[2])
        XCTAssertEqual(deletedFeeds, [1])
    }

    func testTheMenuIsBuiltFromTheRowThatWasClicked() throws {
        let (outlineView, coordinator) = makeOutline()
        let menu = try XCTUnwrap(outlineView.menu)

        coordinator.menuNeedsUpdate(menu)

        XCTAssertEqual(menu.items.map(\.title), ["New Folder…"])
    }

    // MARK: - Helpers

    private func makeOutline(
        counts: UnreadCounts = .empty
    ) -> (NSOutlineView, SidebarOutline.Coordinator) {
        let binding = Binding<SidebarItem?>(
            get: { [weak self] in self?.selection },
            set: { [weak self] in self?.selection = $0 }
        )
        let coordinator = SidebarOutline.Coordinator(selection: binding, actions: actions)
        let outlineView = SidebarOutline.makeOutlineView(coordinator: coordinator)
        outlineView.frame = NSRect(x: 0, y: 0, width: 260, height: 600)
        coordinator.apply(feeds: feeds, folders: folders, counts: counts, selection: selection)
        return (outlineView, coordinator)
    }

    private var actions: SidebarActions {
        SidebarActions(
            newFolder: { [weak self] in self?.newFolderRequests += 1 },
            renameFolder: { [weak self] in self?.renamed.append($0) },
            deleteFolder: { [weak self] in self?.deletedFolders.append($0) },
            newFolderHolding: { _ in },
            moveFeed: { [weak self] feed, folder in self?.moved.append((feed.id, folder)) },
            deleteFeed: { [weak self] in self?.deletedFeeds.append($0.id) },
            placeFeed: { [weak self] id, folder, index in self?.placed.append((id, folder, index)) }
        )
    }

    private func titles(of outlineView: NSOutlineView) -> [String] {
        (0..<outlineView.numberOfRows).compactMap {
            (outlineView.item(atRow: $0) as? SidebarNode)?.title
        }
    }

    private func node(named title: String, in outlineView: NSOutlineView) -> SidebarNode? {
        for row in 0..<outlineView.numberOfRows {
            if let node = outlineView.item(atRow: row) as? SidebarNode, node.title == title {
                return node
            }
        }
        return nil
    }

    private func draggingInfo(for feedID: Int) -> NSDraggingInfo {
        let payload = try! JSONEncoder().encode(FeedTransfer(feedID: feedID))
        return draggingInfo(with: payload, type: .init(UTType.mrrssFeed.identifier))
    }

    private func draggingInfo(with data: Data, type: NSPasteboard.PasteboardType) -> NSDraggingInfo {
        let pasteboard = NSPasteboard(name: .init("MrRSSOutlineTests-\(UUID().uuidString)"))
        pasteboard.clearContents()
        pasteboard.setData(data, forType: type)
        return StubDraggingInfo(pasteboard: pasteboard)
    }
}
