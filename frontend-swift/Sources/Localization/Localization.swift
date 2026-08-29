import Foundation

/// The languages the interface can be shown in. The identifiers match the
/// values the backend stores under the `language` setting.
enum AppLanguage: String, CaseIterable, Identifiable {
    case english = "en-US"
    case chineseSimplified = "zh-CN"

    var id: String { rawValue }

    var displayName: String {
        switch self {
        case .english: "English"
        case .chineseSimplified: "简体中文"
        }
    }

    /// The language to use when the stored setting is missing or unknown.
    static var systemDefault: AppLanguage {
        let preferred = Locale.preferredLanguages.first ?? "en"
        return preferred.hasPrefix("zh") ? .chineseSimplified : .english
    }

    static func from(settingValue: String?) -> AppLanguage {
        guard let settingValue, !settingValue.isEmpty else { return .systemDefault }
        if let exact = AppLanguage(rawValue: settingValue) { return exact }
        return settingValue.hasPrefix("zh") ? .chineseSimplified : .english
    }
}

/// Looks up interface strings by dotted key, mirroring the keys the previous
/// frontend used so the wording stays identical.
final class Localization: ObservableObject {
    static let shared = Localization()

    @Published private(set) var language: AppLanguage

    private var tables: [AppLanguage: [String: String]] = [:]

    init(language: AppLanguage = .systemDefault) {
        self.language = language
    }

    func setLanguage(_ language: AppLanguage) {
        guard language != self.language else { return }
        self.language = language
    }

    /// Returns the string for `key`, falling back to English and then to the
    /// key itself so a missing entry is visible rather than blank.
    func string(_ key: String) -> String {
        if let value = table(for: language)[key] {
            return value
        }
        if language != .english, let value = table(for: .english)[key] {
            return value
        }
        return key
    }

    /// Returns the string for `key` with `{name}` placeholders replaced.
    func string(_ key: String, _ arguments: [String: CustomStringConvertible]) -> String {
        var result = string(key)
        for (name, value) in arguments {
            result = result.replacingOccurrences(of: "{\(name)}", with: value.description)
        }
        return result
    }

    private func table(for language: AppLanguage) -> [String: String] {
        if let cached = tables[language] { return cached }
        let json: String
        switch language {
        case .english: json = LocalizationTables.english
        case .chineseSimplified: json = LocalizationTables.chineseSimplified
        }
        let parsed = (try? JSONSerialization.jsonObject(with: Data(json.utf8))) as? [String: String] ?? [:]
        tables[language] = parsed
        return parsed
    }
}

/// Shorthand for `Localization.shared.string(_:)`.
func t(_ key: String) -> String {
    Localization.shared.string(key)
}

/// Shorthand for `Localization.shared.string(_:_:)`.
func t(_ key: String, _ arguments: [String: CustomStringConvertible]) -> String {
    Localization.shared.string(key, arguments)
}
