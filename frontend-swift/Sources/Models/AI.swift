import Foundation

/// A saved AI configuration. The API key is only returned when the backend is
/// asked for it explicitly, so it is optional here.
struct AIProfile: Identifiable, Codable, Hashable {
    let id: Int
    var name: String
    var apiKey: String
    var endpoint: String
    var model: String
    var customHeaders: String
    var isDefault: Bool

    enum CodingKeys: String, CodingKey {
        case id, name, endpoint, model
        case apiKey = "api_key"
        case customHeaders = "custom_headers"
        case isDefault = "is_default"
    }

    init(
        id: Int = 0,
        name: String = "",
        apiKey: String = "",
        endpoint: String = "",
        model: String = "",
        customHeaders: String = "",
        isDefault: Bool = false
    ) {
        self.id = id
        self.name = name
        self.apiKey = apiKey
        self.endpoint = endpoint
        self.model = model
        self.customHeaders = customHeaders
        self.isDefault = isDefault
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decodeIfPresent(Int.self, forKey: .id) ?? 0
        name = try container.decodeIfPresent(String.self, forKey: .name) ?? ""
        apiKey = try container.decodeIfPresent(String.self, forKey: .apiKey) ?? ""
        endpoint = try container.decodeIfPresent(String.self, forKey: .endpoint) ?? ""
        model = try container.decodeIfPresent(String.self, forKey: .model) ?? ""
        customHeaders = try container.decodeIfPresent(String.self, forKey: .customHeaders) ?? ""
        isDefault = try container.decodeIfPresent(Bool.self, forKey: .isDefault) ?? false
    }
}

/// The outcome of testing one AI profile.
struct AIProfileTestResult: Codable, Equatable, Identifiable {
    let profileID: Int
    let name: String
    let success: Bool
    let message: String

    var id: Int { profileID }

    enum CodingKeys: String, CodingKey {
        case name, success, message
        case profileID = "profile_id"
    }

    init(profileID: Int, name: String, success: Bool, message: String) {
        self.profileID = profileID
        self.name = name
        self.success = success
        self.message = message
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        profileID = try container.decodeIfPresent(Int.self, forKey: .profileID) ?? 0
        name = try container.decodeIfPresent(String.self, forKey: .name) ?? ""
        success = try container.decodeIfPresent(Bool.self, forKey: .success) ?? false
        if let message = try container.decodeIfPresent(String.self, forKey: .message) {
            self.message = message
        } else {
            message = ""
        }
    }
}

/// A saved conversation about one article.
struct ChatSession: Identifiable, Codable, Hashable {
    let id: Int
    let articleID: Int
    var title: String
    let createdAt: String
    let updatedAt: String
    let messageCount: Int

    enum CodingKeys: String, CodingKey {
        case id, title
        case articleID = "article_id"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case messageCount = "message_count"
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(Int.self, forKey: .id)
        articleID = try container.decodeIfPresent(Int.self, forKey: .articleID) ?? 0
        title = try container.decodeIfPresent(String.self, forKey: .title) ?? ""
        createdAt = try container.decodeIfPresent(String.self, forKey: .createdAt) ?? ""
        updatedAt = try container.decodeIfPresent(String.self, forKey: .updatedAt) ?? ""
        messageCount = try container.decodeIfPresent(Int.self, forKey: .messageCount) ?? 0
    }

    init(id: Int, articleID: Int, title: String, createdAt: String = "", updatedAt: String = "", messageCount: Int = 0) {
        self.id = id
        self.articleID = articleID
        self.title = title
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.messageCount = messageCount
    }
}

/// One stored chat message.
struct ChatMessage: Identifiable, Codable, Hashable {
    let id: Int
    let sessionID: Int
    let role: String
    let content: String
    let html: String?
    let thinking: String?
    let createdAt: String

    enum CodingKeys: String, CodingKey {
        case id, role, content, html, thinking
        case sessionID = "session_id"
        case createdAt = "created_at"
    }

    init(
        id: Int,
        sessionID: Int,
        role: String,
        content: String,
        html: String? = nil,
        thinking: String? = nil,
        createdAt: String = ""
    ) {
        self.id = id
        self.sessionID = sessionID
        self.role = role
        self.content = content
        self.html = html
        self.thinking = thinking
        self.createdAt = createdAt
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decodeIfPresent(Int.self, forKey: .id) ?? 0
        sessionID = try container.decodeIfPresent(Int.self, forKey: .sessionID) ?? 0
        role = try container.decodeIfPresent(String.self, forKey: .role) ?? "user"
        content = try container.decodeIfPresent(String.self, forKey: .content) ?? ""
        html = try container.decodeIfPresent(String.self, forKey: .html)?.nilIfBlank
        thinking = try container.decodeIfPresent(String.self, forKey: .thinking)?.nilIfBlank
        createdAt = try container.decodeIfPresent(String.self, forKey: .createdAt) ?? ""
    }

    var isAssistant: Bool { role == "assistant" }
}

/// What the chat endpoint returns for one exchange.
struct ChatResponse: Codable, Equatable {
    let response: String
    let html: String?
    let thinking: String?
    let sessionID: Int?
    let historySaved: Bool

    enum CodingKeys: String, CodingKey {
        case response, html, thinking
        case sessionID = "session_id"
        case historySaved = "history_saved"
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        response = try container.decodeIfPresent(String.self, forKey: .response) ?? ""
        html = try container.decodeIfPresent(String.self, forKey: .html)?.nilIfBlank
        thinking = try container.decodeIfPresent(String.self, forKey: .thinking)?.nilIfBlank
        sessionID = try container.decodeIfPresent(Int.self, forKey: .sessionID)
        historySaved = try container.decodeIfPresent(Bool.self, forKey: .historySaved) ?? false
    }
}

/// A single hit from the AI-assisted article search.
struct AISearchResult: Identifiable, Codable, Hashable {
    let articleID: Int
    let title: String
    let url: String
    let feedTitle: String?
    let reason: String?
    let score: Double?

    var id: Int { articleID }

    enum CodingKeys: String, CodingKey {
        case title, url, reason, score
        case articleID = "article_id"
        case feedTitle = "feed_title"
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        articleID = try container.decodeIfPresent(Int.self, forKey: .articleID) ?? 0
        title = try container.decodeIfPresent(String.self, forKey: .title) ?? ""
        url = try container.decodeIfPresent(String.self, forKey: .url) ?? ""
        feedTitle = try container.decodeIfPresent(String.self, forKey: .feedTitle)?.nilIfBlank
        reason = try container.decodeIfPresent(String.self, forKey: .reason)?.nilIfBlank
        score = try container.decodeIfPresent(Double.self, forKey: .score)
    }
}

struct AIUsage: Codable, Equatable {
    let usage: Int
    let limit: Int
    let limitReached: Bool

    enum CodingKeys: String, CodingKey {
        case usage, limit
        case limitReached = "limit_reached"
    }

    init(usage: Int, limit: Int, limitReached: Bool) {
        self.usage = usage
        self.limit = limit
        self.limitReached = limitReached
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        usage = try container.decodeIfPresent(Int.self, forKey: .usage) ?? 0
        limit = try container.decodeIfPresent(Int.self, forKey: .limit) ?? 0
        limitReached = try container.decodeIfPresent(Bool.self, forKey: .limitReached) ?? false
    }
}
