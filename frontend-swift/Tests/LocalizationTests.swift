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
