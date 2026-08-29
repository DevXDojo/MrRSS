import SwiftUI

/// Manages the saved AI configurations, so each feature can point at a
/// different provider the way the previous interface allowed.
struct AIProfilesView: View {
    @ObservedObject var viewModel: AppViewModel

    @State private var profiles: [AIProfile] = []
    @State private var editing: AIProfile?
    @State private var isCreating = false
    @State private var testResults: [Int: AIProfileTestResult] = [:]
    @State private var isTesting = false
    @State private var profilePendingDeletion: AIProfile?

    var body: some View {
        Section(t("setting.ai.aiProfiles")) {
            if profiles.isEmpty {
                VStack(alignment: .leading, spacing: 4) {
                    Text(t("setting.ai.noProfiles"))
                    Text(t("setting.ai.noProfilesHint"))
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }

            ForEach(profiles) { profile in
                row(for: profile)
            }

            HStack {
                Button(t("setting.ai.addProfile")) { isCreating = true }
                Button(t("setting.ai.testAllProfiles")) {
                    Task { await testAll() }
                }
                .disabled(profiles.isEmpty || isTesting)
                if isTesting {
                    ProgressView().controlSize(.small)
                }
            }
        }
        .task { await load() }
        .sheet(item: $editing) { profile in
            AIProfileEditorView(profile: profile, viewModel: viewModel) {
                Task { await load() }
            }
        }
        .sheet(isPresented: $isCreating) {
            AIProfileEditorView(profile: nil, viewModel: viewModel) {
                Task { await load() }
            }
        }
        .confirmationDialog(
            t("setting.ai.deleteProfileTitle"),
            isPresented: Binding(
                get: { profilePendingDeletion != nil },
                set: { if !$0 { profilePendingDeletion = nil } }
            )
        ) {
            Button(t("setting.ai.deleteProfile"), role: .destructive) {
                guard let profile = profilePendingDeletion else { return }
                profilePendingDeletion = nil
                Task { await delete(profile) }
            }
            Button(t("common.cancel"), role: .cancel) { profilePendingDeletion = nil }
        } message: {
            Text(t("setting.ai.deleteProfileConfirm", ["name": profilePendingDeletion?.name ?? ""]))
        }

        if let usage = viewModel.aiUsage {
            Section(t("setting.ai.setUsageLimit")) {
                LabeledContent(t("setting.ai.setUsageLimit")) {
                    Text(usage.limit > 0 ? "\(usage.usage) / \(usage.limit)" : "\(usage.usage)")
                        .monospacedDigit()
                }
                if usage.limitReached {
                    Text(t("article.summary.aiLimitReached"))
                        .font(.caption)
                        .foregroundStyle(.orange)
                }
                Button(t("common.action.resetToDefault")) {
                    Task { await viewModel.resetAIUsage() }
                }
            }
        }
    }

    private func row(for profile: AIProfile) -> some View {
        HStack(spacing: 10) {
            VStack(alignment: .leading, spacing: 2) {
                HStack(spacing: 6) {
                    Text(profile.name.isEmpty ? profile.model : profile.name)
                        .fontWeight(.medium)
                    if profile.isDefault {
                        Text(t("common.action.resetToDefault"))
                            .font(.caption2)
                            .padding(.horizontal, 5)
                            .padding(.vertical, 1)
                            .background(Color.accentColor.opacity(0.15), in: Capsule())
                    }
                }
                Text(profile.model)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            if let result = testResults[profile.id] {
                Image(systemName: result.succeeded ? "checkmark.circle.fill" : "xmark.circle.fill")
                    .foregroundStyle(result.succeeded ? .green : .red)
                    .help(result.errorMessage ?? t("common.connectionSuccessful"))
            }

            Button(t("setting.ai.editProfile")) { editing = profile }
                .buttonStyle(.borderless)
            Button(t("setting.ai.deleteProfile"), role: .destructive) {
                profilePendingDeletion = profile
            }
            .buttonStyle(.borderless)
        }
    }

    private func load() async {
        profiles = (try? await viewModel.api.fetchAIProfiles()) ?? []
    }

    private func testAll() async {
        isTesting = true
        defer { isTesting = false }
        let results = (try? await viewModel.api.testAIProfiles()) ?? []
        testResults = Dictionary(uniqueKeysWithValues: results.map { ($0.profileID, $0) })
    }

    private func delete(_ profile: AIProfile) async {
        do {
            try await viewModel.api.deleteAIProfile(id: profile.id)
            viewModel.statusMessage = t("setting.ai.profileDeleted")
            await load()
        } catch {
            viewModel.errorMessage = "\(t("setting.ai.deleteProfileFailed")): \(error.localizedDescription)"
        }
    }
}

/// Creates or edits one AI configuration.
struct AIProfileEditorView: View {
    let profile: AIProfile?
    @ObservedObject var viewModel: AppViewModel
    let onSave: () -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var draft = AIProfile()
    @State private var isSaving = false

    var body: some View {
        VStack(spacing: 0) {
            Text(profile == nil ? t("setting.ai.addProfile") : t("setting.ai.editProfile"))
                .font(.title2.bold())
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(20)

            Divider()

            Form {
                LabeledContent(t("setting.ai.profileName")) {
                    TextField("", text: $draft.name, prompt: Text(t("setting.ai.profileNamePlaceholder")))
                        .textFieldStyle(.roundedBorder)
                }
                LabeledContent(t("setting.ai.aiEndpoint")) {
                    TextField("", text: $draft.endpoint, prompt: Text("https://api.openai.com/v1"))
                        .textFieldStyle(.roundedBorder)
                }
                LabeledContent(t("setting.ai.aiModel")) {
                    TextField("", text: $draft.model, prompt: Text("gpt-4o-mini"))
                        .textFieldStyle(.roundedBorder)
                }
                LabeledContent(t("setting.ai.aiApiKey")) {
                    SecureField("", text: $draft.apiKey)
                        .textFieldStyle(.roundedBorder)
                }
                LabeledContent(t("setting.ai.aiCustomHeaders")) {
                    TextField("", text: $draft.customHeaders, prompt: Text("{}"))
                        .textFieldStyle(.roundedBorder)
                }
                Toggle(t("common.action.resetToDefault"), isOn: $draft.isDefault)
            }
            .formStyle(.grouped)

            Divider()

            HStack {
                Spacer()
                Button(t("common.cancel"), role: .cancel) { dismiss() }
                Button(t("common.action.save")) {
                    Task { await save() }
                }
                .keyboardShortcut(.defaultAction)
                .disabled(isSaving || draft.name.trimmingCharacters(in: .whitespaces).isEmpty)
            }
            .padding(18)
        }
        .frame(width: 560, height: 460)
        .onAppear { draft = profile ?? AIProfile() }
    }

    private func save() async {
        isSaving = true
        defer { isSaving = false }
        do {
            _ = try await viewModel.api.saveAIProfile(draft)
            viewModel.statusMessage = profile == nil
                ? t("setting.ai.profileCreated")
                : t("setting.ai.profileUpdated")
            onSave()
            dismiss()
        } catch {
            viewModel.errorMessage = error.localizedDescription
        }
    }
}
