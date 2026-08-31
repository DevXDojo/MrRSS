import AppKit
import SwiftUI

/// Version information and the update check.
struct AboutView: View {
    @ObservedObject var viewModel: AppViewModel

    @State private var backendVersion = ""
    @State private var update: UpdateInfo?
    @State private var isChecking = false
    @State private var checkError: String?

    var body: some View {
        Form {
            Section {
                LabeledContent(t("setting.about.version")) {
                    Text(backendVersion.isEmpty ? "—" : backendVersion)
                }
                LabeledContent(t("client.connection.serverAddress")) {
                    Text(viewModel.api.baseURL.absoluteString)
                        .textSelection(.enabled)
                        .foregroundStyle(.secondary)
                }
            } header: {
                Text("MrRSS")
            }

            Section(t("setting.update.updates")) {
                if let update {
                    LabeledContent(t("setting.update.currentVersion")) {
                        Text(update.currentVersion)
                    }
                    LabeledContent(t("setting.update.latestVersion")) {
                        Text(update.latestVersion)
                    }
                    if update.hasUpdate {
                        Text(t("setting.update.updateAvailable"))
                            .foregroundStyle(.orange)
                        if let downloadURL = update.downloadURL, let url = URL(string: downloadURL) {
                            Button(t("modal.update.downloadUpdate")) {
                                NSWorkspace.shared.open(url)
                            }
                        }
                    } else {
                        Text(t("setting.update.upToDate"))
                            .foregroundStyle(.secondary)
                    }
                }

                if let checkError {
                    Text(checkError)
                        .font(.caption)
                        .foregroundStyle(.red)
                }

                HStack {
                    Button(t("setting.update.checkForUpdates")) {
                        Task { await checkForUpdates() }
                    }
                    .disabled(isChecking)
                    if isChecking {
                        ProgressView().controlSize(.small)
                    }
                }
            }

            Section {
                Button(t("setting.about.viewOnGitHub")) {
                    if let url = URL(string: "https://github.com/DevXDojo/MrRSS") {
                        NSWorkspace.shared.open(url)
                    }
                }
            }
        }
        .formStyle(.grouped)
        .task {
            backendVersion = (try? await viewModel.api.fetchVersion()) ?? ""
        }
    }

    private func checkForUpdates() async {
        isChecking = true
        checkError = nil
        defer { isChecking = false }
        do {
            update = try await viewModel.api.checkForUpdates()
            if let error = update?.error {
                checkError = error
            }
        } catch {
            checkError = "\(t("common.errors.errorCheckingUpdates")): \(error.localizedDescription)"
        }
    }
}
