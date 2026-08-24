import SwiftUI

struct ContentView: View {
    @ObservedObject var viewModel: AppViewModel

    var body: some View {
        NavigationSplitView {
            SidebarView(viewModel: viewModel)
                .navigationSplitViewColumnWidth(min: 210, ideal: 240, max: 300)
        } content: {
            ArticleListView(viewModel: viewModel)
                .navigationSplitViewColumnWidth(min: 330, ideal: 410, max: 520)
        } detail: {
            if let article = viewModel.article(withID: viewModel.selectedArticleID) {
                ArticleDetailView(article: article, viewModel: viewModel)
                    .id(article.id)
            } else {
                EmptyArticleView()
            }
        }
        .frame(minWidth: 980, minHeight: 620)
        .alert(
            "MrRSS",
            isPresented: Binding(
                get: { viewModel.errorMessage != nil },
                set: { if !$0 { viewModel.clearError() } }
            )
        ) {
            Button("Dismiss", role: .cancel) {
                viewModel.clearError()
            }
            Button("Retry") {
                viewModel.clearError()
                viewModel.refreshAll()
            }
        } message: {
            Text(viewModel.errorMessage ?? "")
        }
    }
}

private struct EmptyArticleView: View {
    var body: some View {
        ContentUnavailableView {
            Label("Choose an article", systemImage: "newspaper")
        } description: {
            Text("Select an article from the list to read it here.")
        }
    }
}
