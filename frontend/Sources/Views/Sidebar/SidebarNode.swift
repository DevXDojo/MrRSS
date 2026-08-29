import AppKit

/// One row of the sidebar outline.
///
/// `NSOutlineView` identifies rows by object equality, so equality here is the
/// row's identity rather than its contents: a feed keeps its expansion and
/// selection across a reload even though the tree is rebuilt from scratch.
final class SidebarNode: NSObject {
    enum Kind {
        case group(String)
        case filter(ArticleFilter)
        case savedFilter(SavedFilter)
        case folder(String)
        case feed(Feed)
    }

    let kind: Kind
    let children: [SidebarNode]
    let badge: Int

    init(kind: Kind, children: [SidebarNode] = [], badge: Int = 0) {
        self.kind = kind
        self.children = children
        self.badge = badge
    }

    var isGroup: Bool {
        if case .group = kind { return true }
        return false
    }

    var title: String {
        switch kind {
        case .group(let title): title
        case .filter(let filter): filter.title
        case .savedFilter(let filter): filter.name
        case .folder(let name): name
        case .feed(let feed): feed.title
        }
    }

    /// What selecting this row means. Group headings select nothing.
    var item: SidebarItem? {
        switch kind {
        case .group: nil
        case .filter(let filter): .filter(filter)
        case .savedFilter(let filter): .savedFilter(filter.id)
        case .folder(let name): .folder(name)
        case .feed(let feed): .feed(feed.id)
        }
    }

    var feed: Feed? {
        if case .feed(let feed) = kind { return feed }
        return nil
    }

    var folderName: String? {
        if case .folder(let name) = kind { return name }
        return nil
    }

    var savedFilter: SavedFilter? {
        if case .savedFilter(let filter) = kind { return filter }
        return nil
    }

    private var identity: String {
        switch kind {
        case .group(let title): "group:\(title)"
        case .filter(let filter): "filter:\(filter.rawValue)"
        case .savedFilter(let filter): "savedFilter:\(filter.id)"
        case .folder(let name): "folder:\(name)"
        case .feed(let feed): "feed:\(feed.id)"
        }
    }

    override func isEqual(_ object: Any?) -> Bool {
        guard let other = object as? SidebarNode else { return false }
        return identity == other.identity
    }

    override var hash: Int {
        identity.hashValue
    }

    // MARK: - Building

    static var libraryTitle: String { t("client.sidebar.library") }
    static var feedsTitle: String { t("sidebar.feedList.feeds") }
    static var savedFiltersTitle: String { t("sidebar.savedFilters.title") }

    /// Builds the outline. `activity` decides which per-feed count is shown as
    /// a badge, matching the activity the reader is currently in.
    static func tree(
        feeds: [Feed],
        folders: [String],
        counts: UnreadCounts,
        filterCounts: FilterCounts = .empty,
        savedFilters: [SavedFilter] = [],
        showImageGallery: Bool = true
    ) -> [SidebarNode] {
        var sections = [
            library(counts: counts, filterCounts: filterCounts, showImageGallery: showImageGallery)
        ]
        if !savedFilters.isEmpty {
            sections.append(
                SidebarNode(
                    kind: .group(savedFiltersTitle),
                    children: savedFilters.map { SidebarNode(kind: .savedFilter($0)) }
                )
            )
        }
        sections.append(feedsSection(feeds: feeds, folders: folders, counts: counts))
        return sections
    }

    private static func library(
        counts: UnreadCounts,
        filterCounts: FilterCounts,
        showImageGallery: Bool
    ) -> SidebarNode {
        let activities = ArticleFilter.allCases.filter { showImageGallery || $0 != .imageGallery }
        return SidebarNode(
            kind: .group(libraryTitle),
            children: activities.map { filter in
                let badge: Int
                switch filter {
                case .all, .unread:
                    badge = counts.total
                case .favorites:
                    badge = filterCounts.total(for: \.favorites)
                case .readLater:
                    badge = filterCounts.total(for: \.readLater)
                case .imageGallery:
                    badge = filterCounts.total(for: \.imagesUnread)
                }
                return SidebarNode(kind: .filter(filter), badge: badge)
            }
        )
    }

    private static func feedsSection(feeds: [Feed], folders: [String], counts: UnreadCounts) -> SidebarNode {
        var children = folders.map { folder -> SidebarNode in
            let members = feeds.filter { $0.category == folder }
            return SidebarNode(
                kind: .folder(folder),
                children: members.map { node(for: $0, counts: counts) },
                badge: members.reduce(0) { $0 + (counts.feedCounts[$1.id] ?? 0) }
            )
        }
        children.append(contentsOf: feeds.filter { $0.category.isEmpty }.map { node(for: $0, counts: counts) })

        return SidebarNode(kind: .group(feedsTitle), children: children)
    }

    private static func node(for feed: Feed, counts: UnreadCounts) -> SidebarNode {
        SidebarNode(kind: .feed(feed), badge: counts.feedCounts[feed.id] ?? 0)
    }
}
