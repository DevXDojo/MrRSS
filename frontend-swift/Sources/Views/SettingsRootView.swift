import SwiftUI

struct SettingsRootView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var selectedPane: SettingsPane? = .general

    var body: some View {
        NavigationSplitView {
            List(SettingsPane.allCases, selection: $selectedPane) { pane in
                Label(pane.title, systemImage: pane.icon)
                    .tag(pane)
            }
            .navigationTitle("Settings")
            .navigationSplitViewColumnWidth(min: 165, ideal: 185, max: 220)
        } detail: {
            detailView
                .navigationTitle(selectedPane?.title ?? "Settings")
                .toolbar {
                    ToolbarItem(placement: .confirmationAction) {
                        Button("Save") {
                            Task { await viewModel.saveSettings() }
                        }
                        .disabled(viewModel.isSavingSettings || selectedPane == .connection || selectedPane == .rules)
                    }
                }
        }
        .frame(minWidth: 760, idealWidth: 860, minHeight: 540, idealHeight: 620)
        .task {
            if viewModel.settings.isEmpty {
                await viewModel.loadSettings()
            }
        }
        .overlay(alignment: .bottom) {
            if let message = viewModel.statusMessage {
                Text(message)
                    .font(.callout)
                    .padding(.horizontal, 14)
                    .padding(.vertical, 9)
                    .background(.regularMaterial, in: Capsule())
                    .padding(.bottom, 12)
                    .task(id: message) {
                        try? await Task.sleep(for: .seconds(3))
                        viewModel.clearStatusMessage()
                    }
            }
        }
    }

    @ViewBuilder
    private var detailView: some View {
        switch selectedPane ?? .general {
        case .connection:
            ServerSettingsView(viewModel: viewModel)
        case .rules:
            RulesSettingsView(viewModel: viewModel)
        case .advanced:
            AllSettingsView(viewModel: viewModel)
        case let pane:
            DynamicSettingsForm(viewModel: viewModel, definitions: SettingsCatalog.definitions(for: pane))
        }
    }
}

private struct DynamicSettingsForm: View {
    @ObservedObject var viewModel: AppViewModel
    let definitions: [SettingDefinition]

    private var sections: [String] {
        definitions.reduce(into: []) { result, definition in
            if !result.contains(definition.section) { result.append(definition.section) }
        }
    }

    var body: some View {
        Form {
            ForEach(sections, id: \.self) { section in
                Section(section) {
                    ForEach(definitions.filter { $0.section == section }) { definition in
                        settingControl(definition)
                    }
                }
            }

            if definitions.contains(where: { $0.pane == .summaryAI }) {
                Section("AI usage") {
                    if let usage = viewModel.aiUsage {
                        LabeledContent("Tokens", value: "\(usage.usage) / \(usage.limit)")
                        ProgressView(value: Double(usage.usage), total: Double(max(usage.limit, 1)))
                        Button("Reset usage counter") {
                            Task { await viewModel.resetAIUsage() }
                        }
                    } else {
                        Text("Usage information is unavailable.")
                            .foregroundStyle(.secondary)
                    }
                }

                Section("Generated content") {
                    Button("Clear cached translations and summaries", role: .destructive) {
                        Task { await viewModel.clearGeneratedContent() }
                    }
                }
            }
        }
        .formStyle(.grouped)
        .overlay {
            if viewModel.isLoadingSettings {
                ProgressView("Loading settings")
            }
        }
    }

    @ViewBuilder
    private func settingControl(_ definition: SettingDefinition) -> some View {
        switch definition.control {
        case .toggle:
            Toggle(definition.title, isOn: boolBinding(definition.key))
        case .text:
            LabeledContent(definition.title) {
                TextField(definition.title, text: textBinding(definition.key))
                    .multilineTextAlignment(.trailing)
                    .frame(minWidth: 240)
            }
        case .secure:
            LabeledContent(definition.title) {
                SecureField(definition.title, text: textBinding(definition.key))
                    .multilineTextAlignment(.trailing)
                    .frame(minWidth: 240)
            }
        case .picker(let options):
            Picker(definition.title, selection: textBinding(definition.key)) {
                ForEach(options, id: \.0) { value, label in
                    Text(label).tag(value)
                }
            }
        }

        if let detail = definition.detail {
            Text(detail)
                .font(.caption)
                .foregroundStyle(.secondary)
        }
    }

    private func textBinding(_ key: String) -> Binding<String> {
        Binding(
            get: { viewModel.setting(key) },
            set: { viewModel.updateSetting(key, value: $0) }
        )
    }

    private func boolBinding(_ key: String) -> Binding<Bool> {
        Binding(
            get: { viewModel.boolSetting(key) },
            set: { viewModel.updateBoolSetting(key, value: $0) }
        )
    }
}

private struct AllSettingsView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var searchText = ""

    private var keys: [String] {
        viewModel.settings.keys
            .filter { searchText.isEmpty || $0.localizedCaseInsensitiveContains(searchText) }
            .sorted()
    }

    var body: some View {
        Form {
            Section {
                Text("This pane exposes every backend setting, including internal state. Use the categorized panes for normal configuration.")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Section("Backend settings") {
                ForEach(keys, id: \.self) { key in
                    if SettingsCatalog.boolKeys.contains(key) {
                        Toggle(key, isOn: Binding(
                            get: { viewModel.boolSetting(key) },
                            set: { viewModel.updateBoolSetting(key, value: $0) }
                        ))
                    } else if SettingsCatalog.secureKeys.contains(key) {
                        LabeledContent(key) {
                            SecureField(key, text: valueBinding(key))
                                .multilineTextAlignment(.trailing)
                                .frame(minWidth: 300)
                        }
                    } else {
                        LabeledContent(key) {
                            TextField(key, text: valueBinding(key), axis: .vertical)
                                .multilineTextAlignment(.trailing)
                                .lineLimit(1...5)
                                .frame(minWidth: 300)
                        }
                    }
                }
            }
        }
        .formStyle(.grouped)
        .searchable(text: $searchText, prompt: "Filter setting keys")
    }

    private func valueBinding(_ key: String) -> Binding<String> {
        Binding(
            get: { viewModel.setting(key) },
            set: { viewModel.updateSetting(key, value: $0) }
        )
    }
}
