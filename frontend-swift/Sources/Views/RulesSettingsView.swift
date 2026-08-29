import SwiftUI

struct RulesSettingsView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var editingRule: AutomationRule?

    var body: some View {
        VStack(spacing: 0) {
            HStack {
                VStack(alignment: .leading, spacing: 3) {
                    Text(t("modal.rule.rules"))
                        .font(.title2.weight(.semibold))
                    Text(t("modal.rule.rulesDesc"))
                        .foregroundStyle(.secondary)
                }
                Spacer()
                Button {
                    editingRule = .empty()
                } label: {
                    Label(t("modal.rule.addRule"), systemImage: "plus")
                }
                .buttonStyle(.borderedProminent)
            }
            .padding()

            Divider()

            if viewModel.rules.isEmpty {
                ContentUnavailableView {
                    Label(t("setting.rule.noRules"), systemImage: "bolt.badge.clock")
                } description: {
                    Text(t("setting.rule.noRulesHint"))
                } actions: {
                    Button(t("setting.rule.addRule")) { editingRule = .empty() }
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

    /// How many conditions and actions the rule carries.
    private var summary: String {
        t("client.rule.summary", ["conditions": rule.conditions.count, "actions": rule.actions.count])
    }

    var body: some View {
        HStack(spacing: 12) {
            Toggle("", isOn: Binding(get: { rule.enabled }, set: onToggle))
                .labelsHidden()

            VStack(alignment: .leading, spacing: 4) {
                Text(rule.name)
                    .font(.headline)
                Text(summary)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            Button(t("setting.rule.applyRuleNow"), action: onApply)
            Button(t("modal.rule.editRule"), action: onEdit)
            Menu {
                Button(t("common.delete"), role: .destructive, action: onDelete)
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

    private var fields: [(String, String)] {
        [
            ("article_title", t("common.form.title")),
            ("feed_name", t("modal.filter.fromFeed")),
            ("feed_category", t("sidebar.sort.byCategory")),
            ("feed_type", t("modal.filter.feedType")),
            ("is_freshrss_feed", t("modal.feed.typeFreshRSS")),
            ("is_image_mode_feed", t("modal.filter.isImageModeFeed")),
            ("published_after", t("modal.filter.publishedAfter")),
            ("published_before", t("modal.filter.publishedBefore")),
            ("is_read", t("modal.filter.readStatus")),
            ("is_favorite", t("modal.filter.favoriteStatus")),
            ("is_hidden", t("modal.filter.hiddenStatus")),
            ("is_read_later", t("modal.filter.readLaterStatus"))
        ]
    }

    private var actions: [(String, String)] {
        [
            ("favorite", t("setting.rule.actionFavorite")),
            ("unfavorite", t("setting.rule.actionUnfavorite")),
            ("hide", t("setting.rule.actionHide")),
            ("unhide", t("setting.rule.actionUnhide")),
            ("mark_read", t("setting.rule.actionMarkRead")),
            ("mark_unread", t("setting.rule.actionMarkUnread")),
            ("read_later", t("setting.rule.actionReadLater")),
            ("remove_read_later", t("setting.rule.actionRemoveReadLater"))
        ]
    }

    var body: some View {
        NavigationStack {
            Form {
                Section(t("modal.rule.name")) {
                    TextField(t("modal.rule.name"), text: $rule.name, prompt: Text(t("modal.rule.namePlaceholder")))
                    Toggle(t("client.rule.enabled"), isOn: $rule.enabled)
                }

                Section(t("modal.filter.filterConditions")) {
                    if rule.conditions.isEmpty {
                        Text(t("modal.filter.conditionAlways"))
                            .foregroundStyle(.secondary)
                    }

                    ForEach($rule.conditions) { $condition in
                        VStack(alignment: .leading, spacing: 8) {
                            HStack {
                                if condition.id != rule.conditions.first?.id {
                                    Picker(t("modal.filter.and"), selection: Binding(
                                        get: { condition.logic ?? "and" },
                                        set: { condition.logic = $0 }
                                    )) {
                                        Text(t("modal.filter.and")).tag("and")
                                        Text(t("modal.filter.or")).tag("or")
                                    }
                                    .labelsHidden()
                                    .frame(width: 76)
                                }

                                Toggle(t("modal.filter.not"), isOn: $condition.negate)
                                    .toggleStyle(.checkbox)

                                Picker(t("modal.filter.filterField"), selection: $condition.field) {
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
                        Label(t("modal.rule.addCondition"), systemImage: "plus")
                    }
                }

                Section(t("modal.rule.actions")) {
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
            .navigationTitle(t("modal.rule.editRule"))
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(t("common.cancel")) { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(t("common.action.save")) {
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
                Picker(t("modal.filter.filterOperator"), selection: condition.operator) {
                    Text(t("modal.filter.contains")).tag("contains")
                    Text(t("modal.filter.exactMatch")).tag("exact")
                    Text(t("modal.filter.regex")).tag("regex")
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
            Picker(t("modal.filter.filterValue"), selection: condition.value) {
                Text(t("common.action.yes")).tag("true")
                Text(t("common.action.no")).tag("false")
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
