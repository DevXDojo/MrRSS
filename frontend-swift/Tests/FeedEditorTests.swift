import AppKit
import SwiftUI
import XCTest
@testable import MrRSS

@MainActor
final class FeedEditorLayoutTests: XCTestCase {
    override func setUp() {
        super.setUp()
        Localization.shared.setLanguage(.english)
    }

    func testTheSheetLaysItsFieldsOutAcrossTheFullWidth() async throws {
        let viewModel = AppViewModel(api: DelayedAPIClient(), autoLoad: false)
        let hostingView = NSHostingView(rootView: FeedEditorView(mode: .add, viewModel: viewModel))
        hostingView.frame = NSRect(x: 0, y: 0, width: 620, height: 640)

        let window = NSWindow(
            contentRect: hostingView.frame,
            styleMask: [.titled, .closable],
            backing: .buffered,
            defer: false
        )
        window.contentView = hostingView
        window.makeKeyAndOrderFront(nil)
        hostingView.layoutSubtreeIfNeeded()

        try await waitUntil("the editor to lay its fields out") {
            !descendants(of: NSTextField.self, in: hostingView).filter(\.isEditable).isEmpty
        }

        let fields = descendants(of: NSTextField.self, in: hostingView).filter(\.isEditable)
        XCTAssertGreaterThanOrEqual(fields.count, 2, "the address and the title are always shown")

        // The form is wide enough that no field is squeezed into a narrow column
        // beside its label.
        for field in fields {
            XCTAssertGreaterThan(field.frame.width, 300)
        }
    }

    private func descendants<View: NSView>(of type: View.Type, in root: NSView) -> [View] {
        var found: [View] = []
        if let match = root as? View {
            found.append(match)
        }
        for subview in root.subviews {
            found.append(contentsOf: descendants(of: type, in: subview))
        }
        return found
    }
}

final class FeedDraftTests: XCTestCase {
    func testADraftBuiltFromAFeedKeepsEveryConfiguredField() {
        var feed = Feed(id: 3, url: "https://example.com/feed", title: "Example", category: "Tech")
        feed.hideFromTimeline = true
        feed.refreshInterval = 30
        feed.proxyEnabled = true
        feed.proxyURL = "http://127.0.0.1:7890"
        feed.type = "HTML+XPath"
        feed.xPathItem = "//article"
        feed.emailIMAPServer = "imap.example.com"
        feed.emailIMAPPort = 143

        let draft = FeedDraft(feed: feed, tags: [1, 2])

        XCTAssertEqual(draft.id, 3)
        XCTAssertEqual(draft.category, "Tech")
        XCTAssertTrue(draft.hideFromTimeline)
        XCTAssertEqual(draft.refreshInterval, 30)
        XCTAssertTrue(draft.proxyEnabled)
        XCTAssertEqual(draft.xPathItem, "//article")
        XCTAssertEqual(draft.emailIMAPPort, 143)
        XCTAssertEqual(draft.tags, [1, 2])
    }

    func testTheJSONBodyUsesTheKeysTheBackendReads() throws {
        var draft = FeedDraft(url: "https://example.com/feed", title: "Example", category: "News")
        draft.isImageMode = true
        draft.articleViewMode = "webpage"
        draft.tags = [5]

        let body = draft.jsonBody

        XCTAssertEqual(body["url"] as? String, "https://example.com/feed")
        XCTAssertEqual(body["is_image_mode"] as? Bool, true)
        XCTAssertEqual(body["article_view_mode"] as? String, "webpage")
        XCTAssertEqual(body["tags"] as? [Int], [5])
        XCTAssertNil(body["id"], "a new subscription carries no identifier")
        XCTAssertTrue(JSONSerialization.isValidJSONObject(body))
    }

    func testASavedFilterEncodesItsConditionsAsAJSONString() throws {
        let filter = SavedFilter(
            id: 1,
            name: "Swift",
            conditions: [FilterCondition(field: "article_title", operator: "contains", value: "swift")]
        )

        let data = try JSONEncoder().encode(filter)
        let object = try XCTUnwrap(try JSONSerialization.jsonObject(with: data) as? [String: Any])
        let conditions = try XCTUnwrap(object["conditions"] as? String)

        XCTAssertTrue(conditions.contains("article_title"))

        let roundTripped = try JSONDecoder().decode(SavedFilter.self, from: data)
        XCTAssertEqual(roundTripped.conditions.first?.value, "swift")
    }
}

final class ColorConversionTests: XCTestCase {
    func testHexStringsAreParsedIntoColors() throws {
        XCTAssertNotNil(Color(hex: "#3b82f6"))
        XCTAssertNotNil(Color(hex: "3b82f6"), "the leading hash is optional")
        XCTAssertNil(Color(hex: "#zzzzzz"))
        XCTAssertNil(Color(hex: "#fff"), "only the six digit form is stored")
        XCTAssertNil(Color(hex: ""))
    }

    func testAColorRoundTripsThroughItsHexForm() throws {
        for hex in ["#3b82f6", "#ff0000", "#000000", "#ffffff"] {
            let color = try XCTUnwrap(Color(hex: hex))
            XCTAssertEqual(color.hexString, hex, "\(hex) should survive the round trip")
        }
    }
}

@MainActor
final class BulkFeedSaveTests: XCTestCase {
    func testSavingSeveralFeedsReloadsOnlyOnce() async throws {
        let client = CountingFeedClient()
        let defaults = UserDefaults(suiteName: "BulkFeedSaveTests")!
        defaults.removePersistentDomain(forName: "BulkFeedSaveTests")
        let viewModel = AppViewModel(api: client, autoLoad: false, defaults: defaults)

        for index in 1...3 {
            await viewModel.saveFeed(
                FeedDraft(url: "https://example.com/\(index)"),
                isEditing: false,
                reloading: false
            )
        }

        XCTAssertEqual(client.addedFeeds, 3)
        XCTAssertEqual(client.feedListLoads, 0, "no reload should happen while saving")

        await viewModel.reloadAfterFeedChange()

        // The reload runs off the main task, so wait for it to land.
        try await waitUntil("the single reload to run") { client.feedListLoads == 1 }
    }
}

final class CountingFeedClient: StubAPIClient {
    private(set) var addedFeeds = 0
    private(set) var feedListLoads = 0

    override func addFeed(_ draft: FeedDraft) async throws {
        addedFeeds += 1
    }

    override func fetchFeeds() async throws -> [Feed] {
        feedListLoads += 1
        return []
    }

    override func fetchArticles(
        feedID: Int?,
        category: String?,
        filter: String,
        page: Int,
        limit: Int
    ) async throws -> [Article] { [] }
}
