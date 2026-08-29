import SwiftUI

struct ArticleRowView: View {
    let article: Article
    var feed: Feed?
    var layout: ArticleListLayout = .comfortable
    var showsPreviewImage = true
    var showsTranslatedTitle = false

    var body: some View {
        switch layout {
        case .compact:
            compactBody
        case .comfortable:
            comfortableBody
        case .cards:
            cardBody
        }
    }

    // MARK: - Layouts

    private var compactBody: some View {
        HStack(alignment: .firstTextBaseline, spacing: 8) {
            unreadDot
            Text(title)
                .font(.body)
                .fontWeight(article.isRead ? .regular : .semibold)
                .lineLimit(1)
            Spacer(minLength: 8)
            statusIcons
            Text(ArticleDateFormatter.relativeDescription(for: article.publishedAt))
                .font(.caption)
                .foregroundStyle(.secondary)
        }
        .padding(.vertical, 2)
    }

    private var comfortableBody: some View {
        HStack(alignment: .top, spacing: 10) {
            unreadDot
                .padding(.top, 6)

            VStack(alignment: .leading, spacing: 4) {
                Text(title)
                    .font(.body)
                    .fontWeight(article.isRead ? .regular : .semibold)
                    .lineLimit(2)
                    .foregroundStyle(article.isRead ? .secondary : .primary)

                metadata
            }

            Spacer(minLength: 8)

            if showsPreviewImage, let imageURL = article.imageURL, let url = URL(string: imageURL) {
                RemoteImage(url: url, displaySize: CGSize(width: 64, height: 48)) {
                    imagePlaceholder
                }
                .frame(width: 64, height: 48)
                .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
            }
        }
        .padding(.vertical, 4)
    }

    private var cardBody: some View {
        VStack(alignment: .leading, spacing: 8) {
            if showsPreviewImage, let imageURL = article.imageURL, let url = URL(string: imageURL) {
                RemoteImage(url: url, displaySize: CGSize(width: 360, height: 140)) {
                    imagePlaceholder
                }
                .frame(height: 140)
                .frame(maxWidth: .infinity)
                .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
            }

            HStack(alignment: .firstTextBaseline, spacing: 8) {
                unreadDot
                Text(title)
                    .font(.headline)
                    .fontWeight(article.isRead ? .regular : .semibold)
                    .lineLimit(3)
            }

            if let excerpt {
                Text(excerpt)
                    .font(.callout)
                    .foregroundStyle(.secondary)
                    .lineLimit(3)
            }

            metadata
        }
        .padding(.vertical, 8)
    }

    // MARK: - Pieces

    private var title: String {
        article.displayTitle(preferTranslation: showsTranslatedTitle)
    }

    private var excerpt: String? {
        guard let summary = article.summary ?? article.originalSummary else { return nil }
        return summary.strippingHTML
    }

    private var metadata: some View {
        HStack(spacing: 6) {
            if let source = feed?.title ?? article.feedTitle {
                Text(source)
                    .foregroundStyle(feed.map(FeedAccentColor.color(for:)) ?? .secondary)
                    .lineLimit(1)
            }
            if let author = article.author {
                Text("·")
                Text(author).lineLimit(1)
            }
            Text("·")
            Text(ArticleDateFormatter.relativeDescription(for: article.publishedAt))
            statusIcons
        }
        .font(.caption)
        .foregroundStyle(.secondary)
    }

    @ViewBuilder
    private var statusIcons: some View {
        HStack(spacing: 4) {
            if article.isFavorite {
                Image(systemName: "star.fill").foregroundStyle(.yellow)
            }
            if article.isReadLater {
                Image(systemName: "clock.fill").foregroundStyle(.orange)
            }
            if article.audioURL != nil {
                Image(systemName: "waveform")
            }
            if article.videoURL != nil {
                Image(systemName: "play.rectangle")
            }
            if article.summary != nil {
                Image(systemName: "sparkles")
            }
        }
        .font(.caption2)
        .imageScale(.small)
    }

    private var unreadDot: some View {
        Circle()
            .fill(article.isRead ? Color.clear : Color.accentColor)
            .frame(width: 6, height: 6)
    }

    private var imagePlaceholder: some View {
        Rectangle()
            .fill(.quaternary)
            .foregroundStyle(.secondary)
    }
}

extension String {
    /// A rough plain-text version of an HTML snippet, good enough for the
    /// one-line excerpts the list shows.
    var strippingHTML: String {
        replacingOccurrences(of: "<[^>]+>", with: " ", options: .regularExpression)
            .replacingOccurrences(of: "&nbsp;", with: " ")
            .replacingOccurrences(of: "&amp;", with: "&")
            .replacingOccurrences(of: "&lt;", with: "<")
            .replacingOccurrences(of: "&gt;", with: ">")
            .replacingOccurrences(of: "&quot;", with: "\"")
            .replacingOccurrences(of: "&#39;", with: "'")
            .replacingOccurrences(of: "\\s+", with: " ", options: .regularExpression)
            .trimmingCharacters(in: .whitespacesAndNewlines)
    }
}
