import AppKit
import SwiftUI
import WebKit

struct WebView: NSViewRepresentable {
    let html: String
    let baseURL: URL?

    func makeCoordinator() -> Coordinator {
        Coordinator()
    }

    func makeNSView(context: Context) -> ClippedWebViewContainer {
        let configuration = WKWebViewConfiguration()
        let webpagePreferences = WKWebpagePreferences()
        webpagePreferences.allowsContentJavaScript = false
        configuration.defaultWebpagePreferences = webpagePreferences
        configuration.preferences.javaScriptCanOpenWindowsAutomatically = false
        configuration.websiteDataStore = .nonPersistent()
        configuration.mediaTypesRequiringUserActionForPlayback = .all

        let webView = WKWebView(frame: .zero, configuration: configuration)
        webView.navigationDelegate = context.coordinator
        webView.underPageBackgroundColor = .textBackgroundColor
        return ClippedWebViewContainer(webView: webView)
    }

    func updateNSView(_ container: ClippedWebViewContainer, context: Context) {
        let webView = container.webView
        let document = HTMLDocument.build(from: html)
        guard document != context.coordinator.loadedDocument || baseURL != context.coordinator.loadedBaseURL else {
            return
        }

        context.coordinator.loadedDocument = document
        context.coordinator.loadedBaseURL = baseURL
        webView.loadHTMLString(document, baseURL: baseURL)
    }

    /// The container never reports a size of its own. Web content can be far
    /// taller than the window, and any size derived from the loaded document
    /// would be propagated up to `NavigationSplitView`, which then lays the
    /// whole window out around that height.
    func sizeThatFits(
        _ proposal: ProposedViewSize,
        nsView: ClippedWebViewContainer,
        context: Context
    ) -> CGSize? {
        CGSize(
            width: WebView.resolvedLength(proposal.width),
            height: WebView.resolvedLength(proposal.height)
        )
    }

    /// Maps an unspecified or infinite proposal to zero so the surrounding
    /// `frame(maxWidth:maxHeight:)` decides how much space the view receives.
    private static func resolvedLength(_ value: CGFloat?) -> CGFloat {
        guard let value, value.isFinite, value > 0 else { return 0 }
        return value
    }

    final class Coordinator: NSObject, WKNavigationDelegate {
        var loadedDocument = ""
        var loadedBaseURL: URL?

        func webView(
            _ webView: WKWebView,
            decidePolicyFor navigationAction: WKNavigationAction,
            decisionHandler: @escaping (WKNavigationActionPolicy) -> Void
        ) {
            guard let url = navigationAction.request.url else {
                decisionHandler(.allow)
                return
            }

            if navigationAction.navigationType == .linkActivated {
                if ["http", "https"].contains(url.scheme?.lowercased() ?? "") {
                    NSWorkspace.shared.open(url)
                }
                decisionHandler(.cancel)
                return
            }

            if url.scheme == "about" || navigationAction.targetFrame == nil {
                decisionHandler(.allow)
            } else if navigationAction.targetFrame?.isMainFrame == true,
                      url == loadedBaseURL {
                decisionHandler(.allow)
            } else {
                decisionHandler(.cancel)
            }
        }
    }
}

final class ClippedWebViewContainer: NSView {
    let webView: WKWebView

    init(webView: WKWebView) {
        self.webView = webView
        super.init(frame: .zero)

        wantsLayer = true
        layer?.masksToBounds = true
        layer?.backgroundColor = NSColor.textBackgroundColor.cgColor

        setContentHuggingPriority(.defaultLow, for: .horizontal)
        setContentHuggingPriority(.defaultLow, for: .vertical)
        setContentCompressionResistancePriority(.defaultLow, for: .horizontal)
        setContentCompressionResistancePriority(.defaultLow, for: .vertical)

        webView.frame = bounds
        webView.autoresizingMask = [.width, .height]
        webView.wantsLayer = true
        webView.layer?.masksToBounds = true
        addSubview(webView)
    }

    override var intrinsicContentSize: NSSize {
        NSSize(width: NSView.noIntrinsicMetric, height: NSView.noIntrinsicMetric)
    }

    override func layout() {
        super.layout()
        webView.frame = bounds
    }

    @available(*, unavailable)
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }
}

enum HTMLDocument {
    static func build(from input: String) -> String {
        let sanitized = sanitize(input)
        return """
        <!doctype html>
        <html>
        <head>
          <meta charset="utf-8">
          <meta name="viewport" content="width=device-width, initial-scale=1">
          <meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src http: https: data: blob:; media-src http: https: data:; style-src 'unsafe-inline'; font-src data:; script-src 'none'; connect-src 'none'; frame-src 'none'; object-src 'none'; base-uri 'none'; form-action 'none'">
          <style>
            :root { color-scheme: light dark; }
            * { box-sizing: border-box; }
            html { background: -apple-system-text-background; }
            body {
              max-width: 780px;
              margin: 0 auto;
              padding: 30px 36px 80px;
              background: -apple-system-text-background;
              color: -apple-system-label;
              font: 16px/1.72 -apple-system, BlinkMacSystemFont, "SF Pro Text", sans-serif;
              overflow-wrap: anywhere;
              -webkit-font-smoothing: antialiased;
            }
            p, ul, ol, blockquote, pre, figure, table { margin: 0 0 1.2em; }
            h1, h2, h3, h4, h5, h6 {
              margin: 1.7em 0 .65em;
              color: -apple-system-label;
              font-weight: 700;
              line-height: 1.25;
            }
            h1 { font-size: 1.65em; }
            h2 { font-size: 1.4em; }
            h3 { font-size: 1.2em; }
            a { color: -apple-system-link; text-decoration: none; }
            a:hover { text-decoration: underline; }
            img, video { display: block; max-width: 100%; height: auto; margin: 1.4em auto; border-radius: 10px; }
            iframe, object, embed, form { display: none !important; }
            blockquote {
              margin-left: 0;
              padding: .2em 0 .2em 1.1em;
              border-left: 3px solid -apple-system-separator;
              color: -apple-system-secondary-label;
            }
            pre, code {
              font-family: "SF Mono", ui-monospace, monospace;
              background: -apple-system-quaternary-fill;
              border-radius: 6px;
            }
            code { padding: .15em .35em; font-size: .9em; }
            pre { padding: 14px; overflow-x: auto; }
            pre code { padding: 0; background: transparent; }
            table { width: 100%; border-collapse: collapse; }
            th, td { padding: 8px 10px; border-bottom: 1px solid -apple-system-separator; text-align: left; }
            hr { border: 0; border-top: 1px solid -apple-system-separator; margin: 2em 0; }
            @media (max-width: 640px) { body { padding: 24px 22px 64px; } }
          </style>
        </head>
        <body>\(sanitized)</body>
        </html>
        """
    }

    static func escape(_ input: String) -> String {
        input
            .replacingOccurrences(of: "&", with: "&amp;")
            .replacingOccurrences(of: "<", with: "&lt;")
            .replacingOccurrences(of: ">", with: "&gt;")
            .replacingOccurrences(of: "\"", with: "&quot;")
            .replacingOccurrences(of: "'", with: "&#39;")
    }

    private static func sanitize(_ input: String) -> String {
        var output = input
        let patterns = [
            #"(?is)<(script|iframe|object|embed|form|base)\b[^>]*>.*?</\1\s*>"#,
            #"(?is)<(script|iframe|object|embed|form|base|meta)\b[^>]*?/?>"#,
            #"(?i)\s+on[a-z]+\s*=\s*(?:\"[^\"]*\"|'[^']*'|[^\s>]+)"#,
            #"(?i)(href|src)\s*=\s*([\"'])\s*javascript:.*?\2"#
        ]

        for pattern in patterns {
            guard let regex = try? NSRegularExpression(pattern: pattern) else { continue }
            let range = NSRange(output.startIndex..., in: output)
            output = regex.stringByReplacingMatches(in: output, range: range, withTemplate: "")
        }
        return output
    }
}
