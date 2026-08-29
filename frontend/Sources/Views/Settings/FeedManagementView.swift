import SwiftUI

/// Lists every subscription with its status, and applies changes to one or to a
/// whole selection, as the previous "Manage Feeds" pane did.
struct FeedManagementView: View {
    @ObservedObject var viewModel: AppViewModel

    @State private var search = ""
    @State private var sort: FeedSort = .original
    @State private var ascending = true
    @State private var selection: Set<Int> = []
    @State private var feedBeingEdited: Feed?
    @State private var isConfirmingDelete = false
    @State private var moveTarget: String?

    /// How the list is ordered.
    enum FeedSort: String, CaseIterable, Identifiable {
        case original
        case name
        case category
        case latestArticle
        case updateStatus

        var id: String { rawValue }

        var title: String {
            switch self {
            case .original: t("modal.feed.originalOrder")
            case .name: t("sidebar.sort.byName")
            case .category: t("sidebar.sort.byCategory")
            case .latestArticle: t("sidebar.sort.byLatestArticle")
            case .updateStatus: t("sidebar.sort.byUpdateStatus")
            }
        }
    }

    var body: some View {
        VStack(spacing: 0) {
            controls
            Divider()
            list
            Divider()
            footer
        }
        .sheet(item: $feedBeingEdited) { feed in
            FeedEditorView(mode: .edit(feed), viewModel: viewModel)
        }
        .confirmationDialog(
            t("modal.feed.deleteMultipleFeedsTitle"),
            isPresented: $isConfirmingDelete
        ) {
            Button(t("common.action.deleteSelected"), role: .destructive) {
                Task { await deleteSelected() }
            }
            Button(t("common.cancel"), role: .cancel) {}
        } message: {
            Text(t("modal.feed.deleteMultipleFeedsMessage", ["count": selection.count]))
        }
    }

    // MARK: - Pieces

    private var controls: some View {
        HStack(spacing: 10) {
            HStack(spacing: 6) {
                Image(systemName: "magnifyingglass")
                    .foregroundStyle(.secondary)
                TextField(t("common.search.searchFeeds"), text: $search)
                    .textFieldStyle(.plain)
            }
            .padding(.horizontal, 8)
            .padding(.vertical, 5)
            .background(.quaternary, in: RoundedRectangle(cornerRadius: 6, style: .continuous))

            Picker("", selection: $sort) {
                ForEach(FeedSort.allCases) { option in
                    Text(option.title).tag(option)
                }
            }
            .labelsHidden()
            .frame(width: 190)

            Button {
                ascending.toggle()
            } label: {
                Image(systemName: ascending ? "arrow.up" : "arrow.down")
            }
            .help(ascending ? t("modal.feed.sortAscending") : t("modal.feed.sortDescending"))
        }
        .padding(12)
    }

    @ViewBuilder
    private var list: some View {
        if viewModel.feeds.isEmpty {
            ContentUnavailableView(t("modal.feed.noFeeds"), systemImage: "dot.radiowaves.left.and.right")
                .frame(maxHeight: .infinity)
        } else if orderedFeeds.isEmpty {
            ContentUnavailableView(t("common.search.noSearchResults"), systemImage: "magnifyingglass")
                .frame(maxHeight: .infinity)
        } else {
            List {
                ForEach(orderedFeeds) { feed in
                    row(for: feed)
                }
            }
            .listStyle(.inset)
        }
    }

    private func row(for feed: Feed) -> some View {
        HStack(spacing: 10) {
            Toggle("", isOn: selectionBinding(for: feed))
                .labelsHidden()
                .toggleStyle(.checkbox)

            VStack(alignment: .leading, spacing: 2) {
                HStack(spacing: 6) {
                    Text(feed.title)
                        .fontWeight(.medium)
                        .lineLimit(1)
                    if feed.isFreshRSSSource {
                        badge(t("modal.feed.typeFreshRSS"))
                    } else if feed.isXPathFeed {
                        badge(t("modal.feed.typeXPath"))
                    } else if feed.isEmailFeed {
                        badge(t("modal.feed.typeEmail"))
                    }
                    if feed.isImageMode {
                        badge(t("setting.feed.imageMode"))
                    }
                }

                HStack(spacing: 6) {
                    if !feed.category.isEmpty {
                        Text(feed.category)
                    }
                    Text(feed.url)
                        .lineLimit(1)
                        .truncationMode(.middle)
                }
                .font(.caption)
                .foregroundStyle(.secondary)

                if !feed.lastError.isEmpty {
                    Text(errorDescription(for: feed.lastError))
                        .font(.caption)
                        .foregroundStyle(.red)
                        .lineLimit(1)
                }
            }

            Spacer()

            Text("\(viewModel.unreadCounts.feedCounts[feed.id] ?? 0)")
                .font(.caption)
                .monospacedDigit()
                .foregroundStyle(.secondary)

            Button(t("common.edit")) { feedBeingEdited = feed }
                .buttonStyle(.borderless)
        }
        .padding(.vertical, 2)
    }

    private var footer: some View {
        HStack(spacing: 10) {
            Button(selection.count == orderedFeeds.count ? t("common.action.deselectAll") : t("common.search.selectAll")) {
                if selection.count == orderedFeeds.count {
                    selection.removeAll()
                } else {
                    selection = Set(orderedFeeds.map(\.id))
                }
            }

            if !selection.isEmpty {
                Text(t("common.search.itemsSelected", ["count": selection.count]))
                    .font(.caption)
                    .foregroundStyle(.secondary)

                Menu(t("common.action.moveSelected")) {
                    Button(t("sidebar.feedList.uncategorized")) {
                        Task { await move(to: "") }
                    }
                    ForEach(viewModel.folders, id: \.self) { folder in
                        Button(folder) {
                            Task { await move(to: folder) }
                        }
                    }
                }
                .frame(width: 150)

                Button(t("common.action.deleteSelected"), role: .destructive) {
                    isConfirmingDelete = true
                }
            }

            Spacer()

            Button(t("sidebar.activity.addFeed")) {
                viewModel.isPresentingAddFeed = true
            }
        }
        .padding(12)
    }

    private func badge(_ text: String) -> some View {
        Text(text)
            .font(.caption2)
            .padding(.horizontal, 5)
            .padding(.vertical, 1)
            .background(Color.accentColor.opacity(0.15), in: Capsule())
    }

    // MARK: - Behaviour

    /// The subscriptions matching the search, in the chosen order.
    private var orderedFeeds: [Feed] {
        let query = search.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        let matching = query.isEmpty
            ? viewModel.feeds
            : viewModel.feeds.filter {
                $0.title.lowercased().contains(query)
                    || $0.url.lowercased().contains(query)
                    || $0.category.lowercased().contains(query)
            }

        let sorted: [Feed]
        switch sort {
        case .original:
            sorted = matching
        case .name:
            sorted = matching.sorted { $0.title.localizedStandardCompare($1.title) == .orderedAscending }
        case .category:
            sorted = matching.sorted {
                $0.category.localizedStandardCompare($1.category) == .orderedAscending
            }
        case .latestArticle:
            sorted = matching.sorted {
                (ArticleDateFormatter.date(from: $0.lastUpdated) ?? .distantPast)
                    > (ArticleDateFormatter.date(from: $1.lastUpdated) ?? .distantPast)
            }
        case .updateStatus:
            sorted = matching.sorted { lhs, rhs in
                // Feeds that failed come first, so they are easy to find.
                (!lhs.lastError.isEmpty ? 0 : 1) < (!rhs.lastError.isEmpty ? 0 : 1)
            }
        }

        return ascending ? sorted : sorted.reversed()
    }

    private func selectionBinding(for feed: Feed) -> Binding<Bool> {
        Binding(
            get: { selection.contains(feed.id) },
            set: { isOn in
                if isOn {
                    selection.insert(feed.id)
                } else {
                    selection.remove(feed.id)
                }
            }
        )
    }

    /// Turns a stored fetch failure into the wording the previous interface used.
    private func errorDescription(for error: String) -> String {
        let lowered = error.lowercased()
        if lowered.contains("timeout") || lowered.contains("deadline") {
            return t("modal.feed.errorTimeout")
        }
        if lowered.contains("no such host") || lowered.contains("dns") {
            return t("modal.feed.errorDNS")
        }
        if lowered.contains("certificate") || lowered.contains("x509") {
            return t("modal.feed.errorCertificate")
        }
        if lowered.contains("404") {
            return t("modal.feed.errorNotFound")
        }
        if lowered.contains("401") || lowered.contains("403") {
            return t("modal.feed.errorUnauthorized")
        }
        if lowered.contains("500") || lowered.contains("502") || lowered.contains("503") {
            return t("modal.feed.errorServer")
        }
        if lowered.contains("parse") || lowered.contains("invalid") {
            return t("modal.feed.errorInvalidFormat")
        }
        return error
    }

    private func move(to folder: String) async {
        for id in selection {
            guard let feed = viewModel.feeds.first(where: { $0.id == id }) else { continue }
            await viewModel.moveFeed(feed, toFolder: folder)
        }
        viewModel.statusMessage = t("modal.feed.feedsMovedSuccess")
        selection.removeAll()
    }

    private func deleteSelected() async {
        for id in selection {
            guard let feed = viewModel.feeds.first(where: { $0.id == id }) else { continue }
            await viewModel.deleteFeed(feed, reloading: false)
        }
        await viewModel.reloadAfterFeedChange()
        viewModel.statusMessage = t("modal.feed.feedsDeletedSuccess")
        selection.removeAll()
    }
}
