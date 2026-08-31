import Foundation

/// A blog found by the discovery engine.
struct DiscoveredBlog: Codable, Hashable, Identifiable {
    let name: String
    let homepage: String
    let rssFeed: String
    let iconURL: String?
    let recentArticles: [DiscoveredArticle]

    var id: String { rssFeed.isEmpty ? homepage : rssFeed }

    enum CodingKeys: String, CodingKey {
        case name, homepage
        case rssFeed = "rss_feed"
        case iconURL = "icon_url"
        case recentArticles = "recent_articles"
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        name = try container.decodeIfPresent(String.self, forKey: .name) ?? ""
        homepage = try container.decodeIfPresent(String.self, forKey: .homepage) ?? ""
        rssFeed = try container.decodeIfPresent(String.self, forKey: .rssFeed) ?? ""
        iconURL = try container.decodeIfPresent(String.self, forKey: .iconURL)?.nilIfBlank
        recentArticles = try container.decodeIfPresent([DiscoveredArticle].self, forKey: .recentArticles) ?? []
    }
}

struct DiscoveredArticle: Codable, Hashable, Identifiable {
    let title: String
    let date: String

    var id: String { title + date }
}

/// How far a discovery run has progressed.
struct DiscoveryProgress: Codable, Hashable {
    var stage: String = ""
    var message: String = ""
    var detail: String = ""
    var current: Int = 0
    var total: Int = 0
    var feedName: String = ""
    var foundCount: Int = 0

    enum CodingKeys: String, CodingKey {
        case stage, message, detail, current, total
        case feedName = "feed_name"
        case foundCount = "found_count"
    }

    init() {}

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        stage = try container.decodeIfPresent(String.self, forKey: .stage) ?? ""
        message = try container.decodeIfPresent(String.self, forKey: .message) ?? ""
        detail = try container.decodeIfPresent(String.self, forKey: .detail) ?? ""
        current = try container.decodeIfPresent(Int.self, forKey: .current) ?? 0
        total = try container.decodeIfPresent(Int.self, forKey: .total) ?? 0
        feedName = try container.decodeIfPresent(String.self, forKey: .feedName) ?? ""
        foundCount = try container.decodeIfPresent(Int.self, forKey: .foundCount) ?? 0
    }

    /// A fraction between 0 and 1, or nil when the total is unknown.
    var fraction: Double? {
        guard total > 0 else { return nil }
        return min(1, Double(current) / Double(total))
    }
}

/// The polled state of a discovery run.
struct DiscoveryState: Codable, Hashable {
    var isRunning: Bool = false
    var isComplete: Bool = false
    var progress = DiscoveryProgress()
    var feeds: [DiscoveredBlog] = []
    var error: String = ""

    enum CodingKeys: String, CodingKey {
        case progress, feeds, error
        case isRunning = "is_running"
        case isComplete = "is_complete"
    }

    init() {}

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        isRunning = try container.decodeIfPresent(Bool.self, forKey: .isRunning) ?? false
        isComplete = try container.decodeIfPresent(Bool.self, forKey: .isComplete) ?? false
        progress = try container.decodeIfPresent(DiscoveryProgress.self, forKey: .progress) ?? DiscoveryProgress()
        feeds = try container.decodeIfPresent([DiscoveredBlog].self, forKey: .feeds) ?? []
        error = try container.decodeIfPresent(String.self, forKey: .error) ?? ""
    }
}

/// Progress of a feed refresh run, as `/api/progress` reports it.
struct RefreshProgress: Codable, Equatable {
    let isRunning: Bool
    /// Refreshes currently being worked on.
    var poolTaskCount: Int = 0
    /// Refreshes waiting for a slot.
    var queueTaskCount: Int = 0
    /// Content fetches triggered by opening an article.
    var articleClickCount: Int = 0
    /// Feeds that failed, keyed by identifier.
    var errors: [String: String] = [:]

    enum CodingKeys: String, CodingKey {
        case errors
        case isRunning = "is_running"
        case poolTaskCount = "pool_task_count"
        case queueTaskCount = "queue_task_count"
        case articleClickCount = "article_click_count"
    }

    init(
        isRunning: Bool,
        poolTaskCount: Int = 0,
        queueTaskCount: Int = 0,
        articleClickCount: Int = 0,
        errors: [String: String] = [:]
    ) {
        self.isRunning = isRunning
        self.poolTaskCount = poolTaskCount
        self.queueTaskCount = queueTaskCount
        self.articleClickCount = articleClickCount
        self.errors = errors
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        isRunning = try container.decodeIfPresent(Bool.self, forKey: .isRunning) ?? false
        poolTaskCount = try container.decodeIfPresent(Int.self, forKey: .poolTaskCount) ?? 0
        queueTaskCount = try container.decodeIfPresent(Int.self, forKey: .queueTaskCount) ?? 0
        articleClickCount = try container.decodeIfPresent(Int.self, forKey: .articleClickCount) ?? 0
        errors = try container.decodeIfPresent([String: String].self, forKey: .errors) ?? [:]
    }

    /// How much work is outstanding altogether.
    var outstandingCount: Int {
        poolTaskCount + queueTaskCount + articleClickCount
    }
}

/// The result of asking the backend whether a newer release exists.
struct UpdateInfo: Codable, Equatable {
    let currentVersion: String
    let latestVersion: String
    let hasUpdate: Bool
    let platform: String
    let arch: String
    let isPortable: Bool
    let downloadURL: String?
    let assetName: String?
    let assetSize: Int?
    let error: String?

    enum CodingKeys: String, CodingKey {
        case platform, arch, error
        case currentVersion = "current_version"
        case latestVersion = "latest_version"
        case hasUpdate = "has_update"
        case isPortable = "is_portable"
        case downloadURL = "download_url"
        case assetName = "asset_name"
        case assetSize = "asset_size"
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        currentVersion = try container.decodeIfPresent(String.self, forKey: .currentVersion) ?? ""
        latestVersion = try container.decodeIfPresent(String.self, forKey: .latestVersion) ?? ""
        hasUpdate = try container.decodeIfPresent(Bool.self, forKey: .hasUpdate) ?? false
        platform = try container.decodeIfPresent(String.self, forKey: .platform) ?? ""
        arch = try container.decodeIfPresent(String.self, forKey: .arch) ?? ""
        isPortable = try container.decodeIfPresent(Bool.self, forKey: .isPortable) ?? false
        downloadURL = try container.decodeIfPresent(String.self, forKey: .downloadURL)?.nilIfBlank
        assetName = try container.decodeIfPresent(String.self, forKey: .assetName)?.nilIfBlank
        assetSize = try container.decodeIfPresent(Int.self, forKey: .assetSize)
        error = try container.decodeIfPresent(String.self, forKey: .error)?.nilIfBlank
    }
}

/// Reading statistics for one period.
struct StatisticsSummary: Codable, Equatable {
    let period: String
    let startDate: String
    let endDate: String
    let totals: [String: Int]
    let dailyData: [String: [String: Int]]
    let canNavigate: Bool
    let hasPrevious: Bool
    let hasNext: Bool
    let displayLabel: String

    enum CodingKeys: String, CodingKey {
        case period, totals
        case startDate = "start_date"
        case endDate = "end_date"
        case dailyData = "daily_data"
        case canNavigate = "can_navigate"
        case hasPrevious = "has_previous"
        case hasNext = "has_next"
        case displayLabel = "display_label"
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        period = try container.decodeIfPresent(String.self, forKey: .period) ?? ""
        startDate = try container.decodeIfPresent(String.self, forKey: .startDate) ?? ""
        endDate = try container.decodeIfPresent(String.self, forKey: .endDate) ?? ""
        totals = try container.decodeIfPresent([String: Int].self, forKey: .totals) ?? [:]
        dailyData = try container.decodeIfPresent([String: [String: Int]].self, forKey: .dailyData) ?? [:]
        canNavigate = try container.decodeIfPresent(Bool.self, forKey: .canNavigate) ?? false
        hasPrevious = try container.decodeIfPresent(Bool.self, forKey: .hasPrevious) ?? false
        hasNext = try container.decodeIfPresent(Bool.self, forKey: .hasNext) ?? false
        displayLabel = try container.decodeIfPresent(String.self, forKey: .displayLabel) ?? ""
    }
}

/// How many articles have cached content.
struct ContentCacheInfo: Codable, Equatable {
    let cachedArticles: Int

    enum CodingKeys: String, CodingKey {
        case cachedArticles = "cached_articles"
    }

    init(cachedArticles: Int) {
        self.cachedArticles = cachedArticles
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        cachedArticles = try container.decodeIfPresent(Int.self, forKey: .cachedArticles) ?? 0
    }
}

/// How much disk the media cache uses.
struct MediaCacheInfo: Codable, Equatable {
    let cacheSizeMB: Double

    enum CodingKeys: String, CodingKey {
        case cacheSizeMB = "cache_size_mb"
    }

    init(cacheSizeMB: Double) {
        self.cacheSizeMB = cacheSizeMB
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        cacheSizeMB = try container.decodeIfPresent(Double.self, forKey: .cacheSizeMB) ?? 0
    }
}

/// What the FreshRSS integration reports about pending synchronisation.
struct FreshRSSStatus: Codable, Equatable {
    let pendingChanges: Int
    let failedItems: Int
    let lastSyncTime: String?

    enum CodingKeys: String, CodingKey {
        case pendingChanges = "pending_changes"
        case failedItems = "failed_items"
        case lastSyncTime = "last_sync_time"
    }

    init(pendingChanges: Int, failedItems: Int, lastSyncTime: String?) {
        self.pendingChanges = pendingChanges
        self.failedItems = failedItems
        self.lastSyncTime = lastSyncTime
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        pendingChanges = try container.decodeIfPresent(Int.self, forKey: .pendingChanges) ?? 0
        failedItems = try container.decodeIfPresent(Int.self, forKey: .failedItems) ?? 0
        lastSyncTime = try container.decodeIfPresent(String.self, forKey: .lastSyncTime)?.nilIfBlank
    }
}

struct TitleTranslationResponse: Codable, Equatable {
    let translatedTitle: String
    let limitReached: Bool

    enum CodingKeys: String, CodingKey {
        case translatedTitle = "translated_title"
        case limitReached = "limit_reached"
    }

    init(translatedTitle: String, limitReached: Bool) {
        self.translatedTitle = translatedTitle
        self.limitReached = limitReached
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        translatedTitle = try container.decodeIfPresent(String.self, forKey: .translatedTitle) ?? ""
        limitReached = try container.decodeIfPresent(Bool.self, forKey: .limitReached) ?? false
    }
}

struct TextTranslationResponse: Codable, Equatable {
    let translatedText: String
    let html: String

    enum CodingKeys: String, CodingKey {
        case translatedText = "translated_text"
        case html
    }

    init(translatedText: String, html: String) {
        self.translatedText = translatedText
        self.html = html
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        translatedText = try container.decodeIfPresent(String.self, forKey: .translatedText) ?? ""
        html = try container.decodeIfPresent(String.self, forKey: .html) ?? ""
    }
}

struct SummaryResult: Codable, Equatable {
    let summary: String
    let html: String?
    let sentenceCount: Int?
    let isTooShort: Bool
    let limitReached: Bool?
    let usedFallback: Bool?
    let thinking: String?
    let error: String?
    let cached: Bool?

    enum CodingKeys: String, CodingKey {
        case summary, html, thinking, error, cached
        case sentenceCount = "sentence_count"
        case isTooShort = "is_too_short"
        case limitReached = "limit_reached"
        case usedFallback = "used_fallback"
    }
}

/// The window geometry the backend remembers between launches.
struct WindowState: Codable, Equatable {
    var x: Int
    var y: Int
    var width: Int
    var height: Int
    var maximized: Bool

    enum CodingKeys: String, CodingKey {
        case x, y, width, height, maximized
    }

    init(x: Int, y: Int, width: Int, height: Int, maximized: Bool = false) {
        self.x = x
        self.y = y
        self.width = width
        self.height = height
        self.maximized = maximized
    }

    /// The values arrive as strings, so each one is read leniently.
    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        func number(_ key: CodingKeys, default fallback: Int) -> Int {
            if let value = try? container.decode(Int.self, forKey: key) { return value }
            if let text = try? container.decode(String.self, forKey: key), let value = Int(text) {
                return value
            }
            return fallback
        }
        x = number(.x, default: 0)
        y = number(.y, default: 0)
        width = number(.width, default: 1_280)
        height = number(.height, default: 780)
        if let value = try? container.decode(Bool.self, forKey: .maximized) {
            maximized = value
        } else if let text = try? container.decode(String.self, forKey: .maximized) {
            maximized = text == "true" || text == "1"
        } else {
            maximized = false
        }
    }

    var jsonBody: [String: Any] {
        ["x": x, "y": y, "width": width, "height": height, "maximized": maximized]
    }
}

/// The custom fetch scripts the backend has available.
struct ScriptList: Codable, Equatable {
    let scripts: [String]
    let scriptsDir: String

    enum CodingKeys: String, CodingKey {
        case scripts
        case scriptsDir = "scripts_dir"
    }

    init(scripts: [String], scriptsDir: String) {
        self.scripts = scripts
        self.scriptsDir = scriptsDir
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        scripts = try container.decodeIfPresent([String].self, forKey: .scripts) ?? []
        scriptsDir = try container.decodeIfPresent(String.self, forKey: .scriptsDir) ?? ""
    }
}
