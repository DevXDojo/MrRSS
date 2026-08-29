import Foundation

/// Wording that only the native client needs, kept apart from the catalogue
/// ported from the previous frontend so the two stay easy to tell apart.
enum ClientStrings {
    static let english: [String: String] = [
        "client.error.invalidServerAddress": "The server address is invalid.",
        "client.error.invalidResponse": "The server returned an invalid response.",
        "client.error.httpStatus": "The server returned HTTP {status}.",
        "client.error.unreadableResponse": "The server response could not be read",
        "client.connection.connecting": "Connecting",
        "client.connection.connected": "Connected",
        "client.connection.offline": "Offline",
        "client.connection.serverAddress": "Server address",
        "client.connection.serverAddressHelp":
            "The address of the MrRSS backend, for example http://127.0.0.1:1234.",
        "client.connection.apply": "Apply",
        "client.article.chooseArticle": "Choose an article",
        "client.article.chooseArticleDetail": "Select an article from the list to read it here.",
        "client.article.noArticles": "No articles",
        "client.article.noArticlesDetail": "Nothing matches the current selection.",
        "client.article.loadingContent": "Loading content",
        "client.sidebar.uncategorised": "Uncategorised",
        "client.sidebar.newFolder": "New Folder",
        "client.settings.allSettings": "All Settings",
        "client.settings.searchSettings": "Search settings",
        "client.settings.connection": "Connection",
        "client.settings.noResults": "No settings match the search.",
        "client.action.retry": "Retry",
        "client.action.dismiss": "Dismiss",
        "client.action.showInFinder": "Show in Finder",
        "client.feed.invalidURL": "Enter a valid HTTP or HTTPS feed address."
    ]

    static let chineseSimplified: [String: String] = [
        "client.error.invalidServerAddress": "服务器地址无效。",
        "client.error.invalidResponse": "服务器返回了无效的响应。",
        "client.error.httpStatus": "服务器返回 HTTP {status}。",
        "client.error.unreadableResponse": "无法读取服务器响应",
        "client.connection.connecting": "正在连接",
        "client.connection.connected": "已连接",
        "client.connection.offline": "离线",
        "client.connection.serverAddress": "服务器地址",
        "client.connection.serverAddressHelp": "MrRSS 后端的地址，例如 http://127.0.0.1:1234。",
        "client.connection.apply": "应用",
        "client.article.chooseArticle": "选择一篇文章",
        "client.article.chooseArticleDetail": "从列表中选择一篇文章即可在此阅读。",
        "client.article.noArticles": "没有文章",
        "client.article.noArticlesDetail": "没有符合当前选择的内容。",
        "client.article.loadingContent": "正在载入内容",
        "client.sidebar.uncategorised": "未分类",
        "client.sidebar.newFolder": "新建文件夹",
        "client.settings.allSettings": "全部设置",
        "client.settings.searchSettings": "搜索设置",
        "client.settings.connection": "连接",
        "client.settings.noResults": "没有符合搜索条件的设置。",
        "client.action.retry": "重试",
        "client.action.dismiss": "关闭",
        "client.action.showInFinder": "在访达中显示",
        "client.feed.invalidURL": "请输入有效的 HTTP 或 HTTPS 订阅地址。"
    ]

    static func table(for language: AppLanguage) -> [String: String] {
        switch language {
        case .english: english
        case .chineseSimplified: chineseSimplified
        }
    }
}
