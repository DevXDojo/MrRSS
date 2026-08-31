import Foundation

extension APIService {
    // MARK: - Usage

    func fetchAIUsage() async throws -> AIUsage {
        try await get("ai-usage")
    }

    func resetAIUsage() async throws {
        try await post("ai-usage/reset")
    }

    // MARK: - Profiles

    func fetchAIProfiles() async throws -> [AIProfile] {
        try await get("ai/profiles")
    }

    func saveAIProfile(_ profile: AIProfile) async throws -> AIProfile {
        var body: [String: Any] = [
            "name": profile.name,
            "endpoint": profile.endpoint,
            "model": profile.model,
            "custom_headers": profile.customHeaders,
            "is_default": profile.isDefault
        ]
        if !profile.apiKey.isEmpty {
            body["api_key"] = profile.apiKey
        }

        if profile.id > 0 {
            body["id"] = profile.id
            let data = try await send("ai/profiles/\(profile.id)", method: "PUT", jsonBody: body)
            return (try? decode(data) as AIProfile) ?? profile
        }
        return try await postDecoding("ai/profiles", jsonBody: body)
    }

    func deleteAIProfile(id: Int) async throws {
        try await send("ai/profiles/\(id)", method: "DELETE")
    }

    func setDefaultAIProfile(id: Int) async throws {
        try await post("ai/profiles/\(id)/default")
    }

    func testAIProfiles() async throws -> [AIProfileTestResult] {
        try await postDecoding("ai/profiles/test-all")
    }

    // MARK: - Search

    func aiSearch(query: String) async throws -> AISearchResponse {
        struct Request: Encodable { let query: String }
        return try await postJSON("ai/search", body: Request(query: query))
    }

    // MARK: - Chat

    func sendChatMessage(_ request: ChatRequest) async throws -> ChatResponse {
        try await postJSON("ai-chat", body: request)
    }

    func fetchChatSessions(articleID: Int) async throws -> [ChatSession] {
        try await get(
            "ai/chat/sessions",
            queryItems: [URLQueryItem(name: "article_id", value: String(articleID))]
        )
    }

    func createChatSession(articleID: Int, title: String) async throws -> ChatSession {
        try await postDecoding(
            "ai/chat/session/create",
            jsonBody: ["article_id": articleID, "title": title]
        )
    }

    func fetchChatMessages(sessionID: Int) async throws -> [ChatMessage] {
        try await get(
            "ai/chat/messages",
            queryItems: [URLQueryItem(name: "session_id", value: String(sessionID))]
        )
    }

    func deleteChatSession(id: Int) async throws {
        try await send(
            "ai/chat/session",
            method: "DELETE",
            queryItems: [URLQueryItem(name: "session_id", value: String(id))]
        )
    }

    /// Deletes every stored conversation. The backend clears all sessions at
    /// once rather than only those for one article.
    func deleteAllChatSessions() async throws {
        try await send("ai/chat/sessions/delete-all", method: "DELETE")
    }
}
