import SwiftUI

struct SidebarView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var isShowingAddFeed = false
    @State private var feedPendingDeletion: Feed?

    var body: some View {
        VStack(spacing: 0) {
            List {
                Section("Library") {
                    ForEach(ArticleFilter.allCases) { filter in
                        sidebarButton(for: .filter(filter)) {
                            Label(filter.title, systemImage: filter.icon)
                        }
                        .badge(badge(for: filter))
                    }
                }

                Section("Feeds") {
                    if viewModel.isLoadingFeeds && viewModel.feeds.isEmpty {
                        HStack(spacing: 8) {
                            ProgressView()
                                .controlSize(.small)
                            Text("Loading feeds")
                                .foregroundStyle(.secondary)
                        }
                    } else if viewModel.feeds.isEmpty {
                        Text("No feeds available")
                            .font(.callout)
                            .foregroundStyle(.secondary)
                    } else {
                        ForEach(viewModel.feeds) { feed in
                            sidebarButton(for: .feed(feed.id)) {
                                FeedLabel(feed: feed)
                            }
                            .badge(viewModel.unreadCounts.feedCounts[feed.id] ?? 0)
                            .contextMenu {
                                Button("Delete Feed", systemImage: "trash", role: .destructive) {
                                    feedPendingDeletion = feed
                                }
                            }
                        }
                    }
                }
            }
            .listStyle(.sidebar)

            Divider()

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
        .navigationTitle("MrRSS")
        .toolbar {
            ToolbarItemGroup(placement: .primaryAction) {
                Button("Add feed", systemImage: "plus") {
                    isShowingAddFeed = true
                }
                .help("Add a feed")

                Button("Refresh feeds", systemImage: "arrow.clockwise") {
                    viewModel.refreshFromSources()
                }
                .disabled(viewModel.isRefreshingSources)
                .help("Refresh feeds and articles")
            }
        }
        .sheet(isPresented: $isShowingAddFeed) {
            AddFeedView(viewModel: viewModel)
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
    }

    private func badge(for filter: ArticleFilter) -> Int {
        switch filter {
        case .all, .unread:
            viewModel.unreadCounts.total
        case .favorites, .readLater:
            0
        }
    }

    private func sidebarButton<Label: View>(
        for item: SidebarItem,
        @ViewBuilder label: () -> Label
    ) -> some View {
        Button {
            viewModel.selection = item
        } label: {
            label()
                .frame(maxWidth: .infinity, alignment: .leading)
                .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .listRowBackground(
            viewModel.selection == item
                ? Color.accentColor.opacity(0.16)
                : Color.clear
        )
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
                    Image(systemName: "dot.radiowaves.left.and.right")
                }
                .frame(width: 16, height: 16)
                .clipShape(RoundedRectangle(cornerRadius: 3))
            } else {
                Image(systemName: "dot.radiowaves.left.and.right")
            }
        }
    }
}
