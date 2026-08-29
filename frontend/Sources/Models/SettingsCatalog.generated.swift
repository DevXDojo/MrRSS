// Generated from internal/config/settings_schema.json.
// Regenerate with: python3 tools/settings-swift/generate.py

import Foundation

extension SettingsCatalog {
    /// Every setting the backend stores, paired with the wording the
    /// previous interface used for it.
    static let generated: [SettingDefinition] = [
        SettingDefinition(
            key: "ai_api_key",
            pane: .ai,
            control: .secret,
            defaultValue: "",
            titleKey: "setting.ai.aiApiKey",
            fallbackTitle: "AI API Key",
            detailKey: "setting.ai.aiApiKeyDesc",
            choices: []
        ),
        SettingDefinition(
            key: "ai_chat_enabled",
            pane: .ai,
            control: .toggle,
            defaultValue: "false",
            titleKey: "setting.ai.aiChatEnabled",
            fallbackTitle: "AI Chat Enabled",
            detailKey: "setting.ai.aiChatEnabledDesc",
            choices: []
        ),
        SettingDefinition(
            key: "ai_chat_profile_id",
            pane: .ai,
            control: .text,
            defaultValue: "",
            titleKey: "setting.ai.selectProfileForChat",
            fallbackTitle: "AI Chat Profile ID",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "ai_custom_headers",
            pane: .ai,
            control: .text,
            defaultValue: "",
            titleKey: "setting.ai.aiCustomHeaders",
            fallbackTitle: "AI Custom Headers",
            detailKey: "setting.ai.aiCustomHeadersDesc",
            choices: []
        ),
        SettingDefinition(
            key: "ai_endpoint",
            pane: .ai,
            control: .text,
            defaultValue: "https://api.openai.com/v1/chat/completions",
            titleKey: "setting.ai.aiEndpoint",
            fallbackTitle: "AI Endpoint",
            detailKey: "setting.ai.aiEndpointDesc",
            choices: []
        ),
        SettingDefinition(
            key: "ai_model",
            pane: .ai,
            control: .text,
            defaultValue: "gpt-4o-mini",
            titleKey: "setting.ai.aiModel",
            fallbackTitle: "AI Model",
            detailKey: "setting.ai.aiModelDesc",
            choices: []
        ),
        SettingDefinition(
            key: "ai_search_enabled",
            pane: .ai,
            control: .toggle,
            defaultValue: "false",
            titleKey: "setting.ai.aiSearchEnabled",
            fallbackTitle: "AI Search Enabled",
            detailKey: "setting.ai.aiSearchEnabledDesc",
            choices: []
        ),
        SettingDefinition(
            key: "ai_search_profile_id",
            pane: .ai,
            control: .text,
            defaultValue: "",
            titleKey: "setting.ai.selectProfileForSearch",
            fallbackTitle: "AI Search Profile ID",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "ai_summary_profile_id",
            pane: .ai,
            control: .text,
            defaultValue: "",
            titleKey: "setting.ai.selectProfileForSummary",
            fallbackTitle: "AI Summary Profile ID",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "ai_summary_prompt",
            pane: .ai,
            control: .text,
            defaultValue: "You are a summarizer. Generate a concise summary of the given text. Output ONLY the summary, nothing else.",
            titleKey: "setting.content.aiSummaryPrompt",
            fallbackTitle: "AI Summary Prompt",
            detailKey: "setting.content.aiSummaryPromptDesc",
            choices: []
        ),
        SettingDefinition(
            key: "ai_translation_profile_id",
            pane: .ai,
            control: .text,
            defaultValue: "",
            titleKey: "setting.ai.selectProfileForTranslation",
            fallbackTitle: "AI Translation Profile ID",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "ai_translation_prompt",
            pane: .ai,
            control: .text,
            defaultValue: "You are a translator. Translate the given text accurately. Output ONLY the translated text, nothing else.",
            titleKey: "setting.content.aiTranslationPrompt",
            fallbackTitle: "AI Translation Prompt",
            detailKey: "setting.content.aiTranslationPromptDesc",
            choices: []
        ),
        SettingDefinition(
            key: "ai_usage_limit",
            pane: .ai,
            control: .text,
            defaultValue: "20000",
            titleKey: "setting.ai.setUsageLimit",
            fallbackTitle: "AI Usage Limit",
            detailKey: "setting.ai.setUsageLimitDesc",
            choices: []
        ),
        SettingDefinition(
            key: "ai_usage_tokens",
            pane: .ai,
            control: .text,
            defaultValue: "0",
            titleKey: "setting.ai.aiUsageTokens",
            fallbackTitle: "AI Usage Tokens",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "auto_cleanup_enabled",
            pane: .storage,
            control: .toggle,
            defaultValue: "true",
            titleKey: "setting.database.autoCleanup",
            fallbackTitle: "Auto Cleanup Enabled",
            detailKey: "setting.database.autoCleanupDesc",
            choices: []
        ),
        SettingDefinition(
            key: "auto_show_all_content",
            pane: .reading,
            control: .toggle,
            defaultValue: "false",
            titleKey: "setting.reading.autoShowAllContent",
            fallbackTitle: "Auto Show All Content",
            detailKey: "setting.reading.autoShowAllContentDesc",
            choices: []
        ),
        SettingDefinition(
            key: "baidu_app_id",
            pane: .translation,
            control: .text,
            defaultValue: "",
            titleKey: "setting.content.baiduAppId",
            fallbackTitle: "Baidu App ID",
            detailKey: "setting.content.baiduAppIdDesc",
            choices: []
        ),
        SettingDefinition(
            key: "baidu_secret_key",
            pane: .translation,
            control: .secret,
            defaultValue: "",
            titleKey: "setting.content.baiduSecretKey",
            fallbackTitle: "Baidu Secret Key",
            detailKey: "setting.content.baiduSecretKeyDesc",
            choices: []
        ),
        SettingDefinition(
            key: "close_to_tray",
            pane: .general,
            control: .toggle,
            defaultValue: "true",
            titleKey: "setting.general.closeToTray",
            fallbackTitle: "Close To Tray",
            detailKey: "setting.general.closeToTrayDesc",
            choices: []
        ),
        SettingDefinition(
            key: "content_font_family",
            pane: .typography,
            control: .choice,
            defaultValue: "system",
            titleKey: "setting.typography.contentFontFamily",
            fallbackTitle: "Content Font Family",
            detailKey: "setting.typography.contentFontFamilyDesc",
            choices: [SettingChoice(value: "system", titleKey: "setting.typography.fontSystem"), SettingChoice(value: "serif", titleKey: "setting.typography.fontSerif"), SettingChoice(value: "monospace", titleKey: "setting.typography.fontMonospace")]
        ),
        SettingDefinition(
            key: "content_font_size",
            pane: .typography,
            control: .number,
            defaultValue: "16",
            titleKey: "setting.typography.contentFontSize",
            fallbackTitle: "Content Font Size",
            detailKey: "setting.typography.contentFontSizeDesc",
            choices: []
        ),
        SettingDefinition(
            key: "content_line_height",
            pane: .typography,
            control: .text,
            defaultValue: "1.6",
            titleKey: "setting.typography.contentLineHeight",
            fallbackTitle: "Content Line Height",
            detailKey: "setting.typography.contentLineHeightDesc",
            choices: []
        ),
        SettingDefinition(
            key: "custom_css_file",
            pane: .customization,
            control: .text,
            defaultValue: "",
            titleKey: nil,
            fallbackTitle: "Custom CSS File",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "custom_translation_body_template",
            pane: .translation,
            control: .text,
            defaultValue: "",
            titleKey: "setting.translation.custom.bodyTemplate",
            fallbackTitle: "Custom Translation Body Template",
            detailKey: "setting.translation.custom.bodyTemplateDesc",
            choices: []
        ),
        SettingDefinition(
            key: "custom_translation_enabled",
            pane: .translation,
            control: .toggle,
            defaultValue: "false",
            titleKey: nil,
            fallbackTitle: "Custom Translation Enabled",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "custom_translation_endpoint",
            pane: .translation,
            control: .text,
            defaultValue: "",
            titleKey: "setting.translation.custom.endpoint",
            fallbackTitle: "Custom Translation Endpoint",
            detailKey: "setting.translation.custom.endpointDesc",
            choices: []
        ),
        SettingDefinition(
            key: "custom_translation_headers",
            pane: .translation,
            control: .text,
            defaultValue: "",
            titleKey: "setting.translation.custom.headers",
            fallbackTitle: "Custom Translation Headers",
            detailKey: "setting.translation.custom.headersDesc",
            choices: []
        ),
        SettingDefinition(
            key: "custom_translation_lang_mapping",
            pane: .translation,
            control: .text,
            defaultValue: "",
            titleKey: "setting.translation.custom.langMapping",
            fallbackTitle: "Custom Translation Lang Mapping",
            detailKey: "setting.translation.custom.langMappingDesc",
            choices: []
        ),
        SettingDefinition(
            key: "custom_translation_method",
            pane: .translation,
            control: .choice,
            defaultValue: "POST",
            titleKey: "setting.translation.custom.method",
            fallbackTitle: "Custom Translation Method",
            detailKey: "setting.translation.custom.methodDesc",
            choices: [SettingChoice(value: "POST", titleKey: nil), SettingChoice(value: "GET", titleKey: nil)]
        ),
        SettingDefinition(
            key: "custom_translation_name",
            pane: .translation,
            control: .text,
            defaultValue: "",
            titleKey: nil,
            fallbackTitle: "Custom Translation Name",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "custom_translation_response_path",
            pane: .translation,
            control: .text,
            defaultValue: "",
            titleKey: "setting.translation.custom.responsePath",
            fallbackTitle: "Custom Translation Response Path",
            detailKey: "setting.translation.custom.responsePathDesc",
            choices: []
        ),
        SettingDefinition(
            key: "custom_translation_timeout",
            pane: .translation,
            control: .number,
            defaultValue: "10",
            titleKey: "setting.translation.custom.timeout",
            fallbackTitle: "Custom Translation Timeout",
            detailKey: "setting.translation.custom.timeoutDesc",
            choices: []
        ),
        SettingDefinition(
            key: "deepl_api_key",
            pane: .translation,
            control: .secret,
            defaultValue: "",
            titleKey: "setting.content.deeplApiKey",
            fallbackTitle: "Deepl API Key",
            detailKey: "setting.content.deeplApiKeyDesc",
            choices: []
        ),
        SettingDefinition(
            key: "deepl_endpoint",
            pane: .translation,
            control: .text,
            defaultValue: "",
            titleKey: "setting.content.deeplEndpoint",
            fallbackTitle: "Deepl Endpoint",
            detailKey: "setting.content.deeplEndpointDesc",
            choices: []
        ),
        SettingDefinition(
            key: "default_view_mode",
            pane: .reading,
            control: .choice,
            defaultValue: "rendered",
            titleKey: "setting.reading.defaultViewMode",
            fallbackTitle: "Default View Mode",
            detailKey: "setting.reading.defaultViewModeDesc",
            choices: [SettingChoice(value: "rendered", titleKey: "setting.reading.viewAsRendered"), SettingChoice(value: "webpage", titleKey: "setting.reading.viewAsWebpage")]
        ),
        SettingDefinition(
            key: "freshrss_api_password",
            pane: .integrations,
            control: .secret,
            defaultValue: "",
            titleKey: "setting.freshrss.apiPassword",
            fallbackTitle: "Freshrss API Password",
            detailKey: "setting.freshrss.apiPasswordDesc",
            choices: []
        ),
        SettingDefinition(
            key: "freshrss_auto_sync_interval",
            pane: .integrations,
            control: .number,
            defaultValue: "0",
            titleKey: nil,
            fallbackTitle: "Freshrss Auto Sync Interval",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "freshrss_enabled",
            pane: .integrations,
            control: .toggle,
            defaultValue: "false",
            titleKey: "setting.freshrss.enabled",
            fallbackTitle: "Freshrss Enabled",
            detailKey: "setting.freshrss.enabledDesc",
            choices: []
        ),
        SettingDefinition(
            key: "freshrss_server_url",
            pane: .integrations,
            control: .text,
            defaultValue: "",
            titleKey: "setting.freshrss.serverUrl",
            fallbackTitle: "Freshrss Server URL",
            detailKey: "setting.freshrss.serverUrlDesc",
            choices: []
        ),
        SettingDefinition(
            key: "freshrss_sync_on_startup",
            pane: .integrations,
            control: .toggle,
            defaultValue: "false",
            titleKey: nil,
            fallbackTitle: "Freshrss Sync On Startup",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "freshrss_username",
            pane: .integrations,
            control: .text,
            defaultValue: "",
            titleKey: "setting.freshrss.username",
            fallbackTitle: "Freshrss Username",
            detailKey: "setting.freshrss.usernameDesc",
            choices: []
        ),
        SettingDefinition(
            key: "full_text_fetch_enabled",
            pane: .reading,
            control: .toggle,
            defaultValue: "true",
            titleKey: "setting.feed.enableFullTextFetch",
            fallbackTitle: "Full Text Fetch Enabled",
            detailKey: "setting.feed.enableFullTextFetchDesc",
            choices: []
        ),
        SettingDefinition(
            key: "google_translate_endpoint",
            pane: .translation,
            control: .text,
            defaultValue: "translate.googleapis.com",
            titleKey: "setting.content.googleTranslateEndpoint",
            fallbackTitle: "Google Translate Endpoint",
            detailKey: "setting.content.googleTranslateEndpointDesc",
            choices: []
        ),
        SettingDefinition(
            key: "hover_mark_as_read",
            pane: .reading,
            control: .toggle,
            defaultValue: "false",
            titleKey: "setting.reading.hoverMarkAsRead",
            fallbackTitle: "Hover Mark As Read",
            detailKey: "setting.reading.hoverMarkAsReadDesc",
            choices: []
        ),
        SettingDefinition(
            key: "image_gallery_enabled",
            pane: .general,
            control: .toggle,
            defaultValue: "true",
            titleKey: "setting.reading.imageGalleryEnabled",
            fallbackTitle: "Image Gallery Enabled",
            detailKey: "setting.reading.imageGalleryEnabledDesc",
            choices: []
        ),
        SettingDefinition(
            key: "language",
            pane: .general,
            control: .choice,
            defaultValue: "en-US",
            titleKey: "setting.general.language",
            fallbackTitle: "Language",
            detailKey: "setting.general.languageDesc",
            choices: [SettingChoice(value: "en-US", titleKey: nil), SettingChoice(value: "zh-CN", titleKey: nil)]
        ),
        SettingDefinition(
            key: "layout_mode",
            pane: .customization,
            control: .choice,
            defaultValue: "normal",
            titleKey: "setting.typography.layoutMode",
            fallbackTitle: "Layout Mode",
            detailKey: "setting.typography.layoutModeDesc",
            choices: [SettingChoice(value: "normal", titleKey: nil), SettingChoice(value: "compact", titleKey: nil), SettingChoice(value: "wide", titleKey: nil)]
        ),
        SettingDefinition(
            key: "max_article_age_days",
            pane: .storage,
            control: .number,
            defaultValue: "30",
            titleKey: "setting.database.maxArticleAge",
            fallbackTitle: "Max Article Age Days",
            detailKey: "setting.database.maxArticleAgeDesc",
            choices: []
        ),
        SettingDefinition(
            key: "max_cache_size_mb",
            pane: .storage,
            control: .number,
            defaultValue: "500",
            titleKey: "setting.database.maxCacheSize",
            fallbackTitle: "Max Cache Size MB",
            detailKey: "setting.database.maxCacheSizeDesc",
            choices: []
        ),
        SettingDefinition(
            key: "max_concurrent_refreshes",
            pane: .network,
            control: .text,
            defaultValue: "5",
            titleKey: nil,
            fallbackTitle: "Max Concurrent Refreshes",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "media_cache_enabled",
            pane: .storage,
            control: .toggle,
            defaultValue: "false",
            titleKey: "setting.database.mediaCacheEnabled",
            fallbackTitle: "Media Cache Enabled",
            detailKey: "setting.database.mediaCacheEnabledDesc",
            choices: []
        ),
        SettingDefinition(
            key: "media_cache_max_age_days",
            pane: .storage,
            control: .number,
            defaultValue: "7",
            titleKey: "setting.database.mediaCacheMaxAge",
            fallbackTitle: "Media Cache Max Age Days",
            detailKey: "setting.database.mediaCacheMaxAgeDesc",
            choices: []
        ),
        SettingDefinition(
            key: "media_cache_max_size_mb",
            pane: .storage,
            control: .number,
            defaultValue: "200",
            titleKey: "setting.database.mediaCacheMaxSize",
            fallbackTitle: "Media Cache Max Size MB",
            detailKey: "setting.database.mediaCacheMaxSizeDesc",
            choices: []
        ),
        SettingDefinition(
            key: "media_proxy_fallback",
            pane: .storage,
            control: .toggle,
            defaultValue: "true",
            titleKey: nil,
            fallbackTitle: "Media Proxy Fallback",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "microsoft_api_key",
            pane: .translation,
            control: .secret,
            defaultValue: "",
            titleKey: "setting.content.microsoftApiKey",
            fallbackTitle: "Microsoft API Key",
            detailKey: "setting.content.microsoftApiKeyDesc",
            choices: []
        ),
        SettingDefinition(
            key: "microsoft_endpoint",
            pane: .translation,
            control: .text,
            defaultValue: "",
            titleKey: "setting.content.microsoftEndpoint",
            fallbackTitle: "Microsoft Endpoint",
            detailKey: "setting.content.microsoftEndpointDesc",
            choices: []
        ),
        SettingDefinition(
            key: "microsoft_region",
            pane: .translation,
            control: .text,
            defaultValue: "",
            titleKey: "setting.content.microsoftRegion",
            fallbackTitle: "Microsoft Region",
            detailKey: "setting.content.microsoftRegionDesc",
            choices: []
        ),
        SettingDefinition(
            key: "notion_api_key",
            pane: .integrations,
            control: .secret,
            defaultValue: "",
            titleKey: "setting.plugins.notion.apiKey",
            fallbackTitle: "Notion API Key",
            detailKey: "setting.plugins.notion.apiKeyDesc",
            choices: []
        ),
        SettingDefinition(
            key: "notion_enabled",
            pane: .integrations,
            control: .toggle,
            defaultValue: "false",
            titleKey: nil,
            fallbackTitle: "Notion Enabled",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "notion_page_id",
            pane: .integrations,
            control: .text,
            defaultValue: "",
            titleKey: "setting.plugins.notion.pageId",
            fallbackTitle: "Notion Page ID",
            detailKey: "setting.plugins.notion.pageIdDesc",
            choices: []
        ),
        SettingDefinition(
            key: "obsidian_enabled",
            pane: .integrations,
            control: .toggle,
            defaultValue: "false",
            titleKey: nil,
            fallbackTitle: "Obsidian Enabled",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "obsidian_vault",
            pane: .integrations,
            control: .text,
            defaultValue: "",
            titleKey: "setting.plugins.obsidian.vaultName",
            fallbackTitle: "Obsidian Vault",
            detailKey: "setting.plugins.obsidian.vaultNameDesc",
            choices: []
        ),
        SettingDefinition(
            key: "obsidian_vault_path",
            pane: .integrations,
            control: .text,
            defaultValue: "",
            titleKey: "setting.plugins.obsidian.vaultPath",
            fallbackTitle: "Obsidian Vault Path",
            detailKey: "setting.plugins.obsidian.vaultPathDesc",
            choices: []
        ),
        SettingDefinition(
            key: "proxy_enabled",
            pane: .network,
            control: .toggle,
            defaultValue: "false",
            titleKey: "setting.network.enableProxy",
            fallbackTitle: "Proxy Enabled",
            detailKey: "setting.network.enableProxyDesc",
            choices: []
        ),
        SettingDefinition(
            key: "proxy_host",
            pane: .network,
            control: .text,
            defaultValue: "127.0.0.1",
            titleKey: "setting.network.proxyHost",
            fallbackTitle: "Proxy Host",
            detailKey: "setting.network.proxyHostDesc",
            choices: []
        ),
        SettingDefinition(
            key: "proxy_password",
            pane: .network,
            control: .secret,
            defaultValue: "",
            titleKey: "setting.network.proxyPassword",
            fallbackTitle: "Proxy Password",
            detailKey: "setting.network.proxyPasswordDesc",
            choices: []
        ),
        SettingDefinition(
            key: "proxy_port",
            pane: .network,
            control: .text,
            defaultValue: "7890",
            titleKey: "setting.network.proxyPort",
            fallbackTitle: "Proxy Port",
            detailKey: "setting.network.proxyPortDesc",
            choices: []
        ),
        SettingDefinition(
            key: "proxy_type",
            pane: .network,
            control: .choice,
            defaultValue: "https",
            titleKey: "setting.network.proxyType",
            fallbackTitle: "Proxy Type",
            detailKey: "setting.network.proxyTypeDesc",
            choices: [SettingChoice(value: "http", titleKey: nil), SettingChoice(value: "https", titleKey: nil), SettingChoice(value: "socks5", titleKey: nil)]
        ),
        SettingDefinition(
            key: "proxy_username",
            pane: .network,
            control: .secret,
            defaultValue: "",
            titleKey: "setting.network.proxyUsername",
            fallbackTitle: "Proxy Username",
            detailKey: "setting.network.proxyUsernameDesc",
            choices: []
        ),
        SettingDefinition(
            key: "refresh_mode",
            pane: .general,
            control: .choice,
            defaultValue: "fixed",
            titleKey: "setting.feed.refreshMode",
            fallbackTitle: "Refresh Mode",
            detailKey: "setting.feed.refreshModeDesc",
            choices: [SettingChoice(value: "fixed", titleKey: "setting.feed.fixedInterval"), SettingChoice(value: "smart", titleKey: "setting.feed.intelligentInterval")]
        ),
        SettingDefinition(
            key: "retry_timeout_seconds",
            pane: .network,
            control: .number,
            defaultValue: "60",
            titleKey: "setting.feed.retryTimeout",
            fallbackTitle: "Retry Timeout Seconds",
            detailKey: "setting.feed.retryTimeoutDesc",
            choices: []
        ),
        SettingDefinition(
            key: "rsshub_api_key",
            pane: .integrations,
            control: .secret,
            defaultValue: "",
            titleKey: "setting.rsshub.apiKey",
            fallbackTitle: "Rsshub API Key",
            detailKey: "setting.rsshub.apiKeyDesc",
            choices: []
        ),
        SettingDefinition(
            key: "rsshub_enabled",
            pane: .integrations,
            control: .toggle,
            defaultValue: "false",
            titleKey: "setting.rsshub.enabled",
            fallbackTitle: "Rsshub Enabled",
            detailKey: "setting.rsshub.enabledDesc",
            choices: []
        ),
        SettingDefinition(
            key: "rsshub_endpoint",
            pane: .integrations,
            control: .text,
            defaultValue: "https://rss.spriple.org",
            titleKey: "setting.rsshub.endpoint",
            fallbackTitle: "Rsshub Endpoint",
            detailKey: "setting.rsshub.endpointDesc",
            choices: []
        ),
        SettingDefinition(
            key: "shortcuts",
            pane: .general,
            control: .text,
            defaultValue: "",
            titleKey: "setting.shortcut.shortcuts",
            fallbackTitle: "Shortcuts",
            detailKey: "setting.shortcut.shortcutsDesc",
            choices: []
        ),
        SettingDefinition(
            key: "shortcuts_enabled",
            pane: .general,
            control: .toggle,
            defaultValue: "true",
            titleKey: "setting.shortcut.shortcutsEnabled",
            fallbackTitle: "Shortcuts Enabled",
            detailKey: "setting.shortcut.shortcutsEnabledDesc",
            choices: []
        ),
        SettingDefinition(
            key: "show_article_preview_images",
            pane: .reading,
            control: .toggle,
            defaultValue: "true",
            titleKey: "setting.reading.showArticlePreviewImages",
            fallbackTitle: "Show Article Preview Images",
            detailKey: "setting.reading.showArticlePreviewImagesDesc",
            choices: []
        ),
        SettingDefinition(
            key: "show_floating_toc",
            pane: .reading,
            control: .toggle,
            defaultValue: "false",
            titleKey: "setting.reading.showFloatingToc",
            fallbackTitle: "Show Floating Toc",
            detailKey: "setting.reading.showFloatingTocDesc",
            choices: []
        ),
        SettingDefinition(
            key: "show_hidden_articles",
            pane: .reading,
            control: .toggle,
            defaultValue: "false",
            titleKey: "setting.reading.showHiddenArticles",
            fallbackTitle: "Show Hidden Articles",
            detailKey: "setting.reading.showHiddenArticlesDesc",
            choices: []
        ),
        SettingDefinition(
            key: "startup_on_boot",
            pane: .general,
            control: .toggle,
            defaultValue: "false",
            titleKey: "setting.general.startupOnBoot",
            fallbackTitle: "Startup On Boot",
            detailKey: "setting.general.startupOnBootDesc",
            choices: []
        ),
        SettingDefinition(
            key: "summary_enabled",
            pane: .summary,
            control: .toggle,
            defaultValue: "true",
            titleKey: "setting.content.enableSummary",
            fallbackTitle: "Summary Enabled",
            detailKey: "setting.content.enableSummaryDesc",
            choices: []
        ),
        SettingDefinition(
            key: "summary_length",
            pane: .summary,
            control: .choice,
            defaultValue: "medium",
            titleKey: "setting.content.summaryLength",
            fallbackTitle: "Summary Length",
            detailKey: "setting.content.summaryLengthDesc",
            choices: [SettingChoice(value: "short", titleKey: nil), SettingChoice(value: "medium", titleKey: nil), SettingChoice(value: "long", titleKey: nil)]
        ),
        SettingDefinition(
            key: "summary_provider",
            pane: .summary,
            control: .choice,
            defaultValue: "local",
            titleKey: "setting.content.summaryProvider",
            fallbackTitle: "Summary Provider",
            detailKey: "setting.content.summaryProviderDesc",
            choices: [SettingChoice(value: "local", titleKey: nil), SettingChoice(value: "ai", titleKey: "setting.content.aiSummary")]
        ),
        SettingDefinition(
            key: "summary_trigger_mode",
            pane: .summary,
            control: .choice,
            defaultValue: "manual",
            titleKey: "setting.content.summaryTriggerMode",
            fallbackTitle: "Summary Trigger Mode",
            detailKey: "setting.content.summaryTriggerModeDesc",
            choices: [SettingChoice(value: "manual", titleKey: nil), SettingChoice(value: "auto", titleKey: nil)]
        ),
        SettingDefinition(
            key: "target_language",
            pane: .translation,
            control: .text,
            defaultValue: "zh",
            titleKey: "setting.content.targetLanguage",
            fallbackTitle: "Target Language",
            detailKey: "setting.content.targetLanguageDesc",
            choices: []
        ),
        SettingDefinition(
            key: "tencent_region",
            pane: .translation,
            control: .text,
            defaultValue: "ap-guangzhou",
            titleKey: "setting.content.tencentRegion",
            fallbackTitle: "Tencent Region",
            detailKey: "setting.content.tencentRegionDesc",
            choices: []
        ),
        SettingDefinition(
            key: "tencent_secret_id",
            pane: .translation,
            control: .text,
            defaultValue: "",
            titleKey: "setting.content.tencentSecretId",
            fallbackTitle: "Tencent Secret ID",
            detailKey: "setting.content.tencentSecretIdDesc",
            choices: []
        ),
        SettingDefinition(
            key: "tencent_secret_key",
            pane: .translation,
            control: .secret,
            defaultValue: "",
            titleKey: "setting.content.tencentSecretKey",
            fallbackTitle: "Tencent Secret Key",
            detailKey: "setting.content.tencentSecretKeyDesc",
            choices: []
        ),
        SettingDefinition(
            key: "theme",
            pane: .general,
            control: .choice,
            defaultValue: "auto",
            titleKey: "setting.general.theme",
            fallbackTitle: "Theme",
            detailKey: "setting.general.themeDesc",
            choices: [SettingChoice(value: "auto", titleKey: nil), SettingChoice(value: "light", titleKey: nil), SettingChoice(value: "dark", titleKey: nil)]
        ),
        SettingDefinition(
            key: "translation_enabled",
            pane: .translation,
            control: .toggle,
            defaultValue: "false",
            titleKey: "setting.content.enableTranslation",
            fallbackTitle: "Translation Enabled",
            detailKey: "setting.content.enableTranslationDesc",
            choices: []
        ),
        SettingDefinition(
            key: "translation_only_mode",
            pane: .translation,
            control: .toggle,
            defaultValue: "false",
            titleKey: "setting.content.translationOnlyMode",
            fallbackTitle: "Translation Only Mode",
            detailKey: "setting.content.translationOnlyModeDesc",
            choices: []
        ),
        SettingDefinition(
            key: "translation_provider",
            pane: .translation,
            control: .choice,
            defaultValue: "google",
            titleKey: "setting.content.translationProvider",
            fallbackTitle: "Translation Provider",
            detailKey: "setting.content.translationProviderDesc",
            choices: [SettingChoice(value: "google", titleKey: "setting.content.googleTranslate"), SettingChoice(value: "deepl", titleKey: nil), SettingChoice(value: "baidu", titleKey: "setting.content.baiduTranslate"), SettingChoice(value: "microsoft", titleKey: "setting.content.microsoftTranslate"), SettingChoice(value: "tencent", titleKey: "setting.content.tencentTranslate"), SettingChoice(value: "ai", titleKey: "setting.content.aiTranslation"), SettingChoice(value: "custom", titleKey: nil)]
        ),
        SettingDefinition(
            key: "ui_font_family",
            pane: .typography,
            control: .choice,
            defaultValue: "system",
            titleKey: "setting.general.uiFontFamily",
            fallbackTitle: "UI Font Family",
            detailKey: "setting.general.uiFontFamilyDesc",
            choices: [SettingChoice(value: "system", titleKey: "setting.typography.fontSystem"), SettingChoice(value: "serif", titleKey: "setting.typography.fontSerif"), SettingChoice(value: "monospace", titleKey: "setting.typography.fontMonospace")]
        ),
        SettingDefinition(
            key: "ui_font_size",
            pane: .typography,
            control: .number,
            defaultValue: "16",
            titleKey: "setting.general.uiFontSize",
            fallbackTitle: "UI Font Size",
            detailKey: "setting.general.uiFontSizeDesc",
            choices: []
        ),
        SettingDefinition(
            key: "update_check_enabled",
            pane: .general,
            control: .toggle,
            defaultValue: "true",
            titleKey: "setting.update.updateCheckEnabled",
            fallbackTitle: "Update Check Enabled",
            detailKey: "setting.update.updateCheckEnabledDesc",
            choices: []
        ),
        SettingDefinition(
            key: "update_interval",
            pane: .general,
            control: .number,
            defaultValue: "30",
            titleKey: nil,
            fallbackTitle: "Update Interval",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "zotero_api_key",
            pane: .integrations,
            control: .secret,
            defaultValue: "",
            titleKey: "setting.plugins.zotero.apiKey",
            fallbackTitle: "Zotero API Key",
            detailKey: "setting.plugins.zotero.apiKeyDesc",
            choices: []
        ),
        SettingDefinition(
            key: "zotero_enabled",
            pane: .integrations,
            control: .toggle,
            defaultValue: "false",
            titleKey: nil,
            fallbackTitle: "Zotero Enabled",
            detailKey: nil,
            choices: []
        ),
        SettingDefinition(
            key: "zotero_user_id",
            pane: .integrations,
            control: .text,
            defaultValue: "",
            titleKey: "setting.plugins.zotero.userId",
            fallbackTitle: "Zotero User ID",
            detailKey: "setting.plugins.zotero.userIdDesc",
            choices: []
        )
    ]
}
