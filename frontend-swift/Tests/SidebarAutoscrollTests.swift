import AppKit
import XCTest
@testable import MrRSS

@MainActor
final class SidebarAutoscrollTests: XCTestCase {
    private let margin: CGFloat = 28

    func testAPointerInTheMiddleDoesNotScroll() {
        XCTAssertNil(offset(currentOffset: 100, pointerY: 200))
    }

    func testAPointerNearTheTopScrollsUp() throws {
        let result = try XCTUnwrap(offset(currentOffset: 100, pointerY: 110))
        XCTAssertLessThan(result, 100)
    }

    func testAPointerNearTheBottomScrollsDown() throws {
        // The visible band is 100...400, so 390 sits inside the lower margin.
        let result = try XCTUnwrap(offset(currentOffset: 100, pointerY: 390))
        XCTAssertGreaterThan(result, 100)
    }

    func testTheCloserToTheEdgeTheFasterItGoes() throws {
        let shallow = try XCTUnwrap(offset(currentOffset: 100, pointerY: 380))
        let deep = try XCTUnwrap(offset(currentOffset: 100, pointerY: 399))

        XCTAssertGreaterThan(deep - 100, shallow - 100)
    }

    func testItStopsAtTheTopOfTheList() {
        XCTAssertNil(offset(currentOffset: 0, pointerY: 5))
    }

    func testItStopsAtTheBottomOfTheList() {
        // 600 of content in a 300 tall list leaves 300 of travel.
        XCTAssertNil(offset(currentOffset: 300, pointerY: 595))
    }

    func testAListThatFitsNeverScrolls() {
        XCTAssertNil(
            SidebarDragCoordinator.autoscrollOffset(
                currentOffset: 0,
                visibleHeight: 300,
                contentHeight: 200,
                pointerY: 5,
                margin: margin
            )
        )
    }

    private func offset(currentOffset: CGFloat, pointerY: CGFloat) -> CGFloat? {
        SidebarDragCoordinator.autoscrollOffset(
            currentOffset: currentOffset,
            visibleHeight: 300,
            contentHeight: 600,
            pointerY: pointerY,
            margin: margin
        )
    }
}
