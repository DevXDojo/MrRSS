import AppKit
import SwiftUI

/// Gives the window's toolbar buttons their hover descriptions.
///
/// SwiftUI's `help(_:)` never reaches `NSToolbarItem`, so a toolbar rendered as
/// icons alone leaves the reader with no way to tell what a button does. The
/// label SwiftUI already puts on each item is the text that belongs there, so it
/// is copied across.
///
/// The copy follows the label through key-value observation rather than being
/// repeated on a schedule. SwiftUI reuses toolbar items as the interface
/// changes, and an item can hold a neighbour's label for an instant while it is
/// being rearranged; watching the label means the tooltip is corrected the
/// moment the label settles, instead of keeping whatever was read in passing.
struct ToolbarTooltips: NSViewRepresentable {
    func makeNSView(context: Context) -> NSView {
        ToolbarTooltipView()
    }

    func updateNSView(_ nsView: NSView, context: Context) {}

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

/// Keeps every toolbar item's tooltip in step with its label.
@MainActor
final class ToolbarTooltipBinder {
    private var observations: [ObjectIdentifier: NSKeyValueObservation] = [:]

    /// Describes the items the toolbar holds now, and forgets the ones it no
    /// longer does.
    func bind(_ window: NSWindow?) {
        guard let toolbar = window?.toolbar else {
            observations.removeAll()
            return
        }

        var present: Set<ObjectIdentifier> = []

        for item in toolbar.items {
            let key = ObjectIdentifier(item)
            present.insert(key)
            ToolbarTooltipBinder.describe(item)

            guard observations[key] == nil else { continue }
            observations[key] = item.observe(\.label, options: [.new]) { item, _ in
                MainActor.assumeIsolated { ToolbarTooltipBinder.describe(item) }
            }
        }

        observations = observations.filter { present.contains($0.key) }
    }

    /// Writes one item's description, leaving it alone when nothing applies.
    static func describe(_ item: NSToolbarItem) {
        guard let tooltip = ToolbarTooltips.tooltip(for: item) else { return }
        guard item.toolTip != tooltip else { return }
        item.toolTip = tooltip
    }
}

/// Watches its window and keeps the toolbar descriptions bound.
private final class ToolbarTooltipView: NSView {
    private let binder = ToolbarTooltipBinder()
    private var observer: NSObjectProtocol?

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()

        if let observer {
            NotificationCenter.default.removeObserver(observer)
            self.observer = nil
        }

        guard let window else {
            binder.bind(nil)
            return
        }

        binder.bind(window)
        // Items are added and removed as the interface changes, most visibly
        // when an article is opened, so the set is re-examined as the window
        // updates. Their wording is then kept current by the observations.
        observer = NotificationCenter.default.addObserver(
            forName: NSWindow.didUpdateNotification,
            object: window,
            queue: .main
        ) { [weak self, weak window] _ in
            MainActor.assumeIsolated { self?.binder.bind(window) }
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
