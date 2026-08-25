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
        case .folder(let name): name
        case .feed(let feed): feed.title
        }
    }

    /// What selecting this row means. Group headings select nothing.
    var item: SidebarItem? {
        switch kind {
        case .group: nil
        case .filter(let filter): .filter(filter)
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

    private var identity: String {
        switch kind {
        case .group(let title): "group:\(title)"
        case .filter(let filter): "filter:\(filter.rawValue)"
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

    static let libraryTitle = "Library"
    static let feedsTitle = "Feeds"

    static func tree(feeds: [Feed], folders: [String], counts: UnreadCounts) -> [SidebarNode] {
        [library(counts: counts), feedsSection(feeds: feeds, folders: folders, counts: counts)]
    }

    private static func library(counts: UnreadCounts) -> SidebarNode {
        SidebarNode(
            kind: .group(libraryTitle),
            children: ArticleFilter.allCases.map { filter in
                let badge: Int
                switch filter {
                case .all, .unread: badge = counts.total
                case .favorites, .readLater: badge = 0
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
