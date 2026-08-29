import WebKit
import XCTest
@testable import MrRSS

final class TableOfContentsTests: XCTestCase {
    func testHeadingsAreListedWithTheirLevels() {
        let html = """
        <h1>Introduction</h1><p>Text</p>
        <h2>First part</h2><p>Text</p>
        <h3>Detail</h3>
        """

        let entries = ArticleTableOfContents.entries(in: html)

        XCTAssertEqual(entries.map(\.text), ["Introduction", "First part", "Detail"])
        XCTAssertEqual(entries.map(\.level), [1, 2, 3])
    }

    func testLevelsShiftUpWhenAnArticleStartsAtTheSecondLevel() {
        let html = "<h2>One</h2><h3>Two</h3>"

        let entries = ArticleTableOfContents.entries(in: html)

        XCTAssertEqual(entries.map(\.level), [1, 2])
    }

    func testMarkupAndMarkdownHashesAreStrippedFromHeadings() {
        let html = "<h2><a href=\"#x\">## Section&nbsp;title</a></h2>"

        let entries = ArticleTableOfContents.entries(in: html)

        XCTAssertEqual(entries.first?.text, "Section title")
    }

    func testHeadingsWithoutTextAreSkipped() {
        let html = "<h1></h1><h2>Real</h2>"

        let entries = ArticleTableOfContents.entries(in: html)

        XCTAssertEqual(entries.map(\.text), ["Real"])
    }

    func testAnchorsMatchTheListedIdentifiers() {
        let html = "<h1>One</h1><h2 class=\"x\">Two</h2>"

        let entries = ArticleTableOfContents.entries(in: html)
        let anchored = ArticleTableOfContents.anchored(html)

        for entry in entries {
            XCTAssertTrue(
                anchored.contains("id=\"\(entry.id)\""),
                "\(entry.id) should be anchored in the rendered markup"
            )
        }
        XCTAssertTrue(anchored.contains("class=\"x\""), "existing attributes should survive")
    }
}

@MainActor
final class WebViewScriptingTests: XCTestCase {
    /// The rendered document has page scripting switched off. This checks that
    /// the application can still run a script itself, which is how the table of
    /// contents scrolls the article.
    func testTheHostCanRunScriptsWhilePageScriptingIsOff() async throws {
        let configuration = WKWebViewConfiguration()
        let preferences = WKWebpagePreferences()
        preferences.allowsContentJavaScript = false
        configuration.defaultWebpagePreferences = preferences

        let webView = WKWebView(frame: .init(x: 0, y: 0, width: 400, height: 400), configuration: configuration)
        webView.loadHTMLString("<html><body><h1 id=\"target\">Heading</h1></body></html>", baseURL: nil)

        try await waitUntil("the document to load") { !webView.isLoading }

        let result = try await webView.evaluateJavaScript("document.getElementById('target') !== null")

        XCTAssertEqual(result as? Bool, true)
    }
}

final class MediaPlayerTests: XCTestCase {
    func testDurationsAreWrittenAsMinutesAndSeconds() {
        XCTAssertEqual(AudioPlayerBar.format(0), "0:00")
        XCTAssertEqual(AudioPlayerBar.format(9), "0:09")
        XCTAssertEqual(AudioPlayerBar.format(75), "1:15")
        XCTAssertEqual(AudioPlayerBar.format(3_725), "1:02:05")
    }

    func testAnUnknownDurationDoesNotProduceNonsense() {
        XCTAssertEqual(AudioPlayerBar.format(.nan), "0:00")
        XCTAssertEqual(AudioPlayerBar.format(.infinity), "0:00")
        XCTAssertEqual(AudioPlayerBar.format(-5), "0:00")
    }

    func testYouTubeAddressesBecomeEmbedAddresses() throws {
        let watch = try XCTUnwrap(URL(string: "https://www.youtube.com/watch?v=abc123&t=30"))
        let short = try XCTUnwrap(URL(string: "https://youtu.be/abc123"))
        let embed = try XCTUnwrap(URL(string: "https://www.youtube.com/embed/abc123"))

        XCTAssertEqual(VideoPlayerBar.embedURL(for: watch)?.absoluteString, "https://www.youtube.com/embed/abc123")
        XCTAssertEqual(VideoPlayerBar.embedURL(for: short)?.absoluteString, "https://www.youtube.com/embed/abc123")
        XCTAssertEqual(VideoPlayerBar.embedURL(for: embed), embed)
    }

    func testVimeoAddressesBecomeEmbedAddresses() throws {
        let url = try XCTUnwrap(URL(string: "https://vimeo.com/123456789"))

        XCTAssertEqual(
            VideoPlayerBar.embedURL(for: url)?.absoluteString,
            "https://player.vimeo.com/video/123456789"
        )
    }

    func testADirectFileHasNoEmbedAddress() throws {
        let url = try XCTUnwrap(URL(string: "https://example.com/clip.mp4"))

        XCTAssertNil(VideoPlayerBar.embedURL(for: url))
    }
}
