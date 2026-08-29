import XCTest
@testable import MrRSS

/// Checks the request each new endpoint builds, so a wrong path or body shows
/// up here rather than as an empty screen.
final class APIServiceCoverageTests: XCTestCase {
    private var service: APIService!

    override func setUp() {
        super.setUp()
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [MockURLProtocol.self]
        service = APIService(
            baseURL: URL(string: "http://127.0.0.1:1234/api")!,
            session: URLSession(configuration: configuration)
        )
    }

    override func tearDown() {
        MockURLProtocol.handler = nil
        service = nil
        super.tearDown()
    }

    private func respond(
        _ body: String,
        status: Int = 200,
        inspect: @escaping (URLRequest, URLComponents, [String: String]) -> Void = { _, _, _ in }
    ) {
        MockURLProtocol.handler = { request in
            let components = URLComponents(url: request.url!, resolvingAgainstBaseURL: false)!
            let query = Dictionary(
                uniqueKeysWithValues: (components.queryItems ?? []).map { ($0.name, $0.value ?? "") }
            )
            inspect(request, components, query)
            return (
                HTTPURLResponse(url: request.url!, statusCode: status, httpVersion: nil, headerFields: nil)!,
                Data(body.utf8)
            )
        }
    }

    func testUpdateFeedPostsEveryConfiguredField() async throws {
        var draft = FeedDraft(id: 12, url: "https://example.com/feed", title: "Example", category: "News")
        draft.hideFromTimeline = true
        draft.refreshInterval = 45
        draft.type = "HTML+XPath"
        draft.xPathItem = "//article"
        draft.tags = [3, 4]

        var body: [String: Any] = [:]
        respond("{}") { request, components, _ in
            XCTAssertEqual(components.path, "/api/feeds/update")
            XCTAssertEqual(request.httpMethod, "POST")
            let data = request.bodyData ?? Data()
            body = (try? JSONSerialization.jsonObject(with: data)) as? [String: Any] ?? [:]
        }

        try await service.updateFeed(draft)

        XCTAssertEqual(body["id"] as? Int, 12)
        XCTAssertEqual(body["title"] as? String, "Example")
        XCTAssertEqual(body["hide_from_timeline"] as? Bool, true)
        XCTAssertEqual(body["refresh_interval"] as? Int, 45)
        XCTAssertEqual(body["type"] as? String, "HTML+XPath")
        XCTAssertEqual(body["xpath_item"] as? String, "//article")
        XCTAssertEqual(body["tags"] as? [Int], [3, 4])
    }

    func testMarkAllReadPassesTheCategory() async throws {
        respond("{}") { _, components, query in
            XCTAssertEqual(components.path, "/api/articles/mark-all-read")
            XCTAssertEqual(query["category"], "Tech")
            XCTAssertNil(query["feed_id"])
        }

        try await service.markAllRead(feedID: nil, category: "Tech")
    }

    func testMarkRelativeReadsTheAffectedCount() async throws {
        respond(#"{"count": 7}"#) { _, components, query in
            XCTAssertEqual(components.path, "/api/articles/mark-relative")
            XCTAssertEqual(query["direction"], "above")
            XCTAssertEqual(query["id"], "42")
        }

        let count = try await service.markRelative(id: 42, direction: "above")

        XCTAssertEqual(count, 7)
    }

    func testFilterArticlesDecodesAPage() async throws {
        let payload = """
        {
          "articles": [
            {"id": 1, "feed_id": 2, "title": "One", "url": "https://example.com/1", "published_at": "2026-08-16T08:00:00Z"}
          ],
          "total": 1,
          "page": 1,
          "limit": 50,
          "has_more": false
        }
        """
        respond(payload) { _, components, _ in
            XCTAssertEqual(components.path, "/api/articles/filter")
        }

        let page = try await service.filterArticles(conditions: [FilterCondition()], page: 1, limit: 50)

        XCTAssertEqual(page.total, 1)
        XCTAssertFalse(page.hasMore)
        XCTAssertEqual(page.articles.first?.title, "One")
    }

    func testSavedFilterUpdateUsesTheFilterEndpoint() async throws {
        var body: [String: Any] = [:]
        respond("{}") { request, components, query in
            XCTAssertEqual(components.path, "/api/saved-filters/filter")
            XCTAssertEqual(request.httpMethod, "PUT")
            XCTAssertEqual(query["id"], "5")
            body = (try? JSONSerialization.jsonObject(with: request.bodyData ?? Data())) as? [String: Any] ?? [:]
        }

        let filter = SavedFilter(id: 5, name: "Weekly", conditions: [FilterCondition(value: "swift")])
        try await service.updateSavedFilter(filter)

        XCTAssertEqual(body["name"] as? String, "Weekly")
        let conditions = body["conditions"] as? String ?? ""
        XCTAssertTrue(conditions.contains("swift"), "conditions should be sent as encoded JSON")
    }

    func testFilterCountsDecodeIntegerKeyedMaps() async throws {
        respond(#"{"unread":{"1":3},"favorites":{"2":1},"images_unread":{"3":9}}"#)

        let counts = try await service.fetchFilterCounts()

        XCTAssertEqual(counts.unread[1], 3)
        XCTAssertEqual(counts.favorites[2], 1)
        XCTAssertEqual(counts.imagesUnread[3], 9)
    }

    func testUpdateInfoDecodesTheDownloadDetails() async throws {
        let payload = """
        {
          "current_version": "1.3.28",
          "latest_version": "1.4.0",
          "has_update": true,
          "platform": "darwin",
          "arch": "arm64",
          "is_portable": false,
          "download_url": "https://example.com/app.dmg",
          "asset_name": "app.dmg",
          "asset_size": 1024
        }
        """
        respond(payload)

        let info = try await service.checkForUpdates()

        XCTAssertTrue(info.hasUpdate)
        XCTAssertEqual(info.latestVersion, "1.4.0")
        XCTAssertEqual(info.assetSize, 1024)
    }

    func testAISearchDecodesHitsWithMatchDetail() async throws {
        let payload = """
        {
          "success": true,
          "total_count": 1,
          "search_terms": "swift concurrency",
          "articles": [
            {
              "id": 8,
              "feed_id": 2,
              "title": "Concurrency",
              "url": "https://example.com/8",
              "published_at": "2026-08-16T08:00:00Z",
              "relevance_score": 0.75,
              "matched_terms": ["swift"],
              "matched_fields": ["title"],
              "excerpt": "…structured concurrency…"
            }
          ]
        }
        """
        respond(payload) { _, components, _ in
            XCTAssertEqual(components.path, "/api/ai/search")
        }

        let response = try await service.aiSearch(query: "swift concurrency")

        XCTAssertTrue(response.success)
        XCTAssertEqual(response.articles.count, 1)
        XCTAssertEqual(response.articles.first?.article.title, "Concurrency")
        XCTAssertEqual(response.articles.first?.matchedTerms, ["swift"])
    }

    func testErrorBodyPrefersTheErrorField() async {
        respond(#"{"error":"AI limit reached"}"#, status: 403)

        do {
            _ = try await service.fetchFeeds()
            XCTFail("Expected the request to fail")
        } catch let error as APIError {
            XCTAssertEqual(error, .server(statusCode: 403, message: "AI limit reached"))
        } catch {
            XCTFail("Unexpected error: \(error)")
        }
    }

    func testOPMLImportSendsTheFilenameHint() async throws {
        respond("") { request, components, query in
            XCTAssertEqual(components.path, "/api/opml/import")
            XCTAssertEqual(query["filename"], "subscriptions.opml")
            XCTAssertEqual(request.value(forHTTPHeaderField: "Content-Type"), "text/xml")
        }

        try await service.importOPML(data: Data("<opml/>".utf8), filename: "subscriptions.opml")
    }
}

private extension URLRequest {
    /// `URLProtocol` moves an uploaded body into a stream, so read whichever is set.
    var bodyData: Data? {
        if let httpBody { return httpBody }
        guard let stream = httpBodyStream else { return nil }
        stream.open()
        defer { stream.close() }
        var data = Data()
        let bufferSize = 4_096
        var buffer = [UInt8](repeating: 0, count: bufferSize)
        while stream.hasBytesAvailable {
            let read = stream.read(&buffer, maxLength: bufferSize)
            if read <= 0 { break }
            data.append(buffer, count: read)
        }
        return data
    }
}
