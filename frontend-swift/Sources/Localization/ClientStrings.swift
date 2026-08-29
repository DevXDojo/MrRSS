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
        "client.sidebar.library": "Library",
        "client.article.showArticle": "Show Article",
        "client.article.translateTitle": "Translate Title",
        "client.article.translateContent": "Translate Content",
        "client.article.more": "More",
        "client.article.noImages": "This article has no images.",
        "client.settings.searchSettings": "Search settings",
        "client.settings.connection": "Connection",
        "client.settings.noResults": "No settings match the search.",
        "client.action.retry": "Retry",
        "client.action.dismiss": "Dismiss",
        "client.action.showInFinder": "Show in Finder",
        "client.feed.invalidURL": "Enter a valid HTTP or HTTPS feed address.",
        "client.sort.oldestFirst": "Oldest first",
        "client.sort.unreadFirst": "Unread first",
        "client.sort.title": "Sort order",
        "client.layout.compact": "Compact",
        "client.layout.comfortable": "Comfortable",
        "client.layout.cards": "Cards",
        "client.layout.title": "Layout",
        "client.folder.nameRequired": "Enter a folder name.",
        "client.folder.alreadyExists": "A folder named {name} already exists.",
        "client.server.invalidAddress": "Enter a valid HTTP or HTTPS server address.",
        "client.settings.saved": "Settings saved.",
        "client.rule.applied": "The rule was applied to {count} articles.",
        "client.ai.limitReachedFallback":
            "The AI usage limit was reached, so a fallback provider was used.",
        "client.ai.usageReset": "AI usage was reset.",
        "client.maintenance.cleared": "Cached translations and summaries were cleared."
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
        "client.sidebar.library": "资料库",
        "client.article.showArticle": "显示文章",
        "client.article.translateTitle": "翻译标题",
        "client.article.translateContent": "翻译正文",
        "client.article.more": "更多",
        "client.article.noImages": "这篇文章没有图片。",
        "client.settings.searchSettings": "搜索设置",
        "client.settings.connection": "连接",
        "client.settings.noResults": "没有符合搜索条件的设置。",
        "client.action.retry": "重试",
        "client.action.dismiss": "关闭",
        "client.action.showInFinder": "在访达中显示",
        "client.feed.invalidURL": "请输入有效的 HTTP 或 HTTPS 订阅地址。",
        "client.sort.oldestFirst": "由旧到新",
        "client.sort.unreadFirst": "未读优先",
        "client.sort.title": "排序方式",
        "client.layout.compact": "紧凑",
        "client.layout.comfortable": "标准",
        "client.layout.cards": "卡片",
        "client.layout.title": "布局",
        "client.folder.nameRequired": "请输入文件夹名称。",
        "client.folder.alreadyExists": "名为 {name} 的文件夹已存在。",
        "client.server.invalidAddress": "请输入有效的 HTTP 或 HTTPS 服务器地址。",
        "client.settings.saved": "设置已保存。",
        "client.rule.applied": "规则已应用于 {count} 篇文章。",
        "client.ai.limitReachedFallback": "已达到 AI 使用上限，因此使用了备用服务。",
        "client.ai.usageReset": "AI 用量已重置。",
        "client.maintenance.cleared": "已清除缓存的翻译和摘要。"
    ]

    static func table(for language: AppLanguage) -> [String: String] {
        switch language {
        case .english: english
        case .chineseSimplified: chineseSimplified
        }
    }
}
