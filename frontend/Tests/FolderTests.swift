import XCTest
@testable import MrRSS

@MainActor
final class FolderTests: XCTestCase {
    private var defaults: UserDefaults!
    private var suiteName: String!

    override func setUp() {
        super.setUp()
        suiteName = "MrRSSFolderTests-\(UUID().uuidString)"
        defaults = UserDefaults(suiteName: suiteName)
    }

    override func tearDown() {
        defaults.removePersistentDomain(forName: suiteName)
        defaults = nil
        suiteName = nil
        super.tearDown()
    }

    private func makeViewModel(_ client: DelayedAPIClient) async throws -> AppViewModel {
        client.defaultFeeds = [
            Feed(id: 1, url: "https://a.example.com/feed", title: "Feed A", category: ""),
            Feed(id: 2, url: "https://b.example.com/feed", title: "Feed B", category: "Tech"),
            Feed(id: 3, url: "https://c.example.com/feed", title: "Feed C", category: "Tech")
        ]
        let viewModel = AppViewModel(api: client, autoLoad: false, defaults: defaults)
        viewModel.refreshFeeds()
        try await waitUntil("the feeds to load") { !viewModel.feeds.isEmpty }
        return viewModel
    }

    func testFoldersComeFromTheCategoryOnEachFeed() async throws {
        let viewModel = try await makeViewModel(DelayedAPIClient())

        XCTAssertEqual(viewModel.folders, ["Tech"])
        XCTAssertEqual(viewModel.feeds(inFolder: "Tech").map(\.id), [2, 3])
        XCTAssertEqual(viewModel.unfiledFeeds.map(\.id), [1])
    }

    func testAnEmptyFolderSurvivesUntilAFeedMovesIntoIt() async throws {
        let client = DelayedAPIClient()
        let viewModel = try await makeViewModel(client)

        XCTAssertTrue(viewModel.createFolder(named: "  Reading  "))
        XCTAssertEqual(viewModel.folders, ["Reading", "Tech"])

        // A folder with no feeds has nothing to store on the server, so it is
        // remembered locally and reappears on the next launch.
        let reopened = AppViewModel(api: client, autoLoad: false, defaults: defaults)
        XCTAssertEqual(reopened.folders, ["Reading"])
    }

    func testCreateFolderRejectsBlankAndDuplicateNames() async throws {
        let viewModel = try await makeViewModel(DelayedAPIClient())

        XCTAssertFalse(viewModel.createFolder(named: "   "))
        XCTAssertNotNil(viewModel.errorMessage)

        viewModel.clearError()
        XCTAssertFalse(viewModel.createFolder(named: "Tech"))
        XCTAssertNotNil(viewModel.errorMessage)
        XCTAssertEqual(viewModel.folders, ["Tech"])
    }

    func testMovingAFeedUpdatesTheServerAndTheSidebar() async throws {
        let client = DelayedAPIClient()
        let viewModel = try await makeViewModel(client)

        await viewModel.moveFeed(viewModel.feeds[0], toFolder: "Tech")

        XCTAssertEqual(client.categoryMutations.map(\.id), [1])
        XCTAssertEqual(client.categoryMutations.map(\.category), ["Tech"])
        XCTAssertEqual(viewModel.feeds(inFolder: "Tech").map(\.id), [1, 2, 3])
        XCTAssertTrue(viewModel.unfiledFeeds.isEmpty)
    }

    func testMovingAFeedOutOfAFolderClearsItsCategory() async throws {
        let client = DelayedAPIClient()
        let viewModel = try await makeViewModel(client)
        let feed = try XCTUnwrap(viewModel.feeds.first(where: { $0.id == 2 }))

        await viewModel.moveFeed(feed, toFolder: nil)

        XCTAssertEqual(client.categoryMutations.map(\.category), [""])
        XCTAssertEqual(viewModel.unfiledFeeds.map(\.id), [1, 2])
    }

    func testAFailedMoveLeavesTheSidebarUnchanged() async throws {
        let client = DelayedAPIClient()
        let viewModel = try await makeViewModel(client)
        client.categoryUpdateError = URLError(.notConnectedToInternet)

        await viewModel.moveFeed(viewModel.feeds[0], toFolder: "Tech")

        XCTAssertNotNil(viewModel.errorMessage)
        XCTAssertEqual(viewModel.unfiledFeeds.map(\.id), [1])
    }

    func testRenamingAFolderMovesEveryFeedAndFollowsTheSelection() async throws {
        let client = DelayedAPIClient()
        let viewModel = try await makeViewModel(client)
        viewModel.selection = .folder("Tech")

        await viewModel.renameFolder("Tech", to: "Technology")

        XCTAssertEqual(client.categoryMutations.map(\.id), [2, 3])
        XCTAssertEqual(Set(client.categoryMutations.map(\.category)), ["Technology"])
        XCTAssertEqual(viewModel.folders, ["Technology"])
        XCTAssertEqual(viewModel.selection, .folder("Technology"))
    }

    func testDeletingAFolderKeepsItsFeedsSubscribed() async throws {
        let client = DelayedAPIClient()
        let viewModel = try await makeViewModel(client)
        viewModel.selection = .folder("Tech")

        await viewModel.deleteFolder("Tech")

        XCTAssertEqual(Set(client.categoryMutations.map(\.category)), [""])
        XCTAssertTrue(viewModel.folders.isEmpty)
        XCTAssertEqual(viewModel.feeds.count, 3)
        XCTAssertEqual(viewModel.selection, .filter(.all))
    }

    func testFolderBadgeSumsTheUnreadCountsOfItsFeeds() async throws {
        let client = DelayedAPIClient()
        client.unreadCounts = UnreadCounts(total: 30, feedCounts: [1: 5, 2: 7, 3: 18])
        let viewModel = try await makeViewModel(client)

        XCTAssertEqual(viewModel.unreadCount(forFolder: "Tech"), 25)
    }

    func testSelectingAFolderAsksTheServerForThatCategory() async throws {
        let client = DelayedAPIClient()
        let viewModel = try await makeViewModel(client)

        viewModel.selection = .folder("Tech")
        try await waitUntil("the folder's articles to be requested") {
            client.lastArticleQuery?.category == "Tech"
        }

        XCTAssertEqual(client.lastArticleQuery?.category, "Tech")
        XCTAssertNil(client.lastArticleQuery?.feedID)
    }
}
