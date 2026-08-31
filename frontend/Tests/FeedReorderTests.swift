import XCTest
@testable import MrRSS

@MainActor
final class FeedReorderTests: XCTestCase {
    private var defaults: UserDefaults!
    private var suiteName: String!

    override func setUp() {
        super.setUp()
        suiteName = "MrRSSReorderTests-\(UUID().uuidString)"
        defaults = UserDefaults(suiteName: suiteName)
    }

    override func tearDown() {
        defaults.removePersistentDomain(forName: suiteName)
        defaults = nil
        suiteName = nil
        super.tearDown()
    }

    func testFeedsAreListedInTheOrderTheServerKeeps() async throws {
        let client = DelayedAPIClient()
        client.defaultFeeds = [
            Feed(id: 1, url: "https://a.example.com/feed", title: "A", category: "", position: 2),
            Feed(id: 2, url: "https://b.example.com/feed", title: "B", category: "", position: 0),
            Feed(id: 3, url: "https://c.example.com/feed", title: "C", category: "", position: 1)
        ]
        let viewModel = AppViewModel(api: client, autoLoad: false, defaults: defaults)
        viewModel.refreshFeeds()
        try await waitUntil("the feeds to load") { !viewModel.feeds.isEmpty }

        XCTAssertEqual(viewModel.feeds.map(\.id), [2, 3, 1])
    }

    func testDroppingAboveARowRanksTheFeedBeforeIt() async throws {
        let (client, viewModel) = try await makeViewModel()

        await viewModel.moveFeeds(ids: [3], relativeTo: 1, placeAbove: true)

        XCTAssertEqual(client.reorderMutations.map(\.id), [3])
        XCTAssertEqual(client.reorderMutations.map(\.position), [0])
        XCTAssertEqual(viewModel.unfiledFeeds.map(\.id), [3, 1, 2])
    }

    func testDroppingBelowARowRanksTheFeedAfterIt() async throws {
        let (client, viewModel) = try await makeViewModel()

        await viewModel.moveFeeds(ids: [3], relativeTo: 1, placeAbove: false)

        XCTAssertEqual(client.reorderMutations.map(\.position), [1])
        XCTAssertEqual(viewModel.unfiledFeeds.map(\.id), [1, 3, 2])
    }

    func testDroppingOnARowInsideAFolderFilesTheFeedThere() async throws {
        let (client, viewModel) = try await makeViewModel()

        await viewModel.moveFeeds(ids: [1], relativeTo: 4, placeAbove: true)

        XCTAssertEqual(client.reorderMutations.map(\.category), ["Tech"])
        XCTAssertEqual(viewModel.feeds(inFolder: "Tech").map(\.id), [1, 4])
        XCTAssertEqual(viewModel.unfiledFeeds.map(\.id), [2, 3])
    }

    func testDroppingARowOnItselfChangesNothing() async throws {
        let (client, viewModel) = try await makeViewModel()

        await viewModel.moveFeeds(ids: [1], relativeTo: 1, placeAbove: true)

        XCTAssertTrue(client.reorderMutations.isEmpty)
        XCTAssertEqual(viewModel.unfiledFeeds.map(\.id), [1, 2, 3])
    }

    func testSeveralFeedsKeepTheOrderTheyWereDraggedIn() async throws {
        let (client, viewModel) = try await makeViewModel()

        await viewModel.moveFeeds(ids: [2, 3], relativeTo: 1, placeAbove: true)

        XCTAssertEqual(client.reorderMutations.map(\.id), [2, 3])
        XCTAssertEqual(viewModel.unfiledFeeds.map(\.id), [2, 3, 1])
    }

    func testAFailedReorderLeavesTheListAlone() async throws {
        let (client, viewModel) = try await makeViewModel()
        client.categoryUpdateError = URLError(.notConnectedToInternet)

        await viewModel.moveFeeds(ids: [3], relativeTo: 1, placeAbove: true)

        XCTAssertNotNil(viewModel.errorMessage)
        XCTAssertEqual(viewModel.unfiledFeeds.map(\.id), [1, 2, 3])
    }

    func testCommittingWhereTheRowsOpenedUpRanksTheSubscriptionThere() async throws {
        let (client, viewModel) = try await makeViewModel()

        await viewModel.placeFeed(id: 3, inFolder: "", at: 0)

        XCTAssertEqual(client.reorderMutations.map(\.id), [3])
        XCTAssertEqual(client.reorderMutations.map(\.position), [0])
        XCTAssertEqual(viewModel.unfiledFeeds.map(\.id), [3, 1, 2])
    }

    func testCommittingAPositionThatChangesNothingSkipsTheServer() async throws {
        let (client, viewModel) = try await makeViewModel()

        await viewModel.placeFeed(id: 2, inFolder: "", at: 1)

        XCTAssertTrue(client.reorderMutations.isEmpty)
    }

    func testCommittingIntoAnotherFolderFilesItThere() async throws {
        let (client, viewModel) = try await makeViewModel()

        await viewModel.placeFeed(id: 1, inFolder: "Tech", at: 0)

        XCTAssertEqual(client.reorderMutations.map(\.category), ["Tech"])
        XCTAssertEqual(viewModel.feeds(inFolder: "Tech").map(\.id), [1, 4])
    }

    private func makeViewModel() async throws -> (DelayedAPIClient, AppViewModel) {
        let client = DelayedAPIClient()
        client.defaultFeeds = [
            Feed(id: 1, url: "https://a.example.com/feed", title: "A", category: "", position: 0),
            Feed(id: 2, url: "https://b.example.com/feed", title: "B", category: "", position: 1),
            Feed(id: 3, url: "https://c.example.com/feed", title: "C", category: "", position: 2),
            Feed(id: 4, url: "https://d.example.com/feed", title: "D", category: "Tech", position: 0)
        ]
        let viewModel = AppViewModel(api: client, autoLoad: false, defaults: defaults)
        viewModel.refreshFeeds()
        try await waitUntil("the feeds to load") { !viewModel.feeds.isEmpty }
        return (client, viewModel)
    }
}
