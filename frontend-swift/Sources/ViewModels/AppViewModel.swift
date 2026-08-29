import Foundation
import SwiftUI

enum ArticleFilter: String, CaseIterable, Identifiable {
    case all
    case unread
    case favorites
    case readLater
    case imageGallery

    var id: String { rawValue }

    var title: String {
        switch self {
        case .all: t("sidebar.activity.allArticles")
        case .unread: t("sidebar.activity.unreadArticles")
        case .favorites: t("sidebar.activity.favorites")
        case .readLater: t("sidebar.activity.readLater")
        case .imageGallery: t("sidebar.activity.imageGallery")
        }
    }

    var icon: String {
        switch self {
        case .all: "rectangle.stack"
        case .unread: "circle.fill"
        case .favorites: "star.fill"
        case .readLater: "clock.fill"
        case .imageGallery: "photo.on.rectangle.angled"
        }
    }

    /// The value `/api/articles` expects for this activity.
    var queryValue: String {
        switch self {
        case .all: "all"
        case .unread: "unread"
        case .favorites: "favorites"
        case .readLater: "readLater"
        case .imageGallery: "all"
        }
    }

    /// Which per-feed count in `/api/articles/filter-counts` belongs to this activity.
    var countsKeyPath: KeyPath<FilterCounts, [Int: Int]>? {
        switch self {
        case .all: nil
        case .unread: \FilterCounts.unread
        case .favorites: \FilterCounts.favorites
        case .readLater: \FilterCounts.readLater
        case .imageGallery: \FilterCounts.images
        }
    }

    /// The count to show when only unread items are being counted.
    var unreadCountsKeyPath: KeyPath<FilterCounts, [Int: Int]>? {
        switch self {
        case .all, .unread: \FilterCounts.unread
        case .favorites: \FilterCounts.favoritesUnread
        case .readLater: \FilterCounts.readLaterUnread
        case .imageGallery: \FilterCounts.imagesUnread
        }
    }
}

/// How the article list is ordered.
enum ArticleSortOrder: String, CaseIterable, Identifiable {
    case newestFirst
    case oldestFirst
    case unreadFirst
    case byTitle

    var id: String { rawValue }

    var title: String {
        switch self {
        case .newestFirst: t("sidebar.sort.latest")
        case .oldestFirst: t("client.sort.oldestFirst")
        case .unreadFirst: t("client.sort.unreadFirst")
        case .byTitle: t("sidebar.sort.byName")
        }
    }
}

/// How each article is presented in the list.
enum ArticleListLayout: String, CaseIterable, Identifiable {
    case compact
    case comfortable
    case cards

    var id: String { rawValue }

    var title: String {
        switch self {
        case .compact: t("client.layout.compact")
        case .comfortable: t("client.layout.comfortable")
        case .cards: t("client.layout.cards")
        }
    }

    var icon: String {
        switch self {
        case .compact: "list.bullet"
        case .comfortable: "list.dash"
        case .cards: "square.grid.2x2"
        }
    }
}

enum SidebarItem: Hashable {
    case filter(ArticleFilter)
    case folder(String)
    case feed(Int)
    case savedFilter(Int)
}

enum ConnectionState: Equatable {
    case connecting
    case connected
    case disconnected

    var title: String {
        switch self {
        case .connecting: t("client.connection.connecting")
        case .connected: t("client.connection.connected")
        case .disconnected: t("client.connection.offline")
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

/// What the article list should ask the backend for.
struct ArticleQuery: Equatable {
    var filter: ArticleFilter?
    var feedID: Int?
    var category: String?
    var conditions: [FilterCondition]

    init(
        filter: ArticleFilter? = nil,
        feedID: Int? = nil,
        category: String? = nil,
        conditions: [FilterCondition] = []
    ) {
        self.filter = filter
        self.feedID = feedID
        self.category = category
        self.conditions = conditions
    }

    /// True when the request goes to the saved-filter endpoint instead of the
    /// plain listing endpoint.
    var usesConditions: Bool { !conditions.isEmpty }

    /// True when the request should come from the multimedia listing.
    var usesImageGallery: Bool { filter == .imageGallery }
}

@MainActor
final class AppViewModel: ObservableObject {
    @Published var feeds: [Feed] = []
    @Published private(set) var folders: [String] = []
    @Published var articles: [Article] = []
    @Published var unreadCounts = UnreadCounts.empty
    @Published var isLoadingArticles = false
    @Published private(set) var isLoadingFeeds = false
    @Published private(set) var isRefreshingSources = false
    @Published private(set) var connectionState: ConnectionState = .connecting
    @Published private(set) var settings: [String: String] = [:]
    @Published private(set) var rules: [AutomationRule] = []
    @Published var aiUsage: AIUsage?
    @Published private(set) var isLoadingSettings = false
    @Published private(set) var isSavingSettings = false
    @Published var statusMessage: String?
    @Published var errorMessage: String?
    @Published var selectedArticleID: Int?
    @Published var serverURLText: String

    /// Per-feed counts for each activity, used for the sidebar badges.
    @Published var filterCounts = FilterCounts.empty
    /// Saved filters, which appear in the sidebar under their own heading.
    @Published var savedFilters: [SavedFilter] = []
    /// Tags, which are assigned to feeds in the feed editor.
    @Published var tags: [Tag] = []
    /// The tag identifiers assigned to each feed.
    @Published var feedTags: [Int: [Int]] = [:]
    /// How far the running refresh has progressed.
    @Published var refreshProgress = RefreshProgress(isRunning: false)

    /// Restrict the current activity to unread items only.
    @Published var showOnlyUnread = false {
        didSet {
            guard showOnlyUnread != oldValue else { return }
            defaults.set(showOnlyUnread, forKey: Self.showOnlyUnreadKey)
            reloadArticles()
        }
    }

    /// How the list is ordered. Ordering is applied on this Mac so switching is
    /// instant and does not refetch.
    @Published var sortOrder: ArticleSortOrder = .newestFirst {
        didSet {
            guard sortOrder != oldValue else { return }
            defaults.set(sortOrder.rawValue, forKey: Self.sortOrderKey)
        }
    }

    /// Raised when a shortcut asks the reading pane to switch between the
    /// rendered article and the original page.
    @Published var requestedViewModeToggle = 0
    /// Drives the add-subscription sheet, which a shortcut can also open.
    @Published var isPresentingAddFeed = false

    /// How much of each article the list shows.
    @Published var listLayout: ArticleListLayout = .comfortable {
        didSet {
            guard listLayout != oldValue else { return }
            defaults.set(listLayout.rawValue, forKey: Self.listLayoutKey)
        }
    }

    @Published var selection: SidebarItem? = .filter(.all) {
        didSet {
            guard selection != oldValue else { return }
            selectedArticleID = nil
            reloadArticles()
        }
    }

    static let showOnlyUnreadKey = "MrRSS.showOnlyUnread"
    static let sortOrderKey = "MrRSS.sortOrder"
    static let listLayoutKey = "MrRSS.listLayout"

    private var page = 0
    private var hasMore = true
    private let limit: Int
    /// Not private so the feature extensions in the neighbouring files can
    /// reach the backend.
    private(set) var api: APIClient
    private let defaults: UserDefaults
    private var feedTask: Task<Void, Never>?
    private var sourceRefreshTask: Task<Void, Never>?
    private var feedRequestID = UUID()
    private var articleTask: Task<Void, Never>?
    private var articleRequestID = UUID()

    init(
        api: APIClient = APIService.shared,
        limit: Int = 50,
        autoLoad: Bool = true,
        defaults: UserDefaults = .standard
    ) {
        self.api = api
        self.limit = limit
        self.defaults = defaults
        serverURLText = api.baseURL.absoluteString
        showOnlyUnread = defaults.bool(forKey: Self.showOnlyUnreadKey)
        sortOrder = ArticleSortOrder(rawValue: defaults.string(forKey: Self.sortOrderKey) ?? "")
            ?? .newestFirst
        listLayout = ArticleListLayout(rawValue: defaults.string(forKey: Self.listLayoutKey) ?? "")
            ?? .comfortable
        refreshFolders()

        if autoLoad {
            refreshAll()
        }
    }

    func refreshAll() {
        refreshFeeds()
        reloadArticles()
        Task { [weak self] in
            guard let self else { return }
            await loadSettings()
            await refreshCounts()
            await loadSavedFilters()
            await loadTags()
        }
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
                feeds = AppViewModel.ordered(loadedFeeds)
                refreshFolders()
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
            errorMessage = t("client.feed.invalidURL")
            return false
        }

        do {
            try await api.addFeed(
                FeedDraft(
                    url: url.trimmingCharacters(in: .whitespacesAndNewlines),
                    title: title.trimmingCharacters(in: .whitespacesAndNewlines),
                    category: category.trimmingCharacters(in: .whitespacesAndNewlines)
                )
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


    // MARK: - Folders

    /// A folder is the category stored on each feed. A folder that somebody
    /// created but has not filled yet has nowhere to live on the server, so its
    /// name is remembered on this Mac until a feed moves into it.
    private static let pendingFoldersKey = "MrRSS.pendingFolders"

    private var pendingFolders: Set<String> {
        get { Set(defaults.stringArray(forKey: Self.pendingFoldersKey) ?? []) }
        set { defaults.set(newValue.sorted(), forKey: Self.pendingFoldersKey) }
    }

    func feeds(inFolder folder: String) -> [Feed] {
        feeds.filter { $0.category == folder }
    }

    var unfiledFeeds: [Feed] {
        feeds.filter { $0.category.isEmpty }
    }

    func unreadCount(forFolder folder: String) -> Int {
        feeds(inFolder: folder).reduce(0) { $0 + (unreadCounts.feedCounts[$1.id] ?? 0) }
    }

    @discardableResult
    func createFolder(named name: String) -> Bool {
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else {
            errorMessage = t("client.folder.nameRequired")
            return false
        }
        guard !folders.contains(trimmed) else {
            errorMessage = t("client.folder.alreadyExists", ["name": trimmed])
            return false
        }

        pendingFolders.insert(trimmed)
        refreshFolders()
        return true
    }

    func moveFeed(_ feed: Feed, toFolder folder: String?) async {
        let target = (folder ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
        guard feed.category != target else { return }

        do {
            try await api.updateFeedCategory(id: feed.id, category: target)
        } catch {
            errorMessage = error.localizedDescription
            return
        }

        if let index = feeds.firstIndex(where: { $0.id == feed.id }) {
            feeds[index].category = target
        }
        refreshFolders()
    }

    /// Moves everything a drag carried. Identifiers that no longer match a
    /// subscription are skipped rather than failing the whole drop.
    func moveFeeds(ids: [Int], toFolder folder: String?) async {
        for id in ids {
            guard let feed = feeds.first(where: { $0.id == id }) else { continue }
            await moveFeed(feed, toFolder: folder)
        }
    }

    /// Files everything a drag carried directly above or below one of the rows
    /// it was dropped on, which is how the sidebar's own order is changed.
    func moveFeeds(ids: [Int], relativeTo referenceID: Int, placeAbove: Bool) async {
        var anchorID = referenceID
        var above = placeAbove

        for id in ids where id != referenceID {
            guard let anchor = feeds.first(where: { $0.id == anchorID }),
                  feeds.contains(where: { $0.id == id }) else { continue }

            let siblings = feeds(inFolder: anchor.category).filter { $0.id != id }
            let anchorIndex = siblings.firstIndex(where: { $0.id == anchorID }) ?? siblings.count
            await reorderFeed(id: id, category: anchor.category, index: above ? anchorIndex : anchorIndex + 1)

            // Anything after the first lands just below what came before it, so
            // a multiple selection keeps the order it was dragged in.
            anchorID = id
            above = false
        }
    }

    /// Commits what the drag preview was showing.
    func placeFeed(id: Int, inFolder folder: String, at index: Int) async {
        guard let feed = feeds.first(where: { $0.id == id }) else { return }

        let siblings = feeds(inFolder: folder).filter { $0.id != id }
        let clamped = min(max(0, index), siblings.count)
        let currentIndex = feeds(inFolder: folder).firstIndex(where: { $0.id == id })
        guard feed.category != folder || currentIndex != clamped else { return }

        await reorderFeed(id: id, category: folder, index: clamped)
    }

    private func reorderFeed(id: Int, category: String, index: Int) async {
        do {
            try await api.reorderFeed(id: id, category: category, position: index)
        } catch {
            errorMessage = error.localizedDescription
            refreshFeeds()
            return
        }

        applyLocalOrder(feedID: id, category: category, index: index)
        refreshFolders()
    }

    /// Mirrors the ranking the server just performed so the sidebar settles
    /// immediately instead of waiting for the next load.
    private func applyLocalOrder(feedID: Int, category: String, index: Int) {
        guard var moving = feeds.first(where: { $0.id == feedID }) else { return }
        moving.category = category

        var siblings = feeds.filter { $0.category == category && $0.id != feedID }
        siblings.insert(moving, at: min(max(0, index), siblings.count))

        for (rank, sibling) in siblings.enumerated() {
            guard let position = feeds.firstIndex(where: { $0.id == sibling.id }) else { continue }
            feeds[position].category = category
            feeds[position].position = rank
        }
        feeds = AppViewModel.ordered(feeds)
    }

    private static func ordered(_ feeds: [Feed]) -> [Feed] {
        feeds.sorted { lhs, rhs in
            lhs.position == rhs.position ? lhs.id < rhs.id : lhs.position < rhs.position
        }
    }

    func renameFolder(_ folder: String, to newName: String) async {
        let trimmed = newName.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty, trimmed != folder else { return }
        guard !folders.contains(trimmed) else {
            errorMessage = t("client.folder.alreadyExists", ["name": trimmed])
            return
        }

        for feed in feeds(inFolder: folder) {
            do {
                try await api.updateFeedCategory(id: feed.id, category: trimmed)
            } catch {
                errorMessage = error.localizedDescription
                refreshFeeds()
                return
            }
        }

        for index in feeds.indices where feeds[index].category == folder {
            feeds[index].category = trimmed
        }
        var stored = pendingFolders
        stored.remove(folder)
        stored.insert(trimmed)
        pendingFolders = stored
        if selection == .folder(folder) {
            selection = .folder(trimmed)
        }
        refreshFolders()
    }

    /// Removes the folder itself. The feeds it held stay subscribed and move
    /// back out of any folder.
    func deleteFolder(_ folder: String) async {
        for feed in feeds(inFolder: folder) {
            do {
                try await api.updateFeedCategory(id: feed.id, category: "")
            } catch {
                errorMessage = error.localizedDescription
                refreshFeeds()
                return
            }
        }

        for index in feeds.indices where feeds[index].category == folder {
            feeds[index].category = ""
        }
        pendingFolders.remove(folder)
        if selection == .folder(folder) {
            selection = .filter(.all)
        }
        refreshFolders()
    }

    private func refreshFolders() {
        let assigned = Set(feeds.map(\.category).filter { !$0.isEmpty })
        let stored = pendingFolders.subtracting(assigned)
        pendingFolders = stored
        folders = assigned.union(stored).sorted { $0.localizedStandardCompare($1) == .orderedAscending }
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
        markReadOnOpen(article)
    }

    /// Opening an article marks it read, which is what the previous interface
    /// did. Articles kept for later are left alone.
    func markReadOnOpen(_ article: Article) {
        guard !article.isRead, !article.isReadLater else { return }
        setArticleRead(article, read: true)
    }

    /// True while more pages remain for the current selection.
    var hasMoreArticles: Bool { hasMore }

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
            errorMessage = t("client.server.invalidAddress")
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
            applyLanguageSetting()
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
            statusMessage = t("client.settings.saved")
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
            statusMessage = t("client.rule.applied", ["count": result.affected])
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
            statusMessage = t("client.ai.limitReachedFallback")
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
            statusMessage = t("client.maintenance.cleared")
            reloadArticles()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func resetAIUsage() async {
        do {
            try await api.resetAIUsage()
            aiUsage = try await api.fetchAIUsage()
            statusMessage = t("client.ai.usageReset")
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

    /// Follows the language stored on the server, which is where the previous
    /// interface kept it too.
    func applyLanguageSetting() {
        Localization.shared.setLanguage(AppLanguage.from(settingValue: settings["language"]))
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
                let loadedArticles = try await load(query: query, page: targetPage)
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

    /// Runs one page of whichever request the current selection needs.
    private func load(query: ArticleQuery, page: Int) async throws -> [Article] {
        if query.usesConditions {
            let response = try await api.filterArticles(
                conditions: query.conditions,
                page: page,
                limit: limit
            )
            return applyUnreadRestriction(response.articles)
        }

        if query.usesImageGallery {
            return applyUnreadRestriction(try await api.fetchImageArticles(page: page, limit: limit))
        }

        let filterValue = query.filter?.queryValue ?? (showOnlyUnread ? "unread" : "")
        let articles = try await api.fetchArticles(
            feedID: query.feedID,
            category: query.category,
            filter: filterValue,
            page: page,
            limit: limit
        )
        return applyUnreadRestriction(articles)
    }

    /// The plain listing understands "unread only" through its filter, but the
    /// multimedia and saved-filter endpoints do not, so the restriction is
    /// applied here for those.
    private func applyUnreadRestriction(_ articles: [Article]) -> [Article] {
        guard showOnlyUnread else { return articles }
        return articles.filter { !$0.isRead }
    }

    /// The request the current sidebar selection maps to.
    var articleQuery: ArticleQuery {
        switch selection {
        case .filter(let filter):
            return ArticleQuery(filter: filter)
        case .folder(let name):
            return ArticleQuery(category: name)
        case .feed(let id):
            return ArticleQuery(feedID: id)
        case .savedFilter(let id):
            let conditions = savedFilters.first(where: { $0.id == id })?.conditions ?? []
            return ArticleQuery(conditions: conditions)
        case .none:
            return ArticleQuery(filter: .all)
        }
    }
}
