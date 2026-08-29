import SwiftUI
import AppKit

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
                contentView
            }
            .frame(width: geometry.size.width, height: geometry.size.height)
            .clipped()
        }
        .background(Color(nsColor: .textBackgroundColor))
        .toolbar {
            ToolbarItemGroup(placement: .primaryAction) {
                Menu {
                    Button {
                        Task { await translateTitle() }
                    } label: {
                        Label("Translate Title", systemImage: "character.book.closed")
                    }

                    Button {
                        Task { await translateArticle() }
                    } label: {
                        Label("Translate Content", systemImage: "text.bubble")
                    }

                    Button {
                        Task { await summarizeArticle() }
                    } label: {
                        Label("Generate Summary", systemImage: "text.quote")
                    }

                    if translatedContent != nil {
                        Divider()
                        Toggle("Show Translation", isOn: $displayTranslation)
                    }
                } label: {
                    if activeOperation != nil {
                        ProgressView()
                            .controlSize(.small)
                    } else {
                        Label("Language Tools", systemImage: "sparkles")
                    }
                }
                .disabled(activeOperation != nil)

                Button {
                    viewModel.toggleFavorite(article)
                } label: {
                    Label(
                        article.isFavorite ? "Remove Favorite" : "Favorite",
                        systemImage: article.isFavorite ? "star.fill" : "star"
                    )
                }
                .help(article.isFavorite ? "Remove from favorites" : "Add to favorites")

                if let url = URL(string: article.url) {
                    Link(destination: url) {
                        Label("Open in Browser", systemImage: "safari")
                    }
                    .help("Open the original article in your browser")
                }
            }
        }
        .task(id: article.id) {
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
            viewModel.setArticleRead(article, read: true)
            await loadContent()
            if viewModel.boolSetting("summary_enabled"),
               viewModel.setting("summary_trigger_mode", default: "manual") == "auto",
               summaryResult == nil {
                await summarizeArticle()
            }
        }
        .alert("Language Tool Error", isPresented: Binding(
            get: { operationError != nil },
            set: { if !$0 { operationError = nil } }
        )) {
            Button("Dismiss", role: .cancel) { operationError = nil }
        } message: {
            Text(operationError ?? "")
        }
        .alert("Language Tools", isPresented: Binding(
            get: { operationNotice != nil },
            set: { if !$0 { operationNotice = nil } }
        )) {
            Button("OK", role: .cancel) { operationNotice = nil }
        } message: {
            Text(operationNotice ?? "")
        }
    }

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
                if let feedTitle = article.feedTitle {
                    Text(feedTitle)
                        .fontWeight(.medium)
                        .foregroundStyle(.tint)
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

    @ViewBuilder
    private var contentView: some View {
        if isLoading {
            ProgressView("Loading article")
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        } else if let loadError {
            ContentUnavailableView {
                Label("Unable to Load Article", systemImage: "exclamationmark.triangle")
            } description: {
                Text(loadError)
            } actions: {
                Button("Retry") {
                    Task { await loadContent() }
                }
            }
        } else {
            VStack(spacing: 0) {
                if let summaryResult, !summaryResult.summary.isEmpty {
                    SummaryCard(result: summaryResult)
                    Divider()
                }

                if translatedContent != nil {
                    Picker("Content", selection: $displayTranslation) {
                        Text("Original").tag(false)
                        Text("Translation").tag(true)
                    }
                    .pickerStyle(.segmented)
                    .labelsHidden()
                    .frame(maxWidth: 280)
                    .padding(.vertical, 8)
                }

                WebView(
                    html: displayedContent,
                    baseURL: URL(string: articleContent.feedURL ?? article.url)
                )
                .frame(minWidth: 0, maxWidth: .infinity, minHeight: 0, maxHeight: .infinity)
                .clipped()
            }
        }
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
        return "<p>No article content is available.</p>"
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

    private func translateTitle() async {
        guard activeOperation == nil else { return }
        activeOperation = .translateTitle
        defer { activeOperation = nil }
        do {
            translatedTitle = try await viewModel.translateTitle(for: article)
            if translatedTitle == article.title {
                operationNotice = "The title already matches the selected target language."
            }
        } catch {
            operationError = error.localizedDescription
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
                operationNotice = "The article already primarily matches the selected target language."
            } else {
                displayTranslation = true
            }
        } catch {
            operationError = error.localizedDescription
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

private enum ArticleOperation {
    case translateTitle
    case translateContent
    case summarize
}

private struct SummaryCard: View {
    let result: SummaryResult

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Label("Summary", systemImage: "text.quote")
                    .font(.headline)
                Spacer()
                if result.cached == true {
                    Text("Cached")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                if result.usedFallback == true {
                    Text("Local fallback")
                        .font(.caption)
                        .foregroundStyle(.orange)
                }
            }
            Text(result.summary)
                .font(.callout)
                .textSelection(.enabled)
                .lineLimit(8)
                .fixedSize(horizontal: false, vertical: true)
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
