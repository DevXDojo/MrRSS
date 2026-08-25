import SwiftUI
import UniformTypeIdentifiers

struct SidebarView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var isShowingAddFeed = false
    @State private var feedPendingDeletion: Feed?
    @State private var folderPendingDeletion: String?
    @State private var expandedFolders: Set<String> = []
    @State private var folderPrompt: FolderPrompt?
    @State private var dropTargetFolder: String?
    @State private var isDropTargetingRoot = false
    @State private var dragSession: FeedDragSession?
    @State private var dragCoordinator = SidebarDragCoordinator()

    var body: some View {
        VStack(spacing: 0) {
            feedList
            Divider()
            statusBar
        }
        .navigationTitle("MrRSS")
        .toolbar { toolbarContent }
        .sheet(isPresented: $isShowingAddFeed) {
            AddFeedView(viewModel: viewModel)
        }
        .sheet(item: $folderPrompt) { prompt in
            FolderNameView(prompt: prompt, viewModel: viewModel)
        }
        .confirmationDialog(
            "Delete \(feedPendingDeletion?.title ?? "this feed")?",
            isPresented: Binding(
                get: { feedPendingDeletion != nil },
                set: { if !$0 { feedPendingDeletion = nil } }
            )
        ) {
            Button("Delete Feed", role: .destructive) {
                guard let feed = feedPendingDeletion else { return }
                feedPendingDeletion = nil
                Task { await viewModel.deleteFeed(feed) }
            }
            Button("Cancel", role: .cancel) {
                feedPendingDeletion = nil
            }
        } message: {
            Text("The subscription and its stored articles will be removed.")
        }
        .confirmationDialog(
            "Delete the folder \(folderPendingDeletion ?? "")?",
            isPresented: Binding(
                get: { folderPendingDeletion != nil },
                set: { if !$0 { folderPendingDeletion = nil } }
            )
        ) {
            Button("Delete Folder", role: .destructive) {
                guard let folder = folderPendingDeletion else { return }
                folderPendingDeletion = nil
                Task { await viewModel.deleteFolder(folder) }
            }
            Button("Cancel", role: .cancel) {
                folderPendingDeletion = nil
            }
        } message: {
            Text("The feeds it holds stay subscribed and move back out of any folder.")
        }
    }

    private var feedList: some View {
        List(selection: $viewModel.selection) {
            Section("Library") {
                ForEach(ArticleFilter.allCases) { filter in
                    Label(filter.title, systemImage: filter.icon)
                        .badge(badge(for: filter))
                        .tag(SidebarItem.filter(filter))
                }
            }

            Section {
                feedSection
            } header: {
                // Dropping on the section header takes a subscription back out
                // of whatever folder it is in.
                Text("Feeds")
                    .padding(.horizontal, 4)
                    .background(dropHighlight(active: isDropTargetingRoot))
                    .onDrop(
                        of: [.mrrssFeed],
                        isTargeted: Binding(
                            get: { isDropTargetingRoot },
                            set: { isTargeted in
                                isDropTargetingRoot = isTargeted
                                if isTargeted {
                                    previewMove(intoFolder: "")
                                } else {
                                    scheduleDragSessionClear()
                                }
                            }
                        )
                    ) { providers in
                        move(providers, toFolder: nil)
                    }
            }
        }
        .listStyle(.sidebar)
        .background {
            SidebarBackgroundMenu(
                title: "New Folder…",
                action: { folderPrompt = .create },
                onResolveScrollView: { dragCoordinator.scrollView = $0 }
            )
        }
    }

    @ViewBuilder
    private var feedSection: some View {
        if viewModel.isLoadingFeeds && viewModel.feeds.isEmpty {
            HStack(spacing: 8) {
                ProgressView()
                    .controlSize(.small)
                Text("Loading feeds")
                    .foregroundStyle(.secondary)
            }
        } else if viewModel.feeds.isEmpty && viewModel.folders.isEmpty {
            Text("No feeds available")
                .font(.callout)
                .foregroundStyle(.secondary)
        } else {
            ForEach(viewModel.folders, id: \.self) { folder in
                folderRow(folder)
            }

            ForEach(viewModel.unfiledFeeds) { feed in
                feedRow(feed)
            }
        }
    }

    private var statusBar: some View {
        HStack(spacing: 8) {
            Circle()
                .fill(viewModel.connectionState.color)
                .frame(width: 7, height: 7)
            Text(viewModel.connectionState.title)
                .font(.caption)
                .foregroundStyle(.secondary)
            Spacer()
            if viewModel.isLoadingFeeds || viewModel.isRefreshingSources {
                ProgressView()
                    .controlSize(.mini)
            }
        }
        .padding(.horizontal, 12)
        .frame(height: 36)
        .background(.ultraThinMaterial)
    }

    @ToolbarContentBuilder
    private var toolbarContent: some ToolbarContent {
        ToolbarItemGroup(placement: .primaryAction) {
            Menu {
                Button("Add Feed…", systemImage: "plus") {
                    isShowingAddFeed = true
                }
                Button("New Folder…", systemImage: "folder.badge.plus") {
                    folderPrompt = .create
                }
            } label: {
                Label("Add", systemImage: "plus")
            }
            .help("Add a feed or a folder")

            Button("Refresh feeds", systemImage: "arrow.clockwise") {
                viewModel.refreshFromSources()
            }
            .disabled(viewModel.isRefreshingSources)
            .help("Refresh feeds and articles")
        }
    }

    private func folderRow(_ folder: String) -> some View {
        DisclosureGroup(
            isExpanded: Binding(
                get: { expandedFolders.contains(folder) },
                set: { isExpanded in
                    if isExpanded {
                        expandedFolders.insert(folder)
                    } else {
                        expandedFolders.remove(folder)
                    }
                }
            )
        ) {
            ForEach(viewModel.arrangedFeeds(inFolder: folder, previewing: dragSession)) { feed in
                feedRow(feed)
            }
        } label: {
            Label {
                Text(folder)
                    .lineLimit(1)
            } icon: {
                Image(systemName: expandedFolders.contains(folder) ? "folder.fill" : "folder")
                    .foregroundStyle(.tint)
            }
            .badge(viewModel.unreadCount(forFolder: folder))
            .tag(SidebarItem.folder(folder))
            .padding(.vertical, 2)
            .padding(.horizontal, 4)
            .background(dropHighlight(active: dropTargetFolder == folder))
            .onDrop(
                of: [.mrrssFeed],
                isTargeted: Binding(
                    get: { dropTargetFolder == folder },
                    set: { isTargeted in
                        if isTargeted {
                            dropTargetFolder = folder
                            expandedFolders.insert(folder)
                            previewMove(intoFolder: folder)
                        } else if dropTargetFolder == folder {
                            dropTargetFolder = nil
                            scheduleDragSessionClear()
                        }
                    }
                )
            ) { providers in
                move(providers, toFolder: folder)
            }
            .contextMenu {
                Button("Rename Folder…", systemImage: "pencil") {
                    folderPrompt = .rename(folder)
                }
                Button("Delete Folder", systemImage: "trash", role: .destructive) {
                    folderPendingDeletion = folder
                }
            }
        }
    }

    private func feedRow(_ feed: Feed) -> some View {
        FeedLabel(feed: feed)
            .badge(viewModel.unreadCounts.feedCounts[feed.id] ?? 0)
            .tag(SidebarItem.feed(feed.id))
            .opacity(dragSession?.feedID == feed.id ? 0.4 : 1)
            .background {
                GeometryReader { geometry in
                    Color.clear.onDrop(
                        of: [.mrrssFeed],
                        delegate: FeedRowDropDelegate(
                            rowHeight: geometry.size.height,
                            onHover: { above in previewInsertion(near: feed, above: above) },
                            onExit: { scheduleDragSessionClear() },
                            onDrop: { above, providers in
                                drop(providers, near: feed, above: above)
                            }
                        )
                    )
                }
            }
            .onDrag {
                startDragSession(for: feed)
                return FeedTransfer(feedID: feed.id).itemProvider
            } preview: {
                FeedLabel(feed: feed)
                    .padding(6)
                    .background(.regularMaterial, in: RoundedRectangle(cornerRadius: 6))
            }
            .contextMenu {
                Menu("Move to Folder") {
                    ForEach(viewModel.folders, id: \.self) { folder in
                        Button(folder) {
                            Task { await viewModel.moveFeed(feed, toFolder: folder) }
                        }
                        .disabled(feed.category == folder)
                    }

                    if !viewModel.folders.isEmpty {
                        Divider()
                    }

                    Button("New Folder…") {
                        folderPrompt = .createWithFeed(feed)
                    }

                    if !feed.category.isEmpty {
                        Divider()
                        Button("Remove from Folder") {
                            Task { await viewModel.moveFeed(feed, toFolder: nil) }
                        }
                    }
                }

                Divider()

                Button("Delete Feed", systemImage: "trash", role: .destructive) {
                    feedPendingDeletion = feed
                }
            }
    }

    private func startDragSession(for feed: Feed) {
        let siblings = viewModel.feeds(inFolder: feed.category)
        dragSession = FeedDragSession(
            feedID: feed.id,
            folder: feed.category,
            index: siblings.firstIndex(where: { $0.id == feed.id }) ?? siblings.count
        )
    }

    /// Opens a gap where the subscription would land. Hovering the travelling
    /// row itself changes nothing, which is what keeps the preview from
    /// oscillating as the rows move out of the way.
    private func previewInsertion(near feed: Feed, above: Bool) {
        dragCoordinator.hoverGeneration += 1
        dragCoordinator.beginAutoscroll()
        guard let session = dragSession, session.feedID != feed.id else { return }

        let siblings = viewModel.feeds(inFolder: feed.category).filter { $0.id != session.feedID }
        let anchor = siblings.firstIndex(where: { $0.id == feed.id }) ?? siblings.count
        let updated = FeedDragSession(
            feedID: session.feedID,
            folder: feed.category,
            index: above ? anchor : anchor + 1
        )

        guard updated != session else { return }
        withAnimation(SidebarView.reflow) {
            dragSession = updated
        }
    }

    /// A drag that leaves one row usually enters another straight away, so the
    /// preview is only dropped once nothing has claimed it.
    private func scheduleDragSessionClear() {
        dragCoordinator.hoverGeneration += 1
        let generation = dragCoordinator.hoverGeneration
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.12) {
            guard generation == dragCoordinator.hoverGeneration else { return }
            dragCoordinator.endAutoscroll()
            withAnimation(SidebarView.reflow) {
                dragSession = nil
            }
        }
    }

    /// A calm spring, so the rows settle the way the system's own lists do
    /// rather than snapping into place.
    private static let reflow = Animation.smooth(duration: 0.22)

    private func drop(_ providers: [NSItemProvider], near feed: Feed, above: Bool) -> Bool {
        guard !providers.isEmpty else { return false }

        let session = dragSession
        dragCoordinator.endAutoscroll()
        withAnimation(SidebarView.reflow) {
            dragSession = nil
        }

        Task {
            let ids = await FeedTransfer.feedIDs(from: providers)
            // Commit what the preview was showing; a drag from elsewhere has no
            // preview, so fall back to the row the pointer is on.
            if let session, ids == [session.feedID] {
                await viewModel.placeFeed(id: session.feedID, inFolder: session.folder, at: session.index)
            } else {
                await viewModel.moveFeeds(ids: ids, relativeTo: feed.id, placeAbove: above)
            }
        }
        return true
    }

    /// The system's own drop tint, rather than a colour of our choosing.
    private func dropHighlight(active: Bool) -> some View {
        RoundedRectangle(cornerRadius: 6, style: .continuous)
            .fill(Color(nsColor: .selectedContentBackgroundColor).opacity(active ? 0.28 : 0))
            .animation(.easeOut(duration: 0.12), value: active)
    }

    /// Lifts the travelling subscription out of the list it came from while the
    /// pointer rests on a folder, so the gap it leaves behind closes up.
    private func previewMove(intoFolder folder: String) {
        dragCoordinator.hoverGeneration += 1
        dragCoordinator.beginAutoscroll()
        guard let session = dragSession else { return }

        let updated = FeedDragSession(
            feedID: session.feedID,
            folder: folder,
            index: viewModel.feeds(inFolder: folder).filter { $0.id != session.feedID }.count
        )
        guard updated != session else { return }
        withAnimation(SidebarView.reflow) {
            dragSession = updated
        }
    }

    private func move(_ providers: [NSItemProvider], toFolder folder: String?) -> Bool {
        guard !providers.isEmpty else { return false }
        dragCoordinator.endAutoscroll()
        withAnimation(SidebarView.reflow) {
            dragSession = nil
        }
        Task {
            let ids = await FeedTransfer.feedIDs(from: providers)
            await viewModel.moveFeeds(ids: ids, toFolder: folder)
        }
        return true
    }

    private func badge(for filter: ArticleFilter) -> Int {
        switch filter {
        case .all, .unread:
            viewModel.unreadCounts.total
        case .favorites, .readLater:
            0
        }
    }
}

enum FolderPrompt: Identifiable {
    case create
    case createWithFeed(Feed)
    case rename(String)

    var id: String {
        switch self {
        case .create: "create"
        case .createWithFeed(let feed): "create-\(feed.id)"
        case .rename(let folder): "rename-\(folder)"
        }
    }

    var title: String {
        switch self {
        case .create, .createWithFeed: "New Folder"
        case .rename: "Rename Folder"
        }
    }

    var message: String {
        switch self {
        case .create:
            "Folders group your subscriptions in the sidebar."
        case .createWithFeed(let feed):
            "\(feed.title) will move into the new folder."
        case .rename(let folder):
            "Every feed in \(folder) moves to the new name."
        }
    }

    var initialName: String {
        switch self {
        case .create, .createWithFeed: ""
        case .rename(let folder): folder
        }
    }

    var confirmTitle: String {
        switch self {
        case .create, .createWithFeed: "Create"
        case .rename: "Rename"
        }
    }
}

private struct FeedLabel: View {
    let feed: Feed

    var body: some View {
        Label {
            Text(feed.title)
                .lineLimit(1)
        } icon: {
            if let iconURL = feed.iconURL,
               let url = URL(string: iconURL) {
                RemoteImage(url: url, displaySize: CGSize(width: 16, height: 16)) {
                    fallbackIcon
                }
                .frame(width: 16, height: 16)
                .clipped()
                .clipShape(RoundedRectangle(cornerRadius: 3))
            } else {
                fallbackIcon
            }
        }
    }

    private var fallbackIcon: some View {
        Image(systemName: "dot.radiowaves.left.and.right")
            .foregroundStyle(FeedAccentColor.color(for: feed))
    }
}
