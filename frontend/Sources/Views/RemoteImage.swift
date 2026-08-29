import AppKit
import ImageIO
import SwiftUI

/// Loads a remote image at the size it will actually be drawn at.
///
/// `AsyncImage` decodes whatever the server sends at full resolution on the
/// main thread and keeps nothing between appearances, so a list of thumbnails
/// stalls the interface every time it is rebuilt.
final class RemoteImageLoader: @unchecked Sendable {
    static let shared = RemoteImageLoader()

    private let cache = NSCache<NSString, NSImage>()
    private let session: URLSession

    init(session: URLSession = .shared) {
        self.session = session
        cache.countLimit = 400
    }

    func image(for url: URL, maxPixelSize: CGFloat) async -> NSImage? {
        let key = RemoteImageLoader.key(for: url, maxPixelSize: maxPixelSize)
        if let cached = cache.object(forKey: key) {
            return cached
        }

        guard let data = try? await session.data(from: url).0,
              let image = RemoteImageLoader.downsample(data, maxPixelSize: maxPixelSize) else {
            return nil
        }

        cache.setObject(image, forKey: key)
        return image
    }

    func cachedImage(for url: URL, maxPixelSize: CGFloat) -> NSImage? {
        cache.object(forKey: RemoteImageLoader.key(for: url, maxPixelSize: maxPixelSize))
    }

    private static func key(for url: URL, maxPixelSize: CGFloat) -> NSString {
        "\(url.absoluteString)|\(Int(maxPixelSize))" as NSString
    }

    /// Decodes straight to the drawn size. ImageIO does the work off the main
    /// thread and never holds the full sized bitmap.
    static func downsample(_ data: Data, maxPixelSize: CGFloat) -> NSImage? {
        guard maxPixelSize > 0,
              let source = CGImageSourceCreateWithData(
                data as CFData,
                [kCGImageSourceShouldCache: false] as CFDictionary
              ) else {
            return nil
        }

        let options: [CFString: Any] = [
            kCGImageSourceCreateThumbnailFromImageAlways: true,
            kCGImageSourceCreateThumbnailWithTransform: true,
            kCGImageSourceShouldCacheImmediately: true,
            kCGImageSourceThumbnailMaxPixelSize: maxPixelSize
        ]

        guard let thumbnail = CGImageSourceCreateThumbnailAtIndex(source, 0, options as CFDictionary) else {
            return nil
        }
        return NSImage(
            cgImage: thumbnail,
            size: NSSize(width: thumbnail.width, height: thumbnail.height)
        )
    }
}

struct RemoteImage<Placeholder: View>: View {
    let url: URL?
    /// The size the image is drawn at, in points.
    let displaySize: CGSize
    @ViewBuilder let placeholder: () -> Placeholder

    @State private var image: NSImage?

    var body: some View {
        content
            .task(id: url) {
                guard let url else {
                    image = nil
                    return
                }
                image = await RemoteImageLoader.shared.image(for: url, maxPixelSize: maxPixelSize)
            }
    }

    @ViewBuilder
    private var content: some View {
        if let image {
            Image(nsImage: image)
                .resizable()
                .scaledToFill()
        } else {
            placeholder()
        }
    }

    private var maxPixelSize: CGFloat {
        max(displaySize.width, displaySize.height) * 2
    }
}
