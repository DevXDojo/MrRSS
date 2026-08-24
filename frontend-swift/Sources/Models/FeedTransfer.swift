import SwiftUI
import UniformTypeIdentifiers

extension UTType {
    /// Private to MrRSS, so a drag coming from another application is never
    /// mistaken for one of our subscriptions.
    static let mrrssFeed = UTType(exportedAs: "com.devxdojo.mrrss.feed", conformingTo: .data)
}

/// What a sidebar row carries while it is being dragged onto a folder.
struct FeedTransfer: Codable, Transferable, Equatable {
    let feedID: Int

    static var transferRepresentation: some TransferRepresentation {
        CodableRepresentation(contentType: .mrrssFeed)
    }
}
