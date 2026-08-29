import SwiftUI
import UniformTypeIdentifiers

struct SidebarView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var isShowingAddFeed = false
    @State private var feedPendingDeletion: Feed?
    @State private var folderPendingDeletion: String?
    @State private var folderPrompt: FolderPrompt?

    var body: some View {
        VStack(spacing: 0) {
            SidebarOutline(
                feeds: viewModel.feeds,
                folders: viewModel.folders,
                counts: viewModel.unreadCounts,
                filterCounts: viewModel.filterCounts,
                savedFilters: viewModel.savedFilters,
                showsImageGallery: viewModel.boolSetting("image_gallery_enabled", default: true),
                selection: $viewModel.selection,
                actions: actions
            )

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

    private var actions: SidebarActions {
        SidebarActions(
            newFolder: { folderPrompt = .create },
            renameFolder: { folderPrompt = .rename($0) },
            deleteFolder: { folderPendingDeletion = $0 },
            newFolderHolding: { folderPrompt = .createWithFeed($0) },
            moveFeed: { feed, folder in
                Task { await viewModel.moveFeed(feed, toFolder: folder) }
            },
            deleteFeed: { feedPendingDeletion = $0 },
            placeFeed: { id, folder, index in
                Task { await viewModel.placeFeed(id: id, inFolder: folder, at: index) }
            }
        )
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
