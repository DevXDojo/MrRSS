import SwiftUI

struct ArticleListView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var isConfirmingMarkAllRead = false
    @State private var isConfirmingClearReadLater = false
    @State private var relativeMarkRequest: RelativeMarkRequest?

    var body: some View {
        List(selection: selectionBinding) {
            ForEach(viewModel.displayedArticles) { article in
                row(for: article)
            }

            if viewModel.isLoadingArticles && !viewModel.articles.isEmpty {
                loadingFooter
            } else if !viewModel.articles.isEmpty && !viewModel.hasMoreArticles {
                Text(t("article.list.allArticlesLoaded"))
                    .font(.footnote)
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity)
                    .listRowSeparator(.hidden)
            }
        }
        .listStyle(.inset)
        .overlay { emptyState }
        .navigationTitle(viewModel.articleListTitle)
        .toolbar { toolbarContent }
        .confirmationDialog(
            t("article.action.markAllReadConfirmTitle"),
            isPresented: $isConfirmingMarkAllRead
        ) {
            Button(t("article.action.markAllRead")) {
                Task { await viewModel.markAllRead() }
            }
            Button(t("common.action.cancel"), role: .cancel) {}
        } message: {
            Text(t("article.action.markAllReadConfirmMessage"))
        }
        .confirmationDialog(
            t("common.clearReadLater"),
            isPresented: $isConfirmingClearReadLater
        ) {
            Button(t("common.clearReadLater"), role: .destructive) {
                Task { await viewModel.clearReadLater() }
            }
            Button(t("common.action.cancel"), role: .cancel) {}
        }
        .confirmationDialog(
            relativeMarkRequest?.title ?? "",
            isPresented: Binding(
                get: { relativeMarkRequest != nil },
                set: { if !$0 { relativeMarkRequest = nil } }
            )
        ) {
            Button(t("common.action.confirm")) {
                guard let request = relativeMarkRequest else { return }
                relativeMarkRequest = nil
                Task { await viewModel.markRelative(to: request.article, direction: request.direction) }
            }
            Button(t("common.action.cancel"), role: .cancel) { relativeMarkRequest = nil }
        } message: {
            Text(relativeMarkRequest?.message ?? "")
        }
    }

    // MARK: - Rows

    private func row(for article: Article) -> some View {
        ArticleRowView(
            article: article,
            feed: viewModel.feed(for: article),
            layout: viewModel.listLayout,
            showsPreviewImage: viewModel.boolSetting("show_article_preview_images", default: true),
            showsTranslatedTitle: viewModel.boolSetting("translation_enabled")
        )
        .tag(article.id)
        .onAppear {
            if article.id == viewModel.displayedArticles.last?.id {
                viewModel.loadMore()
            }
        }
        .onHover { hovering in
            guard hovering,
                  viewModel.boolSetting("hover_mark_as_read"),
                  !article.isRead,
                  !article.isReadLater else { return }
            viewModel.setArticleRead(article, read: true)
        }
        .contextMenu { contextMenu(for: article) }
    }

    @ViewBuilder
    private func contextMenu(for article: Article) -> some View {
        Button {
            viewModel.setArticleRead(article, read: !article.isRead)
        } label: {
            Label(
                article.isRead ? t("article.action.markAsUnread") : t("article.action.markAsRead"),
                systemImage: article.isRead ? "circle" : "checkmark.circle"
            )
        }

        Button {
            viewModel.toggleFavorite(article)
        } label: {
            Label(
                article.isFavorite
                    ? t("article.action.removeFromFavorite")
                    : t("article.toolbar.addToFavorite"),
                systemImage: article.isFavorite ? "star.slash" : "star"
            )
        }

        Button {
            viewModel.toggleReadLater(article)
        } label: {
            Label(
                article.isReadLater
                    ? t("article.action.removeFromReadLater")
                    : t("article.toolbar.addToReadLater"),
                systemImage: article.isReadLater ? "clock.badge.xmark" : "clock"
            )
        }

        Button {
            viewModel.toggleHidden(article)
        } label: {
            Label(
                article.isHidden ? t("client.article.showArticle") : t("article.action.hideArticle"),
                systemImage: article.isHidden ? "eye" : "eye.slash"
            )
        }

        Divider()

        Button {
            relativeMarkRequest = RelativeMarkRequest(article: article, direction: .above)
        } label: {
            Label(t("article.action.markAboveAsRead"), systemImage: "arrow.up.to.line")
        }

        Button {
            relativeMarkRequest = RelativeMarkRequest(article: article, direction: .below)
        } label: {
            Label(t("article.action.markBelowAsRead"), systemImage: "arrow.down.to.line")
        }

        Divider()

        Button {
            viewModel.openInBrowser(article)
        } label: {
            Label(t("article.action.openInBrowser"), systemImage: "safari")
        }

        Button {
            copyToPasteboard(article.url)
        } label: {
            Label(t("common.contextMenu.copyLink"), systemImage: "link")
        }

        Button {
            copyToPasteboard(article.title)
        } label: {
            Label(t("common.contextMenu.copyTitle"), systemImage: "textformat")
        }
    }

    // MARK: - Toolbar

    @ToolbarContentBuilder
    private var toolbarContent: some ToolbarContent {
        ToolbarItemGroup(placement: .primaryAction) {
            if viewModel.selection == .filter(.readLater) {
                Button {
                    isConfirmingClearReadLater = true
                } label: {
                    Label(t("common.clearReadLater"), systemImage: "trash")
                }
                .help(t("common.clearReadLater"))
            }

            Button {
                isConfirmingMarkAllRead = true
            } label: {
                Label(t("article.action.markAllRead"), systemImage: "checkmark.circle")
            }
            .disabled(viewModel.articles.isEmpty)
            .help(t("article.action.markAllRead"))

            Toggle(isOn: $viewModel.showOnlyUnread) {
                Label(
                    viewModel.showOnlyUnread
                        ? t("setting.reading.showAllArticles")
                        : t("setting.reading.showOnlyUnread"),
                    systemImage: viewModel.showOnlyUnread ? "line.3.horizontal.decrease.circle.fill"
                        : "line.3.horizontal.decrease.circle"
                )
            }
            .toggleStyle(.button)
            .help(
                viewModel.showOnlyUnread
                    ? t("setting.reading.showAllArticles")
                    : t("setting.reading.showOnlyUnread")
            )

            Menu {
                Picker(t("client.sort.title"), selection: $viewModel.sortOrder) {
                    ForEach(ArticleSortOrder.allCases) { order in
                        Text(order.title).tag(order)
                    }
                }
                .pickerStyle(.inline)

                Picker(t("client.layout.title"), selection: $viewModel.listLayout) {
                    ForEach(ArticleListLayout.allCases) { layout in
                        Label(layout.title, systemImage: layout.icon).tag(layout)
                    }
                }
                .pickerStyle(.inline)
            } label: {
                Label(t("client.sort.title"), systemImage: "arrow.up.arrow.down")
            }

            Button {
                viewModel.refreshFromSources()
            } label: {
                Label(t("article.action.refresh"), systemImage: "arrow.clockwise")
            }
            .disabled(viewModel.isRefreshingSources)
            .help(t("article.action.refreshFeed"))
        }
    }

    // MARK: - Supporting views

    private var loadingFooter: some View {
        HStack {
            Spacer()
            ProgressView()
                .controlSize(.small)
            Spacer()
        }
        .listRowSeparator(.hidden)
    }

    @ViewBuilder
    private var emptyState: some View {
        if viewModel.isLoadingArticles && viewModel.articles.isEmpty {
            ProgressView(t("client.article.loadingContent"))
        } else if viewModel.articles.isEmpty {
            ContentUnavailableView(
                t("client.article.noArticles"),
                systemImage: "tray",
                description: Text(t("client.article.noArticlesDetail"))
            )
        }
    }

    private var selectionBinding: Binding<Int?> {
        Binding(
            get: { viewModel.selectedArticleID },
            set: { newValue in
                viewModel.selectedArticleID = newValue
                if let newValue, let article = viewModel.article(withID: newValue) {
                    viewModel.markReadOnOpen(article)
                }
            }
        )
    }

    private func copyToPasteboard(_ value: String) {
        NSPasteboard.general.clearContents()
        NSPasteboard.general.setString(value, forType: .string)
        viewModel.statusMessage = t("common.toast.copiedToClipboard")
    }
}

/// A pending "mark everything above or below" confirmation.
private struct RelativeMarkRequest: Identifiable {
    let article: Article
    let direction: MarkDirection

    var id: Int { article.id }

    var title: String {
        direction == .above
            ? t("article.action.markAboveReadConfirmTitle")
            : t("article.action.markBelowReadConfirmTitle")
    }

    var message: String {
        direction == .above
            ? t("article.action.markAboveReadConfirmMessage")
            : t("article.action.markBelowReadConfirmMessage")
    }
}
