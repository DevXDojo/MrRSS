import Foundation

/// An article as the backend returns it.
struct Article: Identifiable, Codable, Hashable {
    let id: Int
    let feedID: Int
    let feedTitle: String?
    let title: String
    let url: String
    let imageURL: String?
    let audioURL: String?
    let videoURL: String?
    let author: String?
    let publishedAt: String
    var isRead: Bool
    var isFavorite: Bool
    var isHidden: Bool
    var isReadLater: Bool
    var translatedTitle: String?
    var summary: String?
    var originalSummary: String?
    var freshRSSItemID: String?

    enum CodingKeys: String, CodingKey {
        case id, title, url, summary, author
        case feedID = "feed_id"
        case feedTitle = "feed_title"
        case imageURL = "image_url"
        case audioURL = "audio_url"
        case videoURL = "video_url"
        case publishedAt = "published_at"
        case isRead = "is_read"
        case isFavorite = "is_favorite"
        case isHidden = "is_hidden"
        case isReadLater = "is_read_later"
        case translatedTitle = "translated_title"
        case originalSummary = "original_summary"
        case freshRSSItemID = "freshrss_item_id"
    }

    init(
        id: Int,
        feedID: Int,
        feedTitle: String? = nil,
        title: String,
        url: String,
        imageURL: String? = nil,
        audioURL: String? = nil,
        videoURL: String? = nil,
        author: String? = nil,
        publishedAt: String,
        isRead: Bool = false,
        isFavorite: Bool = false,
        isHidden: Bool = false,
        isReadLater: Bool = false,
        translatedTitle: String? = nil,
        summary: String? = nil,
        originalSummary: String? = nil,
        freshRSSItemID: String? = nil
    ) {
        self.id = id
        self.feedID = feedID
        self.feedTitle = feedTitle?.nilIfBlank
        self.title = title
        self.url = url
        self.imageURL = imageURL?.nilIfBlank
        self.audioURL = audioURL?.nilIfBlank
        self.videoURL = videoURL?.nilIfBlank
        self.author = author?.nilIfBlank
        self.publishedAt = publishedAt
        self.isRead = isRead
        self.isFavorite = isFavorite
        self.isHidden = isHidden
        self.isReadLater = isReadLater
        self.translatedTitle = translatedTitle?.nilIfBlank
        self.summary = summary?.nilIfBlank
        self.originalSummary = originalSummary?.nilIfBlank
        self.freshRSSItemID = freshRSSItemID?.nilIfBlank
    }

    /// The title to show, preferring the translation when one exists.
    func displayTitle(preferTranslation: Bool) -> String {
        guard preferTranslation, let translatedTitle, !translatedTitle.isEmpty else { return title }
        return translatedTitle
    }

    var publishedDate: Date? {
        ArticleDateFormatter.date(from: publishedAt)
    }

    var hasMedia: Bool {
        audioURL != nil || videoURL != nil
    }
}

/// Parses the timestamps the backend emits, which are RFC 3339 with or without
/// fractional seconds.
enum ArticleDateFormatter {
    private static let withFractionalSeconds: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return formatter
    }()

    private static let withoutFractionalSeconds: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime]
        return formatter
    }()

    static func date(from string: String) -> Date? {
        if let date = withFractionalSeconds.date(from: string) { return date }
        return withoutFractionalSeconds.date(from: string)
    }

    /// A short, human-readable form such as "3h ago" or "12 Mar".
    static func relativeDescription(for string: String, now: Date = Date()) -> String {
        guard let date = date(from: string) else { return "" }
        let interval = now.timeIntervalSince(date)

        if interval < 60 {
            return t("common.time.justNow")
        }
        if interval < 3_600 {
            return t("common.time.minutesAgo", ["count": Int(interval / 60)])
        }
        if interval < 86_400 {
            return t("common.time.hoursAgo", ["count": Int(interval / 3_600)])
        }
        if interval < 7 * 86_400 {
            return t("common.time.daysAgo", ["count": Int(interval / 86_400)])
        }

        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        formatter.timeStyle = .none
        return formatter.string(from: date)
    }

    static func fullDescription(for string: String) -> String {
        guard let date = date(from: string) else { return string }
        let formatter = DateFormatter()
        formatter.dateStyle = .long
        formatter.timeStyle = .short
        return formatter.string(from: date)
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

/// Per-feed counts for each activity filter, as `/api/articles/filter-counts` returns them.
struct FilterCounts: Codable, Equatable {
    var unread: [Int: Int] = [:]
    var favorites: [Int: Int] = [:]
    var favoritesUnread: [Int: Int] = [:]
    var readLater: [Int: Int] = [:]
    var readLaterUnread: [Int: Int] = [:]
    var images: [Int: Int] = [:]
    var imagesUnread: [Int: Int] = [:]

    static let empty = FilterCounts()

    enum CodingKeys: String, CodingKey {
        case unread, favorites, images
        case favoritesUnread = "favorites_unread"
        case readLater = "read_later"
        case readLaterUnread = "read_later_unread"
        case imagesUnread = "images_unread"
    }

    init() {}

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        func counts(_ key: CodingKeys) throws -> [Int: Int] {
            let raw = try container.decodeIfPresent([String: Int].self, forKey: key) ?? [:]
            return Dictionary(uniqueKeysWithValues: raw.compactMap { key, value in
                Int(key).map { ($0, value) }
            })
        }
        unread = try counts(.unread)
        favorites = try counts(.favorites)
        favoritesUnread = try counts(.favoritesUnread)
        readLater = try counts(.readLater)
        readLaterUnread = try counts(.readLaterUnread)
        images = try counts(.images)
        imagesUnread = try counts(.imagesUnread)
    }

    func total(for keyPath: KeyPath<FilterCounts, [Int: Int]>) -> Int {
        self[keyPath: keyPath].values.reduce(0, +)
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

/// The images the backend extracted from an article, used by the gallery view.
struct ArticleImages: Codable, Equatable {
    let images: [String]

    init(images: [String]) {
        self.images = images
    }

    init(from decoder: Decoder) throws {
        if let container = try? decoder.container(keyedBy: CodingKeys.self) {
            images = try container.decodeIfPresent([String].self, forKey: .images) ?? []
        } else {
            let single = try decoder.singleValueContainer()
            images = (try? single.decode([String].self)) ?? []
        }
    }

    enum CodingKeys: String, CodingKey {
        case images
    }
}
