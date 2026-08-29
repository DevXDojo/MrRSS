import AppKit
import SwiftUI
import UniformTypeIdentifiers

struct SidebarView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var feedPendingDeletion: Feed?
    @State private var folderPendingDeletion: String?
    @State private var folderPrompt: FolderPrompt?
    @State private var feedBeingEdited: Feed?
    @State private var filterBeingEdited: SavedFilter?
    @State private var isCreatingSavedFilter = false
    @State private var discoverySource: Feed?
    @State private var isManagingTags = false
    @State private var isDiscoveringAll = false

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
        .sheet(item: $folderPrompt) { prompt in
            FolderNameView(prompt: prompt, viewModel: viewModel)
        }
        .sheet(item: $feedBeingEdited) { feed in
            FeedEditorView(mode: .edit(feed), viewModel: viewModel)
        }
        .sheet(item: $filterBeingEdited) { filter in
            SavedFilterEditorView(filter: filter, viewModel: viewModel)
        }
        .sheet(isPresented: $isCreatingSavedFilter) {
            SavedFilterEditorView(filter: nil, viewModel: viewModel)
        }
        .sheet(item: $discoverySource) { feed in
            DiscoveryView(feed: feed, viewModel: viewModel)
        }
        .sheet(isPresented: $isManagingTags) {
            TagManagerView(viewModel: viewModel)
        }
        .sheet(isPresented: $isDiscoveringAll) {
            DiscoveryView(feed: nil, viewModel: viewModel)
        }
        .confirmationDialog(
            t("modal.feed.unsubscribeTitle"),
            isPresented: Binding(
                get: { feedPendingDeletion != nil },
                set: { if !$0 { feedPendingDeletion = nil } }
            )
        ) {
            Button(t("common.action.unsubscribe"), role: .destructive) {
                guard let feed = feedPendingDeletion else { return }
                feedPendingDeletion = nil
                Task { await viewModel.deleteFeed(feed) }
            }
            Button(t("common.cancel"), role: .cancel) {
                feedPendingDeletion = nil
            }
        } message: {
            Text(t("modal.feed.unsubscribeMessage"))
        }
        .confirmationDialog(
            t("client.folder.deleteTitle"),
            isPresented: Binding(
                get: { folderPendingDeletion != nil },
                set: { if !$0 { folderPendingDeletion = nil } }
            )
        ) {
            Button(t("common.delete"), role: .destructive) {
                guard let folder = folderPendingDeletion else { return }
                folderPendingDeletion = nil
                Task { await viewModel.deleteFolder(folder) }
            }
            Button(t("common.cancel"), role: .cancel) {
                folderPendingDeletion = nil
            }
        } message: {
            Text(t("client.folder.deleteMessage"))
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
            },
            editFeed: { feedBeingEdited = $0 },
            refreshFeed: { feed in
                Task { await viewModel.refreshFeed(feed) }
            },
            discoverFrom: { discoverySource = $0 },
            markFeedRead: { feed in
                Task { await viewModel.markRead(feedID: feed.id) }
            },
            openFeedSite: { feed in
                guard let url = feed.siteURL else { return }
                NSWorkspace.shared.open(url)
            },
            markFolderRead: { folder in
                Task { await viewModel.markRead(category: folder) }
            },
            editSavedFilter: { filterBeingEdited = $0 },
            deleteSavedFilter: { filter in
                Task { await viewModel.deleteSavedFilter(filter) }
            },
            newSavedFilter: { isCreatingSavedFilter = true }
        )
    }

    private func importOPML() {
        let panel = NSOpenPanel()
        panel.allowedContentTypes = [.xml, .json]
        panel.allowsOtherFileTypes = true
        panel.allowsMultipleSelection = false
        panel.prompt = t("modal.opml.import")
        guard panel.runModal() == .OK, let url = panel.url else { return }
        Task { await viewModel.importOPML(from: url) }
    }

    private func exportOPML() {
        let panel = NSSavePanel()
        panel.nameFieldStringValue = "subscriptions.opml"
        panel.allowedContentTypes = [.xml]
        panel.prompt = t("modal.opml.export")
        guard panel.runModal() == .OK, let url = panel.url else { return }
        Task { await viewModel.exportOPML(to: url) }
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
            if viewModel.isRefreshingSources, viewModel.refreshProgress.outstandingCount > 0 {
                Text(
                    t(
                        "client.refresh.outstanding",
                        ["count": viewModel.refreshProgress.outstandingCount]
                    )
                )
                .font(.caption)
                .foregroundStyle(.secondary)
            }
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
                Button(t("sidebar.activity.addFeed"), systemImage: "plus") {
                    viewModel.isPresentingAddFeed = true
                }
                Button(t("client.sidebar.newFolder"), systemImage: "folder.badge.plus") {
                    folderPrompt = .create
                }
                Button(t("sidebar.savedFilters.saveFilter"), systemImage: "line.3.horizontal.decrease.circle") {
                    isCreatingSavedFilter = true
                }
                Button(t("modal.tag.manageTags"), systemImage: "tag") {
                    isManagingTags = true
                }
                Button(t("modal.discovery.discoverAllFeeds"), systemImage: "sparkle.magnifyingglass") {
                    isDiscoveringAll = true
                }
                Divider()
                Button(t("modal.opml.import"), systemImage: "square.and.arrow.down") {
                    importOPML()
                }
                Button(t("modal.opml.export"), systemImage: "square.and.arrow.up") {
                    exportOPML()
                }
            } label: {
                Label(t("common.action.add"), systemImage: "plus")
            }
            .help(t("sidebar.activity.addFeed"))

            Button(t("article.action.refreshFeedsShortcut"), systemImage: "arrow.clockwise") {
                viewModel.refreshFromSources()
            }
            .disabled(viewModel.isRefreshingSources)
            .help(t("article.action.refreshFeedsShortcut"))
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
        case .create, .createWithFeed: t("client.folder.newTitle")
        case .rename: t("client.folder.renameTitle")
        }
    }

    var message: String {
        switch self {
        case .create:
            t("client.folder.newMessage")
        case .createWithFeed(let feed):
            t("client.folder.moveMessage", ["name": feed.title])
        case .rename(let folder):
            t("client.folder.renameMessage", ["name": folder])
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
        case .create, .createWithFeed: t("client.folder.create")
        case .rename: t("client.folder.rename")
        }
    }
}
