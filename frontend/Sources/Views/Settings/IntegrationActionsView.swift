import SwiftUI

/// The connection tests and the synchronisation controls that belong with the
/// integration settings.
struct IntegrationActionsView: View {
    @ObservedObject var viewModel: AppViewModel

    @State private var freshRSSStatus: FreshRSSStatus?
    @State private var rsshubMessage: String?
    @State private var isTestingRSSHub = false
    @State private var isSyncing = false

    var body: some View {
        Section(t("setting.freshrss.enabled")) {
            if let freshRSSStatus {
                LabeledContent(t("client.freshrss.pendingChanges")) {
                    Text("\(freshRSSStatus.pendingChanges)").monospacedDigit()
                }
                LabeledContent(t("setting.freshrss.lastSync")) {
                    Text(
                        freshRSSStatus.lastSyncTime.map {
                            ArticleDateFormatter.relativeDescription(for: $0)
                        } ?? t("setting.freshrss.never")
                    )
                }
            }

            HStack {
                Button(t("setting.freshrss.syncNow")) {
                    Task { await sync() }
                }
                .disabled(isSyncing || !viewModel.boolSetting("freshrss_enabled"))
                if isSyncing {
                    ProgressView().controlSize(.small)
                }
            }
        }
        .task { freshRSSStatus = try? await viewModel.api.fetchFreshRSSStatus() }

        Section(t("setting.rsshub.enabled")) {
            HStack {
                Button(t("setting.rsshub.testConnection")) {
                    Task { await testRSSHub() }
                }
                .disabled(isTestingRSSHub || !viewModel.boolSetting("rsshub_enabled"))
                if isTestingRSSHub {
                    ProgressView().controlSize(.small)
                }
                if let rsshubMessage {
                    Text(rsshubMessage)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
            Text(t("setting.rsshub.testConnectionDesc"))
                .font(.caption)
                .foregroundStyle(.secondary)
        }
    }

    private func sync() async {
        isSyncing = true
        defer { isSyncing = false }
        do {
            try await viewModel.api.syncFreshRSS()
            viewModel.statusMessage = t("setting.freshrss.syncStarted")
            freshRSSStatus = try? await viewModel.api.fetchFreshRSSStatus()
            viewModel.refreshFeeds()
        } catch {
            viewModel.errorMessage = error.localizedDescription
        }
    }

    private func testRSSHub() async {
        isTestingRSSHub = true
        defer { isTestingRSSHub = false }
        do {
            rsshubMessage = try await viewModel.api.testRSSHubConnection()
        } catch {
            rsshubMessage = "\(t("setting.rsshub.connectionFailed")): \(error.localizedDescription)"
        }
    }
}
