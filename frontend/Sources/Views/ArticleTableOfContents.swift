import Foundation
import SwiftUI

/// One heading in the article, as the table of contents lists it.
struct TableOfContentsEntry: Identifiable, Hashable {
    let id: String
    let text: String
    /// 1, 2 or 3, so the list can indent nested headings.
    let level: Int
}

enum ArticleTableOfContents {
    /// Reads the headings out of the article markup.
    ///
    /// The web view renders the same markup, so anchoring by the heading's
    /// position in the document is enough to scroll to it.
    static func entries(in html: String) -> [TableOfContentsEntry] {
        let pattern = #"(?is)<h([1-3])\b[^>]*>(.*?)</h\1\s*>"#
        guard let regex = try? NSRegularExpression(pattern: pattern) else { return [] }

        let range = NSRange(html.startIndex..., in: html)
        var entries: [TableOfContentsEntry] = []

        for (index, match) in regex.matches(in: html, range: range).enumerated() {
            guard match.numberOfRanges == 3,
                  let levelRange = Range(match.range(at: 1), in: html),
                  let textRange = Range(match.range(at: 2), in: html),
                  let level = Int(html[levelRange]) else {
                continue
            }

            let text = sanitize(String(html[textRange]))
            guard !text.isEmpty else { continue }

            entries.append(
                TableOfContentsEntry(id: "mrrss-heading-\(index)", text: text, level: level)
            )
        }

        return normalise(entries)
    }

    /// Adds an anchor to every heading so the web view can be scrolled to one.
    static func anchored(_ html: String) -> String {
        let pattern = #"(?is)<h([1-3])\b"#
        guard let regex = try? NSRegularExpression(pattern: pattern) else { return html }

        var result = ""
        var lastIndex = html.startIndex
        let range = NSRange(html.startIndex..., in: html)

        for (index, match) in regex.matches(in: html, range: range).enumerated() {
            guard let matchRange = Range(match.range, in: html),
                  let levelRange = Range(match.range(at: 1), in: html) else {
                continue
            }
            result += html[lastIndex..<matchRange.lowerBound]
            result += "<h\(html[levelRange]) id=\"mrrss-heading-\(index)\""
            lastIndex = matchRange.upperBound
        }

        result += html[lastIndex...]
        return result
    }

    /// Strips the markup and the markdown hashes a heading may carry.
    static func sanitize(_ text: String) -> String {
        text.strippingHTML
            .replacingOccurrences(of: "^\\s*#+\\s*", with: "", options: .regularExpression)
            .trimmingCharacters(in: .whitespacesAndNewlines)
    }

    /// Shifts the levels so the shallowest heading present sits at the top,
    /// which is what the previous interface did when an article starts at `h2`.
    private static func normalise(_ entries: [TableOfContentsEntry]) -> [TableOfContentsEntry] {
        guard let shallowest = entries.map(\.level).min(), shallowest > 1 else { return entries }
        return entries.map {
            TableOfContentsEntry(id: $0.id, text: $0.text, level: $0.level - shallowest + 1)
        }
    }
}

/// The floating list of headings shown beside the article.
struct TableOfContentsPanel: View {
    let entries: [TableOfContentsEntry]
    let onSelect: (TableOfContentsEntry) -> Void

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 6) {
                ForEach(entries) { entry in
                    Button {
                        onSelect(entry)
                    } label: {
                        HStack(spacing: 6) {
                            Rectangle()
                                .fill(Color.accentColor.opacity(0.5))
                                .frame(width: markerWidth(for: entry.level), height: 2)
                            Text(entry.text)
                                .font(entry.level == 1 ? .callout : .caption)
                                .lineLimit(2)
                                .multilineTextAlignment(.leading)
                        }
                    }
                    .buttonStyle(.plain)
                    .padding(.leading, CGFloat(entry.level - 1) * 10)
                }
            }
            .padding(12)
        }
        .frame(width: 220)
        .background(.regularMaterial, in: RoundedRectangle(cornerRadius: 10, style: .continuous))
        .shadow(radius: 10, y: 2)
    }

    private func markerWidth(for level: Int) -> CGFloat {
        switch level {
        case 1: 20
        case 2: 14
        default: 8
        }
    }
}
