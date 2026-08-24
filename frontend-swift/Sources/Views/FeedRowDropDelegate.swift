import SwiftUI
import UniformTypeIdentifiers

/// Handles a subscription dragged over another row. `DropInfo` reports where
/// the pointer sits inside the row, which is what decides whether the drop
/// lands above or below it; `onDrop(of:isTargeted:)` cannot tell.
struct FeedRowDropDelegate: DropDelegate {
    let rowHeight: CGFloat
    let onHover: (Bool) -> Void
    let onExit: () -> Void
    let onDrop: (Bool, [NSItemProvider]) -> Bool

    func validateDrop(info: DropInfo) -> Bool {
        info.hasItemsConforming(to: [.mrrssFeed])
    }

    func dropEntered(info: DropInfo) {
        onHover(placesAbove(info))
    }

    func dropUpdated(info: DropInfo) -> DropProposal? {
        onHover(placesAbove(info))
        return DropProposal(operation: .move)
    }

    func dropExited(info: DropInfo) {
        onExit()
    }

    func performDrop(info: DropInfo) -> Bool {
        let above = placesAbove(info)
        onExit()
        return onDrop(above, info.itemProviders(for: [.mrrssFeed]))
    }

    private func placesAbove(_ info: DropInfo) -> Bool {
        info.location.y < rowHeight / 2
    }
}

/// Where the sidebar will place a dragged subscription.
struct FeedInsertionPoint: Equatable {
    let feedID: Int
    let above: Bool
}
