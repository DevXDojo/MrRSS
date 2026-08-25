import AppKit
import XCTest
@testable import MrRSS

final class StubURLProtocol: URLProtocol {
    nonisolated(unsafe) static var payload = Data()
    nonisolated(unsafe) static var requestCount = 0

    override class func canInit(with request: URLRequest) -> Bool { true }
    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }

    override func startLoading() {
        StubURLProtocol.requestCount += 1
        let response = HTTPURLResponse(
            url: request.url!,
            statusCode: 200,
            httpVersion: nil,
            headerFields: ["Content-Type": "image/png"]
        )!
        client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
        client?.urlProtocol(self, didLoad: StubURLProtocol.payload)
        client?.urlProtocolDidFinishLoading(self)
    }

    override func stopLoading() {}
}

final class RemoteImageTests: XCTestCase {
    func testAnImageIsDecodedAtTheSizeItIsDrawn() throws {
        let data = try Self.png(width: 1_200, height: 900)

        let thumbnail = try XCTUnwrap(RemoteImageLoader.downsample(data, maxPixelSize: 152))

        // Without downsampling the row would hold a 1200 by 900 bitmap and pay
        // to decode it on the main thread every time the list is rebuilt.
        XCTAssertLessThanOrEqual(max(thumbnail.size.width, thumbnail.size.height), 152)
        XCTAssertGreaterThan(min(thumbnail.size.width, thumbnail.size.height), 0)
    }

    func testTheAspectRatioSurvivesDownsampling() throws {
        let data = try Self.png(width: 1_200, height: 600)

        let thumbnail = try XCTUnwrap(RemoteImageLoader.downsample(data, maxPixelSize: 100))

        XCTAssertEqual(thumbnail.size.width / thumbnail.size.height, 2, accuracy: 0.05)
    }

    func testSomethingThatIsNotAnImageIsRefused() {
        XCTAssertNil(RemoteImageLoader.downsample(Data("not an image".utf8), maxPixelSize: 100))
    }

    func testAnImageIsFetchedOnceAndThenServedFromMemory() async throws {
        StubURLProtocol.payload = try Self.png(width: 400, height: 400)
        StubURLProtocol.requestCount = 0

        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [StubURLProtocol.self]
        let loader = RemoteImageLoader(session: URLSession(configuration: configuration))
        let url = URL(string: "https://example.com/thumbnail.png")!

        let first = await loader.image(for: url, maxPixelSize: 120)
        let second = await loader.image(for: url, maxPixelSize: 120)

        XCTAssertNotNil(first)
        XCTAssertNotNil(second)
        XCTAssertEqual(StubURLProtocol.requestCount, 1, "The same thumbnail should not be downloaded again.")
        XCTAssertNotNil(loader.cachedImage(for: url, maxPixelSize: 120))
    }

    private static func png(width: Int, height: Int) throws -> Data {
        let representation = try XCTUnwrap(NSBitmapImageRep(
            bitmapDataPlanes: nil,
            pixelsWide: width,
            pixelsHigh: height,
            bitsPerSample: 8,
            samplesPerPixel: 4,
            hasAlpha: true,
            isPlanar: false,
            colorSpaceName: .deviceRGB,
            bytesPerRow: 0,
            bitsPerPixel: 0
        ))
        NSGraphicsContext.saveGraphicsState()
        NSGraphicsContext.current = NSGraphicsContext(bitmapImageRep: representation)
        NSColor.systemTeal.setFill()
        NSRect(x: 0, y: 0, width: width, height: height).fill()
        NSGraphicsContext.restoreGraphicsState()
        return try XCTUnwrap(representation.representation(using: .png, properties: [:]))
    }
}
