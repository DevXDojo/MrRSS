import SwiftUI

/// The maintenance actions that sit alongside the storage settings.
struct StorageActionsView: View {
    @ObservedObject var viewModel: AppViewModel

    @State private var contentCache: ContentCacheInfo?
    @State private var mediaCache: MediaCacheInfo?
    @State private var pendingAction: MaintenanceAction?
    @State private var isWorking = false

    enum MaintenanceAction: String, Identifiable {
        case database
        case contentCache
        case mediaCache
        case generatedContent

        var id: String { rawValue }

        var title: String {
            switch self {
            case .database: t("setting.database.cleanDatabaseTitle")
            case .contentCache: t("setting.database.articleContentCacheCleanup")
            case .mediaCache: t("setting.database.mediaCacheCleanup")
            case .generatedContent: t("setting.content.clearTranslationCache")
            }
        }

        var message: String {
            switch self {
            case .database: t("setting.database.cleanDatabaseMessage")
            case .contentCache: t("setting.database.clearArticleContentCacheConfirm")
            case .mediaCache: t("setting.database.clearMediaCacheConfirm")
            case .generatedContent: t("setting.content.clearTranslationCacheConfirm")
            }
        }
    }

    var body: some View {
        Section(t("setting.database.dataManagement")) {
            LabeledContent(t("setting.database.currentCachedArticles")) {
                Text("\(contentCache?.cachedArticles ?? 0)")
                    .monospacedDigit()
            }
            LabeledContent(t("setting.database.currentCacheSize")) {
                Text(String(format: "%.1f MB", mediaCache?.cacheSizeMB ?? 0))
                    .monospacedDigit()
            }

            action(.database, buttonTitle: t("setting.database.cleanDatabase"))
            action(.contentCache, buttonTitle: t("setting.database.cleanupArticleContentCache"))
            action(.mediaCache, buttonTitle: t("setting.database.cleanupMediaCache"))
            action(.generatedContent, buttonTitle: t("setting.content.clearTranslationCacheButton"))
        }
        .task { await refreshInfo() }
        .confirmationDialog(
            pendingAction?.title ?? "",
            isPresented: Binding(
                get: { pendingAction != nil },
                set: { if !$0 { pendingAction = nil } }
            )
        ) {
            Button(t("common.action.confirm"), role: .destructive) {
                guard let action = pendingAction else { return }
                pendingAction = nil
                Task { await run(action) }
            }
            Button(t("common.cancel"), role: .cancel) { pendingAction = nil }
        } message: {
            Text(pendingAction?.message ?? "")
        }
    }

    private func action(_ action: MaintenanceAction, buttonTitle: String) -> some View {
        LabeledContent(action.title) {
            Button(buttonTitle, role: .destructive) {
                pendingAction = action
            }
            .disabled(isWorking)
        }
    }

    private func refreshInfo() async {
        contentCache = try? await viewModel.api.fetchContentCacheInfo()
        mediaCache = try? await viewModel.api.fetchMediaCacheInfo()
    }

    private func run(_ action: MaintenanceAction) async {
        isWorking = true
        defer { isWorking = false }
        do {
            switch action {
            case .database:
                try await viewModel.api.cleanupArticles()
            case .contentCache:
                try await viewModel.api.cleanupContentCache()
            case .mediaCache:
                try await viewModel.api.cleanupMediaCache()
            case .generatedContent:
                await viewModel.clearGeneratedContent()
            }
            viewModel.statusMessage = t("setting.content.clearTranslationCacheSuccess")
            await refreshInfo()
        } catch {
            viewModel.errorMessage = error.localizedDescription
        }
    }
}
