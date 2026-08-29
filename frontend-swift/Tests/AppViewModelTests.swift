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
        try await waitUntil("the newer response to replace the older one") {
            viewModel.articles.map(\.id) == [2]
        }

        XCTAssertEqual(viewModel.articles.map(\.id), [2])
    }

    func testReadActionCanMarkArticleUnread() async throws {
        let client = DelayedAPIClient()
        client.defaultArticles = [Self.article(id: 1, feedID: 1, isRead: true)]
        let viewModel = AppViewModel(api: client, autoLoad: false)

        viewModel.reloadArticles()
        try await waitUntil("the articles to load") { !viewModel.articles.isEmpty }
        let article = try XCTUnwrap(viewModel.articles.first)
        viewModel.setArticleRead(article, read: false)

        XCTAssertEqual(viewModel.articles.first?.isRead, false)
        try await waitUntil("the server to be told") { client.lastReadMutation != nil }
        XCTAssertEqual(client.lastReadMutation, .init(id: 1, read: false))
    }

    func testReloadKeepsCurrentArticlesUntilReplacementArrives() async throws {
        let client = DelayedAPIClient()
        let viewModel = AppViewModel(api: client, autoLoad: false)

        viewModel.reloadArticles()
        try await waitUntil("the first response to arrive") { !viewModel.articles.isEmpty }
        XCTAssertEqual(viewModel.articles.map(\.id), [1])

        client.defaultArticles = [Self.article(id: 3, feedID: 1)]
        viewModel.reloadArticles()

        XCTAssertEqual(viewModel.articles.map(\.id), [1])
        try await waitUntil("the replacement to arrive") { viewModel.articles.map(\.id) == [3] }
        XCTAssertEqual(viewModel.articles.map(\.id), [3])
    }

    func testSelectingAndMarkingArticleReadKeepsItInTheList() async throws {
        let client = DelayedAPIClient()
        client.defaultArticles = [Self.article(id: 1, feedID: 1)]
        let viewModel = AppViewModel(api: client, autoLoad: false)

        viewModel.reloadArticles()
        try await waitUntil("the articles to load") { !viewModel.articles.isEmpty }
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

final class DelayedAPIClient: StubAPIClient {
    struct ReadMutation: Equatable {
        let id: Int
        let read: Bool
    }

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
    private(set) var categoryMutations: [(id: Int, category: String)] = []
    var categoryUpdateError: Error?
    private(set) var lastArticleQuery: (feedID: Int?, category: String?, filter: String)?
    private(set) var reorderMutations: [(id: Int, category: String, position: Int)] = []
    var unreadCounts = UnreadCounts.empty

    override func fetchFeeds() async throws -> [Feed] { defaultFeeds }
    override func fetchUnreadCounts() async throws -> UnreadCounts { unreadCounts }
    override func addFeed(_ draft: FeedDraft) async throws {}
    override func deleteFeed(id: Int) async throws {}

    override func updateFeedCategory(id: Int, category: String) async throws {
        if let categoryUpdateError { throw categoryUpdateError }
        categoryMutations.append((id, category))
    }

    override func reorderFeed(id: Int, category: String, position: Int) async throws {
        if let categoryUpdateError { throw categoryUpdateError }
        reorderMutations.append((id, category, position))
    }

    override func refreshAllFeeds() async throws {}

    override func fetchArticles(
        feedID: Int?,
        category: String?,
        filter: String,
        page: Int,
        limit: Int
    ) async throws -> [Article] {
        lastArticleQuery = (feedID, category, filter)
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

    override func setArticleRead(id: Int, read: Bool) async throws {
        lastReadMutation = ReadMutation(id: id, read: read)
    }

    override func toggleFavorite(id: Int) async throws {}

    override func fetchArticleContent(id: Int) async throws -> ArticleContent {
        articleContent
    }

    override func updateSettings(_ settings: [String: String]) async throws {}

    override func translateTitle(
        articleID: Int,
        title: String,
        targetLanguage: String
    ) async throws -> TitleTranslationResponse {
        TitleTranslationResponse(translatedTitle: title, limitReached: false)
    }

    override func translateText(_ text: String, targetLanguage: String) async throws -> TextTranslationResponse {
        TextTranslationResponse(translatedText: text, html: text)
    }

    override func summarize(articleID: Int, length: String, content: String?) async throws -> SummaryResult {
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

    override func clearTranslations() async throws {}
    override func clearSummaries() async throws {}

    override func applyRule(_ rule: AutomationRule) async throws -> RuleApplicationResult {
        RuleApplicationResult(success: true, affected: 0)
    }

    override func fetchAIUsage() async throws -> AIUsage {
        AIUsage(usage: 0, limit: 20_000, limitReached: false)
    }

    override func resetAIUsage() async throws {}
}

@MainActor
final class ArticleListStateTests: XCTestCase {
    override func setUp() {
        super.setUp()
        Localization.shared.setLanguage(.english)
    }

    private func makeViewModel(
        articles: [Article],
        defaults: UserDefaults = UserDefaults(suiteName: "ArticleListStateTests")!
    ) -> (AppViewModel, ListAPIClient) {
        defaults.removePersistentDomain(forName: "ArticleListStateTests")
        let client = ListAPIClient()
        client.articles = articles
        let viewModel = AppViewModel(api: client, autoLoad: false, defaults: defaults)
        viewModel.articles = articles
        return (viewModel, client)
    }

    func testNewestFirstOrdersByPublicationDate() {
        let (viewModel, _) = makeViewModel(articles: [
            .stub(id: 1, published: "2026-08-10T08:00:00Z"),
            .stub(id: 2, published: "2026-08-16T08:00:00Z")
        ])

        viewModel.sortOrder = .newestFirst

        XCTAssertEqual(viewModel.displayedArticles.map(\.id), [2, 1])
    }

    func testUnreadFirstKeepsUnreadAtTheTop() {
        let (viewModel, _) = makeViewModel(articles: [
            .stub(id: 1, published: "2026-08-16T08:00:00Z", isRead: true),
            .stub(id: 2, published: "2026-08-10T08:00:00Z", isRead: false)
        ])

        viewModel.sortOrder = .unreadFirst

        XCTAssertEqual(viewModel.displayedArticles.map(\.id), [2, 1])
    }

    func testHidingAnArticleRemovesItWhenHiddenArticlesAreNotShown() async {
        let (viewModel, _) = makeViewModel(articles: [.stub(id: 1, published: "2026-08-16T08:00:00Z")])

        viewModel.toggleHidden(viewModel.articles[0])

        XCTAssertTrue(viewModel.articles.isEmpty)
    }

    func testMarkAboveMarksOnlyTheEarlierRows() async {
        let (viewModel, client) = makeViewModel(articles: [
            .stub(id: 1, published: "2026-08-18T08:00:00Z"),
            .stub(id: 2, published: "2026-08-17T08:00:00Z"),
            .stub(id: 3, published: "2026-08-16T08:00:00Z")
        ])
        client.relativeCount = 1

        await viewModel.markRelative(to: viewModel.articles[1], direction: .above)

        XCTAssertEqual(viewModel.articles.filter(\.isRead).map(\.id), [1])
    }

    func testUnreadOnlyRestrictsTheMultimediaListing() async throws {
        let defaults = UserDefaults(suiteName: "ArticleListStateTests")!
        defaults.removePersistentDomain(forName: "ArticleListStateTests")
        let client = ListAPIClient()
        client.imageArticles = [
            .stub(id: 1, published: "2026-08-18T08:00:00Z", isRead: true),
            .stub(id: 2, published: "2026-08-17T08:00:00Z", isRead: false)
        ]
        let viewModel = AppViewModel(api: client, autoLoad: false, defaults: defaults)
        viewModel.showOnlyUnread = true
        viewModel.selection = .filter(.imageGallery)

        try await waitUntil("the multimedia listing to drop read items") {
            viewModel.articles.map(\.id) == [2]
        }
    }

    func testArticleListTitleFollowsTheSelection() {
        let (viewModel, _) = makeViewModel(articles: [])
        viewModel.feeds = [Feed(id: 7, url: "https://example.com/feed", title: "Example", category: "")]

        viewModel.selection = .feed(7)

        XCTAssertEqual(viewModel.articleListTitle, "Example")
    }
}

final class ListAPIClient: StubAPIClient {
    var articles: [Article] = []
    var imageArticles: [Article] = []
    var relativeCount = 0
    private(set) var markedRelative: [(id: Int, direction: String)] = []

    override func fetchFeeds() async throws -> [Feed] { [] }
    override func fetchArticles(
        feedID: Int?,
        category: String?,
        filter: String,
        page: Int,
        limit: Int
    ) async throws -> [Article] { articles }
    override func fetchImageArticles(page: Int, limit: Int) async throws -> [Article] { imageArticles }
    override func toggleHidden(id: Int) async throws {}
    override func setArticleRead(id: Int, read: Bool) async throws {}
    override func markRelative(id: Int, direction: String) async throws -> Int {
        markedRelative.append((id, direction))
        return relativeCount
    }
}

private extension Article {
    static func stub(id: Int, published: String, isRead: Bool = false) -> Article {
        Article(
            id: id,
            feedID: 1,
            feedTitle: "Feed",
            title: "Article \(id)",
            url: "https://example.com/\(id)",
            publishedAt: published,
            isRead: isRead
        )
    }
}
