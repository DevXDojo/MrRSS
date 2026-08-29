import AppKit
import SwiftUI
import UniformTypeIdentifiers

/// What the sidebar asks the rest of the application to do.
struct SidebarActions {
    var newFolder: () -> Void
    var renameFolder: (String) -> Void
    var deleteFolder: (String) -> Void
    var newFolderHolding: (Feed) -> Void
    var moveFeed: (Feed, String?) -> Void
    var deleteFeed: (Feed) -> Void
    var placeFeed: (Int, String, Int) -> Void
}

/// The sidebar itself, as an outline view.
///
/// `List` cannot reproduce what the system's own source lists do while
/// something is dragged over them: the gap that opens between rows, the way a
/// long list follows the pointer past its edges, and the folder that lights up
/// underneath. `NSOutlineView` does all of that on its own.
struct SidebarOutline: NSViewRepresentable {
    let feeds: [Feed]
    let folders: [String]
    let counts: UnreadCounts
    @Binding var selection: SidebarItem?
    let actions: SidebarActions

    func makeCoordinator() -> Coordinator {
        Coordinator(selection: $selection, actions: actions)
    }

    func makeNSView(context: Context) -> NSScrollView {
        let outlineView = SidebarOutline.makeOutlineView(coordinator: context.coordinator)

        let scrollView = NSScrollView()
        scrollView.documentView = outlineView
        scrollView.hasVerticalScroller = true
        scrollView.autohidesScrollers = true
        scrollView.drawsBackground = false

        context.coordinator.apply(
            feeds: feeds,
            folders: folders,
            counts: counts,
            selection: selection
        )
        return scrollView
    }

    /// Shared with the tests, so they exercise the outline the application
    /// actually builds rather than one assembled for the occasion.
    static func makeOutlineView(coordinator: Coordinator) -> NSOutlineView {
        let outlineView = NSOutlineView()
        outlineView.style = .sourceList
        outlineView.headerView = nil
        outlineView.rowSizeStyle = .default
        outlineView.indentationPerLevel = 14
        outlineView.floatsGroupRows = false
        outlineView.allowsMultipleSelection = false
        outlineView.autoresizesOutlineColumn = false
        outlineView.dataSource = coordinator
        outlineView.delegate = coordinator
        outlineView.menu = coordinator.makeMenu()

        let column = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("main"))
        column.resizingMask = .autoresizingMask
        outlineView.addTableColumn(column)
        outlineView.outlineTableColumn = column

        // A gap between the rows is the system's own way of showing where
        // something will land, and it brings edge scrolling with it.
        outlineView.draggingDestinationFeedbackStyle = .gap
        outlineView.registerForDraggedTypes([.init(UTType.mrrssFeed.identifier)])
        outlineView.setDraggingSourceOperationMask(.move, forLocal: true)

        coordinator.outlineView = outlineView
        return outlineView
    }

    func updateNSView(_ nsView: NSScrollView, context: Context) {
        context.coordinator.actions = actions
        context.coordinator.apply(
            feeds: feeds,
            folders: folders,
            counts: counts,
            selection: selection
        )
    }

    @MainActor
    final class Coordinator: NSObject, NSOutlineViewDataSource, NSOutlineViewDelegate, NSMenuDelegate {
        weak var outlineView: NSOutlineView?
        var actions: SidebarActions

        private let selection: Binding<SidebarItem?>
        private var roots: [SidebarNode] = []
        private var expandedFolders: Set<String> = []
        private var snapshot: Snapshot?
        private var isApplyingSelection = false

        private struct Snapshot: Equatable {
            let feeds: [Feed]
            let folders: [String]
            let counts: UnreadCounts
        }

        init(selection: Binding<SidebarItem?>, actions: SidebarActions) {
            self.selection = selection
            self.actions = actions
        }

        // MARK: - Contents

        func apply(feeds: [Feed], folders: [String], counts: UnreadCounts, selection item: SidebarItem?) {
            guard let outlineView else { return }

            let incoming = Snapshot(feeds: feeds, folders: folders, counts: counts)
            if incoming != snapshot {
                snapshot = incoming
                roots = SidebarNode.tree(feeds: feeds, folders: folders, counts: counts)
                outlineView.reloadData()
                restoreExpansion()
            }

            applySelection(item)
        }

        private func restoreExpansion() {
            guard let outlineView else { return }
            for root in roots {
                outlineView.expandItem(root)
                for child in root.children where child.folderName.map(expandedFolders.contains) == true {
                    outlineView.expandItem(child)
                }
            }
        }

        private func applySelection(_ item: SidebarItem?) {
            guard let outlineView, let item else { return }
            guard let node = node(for: item) else { return }

            let row = outlineView.row(forItem: node)
            guard row >= 0, outlineView.selectedRow != row else { return }

            isApplyingSelection = true
            outlineView.selectRowIndexes(IndexSet(integer: row), byExtendingSelection: false)
            isApplyingSelection = false
        }

        private func node(for item: SidebarItem) -> SidebarNode? {
            for root in roots {
                for child in root.children {
                    if child.item == item { return child }
                    if let match = child.children.first(where: { $0.item == item }) { return match }
                }
            }
            return nil
        }

        private func node(from item: Any?) -> SidebarNode? {
            item as? SidebarNode
        }

        // MARK: - Data source

        func outlineView(_ outlineView: NSOutlineView, numberOfChildrenOfItem item: Any?) -> Int {
            node(from: item)?.children.count ?? roots.count
        }

        func outlineView(_ outlineView: NSOutlineView, child index: Int, ofItem item: Any?) -> Any {
            node(from: item)?.children[index] ?? roots[index]
        }

        func outlineView(_ outlineView: NSOutlineView, isItemExpandable item: Any) -> Bool {
            !(node(from: item)?.children.isEmpty ?? true)
        }

        // MARK: - Delegate

        func outlineView(_ outlineView: NSOutlineView, isGroupItem item: Any) -> Bool {
            node(from: item)?.isGroup ?? false
        }

        func outlineView(_ outlineView: NSOutlineView, shouldSelectItem item: Any) -> Bool {
            node(from: item)?.item != nil
        }

        func outlineView(_ outlineView: NSOutlineView, viewFor tableColumn: NSTableColumn?, item: Any) -> NSView? {
            guard let node = node(from: item) else { return nil }

            if node.isGroup {
                let identifier = NSUserInterfaceItemIdentifier("SidebarGroup")
                let view = outlineView.makeView(withIdentifier: identifier, owner: self) as? NSTableCellView
                    ?? {
                        let cell = NSTableCellView()
                        let label = NSTextField(labelWithString: "")
                        label.translatesAutoresizingMaskIntoConstraints = false
                        cell.addSubview(label)
                        cell.textField = label
                        NSLayoutConstraint.activate([
                            label.leadingAnchor.constraint(equalTo: cell.leadingAnchor),
                            label.centerYAnchor.constraint(equalTo: cell.centerYAnchor)
                        ])
                        cell.identifier = identifier
                        return cell
                    }()
                view.textField?.stringValue = node.title
                return view
            }

            let view = outlineView.makeView(withIdentifier: SidebarRowView.reuseIdentifier, owner: self) as? SidebarRowView
                ?? {
                    let row = SidebarRowView()
                    row.identifier = SidebarRowView.reuseIdentifier
                    return row
                }()
            view.configure(with: node)
            return view
        }

        func outlineViewSelectionDidChange(_ notification: Notification) {
            guard !isApplyingSelection, let outlineView else { return }
            guard let node = outlineView.item(atRow: outlineView.selectedRow) as? SidebarNode,
                  let item = node.item else {
                return
            }
            selection.wrappedValue = item
        }

        func outlineViewItemDidExpand(_ notification: Notification) {
            guard let folder = (notification.userInfo?["NSObject"] as? SidebarNode)?.folderName else { return }
            expandedFolders.insert(folder)
        }

        func outlineViewItemDidCollapse(_ notification: Notification) {
            guard let folder = (notification.userInfo?["NSObject"] as? SidebarNode)?.folderName else { return }
            expandedFolders.remove(folder)
        }

        // MARK: - Dragging

        func outlineView(_ outlineView: NSOutlineView, pasteboardWriterForItem item: Any) -> NSPasteboardWriting? {
            guard let feed = node(from: item)?.feed,
                  let payload = try? JSONEncoder().encode(FeedTransfer(feedID: feed.id)) else {
                return nil
            }

            let pasteboardItem = NSPasteboardItem()
            pasteboardItem.setData(payload, forType: .init(UTType.mrrssFeed.identifier))
            return pasteboardItem
        }

        func outlineView(
            _ outlineView: NSOutlineView,
            validateDrop info: NSDraggingInfo,
            proposedItem item: Any?,
            proposedChildIndex index: Int
        ) -> NSDragOperation {
            guard draggedFeedID(from: info) != nil else { return [] }

            let target = node(from: item)

            // A subscription is not a container, so a drop aimed at one becomes
            // a drop between the rows either side of it.
            if let target, target.feed != nil, index == NSOutlineViewDropOnItemIndex {
                let parent = outlineView.parent(forItem: target)
                outlineView.setDropItem(parent, dropChildIndex: outlineView.childIndex(forItem: target))
                return .move
            }

            if target == nil {
                guard let section = roots.first(where: { $0.title == SidebarNode.feedsTitle }) else { return [] }
                outlineView.setDropItem(section, dropChildIndex: section.children.count)
                return .move
            }

            switch target?.kind {
            case .folder:
                return .move
            case .group(let title) where title == SidebarNode.feedsTitle:
                return .move
            default:
                return []
            }
        }

        func outlineView(
            _ outlineView: NSOutlineView,
            acceptDrop info: NSDraggingInfo,
            item: Any?,
            childIndex index: Int
        ) -> Bool {
            guard let feedID = draggedFeedID(from: info),
                  let destination = destination(for: node(from: item), childIndex: index, feedID: feedID) else {
                return false
            }

            actions.placeFeed(feedID, destination.folder, destination.index)
            return true
        }

        /// Turns a row and a child index into a folder and a rank inside it.
        /// The Feeds heading lists its folders first, so an index there has to
        /// step over them, and a subscription moving down inside its own list
        /// leaves a place behind it.
        func destination(for target: SidebarNode?, childIndex: Int, feedID: Int) -> (folder: String, index: Int)? {
            guard let target else { return nil }

            let folder: String
            var index: Int

            switch target.kind {
            case .folder(let name):
                folder = name
                index = childIndex == NSOutlineViewDropOnItemIndex ? target.children.count : childIndex
            case .group(let title) where title == SidebarNode.feedsTitle:
                folder = ""
                let folderRows = target.children.filter { $0.folderName != nil }.count
                index = childIndex == NSOutlineViewDropOnItemIndex
                    ? target.children.count - folderRows
                    : max(0, childIndex - folderRows)
            default:
                return nil
            }

            let siblings = snapshot?.feeds.filter { $0.category == folder } ?? []
            if let current = siblings.firstIndex(where: { $0.id == feedID }), current < index {
                index -= 1
            }
            return (folder, max(0, index))
        }

        private func draggedFeedID(from info: NSDraggingInfo) -> Int? {
            guard let data = info.draggingPasteboard.data(forType: .init(UTType.mrrssFeed.identifier)),
                  let transfer = try? JSONDecoder().decode(FeedTransfer.self, from: data) else {
                return nil
            }
            return transfer.feedID
        }

        // MARK: - Menu

        func makeMenu() -> NSMenu {
            let menu = NSMenu()
            menu.delegate = self
            return menu
        }

        func menuNeedsUpdate(_ menu: NSMenu) {
            menu.removeAllItems()
            guard let outlineView else { return }

            let clicked = outlineView.clickedRow >= 0
                ? outlineView.item(atRow: outlineView.clickedRow) as? SidebarNode
                : nil
            for item in menuItems(for: clicked) {
                menu.addItem(item)
            }
        }

        /// The empty area below the rows belongs to no row, so a click there
        /// offers what applies to the sidebar as a whole.
        func menuItems(for node: SidebarNode?) -> [NSMenuItem] {
            switch node?.kind {
            case .folder(let name):
                [
                    item(title: "Rename Folder…") { [weak self] in self?.actions.renameFolder(name) },
                    item(title: "Delete Folder") { [weak self] in self?.actions.deleteFolder(name) }
                ]
            case .feed(let feed):
                [
                    moveMenuItem(for: feed),
                    .separator(),
                    item(title: "Delete Feed") { [weak self] in self?.actions.deleteFeed(feed) }
                ]
            default:
                [item(title: "New Folder…") { [weak self] in self?.actions.newFolder() }]
            }
        }

        private func moveMenuItem(for feed: Feed) -> NSMenuItem {
            let parent = NSMenuItem(title: "Move to Folder", action: nil, keyEquivalent: "")
            let submenu = NSMenu()

            for folder in snapshot?.folders ?? [] {
                let entry = item(title: folder) { [weak self] in self?.actions.moveFeed(feed, folder) }
                entry.isEnabled = feed.category != folder
                submenu.addItem(entry)
            }

            if !(snapshot?.folders.isEmpty ?? true) {
                submenu.addItem(.separator())
            }
            submenu.addItem(item(title: "New Folder…") { [weak self] in self?.actions.newFolderHolding(feed) })

            if !feed.category.isEmpty {
                submenu.addItem(.separator())
                submenu.addItem(item(title: "Remove from Folder") { [weak self] in self?.actions.moveFeed(feed, nil) })
            }

            parent.submenu = submenu
            return parent
        }

        private func item(title: String, action: @escaping () -> Void) -> NSMenuItem {
            let entry = NSMenuItem(title: title, action: #selector(runAction(_:)), keyEquivalent: "")
            entry.target = self
            entry.representedObject = ActionBox(action)
            return entry
        }

        @objc private func runAction(_ sender: NSMenuItem) {
            (sender.representedObject as? ActionBox)?.action()
        }

        private final class ActionBox {
            let action: () -> Void
            init(_ action: @escaping () -> Void) { self.action = action }
        }
    }
}
