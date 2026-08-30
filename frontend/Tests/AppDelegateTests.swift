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

final class ToolTipDelayTests: XCTestCase {
    func testTheDelayIsRegisteredSoADeliberateHoverIsAnsweredQuickly() {
        let suite = "ToolTipDelayTests"
        let defaults = UserDefaults(suiteName: suite)!
        defaults.removePersistentDomain(forName: suite)

        AppDelegate.shortenToolTipDelay(in: defaults)

        XCTAssertEqual(
            defaults.integer(forKey: AppDelegate.toolTipDelayKey),
            AppDelegate.toolTipDelayMilliseconds
        )
    }

    func testAReadersOwnDelayIsLeftAlone() {
        let suite = "ToolTipDelayTests"
        let defaults = UserDefaults(suiteName: suite)!
        defaults.removePersistentDomain(forName: suite)
        defaults.set(1_500, forKey: AppDelegate.toolTipDelayKey)

        AppDelegate.shortenToolTipDelay(in: defaults)

        XCTAssertEqual(defaults.integer(forKey: AppDelegate.toolTipDelayKey), 1_500)
        defaults.removePersistentDomain(forName: suite)
    }
}
