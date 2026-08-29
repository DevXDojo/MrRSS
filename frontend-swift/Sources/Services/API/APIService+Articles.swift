import Foundation

extension APIService {
    // MARK: - Listing

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

    func fetchImageArticles(page: Int, limit: Int) async throws -> [Article] {
        try await get(
            "articles/images",
            queryItems: [
                URLQueryItem(name: "page", value: String(page)),
                URLQueryItem(name: "limit", value: String(limit))
            ]
        )
    }

    func filterArticles(
        conditions: [FilterCondition],
        page: Int,
        limit: Int
    ) async throws -> FilteredArticles {
        struct Request: Encodable {
            let conditions: [FilterCondition]
            let page: Int
            let limit: Int
        }
        return try await postJSON(
            "articles/filter",
            body: Request(conditions: conditions, page: page, limit: limit)
        )
    }

    // MARK: - State changes

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
        try await post("articles/favorite", queryItems: [URLQueryItem(name: "id", value: String(id))])
    }

    func toggleReadLater(id: Int) async throws {
        try await post("articles/toggle-read-later", queryItems: [URLQueryItem(name: "id", value: String(id))])
    }

    func toggleHidden(id: Int) async throws {
        try await post("articles/toggle-hide", queryItems: [URLQueryItem(name: "id", value: String(id))])
    }

    /// Marks every article published before or after the given one as read, and
    /// returns how many were changed. The feed or category scopes the change to
    /// what the reader is currently looking at.
    func markRelative(
        id: Int,
        direction: String,
        feedID: Int?,
        category: String?
    ) async throws -> Int {
        struct Response: Decodable {
            let count: Int?
            let marked: Int?
            let affected: Int?
        }

        var queryItems = [
            URLQueryItem(name: "id", value: String(id)),
            URLQueryItem(name: "direction", value: direction)
        ]
        if let feedID {
            queryItems.append(URLQueryItem(name: "feed_id", value: String(feedID)))
        } else if let category, !category.isEmpty {
            queryItems.append(URLQueryItem(name: "category", value: category))
        }

        let response: Response = try await postDecoding(
            "articles/mark-relative",
            queryItems: queryItems
        )
        return response.count ?? response.marked ?? response.affected ?? 0
    }

    func markAllRead(feedID: Int?, category: String?) async throws {
        var queryItems: [URLQueryItem] = []
        if let feedID {
            queryItems.append(URLQueryItem(name: "feed_id", value: String(feedID)))
        }
        if let category, !category.isEmpty {
            queryItems.append(URLQueryItem(name: "category", value: category))
        }
        try await post("articles/mark-all-read", queryItems: queryItems)
    }

    func clearReadLater() async throws {
        try await post("articles/clear-read-later")
    }

    // MARK: - Content

    func fetchArticleContent(id: Int) async throws -> ArticleContent {
        try await get("articles/content", queryItems: [URLQueryItem(name: "id", value: String(id))])
    }

    func reloadArticleContent(id: Int) async throws -> ArticleContent {
        try await postDecoding(
            "articles/reload-content",
            queryItems: [URLQueryItem(name: "id", value: String(id))]
        )
    }

    func fetchFullArticle(id: Int) async throws -> ArticleContent {
        try await postDecoding(
            "articles/fetch-full",
            queryItems: [URLQueryItem(name: "id", value: String(id))]
        )
    }

    func extractImages(id: Int) async throws -> ArticleImages {
        try await postDecoding(
            "articles/extract-images",
            queryItems: [URLQueryItem(name: "id", value: String(id))]
        )
    }

    // MARK: - Counts

    func fetchUnreadCounts() async throws -> UnreadCounts {
        try await get("articles/unread-counts")
    }

    func fetchFilterCounts() async throws -> FilterCounts {
        try await get("articles/filter-counts")
    }

    // MARK: - Export

    func exportArticle(id: Int, destination: ArticleExportDestination) async throws -> String {
        struct Response: Decodable {
            let success: Bool?
            let message: String?
            let error: String?
            let path: String?
        }
        let response: Response = try await postDecoding(
            destination.endpoint,
            jsonBody: ["article_id": id]
        )
        if let error = response.error, !error.isEmpty {
            throw APIError.server(statusCode: 200, message: error)
        }
        return response.message ?? response.path ?? destination.localizedSuccess
    }

    // MARK: - Translation and summaries

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
}
