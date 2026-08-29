import AppKit
import SwiftUI

/// Shows every image the backend extracted from an article, with a larger
/// preview for whichever one is chosen.
struct ImageGalleryView: View {
    let images: [String]
    let title: String

    @Environment(\.dismiss) private var dismiss
    @State private var selected: String?

    private let columns = [GridItem(.adaptive(minimum: 132), spacing: 12)]

    var body: some View {
        VStack(spacing: 0) {
            header
            Divider()

            if let selected, let url = URL(string: selected) {
                RemoteImage(url: url, displaySize: CGSize(width: 900, height: 560)) {
                    ProgressView()
                }
                .frame(maxWidth: .infinity, maxHeight: 420)
                .padding(16)
                Divider()
            }

            ScrollView {
                LazyVGrid(columns: columns, spacing: 12) {
                    ForEach(images, id: \.self) { image in
                        thumbnail(for: image)
                    }
                }
                .padding(16)
            }
        }
        .frame(minWidth: 720, minHeight: 560)
        .onAppear { selected = images.first }
    }

    private var header: some View {
        HStack {
            Label(title, systemImage: "photo.on.rectangle.angled")
                .font(.headline)
                .lineLimit(1)
            Spacer()
            if let selected, let url = URL(string: selected) {
                Button(t("common.action.openWebsite")) {
                    NSWorkspace.shared.open(url)
                }
                Button(t("common.contextMenu.copyLink")) {
                    NSPasteboard.general.clearContents()
                    NSPasteboard.general.setString(selected, forType: .string)
                }
            }
            Button(t("common.close")) { dismiss() }
                .keyboardShortcut(.cancelAction)
        }
        .padding(16)
    }

    private func thumbnail(for image: String) -> some View {
        Button {
            selected = image
        } label: {
            RemoteImage(url: URL(string: image), displaySize: CGSize(width: 132, height: 100)) {
                Rectangle().fill(.quaternary)
            }
            .frame(height: 100)
            .frame(maxWidth: .infinity)
            .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: 8, style: .continuous)
                    .strokeBorder(
                        selected == image ? Color.accentColor : Color.clear,
                        lineWidth: 2
                    )
            )
        }
        .buttonStyle(.plain)
    }
}
