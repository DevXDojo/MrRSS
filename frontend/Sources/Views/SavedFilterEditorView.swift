import SwiftUI

/// Creates or edits a saved filter, using the same fields and operators the
/// backend evaluates.
struct SavedFilterEditorView: View {
    let filter: SavedFilter?
    @ObservedObject var viewModel: AppViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var name = ""
    @State private var conditions: [FilterCondition] = []
    @State private var isSaving = false

    var body: some View {
        VStack(spacing: 0) {
            header
            Divider()

            ScrollView {
                VStack(alignment: .leading, spacing: 14) {
                    VStack(alignment: .leading, spacing: 5) {
                        Text(t("sidebar.savedFilters.filterName"))
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                        TextField(
                            "",
                            text: $name,
                            prompt: Text(t("sidebar.savedFilters.filterNamePlaceholder"))
                        )
                        .textFieldStyle(.roundedBorder)
                    }

                    Text(t("modal.filter.filterConditions"))
                        .font(.headline)

                    if conditions.isEmpty {
                        Text(t("modal.filter.noFiltersApplied"))
                            .font(.callout)
                            .foregroundStyle(.secondary)
                    }

                    ForEach($conditions) { $condition in
                        ConditionEditor(
                            condition: $condition,
                            isFirst: conditions.first?.id == condition.id,
                            onDelete: { conditions.removeAll { $0.id == condition.id } }
                        )
                    }

                    Button {
                        conditions.append(FilterCondition(id: nextConditionID))
                    } label: {
                        Label(t("modal.filter.addCondition"), systemImage: "plus.circle")
                    }

                    Text(t("modal.filter.logicPrecedence"))
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                .padding(20)
            }

            Divider()
            footer
        }
        .frame(width: 640, height: 560)
        .onAppear(perform: load)
    }

    private var header: some View {
        Text(filter == nil ? t("sidebar.savedFilters.saveCurrentFilter") : t("sidebar.savedFilters.editFilter"))
            .font(.title2.bold())
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(20)
    }

    private var footer: some View {
        HStack {
            Spacer()
            Button(t("common.cancel"), role: .cancel) { dismiss() }
            Button(t("sidebar.savedFilters.save")) {
                Task { await save() }
            }
            .keyboardShortcut(.defaultAction)
            .disabled(isSaving || name.trimmingCharacters(in: .whitespaces).isEmpty)
        }
        .padding(18)
    }

    /// Condition identifiers only need to be unique inside one filter.
    private var nextConditionID: Int {
        (conditions.map(\.id).max() ?? 0) + 1
    }

    private func load() {
        name = filter?.name ?? ""
        conditions = filter?.conditions ?? []
    }

    private func save() async {
        guard !conditions.isEmpty else {
            viewModel.errorMessage = t("sidebar.savedFilters.conditionsRequired")
            return
        }

        isSaving = true
        defer { isSaving = false }

        let succeeded: Bool
        if var existing = filter {
            existing.name = name
            existing.conditions = conditions
            succeeded = await viewModel.updateSavedFilter(existing)
        } else {
            succeeded = await viewModel.createSavedFilter(name: name, conditions: conditions)
        }

        if succeeded { dismiss() }
    }
}

/// One clause of a saved filter.
private struct ConditionEditor: View {
    @Binding var condition: FilterCondition
    let isFirst: Bool
    let onDelete: () -> Void

    var body: some View {
        GroupBox {
            VStack(alignment: .leading, spacing: 8) {
                HStack(spacing: 8) {
                    if !isFirst {
                        Picker("", selection: $condition.logic) {
                            Text(t("modal.filter.and")).tag("and")
                            Text(t("modal.filter.or")).tag("or")
                        }
                        .labelsHidden()
                        .frame(width: 80)
                    }

                    Toggle(t("modal.filter.not"), isOn: $condition.negate)
                        .toggleStyle(.checkbox)

                    Spacer()

                    Button(role: .destructive, action: onDelete) {
                        Image(systemName: "trash")
                    }
                    .buttonStyle(.borderless)
                }

                Picker(t("modal.filter.filterField"), selection: fieldBinding) {
                    ForEach(FilterField.allCases) { field in
                        Text(field.title).tag(field)
                    }
                }

                if field.supportsOperators {
                    Picker(t("modal.filter.filterOperator"), selection: operatorBinding) {
                        ForEach(FilterOperator.allCases) { op in
                            Text(op.title).tag(op)
                        }
                    }
                }

                valueEditor
            }
            .padding(8)
        }
    }

    private var field: FilterField {
        FilterField(rawValue: condition.field) ?? .articleTitle
    }

    private var fieldBinding: Binding<FilterField> {
        Binding(
            get: { field },
            set: { newValue in
                condition.field = newValue.rawValue
                condition.value = ""
                condition.values = []
                condition.operator = newValue.supportsOperators ? "contains" : ""
            }
        )
    }

    private var operatorBinding: Binding<FilterOperator> {
        Binding(
            get: { FilterOperator(rawValue: condition.operator) ?? .contains },
            set: { condition.operator = $0.rawValue }
        )
    }

    @ViewBuilder
    private var valueEditor: some View {
        switch field.valueKind {
        case .text:
            TextField(t("modal.filter.filterValue"), text: $condition.value)
                .textFieldStyle(.roundedBorder)
        case .number:
            TextField(t("modal.filter.filterValue"), text: $condition.value)
                .textFieldStyle(.roundedBorder)
                .frame(width: 120)
        case .date:
            DatePicker(
                t("modal.filter.filterValue"),
                selection: dateBinding,
                displayedComponents: .date
            )
        case .boolean:
            Picker(t("modal.filter.filterValue"), selection: $condition.value) {
                Text(t("common.action.yes")).tag("true")
                Text(t("common.action.no")).tag("false")
            }
            .pickerStyle(.segmented)
        case .none:
            EmptyView()
        }
    }

    private var dateBinding: Binding<Date> {
        Binding(
            get: {
                let formatter = ISO8601DateFormatter()
                formatter.formatOptions = [.withFullDate]
                return formatter.date(from: condition.value) ?? Date()
            },
            set: { newValue in
                let formatter = ISO8601DateFormatter()
                formatter.formatOptions = [.withFullDate]
                condition.value = formatter.string(from: newValue)
            }
        )
    }
}
