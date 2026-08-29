import Foundation

/// A field a saved filter can test, matching the identifiers the backend
/// understands in `/api/articles/filter`.
enum FilterField: String, CaseIterable, Identifiable {
    case feedName = "feed_name"
    case feedCategory = "feed_category"
    case feedTags = "feed_tags"
    case articleTitle = "article_title"
    case author
    case url
    case articleContent = "article_content"
    case feedType = "feed_type"
    case isImageModeFeed = "is_image_mode_feed"
    case publishedAfter = "published_after"
    case publishedBefore = "published_before"
    case publishedAfterHours = "published_after_hours"
    case publishedAfterDays = "published_after_days"
    case isRead = "is_read"
    case isFavorite = "is_favorite"
    case isHidden = "is_hidden"
    case isReadLater = "is_read_later"
    case hasSummary = "has_summary"
    case hasTranslation = "has_translation"
    case hasImage = "has_image"
    case hasAudio = "has_audio"
    case hasVideo = "has_video"
    case feedArticlesPerMonth = "feed_articles_per_month"
    case feedLastUpdateStatus = "feed_last_update_status"

    var id: String { rawValue }

    var title: String {
        switch self {
        case .feedName: t("modal.filter.fromFeed")
        case .feedCategory: t("sidebar.sort.byCategory")
        case .feedTags: t("modal.feed.feedTags")
        case .articleTitle: t("common.form.title")
        case .author: t("modal.filter.author")
        case .url: t("modal.filter.url")
        case .articleContent: t("modal.filter.articleContent")
        case .feedType: t("modal.filter.feedType")
        case .isImageModeFeed: t("modal.filter.isImageModeFeed")
        case .publishedAfter: t("modal.filter.publishedAfter")
        case .publishedBefore: t("modal.filter.publishedBefore")
        case .publishedAfterHours: t("modal.filter.publishedAfterHours")
        case .publishedAfterDays: t("modal.filter.publishedAfterDays")
        case .isRead: t("modal.filter.readStatus")
        case .isFavorite: t("modal.filter.favoriteStatus")
        case .isHidden: t("modal.filter.hiddenStatus")
        case .isReadLater: t("modal.filter.readLaterStatus")
        case .hasSummary: t("modal.filter.hasSummary")
        case .hasTranslation: t("modal.filter.hasTranslation")
        case .hasImage: t("modal.filter.hasImage")
        case .hasAudio: t("modal.filter.hasAudio")
        case .hasVideo: t("modal.filter.hasVideo")
        case .feedArticlesPerMonth: t("modal.filter.feedArticlesPerMonth")
        case .feedLastUpdateStatus: t("modal.filter.feedLastUpdateStatus")
        }
    }

    /// How the value for this field is entered.
    enum ValueKind {
        case text
        case number
        case date
        case boolean
        case none
    }

    var valueKind: ValueKind {
        switch self {
        case .articleTitle, .author, .url, .articleContent, .feedName, .feedCategory,
             .feedTags, .feedType, .feedLastUpdateStatus:
            .text
        case .publishedAfterHours, .publishedAfterDays, .feedArticlesPerMonth:
            .number
        case .publishedAfter, .publishedBefore:
            .date
        case .isRead, .isFavorite, .isHidden, .isReadLater, .hasSummary, .hasTranslation,
             .hasImage, .hasAudio, .hasVideo, .isImageModeFeed:
            .boolean
        }
    }

    /// Only the free-text fields offer a matching mode.
    var supportsOperators: Bool {
        switch self {
        case .articleTitle, .author, .url, .articleContent: true
        default: false
        }
    }
}

/// How a text field is compared.
enum FilterOperator: String, CaseIterable, Identifiable {
    case contains
    case exact
    case regex

    var id: String { rawValue }

    var title: String {
        switch self {
        case .contains: t("modal.filter.contains")
        case .exact: t("modal.filter.exactMatch")
        case .regex: t("modal.filter.regex")
        }
    }
}
