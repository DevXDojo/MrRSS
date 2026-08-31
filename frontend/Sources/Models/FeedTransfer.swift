import UniformTypeIdentifiers

extension UTType {
    /// Specific to MrRSS, so a drag coming from another application is never
    /// mistaken for one of our subscriptions.
    ///
    /// The type is derived from a tag rather than declared with
    /// `UTType(exportedAs:)`. An exported declaration only carries its
    /// conformances when the bundle's Info.plist declares it as well, and a
    /// type that conforms to nothing matches no drop target at all.
    static let mrrssFeed = UTType(
        tag: "mrrssfeed",
        tagClass: .filenameExtension,
        conformingTo: .data
    ) ?? .data
}

/// What a sidebar row carries while it is being dragged. The outline view
/// writes it straight onto the drag pasteboard under `UTType.mrrssFeed`.
struct FeedTransfer: Codable, Equatable {
    let feedID: Int
}
