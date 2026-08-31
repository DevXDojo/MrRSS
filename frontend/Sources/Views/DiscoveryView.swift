import SwiftUI

/// Runs the discovery engine and offers whatever it finds. Passing a feed
/// searches that site's links; passing none scans every subscription that has
/// not been scanned yet.
struct DiscoveryView: View {
    let feed: Feed?
    @ObservedObject var viewModel: AppViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var state = DiscoveryState()
    @State private var selected: Set<String> = []
    @State private var isSubscribing = false
    @State private var pollTask: Task<Void, Never>?

    var body: some View {
        VStack(spacing: 0) {
            header
            Divider()
            content
            Divider()
            footer
        }
        .frame(width: 640, height: 560)
        .task { await start() }
        .onDisappear { pollTask?.cancel() }
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(feed == nil ? t("modal.discovery.discoverAllFeeds") : t("modal.discovery.discoverFeeds"))
                .font(.title2.bold())
            Text(feed?.title ?? t("modal.discovery.discoverAllFeedsDesc"))
                .font(.subheadline)
                .foregroundStyle(.secondary)
            if feed == nil {
                Text(t("modal.discovery.discoveryLongRunningWarning"))
                    .font(.caption)
                    .foregroundStyle(.orange)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(20)
    }

    @ViewBuilder
    private var content: some View {
        if !state.error.isEmpty {
            ContentUnavailableView {
                Label(t("modal.discovery.discoveryFailed"), systemImage: "exclamationmark.triangle")
            } description: {
                Text(state.error)
            }
        } else if state.feeds.isEmpty {
            VStack(spacing: 12) {
                if state.isRunning {
                    ProgressView(value: state.progress.fraction ?? 0) {
                        Text(progressTitle)
                    }
                    .progressViewStyle(.linear)
                    .frame(maxWidth: 380)
                    if !state.progress.detail.isEmpty {
                        Text(state.progress.detail)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                            .lineLimit(1)
                    }
                } else {
                    ContentUnavailableView(
                        t("modal.discovery.noFriendLinksFound"),
                        systemImage: "magnifyingglass"
                    )
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        } else {
            List {
                ForEach(state.feeds) { blog in
                    row(for: blog)
                }
            }
            .listStyle(.inset)
        }
    }

    private func row(for blog: DiscoveredBlog) -> some View {
        HStack(alignment: .top, spacing: 10) {
            Toggle("", isOn: selectionBinding(for: blog))
                .labelsHidden()
                .toggleStyle(.checkbox)

            VStack(alignment: .leading, spacing: 3) {
                Text(blog.name.isEmpty ? blog.homepage : blog.name)
                    .fontWeight(.medium)
                Text(blog.rssFeed)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .lineLimit(1)
                ForEach(blog.recentArticles.prefix(2)) { article in
                    Text("· \(article.title)")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }
            }
            Spacer()
        }
        .padding(.vertical, 3)
    }

    private var footer: some View {
        HStack {
            if state.isRunning {
                ProgressView().controlSize(.small)
                Text(t("modal.discovery.foundSoFar", ["count": state.progress.foundCount]))
                    .font(.caption)
                    .foregroundStyle(.secondary)
            } else if !state.feeds.isEmpty {
                Text(t("modal.discovery.foundFeeds", ["count": state.feeds.count]))
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            Button(t("common.close"), role: .cancel) {
                pollTask?.cancel()
                dismiss()
            }

            Button(t("modal.feed.subscribeSelected")) {
                Task { await subscribe() }
            }
            .keyboardShortcut(.defaultAction)
            .disabled(selected.isEmpty || isSubscribing)
        }
        .padding(18)
    }

    private var progressTitle: String {
        if !state.progress.message.isEmpty { return state.progress.message }
        return t("modal.discovery.discovering")
    }

    private func selectionBinding(for blog: DiscoveredBlog) -> Binding<Bool> {
        Binding(
            get: { selected.contains(blog.rssFeed) },
            set: { isOn in
                if isOn {
                    selected.insert(blog.rssFeed)
                } else {
                    selected.remove(blog.rssFeed)
                }
            }
        )
    }

    private func start() async {
        do {
            if let feed {
                try await viewModel.api.clearDiscovery()
                try await viewModel.api.startDiscovery(feedID: feed.id)
            } else {
                try await viewModel.api.clearDiscoverAll()
                try await viewModel.api.startDiscoverAll()
            }
        } catch {
            state.error = error.localizedDescription
            return
        }
        poll()
    }

    /// Discovery reports progress by polling, as it did before.
    private func poll() {
        pollTask?.cancel()
        pollTask = Task {
            while !Task.isCancelled {
                do {
                    let latest = feed == nil
                        ? try await viewModel.api.fetchDiscoverAllProgress()
                        : try await viewModel.api.fetchDiscoveryProgress()
                    state = latest
                    if latest.isComplete || (!latest.isRunning && !latest.feeds.isEmpty) { return }
                } catch {
                    state.error = error.localizedDescription
                    return
                }
                try? await Task.sleep(for: .milliseconds(700))
            }
        }
    }

    private func subscribe() async {
        isSubscribing = true
        defer { isSubscribing = false }

        var succeeded = 0
        let targets = state.feeds.filter { selected.contains($0.rssFeed) }
        for blog in targets {
            var draft = FeedDraft(url: blog.rssFeed, title: blog.name)
            draft.category = feed?.category ?? ""
            // Reloading once at the end keeps a long list of subscriptions from
            // refetching everything per feed.
            if await viewModel.saveFeed(draft, isEditing: false, reloading: false) {
                succeeded += 1
            }
        }
        await viewModel.reloadAfterFeedChange()

        viewModel.statusMessage = succeeded == targets.count
            ? t("modal.feed.feedsSubscribedSuccess", ["count": succeeded])
            : t("modal.feed.feedsSubscribedPartial", ["succeeded": succeeded, "total": targets.count])
        dismiss()
    }
}
