import SwiftUI

struct SidebarView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var isShowingAddFeed = false
    @State private var feedPendingDeletion: Feed?
    @State private var folderPendingDeletion: String?
    @State private var expandedFolders: Set<String> = []
    @State private var folderPrompt: FolderPrompt?
    @State private var dropTargetFolder: String?

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

            Section("Feeds") {
                feedSection
            }
        }
        .listStyle(.sidebar)
        .background {
            SidebarBackgroundMenu(title: "New Folder…") {
                folderPrompt = .create
            }
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
            ForEach(viewModel.feeds(inFolder: folder)) { feed in
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
            .background(
                RoundedRectangle(cornerRadius: 5)
                    .fill(dropTargetFolder == folder ? Color.accentColor.opacity(0.25) : .clear)
            )
            .dropDestination(for: FeedTransfer.self) { transfers, _ in
                guard !transfers.isEmpty else { return false }
                Task { await viewModel.moveFeeds(ids: transfers.map(\.feedID), toFolder: folder) }
                return true
            } isTargeted: { isTargeted in
                if isTargeted {
                    dropTargetFolder = folder
                    expandedFolders.insert(folder)
                } else if dropTargetFolder == folder {
                    dropTargetFolder = nil
                }
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
            .draggable(FeedTransfer(feedID: feed.id)) {
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
                AsyncImage(url: url) { image in
                    image
                        .resizable()
                        .scaledToFill()
                } placeholder: {
                    fallbackIcon
                }
                .frame(width: 16, height: 16)
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
