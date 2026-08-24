import SwiftUI

struct FolderNameView: View {
    let prompt: FolderPrompt
    @ObservedObject var viewModel: AppViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var name = ""
    @State private var isSubmitting = false

    var body: some View {
        VStack(alignment: .leading, spacing: 20) {
            VStack(alignment: .leading, spacing: 5) {
                Text(prompt.title)
                    .font(.title2.bold())
                Text(prompt.message)
                    .foregroundStyle(.secondary)
            }

            VStack(alignment: .leading, spacing: 5) {
                Text("Name")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                TextField("Name", text: $name, prompt: Text("Technology"))
                    .textFieldStyle(.roundedBorder)
                    .labelsHidden()
            }

            HStack {
                Spacer()
                Button("Cancel", role: .cancel) {
                    dismiss()
                }
                Button(prompt.confirmTitle) {
                    submit()
                }
                .keyboardShortcut(.defaultAction)
                .disabled(name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || isSubmitting)
            }
        }
        .padding(24)
        .frame(width: 420)
        .onAppear { name = prompt.initialName }
    }

    private func submit() {
        isSubmitting = true
        Task {
            switch prompt {
            case .create:
                if viewModel.createFolder(named: name) { dismiss() }
            case .createWithFeed(let feed):
                if viewModel.createFolder(named: name) {
                    await viewModel.moveFeed(feed, toFolder: name)
                    dismiss()
                }
            case .rename(let folder):
                await viewModel.renameFolder(folder, to: name)
                dismiss()
            }
            isSubmitting = false
        }
    }
}
