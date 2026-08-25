import AppKit
import Foundation

/// Holds the parts of a drag that must not redraw the sidebar. A drag reports
/// its position many times a second, and storing any of that in view state
/// would rebuild every row on each report.
@MainActor
final class SidebarDragCoordinator {
    /// Distinguishes a drag that left one row for another from one that left
    /// the sidebar altogether.
    var hoverGeneration = 0

    weak var scrollView: NSScrollView?
    private var autoscrollTimer: Timer?

    /// How close to an edge the pointer has to be before the list follows it.
    static let autoscrollMargin: CGFloat = 28

    func beginAutoscroll() {
        guard autoscrollTimer == nil else { return }
        autoscrollTimer = Timer.scheduledTimer(withTimeInterval: 1 / 60, repeats: true) { [weak self] _ in
            MainActor.assumeIsolated { self?.autoscrollStep() }
        }
    }

    func endAutoscroll() {
        autoscrollTimer?.invalidate()
        autoscrollTimer = nil
    }

    private func autoscrollStep() {
        guard let scrollView, let window = scrollView.window else { return }

        let clipView = scrollView.contentView
        let windowPoint = window.convertPoint(fromScreen: NSEvent.mouseLocation)
        let pointer = clipView.convert(windowPoint, from: nil)

        guard let offset = SidebarDragCoordinator.autoscrollOffset(
            currentOffset: clipView.bounds.origin.y,
            visibleHeight: clipView.bounds.height,
            contentHeight: scrollView.documentView?.frame.height ?? 0,
            pointerY: pointer.y,
            margin: SidebarDragCoordinator.autoscrollMargin
        ) else {
            return
        }

        clipView.scroll(to: NSPoint(x: clipView.bounds.origin.x, y: offset))
        scrollView.reflectScrolledClipView(clipView)
    }

    /// The offset the list should move to, or nil when the pointer is clear of
    /// both edges or the list is already as far as it goes. Speed grows the
    /// closer the pointer gets to the edge, which is how the pointer keeps up
    /// with a long list.
    static func autoscrollOffset(
        currentOffset: CGFloat,
        visibleHeight: CGFloat,
        contentHeight: CGFloat,
        pointerY: CGFloat,
        margin: CGFloat
    ) -> CGFloat? {
        let maximumOffset = max(0, contentHeight - visibleHeight)
        guard maximumOffset > 0, margin > 0 else { return nil }

        let topEdge = currentOffset
        let bottomEdge = currentOffset + visibleHeight
        let step: CGFloat

        if pointerY < topEdge + margin, pointerY >= topEdge - margin {
            step = -speed(forDepth: topEdge + margin - pointerY, margin: margin)
        } else if pointerY > bottomEdge - margin, pointerY <= bottomEdge + margin {
            step = speed(forDepth: pointerY - (bottomEdge - margin), margin: margin)
        } else {
            return nil
        }

        let target = min(max(0, currentOffset + step), maximumOffset)
        return target == currentOffset ? nil : target
    }

    private static func speed(forDepth depth: CGFloat, margin: CGFloat) -> CGFloat {
        let fraction = min(max(depth / margin, 0), 1)
        return 2 + fraction * 12
    }
}
