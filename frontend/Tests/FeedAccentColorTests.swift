import SwiftUI
import XCTest
@testable import MrRSS

final class FeedAccentColorTests: XCTestCase {
    func testThePaletteOffersMoreThanTheThreePrimaries() {
        XCTAssertGreaterThanOrEqual(FeedAccentColor.palette.count, 12)
        XCTAssertEqual(
            Set(FeedAccentColor.palette.map { String(describing: $0) }).count,
            FeedAccentColor.palette.count,
            "Every entry in the palette should be a distinct colour."
        )
    }

    func testAFeedKeepsTheSameColourAcrossLookups() {
        let feed = Feed(id: 7, url: "https://example.com/feed", title: "Example", category: "")
        let first = FeedAccentColor.color(for: feed)

        // The colour is derived from the address, so renaming or renumbering a
        // feed does not repaint it.
        let renamed = Feed(id: 42, url: "https://example.com/feed", title: "Renamed", category: "Tech")

        XCTAssertEqual(first, FeedAccentColor.color(for: feed))
        XCTAssertEqual(first, FeedAccentColor.color(for: renamed))
    }

    func testTheHashIsStableAcrossProcessesAndStaysInRange() {
        // A fixed expectation catches any switch to Swift's per-process hashing.
        XCTAssertEqual(FeedAccentColor.index(for: "https://example.com/feed"), 6)

        for identity in (0..<200).map({ "https://example.com/\($0)" }) {
            let index = FeedAccentColor.index(for: identity)
            XCTAssertTrue((0..<FeedAccentColor.palette.count).contains(index))
        }
    }

    func testDifferentFeedsSpreadAcrossThePalette() {
        let identities = (0..<60).map { "https://feed\($0).example.com/rss" }
        let used = Set(identities.map(FeedAccentColor.index(for:)))

        XCTAssertGreaterThanOrEqual(used.count, 10)
    }

    func testAFeedWithoutAnAddressFallsBackToItsTitle() {
        let feed = Feed(id: 1, url: "", title: "Local", category: "")
        XCTAssertEqual(FeedAccentColor.color(for: feed), FeedAccentColor.palette[FeedAccentColor.index(for: "Local")])
    }
}
