import SwiftUI

/// Creates, renames, recolours and deletes the tags that feeds can carry.
struct TagManagerView: View {
    @ObservedObject var viewModel: AppViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var newName = ""
    @State private var newColor = Color.accentColor
    @State private var tagPendingDeletion: Tag?
    /// Names being typed, committed when the field is submitted or left, so a
    /// rename is one request rather than one per keystroke.
    @State private var draftNames: [Int: String] = [:]
    /// Colours being chosen. The picker reports every shade the pointer passes
    /// over, so the change is saved once the choice settles.
    @State private var draftColors: [Int: Color] = [:]
    @State private var colorSaveTasks: [Int: Task<Void, Never>] = [:]

    var body: some View {
        VStack(spacing: 0) {
            HStack {
                Text(t("modal.tag.manageTags"))
                    .font(.title2.bold())
                Spacer()
                Button(t("common.close")) { dismiss() }
                    .keyboardShortcut(.cancelAction)
            }
            .padding(20)

            Divider()

            if viewModel.tags.isEmpty {
                ContentUnavailableView(t("modal.tag.noTags"), systemImage: "tag")
                    .frame(maxHeight: .infinity)
            } else {
                List {
                    ForEach(viewModel.tags) { tag in
                        row(for: tag)
                    }
                }
                .listStyle(.inset)
            }

            Divider()

            HStack(spacing: 10) {
                ColorPicker("", selection: $newColor, supportsOpacity: false)
                    .labelsHidden()
                    .frame(width: 44)
                TextField(t("modal.tag.name"), text: $newName)
                    .textFieldStyle(.roundedBorder)
                Button(t("modal.tag.createTag")) {
                    Task { await create() }
                }
                .keyboardShortcut(.defaultAction)
                .disabled(newName.trimmingCharacters(in: .whitespaces).isEmpty)
            }
            .padding(16)
        }
        .frame(width: 520, height: 480)
        .task { await viewModel.loadTags() }
        .confirmationDialog(
            t("modal.tag.confirmDelete"),
            isPresented: Binding(
                get: { tagPendingDeletion != nil },
                set: { if !$0 { tagPendingDeletion = nil } }
            )
        ) {
            Button(t("common.delete"), role: .destructive) {
                guard let tag = tagPendingDeletion else { return }
                tagPendingDeletion = nil
                Task { await viewModel.deleteTag(tag) }
            }
            Button(t("common.cancel"), role: .cancel) { tagPendingDeletion = nil }
        }
    }

    private func row(for tag: Tag) -> some View {
        HStack(spacing: 10) {
            ColorPicker("", selection: colorBinding(for: tag), supportsOpacity: false)
                .labelsHidden()
                .frame(width: 44)

            TextField("", text: nameBinding(for: tag))
                .textFieldStyle(.plain)
                .onSubmit { commitName(for: tag) }
                .onDisappear { commitName(for: tag) }

            Text(t("modal.tag.assignedFeeds", ["count": viewModel.feeds(taggedWith: tag.id).count]))
                .font(.caption)
                .foregroundStyle(.secondary)

            Button(role: .destructive) {
                tagPendingDeletion = tag
            } label: {
                Image(systemName: "trash")
            }
            .buttonStyle(.borderless)
        }
    }

    private func nameBinding(for tag: Tag) -> Binding<String> {
        Binding(
            get: { draftNames[tag.id] ?? tag.name },
            set: { draftNames[tag.id] = $0 }
        )
    }

    /// Saves a renamed tag, if the name actually changed.
    private func commitName(for tag: Tag) {
        guard let name = draftNames[tag.id] else { return }
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        draftNames[tag.id] = nil
        guard !trimmed.isEmpty, trimmed != tag.name else { return }

        var updated = tag
        updated.name = trimmed
        Task { await viewModel.updateTag(updated) }
    }

    private func colorBinding(for tag: Tag) -> Binding<Color> {
        Binding(
            get: { draftColors[tag.id] ?? Color(hex: tag.color) ?? .accentColor },
            set: { newValue in
                draftColors[tag.id] = newValue
                scheduleColorSave(for: tag, color: newValue)
            }
        )
    }

    private func scheduleColorSave(for tag: Tag, color: Color) {
        colorSaveTasks[tag.id]?.cancel()
        colorSaveTasks[tag.id] = Task {
            try? await Task.sleep(for: .milliseconds(400))
            guard !Task.isCancelled else { return }

            let hex = color.hexString
            guard hex != tag.color else { return }
            var updated = tag
            updated.color = hex
            await viewModel.updateTag(updated)
            draftColors[tag.id] = nil
        }
    }

    private func create() async {
        let name = newName.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !name.isEmpty else { return }
        if await viewModel.createTag(name: name, color: newColor.hexString) {
            newName = ""
        }
    }
}

extension Color {
    /// The `#rrggbb` form the backend stores.
    var hexString: String {
        let converted = NSColor(self).usingColorSpace(.sRGB) ?? NSColor.controlAccentColor
        let red = Int((converted.redComponent * 255).rounded())
        let green = Int((converted.greenComponent * 255).rounded())
        let blue = Int((converted.blueComponent * 255).rounded())
        return String(format: "#%02x%02x%02x", red, green, blue)
    }
}
