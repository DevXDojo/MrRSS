import Foundation

/// The tabs the settings window is divided into. The names follow the tabs the
/// previous interface used, plus a connection tab for the server address, which
/// only a separate client needs.
enum SettingsPane: String, CaseIterable, Identifiable {
    case connection
    case feeds
    case general
    case reading
    case typography
    case customization
    case translation
    case summary
    case ai
    case network
    case storage
    case integrations
    case rules
    case statistics
    case about

    var id: String { rawValue }

    var title: String {
        switch self {
        case .connection: t("client.settings.connection")
        case .feeds: t("modal.feed.manageFeeds")
        case .general: t("setting.tab.general")
        case .reading: t("setting.tab.readingAndDisplay")
        case .typography: t("setting.tab.typography")
        case .customization: t("setting.tab.customization")
        case .translation: t("setting.tab.content")
        case .summary: t("article.summary.articleSummary")
        case .ai: t("setting.tab.ai")
        case .network: t("setting.tab.network")
        case .storage: t("setting.database.cleanDatabaseTitle")
        case .integrations: t("setting.tab.plugins")
        case .rules: t("modal.rule.rules")
        case .statistics: t("setting.statistic.statistics")
        case .about: t("setting.tab.about")
        }
    }

    var icon: String {
        switch self {
        case .connection: "server.rack"
        case .feeds: "dot.radiowaves.left.and.right"
        case .general: "gearshape"
        case .reading: "text.book.closed"
        case .typography: "textformat"
        case .customization: "paintbrush"
        case .translation: "character.book.closed"
        case .summary: "text.quote"
        case .ai: "sparkles"
        case .network: "network"
        case .storage: "externaldrive"
        case .integrations: "puzzlepiece.extension"
        case .rules: "bolt.badge.clock"
        case .statistics: "chart.bar"
        case .about: "info.circle"
        }
    }

    /// Panes whose contents are generated from the schema rather than
    /// hand-built.
    var isSettingList: Bool {
        switch self {
        case .connection, .feeds, .rules, .statistics, .about: false
        default: true
        }
    }
}

/// One option of a setting whose value comes from a fixed list.
struct SettingChoice: Identifiable, Hashable {
    let value: String
    let titleKey: String?

    var id: String { value }

    var title: String {
        guard let titleKey else { return value }
        return t(titleKey)
    }
}

/// How a setting's value is entered.
enum SettingControl {
    case toggle
    case text
    case secret
    case number
    case choice
}

/// One setting, as generated from the backend schema.
struct SettingDefinition: Identifiable {
    let key: String
    let pane: SettingsPane
    let control: SettingControl
    let defaultValue: String
    let titleKey: String?
    let fallbackTitle: String
    let detailKey: String?
    let choices: [SettingChoice]

    var id: String { key }

    var title: String {
        guard let titleKey else { return fallbackTitle }
        let translated = t(titleKey)
        return translated == titleKey ? fallbackTitle : translated
    }

    var detail: String? {
        guard let detailKey else { return nil }
        let translated = t(detailKey)
        return translated == detailKey ? nil : translated
    }

    /// True when the value is a password or key that should not be shown.
    var isSecret: Bool {
        if case .secret = control { return true }
        return false
    }
}

enum SettingsCatalog {
    static var definitions: [SettingDefinition] { generated }

    static func definitions(for pane: SettingsPane) -> [SettingDefinition] {
        generated.filter { $0.pane == pane }
    }

    /// The keys whose value is a boolean, used when reading and writing.
    static var boolKeys: Set<String> {
        Set(generated.compactMap { definition in
            if case .toggle = definition.control { return definition.key }
            return nil
        })
    }

    /// Finds the settings matching a search, so the window can offer one.
    static func search(_ query: String) -> [SettingDefinition] {
        let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        guard !trimmed.isEmpty else { return [] }
        return generated.filter { definition in
            definition.key.lowercased().contains(trimmed)
                || definition.title.lowercased().contains(trimmed)
                || (definition.detail?.lowercased().contains(trimmed) ?? false)
        }
    }
}
