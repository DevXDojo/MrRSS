import XCTest
@testable import MrRSS

@MainActor
final class FeedDragPreviewTests: XCTestCase {
    private var defaults: UserDefaults!
    private var suiteName: String!

    override func setUp() {
        super.setUp()
        suiteName = "MrRSSDragPreviewTests-\(UUID().uuidString)"
        defaults = UserDefaults(suiteName: suiteName)
    }

    override func tearDown() {
        defaults.removePersistentDomain(forName: suiteName)
        defaults = nil
        suiteName = nil
        super.tearDown()
    }

    func testWithoutADragTheRowsReadInTheirStoredOrder() async throws {
        let (_, viewModel) = try await makeViewModel()

        XCTAssertEqual(viewModel.arrangedFeeds(inFolder: "", previewing: nil).map(\.id), [1, 2, 3])
    }

    func testTheRowsOpenUpWhereTheSubscriptionWouldLand() async throws {
        let (_, viewModel) = try await makeViewModel()

        // C is on its way to the top, so A and B move down to make room.
        let session = FeedDragSession(feedID: 3, folder: "", index: 0)

        XCTAssertEqual(viewModel.arrangedFeeds(inFolder: "", previewing: session).map(\.id), [3, 1, 2])
    }

    func testTheGapLeftBehindClosesWhileTheSubscriptionIsOverAnotherFolder() async throws {
        let (_, viewModel) = try await makeViewModel()

        let session = FeedDragSession(feedID: 1, folder: "Tech", index: 0)

        XCTAssertEqual(viewModel.arrangedFeeds(inFolder: "", previewing: session).map(\.id), [2, 3])
        XCTAssertEqual(viewModel.arrangedFeeds(inFolder: "Tech", previewing: session).map(\.id), [1, 4])
    }

    func testAnIndexBeyondTheListLandsAtTheEnd() async throws {
        let (_, viewModel) = try await makeViewModel()

        let session = FeedDragSession(feedID: 1, folder: "", index: 99)

        XCTAssertEqual(viewModel.arrangedFeeds(inFolder: "", previewing: session).map(\.id), [2, 3, 1])
    }

    func testCommittingThePreviewRanksTheSubscriptionThere() async throws {
        let (client, viewModel) = try await makeViewModel()

        await viewModel.placeFeed(id: 3, inFolder: "", at: 0)

        XCTAssertEqual(client.reorderMutations.map(\.id), [3])
        XCTAssertEqual(client.reorderMutations.map(\.position), [0])
        XCTAssertEqual(viewModel.unfiledFeeds.map(\.id), [3, 1, 2])
    }

    func testCommittingAPreviewThatChangesNothingSkipsTheServer() async throws {
        let (client, viewModel) = try await makeViewModel()

        await viewModel.placeFeed(id: 2, inFolder: "", at: 1)

        XCTAssertTrue(client.reorderMutations.isEmpty)
    }

    func testCommittingAPreviewIntoAnotherFolderFilesItThere() async throws {
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
        try await Task.sleep(for: .milliseconds(250))
        return (client, viewModel)
    }
}
