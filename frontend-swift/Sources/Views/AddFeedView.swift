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

            // The label sits above the field rather than beside it, so a long
            // address keeps the full width of the sheet instead of scrolling
            // out of a narrow field next to its label.
            VStack(alignment: .leading, spacing: 14) {
                field("Feed URL", text: $url, prompt: "https://example.com/feed.xml")
                field("Title (optional)", text: $title, prompt: "Feed title")
                field("Category (optional)", text: $category, prompt: "Category name")
            }

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
        .frame(width: 520)
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

    private func field(_ label: String, text: Binding<String>, prompt: String) -> some View {
        VStack(alignment: .leading, spacing: 5) {
            Text(label)
                .font(.subheadline)
                .foregroundStyle(.secondary)
            TextField(label, text: text, prompt: Text(prompt))
                .textFieldStyle(.roundedBorder)
                .labelsHidden()
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
