import SwiftUI

struct ArticleListView: View {
    @ObservedObject var viewModel: AppViewModel

    var body: some View {
        List {
            ForEach(viewModel.articles) { article in
                Button {
                    viewModel.selectArticle(article)
                } label: {
                    ArticleRowView(article: article)
                        .frame(maxWidth: .infinity, alignment: .leading)
                }
                .buttonStyle(.plain)
                .listRowBackground(
                    viewModel.selectedArticleID == article.id
                        ? Color.accentColor.opacity(0.14)
                        : Color.clear
                )
                .onAppear {
                    if article.id == viewModel.articles.last?.id {
                        viewModel.loadMore()
                    }
                }
                .contextMenu {
                    Button {
                        viewModel.setArticleRead(article, read: !article.isRead)
                    } label: {
                        Label(
                            article.isRead ? "Mark as Unread" : "Mark as Read",
                            systemImage: article.isRead ? "circle" : "checkmark.circle"
                        )
                    }

                    Button {
                        viewModel.toggleFavorite(article)
                    } label: {
                        Label(
                            article.isFavorite ? "Remove Favorite" : "Favorite",
                            systemImage: article.isFavorite ? "star.slash" : "star"
                        )
                    }
                }
            }

            if viewModel.isLoadingArticles && !viewModel.articles.isEmpty {
                HStack {
                    Spacer()
                    ProgressView()
                        .controlSize(.small)
                    Spacer()
                }
                .listRowSeparator(.hidden)
            }
        }
        .listStyle(.inset)
        .overlay {
            if viewModel.isLoadingArticles && viewModel.articles.isEmpty {
                ProgressView("Loading articles")
            } else if viewModel.articles.isEmpty {
                ContentUnavailableView(
                    "No Articles",
                    systemImage: "tray",
                    description: Text("Refresh or choose another source.")
                )
            }
        }
        .navigationTitle(currentTitle)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button("Refresh articles", systemImage: "arrow.clockwise") {
                    viewModel.reloadArticles()
                }
                .disabled(viewModel.isLoadingArticles)
                .help("Reload the current article list")
            }
        }
    }

    private var currentTitle: String {
        switch viewModel.selection {
        case .filter(let filter):
            filter.title
        case .feed(let id):
            viewModel.feeds.first(where: { $0.id == id })?.title ?? "Feed"
        case .none:
            "Articles"
        }
    }
}
