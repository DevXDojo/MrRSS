import XCTest
@testable import MrRSS

final class APIServiceTests: XCTestCase {
    override func tearDown() {
        MockURLProtocol.handler = nil
        super.tearDown()
    }

    func testFetchArticlesSendsAllFilterAndPagination() async throws {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [MockURLProtocol.self]
        let session = URLSession(configuration: configuration)
        let service = APIService(baseURL: URL(string: "http://127.0.0.1:1234/api")!, session: session)

        MockURLProtocol.handler = { request in
            let components = try XCTUnwrap(URLComponents(url: request.url!, resolvingAgainstBaseURL: false))
            let query = Dictionary(uniqueKeysWithValues: (components.queryItems ?? []).map { ($0.name, $0.value ?? "") })
            XCTAssertEqual(components.path, "/api/articles")
            XCTAssertEqual(query["filter"], "all")
            XCTAssertEqual(query["page"], "2")
            XCTAssertEqual(query["limit"], "25")
            return (HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!, Data("[]".utf8))
        }

        let articles = try await service.fetchArticles(
            feedID: nil,
            category: nil,
            filter: "all",
            page: 2,
            limit: 25
        )

        XCTAssertTrue(articles.isEmpty)
    }

    func testServerErrorIncludesResponseMessage() async {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [MockURLProtocol.self]
        let session = URLSession(configuration: configuration)
        let service = APIService(baseURL: URL(string: "http://127.0.0.1:1234/api")!, session: session)

        MockURLProtocol.handler = { request in
            return (
                HTTPURLResponse(url: request.url!, statusCode: 503, httpVersion: nil, headerFields: nil)!,
                Data("database unavailable".utf8)
            )
        }

        do {
            _ = try await service.fetchFeeds()
            XCTFail("Expected the request to fail")
        } catch let error as APIError {
            XCTAssertEqual(error, .server(statusCode: 503, message: "database unavailable"))
        } catch {
            XCTFail("Unexpected error: \(error)")
        }
    }

    func testServerAddressNormalizationAddsSchemeAndAPIPath() {
        let url = ServerConfiguration.normalizedURL(from: "127.0.0.1:1234")
        XCTAssertEqual(url?.absoluteString, "http://127.0.0.1:1234/api")
    }

    func testUpdateSettingsSendsCompleteJSONMap() async throws {
        let service = makeService()
        MockURLProtocol.handler = { request in
            XCTAssertEqual(request.httpMethod, "POST")
            XCTAssertEqual(request.url?.path, "/api/settings")
            let body = try Self.requestBody(from: request)
            let values = try XCTUnwrap(JSONSerialization.jsonObject(with: body) as? [String: String])
            XCTAssertEqual(values["theme"], "dark")
            XCTAssertEqual(values["translation_enabled"], "true")
            return Self.response(for: request, data: Data())
        }

        try await service.updateSettings(["theme": "dark", "translation_enabled": "true"])
    }

    func testTranslationSummaryAndRuleResponsesDecode() async throws {
        let service = makeService()
        MockURLProtocol.handler = { request in
            switch request.url?.path {
            case "/api/articles/translate":
                return Self.response(
                    for: request,
                    data: Data(#"{"translated_title":"译文","limit_reached":false}"#.utf8)
                )
            case "/api/articles/summarize":
                return Self.response(
                    for: request,
                    data: Data(#"{"summary":"摘要","is_too_short":false}"#.utf8)
                )
            case "/api/rules/apply":
                return Self.response(
                    for: request,
                    data: Data(#"{"success":true,"affected":4}"#.utf8)
                )
            default:
                XCTFail("Unexpected endpoint: \(request.url?.path ?? "nil")")
                return Self.response(for: request, statusCode: 404, data: Data())
            }
        }

        let translation = try await service.translateTitle(articleID: 1, title: "Text", targetLanguage: "zh")
        let summary = try await service.summarize(articleID: 1, length: "medium", content: "Body")
        let ruleResult = try await service.applyRule(
            AutomationRule(id: 1, name: "Test", enabled: true, conditions: [], actions: ["favorite"])
        )

        XCTAssertEqual(translation.translatedTitle, "译文")
        XCTAssertEqual(summary.summary, "摘要")
        XCTAssertNil(summary.sentenceCount)
        XCTAssertEqual(ruleResult.affected, 4)
    }

    private func makeService() -> APIService {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [MockURLProtocol.self]
        return APIService(
            baseURL: URL(string: "http://127.0.0.1:1234/api")!,
            session: URLSession(configuration: configuration)
        )
    }

    private static func response(
        for request: URLRequest,
        statusCode: Int = 200,
        data: Data
    ) -> (HTTPURLResponse, Data) {
        (
            HTTPURLResponse(url: request.url!, statusCode: statusCode, httpVersion: nil, headerFields: nil)!,
            data
        )
    }

    private static func requestBody(from request: URLRequest) throws -> Data {
        if let body = request.httpBody { return body }
        let stream = try XCTUnwrap(request.httpBodyStream)
        stream.open()
        defer { stream.close() }

        var data = Data()
        var buffer = [UInt8](repeating: 0, count: 1_024)
        while stream.hasBytesAvailable {
            let count = stream.read(&buffer, maxLength: buffer.count)
            if count < 0 { throw stream.streamError ?? APIError.invalidResponse }
            if count == 0 { break }
            data.append(buffer, count: count)
        }
        return data
    }
}

private final class MockURLProtocol: URLProtocol {
    static var handler: ((URLRequest) throws -> (HTTPURLResponse, Data))?

    override class func canInit(with request: URLRequest) -> Bool { true }
    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }

    override func startLoading() {
        guard let handler = Self.handler else {
            XCTFail("MockURLProtocol handler was not configured")
            return
        }

        do {
            let (response, data) = try handler(request)
            client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
            client?.urlProtocol(self, didLoad: data)
            client?.urlProtocolDidFinishLoading(self)
        } catch {
            client?.urlProtocol(self, didFailWithError: error)
        }
    }

    override func stopLoading() {}
}
