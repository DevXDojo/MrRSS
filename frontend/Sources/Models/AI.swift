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
    let profileName: String
    let configValid: Bool
    let connectionSuccess: Bool
    let modelAvailable: Bool
    let responseTimeMs: Int
    let errorMessage: String?
    let errorCode: String?

    var id: Int { profileID }

    /// True when the profile is usable end to end.
    var succeeded: Bool { configValid && connectionSuccess }

    enum CodingKeys: String, CodingKey {
        case profileID = "profile_id"
        case profileName = "profile_name"
        case configValid = "config_valid"
        case connectionSuccess = "connection_success"
        case modelAvailable = "model_available"
        case responseTimeMs = "response_time_ms"
        case errorMessage = "error_message"
        case errorCode = "error_code"
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        profileID = try container.decodeIfPresent(Int.self, forKey: .profileID) ?? 0
        profileName = try container.decodeIfPresent(String.self, forKey: .profileName) ?? ""
        configValid = try container.decodeIfPresent(Bool.self, forKey: .configValid) ?? false
        connectionSuccess = try container.decodeIfPresent(Bool.self, forKey: .connectionSuccess) ?? false
        modelAvailable = try container.decodeIfPresent(Bool.self, forKey: .modelAvailable) ?? false
        responseTimeMs = try container.decodeIfPresent(Int.self, forKey: .responseTimeMs) ?? 0
        errorMessage = try container.decodeIfPresent(String.self, forKey: .errorMessage)?.nilIfBlank
        errorCode = try container.decodeIfPresent(String.self, forKey: .errorCode)?.nilIfBlank
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

/// One hit from the AI-assisted search: the article plus why it matched.
struct AISearchHit: Identifiable, Codable, Hashable {
    let article: Article
    let relevanceScore: Double
    let matchedTerms: [String]
    let matchedFields: [String]
    let excerpt: String?

    var id: Int { article.id }

    enum CodingKeys: String, CodingKey {
        case excerpt
        case relevanceScore = "relevance_score"
        case matchedTerms = "matched_terms"
        case matchedFields = "matched_fields"
    }

    init(from decoder: Decoder) throws {
        article = try Article(from: decoder)
        let container = try decoder.container(keyedBy: CodingKeys.self)
        relevanceScore = try container.decodeIfPresent(Double.self, forKey: .relevanceScore) ?? 0
        matchedTerms = try container.decodeIfPresent([String].self, forKey: .matchedTerms) ?? []
        matchedFields = try container.decodeIfPresent([String].self, forKey: .matchedFields) ?? []
        excerpt = try container.decodeIfPresent(String.self, forKey: .excerpt)?.nilIfBlank
    }

    func encode(to encoder: Encoder) throws {
        try article.encode(to: encoder)
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encode(relevanceScore, forKey: .relevanceScore)
        try container.encode(matchedTerms, forKey: .matchedTerms)
        try container.encode(matchedFields, forKey: .matchedFields)
        try container.encodeIfPresent(excerpt, forKey: .excerpt)
    }
}

/// What `/api/ai/search` returns.
struct AISearchResponse: Codable, Equatable {
    let success: Bool
    let articles: [AISearchHit]
    let searchTerms: String?
    let error: String?
    let errorCode: String?
    let totalCount: Int

    enum CodingKeys: String, CodingKey {
        case success, articles, error
        case searchTerms = "search_terms"
        case errorCode = "error_code"
        case totalCount = "total_count"
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        success = try container.decodeIfPresent(Bool.self, forKey: .success) ?? false
        articles = try container.decodeIfPresent([AISearchHit].self, forKey: .articles) ?? []
        searchTerms = try container.decodeIfPresent(String.self, forKey: .searchTerms)?.nilIfBlank
        error = try container.decodeIfPresent(String.self, forKey: .error)?.nilIfBlank
        errorCode = try container.decodeIfPresent(String.self, forKey: .errorCode)?.nilIfBlank
        totalCount = try container.decodeIfPresent(Int.self, forKey: .totalCount) ?? 0
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
