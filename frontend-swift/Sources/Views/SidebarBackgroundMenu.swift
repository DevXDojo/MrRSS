import AppKit
import SwiftUI

/// Puts a menu on the sidebar's table view so a right click on the empty area
/// below the last row offers an action. SwiftUI's `contextMenu` reaches only
/// the rows themselves; the space beneath them belongs to the table view, which
/// has no SwiftUI representation.
struct SidebarBackgroundMenu: NSViewRepresentable {
    let title: String
    let action: () -> Void
    /// Handed the sidebar's scroll view once the list has built it, so a drag
    /// can follow the pointer past the top and bottom of the visible rows.
    let onResolveScrollView: (NSScrollView) -> Void

    func makeCoordinator() -> Coordinator {
        Coordinator(title: title, action: action)
    }

    func makeNSView(context: Context) -> NSView {
        let view = MenuInstaller()
        view.coordinator = context.coordinator
        return view
    }

    func updateNSView(_ nsView: NSView, context: Context) {
        context.coordinator.action = action
        context.coordinator.onResolveScrollView = onResolveScrollView
        (nsView as? MenuInstaller)?.installMenu()
    }

    func sizeThatFits(_ proposal: ProposedViewSize, nsView: NSView, context: Context) -> CGSize? {
        .zero
    }

    final class Coordinator: NSObject {
        let menu: NSMenu
        var action: () -> Void
        var onResolveScrollView: (NSScrollView) -> Void = { _ in }

        init(title: String, action: @escaping () -> Void) {
            self.action = action
            menu = NSMenu()
            super.init()

            let item = NSMenuItem(title: title, action: #selector(runAction), keyEquivalent: "")
            item.target = self
            menu.addItem(item)
        }

        @objc private func runAction() {
            action()
        }
    }

    final class MenuInstaller: NSView {
        weak var coordinator: Coordinator?

        override func viewDidMoveToWindow() {
            super.viewDidMoveToWindow()
            installMenu()
        }

        /// The table view is built while the list lays out, so the first attempt
        /// can run before it exists.
        func installMenu(remainingAttempts: Int = 5) {
            guard let coordinator else { return }

            if let scrollView = nearestTableView()?.enclosingScrollView {
                // The table view itself discards an assigned menu, and AppKit
                // then looks for one on each ancestor in turn, so the clip view
                // is the first place that can answer for the empty area.
                scrollView.contentView.menu = coordinator.menu
                scrollView.menu = coordinator.menu
                coordinator.onResolveScrollView(scrollView)
                return
            }

            guard remainingAttempts > 0, window != nil else { return }
            DispatchQueue.main.async { [weak self] in
                self?.installMenu(remainingAttempts: remainingAttempts - 1)
            }
        }

        /// Searches outwards one ancestor at a time so the closest table view
        /// wins, which is the sidebar's rather than another column's.
        private func nearestTableView() -> NSTableView? {
            var ancestor = superview
            while let current = ancestor {
                if let tableView = MenuInstaller.firstTableView(in: current) {
                    return tableView
                }
                ancestor = current.superview
            }
            return nil
        }

        private static func firstTableView(in view: NSView) -> NSTableView? {
            if let tableView = view as? NSTableView {
                return tableView
            }
            for subview in view.subviews {
                if let tableView = firstTableView(in: subview) {
                    return tableView
                }
            }
            return nil
        }
    }
}
