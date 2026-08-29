import Foundation

enum SettingsPane: String, CaseIterable, Identifiable {
    case connection
    case general
    case reading
    case translation
    case summaryAI
    case storageNetwork
    case integrations
    case rules
    case advanced

    var id: String { rawValue }

    var title: String {
        switch self {
        case .connection: "Connection"
        case .general: "General"
        case .reading: "Reading"
        case .translation: "Translation"
        case .summaryAI: "Summary & AI"
        case .storageNetwork: "Storage & Network"
        case .integrations: "Integrations"
        case .rules: "Rules"
        case .advanced: "All Settings"
        }
    }

    var icon: String {
        switch self {
        case .connection: "server.rack"
        case .general: "gearshape"
        case .reading: "text.book.closed"
        case .translation: "character.book.closed"
        case .summaryAI: "sparkles"
        case .storageNetwork: "externaldrive.connected.to.line.below"
        case .integrations: "link"
        case .rules: "bolt.badge.clock"
        case .advanced: "slider.horizontal.3"
        }
    }
}

enum SettingControl {
    case toggle
    case text
    case secure
    case picker([(String, String)])
}

struct SettingDefinition: Identifiable {
    let key: String
    let pane: SettingsPane
    let section: String
    let control: SettingControl
    let detail: String?

    var id: String { key }
    var title: String { key.settingTitle }
}

enum SettingsCatalog {
    static let definitions: [SettingDefinition] = [
        d("update_interval", .general, "Updates", .text, "Automatic refresh interval in minutes."),
        d("refresh_mode", .general, "Updates", .picker([("fixed", "Fixed interval"), ("smart", "Intelligent")])),
        d("language", .general, "Appearance", .picker([("en-US", "English"), ("zh-CN", "简体中文")])),
        d("theme", .general, "Appearance", .picker([("auto", "System"), ("light", "Light"), ("dark", "Dark")])),
        d("startup_on_boot", .general, "Application", .toggle),
        d("close_to_tray", .general, "Application", .toggle),
        d("auto_update", .general, "Application", .toggle),
        d("shortcuts_enabled", .general, "Application", .toggle),
        d("image_gallery_enabled", .general, "Application", .toggle),

        d("default_view_mode", .reading, "Article display", .picker([("rendered", "Rendered content"), ("webpage", "Original webpage")])),
        d("show_hidden_articles", .reading, "Article display", .toggle),
        d("hover_mark_as_read", .reading, "Article display", .toggle),
        d("show_article_preview_images", .reading, "Article display", .toggle),
        d("full_text_fetch_enabled", .reading, "Full text", .toggle),
        d("auto_show_all_content", .reading, "Full text", .toggle),
        d("custom_css_file", .reading, "Customization", .text),

        d("translation_enabled", .translation, "Translation", .toggle),
        d("target_language", .translation, "Translation", .picker([("zh", "Chinese"), ("en", "English"), ("ja", "Japanese"), ("ko", "Korean"), ("fr", "French"), ("de", "German"), ("es", "Spanish")])),
        d("translation_provider", .translation, "Translation", .picker([("google", "Google"), ("deepl", "DeepL"), ("baidu", "Baidu"), ("ai", "AI")])),
        d("google_translate_endpoint", .translation, "Google", .text),
        d("deepl_endpoint", .translation, "DeepL", .text),
        d("deepl_api_key", .translation, "DeepL", .secure),
        d("baidu_app_id", .translation, "Baidu", .text),
        d("baidu_secret_key", .translation, "Baidu", .secure),

        d("summary_enabled", .summaryAI, "Summary", .toggle),
        d("summary_length", .summaryAI, "Summary", .picker([("short", "Short"), ("medium", "Medium"), ("long", "Long")])),
        d("summary_provider", .summaryAI, "Summary", .picker([("local", "Local TextRank"), ("ai", "AI")])),
        d("summary_trigger_mode", .summaryAI, "Summary", .picker([("manual", "Manual"), ("auto", "Automatic")])),
        d("ai_api_key", .summaryAI, "AI provider", .secure),
        d("ai_endpoint", .summaryAI, "AI provider", .text),
        d("ai_model", .summaryAI, "AI provider", .text),
        d("ai_usage_limit", .summaryAI, "AI provider", .text),
        d("ai_chat_enabled", .summaryAI, "AI provider", .toggle),
        d("ai_custom_headers", .summaryAI, "AI prompts", .text),
        d("ai_translation_prompt", .summaryAI, "AI prompts", .text),
        d("ai_summary_prompt", .summaryAI, "AI prompts", .text),

        d("auto_cleanup_enabled", .storageNetwork, "Storage", .toggle),
        d("max_cache_size_mb", .storageNetwork, "Storage", .text),
        d("max_article_age_days", .storageNetwork, "Storage", .text),
        d("media_cache_enabled", .storageNetwork, "Media cache", .toggle),
        d("media_proxy_fallback", .storageNetwork, "Media cache", .toggle),
        d("media_cache_max_size_mb", .storageNetwork, "Media cache", .text),
        d("media_cache_max_age_days", .storageNetwork, "Media cache", .text),
        d("proxy_enabled", .storageNetwork, "Proxy", .toggle),
        d("proxy_type", .storageNetwork, "Proxy", .picker([("http", "HTTP"), ("https", "HTTPS"), ("socks5", "SOCKS5")])),
        d("proxy_host", .storageNetwork, "Proxy", .text),
        d("proxy_port", .storageNetwork, "Proxy", .text),
        d("proxy_username", .storageNetwork, "Proxy", .secure),
        d("proxy_password", .storageNetwork, "Proxy", .secure),
        d("max_concurrent_refreshes", .storageNetwork, "Network", .text),
        d("retry_timeout_seconds", .storageNetwork, "Network", .text),

        d("obsidian_enabled", .integrations, "Obsidian", .toggle),
        d("obsidian_vault", .integrations, "Obsidian", .text),
        d("obsidian_vault_path", .integrations, "Obsidian", .text),
        d("freshrss_enabled", .integrations, "FreshRSS", .toggle),
        d("freshrss_server_url", .integrations, "FreshRSS", .text),
        d("freshrss_username", .integrations, "FreshRSS", .text),
        d("freshrss_api_password", .integrations, "FreshRSS", .secure),
        d("freshrss_auto_sync_interval", .integrations, "FreshRSS", .text),
        d("freshrss_sync_on_startup", .integrations, "FreshRSS", .toggle)
    ]

    static let secureKeys: Set<String> = [
        "ai_api_key", "baidu_secret_key", "deepl_api_key", "freshrss_api_password",
        "proxy_password", "proxy_username"
    ]

    static let boolKeys: Set<String> = Set(definitions.compactMap { definition in
        if case .toggle = definition.control { return definition.key }
        return nil
    })

    static func definitions(for pane: SettingsPane) -> [SettingDefinition] {
        definitions.filter { $0.pane == pane }
    }

    private static func d(
        _ key: String,
        _ pane: SettingsPane,
        _ section: String,
        _ control: SettingControl,
        _ detail: String? = nil
    ) -> SettingDefinition {
        SettingDefinition(key: key, pane: pane, section: section, control: control, detail: detail)
    }
}

private extension String {
    var settingTitle: String {
        split(separator: "_")
            .map { word in
                switch word.lowercased() {
                case "ai": "AI"
                case "api": "API"
                case "rss": "RSS"
                case "url": "URL"
                case "css": "CSS"
                case "mb": "MB"
                default: word.capitalized
                }
            }
            .joined(separator: " ")
    }
}
