import SwiftUI
import WebKit
import XCTest
@testable import MrRSS

@MainActor
final class WebViewLayoutTests: XCTestCase {
    func testContainerKeepsWebViewInsideItsBounds() {
        let webView = WKWebView(frame: .zero)
        let container = ClippedWebViewContainer(webView: webView)
        container.frame = NSRect(x: 0, y: 0, width: 640, height: 480)

        container.layoutSubtreeIfNeeded()

        XCTAssertEqual(webView.frame, container.bounds)
        XCTAssertEqual(container.intrinsicContentSize.width, NSView.noIntrinsicMetric)
        XCTAssertEqual(container.intrinsicContentSize.height, NSView.noIntrinsicMetric)
        XCTAssertTrue(container.layer?.masksToBounds == true)
        XCTAssertTrue(webView.layer?.masksToBounds == true)
    }

    func testArticleContentDoesNotExpandSplitViewBeyondWindow() async throws {
        let client = DelayedAPIClient()
        client.articleContent = ArticleContent(
            content: String(repeating: "<p>Long article content.</p>", count: 2_000),
            feedURL: "https://example.com"
        )
        client.defaultArticles = [
            Article(
                id: 1,
                feedID: 1,
                feedTitle: "Feed 1",
                title: String(repeating: "Long article title ", count: 20),
                url: "https://example.com/1",
                publishedAt: "2026-08-16T08:00:00Z",
                summary: String(repeating: "Long article summary. ", count: 400)
            )
        ]
        let viewModel = AppViewModel(api: client, autoLoad: false)

        viewModel.reloadArticles()
        try await waitUntil("the articles to load") { !viewModel.articles.isEmpty }
        let article = try XCTUnwrap(viewModel.articles.first)

        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 1_280, height: 780),
            styleMask: [.titled, .closable, .resizable, .fullSizeContentView],
            backing: .buffered,
            defer: false
        )
        let hostingView = NSHostingView(rootView: ContentView(viewModel: viewModel))
        window.contentView = hostingView
        window.makeKeyAndOrderFront(nil)
        hostingView.layoutSubtreeIfNeeded()

        viewModel.selectArticle(article)
        // The detail pane builds its web view once the article content has
        // loaded, which is the point the layout is worth measuring.
        try await waitUntil("the detail pane to be built") {
            firstDescendant(of: WKWebView.self, in: hostingView) != nil
        }
        window.layoutIfNeeded()
        hostingView.layoutSubtreeIfNeeded()

        let splitView = try XCTUnwrap(firstDescendant(of: NSSplitView.self, in: hostingView))
        XCTAssertEqual(splitView.frame.height, hostingView.bounds.height, accuracy: 1)
        XCTAssertEqual(splitView.frame.width, hostingView.bounds.width, accuracy: 1)

        // The split view is hosted by an intermediate AppKit view. When a column
        // reports an ideal height taller than the window, that view keeps its
        // oversized height and is centred, which scrolls the sidebar and the
        // article list out of the window.
        let splitHost = try XCTUnwrap(splitView.superview?.superview)
        XCTAssertEqual(splitHost.frame.origin.y, 0, accuracy: 1)
        XCTAssertEqual(splitHost.frame.height, hostingView.bounds.height, accuracy: 1)

        let visibleBounds = hostingView.bounds
        for column in splitView.subviews where column.frame.height > 100 {
            let columnFrame = column.convert(column.bounds, to: hostingView)
            XCTAssertEqual(
                columnFrame.origin.y, 0, accuracy: 1,
                "A split view column starts outside the window: \(columnFrame)"
            )
            XCTAssertEqual(
                columnFrame.height, visibleBounds.height, accuracy: 1,
                "A split view column is taller than the window: \(columnFrame)"
            )
        }
    }

    func testDetailPaneReportsASizeIndependentOfItsContent() async throws {
        let client = DelayedAPIClient()
        client.articleContent = ArticleContent(
            content: String(repeating: "<p>Long article content.</p>", count: 2_000),
            feedURL: "https://example.com"
        )
        let article = Article(
            id: 1,
            feedID: 1,
            feedTitle: "Feed 1",
            title: String(repeating: "Long article title ", count: 20),
            url: "https://example.com/1",
            publishedAt: "2026-08-16T08:00:00Z",
            summary: String(repeating: "Long article summary. ", count: 400)
        )
        let viewModel = AppViewModel(api: client, autoLoad: false)

        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 300, height: 200),
            styleMask: [.titled, .closable, .resizable],
            backing: .buffered,
            defer: false
        )
        let hostingView = NSHostingView(
            rootView: ArticleDetailView(article: article, viewModel: viewModel)
        )
        window.contentView = hostingView
        window.makeKeyAndOrderFront(nil)
        hostingView.layoutSubtreeIfNeeded()
        try await waitUntil("the detail pane to be built") {
            firstDescendant(of: WKWebView.self, in: hostingView) != nil
        }
        hostingView.layoutSubtreeIfNeeded()

        // A pane whose reported size follows the article text makes
        // NavigationSplitView lay the window out around that size.
        XCTAssertLessThanOrEqual(
            hostingView.fittingSize.width,
            hostingView.bounds.width,
            "The detail pane asks for more width than it is given: \(hostingView.fittingSize)"
        )
        XCTAssertLessThanOrEqual(
            hostingView.fittingSize.height,
            hostingView.bounds.height,
            "The detail pane asks for more height than it is given: \(hostingView.fittingSize)"
        )
    }

    private func firstDescendant<View: NSView>(of type: View.Type, in root: NSView) -> View? {
        if let match = root as? View {
            return match
        }
        for subview in root.subviews {
            if let match = firstDescendant(of: type, in: subview) {
                return match
            }
        }
        return nil
    }
}

final class WebViewSourceTests: XCTestCase {
    func testTypographyFollowsTheReadingSettings() {
        let typography = WebViewTypography(
            fontFamily: "serif",
            fontSize: 19,
            lineHeight: "1.8"
        )

        let document = HTMLDocument.build(from: "<p>Body</p>", typography: typography)

        XCTAssertTrue(document.contains("19px/1.8"), "the size and leading should reach the document")
        XCTAssertTrue(document.contains("New York"), "the serif stack should be used")
    }

    func testTheDefaultTypographyKeepsTheSystemStack() {
        let document = HTMLDocument.build(from: "<p>Body</p>")

        XCTAssertTrue(document.contains("16px/1.6"))
        XCTAssertTrue(document.contains("-apple-system"))
    }

    func testALiveSourceIsDistinguishedFromRenderedText() {
        let live = WebViewSource.url(URL(string: "https://example.com")!)
        let rendered = WebViewSource.html("<p>Body</p>", baseURL: nil)

        XCTAssertTrue(live.isLiveURL)
        XCTAssertFalse(rendered.isLiveURL)
    }

    func testScriptsAndFramesAreStrippedFromRenderedText() {
        let document = HTMLDocument.build(
            from: "<p>Safe</p><script>alert(1)</script><iframe src=\"x\"></iframe><a href=\"javascript:evil()\">x</a>"
        )

        XCTAssertTrue(document.contains("Safe"))
        XCTAssertFalse(document.contains("alert(1)"))
        XCTAssertFalse(document.contains("<iframe"))
        XCTAssertFalse(document.contains("javascript:evil"))
    }
}
