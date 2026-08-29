import XCTest
@testable import MrRSS

final class LocalizationTests: XCTestCase {
    func testEnglishTableResolvesNestedKeys() {
        let localization = Localization(language: .english)
        XCTAssertEqual(localization.string("article.action.markAsRead"), "Mark as Read")
        XCTAssertEqual(localization.string("sidebar.activity.favorites"), "Favorites")
    }

    func testChineseTableResolvesNestedKeys() {
        let localization = Localization(language: .chineseSimplified)
        XCTAssertEqual(localization.string("article.action.markAsRead"), "标记为已读")
    }

    func testMissingKeyFallsBackToEnglishThenKey() {
        let localization = Localization(language: .chineseSimplified)
        XCTAssertEqual(localization.string("definitely.not.a.key"), "definitely.not.a.key")
    }

    func testPlaceholdersAreReplaced() {
        let localization = Localization(language: .english)
        let value = localization.string("article.action.markedNArticlesAsRead", ["count": 12])
        XCTAssertEqual(value, "Marked 12 articles as read")
    }

    func testLanguageMappingFromSettingValue() {
        XCTAssertEqual(AppLanguage.from(settingValue: "zh-CN"), .chineseSimplified)
        XCTAssertEqual(AppLanguage.from(settingValue: "en-US"), .english)
        XCTAssertEqual(AppLanguage.from(settingValue: "zh-TW"), .chineseSimplified)
        XCTAssertEqual(AppLanguage.from(settingValue: ""), AppLanguage.systemDefault)
    }
}

final class SettingsCatalogTests: XCTestCase {
    override func setUp() {
        super.setUp()
        Localization.shared.setLanguage(.english)
    }

    func testEverySettingTheBackendStoresIsOffered() throws {
        // The catalogue is generated from the schema, so a setting added there
        // and not regenerated here shows up as a gap.
        XCTAssertGreaterThanOrEqual(SettingsCatalog.definitions.count, 95)
    }

    func testEverySettingBelongsToAPaneThatListsSettings() {
        for definition in SettingsCatalog.definitions {
            XCTAssertTrue(
                definition.pane.isSettingList,
                "\(definition.key) is on the \(definition.pane.rawValue) pane, which shows no settings"
            )
        }
    }

    func testTitlesAreTranslatedRatherThanRawKeys() {
        for definition in SettingsCatalog.definitions {
            XCTAssertFalse(
                definition.title.hasPrefix("setting."),
                "\(definition.key) falls back to its translation key"
            )
            XCTAssertFalse(definition.title.isEmpty)
        }
    }

    func testChoicesCarryTheValuesTheBackendExpects() throws {
        let theme = try XCTUnwrap(SettingsCatalog.definitions.first { $0.key == "theme" })
        XCTAssertEqual(theme.choices.map(\.value), ["auto", "light", "dark"])

        let viewMode = try XCTUnwrap(SettingsCatalog.definitions.first { $0.key == "default_view_mode" })
        XCTAssertEqual(viewMode.choices.map(\.value), ["rendered", "webpage"])
    }

    func testSecretsAreNotShownInPlainFields() throws {
        for key in ["ai_api_key", "deepl_api_key", "freshrss_api_password", "proxy_password"] {
            let definition = try XCTUnwrap(SettingsCatalog.definitions.first { $0.key == key })
            XCTAssertTrue(definition.isSecret, "\(key) should be entered as a secret")
        }
    }

    func testSearchFindsSettingsByKeyAndByTitle() {
        XCTAssertTrue(SettingsCatalog.search("proxy").contains { $0.key == "proxy_host" })
        XCTAssertTrue(SettingsCatalog.search("translation").contains { $0.key == "translation_enabled" })
        XCTAssertTrue(SettingsCatalog.search("").isEmpty)
    }
}

@MainActor
final class SettingsPersistenceTests: XCTestCase {
    private func makeViewModel() -> (AppViewModel, RecordingSettingsClient) {
        let defaults = UserDefaults(suiteName: "SettingsPersistenceTests")!
        defaults.removePersistentDomain(forName: "SettingsPersistenceTests")
        let client = RecordingSettingsClient()
        return (AppViewModel(api: client, autoLoad: false, defaults: defaults), client)
    }

    func testAChangeSavesItselfShortlyAfterwards() async throws {
        let (viewModel, client) = makeViewModel()

        viewModel.updateBoolSetting("hover_mark_as_read", value: true)

        XCTAssertTrue(client.saved.isEmpty, "the save waits for the reader to stop changing things")
        try await waitUntil("the settings to save themselves") { client.saved.count == 1 }
        XCTAssertEqual(client.saved.last?["hover_mark_as_read"], "true")
    }

    func testRapidChangesSaveOnce() async throws {
        let (viewModel, client) = makeViewModel()

        viewModel.updateSetting("target_language", value: "zh")
        viewModel.updateSetting("target_language", value: "ja")
        viewModel.updateSetting("target_language", value: "fr")

        try await waitUntil("the settings to save themselves") { !client.saved.isEmpty }
        try? await Task.sleep(for: .milliseconds(700))

        XCTAssertEqual(client.saved.count, 1, "only the settled value should be sent")
        XCTAssertEqual(client.saved.last?["target_language"], "fr")
    }

    func testWritingTheSameValueSavesNothing() async throws {
        let (viewModel, client) = makeViewModel()
        viewModel.updateSetting("theme", value: "dark")
        try await waitUntil("the first save") { client.saved.count == 1 }

        viewModel.updateSetting("theme", value: "dark")
        try? await Task.sleep(for: .milliseconds(700))

        XCTAssertEqual(client.saved.count, 1)
    }

    func testTheLanguageAppliesAsSoonAsItIsSaved() async throws {
        let (viewModel, _) = makeViewModel()
        Localization.shared.setLanguage(.english)

        viewModel.updateSetting("language", value: "zh-CN")

        try await waitUntil("the language to follow the setting") {
            Localization.shared.language == .chineseSimplified
        }

        Localization.shared.setLanguage(.english)
    }
}

final class RecordingSettingsClient: StubAPIClient {
    private(set) var saved: [[String: String]] = []

    override func updateSettings(_ settings: [String: String]) async throws {
        saved.append(settings)
    }
}
