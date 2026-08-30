import AppKit
import SwiftUI

struct ContentView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var keyMonitor: Any?

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
        .describingToolbarButtons()
        .overlay(alignment: .bottom) { StatusOverlay(viewModel: viewModel) }
        .onAppear(perform: installKeyMonitor)
        .onDisappear(perform: removeKeyMonitor)
        .sheet(isPresented: $viewModel.isPresentingAddFeed) {
            FeedEditorView(mode: .add, viewModel: viewModel)
        }
        .alert(
            "MrRSS",
            isPresented: Binding(
                get: { viewModel.errorMessage != nil },
                set: { if !$0 { viewModel.clearError() } }
            )
        ) {
            Button(t("client.action.dismiss"), role: .cancel) {
                viewModel.clearError()
            }
            Button(t("client.action.retry")) {
                viewModel.clearError()
                viewModel.refreshAll()
            }
        } message: {
            Text(viewModel.errorMessage ?? "")
        }
    }

    /// Single-key shortcuts cannot be menu commands, so key presses are watched
    /// here and ignored while the reader is typing into a field.
    private func installKeyMonitor() {
        guard keyMonitor == nil else { return }
        keyMonitor = NSEvent.addLocalMonitorForEvents(matching: .keyDown) { event in
            if isEditingText { return event }
            return viewModel.handleKeyPress(event) ? nil : event
        }
    }

    private func removeKeyMonitor() {
        if let keyMonitor {
            NSEvent.removeMonitor(keyMonitor)
        }
        keyMonitor = nil
    }

    private var isEditingText: Bool {
        guard let responder = NSApp.keyWindow?.firstResponder else { return false }
        if responder is NSTextView { return true }
        return responder is NSTextField
    }
}

private struct EmptyArticleView: View {
    var body: some View {
        ContentUnavailableView {
            Label(t("client.article.chooseArticle"), systemImage: "newspaper")
        } description: {
            Text(t("client.article.chooseArticleDetail"))
        }
    }
}
