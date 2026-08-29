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
    case notStubbed(String)

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return t("client.error.invalidServerAddress")
        case .invalidResponse:
            return t("client.error.invalidResponse")
        case .server(let statusCode, let message):
            return message.isEmpty
                ? t("client.error.httpStatus", ["status": statusCode])
                : "\(t("client.error.httpStatus", ["status": statusCode])): \(message)"
        case .decoding(let message):
            return "\(t("client.error.unreadableResponse")): \(message)"
        case .notStubbed(let name):
            return "The test stub does not implement \(name)."
        }
    }
}

/// Talks to the MrRSS HTTP API. Domain-specific calls live in the
/// `APIService+…` extensions next to this file.
final class APIService: APIClient {
    static let shared = APIService()

    private(set) var baseURL: URL
    let session: URLSession

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

    // MARK: - Transport

    func get<T: Decodable>(_ endpoint: String, queryItems: [URLQueryItem] = []) async throws -> T {
        let url = try makeURL(endpoint: endpoint, queryItems: queryItems)
        let data = try await data(for: URLRequest(url: url))
        return try decode(data)
    }

    func getData(_ endpoint: String, queryItems: [URLQueryItem] = []) async throws -> Data {
        let url = try makeURL(endpoint: endpoint, queryItems: queryItems)
        return try await data(for: URLRequest(url: url))
    }

    @discardableResult
    func post(
        _ endpoint: String,
        queryItems: [URLQueryItem] = [],
        jsonBody: [String: Any]? = nil
    ) async throws -> Data {
        let url = try makeURL(endpoint: endpoint, queryItems: queryItems)
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        if let jsonBody {
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
            request.httpBody = try JSONSerialization.data(withJSONObject: jsonBody)
        }
        return try await data(for: request)
    }

    func postDecoding<T: Decodable>(
        _ endpoint: String,
        queryItems: [URLQueryItem] = [],
        jsonBody: [String: Any]? = nil
    ) async throws -> T {
        let data = try await post(endpoint, queryItems: queryItems, jsonBody: jsonBody)
        return try decode(data)
    }

    func postJSON<Response: Decodable, Body: Encodable>(
        _ endpoint: String,
        body: Body,
        method: String = "POST"
    ) async throws -> Response {
        let responseData = try await sendJSONReturningData(endpoint, body: body, method: method)
        return try decode(responseData)
    }

    @discardableResult
    func sendJSONReturningData<Body: Encodable>(
        _ endpoint: String,
        body: Body,
        method: String = "POST"
    ) async throws -> Data {
        let url = try makeURL(endpoint: endpoint)
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(body)
        return try await data(for: request)
    }

    @discardableResult
    func send(
        _ endpoint: String,
        method: String,
        queryItems: [URLQueryItem] = [],
        jsonBody: [String: Any]? = nil
    ) async throws -> Data {
        let url = try makeURL(endpoint: endpoint, queryItems: queryItems)
        var request = URLRequest(url: url)
        request.httpMethod = method
        if let jsonBody {
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
            request.httpBody = try JSONSerialization.data(withJSONObject: jsonBody)
        }
        return try await data(for: request)
    }

    func upload(
        _ endpoint: String,
        data payload: Data,
        contentType: String,
        queryItems: [URLQueryItem] = []
    ) async throws -> Data {
        let url = try makeURL(endpoint: endpoint, queryItems: queryItems)
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue(contentType, forHTTPHeaderField: "Content-Type")
        request.httpBody = payload
        return try await data(for: request)
    }

    /// Decodes a response, treating a literal `null` body as an empty collection
    /// because several endpoints return `null` instead of `[]`.
    func decode<T: Decodable>(_ data: Data) throws -> T {
        if data.trimmingWhitespace == Data("null".utf8), let emptyArray = [] as? T {
            return emptyArray
        }
        do {
            return try JSONDecoder().decode(T.self, from: data)
        } catch {
            throw APIError.decoding(error.localizedDescription)
        }
    }

    func makeURL(endpoint: String, queryItems: [URLQueryItem] = []) throws -> URL {
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

    @discardableResult
    func data(for request: URLRequest) async throws -> Data {
        let (data, response) = try await session.data(for: request)
        guard let response = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }
        guard (200...299).contains(response.statusCode) else {
            let message = APIService.errorMessage(from: data)
            throw APIError.server(statusCode: response.statusCode, message: message)
        }
        return data
    }

    /// Pulls the human-readable part out of an error body, which the backend
    /// sends either as `{"error": "…"}` or as plain text.
    static func errorMessage(from data: Data) -> String {
        if let object = try? JSONSerialization.jsonObject(with: data) as? [String: Any] {
            for key in ["error", "message", "detail"] {
                if let value = object[key] as? String, !value.isEmpty {
                    return value
                }
            }
        }
        return String(data: data, encoding: .utf8)?
            .trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
    }

    // MARK: - Connection

    func checkConnection() async throws {
        _ = try await getData("version")
    }

    func fetchVersion() async throws -> String {
        struct Response: Decodable { let version: String }
        let response: Response = try await get("version")
        return response.version
    }
}

extension Data {
    var trimmingWhitespace: Data {
        guard let string = String(data: self, encoding: .utf8) else { return self }
        return Data(string.trimmingCharacters(in: .whitespacesAndNewlines).utf8)
    }
}
