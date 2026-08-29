import SwiftUI

/// A stable colour for each subscription, so the sidebar reads as a list of
/// distinguishable sources rather than a column of identical glyphs. The hues
/// are spread around the wheel instead of being limited to the primaries.
enum FeedAccentColor {
    static let palette: [Color] = [
        Color(red: 0.90, green: 0.26, blue: 0.28),  // red
        Color(red: 0.95, green: 0.48, blue: 0.16),  // orange
        Color(red: 0.93, green: 0.71, blue: 0.13),  // amber
        Color(red: 0.71, green: 0.75, blue: 0.16),  // olive
        Color(red: 0.30, green: 0.71, blue: 0.31),  // green
        Color(red: 0.16, green: 0.71, blue: 0.58),  // teal
        Color(red: 0.15, green: 0.68, blue: 0.80),  // cyan
        Color(red: 0.20, green: 0.51, blue: 0.89),  // blue
        Color(red: 0.33, green: 0.36, blue: 0.82),  // indigo
        Color(red: 0.55, green: 0.34, blue: 0.80),  // violet
        Color(red: 0.79, green: 0.31, blue: 0.68),  // magenta
        Color(red: 0.88, green: 0.36, blue: 0.48),  // rose
        Color(red: 0.58, green: 0.44, blue: 0.32),  // brown
        Color(red: 0.42, green: 0.53, blue: 0.60)   // slate
    ]

    static func color(for feed: Feed) -> Color {
        palette[index(for: identity(of: feed))]
    }

    static func index(for identity: String) -> Int {
        // FNV-1a keeps the colour stable across launches, which Swift's own
        // hashing does not guarantee because it is seeded per process.
        var hash: UInt64 = 0xcbf2_9ce4_8422_2325
        for byte in identity.utf8 {
            hash ^= UInt64(byte)
            hash = hash &* 0x1000_0000_01b3
        }
        return Int(hash % UInt64(palette.count))
    }

    private static func identity(of feed: Feed) -> String {
        feed.url.isEmpty ? feed.title : feed.url
    }
}
