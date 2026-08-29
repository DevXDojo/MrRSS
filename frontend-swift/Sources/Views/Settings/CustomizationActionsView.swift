import AppKit
import SwiftUI

/// The custom stylesheet and the fetch scripts, alongside the generated
/// customization settings.
struct CustomizationActionsView: View {
    @ObservedObject var viewModel: AppViewModel

    @State private var scripts = ScriptList.empty
    @State private var isUploading = false
    @State private var isConfirmingDelete = false

    var body: some View {
        Section(t("setting.customization.css")) {
            Text(t("setting.customization.cssDesc"))
                .font(.caption)
                .foregroundStyle(.secondary)
                .fixedSize(horizontal: false, vertical: true)

            let current = viewModel.setting("custom_css_file")
            if !current.isEmpty {
                LabeledContent(t("setting.customization.cssApplied")) {
                    Text((current as NSString).lastPathComponent)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }
            }

            HStack(spacing: 10) {
                Button(t("setting.customization.cssUpload")) {
                    uploadCSS()
                }
                .disabled(isUploading)

                if !current.isEmpty {
                    Button(t("setting.customization.deleteCSS"), role: .destructive) {
                        isConfirmingDelete = true
                    }
                }

                if isUploading {
                    ProgressView().controlSize(.small)
                }
            }
        }
        .confirmationDialog(
            t("setting.customization.deleteCSS"),
            isPresented: $isConfirmingDelete
        ) {
            Button(t("common.delete"), role: .destructive) {
                Task { await deleteCSS() }
            }
            Button(t("common.cancel"), role: .cancel) {}
        }

        Section(t("setting.customization.script")) {
            if scripts.scripts.isEmpty {
                Text(t("setting.customization.scriptsNotFound"))
                    .font(.caption)
                    .foregroundStyle(.secondary)
            } else {
                ForEach(scripts.scripts, id: \.self) { script in
                    Text(script)
                        .font(.callout)
                        .textSelection(.enabled)
                }
            }

            LabeledContent(t("setting.customization.scriptsFolder")) {
                Button(t("client.action.showInFinder")) {
                    guard !scripts.scriptsDir.isEmpty else { return }
                    NSWorkspace.shared.selectFile(nil, inFileViewerRootedAtPath: scripts.scriptsDir)
                }
                .disabled(scripts.scriptsDir.isEmpty)
            }
        }
        .task { scripts = (try? await viewModel.api.fetchScripts()) ?? .empty }
    }

    private func uploadCSS() {
        let panel = NSOpenPanel()
        panel.allowedContentTypes = [.init(filenameExtension: "css") ?? .plainText]
        panel.allowsMultipleSelection = false
        panel.prompt = t("setting.customization.cssUpload")
        guard panel.runModal() == .OK, let url = panel.url else { return }

        isUploading = true
        Task {
            defer { isUploading = false }
            do {
                let data = try Data(contentsOf: url)
                try await viewModel.api.uploadCustomCSS(data: data, filename: url.lastPathComponent)
                viewModel.statusMessage = t("setting.customization.cssUploaded")
                await viewModel.loadSettings()
            } catch {
                viewModel.errorMessage =
                    "\(t("setting.customization.cssUploadFailed")): \(error.localizedDescription)"
            }
        }
    }

    private func deleteCSS() async {
        do {
            try await viewModel.api.deleteCustomCSS()
            viewModel.statusMessage = t("setting.customization.cssDeleted")
            await viewModel.loadSettings()
        } catch {
            viewModel.errorMessage =
                "\(t("setting.customization.cssDeleteFailed")): \(error.localizedDescription)"
        }
    }
}

extension ScriptList {
    static let empty = ScriptList(scripts: [], scriptsDir: "")
}
