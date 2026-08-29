import Foundation

enum ServerConfiguration {
    static let storageKey = "mrrss.apiBaseURL"
    static let fallbackURL = URL(string: "http://127.0.0.1:1234/api")!

    static var savedBaseURL: URL {
        if let environmentValue = ProcessInfo.processInfo.environment["MRRSS_API_BASE_URL"],
           let url = normalizedURL(from: environmentValue) {
            return url
        }

        if let savedValue = UserDefaults.standard.string(forKey: storageKey),
           let url = normalizedURL(from: savedValue) {
            return url
        }

        return fallbackURL
    }

    static func normalizedURL(from value: String) -> URL? {
        let trimmed = value.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }

        let candidate = trimmed.contains("://") ? trimmed : "http://\(trimmed)"
        guard var components = URLComponents(string: candidate),
              let scheme = components.scheme?.lowercased(),
              ["http", "https"].contains(scheme),
              components.host != nil else {
            return nil
        }

        var path = components.path
        while path.hasSuffix("/") {
            path.removeLast()
        }
        if !path.hasSuffix("/api") {
            path += "/api"
        }
        components.path = path
        components.query = nil
        components.fragment = nil
        return components.url
    }
}

enum APIError: LocalizedError, Equatable {
    case invalidURL
    case invalidResponse
    case server(statusCode: Int, message: String)
    case decoding(String)

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "The server address is invalid."
        case .invalidResponse:
            return "The server returned an invalid response."
        case .server(let statusCode, let message):
            return message.isEmpty
                ? "The server returned HTTP \(statusCode)."
                : "The server returned HTTP \(statusCode): \(message)"
        case .decoding(let message):
            return "The server response could not be read: \(message)"
        }
    }
}

protocol APIClient: AnyObject {
    var baseURL: URL { get }
    func checkConnection() async throws
    func fetchFeeds() async throws -> [Feed]
    func fetchUnreadCounts() async throws -> UnreadCounts
    func addFeed(url: String, title: String, category: String) async throws
    func deleteFeed(id: Int) async throws
    func updateFeedCategory(id: Int, category: String) async throws
    func reorderFeed(id: Int, category: String, position: Int) async throws
    func refreshAllFeeds() async throws
    func fetchRefreshProgress() async throws -> RefreshProgress
    func fetchArticles(
        feedID: Int?,
        category: String?,
        filter: String,
        page: Int,
        limit: Int
    ) async throws -> [Article]
    func setArticleRead(id: Int, read: Bool) async throws
    func toggleFavorite(id: Int) async throws
    func fetchArticleContent(id: Int) async throws -> ArticleContent
    func fetchSettings() async throws -> [String: String]
    func updateSettings(_ settings: [String: String]) async throws
    func translateTitle(articleID: Int, title: String, targetLanguage: String) async throws -> TitleTranslationResponse
    func translateText(_ text: String, targetLanguage: String) async throws -> TextTranslationResponse
    func summarize(articleID: Int, length: String, content: String?) async throws -> SummaryResult
    func clearTranslations() async throws
    func clearSummaries() async throws
    func applyRule(_ rule: AutomationRule) async throws -> RuleApplicationResult
    func fetchAIUsage() async throws -> AIUsage
    func resetAIUsage() async throws
}

final class APIService: APIClient {
    static let shared = APIService()

    private(set) var baseURL: URL
    private let session: URLSession

    init(baseURL: URL = ServerConfiguration.savedBaseURL, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session
    }

    func updateBaseURL(_ url: URL, persist: Bool = true) {
        baseURL = url
        if persist {
            UserDefaults.standard.set(url.absoluteString, forKey: ServerConfiguration.storageKey)
        }
    }

    func checkConnection() async throws {
        let url = try makeURL(endpoint: "version")
        _ = try await data(for: URLRequest(url: url))
    }

    func fetchFeeds() async throws -> [Feed] {
        try await get("feeds")
    }

    func fetchUnreadCounts() async throws -> UnreadCounts {
        try await get("articles/unread-counts")
    }

    func addFeed(url: String, title: String, category: String) async throws {
        try await post(
            "feeds/add",
            jsonBody: [
                "url": url,
                "title": title,
                "category": category
            ]
        )
    }

    func deleteFeed(id: Int) async throws {
        try await post(
            "feeds/delete",
            queryItems: [URLQueryItem(name: "id", value: String(id))]
        )
    }

    func updateFeedCategory(id: Int, category: String) async throws {
        try await post(
            "feeds/category",
            jsonBody: ["id": id, "category": category]
        )
    }

    func reorderFeed(id: Int, category: String, position: Int) async throws {
        try await post(
            "feeds/reorder",
            jsonBody: ["feed_id": id, "category": category, "position": position]
        )
    }

    func refreshAllFeeds() async throws {
        try await post("refresh")
    }

    func fetchRefreshProgress() async throws -> RefreshProgress {
        try await get("progress")
    }

    func fetchArticles(
        feedID: Int? = nil,
        category: String? = nil,
        filter: String,
        page: Int = 1,
        limit: Int = 50
    ) async throws -> [Article] {
        var queryItems = [
            URLQueryItem(name: "page", value: String(page)),
            URLQueryItem(name: "limit", value: String(limit)),
            URLQueryItem(name: "filter", value: filter)
        ]

        if let feedID {
            queryItems.append(URLQueryItem(name: "feed_id", value: String(feedID)))
        }
        if let category, !category.isEmpty {
            queryItems.append(URLQueryItem(name: "category", value: category))
        }

        return try await get("articles", queryItems: queryItems)
    }

    func setArticleRead(id: Int, read: Bool) async throws {
        try await post(
            "articles/read",
            queryItems: [
                URLQueryItem(name: "id", value: String(id)),
                URLQueryItem(name: "read", value: String(read))
            ]
        )
    }

    func toggleFavorite(id: Int) async throws {
        try await post(
            "articles/favorite",
            queryItems: [URLQueryItem(name: "id", value: String(id))]
        )
    }

    func fetchArticleContent(id: Int) async throws -> ArticleContent {
        try await get(
            "articles/content",
            queryItems: [URLQueryItem(name: "id", value: String(id))]
        )
    }

    func fetchSettings() async throws -> [String: String] {
        try await get("settings")
    }

    func updateSettings(_ settings: [String: String]) async throws {
        try await sendJSON("settings", body: settings)
    }

    func translateTitle(
        articleID: Int,
        title: String,
        targetLanguage: String
    ) async throws -> TitleTranslationResponse {
        struct Request: Encodable {
            let articleID: Int
            let title: String
            let targetLanguage: String

            enum CodingKeys: String, CodingKey {
                case articleID = "article_id"
                case title
                case targetLanguage = "target_language"
            }
        }

        return try await postJSON(
            "articles/translate",
            body: Request(articleID: articleID, title: title, targetLanguage: targetLanguage)
        )
    }

    func translateText(_ text: String, targetLanguage: String) async throws -> TextTranslationResponse {
        struct Request: Encodable {
            let text: String
            let targetLanguage: String

            enum CodingKeys: String, CodingKey {
                case text
                case targetLanguage = "target_language"
            }
        }

        return try await postJSON(
            "articles/translate-text",
            body: Request(text: text, targetLanguage: targetLanguage)
        )
    }

    func summarize(articleID: Int, length: String, content: String?) async throws -> SummaryResult {
        struct Request: Encodable {
            let articleID: Int
            let length: String
            let content: String?

            enum CodingKeys: String, CodingKey {
                case articleID = "article_id"
                case length, content
            }
        }

        return try await postJSON(
            "articles/summarize",
            body: Request(articleID: articleID, length: length, content: content)
        )
    }

    func clearTranslations() async throws {
        try await post("articles/clear-translations")
    }

    func clearSummaries() async throws {
        try await send("articles/clear-summaries", method: "DELETE")
    }

    func applyRule(_ rule: AutomationRule) async throws -> RuleApplicationResult {
        try await postJSON("rules/apply", body: rule)
    }

    func fetchAIUsage() async throws -> AIUsage {
        try await get("ai-usage")
    }

    func resetAIUsage() async throws {
        try await post("ai-usage/reset")
    }

    private func get<T: Decodable>(
        _ endpoint: String,
        queryItems: [URLQueryItem] = []
    ) async throws -> T {
        let url = try makeURL(endpoint: endpoint, queryItems: queryItems)
        let data = try await data(for: URLRequest(url: url))

        if data.trimmingWhitespace == Data("null".utf8),
           let emptyArray = [] as? T {
            return emptyArray
        }

        do {
            return try JSONDecoder().decode(T.self, from: data)
        } catch {
            throw APIError.decoding(error.localizedDescription)
        }
    }

    private func post(
        _ endpoint: String,
        queryItems: [URLQueryItem] = [],
        jsonBody: [String: Any]? = nil
    ) async throws {
        let url = try makeURL(endpoint: endpoint, queryItems: queryItems)
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        if let jsonBody {
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
            request.httpBody = try JSONSerialization.data(withJSONObject: jsonBody)
        }
        _ = try await data(for: request)
    }

    private func postJSON<Response: Decodable, Body: Encodable>(
        _ endpoint: String,
        body: Body
    ) async throws -> Response {
        let url = try makeURL(endpoint: endpoint)
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(body)
        let responseData = try await data(for: request)
        do {
            return try JSONDecoder().decode(Response.self, from: responseData)
        } catch {
            throw APIError.decoding(error.localizedDescription)
        }
    }

    private func sendJSON<Body: Encodable>(
        _ endpoint: String,
        method: String = "POST",
        body: Body
    ) async throws {
        let url = try makeURL(endpoint: endpoint)
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(body)
        _ = try await data(for: request)
    }

    private func send(_ endpoint: String, method: String) async throws {
        let url = try makeURL(endpoint: endpoint)
        var request = URLRequest(url: url)
        request.httpMethod = method
        _ = try await data(for: request)
    }

    private func makeURL(
        endpoint: String,
        queryItems: [URLQueryItem] = []
    ) throws -> URL {
        let cleanEndpoint = endpoint.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        let endpointURL = baseURL.appendingPathComponent(cleanEndpoint)
        guard var components = URLComponents(url: endpointURL, resolvingAgainstBaseURL: false) else {
            throw APIError.invalidURL
        }
        components.queryItems = queryItems.isEmpty ? nil : queryItems
        guard let url = components.url else {
            throw APIError.invalidURL
        }
        return url
    }

    private func data(for request: URLRequest) async throws -> Data {
        let (data, response) = try await session.data(for: request)
        guard let response = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }
        guard (200...299).contains(response.statusCode) else {
            let message = String(data: data, encoding: .utf8)?
                .trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
            throw APIError.server(statusCode: response.statusCode, message: message)
        }
        return data
    }
}

private extension Data {
    var trimmingWhitespace: Data {
        guard let string = String(data: self, encoding: .utf8) else { return self }
        return Data(string.trimmingCharacters(in: .whitespacesAndNewlines).utf8)
    }
}
