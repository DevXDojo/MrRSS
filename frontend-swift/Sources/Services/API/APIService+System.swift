import Foundation

extension APIService {
    // MARK: - Statistics

    func fetchStatistics(period: String, offset: Int) async throws -> StatisticsSummary {
        try await get(
            "statistics",
            queryItems: [
                URLQueryItem(name: "period", value: period),
                URLQueryItem(name: "offset", value: String(offset))
            ]
        )
    }

    func fetchAllTimeStatistics() async throws -> [String: Int] {
        try await get("statistics/all-time")
    }

    // MARK: - Maintenance

    func fetchContentCacheInfo() async throws -> ContentCacheInfo {
        try await get("articles/content-cache-info")
    }

    func cleanupArticles() async throws {
        try await post("articles/cleanup")
    }

    func cleanupContentCache() async throws {
        try await post("articles/cleanup-content")
    }

    func fetchMediaCacheInfo() async throws -> MediaCacheInfo {
        try await get("media/info")
    }

    func cleanupMediaCache() async throws {
        try await post("media/cleanup")
    }

    // MARK: - Updates

    func checkForUpdates() async throws -> UpdateInfo {
        try await get("check-updates")
    }

    // MARK: - FreshRSS

    func fetchFreshRSSStatus() async throws -> FreshRSSStatus {
        try await get("freshrss/status")
    }

    func syncFreshRSS() async throws {
        try await post("freshrss/sync")
    }

    func syncFreshRSSFeed(id: Int) async throws {
        try await post(
            "freshrss/sync-feed",
            queryItems: [URLQueryItem(name: "feed_id", value: String(id))]
        )
    }

    // MARK: - OPML

    func exportOPML() async throws -> Data {
        try await getData("opml/export")
    }

    func importOPML(data: Data, filename: String) async throws {
        _ = try await upload(
            "opml/import",
            data: data,
            contentType: filename.lowercased().hasSuffix(".json") ? "application/json" : "text/xml",
            queryItems: [URLQueryItem(name: "filename", value: filename)]
        )
    }

    // MARK: - Browser

    /// Asks the backend to open a link. It answers with the URL to open, which
    /// the client hands to the system browser.
    func openInBrowser(url: String) async throws -> String? {
        struct Response: Decodable { let redirect: String? }
        let response: Response = try await postDecoding("browser/open", jsonBody: ["url": url])
        return response.redirect
    }
}
