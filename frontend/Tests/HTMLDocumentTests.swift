import XCTest
@testable import MrRSS

final class HTMLDocumentTests: XCTestCase {
    func testDocumentRemovesExecutableMarkup() {
        let input = #"<p onclick="steal()">Safe</p><script>steal()</script><a href="javascript:steal()">Bad link</a>"#

        let document = HTMLDocument.build(from: input)

        XCTAssertTrue(document.contains("<p>Safe</p>"))
        XCTAssertFalse(document.contains("<script>steal()</script>"))
        XCTAssertFalse(document.contains("onclick="))
        XCTAssertFalse(document.contains("javascript:"))
        XCTAssertTrue(document.contains("script-src 'none'"))
        XCTAssertTrue(document.contains("form-action 'none'"))
    }

    func testEscapeProtectsPlainTextSummary() {
        XCTAssertEqual(
            HTMLDocument.escape(#"<hello> & "world""#),
            "&lt;hello&gt; &amp; &quot;world&quot;"
        )
    }
}
