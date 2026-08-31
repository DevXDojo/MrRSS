import AppKit
import Foundation

/// The actions a key can trigger, using the same bindings the previous
/// interface shipped with.
enum ShortcutAction: String, CaseIterable, Identifiable {
    case nextArticle
    case previousArticle
    case toggleReadStatus
    case toggleFavoriteStatus
    case toggleReadLaterStatus
    case openInBrowser
    case toggleContentView
    case refreshFeeds
    case markAllRead
    case addFeed
    case toggleUnreadFilter
    case toggleFavoritesFilter
    case toggleReadLaterFilter
    case goToAllArticles
    case goToUnread
    case goToFavorites
    case goToReadLater

    var id: String { rawValue }

    /// The default binding, written the way the previous interface stored it.
    var defaultBinding: String {
        switch self {
        case .nextArticle: "j"
        case .previousArticle: "k"
        case .toggleReadStatus: "r"
        case .toggleFavoriteStatus: "s"
        case .toggleReadLaterStatus: "l"
        case .openInBrowser: "o"
        case .toggleContentView: "v"
        case .refreshFeeds: "Shift+r"
        case .markAllRead: "Shift+a"
        case .addFeed: "a"
        case .toggleUnreadFilter: "Alt+r"
        case .toggleFavoritesFilter: "Alt+s"
        case .toggleReadLaterFilter: "Alt+l"
        case .goToAllArticles: "1"
        case .goToUnread: "2"
        case .goToFavorites: "3"
        case .goToReadLater: "4"
        }
    }

    var localizedTitle: String {
        switch self {
        case .nextArticle: t("article.navigation.nextArticle")
        case .previousArticle: t("article.navigation.previousArticle")
        case .toggleReadStatus: t("shortcut.toggle.readStatus")
        case .toggleFavoriteStatus: t("article.toolbar.addToFavorite")
        case .toggleReadLaterStatus: t("shortcut.toggle.readLaterStatus")
        case .openInBrowser: t("article.action.openInBrowserShortcut")
        case .toggleContentView: t("shortcut.toggle.contentView")
        case .refreshFeeds: t("article.action.refreshFeedsShortcut")
        case .markAllRead: t("article.action.markAllReadShortcut")
        case .addFeed: t("sidebar.activity.addFeed")
        case .toggleUnreadFilter: t("shortcut.toggle.unreadFilter")
        case .toggleFavoritesFilter: t("shortcut.toggle.favoritesFilter")
        case .toggleReadLaterFilter: t("shortcut.toggle.readLaterFilter")
        case .goToAllArticles: t("article.navigation.goToAllArticles")
        case .goToUnread: t("article.navigation.goToUnread")
        case .goToFavorites: t("article.navigation.goToFavorites")
        case .goToReadLater: t("article.navigation.goToReadLater")
        }
    }
}

/// Turns key presses into actions. Bindings come from the settings the backend
/// stores, falling back to the defaults above.
struct KeyboardShortcutTable {
    private var bindings: [String: ShortcutAction]

    /// Whether key handling is switched on at all.
    let isEnabled: Bool

    init(settings: [String: String] = [:]) {
        // The backend keeps every binding in one JSON object under `shortcuts`,
        // keyed by the same action names the previous interface used.
        var stored: [String: String] = [:]
        if let raw = settings["shortcuts"], let data = raw.data(using: .utf8),
           let decoded = try? JSONDecoder().decode([String: String].self, from: data) {
            stored = decoded
        }

        var table: [String: ShortcutAction] = [:]
        for action in ShortcutAction.allCases {
            let value = stored[action.rawValue]
            let binding = (value?.isEmpty == false ? value! : action.defaultBinding)
            table[KeyboardShortcutTable.normalize(binding)] = action
        }
        bindings = table

        let enabled = settings["shortcuts_enabled"]
        isEnabled = enabled == nil || enabled == "true" || enabled == "1"
    }

    func action(for event: NSEvent) -> ShortcutAction? {
        bindings[KeyboardShortcutTable.combination(for: event)]
    }

    func action(forBinding binding: String) -> ShortcutAction? {
        bindings[KeyboardShortcutTable.normalize(binding)]
    }

    /// Builds the same textual form the previous interface used, so stored
    /// bindings keep working.
    static func combination(for event: NSEvent) -> String {
        var parts = ""
        if event.modifierFlags.contains(.control) { parts += "Ctrl+" }
        if event.modifierFlags.contains(.option) { parts += "Alt+" }
        if event.modifierFlags.contains(.shift) { parts += "Shift+" }
        if event.modifierFlags.contains(.command) { parts += "Meta+" }

        var key = event.charactersIgnoringModifiers ?? ""
        switch event.keyCode {
        case 123: key = "ArrowLeft"
        case 124: key = "ArrowRight"
        case 125: key = "ArrowDown"
        case 126: key = "ArrowUp"
        case 36: key = "Enter"
        case 53: key = "Escape"
        case 49: key = "Space"
        default:
            if key.count == 1 { key = key.lowercased() }
        }

        return normalize(parts + key)
    }

    private static func normalize(_ binding: String) -> String {
        var components = binding.split(separator: "+").map(String.init)
        guard let key = components.popLast() else { return binding }
        let order = ["Ctrl", "Alt", "Shift", "Meta"]
        let modifiers = order.filter { modifier in
            components.contains { $0.caseInsensitiveCompare(modifier) == .orderedSame }
        }
        let normalizedKey = key.count == 1 ? key.lowercased() : key
        return (modifiers + [normalizedKey]).joined(separator: "+")
    }
}
