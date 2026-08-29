import XCTest

/// Waits for something the view model does off the main task to land.
///
/// A fixed sleep has to guess how long the work takes, and a guess that holds
/// on a developer's machine fails on a loaded continuous integration runner.
/// Polling costs the same when things are quick and still passes when they are
/// not.
@MainActor
func waitUntil(
    _ description: String = "condition",
    timeout: Duration = .seconds(10),
    file: StaticString = #filePath,
    line: UInt = #line,
    _ condition: () -> Bool
) async throws {
    let deadline = ContinuousClock.now + timeout

    while ContinuousClock.now < deadline {
        if condition() { return }
        try await Task.sleep(for: .milliseconds(10))
    }

    XCTFail("Timed out waiting for \(description).", file: file, line: line)
}
