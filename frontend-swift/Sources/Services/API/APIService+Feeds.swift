import Foundation

extension APIService {
    // MARK: - Feeds

    func fetchFeeds() async throws -> [Feed] {
        try await get("feeds")
    }

    func addFeed(_ draft: FeedDraft) async throws {
        try await post("feeds/add", jsonBody: draft.jsonBody)
    }

    func updateFeed(_ draft: FeedDraft) async throws {
        try await post("feeds/update", jsonBody: draft.jsonBody)
    }

    func deleteFeed(id: Int) async throws {
        try await post("feeds/delete", queryItems: [URLQueryItem(name: "id", value: String(id))])
    }

    func updateFeedCategory(id: Int, category: String) async throws {
        try await post("feeds/category", jsonBody: ["id": id, "category": category])
    }

    func reorderFeed(id: Int, category: String, position: Int) async throws {
        try await post("feeds/reorder", jsonBody: ["feed_id": id, "category": category, "position": position])
    }

    func refreshAllFeeds() async throws {
        try await post("refresh")
    }

    func refreshFeed(id: Int) async throws {
        try await post("feeds/refresh", queryItems: [URLQueryItem(name: "id", value: String(id))])
    }

    func fetchRefreshProgress() async throws -> RefreshProgress {
        try await get("progress")
    }

    func testIMAPConnection(_ draft: FeedDraft) async throws -> String {
        struct Response: Decodable {
            let success: Bool?
            let message: String?
            let error: String?
        }

        let body: [String: Any] = [
            "email_imap_server": draft.emailIMAPServer,
            "email_imap_port": draft.emailIMAPPort,
            "email_username": draft.emailUsername,
            "email_password": draft.emailPassword,
            "email_folder": draft.emailFolder
        ]
        let response: Response = try await postDecoding("feeds/test-imap", jsonBody: body)
        if let error = response.error, !error.isEmpty {
            throw APIError.server(statusCode: 200, message: error)
        }
        return response.message ?? t("common.connectionSuccessful")
    }

    // MARK: - Discovery

    func startDiscovery(feedID: Int) async throws {
        try await post("feeds/discover/start", jsonBody: ["feed_id": feedID])
    }

    func fetchDiscoveryProgress() async throws -> DiscoveryState {
        try await get("feeds/discover/progress")
    }

    func clearDiscovery() async throws {
        try await post("feeds/discover/clear")
    }

    func startDiscoverAll() async throws {
        try await post("feeds/discover-all/start")
    }

    func fetchDiscoverAllProgress() async throws -> DiscoveryState {
        try await get("feeds/discover-all/progress")
    }

    func clearDiscoverAll() async throws {
        try await post("feeds/discover-all/clear")
    }

    // MARK: - RSSHub

    func testRSSHubConnection() async throws -> String {
        struct Response: Decodable {
            let success: Bool?
            let message: String?
            let error: String?
        }
        let response: Response = try await postDecoding("rsshub/test-connection")
        if let error = response.error, !error.isEmpty {
            throw APIError.server(statusCode: 200, message: error)
        }
        return response.message ?? t("setting.rsshub.connectionSuccessful")
    }

    func transformRSSHubURL(_ url: String) async throws -> String {
        struct Response: Decodable {
            let url: String?
            let transformed_url: String?
        }
        let response: Response = try await postDecoding("rsshub/transform-url", jsonBody: ["url": url])
        return response.transformed_url ?? response.url ?? url
    }
}
