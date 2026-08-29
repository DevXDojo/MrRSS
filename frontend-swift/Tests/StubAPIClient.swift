import Foundation
@testable import MrRSS

/// A client that fails every call. Tests subclass it and override only the
/// calls the case under test is supposed to make, so an unexpected request
/// shows up as a clear failure rather than as silent empty data.
class StubAPIClient: APIClient {
    var baseURL = URL(string: "http://127.0.0.1:1234/api")!

    private func unimplemented(_ name: String = #function) -> Error {
        APIError.notStubbed(name)
    }

    // Connection
    func checkConnection() async throws {}
    func fetchVersion() async throws -> String { throw unimplemented() }

    // Feeds
    func fetchFeeds() async throws -> [Feed] { throw unimplemented() }
    func addFeed(_ draft: FeedDraft) async throws { throw unimplemented() }
    func updateFeed(_ draft: FeedDraft) async throws { throw unimplemented() }
    func deleteFeed(id: Int) async throws { throw unimplemented() }
    func updateFeedCategory(id: Int, category: String) async throws { throw unimplemented() }
    func reorderFeed(id: Int, category: String, position: Int) async throws { throw unimplemented() }
    func refreshAllFeeds() async throws { throw unimplemented() }
    func refreshFeed(id: Int) async throws { throw unimplemented() }
    func fetchRefreshProgress() async throws -> RefreshProgress { RefreshProgress(isRunning: false) }
    func testIMAPConnection(_ draft: FeedDraft) async throws -> String { throw unimplemented() }
    func fetchFeedTags(feedID: Int) async throws -> [Int] { [] }

    // Discovery
    func startDiscovery(feedID: Int) async throws { throw unimplemented() }
    func fetchDiscoveryProgress() async throws -> DiscoveryState { throw unimplemented() }
    func clearDiscovery() async throws { throw unimplemented() }
    func startDiscoverAll() async throws { throw unimplemented() }
    func fetchDiscoverAllProgress() async throws -> DiscoveryState { throw unimplemented() }
    func clearDiscoverAll() async throws { throw unimplemented() }

    // RSSHub
    func testRSSHubConnection() async throws -> String { throw unimplemented() }
    func transformRSSHubURL(_ url: String) async throws -> String { throw unimplemented() }

    // Articles
    func fetchArticles(
        feedID: Int?,
        category: String?,
        filter: String,
        page: Int,
        limit: Int
    ) async throws -> [Article] { throw unimplemented() }
    func fetchImageArticles(page: Int, limit: Int) async throws -> [Article] { throw unimplemented() }
    func filterArticles(
        conditions: [FilterCondition],
        page: Int,
        limit: Int
    ) async throws -> FilteredArticles { throw unimplemented() }
    func setArticleRead(id: Int, read: Bool) async throws { throw unimplemented() }
    func toggleFavorite(id: Int) async throws { throw unimplemented() }
    func toggleReadLater(id: Int) async throws { throw unimplemented() }
    func toggleHidden(id: Int) async throws { throw unimplemented() }
    func markRelative(
        id: Int,
        direction: String,
        feedID: Int?,
        category: String?
    ) async throws -> Int { throw unimplemented() }
    func markAllRead(feedID: Int?, category: String?) async throws { throw unimplemented() }
    func clearReadLater() async throws { throw unimplemented() }
    func fetchArticleContent(id: Int) async throws -> ArticleContent { throw unimplemented() }
    func reloadArticleContent(id: Int) async throws -> ArticleContent { throw unimplemented() }
    func fetchFullArticle(id: Int) async throws -> ArticleContent { throw unimplemented() }
    func extractImages(id: Int) async throws -> ArticleImages { throw unimplemented() }
    func fetchUnreadCounts() async throws -> UnreadCounts { .empty }
    func fetchFilterCounts() async throws -> FilterCounts { .empty }

    // Export
    func exportArticle(id: Int, destination: ArticleExportDestination) async throws -> String {
        throw unimplemented()
    }

    // Translation and summaries
    func translateTitle(
        articleID: Int,
        title: String,
        targetLanguage: String
    ) async throws -> TitleTranslationResponse { throw unimplemented() }
    func translateText(_ text: String, targetLanguage: String) async throws -> TextTranslationResponse {
        throw unimplemented()
    }
    func summarize(articleID: Int, length: String, content: String?) async throws -> SummaryResult {
        throw unimplemented()
    }
    func clearTranslations() async throws { throw unimplemented() }
    func clearSummaries() async throws { throw unimplemented() }

    // Tags
    func fetchTags() async throws -> [Tag] { [] }
    func createTag(name: String, color: String) async throws -> Tag { throw unimplemented() }
    func updateTag(_ tag: Tag) async throws { throw unimplemented() }
    func deleteTag(id: Int) async throws { throw unimplemented() }
    func reorderTag(id: Int, newPosition: Int) async throws { throw unimplemented() }

    // Saved filters
    func fetchSavedFilters() async throws -> [SavedFilter] { [] }
    func createSavedFilter(name: String, conditions: [FilterCondition]) async throws -> SavedFilter {
        throw unimplemented()
    }
    func updateSavedFilter(_ filter: SavedFilter) async throws { throw unimplemented() }
    func deleteSavedFilter(id: Int) async throws { throw unimplemented() }
    func reorderSavedFilter(id: Int, newPosition: Int) async throws { throw unimplemented() }

    // Rules
    func applyRule(_ rule: AutomationRule) async throws -> RuleApplicationResult { throw unimplemented() }

    // Settings
    func fetchSettings() async throws -> [String: String] { [:] }
    func updateSettings(_ settings: [String: String]) async throws { throw unimplemented() }

    // AI
    func fetchAIUsage() async throws -> AIUsage { AIUsage(usage: 0, limit: 0, limitReached: false) }
    func resetAIUsage() async throws { throw unimplemented() }
    func fetchAIProfiles() async throws -> [AIProfile] { [] }
    func saveAIProfile(_ profile: AIProfile) async throws -> AIProfile { throw unimplemented() }
    func deleteAIProfile(id: Int) async throws { throw unimplemented() }
    func setDefaultAIProfile(id: Int) async throws { throw unimplemented() }
    func testAIProfiles() async throws -> [AIProfileTestResult] { throw unimplemented() }
    func aiSearch(query: String) async throws -> AISearchResponse { throw unimplemented() }
    func sendChatMessage(_ request: ChatRequest) async throws -> ChatResponse { throw unimplemented() }
    func fetchChatSessions(articleID: Int) async throws -> [ChatSession] { [] }
    func createChatSession(articleID: Int, title: String) async throws -> ChatSession {
        throw unimplemented()
    }
    func fetchChatMessages(sessionID: Int) async throws -> [ChatMessage] { [] }
    func deleteChatSession(id: Int) async throws { throw unimplemented() }
    func deleteAllChatSessions() async throws { throw unimplemented() }

    // System
    func fetchStatistics(period: String, offset: Int) async throws -> StatisticsSummary {
        throw unimplemented()
    }
    func fetchAllTimeStatistics() async throws -> [String: Int] { [:] }
    func fetchContentCacheInfo() async throws -> ContentCacheInfo { ContentCacheInfo(cachedArticles: 0) }
    func cleanupArticles() async throws { throw unimplemented() }
    func cleanupContentCache() async throws { throw unimplemented() }
    func fetchMediaCacheInfo() async throws -> MediaCacheInfo { MediaCacheInfo(cacheSizeMB: 0) }
    func cleanupMediaCache() async throws { throw unimplemented() }
    func checkForUpdates() async throws -> UpdateInfo { throw unimplemented() }
    func fetchFreshRSSStatus() async throws -> FreshRSSStatus {
        FreshRSSStatus(pendingChanges: 0, failedItems: 0, lastSyncTime: nil)
    }
    func syncFreshRSS() async throws { throw unimplemented() }
    func syncFreshRSSFeed(id: Int) async throws { throw unimplemented() }
    func exportOPML() async throws -> Data { throw unimplemented() }
    func importOPML(data: Data, filename: String) async throws { throw unimplemented() }
    func openInBrowser(url: String) async throws -> String? { nil }
    func fetchWindowState() async throws -> WindowState { throw unimplemented() }
    func saveWindowState(_ state: WindowState) async throws {}
    func fetchScripts() async throws -> ScriptList { throw unimplemented() }
    func uploadCustomCSS(data: Data, filename: String) async throws { throw unimplemented() }
    func deleteCustomCSS() async throws { throw unimplemented() }
    func fetchCustomCSS() async throws -> String { throw unimplemented() }
}
