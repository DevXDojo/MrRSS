import AppKit
import Foundation

extension AppViewModel {
    // MARK: - Presentation

    /// The list as it should be shown, after the chosen ordering is applied.
    var displayedArticles: [Article] {
        switch sortOrder {
        case .newestFirst:
            return articles.sorted { lhs, rhs in
                (lhs.publishedDate ?? .distantPast) > (rhs.publishedDate ?? .distantPast)
            }
        case .oldestFirst:
            return articles.sorted { lhs, rhs in
                (lhs.publishedDate ?? .distantPast) < (rhs.publishedDate ?? .distantPast)
            }
        case .unreadFirst:
            return articles.sorted { lhs, rhs in
                if lhs.isRead != rhs.isRead { return !lhs.isRead }
                return (lhs.publishedDate ?? .distantPast) > (rhs.publishedDate ?? .distantPast)
            }
        case .byTitle:
            return articles.sorted { lhs, rhs in
                lhs.title.localizedStandardCompare(rhs.title) == .orderedAscending
            }
        }
    }

    /// The heading shown above the article list.
    var articleListTitle: String {
        switch selection {
        case .filter(let filter):
            return filter.title
        case .folder(let name):
            return name
        case .feed(let id):
            return feeds.first(where: { $0.id == id })?.title ?? t("sidebar.feedList.articles")
        case .savedFilter(let id):
            return savedFilters.first(where: { $0.id == id })?.name ?? t("modal.filter.filter")
        case .none:
            return t("sidebar.feedList.articles")
        }
    }

    /// The feed a given article came from, used for the list subtitle and for
    /// per-feed reading preferences.
    func feed(for article: Article) -> Feed? {
        feeds.first(where: { $0.id == article.feedID })
    }

    /// The badge to show next to one feed, honouring the current activity.
    func badgeCount(for feedID: Int, activity: ArticleFilter) -> Int {
        let keyPath = showOnlyUnread ? activity.unreadCountsKeyPath : activity.countsKeyPath
        guard let keyPath else { return unreadCounts.feedCounts[feedID] ?? 0 }
        return filterCounts[keyPath: keyPath][feedID] ?? 0
    }

    /// The total for one activity across every feed.
    func totalCount(for activity: ArticleFilter) -> Int {
        switch activity {
        case .all: return unreadCounts.total
        case .unread: return unreadCounts.total
        case .favorites: return filterCounts.total(for: \.favorites)
        case .readLater: return filterCounts.total(for: \.readLater)
        case .imageGallery: return filterCounts.total(for: \.images)
        }
    }

    /// The tags assigned to one feed.
    func tags(forFeed feedID: Int) -> [Tag] {
        let ids = Set(feedTags[feedID] ?? [])
        return tags.filter { ids.contains($0.id) }
    }

    // MARK: - Article actions

    func toggleReadLater(_ article: Article) {
        mutateArticle(article.id, apply: { $0.isReadLater.toggle() }) { [weak self] in
            try await self?.api.toggleReadLater(id: article.id)
        }
    }

    func toggleHidden(_ article: Article) {
        mutateArticle(article.id, apply: { $0.isHidden.toggle() }) { [weak self] in
            try await self?.api.toggleHidden(id: article.id)
        }

        // A hidden article leaves the list unless hidden articles are shown.
        if !boolSetting("show_hidden_articles"), let index = articles.firstIndex(where: { $0.id == article.id }),
           articles[index].isHidden {
            articles.remove(at: index)
            if selectedArticleID == article.id {
                selectedArticleID = nil
            }
        }
    }

    /// Marks everything in one feed or one folder as read.
    func markRead(feedID: Int? = nil, category: String? = nil) async {
        do {
            try await api.markAllRead(feedID: feedID, category: category)
            statusMessage = t("article.action.markedAllAsRead")
            if let feedID {
                for index in articles.indices where articles[index].feedID == feedID {
                    articles[index].isRead = true
                }
            } else if category != nil {
                reloadArticles()
            }
            await refreshCounts()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    /// Marks everything in the current activity as read.
    func markAllRead() async {
        let query = articleQuery
        do {
            if query.usesConditions || query.usesImageGallery {
                // These listings have no server-side bulk action, so the visible
                // articles are marked one by one.
                let unread = articles.filter { !$0.isRead }
                guard !unread.isEmpty else {
                    statusMessage = t("article.action.noArticlesToMark")
                    return
                }
                for article in unread {
                    try await api.setArticleRead(id: article.id, read: true)
                }
                statusMessage = t("article.action.markedNArticlesAsRead", ["count": unread.count])
            } else {
                try await api.markAllRead(feedID: query.feedID, category: query.category)
                statusMessage = t("article.action.markedAllAsRead")
            }
            for index in articles.indices {
                articles[index].isRead = true
            }
            await refreshCounts()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    /// Marks everything published before or after one article as read, scoped to
    /// the feed or folder currently being read.
    func markRelative(to article: Article, direction: MarkDirection) async {
        let query = articleQuery
        do {
            let count = try await api.markRelative(
                id: article.id,
                direction: direction.rawValue,
                feedID: query.feedID,
                category: query.category
            )
            statusMessage = t("article.action.markedNArticlesAsRead", ["count": count])
            applyRelativeReadState(from: article, direction: direction)
            await refreshCounts()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func clearReadLater() async {
        do {
            try await api.clearReadLater()
            statusMessage = t("common.toast.clearedReadLater")
            reloadArticles()
            await refreshCounts()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func exportArticle(_ article: Article, to destination: ArticleExportDestination) async {
        statusMessage = destination.localizedProgress
        do {
            statusMessage = try await api.exportArticle(id: article.id, destination: destination)
        } catch {
            errorMessage = "\(destination.localizedFailure): \(error.localizedDescription)"
        }
    }

    /// Opens the article in the system browser. The backend is asked first so
    /// its own link handling still applies.
    func openInBrowser(_ article: Article) {
        guard !article.url.isEmpty else { return }
        Task { [weak self] in
            guard let self else { return }
            // The backend answers with the address to open, and falls back to
            // the article's own when it has nothing to add.
            let redirect = try? await api.openInBrowser(url: article.url)
            let target = redirect.flatMap { $0 } ?? article.url
            guard let url = URL(string: target) else { return }
            NSWorkspace.shared.open(url)
        }
    }

    // MARK: - Article content

    func reloadArticleContent(id: Int) async throws -> ArticleContent {
        try await api.reloadArticleContent(id: id)
    }

    func fetchFullArticle(id: Int) async throws -> ArticleContent {
        try await api.fetchFullArticle(id: id)
    }

    /// The images the backend can pull out of one article.
    func articleImages(id: Int) async -> [String] {
        ((try? await api.extractImages(id: id))?.images ?? []).filter { !$0.isEmpty }
    }

    // MARK: - AI search

    /// Runs the AI-assisted search and shows the hits in place of the list.
    func runAISearch(_ query: String) async {
        let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else {
            clearAISearch()
            return
        }

        isSearching = true
        defer { isSearching = false }

        do {
            let response = try await api.aiSearch(query: trimmed)
            if !response.success, let error = response.error {
                errorMessage = error
                return
            }
            searchHits = response.articles
            articles = response.articles.map(\.article)
            searchTerms = response.searchTerms
            statusMessage = t("aiSearch.foundResults", ["count": response.totalCount])
        } catch {
            errorMessage = "\(t("aiSearch.searchFailed")) \(error.localizedDescription)"
        }
    }

    /// Puts the ordinary list back.
    func clearAISearch() {
        guard !searchHits.isEmpty || searchTerms != nil else { return }
        searchHits = []
        searchTerms = nil
        reloadArticles()
    }

    /// Why one article matched the search, when it came from one.
    func searchHit(for articleID: Int) -> AISearchHit? {
        searchHits.first { $0.id == articleID }
    }

    // MARK: - Loading side data

    func refreshCounts() async {
        async let unread = try? await api.fetchUnreadCounts()
        async let filters = try? await api.fetchFilterCounts()
        let (loadedUnread, loadedFilters) = await (unread, filters)
        if let loadedUnread { unreadCounts = loadedUnread }
        if let loadedFilters { filterCounts = loadedFilters }
    }

    func loadSavedFilters() async {
        savedFilters = (try? await api.fetchSavedFilters())?
            .sorted { $0.position < $1.position } ?? []
    }

    func loadTags() async {
        tags = (try? await api.fetchTags())?.sorted { $0.position < $1.position } ?? []
        var assignments: [Int: [Int]] = [:]
        for feed in feeds {
            if let ids = try? await api.fetchFeedTags(feedID: feed.id), !ids.isEmpty {
                assignments[feed.id] = ids
            }
        }
        feedTags = assignments
    }

    // MARK: - Helpers

    /// Applies a change straight away and rolls it back if the server rejects it.
    private func mutateArticle(
        _ id: Int,
        apply change: (inout Article) -> Void,
        request: @escaping () async throws -> Void
    ) {
        guard let index = articles.firstIndex(where: { $0.id == id }) else { return }
        let previous = articles[index]
        change(&articles[index])

        Task { [weak self] in
            guard let self else { return }
            do {
                try await request()
            } catch {
                if let currentIndex = articles.firstIndex(where: { $0.id == id }) {
                    articles[currentIndex] = previous
                }
                errorMessage = error.localizedDescription
            }
        }
    }

    /// Mirrors what the server just did. It works on publication time rather
    /// than on the order the list happens to be in, so the two agree however
    /// the reader has sorted the list.
    private func applyRelativeReadState(from article: Article, direction: MarkDirection) {
        guard let pivot = article.publishedDate else { return }
        for index in articles.indices {
            guard let published = articles[index].publishedDate else { continue }
            let isAffected = direction == .above ? published > pivot : published < pivot
            if isAffected {
                articles[index].isRead = true
            }
        }
    }
}

/// Which side of an article a bulk "mark as read" applies to.
enum MarkDirection: String {
    case above
    case below
}
