import AppKit
import SwiftUI

/// Gives the window's toolbar buttons their hover descriptions.
///
/// SwiftUI's `help(_:)` never reaches `NSToolbarItem`, so a toolbar rendered as
/// icons alone leaves the reader with no way to tell what a button does. The
/// label SwiftUI already puts on each item is the text that belongs there, so it
/// is copied across.
///
/// SwiftUI replaces a toolbar item whenever the control it stands for changes —
/// becoming disabled while a refresh runs, for instance — and the replacement
/// arrives with no tooltip again. The window reports each of its update passes,
/// so that is what the copy is driven by.
struct ToolbarTooltips: NSViewRepresentable {
    func makeNSView(context: Context) -> NSView {
        ToolbarTooltipView()
    }

    func updateNSView(_ nsView: NSView, context: Context) {}

    /// Copies each item's label into its tooltip.
    ///
    /// The tooltip is rewritten rather than only filled in, because a control
    /// that describes its own state relabels itself as that state changes: the
    /// read button reads "Mark as Read" until the article is read, and "Mark as
    /// Unread" afterwards.
    static func apply(in window: NSWindow?) {
        guard let toolbar = window?.toolbar else { return }

        for item in toolbar.items {
            guard let tooltip = tooltip(for: item), item.toolTip != tooltip else { continue }
            item.toolTip = tooltip
        }
    }

    /// The text to show for one item, or nil for spacers and separators.
    static func tooltip(for item: NSToolbarItem) -> String? {
        if let known = systemItemTooltips[item.itemIdentifier.rawValue] {
            return known()
        }

        // Anything else macOS supplies is left alone: its wording comes from the
        // system and is not part of this catalogue.
        guard !item.itemIdentifier.rawValue.hasPrefix("com.apple.") else { return nil }

        let label = item.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return label.isEmpty ? nil : label
    }

    /// Items macOS supplies that are worth describing in the reader's language.
    private static let systemItemTooltips: [String: () -> String] = [
        "com.apple.SwiftUI.navigationSplitView.toggleSidebar": { t("shortcut.toggle.sidebar") }
    ]
}

/// Watches its window and refreshes the toolbar tooltips as the window updates.
private final class ToolbarTooltipView: NSView {
    private var observer: NSObjectProtocol?

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()

        if let observer {
            NotificationCenter.default.removeObserver(observer)
            self.observer = nil
        }

        guard let window else { return }

        ToolbarTooltips.apply(in: window)
        observer = NotificationCenter.default.addObserver(
            forName: NSWindow.didUpdateNotification,
            object: window,
            queue: .main
        ) { [weak window] _ in
            MainActor.assumeIsolated { ToolbarTooltips.apply(in: window) }
        }
    }

    deinit {
        if let observer {
            NotificationCenter.default.removeObserver(observer)
        }
    }
}

extension View {
    /// Applies `ToolbarTooltips` behind the view.
    func describingToolbarButtons() -> some View {
        background(ToolbarTooltips())
    }
}
