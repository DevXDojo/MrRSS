import SwiftUI

/// The settings window. Most panes are built from the schema the backend
/// publishes, so a new setting appears here as soon as it is generated.
struct SettingsRootView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var pane: SettingsPane = .connection
    @State private var search = ""

    var body: some View {
        NavigationSplitView {
            List(SettingsPane.allCases, selection: $pane) { entry in
                Label(entry.title, systemImage: entry.icon).tag(entry)
            }
            .navigationSplitViewColumnWidth(min: 190, ideal: 210, max: 250)
        } detail: {
            detail
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        }
        .frame(minWidth: 820, minHeight: 560)
        .searchable(text: $search, prompt: Text(t("client.settings.searchSettings")))
        .toolbar { toolbarContent }
        .task { await viewModel.loadSettings() }
    }

    @ViewBuilder
    private var detail: some View {
        if !search.trimmingCharacters(in: .whitespaces).isEmpty {
            searchResults
        } else {
            switch pane {
            case .connection:
                ServerSettingsView(viewModel: viewModel)
            case .rules:
                RulesSettingsView(viewModel: viewModel)
            case .statistics:
                StatisticsView(viewModel: viewModel)
            case .about:
                AboutView(viewModel: viewModel)
            case .storage:
                SettingListView(pane: pane, viewModel: viewModel) {
                    StorageActionsView(viewModel: viewModel)
                }
            case .integrations:
                SettingListView(pane: pane, viewModel: viewModel) {
                    IntegrationActionsView(viewModel: viewModel)
                }
            case .ai:
                SettingListView(pane: pane, viewModel: viewModel) {
                    AIProfilesView(viewModel: viewModel)
                }
            default:
                SettingListView(pane: pane, viewModel: viewModel) { EmptyView() }
            }
        }
    }

    private var searchResults: some View {
        let matches = SettingsCatalog.search(search)
        return Group {
            if matches.isEmpty {
                ContentUnavailableView(
                    t("client.settings.noResults"),
                    systemImage: "magnifyingglass"
                )
            } else {
                Form {
                    ForEach(matches) { definition in
                        SettingRow(definition: definition, viewModel: viewModel)
                    }
                }
                .formStyle(.grouped)
            }
        }
    }

    @ToolbarContentBuilder
    private var toolbarContent: some ToolbarContent {
        ToolbarItemGroup(placement: .primaryAction) {
            if viewModel.isSavingSettings {
                ProgressView().controlSize(.small)
            }
            Button(t("common.action.save")) {
                Task { await viewModel.saveSettings() }
            }
            .keyboardShortcut("s", modifiers: .command)
            .disabled(viewModel.isSavingSettings)
        }
    }
}

/// A pane built from the generated catalogue, with room for extra controls.
struct SettingListView<Extra: View>: View {
    let pane: SettingsPane
    @ObservedObject var viewModel: AppViewModel
    @ViewBuilder let extra: () -> Extra

    var body: some View {
        Form {
            ForEach(SettingsCatalog.definitions(for: pane)) { definition in
                SettingRow(definition: definition, viewModel: viewModel)
            }
            extra()
        }
        .formStyle(.grouped)
    }
}

/// One setting, drawn according to the control the schema calls for.
struct SettingRow: View {
    let definition: SettingDefinition
    @ObservedObject var viewModel: AppViewModel

    var body: some View {
        switch definition.control {
        case .toggle:
            row {
                Toggle(definition.title, isOn: boolBinding)
                    .toggleStyle(.switch)
            }
        case .text:
            row {
                LabeledContent(definition.title) {
                    TextField("", text: stringBinding)
                        .textFieldStyle(.roundedBorder)
                        .frame(maxWidth: 320)
                }
            }
        case .secret:
            row {
                LabeledContent(definition.title) {
                    SecureField("", text: stringBinding)
                        .textFieldStyle(.roundedBorder)
                        .frame(maxWidth: 320)
                }
            }
        case .number:
            row {
                LabeledContent(definition.title) {
                    TextField("", text: stringBinding)
                        .textFieldStyle(.roundedBorder)
                        .frame(width: 110)
                }
            }
        case .choice:
            row {
                Picker(definition.title, selection: stringBinding) {
                    ForEach(definition.choices) { choice in
                        Text(choice.title).tag(choice.value)
                    }
                }
            }
        }
    }

    @ViewBuilder
    private func row<Content: View>(@ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: 3) {
            content()
            if let detail = definition.detail {
                Text(detail)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }

    private var stringBinding: Binding<String> {
        Binding(
            get: { viewModel.setting(definition.key, default: definition.defaultValue) },
            set: { viewModel.updateSetting(definition.key, value: $0) }
        )
    }

    private var boolBinding: Binding<Bool> {
        Binding(
            get: {
                viewModel.boolSetting(definition.key, default: definition.defaultValue == "true")
            },
            set: { viewModel.updateBoolSetting(definition.key, value: $0) }
        )
    }
}
