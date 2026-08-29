import AppKit
import SwiftUI

struct ArticleDetailView: View {
    let article: Article
    @ObservedObject var viewModel: AppViewModel

    @State private var articleContent = ArticleContent(content: "", feedURL: nil)
    @State private var isLoading = false
    @State private var loadError: String?
    @State private var translatedTitle: String?
    @State private var translatedContent: TextTranslationResponse?
    @State private var summaryResult: SummaryResult?
    @State private var displayTranslation = false
    @State private var activeOperation: ArticleOperation?
    @State private var operationError: String?
    @State private var operationNotice: String?
    @State private var viewMode: ArticleViewMode = .rendered
    @State private var isShowingFindBar = false
    @State private var findQuery = ""
    @State private var galleryImages: [String] = []
    @State private var isShowingGallery = false

    // NavigationSplitView measures its columns, and a column whose ideal size
    // follows the article text makes the split view lay itself out around that
    // size instead of around the window, which pushes the sidebar and the
    // article list out of the visible area. GeometryReader keeps the reported
    // size independent of the header, the summary card, and the web content.
    var body: some View {
        GeometryReader { geometry in
            VStack(spacing: 0) {
                articleHeader
                Divider()
                if isShowingFindBar {
                    findBar
                    Divider()
                }
                contentView
            }
            .frame(width: geometry.size.width, height: geometry.size.height)
            .clipped()
        }
        .background(Color(nsColor: .textBackgroundColor))
        .toolbar { toolbarContent }
        .task(id: article.id) { await prepare() }
        .onChange(of: viewModel.requestedViewModeToggle) { _, _ in
            viewMode = viewMode == .rendered ? .webpage : .rendered
        }
        .sheet(isPresented: $isShowingGallery) {
            ImageGalleryView(images: galleryImages, title: article.title)
        }
        .alert(t("common.errors.unknownError"), isPresented: Binding(
            get: { operationError != nil },
            set: { if !$0 { operationError = nil } }
        )) {
            Button(t("client.action.dismiss"), role: .cancel) { operationError = nil }
        } message: {
            Text(operationError ?? "")
        }
        .alert("MrRSS", isPresented: Binding(
            get: { operationNotice != nil },
            set: { if !$0 { operationNotice = nil } }
        )) {
            Button(t("common.confirm"), role: .cancel) { operationNotice = nil }
        } message: {
            Text(operationNotice ?? "")
        }
    }

    // MARK: - Header

    private var articleHeader: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(translatedTitle ?? article.translatedTitle ?? article.title)
                .font(.system(size: 25, weight: .bold, design: .rounded))
                .textSelection(.enabled)
                .lineLimit(4)
                .fixedSize(horizontal: false, vertical: true)

            if let translatedTitle, translatedTitle != article.title {
                Text(article.title)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .lineLimit(2)
                    .textSelection(.enabled)
            }

            HStack(spacing: 8) {
                if let feedTitle = viewModel.feed(for: article)?.title ?? article.feedTitle {
                    Text(feedTitle)
                        .fontWeight(.medium)
                        .foregroundStyle(.tint)
                }

                if let author = article.author {
                    Text("·").foregroundStyle(.secondary)
                    Text(author).foregroundStyle(.secondary)
                }

                Text(ArticleDateFormatter.fullDescription(for: article.publishedAt))
                    .foregroundStyle(.secondary)

                Spacer()
            }
            .font(.subheadline)
        }
        .padding(.horizontal, 28)
        .padding(.top, 24)
        .padding(.bottom, 20)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(.bar)
    }

    private var findBar: some View {
        HStack(spacing: 8) {
            Image(systemName: "magnifyingglass")
                .foregroundStyle(.secondary)
            TextField(t("common.findInPage.findInPagePlaceholder"), text: $findQuery)
                .textFieldStyle(.plain)
            Button {
                isShowingFindBar = false
                findQuery = ""
            } label: {
                Image(systemName: "xmark.circle.fill")
            }
            .buttonStyle(.plain)
            .foregroundStyle(.secondary)
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 8)
        .background(.bar)
    }

    // MARK: - Content

    @ViewBuilder
    private var contentView: some View {
        if isLoading {
            ProgressView(t("article.content.loadingContent"))
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        } else if let loadError {
            ContentUnavailableView {
                Label(t("common.errors.fetchingArticleContent"), systemImage: "exclamationmark.triangle")
            } description: {
                Text(loadError)
            } actions: {
                Button(t("client.action.retry")) {
                    Task { await loadContent() }
                }
            }
        } else {
            VStack(spacing: 0) {
                if let summaryResult, !summaryResult.summary.isEmpty {
                    SummaryCard(result: summaryResult)
                    Divider()
                }

                if let audioURL = article.audioURL, let url = URL(string: audioURL) {
                    MediaLinkBar(
                        title: t("article.audioPlayer.podcastAudio"),
                        icon: "waveform",
                        url: url
                    )
                    Divider()
                }

                if let videoURL = article.videoURL, let url = URL(string: videoURL) {
                    MediaLinkBar(
                        title: t("article.videoPlayer.youtubeVideo"),
                        icon: "play.rectangle",
                        url: url
                    )
                    Divider()
                }

                if translatedContent != nil {
                    Picker("", selection: $displayTranslation) {
                        Text(t("setting.reading.showOriginal")).tag(false)
                        Text(t("article.summary.translatedSummary")).tag(true)
                    }
                    .pickerStyle(.segmented)
                    .labelsHidden()
                    .frame(maxWidth: 280)
                    .padding(.vertical, 8)
                }

                WebView(
                    source: webSource,
                    typography: typography,
                    findQuery: findQuery
                )
                .frame(minWidth: 0, maxWidth: .infinity, minHeight: 0, maxHeight: .infinity)
                .clipped()
            }
        }
    }

    private var webSource: WebViewSource {
        if viewMode == .webpage, let url = URL(string: article.url) {
            return .url(url)
        }
        return .html(displayedContent, baseURL: URL(string: articleContent.feedURL ?? article.url))
    }

    private var typography: WebViewTypography {
        WebViewTypography(
            fontFamily: viewModel.setting("content_font_family", default: "system"),
            fontSize: Int(viewModel.setting("content_font_size", default: "16")) ?? 16,
            lineHeight: viewModel.setting("content_line_height", default: "1.6")
        )
    }

    private var displayedContent: String {
        if displayTranslation, let translatedContent {
            if !translatedContent.html.isEmpty { return translatedContent.html }
            return "<p>\(HTMLDocument.escape(translatedContent.translatedText))</p>"
        }
        if !articleContent.content.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            return articleContent.content
        }
        if let summary = article.summary, !summary.isEmpty {
            return "<p>\(HTMLDocument.escape(summary))</p>"
        }
        return "<p>\(HTMLDocument.escape(t("article.content.noContentAvailable")))</p>"
    }

    // MARK: - Toolbar

    @ToolbarContentBuilder
    private var toolbarContent: some ToolbarContent {
        ToolbarItemGroup(placement: .primaryAction) {
            Picker("", selection: $viewMode) {
                Label(t("article.content.renderContent"), systemImage: "doc.plaintext")
                    .tag(ArticleViewMode.rendered)
                Label(t("setting.reading.viewAsWebpage"), systemImage: "globe")
                    .tag(ArticleViewMode.webpage)
            }
            .pickerStyle(.segmented)
            .labelsHidden()
            .help(t("shortcut.toggle.contentView"))

            Button {
                viewModel.setArticleRead(article, read: !article.isRead)
            } label: {
                Label(
                    article.isRead ? t("article.action.markAsUnread") : t("article.action.markAsRead"),
                    systemImage: article.isRead ? "circle" : "checkmark.circle"
                )
            }
            .help(t("shortcut.toggle.readStatus"))

            Button {
                viewModel.toggleFavorite(article)
            } label: {
                Label(
                    article.isFavorite
                        ? t("article.action.removeFromFavorite")
                        : t("article.toolbar.addToFavorite"),
                    systemImage: article.isFavorite ? "star.fill" : "star"
                )
            }
            .help(t("article.toolbar.addToFavorite"))

            Button {
                viewModel.toggleReadLater(article)
            } label: {
                Label(
                    article.isReadLater
                        ? t("article.action.removeFromReadLater")
                        : t("article.toolbar.addToReadLater"),
                    systemImage: article.isReadLater ? "clock.fill" : "clock"
                )
            }
            .help(t("shortcut.toggle.readLaterStatus"))

            Menu {
                Button {
                    Task { await translateTitle() }
                } label: {
                    Label(t("client.article.translateTitle"), systemImage: "character.book.closed")
                }

                Button {
                    Task { await translateArticle() }
                } label: {
                    Label(t("client.article.translateContent"), systemImage: "text.bubble")
                }

                Button {
                    Task { await summarizeArticle() }
                } label: {
                    Label(t("article.summary.articleSummary"), systemImage: "text.quote")
                }

                if translatedContent != nil {
                    Divider()
                    Toggle(t("setting.reading.showTranslations"), isOn: $displayTranslation)
                }
            } label: {
                if activeOperation != nil {
                    ProgressView().controlSize(.small)
                } else {
                    Label(t("article.summary.articleSummary"), systemImage: "sparkles")
                }
            }
            .disabled(activeOperation != nil)

            Menu {
                Button {
                    Task { await reloadContent() }
                } label: {
                    Label(t("article.action.reloadContent"), systemImage: "arrow.clockwise")
                }

                Button {
                    Task { await fetchFullArticle() }
                } label: {
                    Label(t("article.action.fetchFullArticle"), systemImage: "doc.text.magnifyingglass")
                }

                Button {
                    Task { await showGallery() }
                } label: {
                    Label(t("sidebar.activity.imageGallery"), systemImage: "photo.on.rectangle.angled")
                }

                Divider()

                ForEach(ArticleExportDestination.allCases) { destination in
                    Button {
                        Task { await viewModel.exportArticle(article, to: destination) }
                    } label: {
                        Label(destination.localizedTitle, systemImage: destination.icon)
                    }
                }

                Divider()

                Button {
                    NSPasteboard.general.clearContents()
                    NSPasteboard.general.setString(article.url, forType: .string)
                    viewModel.statusMessage = t("common.toast.copiedToClipboard")
                } label: {
                    Label(t("common.contextMenu.copyLink"), systemImage: "link")
                }

                Button {
                    isShowingFindBar.toggle()
                } label: {
                    Label(t("common.findInPage.findInPagePlaceholder"), systemImage: "magnifyingglass")
                }
            } label: {
                Label(t("client.article.more"), systemImage: "ellipsis.circle")
            }

            Button {
                viewModel.openInBrowser(article)
            } label: {
                Label(t("article.action.openInBrowser"), systemImage: "safari")
            }
            .help(t("article.action.openInBrowser"))
        }
    }

    // MARK: - Work

    private func prepare() async {
        translatedTitle = article.translatedTitle
        summaryResult = article.summary.map {
            SummaryResult(
                summary: $0,
                html: nil,
                sentenceCount: nil,
                isTooShort: false,
                limitReached: nil,
                usedFallback: nil,
                thinking: nil,
                error: nil,
                cached: true
            )
        }
        viewMode = preferredViewMode
        translatedContent = nil
        displayTranslation = false
        galleryImages = []
        viewModel.markReadOnOpen(article)
        await loadContent()

        if viewModel.boolSetting("summary_enabled"),
           viewModel.setting("summary_trigger_mode", default: "manual") == "auto",
           summaryResult == nil {
            await summarizeArticle()
        }
    }

    /// The feed can override the global reading mode, as it could before.
    private var preferredViewMode: ArticleViewMode {
        let feedPreference = viewModel.feed(for: article)?.articleViewMode ?? "global"
        let value = feedPreference == "global"
            ? viewModel.setting("default_view_mode", default: "rendered")
            : feedPreference
        return value == "webpage" ? .webpage : .rendered
    }

    private func loadContent() async {
        isLoading = true
        loadError = nil
        do {
            articleContent = try await viewModel.fetchArticleContent(id: article.id)
        } catch is CancellationError {
            return
        } catch {
            loadError = error.localizedDescription
        }
        isLoading = false
    }

    private func reloadContent() async {
        guard activeOperation == nil else { return }
        activeOperation = .reload
        defer { activeOperation = nil }
        do {
            articleContent = try await viewModel.reloadArticleContent(id: article.id)
        } catch {
            operationError = error.localizedDescription
        }
    }

    private func fetchFullArticle() async {
        guard activeOperation == nil else { return }
        activeOperation = .fetchFull
        defer { activeOperation = nil }
        do {
            articleContent = try await viewModel.fetchFullArticle(id: article.id)
            operationNotice = t("article.action.fullArticleFetched")
        } catch {
            operationError = "\(t("common.errors.fetchingFullArticle")): \(error.localizedDescription)"
        }
    }

    private func showGallery() async {
        if galleryImages.isEmpty {
            galleryImages = await viewModel.articleImages(id: article.id)
        }
        guard !galleryImages.isEmpty else {
            operationNotice = t("client.article.noImages")
            return
        }
        isShowingGallery = true
    }

    private func translateTitle() async {
        guard activeOperation == nil else { return }
        activeOperation = .translateTitle
        defer { activeOperation = nil }
        do {
            translatedTitle = try await viewModel.translateTitle(for: article)
        } catch {
            operationError = "\(t("client.article.translateTitle")): \(error.localizedDescription)"
        }
    }

    private func translateArticle() async {
        guard activeOperation == nil else { return }
        activeOperation = .translateContent
        defer { activeOperation = nil }
        do {
            let source = plainText(from: articleContent.content)
            translatedContent = try await viewModel.translateContent(source)
            if translatedContent?.translatedText == source {
                operationNotice = t("common.errors.translatingContent")
            } else {
                displayTranslation = true
            }
        } catch {
            operationError = "\(t("common.errors.translating")): \(error.localizedDescription)"
        }
    }

    private func summarizeArticle() async {
        guard activeOperation == nil else { return }
        activeOperation = .summarize
        defer { activeOperation = nil }
        do {
            summaryResult = try await viewModel.summarize(
                article: article,
                content: articleContent.content.isEmpty ? nil : plainText(from: articleContent.content)
            )
        } catch {
            operationError = error.localizedDescription
        }
    }

    private func plainText(from html: String) -> String {
        guard let data = html.data(using: .utf8),
              let attributed = try? NSAttributedString(
                data: data,
                options: [
                    .documentType: NSAttributedString.DocumentType.html,
                    .characterEncoding: String.Encoding.utf8.rawValue
                ],
                documentAttributes: nil
              ) else {
            return html
        }
        return attributed.string
    }
}

/// Whether the reading pane shows the rendered article or the original page.
enum ArticleViewMode: String, Hashable {
    case rendered
    case webpage
}

private enum ArticleOperation {
    case translateTitle
    case translateContent
    case summarize
    case reload
    case fetchFull
}

/// A row offering the article's audio or video in an external player.
private struct MediaLinkBar: View {
    let title: String
    let icon: String
    let url: URL

    var body: some View {
        HStack(spacing: 10) {
            Label(title, systemImage: icon)
                .font(.callout)
            Spacer()
            Button(t("common.action.openWebsite")) {
                NSWorkspace.shared.open(url)
            }
            .controlSize(.small)
        }
        .padding(.horizontal, 24)
        .padding(.vertical, 10)
        .background(Color.accentColor.opacity(0.05))
    }
}

private struct SummaryCard: View {
    let result: SummaryResult

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Label(t("article.summary.articleSummary"), systemImage: "text.quote")
                    .font(.headline)
                Spacer()
                if result.usedFallback == true {
                    Text(t("article.summary.aiSummaryFallback"))
                        .font(.caption)
                        .foregroundStyle(.orange)
                }
                if result.limitReached == true {
                    Text(t("article.summary.aiLimitReached"))
                        .font(.caption)
                        .foregroundStyle(.orange)
                }
            }
            if result.isTooShort {
                Text(t("article.summary.articleTooShort"))
                    .font(.callout)
                    .foregroundStyle(.secondary)
            } else {
                Text(result.summary)
                    .font(.callout)
                    .textSelection(.enabled)
                    .fixedSize(horizontal: false, vertical: true)
            }
            if let error = result.error, !error.isEmpty {
                Text(error)
                    .font(.caption)
                    .foregroundStyle(.red)
            }
        }
        .padding(.horizontal, 24)
        .padding(.vertical, 14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color.accentColor.opacity(0.06))
    }
}
