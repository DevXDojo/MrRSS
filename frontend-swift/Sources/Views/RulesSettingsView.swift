import SwiftUI

struct RulesSettingsView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var editingRule: AutomationRule?

    var body: some View {
        VStack(spacing: 0) {
            HStack {
                VStack(alignment: .leading, spacing: 3) {
                    Text("Automation Rules")
                        .font(.title2.weight(.semibold))
                    Text("Match incoming or existing articles and apply actions automatically.")
                        .foregroundStyle(.secondary)
                }
                Spacer()
                Button {
                    editingRule = .empty()
                } label: {
                    Label("Add Rule", systemImage: "plus")
                }
                .buttonStyle(.borderedProminent)
            }
            .padding()

            Divider()

            if viewModel.rules.isEmpty {
                ContentUnavailableView {
                    Label("No Rules", systemImage: "bolt.badge.clock")
                } description: {
                    Text("Create a rule to favorite, hide, mark, or defer matching articles.")
                } actions: {
                    Button("Create Rule") { editingRule = .empty() }
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                List {
                    ForEach(viewModel.rules) { rule in
                        RuleRow(
                            rule: rule,
                            onToggle: { enabled in update(rule, enabled: enabled) },
                            onEdit: { editingRule = rule },
                            onApply: { Task { await viewModel.applyRule(rule) } },
                            onDelete: { delete(rule) }
                        )
                    }
                }
            }
        }
        .sheet(item: $editingRule) { rule in
            RuleEditorView(rule: rule) { savedRule in
                save(savedRule)
            }
        }
    }

    private func update(_ rule: AutomationRule, enabled: Bool) {
        var updated = rule
        updated.enabled = enabled
        save(updated)
    }

    private func save(_ rule: AutomationRule) {
        var rules = viewModel.rules
        if let index = rules.firstIndex(where: { $0.id == rule.id }) {
            rules[index] = rule
        } else {
            rules.append(rule)
        }
        Task { await viewModel.saveRules(rules) }
    }

    private func delete(_ rule: AutomationRule) {
        Task { await viewModel.saveRules(viewModel.rules.filter { $0.id != rule.id }) }
    }
}

private struct RuleRow: View {
    let rule: AutomationRule
    let onToggle: (Bool) -> Void
    let onEdit: () -> Void
    let onApply: () -> Void
    let onDelete: () -> Void

    var body: some View {
        HStack(spacing: 12) {
            Toggle("", isOn: Binding(get: { rule.enabled }, set: onToggle))
                .labelsHidden()

            VStack(alignment: .leading, spacing: 4) {
                Text(rule.name)
                    .font(.headline)
                Text("\(rule.conditions.count) conditions · \(rule.actions.count) actions")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            Button("Apply Now", action: onApply)
            Button("Edit", action: onEdit)
            Menu {
                Button("Delete", role: .destructive, action: onDelete)
            } label: {
                Image(systemName: "ellipsis.circle")
            }
            .menuStyle(.borderlessButton)
            .fixedSize()
        }
        .padding(.vertical, 5)
    }
}

private struct RuleEditorView: View {
    @Environment(\.dismiss) private var dismiss
    @State var rule: AutomationRule
    let onSave: (AutomationRule) -> Void

    private let fields = [
        ("article_title", "Article title"),
        ("feed_name", "Feed name"),
        ("feed_category", "Feed category"),
        ("feed_type", "Feed type"),
        ("is_freshrss_feed", "FreshRSS feed"),
        ("is_image_mode_feed", "Image-mode feed"),
        ("published_after", "Published after"),
        ("published_before", "Published before"),
        ("is_read", "Read state"),
        ("is_favorite", "Favorite state"),
        ("is_hidden", "Hidden state"),
        ("is_read_later", "Read-later state")
    ]

    private let actions = [
        ("favorite", "Add favorite"),
        ("unfavorite", "Remove favorite"),
        ("hide", "Hide"),
        ("unhide", "Unhide"),
        ("mark_read", "Mark read"),
        ("mark_unread", "Mark unread"),
        ("read_later", "Add to read later"),
        ("remove_read_later", "Remove from read later")
    ]

    var body: some View {
        NavigationStack {
            Form {
                Section("Rule") {
                    TextField("Name", text: $rule.name)
                    Toggle("Enabled", isOn: $rule.enabled)
                }

                Section("Conditions") {
                    if rule.conditions.isEmpty {
                        Text("No conditions means that every article matches.")
                            .foregroundStyle(.secondary)
                    }

                    ForEach($rule.conditions) { $condition in
                        VStack(alignment: .leading, spacing: 8) {
                            HStack {
                                if condition.id != rule.conditions.first?.id {
                                    Picker("Logic", selection: Binding(
                                        get: { condition.logic ?? "and" },
                                        set: { condition.logic = $0 }
                                    )) {
                                        Text("AND").tag("and")
                                        Text("OR").tag("or")
                                    }
                                    .labelsHidden()
                                    .frame(width: 76)
                                }

                                Toggle("Not", isOn: $condition.negate)
                                    .toggleStyle(.checkbox)

                                Picker("Field", selection: $condition.field) {
                                    ForEach(fields, id: \.0) { value, title in
                                        Text(title).tag(value)
                                    }
                                }

                                Button(role: .destructive) {
                                    rule.conditions.removeAll { $0.id == condition.id }
                                } label: {
                                    Image(systemName: "trash")
                                }
                                .buttonStyle(.borderless)
                            }

                            conditionValueEditor($condition)
                        }
                        .padding(.vertical, 5)
                    }

                    Button {
                        var condition = RuleCondition.empty(id: nextConditionID)
                        condition.logic = rule.conditions.isEmpty ? nil : "and"
                        rule.conditions.append(condition)
                    } label: {
                        Label("Add Condition", systemImage: "plus")
                    }
                }

                Section("Actions") {
                    ForEach(actions, id: \.0) { value, title in
                        Toggle(title, isOn: Binding(
                            get: { rule.actions.contains(value) },
                            set: { enabled in
                                if enabled {
                                    if !rule.actions.contains(value) { rule.actions.append(value) }
                                } else {
                                    rule.actions.removeAll { $0 == value }
                                }
                            }
                        ))
                    }
                }
            }
            .formStyle(.grouped)
            .navigationTitle("Edit Rule")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        normalizeConditions()
                        onSave(rule)
                        dismiss()
                    }
                    .disabled(rule.name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || rule.actions.isEmpty)
                }
            }
        }
        .frame(minWidth: 680, minHeight: 580)
    }

    @ViewBuilder
    private func conditionValueEditor(_ condition: Binding<RuleCondition>) -> some View {
        switch condition.wrappedValue.field {
        case "article_title":
            HStack {
                Picker("Operator", selection: condition.operator) {
                    Text("Contains").tag("contains")
                    Text("Equals").tag("exact")
                    Text("Regular expression").tag("regex")
                }
                .frame(width: 180)
                TextField("Value", text: condition.value)
            }
        case "feed_name", "feed_category":
            TextField("Comma-separated values", text: Binding(
                get: { condition.wrappedValue.values.joined(separator: ", ") },
                set: { condition.wrappedValue.values = splitValues($0) }
            ))
        case "is_freshrss_feed", "is_image_mode_feed", "is_read", "is_favorite", "is_hidden", "is_read_later":
            Picker("Value", selection: condition.value) {
                Text("Yes").tag("true")
                Text("No").tag("false")
            }
        default:
            TextField(condition.wrappedValue.field.hasPrefix("published_") ? "YYYY-MM-DD" : "Value", text: condition.value)
        }
    }

    private var nextConditionID: Int {
        max((rule.conditions.map(\.id).max() ?? 0) + 1, Int(Date().timeIntervalSince1970 * 1_000))
    }

    private func splitValues(_ text: String) -> [String] {
        text.split(separator: ",")
            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
            .filter { !$0.isEmpty }
    }

    private func normalizeConditions() {
        for index in rule.conditions.indices {
            rule.conditions[index].logic = index == 0 ? nil : (rule.conditions[index].logic ?? "and")
            if rule.conditions[index].field == "article_title", rule.conditions[index].operator.isEmpty {
                rule.conditions[index].operator = "contains"
            }
        }
    }
}
