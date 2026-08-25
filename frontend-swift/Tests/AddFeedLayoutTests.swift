import AppKit
import SwiftUI
import XCTest
@testable import MrRSS

@MainActor
final class AddFeedLayoutTests: XCTestCase {
    func testSheetShowsEveryFieldAtFullWidthWithoutScrolling() async throws {
        let viewModel = AppViewModel(api: DelayedAPIClient(), autoLoad: false)
        let hostingView = NSHostingView(rootView: AddFeedView(viewModel: viewModel))
        let contentSize = hostingView.fittingSize
        hostingView.frame = NSRect(origin: .zero, size: contentSize)

        let window = NSWindow(
            contentRect: hostingView.frame,
            styleMask: [.titled, .closable],
            backing: .buffered,
            defer: false
        )
        window.contentView = hostingView
        window.makeKeyAndOrderFront(nil)
        hostingView.layoutSubtreeIfNeeded()
        try await waitUntil("the sheet to lay its fields out") {
            descendants(of: NSTextField.self, in: hostingView).filter(\.isEditable).count == 3
        }

        XCTAssertTrue(
            descendants(of: NSScrollView.self, in: hostingView).isEmpty,
            "A three field sheet should size itself to its content instead of scrolling."
        )

        let fields = descendants(of: NSTextField.self, in: hostingView).filter(\.isEditable)
        XCTAssertEqual(fields.count, 3)

        // Each field spans the sheet minus its padding, so a long address never
        // runs into a label placed beside it.
        for field in fields {
            XCTAssertEqual(field.frame.width, contentSize.width - 48, accuracy: 1)
        }
    }

    private func descendants<View: NSView>(of type: View.Type, in root: NSView) -> [View] {
        var found: [View] = []
        if let match = root as? View {
            found.append(match)
        }
        for subview in root.subviews {
            found.append(contentsOf: descendants(of: type, in: subview))
        }
        return found
    }
}
