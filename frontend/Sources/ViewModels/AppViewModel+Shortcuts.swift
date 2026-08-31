import AppKit
import Foundation

extension AppViewModel {
    /// The bindings currently in force, taken from the stored settings.
    var shortcutTable: KeyboardShortcutTable {
        KeyboardShortcutTable(settings: settings)
    }

    /// Runs whatever the key press means. Returns false when nothing matched so
    /// the event can continue on its way.
    @discardableResult
    func handleKeyPress(_ event: NSEvent) -> Bool {
        let table = shortcutTable
        guard table.isEnabled, let action = table.action(for: event) else { return false }
        perform(action)
        return true
    }

    func perform(_ action: ShortcutAction) {
        switch action {
        case .nextArticle:
            selectRelativeArticle(offset: 1)
        case .previousArticle:
            selectRelativeArticle(offset: -1)
        case .toggleReadStatus:
            guard let article = currentArticle else { return }
            setArticleRead(article, read: !article.isRead)
        case .toggleFavoriteStatus:
            guard let article = currentArticle else { return }
            toggleFavorite(article)
        case .toggleReadLaterStatus:
            guard let article = currentArticle else { return }
            toggleReadLater(article)
        case .openInBrowser:
            guard let article = currentArticle else { return }
            openInBrowser(article)
        case .toggleContentView:
            requestedViewModeToggle += 1
        case .refreshFeeds:
            refreshFromSources()
        case .markAllRead:
            Task { await markAllRead() }
        case .addFeed:
            isPresentingAddFeed = true
        case .toggleUnreadFilter:
            showOnlyUnread.toggle()
        case .toggleFavoritesFilter:
            selection = selection == .filter(.favorites) ? .filter(.all) : .filter(.favorites)
        case .toggleReadLaterFilter:
            selection = selection == .filter(.readLater) ? .filter(.all) : .filter(.readLater)
        case .goToAllArticles:
            selection = .filter(.all)
        case .goToUnread:
            selection = .filter(.unread)
        case .goToFavorites:
            selection = .filter(.favorites)
        case .goToReadLater:
            selection = .filter(.readLater)
        }
    }

    /// True when there is an article before the current one.
    var hasPreviousArticle: Bool {
        guard let current = selectedArticleID,
              let index = displayedArticles.firstIndex(where: { $0.id == current }) else {
            return false
        }
        return index > 0
    }

    /// True when there is an article after the current one, or more to load.
    var hasNextArticle: Bool {
        guard let current = selectedArticleID,
              let index = displayedArticles.firstIndex(where: { $0.id == current }) else {
            return !displayedArticles.isEmpty
        }
        return index + 1 < displayedArticles.count || hasMoreArticles
    }

    /// The article the reading pane is showing.
    var currentArticle: Article? {
        article(withID: selectedArticleID)
    }

    /// Moves the selection through the list as it is currently ordered.
    func selectRelativeArticle(offset: Int) {
        let ordered = displayedArticles
        guard !ordered.isEmpty else { return }

        guard let current = selectedArticleID,
              let index = ordered.firstIndex(where: { $0.id == current }) else {
            selectArticle(ordered[0])
            return
        }

        let target = index + offset
        guard ordered.indices.contains(target) else {
            if target >= ordered.count { loadMore() }
            return
        }
        selectArticle(ordered[target])
    }
}
