import SwiftUI

struct AddFeedView: View {
    @ObservedObject var viewModel: AppViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var url = ""
    @State private var title = ""
    @State private var category = ""
    @State private var isSubmitting = false

    var body: some View {
        VStack(alignment: .leading, spacing: 20) {
            VStack(alignment: .leading, spacing: 5) {
                Text("Add Feed")
                    .font(.title2.bold())
                Text("Subscribe using an RSS, Atom, or compatible feed URL.")
                    .foregroundStyle(.secondary)
            }

            Form {
                TextField("Feed URL", text: $url, prompt: Text("https://example.com/feed.xml"))
                TextField("Title (optional)", text: $title)
                TextField("Category (optional)", text: $category)
            }
            .formStyle(.grouped)

            HStack {
                Spacer()
                Button("Cancel", role: .cancel) {
                    dismiss()
                }
                Button("Add Feed") {
                    submit()
                }
                .keyboardShortcut(.defaultAction)
                .disabled(url.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || isSubmitting)
            }
        }
        .padding(24)
        .frame(width: 500, height: 300)
        .overlay {
            if isSubmitting {
                ZStack {
                    Color.black.opacity(0.08)
                    ProgressView("Adding feed")
                        .padding(20)
                        .background(.regularMaterial, in: RoundedRectangle(cornerRadius: 12))
                }
            }
        }
    }

    private func submit() {
        isSubmitting = true
        Task {
            if await viewModel.addFeed(url: url, title: title, category: category) {
                dismiss()
            }
            isSubmitting = false
        }
    }
}
