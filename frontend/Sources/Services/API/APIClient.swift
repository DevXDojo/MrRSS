import Foundation

/// Everything the interface needs from the backend. `APIService` is the live
/// implementation; tests use `StubAPIClient`, which fails any call a test has
/// not explicitly prepared.
protocol APIClient: AnyObject {
    var baseURL: URL { get }

    // Connection
    func checkConnection() async throws
    func fetchVersion() async throws -> String

    // Feeds
    func fetchFeeds() async throws -> [Feed]
    func addFeed(_ draft: FeedDraft) async throws
    func updateFeed(_ draft: FeedDraft) async throws
    func deleteFeed(id: Int) async throws
    func updateFeedCategory(id: Int, category: String) async throws
    func reorderFeed(id: Int, category: String, position: Int) async throws
    func refreshAllFeeds() async throws
    func refreshFeed(id: Int) async throws
    func fetchRefreshProgress() async throws -> RefreshProgress
    func testIMAPConnection(_ draft: FeedDraft) async throws -> String

    // Discovery
    func startDiscovery(feedID: Int) async throws
    func fetchDiscoveryProgress() async throws -> DiscoveryState
    func clearDiscovery() async throws
    func startDiscoverAll() async throws
    func fetchDiscoverAllProgress() async throws -> DiscoveryState
    func clearDiscoverAll() async throws

    // RSSHub
    func testRSSHubConnection() async throws -> String
    func transformRSSHubURL(_ url: String) async throws -> String

    // Articles
    func fetchArticles(
        feedID: Int?,
        category: String?,
        filter: String,
        page: Int,
        limit: Int
    ) async throws -> [Article]
    func fetchImageArticles(page: Int, limit: Int) async throws -> [Article]
    func filterArticles(conditions: [FilterCondition], page: Int, limit: Int) async throws -> FilteredArticles
    func setArticleRead(id: Int, read: Bool) async throws
    func toggleFavorite(id: Int) async throws
    func toggleReadLater(id: Int) async throws
    func toggleHidden(id: Int) async throws
    func markRelative(id: Int, direction: String, feedID: Int?, category: String?) async throws -> Int
    func markAllRead(feedID: Int?, category: String?) async throws
    func clearReadLater() async throws
    func fetchArticleContent(id: Int) async throws -> ArticleContent
    func reloadArticleContent(id: Int) async throws -> ArticleContent
    func fetchFullArticle(id: Int) async throws -> ArticleContent
    func extractImages(id: Int) async throws -> ArticleImages
    func fetchUnreadCounts() async throws -> UnreadCounts
    func fetchFilterCounts() async throws -> FilterCounts

    // Article export
    func exportArticle(id: Int, destination: ArticleExportDestination) async throws -> String

    // Translation and summaries
    func translateTitle(articleID: Int, title: String, targetLanguage: String) async throws -> TitleTranslationResponse
    func translateText(_ text: String, targetLanguage: String) async throws -> TextTranslationResponse
    func summarize(articleID: Int, length: String, content: String?) async throws -> SummaryResult
    func clearTranslations() async throws
    func clearSummaries() async throws

    // Tags
    func fetchTags() async throws -> [Tag]
    func createTag(name: String, color: String) async throws -> Tag
    func updateTag(_ tag: Tag) async throws
    func deleteTag(id: Int) async throws
    func reorderTag(id: Int, newPosition: Int) async throws

    // Saved filters
    func fetchSavedFilters() async throws -> [SavedFilter]
    func createSavedFilter(name: String, conditions: [FilterCondition]) async throws -> SavedFilter
    func updateSavedFilter(_ filter: SavedFilter) async throws
    func deleteSavedFilter(id: Int) async throws
    func reorderSavedFilters(_ filters: [SavedFilter]) async throws

    // Rules
    func applyRule(_ rule: AutomationRule) async throws -> RuleApplicationResult

    // Settings
    func fetchSettings() async throws -> [String: String]
    func updateSettings(_ settings: [String: String]) async throws

    // AI
    func fetchAIUsage() async throws -> AIUsage
    func resetAIUsage() async throws
    func fetchAIProfiles() async throws -> [AIProfile]
    func saveAIProfile(_ profile: AIProfile) async throws -> AIProfile
    func deleteAIProfile(id: Int) async throws
    func setDefaultAIProfile(id: Int) async throws
    func testAIProfiles() async throws -> [AIProfileTestResult]
    func aiSearch(query: String) async throws -> AISearchResponse
    func sendChatMessage(_ request: ChatRequest) async throws -> ChatResponse
    func fetchChatSessions(articleID: Int) async throws -> [ChatSession]
    func createChatSession(articleID: Int, title: String) async throws -> ChatSession
    func fetchChatMessages(sessionID: Int) async throws -> [ChatMessage]
    func deleteChatSession(id: Int) async throws
    func deleteAllChatSessions() async throws

    // Maintenance and system
    func fetchStatistics(period: String, offset: Int) async throws -> StatisticsSummary
    func fetchAllTimeStatistics() async throws -> [String: Int]
    func fetchContentCacheInfo() async throws -> ContentCacheInfo
    func cleanupArticles() async throws
    func cleanupContentCache() async throws
    func fetchMediaCacheInfo() async throws -> MediaCacheInfo
    func cleanupMediaCache() async throws
    func checkForUpdates() async throws -> UpdateInfo
    func fetchFreshRSSStatus() async throws -> FreshRSSStatus
    func syncFreshRSS() async throws
    func syncFreshRSSFeed(id: Int) async throws
    func exportOPML() async throws -> Data
    func importOPML(data: Data, filename: String) async throws
    func openInBrowser(url: String) async throws -> String?

    // Window and scripts
    func fetchWindowState() async throws -> WindowState
    func saveWindowState(_ state: WindowState) async throws
    func fetchScripts() async throws -> ScriptList
    func uploadCustomCSS(data: Data, filename: String) async throws
    func deleteCustomCSS() async throws
    func fetchCustomCSS() async throws -> String
}

/// Where an article can be sent from the reading view.
enum ArticleExportDestination: String, CaseIterable, Identifiable {
    case obsidian
    case notion
    case zotero

    var id: String { rawValue }

    var endpoint: String { "articles/export/\(rawValue)" }

    var localizedTitle: String {
        switch self {
        case .obsidian: t("setting.plugins.obsidian.exportTo")
        case .notion: t("setting.plugins.notion.exportTo")
        case .zotero: t("setting.plugins.zotero.exportTo")
        }
    }

    var icon: String {
        switch self {
        case .obsidian: "square.stack.3d.up"
        case .notion: "note.text"
        case .zotero: "books.vertical"
        }
    }

    var localizedSuccess: String {
        switch self {
        case .obsidian: t("setting.plugins.obsidian.exported")
        case .notion: t("setting.plugins.notion.exported")
        case .zotero: t("setting.plugins.zotero.exported")
        }
    }

    var localizedFailure: String {
        switch self {
        case .obsidian: t("setting.plugins.obsidian.exportFailed")
        case .notion: t("setting.plugins.notion.exportFailed")
        case .zotero: t("setting.plugins.zotero.exportFailed")
        }
    }

    var localizedProgress: String {
        switch self {
        case .obsidian: t("setting.plugins.obsidian.exporting")
        case .notion: t("setting.plugins.notion.exporting")
        case .zotero: t("setting.plugins.zotero.exporting")
        }
    }
}

/// The payload the add and edit feed forms send.
struct FeedDraft: Equatable {
    var id: Int?
    var url: String = ""
    var title: String = ""
    var category: String = ""
    var scriptPath: String = ""
    var hideFromTimeline: Bool = false
    var proxyURL: String = ""
    var proxyEnabled: Bool = false
    var refreshInterval: Int = 0
    var isImageMode: Bool = false
    var type: String = ""
    var xPathItem: String = ""
    var xPathItemTitle: String = ""
    var xPathItemContent: String = ""
    var xPathItemURI: String = ""
    var xPathItemAuthor: String = ""
    var xPathItemTimestamp: String = ""
    var xPathItemTimeFormat: String = ""
    var xPathItemThumbnail: String = ""
    var xPathItemCategories: String = ""
    var xPathItemUID: String = ""
    var articleViewMode: String = "global"
    var autoExpandContent: String = "global"
    var emailAddress: String = ""
    var emailIMAPServer: String = ""
    var emailIMAPPort: Int = 993
    var emailUsername: String = ""
    var emailPassword: String = ""
    var emailFolder: String = "INBOX"
    var tags: [Int] = []

    init(id: Int? = nil, url: String = "", title: String = "", category: String = "") {
        self.id = id
        self.url = url
        self.title = title
        self.category = category
    }

    /// Builds a draft that round-trips an existing feed through the edit form.
    init(feed: Feed, tags: [Int] = []) {
        id = feed.id
        url = feed.url
        title = feed.title
        category = feed.category
        scriptPath = feed.scriptPath
        hideFromTimeline = feed.hideFromTimeline
        proxyURL = feed.proxyURL
        proxyEnabled = feed.proxyEnabled
        refreshInterval = feed.refreshInterval
        isImageMode = feed.isImageMode
        type = feed.type
        xPathItem = feed.xPathItem
        xPathItemTitle = feed.xPathItemTitle
        xPathItemContent = feed.xPathItemContent
        xPathItemURI = feed.xPathItemURI
        xPathItemAuthor = feed.xPathItemAuthor
        xPathItemTimestamp = feed.xPathItemTimestamp
        xPathItemTimeFormat = feed.xPathItemTimeFormat
        xPathItemThumbnail = feed.xPathItemThumbnail
        xPathItemCategories = feed.xPathItemCategories
        xPathItemUID = feed.xPathItemUID
        articleViewMode = feed.articleViewMode
        autoExpandContent = feed.autoExpandContent
        emailAddress = feed.emailAddress
        emailIMAPServer = feed.emailIMAPServer
        emailIMAPPort = feed.emailIMAPPort
        emailUsername = feed.emailUsername
        emailPassword = feed.emailPassword
        emailFolder = feed.emailFolder
        self.tags = tags
    }

    /// The JSON body shared by `/feeds/add` and `/feeds/update`.
    var jsonBody: [String: Any] {
        var body: [String: Any] = [
            "url": url,
            "title": title,
            "category": category,
            "script_path": scriptPath,
            "hide_from_timeline": hideFromTimeline,
            "proxy_url": proxyURL,
            "proxy_enabled": proxyEnabled,
            "refresh_interval": refreshInterval,
            "is_image_mode": isImageMode,
            "type": type,
            "xpath_item": xPathItem,
            "xpath_item_title": xPathItemTitle,
            "xpath_item_content": xPathItemContent,
            "xpath_item_uri": xPathItemURI,
            "xpath_item_author": xPathItemAuthor,
            "xpath_item_timestamp": xPathItemTimestamp,
            "xpath_item_time_format": xPathItemTimeFormat,
            "xpath_item_thumbnail": xPathItemThumbnail,
            "xpath_item_categories": xPathItemCategories,
            "xpath_item_uid": xPathItemUID,
            "article_view_mode": articleViewMode,
            "auto_expand_content": autoExpandContent,
            "email_address": emailAddress,
            "email_imap_server": emailIMAPServer,
            "email_imap_port": emailIMAPPort,
            "email_username": emailUsername,
            "email_password": emailPassword,
            "email_folder": emailFolder,
            "tags": tags
        ]
        if let id {
            body["id"] = id
        }
        return body
    }
}

/// A page of results from the saved-filter endpoint.
struct FilteredArticles: Codable, Equatable {
    let articles: [Article]
    let total: Int
    let page: Int
    let limit: Int
    let hasMore: Bool

    enum CodingKeys: String, CodingKey {
        case articles, total, page, limit
        case hasMore = "has_more"
    }

    init(articles: [Article], total: Int, page: Int, limit: Int, hasMore: Bool) {
        self.articles = articles
        self.total = total
        self.page = page
        self.limit = limit
        self.hasMore = hasMore
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        articles = try container.decodeIfPresent([Article].self, forKey: .articles) ?? []
        total = try container.decodeIfPresent(Int.self, forKey: .total) ?? 0
        page = try container.decodeIfPresent(Int.self, forKey: .page) ?? 1
        limit = try container.decodeIfPresent(Int.self, forKey: .limit) ?? 0
        hasMore = try container.decodeIfPresent(Bool.self, forKey: .hasMore) ?? false
    }
}

/// One turn of an AI conversation, as the chat endpoint expects it.
struct ChatRequest: Encodable {
    struct Turn: Encodable {
        let role: String
        let content: String
    }

    var messages: [Turn]
    var sessionID: Int?
    var articleID: Int?
    var articleTitle: String?
    var articleURL: String?
    var articleContent: String?
    var isFirstMessage: Bool

    enum CodingKeys: String, CodingKey {
        case messages
        case sessionID = "session_id"
        case articleID = "article_id"
        case articleTitle = "article_title"
        case articleURL = "article_url"
        case articleContent = "article_content"
        case isFirstMessage = "is_first_message"
    }
}
