import XCTest
@testable import MrRSS

@MainActor
final class AppViewModelTests: XCTestCase {
    func testChangingSelectionDiscardsPreviousResponse() async throws {
        let client = DelayedAPIClient()
        let viewModel = AppViewModel(api: client, autoLoad: false)

        viewModel.reloadArticles()
        try await Task.sleep(for: .milliseconds(20))
        viewModel.selection = .feed(2)
        try await Task.sleep(for: .milliseconds(300))

        XCTAssertEqual(viewModel.articles.map(\.id), [2])
    }

    func testReadActionCanMarkArticleUnread() async throws {
        let client = DelayedAPIClient()
        client.defaultArticles = [Self.article(id: 1, feedID: 1, isRead: true)]
        let viewModel = AppViewModel(api: client, autoLoad: false)

        viewModel.reloadArticles()
        try await Task.sleep(for: .milliseconds(220))
        let article = try XCTUnwrap(viewModel.articles.first)
        viewModel.setArticleRead(article, read: false)

        XCTAssertEqual(viewModel.articles.first?.isRead, false)
        try await Task.sleep(for: .milliseconds(20))
        XCTAssertEqual(client.lastReadMutation, .init(id: 1, read: false))
    }

    func testReloadKeepsCurrentArticlesUntilReplacementArrives() async throws {
        let client = DelayedAPIClient()
        let viewModel = AppViewModel(api: client, autoLoad: false)

        viewModel.reloadArticles()
        try await Task.sleep(for: .milliseconds(220))
        XCTAssertEqual(viewModel.articles.map(\.id), [1])

        client.defaultArticles = [Self.article(id: 3, feedID: 1)]
        viewModel.reloadArticles()

        XCTAssertEqual(viewModel.articles.map(\.id), [1])
        try await Task.sleep(for: .milliseconds(220))
        XCTAssertEqual(viewModel.articles.map(\.id), [3])
    }

    func testSelectingAndMarkingArticleReadKeepsItInTheList() async throws {
        let client = DelayedAPIClient()
        client.defaultArticles = [Self.article(id: 1, feedID: 1)]
        let viewModel = AppViewModel(api: client, autoLoad: false)

        viewModel.reloadArticles()
        try await Task.sleep(for: .milliseconds(220))
        let article = try XCTUnwrap(viewModel.articles.first)

        viewModel.selectArticle(article)
        viewModel.setArticleRead(article, read: true)

        XCTAssertEqual(viewModel.selectedArticleID, 1)
        XCTAssertEqual(viewModel.articles.map(\.id), [1])
        XCTAssertEqual(viewModel.articles.first?.isRead, true)
    }

    private static func article(id: Int, feedID: Int, isRead: Bool = false) -> Article {
        Article(
            id: id,
            feedID: feedID,
            feedTitle: "Feed \(feedID)",
            title: "Article \(id)",
            url: "https://example.com/\(id)",
            publishedAt: "2026-08-16T08:00:00Z",
            isRead: isRead
        )
    }
}

final class DelayedAPIClient: APIClient {
    struct ReadMutation: Equatable {
        let id: Int
        let read: Bool
    }

    let baseURL = URL(string: "http://127.0.0.1:1234/api")!
    var defaultArticles = [
        Article(
            id: 1,
            feedID: 1,
            feedTitle: "Feed 1",
            title: "Old article",
            url: "https://example.com/1",
            publishedAt: "2026-08-16T08:00:00Z"
        )
    ]
    private(set) var lastReadMutation: ReadMutation?
    var articleContent = ArticleContent(content: "", feedURL: nil)
    var defaultFeeds: [Feed] = []

    func checkConnection() async throws {}
    func fetchFeeds() async throws -> [Feed] { defaultFeeds }
    func fetchUnreadCounts() async throws -> UnreadCounts { .empty }
    func addFeed(url: String, title: String, category: String) async throws {}
    func deleteFeed(id: Int) async throws {}
    func refreshAllFeeds() async throws {}
    func fetchRefreshProgress() async throws -> RefreshProgress {
        RefreshProgress(isRunning: false)
    }

    func fetchArticles(
        feedID: Int?,
        category: String?,
        filter: String,
        page: Int,
        limit: Int
    ) async throws -> [Article] {
        if feedID == 2 {
            try await Task.sleep(for: .milliseconds(10))
            return [
                Article(
                    id: 2,
                    feedID: 2,
                    feedTitle: "Feed 2",
                    title: "Current article",
                    url: "https://example.com/2",
                    publishedAt: "2026-08-16T08:00:00Z"
                )
            ]
        }

        try? await Task.sleep(for: .milliseconds(180))
        return defaultArticles
    }

    func setArticleRead(id: Int, read: Bool) async throws {
        lastReadMutation = ReadMutation(id: id, read: read)
    }

    func toggleFavorite(id: Int) async throws {}

    func fetchArticleContent(id: Int) async throws -> ArticleContent {
        articleContent
    }

    func fetchSettings() async throws -> [String: String] { [:] }
    func updateSettings(_ settings: [String: String]) async throws {}
    func translateTitle(
        articleID: Int,
        title: String,
        targetLanguage: String
    ) async throws -> TitleTranslationResponse {
        TitleTranslationResponse(translatedTitle: title, limitReached: false)
    }
    func translateText(_ text: String, targetLanguage: String) async throws -> TextTranslationResponse {
        TextTranslationResponse(translatedText: text, html: text)
    }
    func summarize(articleID: Int, length: String, content: String?) async throws -> SummaryResult {
        SummaryResult(
            summary: "Summary",
            html: nil,
            sentenceCount: 1,
            isTooShort: false,
            limitReached: false,
            usedFallback: false,
            thinking: nil,
            error: nil,
            cached: false
        )
    }
    func clearTranslations() async throws {}
    func clearSummaries() async throws {}
    func applyRule(_ rule: AutomationRule) async throws -> RuleApplicationResult {
        RuleApplicationResult(success: true, affected: 0)
    }
    func fetchAIUsage() async throws -> AIUsage {
        AIUsage(usage: 0, limit: 20_000, limitReached: false)
    }
    func resetAIUsage() async throws {}
}
