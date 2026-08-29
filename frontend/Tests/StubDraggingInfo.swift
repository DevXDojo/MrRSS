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
