import Foundation

/// A subscription as the backend stores it. Every field the API returns is kept
/// so the feed editor can round-trip a feed without losing configuration.
struct Feed: Identifiable, Codable, Hashable {
    let id: Int
    let url: String
    let title: String
    var category: String
    /// Rank inside its category, as the server keeps it.
    var position: Int
    let lastUpdated: String
    let iconURL: String?
    var link: String
    var feedDescription: String
    var lastError: String
    var hideFromTimeline: Bool
    var refreshInterval: Int
    var isImageMode: Bool
    var scriptPath: String
    var proxyURL: String
    var proxyEnabled: Bool
    var articleViewMode: String
    var autoExpandContent: String
    var type: String
    var xPathItem: String
    var xPathItemTitle: String
    var xPathItemContent: String
    var xPathItemURI: String
    var xPathItemAuthor: String
    var xPathItemTimestamp: String
    var xPathItemTimeFormat: String
    var xPathItemThumbnail: String
    var xPathItemCategories: String
    var xPathItemUID: String
    var emailAddress: String
    var emailIMAPServer: String
    var emailIMAPPort: Int
    var emailUsername: String
    var emailPassword: String
    var emailFolder: String
    var isFreshRSSSource: Bool
    /// The tags assigned to this feed, which the feed listing includes.
    var tags: [Tag]

    enum CodingKeys: String, CodingKey {
        case id, url, title, category, position, link, type
        case feedDescription = "description"
        case lastUpdated = "last_updated"
        case iconURL = "image_url"
        case lastError = "last_error"
        case hideFromTimeline = "hide_from_timeline"
        case refreshInterval = "refresh_interval"
        case isImageMode = "is_image_mode"
        case scriptPath = "script_path"
        case proxyURL = "proxy_url"
        case proxyEnabled = "proxy_enabled"
        case articleViewMode = "article_view_mode"
        case autoExpandContent = "auto_expand_content"
        case xPathItem = "xpath_item"
        case xPathItemTitle = "xpath_item_title"
        case xPathItemContent = "xpath_item_content"
        case xPathItemURI = "xpath_item_uri"
        case xPathItemAuthor = "xpath_item_author"
        case xPathItemTimestamp = "xpath_item_timestamp"
        case xPathItemTimeFormat = "xpath_item_time_format"
        case xPathItemThumbnail = "xpath_item_thumbnail"
        case xPathItemCategories = "xpath_item_categories"
        case xPathItemUID = "xpath_item_uid"
        case emailAddress = "email_address"
        case emailIMAPServer = "email_imap_server"
        case emailIMAPPort = "email_imap_port"
        case emailUsername = "email_username"
        case emailPassword = "email_password"
        case emailFolder = "email_folder"
        case isFreshRSSSource = "is_freshrss_source"
        case tags
    }

    init(
        id: Int,
        url: String,
        title: String,
        category: String,
        position: Int = 0,
        lastUpdated: String = "",
        iconURL: String? = nil,
        link: String = "",
        feedDescription: String = "",
        lastError: String = "",
        hideFromTimeline: Bool = false,
        refreshInterval: Int = 0,
        isImageMode: Bool = false,
        scriptPath: String = "",
        proxyURL: String = "",
        proxyEnabled: Bool = false,
        articleViewMode: String = "global",
        autoExpandContent: String = "global",
        type: String = "",
        xPathItem: String = "",
        xPathItemTitle: String = "",
        xPathItemContent: String = "",
        xPathItemURI: String = "",
        xPathItemAuthor: String = "",
        xPathItemTimestamp: String = "",
        xPathItemTimeFormat: String = "",
        xPathItemThumbnail: String = "",
        xPathItemCategories: String = "",
        xPathItemUID: String = "",
        emailAddress: String = "",
        emailIMAPServer: String = "",
        emailIMAPPort: Int = 993,
        emailUsername: String = "",
        emailPassword: String = "",
        emailFolder: String = "INBOX",
        isFreshRSSSource: Bool = false,
        tags: [Tag] = []
    ) {
        self.id = id
        self.url = url
        self.title = title
        self.category = category
        self.position = position
        self.lastUpdated = lastUpdated
        self.iconURL = iconURL?.nilIfBlank
        self.link = link
        self.feedDescription = feedDescription
        self.lastError = lastError
        self.hideFromTimeline = hideFromTimeline
        self.refreshInterval = refreshInterval
        self.isImageMode = isImageMode
        self.scriptPath = scriptPath
        self.proxyURL = proxyURL
        self.proxyEnabled = proxyEnabled
        self.articleViewMode = articleViewMode
        self.autoExpandContent = autoExpandContent
        self.type = type
        self.xPathItem = xPathItem
        self.xPathItemTitle = xPathItemTitle
        self.xPathItemContent = xPathItemContent
        self.xPathItemURI = xPathItemURI
        self.xPathItemAuthor = xPathItemAuthor
        self.xPathItemTimestamp = xPathItemTimestamp
        self.xPathItemTimeFormat = xPathItemTimeFormat
        self.xPathItemThumbnail = xPathItemThumbnail
        self.xPathItemCategories = xPathItemCategories
        self.xPathItemUID = xPathItemUID
        self.emailAddress = emailAddress
        self.emailIMAPServer = emailIMAPServer
        self.emailIMAPPort = emailIMAPPort
        self.emailUsername = emailUsername
        self.emailPassword = emailPassword
        self.emailFolder = emailFolder
        self.isFreshRSSSource = isFreshRSSSource
        self.tags = tags
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(Int.self, forKey: .id)
        url = try container.decodeIfPresent(String.self, forKey: .url) ?? ""
        title = try container.decodeIfPresent(String.self, forKey: .title) ?? url
        category = try container.decodeIfPresent(String.self, forKey: .category) ?? ""
        position = try container.decodeIfPresent(Int.self, forKey: .position) ?? 0
        lastUpdated = try container.decodeIfPresent(String.self, forKey: .lastUpdated) ?? ""
        iconURL = try container.decodeIfPresent(String.self, forKey: .iconURL)?.nilIfBlank
        link = try container.decodeIfPresent(String.self, forKey: .link) ?? ""
        feedDescription = try container.decodeIfPresent(String.self, forKey: .feedDescription) ?? ""
        lastError = try container.decodeIfPresent(String.self, forKey: .lastError) ?? ""
        hideFromTimeline = try container.decodeIfPresent(Bool.self, forKey: .hideFromTimeline) ?? false
        refreshInterval = try container.decodeIfPresent(Int.self, forKey: .refreshInterval) ?? 0
        isImageMode = try container.decodeIfPresent(Bool.self, forKey: .isImageMode) ?? false
        scriptPath = try container.decodeIfPresent(String.self, forKey: .scriptPath) ?? ""
        proxyURL = try container.decodeIfPresent(String.self, forKey: .proxyURL) ?? ""
        proxyEnabled = try container.decodeIfPresent(Bool.self, forKey: .proxyEnabled) ?? false
        articleViewMode = try container.decodeIfPresent(String.self, forKey: .articleViewMode) ?? "global"
        autoExpandContent = try container.decodeIfPresent(String.self, forKey: .autoExpandContent) ?? "global"
        type = try container.decodeIfPresent(String.self, forKey: .type) ?? ""
        xPathItem = try container.decodeIfPresent(String.self, forKey: .xPathItem) ?? ""
        xPathItemTitle = try container.decodeIfPresent(String.self, forKey: .xPathItemTitle) ?? ""
        xPathItemContent = try container.decodeIfPresent(String.self, forKey: .xPathItemContent) ?? ""
        xPathItemURI = try container.decodeIfPresent(String.self, forKey: .xPathItemURI) ?? ""
        xPathItemAuthor = try container.decodeIfPresent(String.self, forKey: .xPathItemAuthor) ?? ""
        xPathItemTimestamp = try container.decodeIfPresent(String.self, forKey: .xPathItemTimestamp) ?? ""
        xPathItemTimeFormat = try container.decodeIfPresent(String.self, forKey: .xPathItemTimeFormat) ?? ""
        xPathItemThumbnail = try container.decodeIfPresent(String.self, forKey: .xPathItemThumbnail) ?? ""
        xPathItemCategories = try container.decodeIfPresent(String.self, forKey: .xPathItemCategories) ?? ""
        xPathItemUID = try container.decodeIfPresent(String.self, forKey: .xPathItemUID) ?? ""
        emailAddress = try container.decodeIfPresent(String.self, forKey: .emailAddress) ?? ""
        emailIMAPServer = try container.decodeIfPresent(String.self, forKey: .emailIMAPServer) ?? ""
        emailIMAPPort = try container.decodeIfPresent(Int.self, forKey: .emailIMAPPort) ?? 993
        emailUsername = try container.decodeIfPresent(String.self, forKey: .emailUsername) ?? ""
        emailPassword = try container.decodeIfPresent(String.self, forKey: .emailPassword) ?? ""
        emailFolder = try container.decodeIfPresent(String.self, forKey: .emailFolder) ?? "INBOX"
        isFreshRSSSource = try container.decodeIfPresent(Bool.self, forKey: .isFreshRSSSource) ?? false
        tags = try container.decodeIfPresent([Tag].self, forKey: .tags) ?? []
    }

    /// True when the feed is scraped from a page rather than parsed from a feed document.
    var isXPathFeed: Bool {
        type == "HTML+XPath" || type == "XML+XPath"
    }

    /// True when the feed collects newsletters over IMAP.
    var isEmailFeed: Bool {
        !emailAddress.isEmpty || !emailIMAPServer.isEmpty
    }

    /// The site the feed belongs to, used for favicons and "open home page".
    var siteURL: URL? {
        if !link.isEmpty, let url = URL(string: link) { return url }
        guard let feedURL = URL(string: url), let host = feedURL.host else { return nil }
        var components = URLComponents()
        components.scheme = feedURL.scheme ?? "https"
        components.host = host
        return components.url
    }
}

extension String {
    var nilIfBlank: String? {
        trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? nil : self
    }
}
