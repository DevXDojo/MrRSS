import Foundation
import SwiftUI

enum ArticleFilter: String, CaseIterable, Identifiable {
    case all
    case unread
    case favorites
    case readLater

    var id: String { rawValue }

    var title: String {
        switch self {
        case .all: "All Articles"
        case .unread: "Unread"
        case .favorites: "Favorites"
        case .readLater: "Read Later"
        }
    }

    var icon: String {
        switch self {
        case .all: "rectangle.stack"
        case .unread: "circle.fill"
        case .favorites: "star.fill"
        case .readLater: "clock.fill"
        }
    }
}

enum SidebarItem: Hashable {
    case filter(ArticleFilter)
    case feed(Int)
}

enum ConnectionState: Equatable {
    case connecting
    case connected
    case disconnected

    var title: String {
        switch self {
        case .connecting: "Connecting"
        case .connected: "Connected"
        case .disconnected: "Offline"
        }
    }

    var color: Color {
        switch self {
        case .connecting: .orange
        case .connected: .green
        case .disconnected: .red
        }
    }
}

@MainActor
final class AppViewModel: ObservableObject {
    @Published private(set) var feeds: [Feed] = []
    @Published private(set) var articles: [Article] = []
    @Published private(set) var unreadCounts = UnreadCounts.empty
    @Published private(set) var isLoadingArticles = false
    @Published private(set) var isLoadingFeeds = false
    @Published private(set) var isRefreshingSources = false
    @Published private(set) var connectionState: ConnectionState = .connecting
    @Published private(set) var settings: [String: String] = [:]
    @Published private(set) var rules: [AutomationRule] = []
    @Published private(set) var aiUsage: AIUsage?
    @Published private(set) var isLoadingSettings = false
    @Published private(set) var isSavingSettings = false
    @Published var statusMessage: String?
    @Published var errorMessage: String?
    @Published var selectedArticleID: Int?
    @Published var serverURLText: String

    @Published var selection: SidebarItem? = .filter(.all) {
        didSet {
            guard selection != oldValue else { return }
            selectedArticleID = nil
            reloadArticles()
        }
    }

    private var page = 0
    private var hasMore = true
    private let limit: Int
    private var api: APIClient
    private var feedTask: Task<Void, Never>?
    private var sourceRefreshTask: Task<Void, Never>?
    private var feedRequestID = UUID()
    private var articleTask: Task<Void, Never>?
    private var articleRequestID = UUID()

    init(api: APIClient = APIService.shared, limit: Int = 50, autoLoad: Bool = true) {
        self.api = api
        self.limit = limit
        serverURLText = api.baseURL.absoluteString

        if autoLoad {
            refreshAll()
        }
    }

    func refreshAll() {
        refreshFeeds()
        reloadArticles()
        Task { await loadSettings() }
    }

    func start() async {
        connectionState = .connecting
        for attempt in 0..<30 {
            do {
                try await api.checkConnection()
                connectionState = .connected
                refreshAll()
                return
            } catch {
                if attempt == 29 {
                    connectionState = .disconnected
                    errorMessage = error.localizedDescription
                    return
                }
                try? await Task.sleep(for: .milliseconds(200))
            }
        }
    }

    func refreshFeeds() {
        feedTask?.cancel()
        feedRequestID = UUID()
        let requestID = feedRequestID
        isLoadingFeeds = true
        connectionState = .connecting

        feedTask = Task { [weak self] in
            guard let self else { return }
            do {
                async let feedsRequest = api.fetchFeeds()
                async let countsRequest = api.fetchUnreadCounts()
                let (loadedFeeds, loadedCounts) = try await (feedsRequest, countsRequest)
                try Task.checkCancellation()
                guard requestID == feedRequestID else { return }
                feeds = loadedFeeds
                unreadCounts = loadedCounts
                connectionState = .connected
                isLoadingFeeds = false
            } catch is CancellationError {
                guard requestID == feedRequestID else { return }
                isLoadingFeeds = false
            } catch {
                guard requestID == feedRequestID else { return }
                connectionState = .disconnected
                errorMessage = error.localizedDescription
                isLoadingFeeds = false
            }
        }
    }

    func refreshFromSources() {
        guard !isRefreshingSources else { return }
        sourceRefreshTask?.cancel()
        isRefreshingSources = true

        sourceRefreshTask = Task { [weak self] in
            guard let self else { return }
            do {
                try await api.refreshAllFeeds()

                for _ in 0..<180 {
                    try Task.checkCancellation()
                    let progress = try await api.fetchRefreshProgress()
                    if !progress.isRunning { break }
                    try await Task.sleep(for: .seconds(1))
                }

                try Task.checkCancellation()
                isRefreshingSources = false
                refreshAll()
            } catch is CancellationError {
                isRefreshingSources = false
            } catch {
                isRefreshingSources = false
                errorMessage = error.localizedDescription
            }
        }
    }

    func addFeed(url: String, title: String, category: String) async -> Bool {
        guard let parsedURL = URL(string: url),
              let scheme = parsedURL.scheme?.lowercased(),
              ["http", "https"].contains(scheme),
              parsedURL.host != nil else {
            errorMessage = "Enter a valid HTTP or HTTPS feed URL."
            return false
        }

        do {
            try await api.addFeed(
                url: url.trimmingCharacters(in: .whitespacesAndNewlines),
                title: title.trimmingCharacters(in: .whitespacesAndNewlines),
                category: category.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            refreshAll()
            return true
        } catch {
            errorMessage = error.localizedDescription
            return false
        }
    }

    func deleteFeed(_ feed: Feed) async {
        do {
            try await api.deleteFeed(id: feed.id)
            if selection == .feed(feed.id) {
                selection = .filter(.all)
            }
            refreshAll()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func reloadArticles() {
        articleTask?.cancel()
        articleRequestID = UUID()
        page = 0
        hasMore = true
        fetchPage(1, replacing: true)
    }

    func selectArticle(_ article: Article) {
        selectedArticleID = article.id
    }

    func loadMore() {
        guard hasMore, !isLoadingArticles else { return }
        fetchPage(page + 1, replacing: false)
    }

    func setArticleRead(_ article: Article, read: Bool) {
        guard article.isRead != read,
              let index = articles.firstIndex(where: { $0.id == article.id }) else {
            return
        }

        let previousArticle = articles[index]
        articles[index].isRead = read
        if read {
            articles[index].isReadLater = false
        }

        Task { [weak self] in
            guard let self else { return }
            do {
                try await api.setArticleRead(id: article.id, read: read)
                unreadCounts = try await api.fetchUnreadCounts()
            } catch {
                if let currentIndex = articles.firstIndex(where: { $0.id == article.id }),
                   articles[currentIndex].isRead == read {
                    articles[currentIndex] = previousArticle
                }
                errorMessage = error.localizedDescription
            }
        }
    }

    func setArticleRead(id: Int, read: Bool) {
        guard let article = articles.first(where: { $0.id == id }) else { return }
        setArticleRead(article, read: read)
    }

    func toggleFavorite(_ article: Article) {
        guard let index = articles.firstIndex(where: { $0.id == article.id }) else { return }
        let previousValue = articles[index].isFavorite
        articles[index].isFavorite.toggle()

        Task { [weak self] in
            guard let self else { return }
            do {
                try await api.toggleFavorite(id: article.id)
            } catch {
                if let currentIndex = articles.firstIndex(where: { $0.id == article.id }),
                   articles[currentIndex].isFavorite != previousValue {
                    articles[currentIndex].isFavorite = previousValue
                }
                errorMessage = error.localizedDescription
            }
        }
    }

    @discardableResult
    func saveServerAddress() -> Bool {
        guard let url = ServerConfiguration.normalizedURL(from: serverURLText) else {
            errorMessage = "Enter a valid HTTP or HTTPS server address."
            return false
        }

        UserDefaults.standard.set(url.absoluteString, forKey: ServerConfiguration.storageKey)
        api = APIService(baseURL: url)
        serverURLText = url.absoluteString
        selectedArticleID = nil
        errorMessage = nil
        refreshAll()
        return true
    }

    func clearError() {
        errorMessage = nil
    }

    func article(withID id: Int?) -> Article? {
        guard let id else { return nil }
        return articles.first(where: { $0.id == id })
    }

    func fetchArticleContent(id: Int) async throws -> ArticleContent {
        try await api.fetchArticleContent(id: id)
    }

    func loadSettings() async {
        guard !isLoadingSettings else { return }
        isLoadingSettings = true
        do {
            settings = try await api.fetchSettings()
            decodeRules()
            aiUsage = try? await api.fetchAIUsage()
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoadingSettings = false
    }

    func setting(_ key: String, default defaultValue: String = "") -> String {
        settings[key] ?? defaultValue
    }

    func updateSetting(_ key: String, value: String) {
        settings[key] = value
    }

    func boolSetting(_ key: String, default defaultValue: Bool = false) -> Bool {
        guard let value = settings[key] else { return defaultValue }
        return value == "true" || value == "1"
    }

    func updateBoolSetting(_ key: String, value: Bool) {
        settings[key] = value ? "true" : "false"
    }

    @discardableResult
    func saveSettings() async -> Bool {
        guard !isSavingSettings else { return false }
        isSavingSettings = true
        do {
            try await api.updateSettings(settings)
            statusMessage = "Settings saved."
            decodeRules()
            isSavingSettings = false
            return true
        } catch {
            errorMessage = error.localizedDescription
            isSavingSettings = false
            return false
        }
    }

    func saveRules(_ newRules: [AutomationRule]) async -> Bool {
        do {
            let data = try JSONEncoder().encode(newRules)
            guard let json = String(data: data, encoding: .utf8) else { return false }
            rules = newRules
            settings["rules"] = json
            return await saveSettings()
        } catch {
            errorMessage = error.localizedDescription
            return false
        }
    }

    func applyRule(_ rule: AutomationRule) async -> RuleApplicationResult? {
        do {
            let result = try await api.applyRule(rule)
            statusMessage = "Rule applied to \(result.affected) articles."
            refreshAll()
            return result
        } catch {
            errorMessage = error.localizedDescription
            return nil
        }
    }

    func translateTitle(for article: Article) async throws -> String {
        let language = setting("target_language", default: "zh")
        let result = try await api.translateTitle(
            articleID: article.id,
            title: article.title,
            targetLanguage: language
        )
        if let index = articles.firstIndex(where: { $0.id == article.id }) {
            articles[index].translatedTitle = result.translatedTitle
        }
        if result.limitReached {
            statusMessage = "The AI usage limit was reached; a fallback provider was used."
        }
        return result.translatedTitle
    }

    func translateContent(_ content: String) async throws -> TextTranslationResponse {
        try await api.translateText(
            content,
            targetLanguage: setting("target_language", default: "zh")
        )
    }

    func summarize(article: Article, content: String?) async throws -> SummaryResult {
        let result = try await api.summarize(
            articleID: article.id,
            length: setting("summary_length", default: "medium"),
            content: content
        )
        if let index = articles.firstIndex(where: { $0.id == article.id }) {
            articles[index].summary = result.summary
        }
        return result
    }

    func clearGeneratedContent() async {
        do {
            try await api.clearTranslations()
            try await api.clearSummaries()
            statusMessage = "Cached translations and summaries cleared."
            reloadArticles()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func resetAIUsage() async {
        do {
            try await api.resetAIUsage()
            aiUsage = try await api.fetchAIUsage()
            statusMessage = "AI usage reset."
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    var preferredColorScheme: ColorScheme? {
        switch setting("theme", default: "auto") {
        case "light": .light
        case "dark": .dark
        default: nil
        }
    }

    func clearStatusMessage() {
        statusMessage = nil
    }

    private func decodeRules() {
        guard let rawRules = settings["rules"], !rawRules.isEmpty,
              let data = rawRules.data(using: .utf8),
              let decoded = try? JSONDecoder().decode([AutomationRule].self, from: data) else {
            rules = []
            return
        }
        rules = decoded
    }

    private func fetchPage(_ targetPage: Int, replacing: Bool) {
        guard !isLoadingArticles || replacing else { return }

        let requestID = articleRequestID
        let query = articleQuery
        isLoadingArticles = true
        if replacing {
            connectionState = .connecting
        }

        articleTask = Task { [weak self] in
            guard let self else { return }
            do {
                let loadedArticles = try await api.fetchArticles(
                    feedID: query.feedID,
                    category: nil,
                    filter: query.filter,
                    page: targetPage,
                    limit: limit
                )
                try Task.checkCancellation()
                guard requestID == articleRequestID else { return }

                if replacing {
                    articles = loadedArticles
                } else {
                    let existingIDs = Set(articles.map(\.id))
                    articles.append(contentsOf: loadedArticles.filter { !existingIDs.contains($0.id) })
                }
                page = targetPage
                hasMore = loadedArticles.count == limit
                connectionState = .connected
                isLoadingArticles = false
            } catch is CancellationError {
                guard requestID == articleRequestID else { return }
                isLoadingArticles = false
            } catch {
                guard requestID == articleRequestID else { return }
                connectionState = .disconnected
                errorMessage = error.localizedDescription
                isLoadingArticles = false
            }
        }
    }

    private var articleQuery: (feedID: Int?, filter: String) {
        switch selection {
        case .filter(let filter):
            return (nil, filter.rawValue)
        case .feed(let id):
            return (id, "")
        case .none:
            return (nil, ArticleFilter.all.rawValue)
        }
    }
}
