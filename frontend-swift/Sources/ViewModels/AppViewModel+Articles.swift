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

    /// Marks everything above or below one article as read.
    func markRelative(to article: Article, direction: MarkDirection) async {
        do {
            let count = try await api.markRelative(id: article.id, direction: direction.rawValue)
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
            let target = (try? await api.openInBrowser(url: article.url)) ?? article.url
            guard let url = URL(string: target ?? article.url) else { return }
            NSWorkspace.shared.open(url)
        }
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

    private func applyRelativeReadState(from article: Article, direction: MarkDirection) {
        let ordered = displayedArticles
        guard let pivot = ordered.firstIndex(where: { $0.id == article.id }) else { return }
        let affected = direction == .above ? ordered.prefix(pivot) : ordered.suffix(from: pivot + 1)
        let affectedIDs = Set(affected.map(\.id))
        for index in articles.indices where affectedIDs.contains(articles[index].id) {
            articles[index].isRead = true
        }
    }
}

/// Which side of an article a bulk "mark as read" applies to.
enum MarkDirection: String {
    case above
    case below
}
