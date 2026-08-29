import AppKit
import XCTest
@testable import MrRSS

final class AppDelegateTests: XCTestCase {
    func testRepositoryIconIsFoundFromTheBuildDirectory() throws {
        let url = try XCTUnwrap(
            AppDelegate.repositoryIconURL(startingAt: Bundle(for: type(of: self)).bundleURL),
            "The application icon should be reachable from the built executable."
        )

        XCTAssertEqual(url.lastPathComponent, "icons.icns")
        XCTAssertNotNil(NSImage(contentsOf: url))
    }

    func testRepositoryIconLookupStopsAtTheFilesystemRoot() {
        XCTAssertNil(AppDelegate.repositoryIconURL(startingAt: URL(fileURLWithPath: "/")))
    }
}
