import SwiftUI
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

/// What a sidebar row carries while it is being dragged onto a folder.
///
/// The payload travels under `UTType.mrrssFeed` alone. A drag registered under
/// a broad type such as `public.data` is claimed by the enclosing list, which
/// registers itself for row reordering, and never reaches the folder underneath
/// the pointer.
struct FeedTransfer: Codable, Equatable {
    let feedID: Int

    var itemProvider: NSItemProvider {
        let provider = NSItemProvider()
        provider.registerDataRepresentation(
            forTypeIdentifier: UTType.mrrssFeed.identifier,
            visibility: .ownProcess
        ) { completion in
            completion(try? JSONEncoder().encode(self), nil)
            return nil
        }
        return provider
    }

    static func feedIDs(from providers: [NSItemProvider]) async -> [Int] {
        var ids: [Int] = []
        for provider in providers {
            guard let transfer = await load(from: provider) else { continue }
            ids.append(transfer.feedID)
        }
        return ids
    }

    private static func load(from provider: NSItemProvider) async -> FeedTransfer? {
        await withCheckedContinuation { continuation in
            _ = provider.loadDataRepresentation(
                forTypeIdentifier: UTType.mrrssFeed.identifier
            ) { data, _ in
                guard let data, let transfer = try? JSONDecoder().decode(FeedTransfer.self, from: data) else {
                    continuation.resume(returning: nil)
                    return
                }
                continuation.resume(returning: transfer)
            }
        }
    }
}
