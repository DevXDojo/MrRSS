import AppKit
import UniformTypeIdentifiers
@testable import MrRSS

/// Enough of a dragging session to drive a real drop through AppKit. The drop
/// path cannot be reached from SwiftUI alone, and it is exactly where a
/// mismatch between the dragged type and a target's registered types goes
/// unnoticed: the drag still starts, and nothing accepts it.
final class StubDraggingInfo: NSObject, NSDraggingInfo {
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

enum SidebarDropProbe {
    /// Drops `transfer` at a point given in the target's own coordinates. The
    /// sidebar's views are flipped, so a larger y is further down the screen.
    @MainActor
    static func drop(
        _ transfer: FeedTransfer,
        on target: NSView,
        at point: NSPoint
    ) throws -> (entered: NSDragOperation, performed: Bool) {
        let item = NSPasteboardItem()
        item.setData(try JSONEncoder().encode(transfer), forType: .init(UTType.mrrssFeed.identifier))

        let pasteboard = NSPasteboard(name: .init("MrRSSDropProbe-\(UUID().uuidString)"))
        pasteboard.clearContents()
        pasteboard.writeObjects([item])

        let info = StubDraggingInfo(pasteboard: pasteboard)
        info.draggingDestinationWindow = target.window
        info.draggingLocation = target.convert(point, to: nil)
        info.items = [NSDraggingItem(pasteboardWriter: item)]

        let entered = target.draggingEntered(info)
        _ = target.draggingUpdated(info)
        return (entered, target.performDragOperation(info))
    }

    @MainActor
    static func destinations(in root: NSView) -> [NSView] {
        var found: [NSView] = []
        if !root.registeredDraggedTypes.isEmpty, !(root is NSTableView) {
            found.append(root)
        }
        for subview in root.subviews {
            found.append(contentsOf: destinations(in: subview))
        }
        return found
    }
}
