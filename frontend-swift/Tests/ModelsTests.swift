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
