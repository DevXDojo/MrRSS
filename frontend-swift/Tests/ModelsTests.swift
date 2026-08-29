import XCTest
@testable import MrRSS

final class ModelsTests: XCTestCase {
    func testFeedDecodesBackendResponse() throws {
        let data = Data(#"""
        {
            "id": 7,
            "title": "Example",
            "url": "https://example.com/feed.xml",
            "category": "Technology",
            "image_url": "",
            "last_updated": "2026-08-16T08:00:00Z",
            "description": "Additional backend fields are ignored"
        }
        """#.utf8)

        let feed = try JSONDecoder().decode(Feed.self, from: data)

        XCTAssertEqual(feed.id, 7)
        XCTAssertEqual(feed.lastUpdated, "2026-08-16T08:00:00Z")
        XCTAssertNil(feed.iconURL)
    }

    func testFeedUsesSafeDefaultsForOlderResponses() throws {
        let data = Data(#"{"id":1,"url":"https://example.com/rss","title":"Example"}"#.utf8)

        let feed = try JSONDecoder().decode(Feed.self, from: data)

        XCTAssertEqual(feed.category, "")
        XCTAssertEqual(feed.lastUpdated, "")
    }

    func testUnreadCountsDecodeIntegerFeedIdentifiers() throws {
        let data = Data(#"{"total":5,"feed_counts":{"1":2,"42":3}}"#.utf8)

        let counts = try JSONDecoder().decode(UnreadCounts.self, from: data)

        XCTAssertEqual(counts, UnreadCounts(total: 5, feedCounts: [1: 2, 42: 3]))
    }
}

final class ArticleDateFormatterTests: XCTestCase {
    override func setUp() {
        super.setUp()
        Localization.shared.setLanguage(.english)
    }

    func testTimestampsWithAndWithoutFractionalSecondsBothParse() {
        XCTAssertNotNil(ArticleDateFormatter.date(from: "2026-08-16T08:00:00Z"))
        XCTAssertNotNil(ArticleDateFormatter.date(from: "2026-08-16T08:00:00.123Z"))
        XCTAssertNil(ArticleDateFormatter.date(from: "not a date"))
    }

    func testRelativeDescriptionsUseTheTranslatedWording() {
        let now = ArticleDateFormatter.date(from: "2026-08-16T12:00:00Z")!

        XCTAssertEqual(
            ArticleDateFormatter.relativeDescription(for: "2026-08-16T11:59:30Z", now: now),
            "Just now"
        )
        XCTAssertEqual(
            ArticleDateFormatter.relativeDescription(for: "2026-08-16T11:30:00Z", now: now),
            "30 minutes ago"
        )
        XCTAssertEqual(
            ArticleDateFormatter.relativeDescription(for: "2026-08-16T09:00:00Z", now: now),
            "3 hours ago"
        )
        XCTAssertEqual(
            ArticleDateFormatter.relativeDescription(for: "2026-08-14T12:00:00Z", now: now),
            "2 days ago"
        )
    }

    func testFullDescriptionsFollowTheChosenLanguage() {
        Localization.shared.setLanguage(.chineseSimplified)
        let chinese = ArticleDateFormatter.fullDescription(for: "2026-08-16T08:00:00Z")

        Localization.shared.setLanguage(.english)
        let english = ArticleDateFormatter.fullDescription(for: "2026-08-16T08:00:00Z")

        XCTAssertNotEqual(chinese, english, "the date should be written in the chosen language")
        XCTAssertTrue(english.contains("2026"))
    }

    func testTranslatedTitlesAreUsedOnlyWhenAskedFor() {
        let article = Article(
            id: 1,
            feedID: 1,
            title: "Original",
            url: "https://example.com/1",
            publishedAt: "2026-08-16T08:00:00Z",
            translatedTitle: "翻译"
        )

        XCTAssertEqual(article.displayTitle(preferTranslation: true), "翻译")
        XCTAssertEqual(article.displayTitle(preferTranslation: false), "Original")
    }
}
