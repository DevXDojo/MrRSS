import Foundation

struct Feed: Identifiable, Codable, Hashable {
    let id: Int
    let url: String
    let title: String
    let category: String
    let lastUpdated: String
    let iconURL: String?

    enum CodingKeys: String, CodingKey {
        case id, url, title, category
        case lastUpdated = "last_updated"
        case iconURL = "image_url"
    }

    init(
        id: Int,
        url: String,
        title: String,
        category: String,
        lastUpdated: String = "",
        iconURL: String? = nil
    ) {
        self.id = id
        self.url = url
        self.title = title
        self.category = category
        self.lastUpdated = lastUpdated
        self.iconURL = iconURL?.nilIfEmpty
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(Int.self, forKey: .id)
        url = try container.decodeIfPresent(String.self, forKey: .url) ?? ""
        title = try container.decodeIfPresent(String.self, forKey: .title) ?? url
        category = try container.decodeIfPresent(String.self, forKey: .category) ?? ""
        lastUpdated = try container.decodeIfPresent(String.self, forKey: .lastUpdated) ?? ""
        iconURL = try container.decodeIfPresent(String.self, forKey: .iconURL)?.nilIfEmpty
    }
}

struct Article: Identifiable, Codable, Hashable {
    let id: Int
    let feedID: Int
    let feedTitle: String?
    let title: String
    let url: String
    let imageURL: String?
    let publishedAt: String
    var isRead: Bool
    var isFavorite: Bool
    var isHidden: Bool
    var isReadLater: Bool
    var translatedTitle: String?
    var summary: String?

    enum CodingKeys: String, CodingKey {
        case id, title, url, summary
        case feedID = "feed_id"
        case feedTitle = "feed_title"
        case imageURL = "image_url"
        case publishedAt = "published_at"
        case isRead = "is_read"
        case isFavorite = "is_favorite"
        case isHidden = "is_hidden"
        case isReadLater = "is_read_later"
        case translatedTitle = "translated_title"
    }

    init(
        id: Int,
        feedID: Int,
        feedTitle: String? = nil,
        title: String,
        url: String,
        imageURL: String? = nil,
        publishedAt: String,
        isRead: Bool = false,
        isFavorite: Bool = false,
        isHidden: Bool = false,
        isReadLater: Bool = false,
        translatedTitle: String? = nil,
        summary: String? = nil
    ) {
        self.id = id
        self.feedID = feedID
        self.feedTitle = feedTitle?.nilIfEmpty
        self.title = title
        self.url = url
        self.imageURL = imageURL?.nilIfEmpty
        self.publishedAt = publishedAt
        self.isRead = isRead
        self.isFavorite = isFavorite
        self.isHidden = isHidden
        self.isReadLater = isReadLater
        self.translatedTitle = translatedTitle?.nilIfEmpty
        self.summary = summary?.nilIfEmpty
    }
}

struct UnreadCounts: Codable, Equatable {
    let total: Int
    let feedCounts: [Int: Int]

    enum CodingKeys: String, CodingKey {
        case total
        case feedCounts = "feed_counts"
    }

    static let empty = UnreadCounts(total: 0, feedCounts: [:])

    init(total: Int, feedCounts: [Int: Int]) {
        self.total = total
        self.feedCounts = feedCounts
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        total = try container.decodeIfPresent(Int.self, forKey: .total) ?? 0

        let rawCounts = try container.decodeIfPresent([String: Int].self, forKey: .feedCounts) ?? [:]
        feedCounts = Dictionary(uniqueKeysWithValues: rawCounts.compactMap { key, value in
            Int(key).map { ($0, value) }
        })
    }
}

struct ArticleContent: Codable, Equatable {
    let content: String
    let feedURL: String?

    enum CodingKeys: String, CodingKey {
        case content
        case feedURL = "feed_url"
    }
}

struct RefreshProgress: Codable, Equatable {
    let isRunning: Bool

    enum CodingKeys: String, CodingKey {
        case isRunning = "is_running"
    }
}

struct TitleTranslationResponse: Codable, Equatable {
    let translatedTitle: String
    let limitReached: Bool

    enum CodingKeys: String, CodingKey {
        case translatedTitle = "translated_title"
        case limitReached = "limit_reached"
    }
}

struct TextTranslationResponse: Codable, Equatable {
    let translatedText: String
    let html: String

    enum CodingKeys: String, CodingKey {
        case translatedText = "translated_text"
        case html
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

struct RuleCondition: Codable, Hashable, Identifiable {
    var id: Int
    var logic: String?
    var negate: Bool
    var field: String
    var `operator`: String
    var value: String
    var values: [String]

    static func empty(id: Int = Int(Date().timeIntervalSince1970 * 1_000)) -> RuleCondition {
        RuleCondition(
            id: id,
            logic: "and",
            negate: false,
            field: "article_title",
            operator: "contains",
            value: "",
            values: []
        )
    }
}

struct AutomationRule: Codable, Hashable, Identifiable {
    var id: Int
    var name: String
    var enabled: Bool
    var conditions: [RuleCondition]
    var actions: [String]

    static func empty(id: Int = Int(Date().timeIntervalSince1970 * 1_000)) -> AutomationRule {
        AutomationRule(id: id, name: "New Rule", enabled: true, conditions: [], actions: [])
    }
}

struct RuleApplicationResult: Codable, Equatable {
    let success: Bool
    let affected: Int
}

struct AIUsage: Codable, Equatable {
    let usage: Int
    let limit: Int
    let limitReached: Bool

    enum CodingKeys: String, CodingKey {
        case usage, limit
        case limitReached = "limit_reached"
    }
}

private extension String {
    var nilIfEmpty: String? {
        trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? nil : self
    }
}
