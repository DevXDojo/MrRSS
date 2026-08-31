import AppKit
import XCTest
@testable import MrRSS

final class KeyboardShortcutTests: XCTestCase {
    func testDefaultBindingsResolveToTheirActions() {
        let table = KeyboardShortcutTable()

        XCTAssertEqual(table.action(forBinding: "j"), .nextArticle)
        XCTAssertEqual(table.action(forBinding: "k"), .previousArticle)
        XCTAssertEqual(table.action(forBinding: "Shift+r"), .refreshFeeds)
        XCTAssertEqual(table.action(forBinding: "Alt+s"), .toggleFavoritesFilter)
        XCTAssertEqual(table.action(forBinding: "4"), .goToReadLater)
    }

    func testStoredBindingsOverrideTheDefaults() {
        let table = KeyboardShortcutTable(settings: [
            "shortcuts": #"{"nextArticle":"n","toggleReadStatus":"Shift+m"}"#
        ])

        XCTAssertEqual(table.action(forBinding: "n"), .nextArticle)
        XCTAssertEqual(table.action(forBinding: "Shift+m"), .toggleReadStatus)
        XCTAssertNil(table.action(forBinding: "j"), "the replaced binding should no longer match")
        XCTAssertEqual(table.action(forBinding: "k"), .previousArticle, "untouched bindings stay")
    }

    func testModifierOrderDoesNotMatter() {
        let table = KeyboardShortcutTable(settings: [
            "shortcuts": #"{"markAllRead":"Shift+Alt+a"}"#
        ])

        XCTAssertEqual(table.action(forBinding: "Alt+Shift+a"), .markAllRead)
    }

    func testShortcutsCanBeSwitchedOff() {
        XCTAssertFalse(KeyboardShortcutTable(settings: ["shortcuts_enabled": "false"]).isEnabled)
        XCTAssertTrue(KeyboardShortcutTable(settings: ["shortcuts_enabled": "true"]).isEnabled)
        XCTAssertTrue(KeyboardShortcutTable().isEnabled)
    }

    func testArrowKeysAreNamedTheSameWayTheSettingsStoreThem() throws {
        let event = try XCTUnwrap(
            NSEvent.keyEvent(
                with: .keyDown,
                location: .zero,
                modifierFlags: [],
                timestamp: 0,
                windowNumber: 0,
                context: nil,
                characters: "",
                charactersIgnoringModifiers: "",
                isARepeat: false,
                keyCode: 125
            )
        )

        XCTAssertEqual(KeyboardShortcutTable.combination(for: event), "ArrowDown")
    }
}

@MainActor
final class ShortcutActionTests: XCTestCase {
    func testNextAndPreviousMoveThroughTheListInOrder() {
        let defaults = UserDefaults(suiteName: "ShortcutActionTests")!
        defaults.removePersistentDomain(forName: "ShortcutActionTests")
        let viewModel = AppViewModel(api: StubAPIClient(), autoLoad: false, defaults: defaults)
        viewModel.articles = [
            Article(id: 1, feedID: 1, title: "A", url: "https://example.com/1", publishedAt: "2026-08-18T08:00:00Z"),
            Article(id: 2, feedID: 1, title: "B", url: "https://example.com/2", publishedAt: "2026-08-17T08:00:00Z")
        ]

        viewModel.perform(.nextArticle)
        XCTAssertEqual(viewModel.selectedArticleID, 1)

        viewModel.perform(.nextArticle)
        XCTAssertEqual(viewModel.selectedArticleID, 2)

        viewModel.perform(.previousArticle)
        XCTAssertEqual(viewModel.selectedArticleID, 1)
    }

    func testActivityShortcutsChangeTheSelection() {
        let defaults = UserDefaults(suiteName: "ShortcutActionTests")!
        defaults.removePersistentDomain(forName: "ShortcutActionTests")
        let viewModel = AppViewModel(api: StubAPIClient(), autoLoad: false, defaults: defaults)

        viewModel.perform(.goToFavorites)
        XCTAssertEqual(viewModel.selection, .filter(.favorites))

        viewModel.perform(.toggleReadLaterFilter)
        XCTAssertEqual(viewModel.selection, .filter(.readLater))

        viewModel.perform(.toggleReadLaterFilter)
        XCTAssertEqual(viewModel.selection, .filter(.all))
    }
}
