import SwiftUI

struct ArticleRowView: View {
    let article: Article

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            unreadIndicator

            VStack(alignment: .leading, spacing: 7) {
                Text(article.translatedTitle?.isEmpty == false ? article.translatedTitle! : article.title)
                    .font(.system(size: 14, weight: article.isRead ? .regular : .semibold))
                    .foregroundStyle(article.isRead ? .secondary : .primary)
                    .lineLimit(3)
                    .fixedSize(horizontal: false, vertical: true)

                if let translatedTitle = article.translatedTitle,
                   !translatedTitle.isEmpty,
                   translatedTitle != article.title {
                    Text(article.title)
                        .font(.caption)
                        .foregroundStyle(.tertiary)
                        .lineLimit(1)
                }

                HStack(spacing: 6) {
                    if let feedTitle = article.feedTitle {
                        Text(feedTitle)
                            .lineLimit(1)
                    }

                    if article.feedTitle != nil {
                        Text("·")
                    }

                    Text(ArticleDateFormatter.relativeDescription(for: article.publishedAt))

                    Spacer(minLength: 4)

                    if article.isReadLater {
                        Image(systemName: "clock.fill")
                            .foregroundStyle(.tint)
                    }
                    if article.isFavorite {
                        Image(systemName: "star.fill")
                            .foregroundStyle(.yellow)
                    }
                }
                .font(.caption)
                .foregroundStyle(.secondary)
            }

            if let imageURL = article.imageURL,
               let url = URL(string: imageURL) {
                RemoteImage(url: url, displaySize: CGSize(width: 76, height: 58)) {
                    imagePlaceholder
                }
                .frame(width: 76, height: 58)
                .clipped()
                .clipShape(RoundedRectangle(cornerRadius: 7, style: .continuous))
            }
        }
        .padding(.vertical, 7)
        .contentShape(Rectangle())
    }

    private var unreadIndicator: some View {
        Circle()
            .fill(article.isRead ? Color.clear : Color.accentColor)
            .frame(width: 6, height: 6)
            .padding(.top, 6)
    }

    private var imagePlaceholder: some View {
        Rectangle()
            .fill(.quaternary)
            .foregroundStyle(.secondary)
    }
}
