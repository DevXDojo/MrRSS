package settings

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"MrRSS/internal/handlers/core"
)

// safeGetEncryptedSetting safely retrieves an encrypted setting, returning empty string on error.
// This prevents JSON encoding errors when encrypted data is corrupted or cannot be decrypted.
func safeGetEncryptedSetting(h *core.Handler, key string) string {
	value, err := h.DB.GetEncryptedSetting(key)
	if err != nil {
		log.Printf("Warning: Failed to decrypt setting %s: %v. Returning empty string.", key, err)
		return ""
	}
	return sanitizeValue(value)
}

// safeGetSetting safely retrieves a setting, returning empty string on error.
func safeGetSetting(h *core.Handler, key string) string {
	value, err := h.DB.GetSetting(key)
	if err != nil {
		log.Printf("Warning: Failed to retrieve setting %s: %v. Returning empty string.", key, err)
		return ""
	}
	return sanitizeValue(value)
}

// sanitizeValue removes control characters that could break JSON encoding.
func sanitizeValue(value string) string {
	// Remove control characters that could break JSON
	return strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return -1 // Remove control characters except tab, newline, carriage return
		}
		return r
	}, value)
}

// HandleSettings handles GET and POST requests for application settings.
// CODE GENERATED - DO NOT EDIT MANUALLY
// To add new settings, edit internal/config/settings_schema.json and run: go run tools/settings-generator/main.go
func HandleSettings(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		aiApiKey := safeGetEncryptedSetting(h, "ai_api_key")
		aiChatEnabled := safeGetSetting(h, "ai_chat_enabled")
		aiCustomHeaders := safeGetSetting(h, "ai_custom_headers")
		aiEndpoint := safeGetSetting(h, "ai_endpoint")
		aiModel := safeGetSetting(h, "ai_model")
		aiSearchEnabled := safeGetSetting(h, "ai_search_enabled")
		aiSummaryPrompt := safeGetSetting(h, "ai_summary_prompt")
		aiTranslationPrompt := safeGetSetting(h, "ai_translation_prompt")
		aiUsageLimit := safeGetSetting(h, "ai_usage_limit")
		aiUsageTokens := safeGetSetting(h, "ai_usage_tokens")
		autoCleanupEnabled := safeGetSetting(h, "auto_cleanup_enabled")
		autoShowAllContent := safeGetSetting(h, "auto_show_all_content")
		baiduAppId := safeGetSetting(h, "baidu_app_id")
		baiduSecretKey := safeGetEncryptedSetting(h, "baidu_secret_key")
		closeToTray := safeGetSetting(h, "close_to_tray")
		contentFontFamily := safeGetSetting(h, "content_font_family")
		contentFontSize := safeGetSetting(h, "content_font_size")
		contentLineHeight := safeGetSetting(h, "content_line_height")
		customCssFile := safeGetSetting(h, "custom_css_file")
		customTranslationBodyTemplate := safeGetSetting(h, "custom_translation_body_template")
		customTranslationEnabled := safeGetSetting(h, "custom_translation_enabled")
		customTranslationEndpoint := safeGetSetting(h, "custom_translation_endpoint")
		customTranslationHeaders := safeGetSetting(h, "custom_translation_headers")
		customTranslationLangMapping := safeGetSetting(h, "custom_translation_lang_mapping")
		customTranslationMethod := safeGetSetting(h, "custom_translation_method")
		customTranslationName := safeGetSetting(h, "custom_translation_name")
		customTranslationResponsePath := safeGetSetting(h, "custom_translation_response_path")
		customTranslationTimeout := safeGetSetting(h, "custom_translation_timeout")
		deeplApiKey := safeGetEncryptedSetting(h, "deepl_api_key")
		deeplEndpoint := safeGetSetting(h, "deepl_endpoint")
		defaultViewMode := safeGetSetting(h, "default_view_mode")
		feedDrawerExpanded := safeGetSetting(h, "feed_drawer_expanded")
		feedDrawerPinned := safeGetSetting(h, "feed_drawer_pinned")
		freshrssApiPassword := safeGetEncryptedSetting(h, "freshrss_api_password")
		freshrssAutoSyncInterval := safeGetSetting(h, "freshrss_auto_sync_interval")
		freshrssEnabled := safeGetSetting(h, "freshrss_enabled")
		freshrssLastSyncTime := safeGetSetting(h, "freshrss_last_sync_time")
		freshrssServerUrl := safeGetSetting(h, "freshrss_server_url")
		freshrssSyncOnStartup := safeGetSetting(h, "freshrss_sync_on_startup")
		freshrssUsername := safeGetSetting(h, "freshrss_username")
		fullTextFetchEnabled := safeGetSetting(h, "full_text_fetch_enabled")
		googleTranslateEndpoint := safeGetSetting(h, "google_translate_endpoint")
		hoverMarkAsRead := safeGetSetting(h, "hover_mark_as_read")
		imageGalleryEnabled := safeGetSetting(h, "image_gallery_enabled")
		language := safeGetSetting(h, "language")
		lastGlobalRefresh := safeGetSetting(h, "last_global_refresh")
		lastNetworkTest := safeGetSetting(h, "last_network_test")
		layoutMode := safeGetSetting(h, "layout_mode")
		maxArticleAgeDays := safeGetSetting(h, "max_article_age_days")
		maxCacheSizeMb := safeGetSetting(h, "max_cache_size_mb")
		maxConcurrentRefreshes := safeGetSetting(h, "max_concurrent_refreshes")
		mediaCacheEnabled := safeGetSetting(h, "media_cache_enabled")
		mediaCacheMaxAgeDays := safeGetSetting(h, "media_cache_max_age_days")
		mediaCacheMaxSizeMb := safeGetSetting(h, "media_cache_max_size_mb")
		mediaProxyFallback := safeGetSetting(h, "media_proxy_fallback")
		networkBandwidthMbps := safeGetSetting(h, "network_bandwidth_mbps")
		networkLatencyMs := safeGetSetting(h, "network_latency_ms")
		networkSpeed := safeGetSetting(h, "network_speed")
		obsidianEnabled := safeGetSetting(h, "obsidian_enabled")
		obsidianVault := safeGetSetting(h, "obsidian_vault")
		obsidianVaultPath := safeGetSetting(h, "obsidian_vault_path")
		proxyEnabled := safeGetSetting(h, "proxy_enabled")
		proxyHost := safeGetSetting(h, "proxy_host")
		proxyPassword := safeGetEncryptedSetting(h, "proxy_password")
		proxyPort := safeGetSetting(h, "proxy_port")
		proxyType := safeGetSetting(h, "proxy_type")
		proxyUsername := safeGetEncryptedSetting(h, "proxy_username")
		refreshMode := safeGetSetting(h, "refresh_mode")
		retryTimeoutSeconds := safeGetSetting(h, "retry_timeout_seconds")
		rsshubApiKey := safeGetEncryptedSetting(h, "rsshub_api_key")
		rsshubEnabled := safeGetSetting(h, "rsshub_enabled")
		rsshubEndpoint := safeGetSetting(h, "rsshub_endpoint")
		rules := safeGetSetting(h, "rules")
		shortcuts := safeGetSetting(h, "shortcuts")
		shortcutsEnabled := safeGetSetting(h, "shortcuts_enabled")
		showArticlePreviewImages := safeGetSetting(h, "show_article_preview_images")
		showHiddenArticles := safeGetSetting(h, "show_hidden_articles")
		startupOnBoot := safeGetSetting(h, "startup_on_boot")
		summaryEnabled := safeGetSetting(h, "summary_enabled")
		summaryLength := safeGetSetting(h, "summary_length")
		summaryProvider := safeGetSetting(h, "summary_provider")
		summaryTriggerMode := safeGetSetting(h, "summary_trigger_mode")
		targetLanguage := safeGetSetting(h, "target_language")
		theme := safeGetSetting(h, "theme")
		translationEnabled := safeGetSetting(h, "translation_enabled")
		translationOnlyMode := safeGetSetting(h, "translation_only_mode")
		translationProvider := safeGetSetting(h, "translation_provider")
		updateInterval := safeGetSetting(h, "update_interval")
		windowHeight := safeGetSetting(h, "window_height")
		windowMaximized := safeGetSetting(h, "window_maximized")
		windowWidth := safeGetSetting(h, "window_width")
		windowX := safeGetSetting(h, "window_x")
		windowY := safeGetSetting(h, "window_y")
		json.NewEncoder(w).Encode(map[string]string{
			"ai_api_key":                       aiApiKey,
			"ai_chat_enabled":                  aiChatEnabled,
			"ai_custom_headers":                aiCustomHeaders,
			"ai_endpoint":                      aiEndpoint,
			"ai_model":                         aiModel,
			"ai_search_enabled":                aiSearchEnabled,
			"ai_summary_prompt":                aiSummaryPrompt,
			"ai_translation_prompt":            aiTranslationPrompt,
			"ai_usage_limit":                   aiUsageLimit,
			"ai_usage_tokens":                  aiUsageTokens,
			"auto_cleanup_enabled":             autoCleanupEnabled,
			"auto_show_all_content":            autoShowAllContent,
			"baidu_app_id":                     baiduAppId,
			"baidu_secret_key":                 baiduSecretKey,
			"close_to_tray":                    closeToTray,
			"content_font_family":              contentFontFamily,
			"content_font_size":                contentFontSize,
			"content_line_height":              contentLineHeight,
			"custom_css_file":                  customCssFile,
			"custom_translation_body_template": customTranslationBodyTemplate,
			"custom_translation_enabled":       customTranslationEnabled,
			"custom_translation_endpoint":      customTranslationEndpoint,
			"custom_translation_headers":       customTranslationHeaders,
			"custom_translation_lang_mapping":  customTranslationLangMapping,
			"custom_translation_method":        customTranslationMethod,
			"custom_translation_name":          customTranslationName,
			"custom_translation_response_path": customTranslationResponsePath,
			"custom_translation_timeout":       customTranslationTimeout,
			"deepl_api_key":                    deeplApiKey,
			"deepl_endpoint":                   deeplEndpoint,
			"default_view_mode":                defaultViewMode,
			"feed_drawer_expanded":             feedDrawerExpanded,
			"feed_drawer_pinned":               feedDrawerPinned,
			"freshrss_api_password":            freshrssApiPassword,
			"freshrss_auto_sync_interval":      freshrssAutoSyncInterval,
			"freshrss_enabled":                 freshrssEnabled,
			"freshrss_last_sync_time":          freshrssLastSyncTime,
			"freshrss_server_url":              freshrssServerUrl,
			"freshrss_sync_on_startup":         freshrssSyncOnStartup,
			"freshrss_username":                freshrssUsername,
			"full_text_fetch_enabled":          fullTextFetchEnabled,
			"google_translate_endpoint":        googleTranslateEndpoint,
			"hover_mark_as_read":               hoverMarkAsRead,
			"image_gallery_enabled":            imageGalleryEnabled,
			"language":                         language,
			"last_global_refresh":              lastGlobalRefresh,
			"last_network_test":                lastNetworkTest,
			"layout_mode":                      layoutMode,
			"max_article_age_days":             maxArticleAgeDays,
			"max_cache_size_mb":                maxCacheSizeMb,
			"max_concurrent_refreshes":         maxConcurrentRefreshes,
			"media_cache_enabled":              mediaCacheEnabled,
			"media_cache_max_age_days":         mediaCacheMaxAgeDays,
			"media_cache_max_size_mb":          mediaCacheMaxSizeMb,
			"media_proxy_fallback":             mediaProxyFallback,
			"network_bandwidth_mbps":           networkBandwidthMbps,
			"network_latency_ms":               networkLatencyMs,
			"network_speed":                    networkSpeed,
			"obsidian_enabled":                 obsidianEnabled,
			"obsidian_vault":                   obsidianVault,
			"obsidian_vault_path":              obsidianVaultPath,
			"proxy_enabled":                    proxyEnabled,
			"proxy_host":                       proxyHost,
			"proxy_password":                   proxyPassword,
			"proxy_port":                       proxyPort,
			"proxy_type":                       proxyType,
			"proxy_username":                   proxyUsername,
			"refresh_mode":                     refreshMode,
			"retry_timeout_seconds":            retryTimeoutSeconds,
			"rsshub_api_key":                   rsshubApiKey,
			"rsshub_enabled":                   rsshubEnabled,
			"rsshub_endpoint":                  rsshubEndpoint,
			"rules":                            rules,
			"shortcuts":                        shortcuts,
			"shortcuts_enabled":                shortcutsEnabled,
			"show_article_preview_images":      showArticlePreviewImages,
			"show_hidden_articles":             showHiddenArticles,
			"startup_on_boot":                  startupOnBoot,
			"summary_enabled":                  summaryEnabled,
			"summary_length":                   summaryLength,
			"summary_provider":                 summaryProvider,
			"summary_trigger_mode":             summaryTriggerMode,
			"target_language":                  targetLanguage,
			"theme":                            theme,
			"translation_enabled":              translationEnabled,
			"translation_only_mode":            translationOnlyMode,
			"translation_provider":             translationProvider,
			"update_interval":                  updateInterval,
			"window_height":                    windowHeight,
			"window_maximized":                 windowMaximized,
			"window_width":                     windowWidth,
			"window_x":                         windowX,
			"window_y":                         windowY,
		})
	case http.MethodPost:
		var req struct {
			AIAPIKey                      string `json:"ai_api_key"`
			AIChatEnabled                 string `json:"ai_chat_enabled"`
			AICustomHeaders               string `json:"ai_custom_headers"`
			AIEndpoint                    string `json:"ai_endpoint"`
			AIModel                       string `json:"ai_model"`
			AISearchEnabled               string `json:"ai_search_enabled"`
			AISummaryPrompt               string `json:"ai_summary_prompt"`
			AITranslationPrompt           string `json:"ai_translation_prompt"`
			AIUsageLimit                  string `json:"ai_usage_limit"`
			AIUsageTokens                 string `json:"ai_usage_tokens"`
			AutoCleanupEnabled            string `json:"auto_cleanup_enabled"`
			AutoShowAllContent            string `json:"auto_show_all_content"`
			BaiduAppId                    string `json:"baidu_app_id"`
			BaiduSecretKey                string `json:"baidu_secret_key"`
			CloseToTray                   string `json:"close_to_tray"`
			ContentFontFamily             string `json:"content_font_family"`
			ContentFontSize               string `json:"content_font_size"`
			ContentLineHeight             string `json:"content_line_height"`
			CustomCssFile                 string `json:"custom_css_file"`
			CustomTranslationBodyTemplate string `json:"custom_translation_body_template"`
			CustomTranslationEnabled      string `json:"custom_translation_enabled"`
			CustomTranslationEndpoint     string `json:"custom_translation_endpoint"`
			CustomTranslationHeaders      string `json:"custom_translation_headers"`
			CustomTranslationLangMapping  string `json:"custom_translation_lang_mapping"`
			CustomTranslationMethod       string `json:"custom_translation_method"`
			CustomTranslationName         string `json:"custom_translation_name"`
			CustomTranslationResponsePath string `json:"custom_translation_response_path"`
			CustomTranslationTimeout      string `json:"custom_translation_timeout"`
			DeeplAPIKey                   string `json:"deepl_api_key"`
			DeeplEndpoint                 string `json:"deepl_endpoint"`
			DefaultViewMode               string `json:"default_view_mode"`
			FeedDrawerExpanded            string `json:"feed_drawer_expanded"`
			FeedDrawerPinned              string `json:"feed_drawer_pinned"`
			FreshRSSAPIPassword           string `json:"freshrss_api_password"`
			FreshRSSAutoSyncInterval      string `json:"freshrss_auto_sync_interval"`
			FreshRSSEnabled               string `json:"freshrss_enabled"`
			FreshRSSLastSyncTime          string `json:"freshrss_last_sync_time"`
			FreshRSSServerUrl             string `json:"freshrss_server_url"`
			FreshRSSSyncOnStartup         string `json:"freshrss_sync_on_startup"`
			FreshRSSUsername              string `json:"freshrss_username"`
			FullTextFetchEnabled          string `json:"full_text_fetch_enabled"`
			GoogleTranslateEndpoint       string `json:"google_translate_endpoint"`
			HoverMarkAsRead               string `json:"hover_mark_as_read"`
			ImageGalleryEnabled           string `json:"image_gallery_enabled"`
			Language                      string `json:"language"`
			LastGlobalRefresh             string `json:"last_global_refresh"`
			LastNetworkTest               string `json:"last_network_test"`
			LayoutMode                    string `json:"layout_mode"`
			MaxArticleAgeDays             string `json:"max_article_age_days"`
			MaxCacheSizeMb                string `json:"max_cache_size_mb"`
			MaxConcurrentRefreshes        string `json:"max_concurrent_refreshes"`
			MediaCacheEnabled             string `json:"media_cache_enabled"`
			MediaCacheMaxAgeDays          string `json:"media_cache_max_age_days"`
			MediaCacheMaxSizeMb           string `json:"media_cache_max_size_mb"`
			MediaProxyFallback            string `json:"media_proxy_fallback"`
			NetworkBandwidthMbps          string `json:"network_bandwidth_mbps"`
			NetworkLatencyMs              string `json:"network_latency_ms"`
			NetworkSpeed                  string `json:"network_speed"`
			ObsidianEnabled               string `json:"obsidian_enabled"`
			ObsidianVault                 string `json:"obsidian_vault"`
			ObsidianVaultPath             string `json:"obsidian_vault_path"`
			ProxyEnabled                  string `json:"proxy_enabled"`
			ProxyHost                     string `json:"proxy_host"`
			ProxyPassword                 string `json:"proxy_password"`
			ProxyPort                     string `json:"proxy_port"`
			ProxyType                     string `json:"proxy_type"`
			ProxyUsername                 string `json:"proxy_username"`
			RefreshMode                   string `json:"refresh_mode"`
			RetryTimeoutSeconds           string `json:"retry_timeout_seconds"`
			RsshubAPIKey                  string `json:"rsshub_api_key"`
			RsshubEnabled                 string `json:"rsshub_enabled"`
			RsshubEndpoint                string `json:"rsshub_endpoint"`
			Rules                         string `json:"rules"`
			Shortcuts                     string `json:"shortcuts"`
			ShortcutsEnabled              string `json:"shortcuts_enabled"`
			ShowArticlePreviewImages      string `json:"show_article_preview_images"`
			ShowHiddenArticles            string `json:"show_hidden_articles"`
			StartupOnBoot                 string `json:"startup_on_boot"`
			SummaryEnabled                string `json:"summary_enabled"`
			SummaryLength                 string `json:"summary_length"`
			SummaryProvider               string `json:"summary_provider"`
			SummaryTriggerMode            string `json:"summary_trigger_mode"`
			TargetLanguage                string `json:"target_language"`
			Theme                         string `json:"theme"`
			TranslationEnabled            string `json:"translation_enabled"`
			TranslationOnlyMode           string `json:"translation_only_mode"`
			TranslationProvider           string `json:"translation_provider"`
			UpdateInterval                string `json:"update_interval"`
			WindowHeight                  string `json:"window_height"`
			WindowMaximized               string `json:"window_maximized"`
			WindowWidth                   string `json:"window_width"`
			WindowX                       string `json:"window_x"`
			WindowY                       string `json:"window_y"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := h.DB.SetEncryptedSetting("ai_api_key", req.AIAPIKey); err != nil {
			log.Printf("Failed to save ai_api_key: %v", err)
			http.Error(w, "Failed to save ai_api_key", http.StatusInternalServerError)
			return
		}

		if req.AIChatEnabled != "" {
			h.DB.SetSetting("ai_chat_enabled", req.AIChatEnabled)
		}

		if req.AICustomHeaders != "" {
			h.DB.SetSetting("ai_custom_headers", req.AICustomHeaders)
		}

		if req.AIEndpoint != "" {
			h.DB.SetSetting("ai_endpoint", req.AIEndpoint)
		}

		if req.AIModel != "" {
			h.DB.SetSetting("ai_model", req.AIModel)
		}

		if req.AISearchEnabled != "" {
			h.DB.SetSetting("ai_search_enabled", req.AISearchEnabled)
		}

		if req.AISummaryPrompt != "" {
			h.DB.SetSetting("ai_summary_prompt", req.AISummaryPrompt)
		}

		if req.AITranslationPrompt != "" {
			h.DB.SetSetting("ai_translation_prompt", req.AITranslationPrompt)
		}

		if req.AIUsageLimit != "" {
			h.DB.SetSetting("ai_usage_limit", req.AIUsageLimit)
		}

		if req.AIUsageTokens != "" {
			h.DB.SetSetting("ai_usage_tokens", req.AIUsageTokens)
		}

		if req.AutoCleanupEnabled != "" {
			h.DB.SetSetting("auto_cleanup_enabled", req.AutoCleanupEnabled)
		}

		if req.AutoShowAllContent != "" {
			h.DB.SetSetting("auto_show_all_content", req.AutoShowAllContent)
		}

		if req.BaiduAppId != "" {
			h.DB.SetSetting("baidu_app_id", req.BaiduAppId)
		}

		if err := h.DB.SetEncryptedSetting("baidu_secret_key", req.BaiduSecretKey); err != nil {
			log.Printf("Failed to save baidu_secret_key: %v", err)
			http.Error(w, "Failed to save baidu_secret_key", http.StatusInternalServerError)
			return
		}

		if req.CloseToTray != "" {
			h.DB.SetSetting("close_to_tray", req.CloseToTray)
		}

		if req.ContentFontFamily != "" {
			h.DB.SetSetting("content_font_family", req.ContentFontFamily)
		}

		if req.ContentFontSize != "" {
			h.DB.SetSetting("content_font_size", req.ContentFontSize)
		}

		if req.ContentLineHeight != "" {
			h.DB.SetSetting("content_line_height", req.ContentLineHeight)
		}

		if req.CustomCssFile != "" {
			h.DB.SetSetting("custom_css_file", req.CustomCssFile)
		}

		if req.CustomTranslationBodyTemplate != "" {
			h.DB.SetSetting("custom_translation_body_template", req.CustomTranslationBodyTemplate)
		}

		if req.CustomTranslationEnabled != "" {
			h.DB.SetSetting("custom_translation_enabled", req.CustomTranslationEnabled)
		}

		if req.CustomTranslationEndpoint != "" {
			h.DB.SetSetting("custom_translation_endpoint", req.CustomTranslationEndpoint)
		}

		if req.CustomTranslationHeaders != "" {
			h.DB.SetSetting("custom_translation_headers", req.CustomTranslationHeaders)
		}

		if req.CustomTranslationLangMapping != "" {
			h.DB.SetSetting("custom_translation_lang_mapping", req.CustomTranslationLangMapping)
		}

		if req.CustomTranslationMethod != "" {
			h.DB.SetSetting("custom_translation_method", req.CustomTranslationMethod)
		}

		if req.CustomTranslationName != "" {
			h.DB.SetSetting("custom_translation_name", req.CustomTranslationName)
		}

		if req.CustomTranslationResponsePath != "" {
			h.DB.SetSetting("custom_translation_response_path", req.CustomTranslationResponsePath)
		}

		if req.CustomTranslationTimeout != "" {
			h.DB.SetSetting("custom_translation_timeout", req.CustomTranslationTimeout)
		}

		if err := h.DB.SetEncryptedSetting("deepl_api_key", req.DeeplAPIKey); err != nil {
			log.Printf("Failed to save deepl_api_key: %v", err)
			http.Error(w, "Failed to save deepl_api_key", http.StatusInternalServerError)
			return
		}

		if req.DeeplEndpoint != "" {
			h.DB.SetSetting("deepl_endpoint", req.DeeplEndpoint)
		}

		if req.DefaultViewMode != "" {
			h.DB.SetSetting("default_view_mode", req.DefaultViewMode)
		}

		if req.FeedDrawerExpanded != "" {
			h.DB.SetSetting("feed_drawer_expanded", req.FeedDrawerExpanded)
		}

		if req.FeedDrawerPinned != "" {
			h.DB.SetSetting("feed_drawer_pinned", req.FeedDrawerPinned)
		}

		if err := h.DB.SetEncryptedSetting("freshrss_api_password", req.FreshRSSAPIPassword); err != nil {
			log.Printf("Failed to save freshrss_api_password: %v", err)
			http.Error(w, "Failed to save freshrss_api_password", http.StatusInternalServerError)
			return
		}

		if req.FreshRSSAutoSyncInterval != "" {
			h.DB.SetSetting("freshrss_auto_sync_interval", req.FreshRSSAutoSyncInterval)
		}

		if req.FreshRSSEnabled != "" {
			h.DB.SetSetting("freshrss_enabled", req.FreshRSSEnabled)
		}

		if req.FreshRSSLastSyncTime != "" {
			h.DB.SetSetting("freshrss_last_sync_time", req.FreshRSSLastSyncTime)
		}

		if req.FreshRSSServerUrl != "" {
			h.DB.SetSetting("freshrss_server_url", req.FreshRSSServerUrl)
		}

		if req.FreshRSSSyncOnStartup != "" {
			h.DB.SetSetting("freshrss_sync_on_startup", req.FreshRSSSyncOnStartup)
		}

		if req.FreshRSSUsername != "" {
			h.DB.SetSetting("freshrss_username", req.FreshRSSUsername)
		}

		if req.FullTextFetchEnabled != "" {
			h.DB.SetSetting("full_text_fetch_enabled", req.FullTextFetchEnabled)
		}

		if req.GoogleTranslateEndpoint != "" {
			h.DB.SetSetting("google_translate_endpoint", req.GoogleTranslateEndpoint)
		}

		if req.HoverMarkAsRead != "" {
			h.DB.SetSetting("hover_mark_as_read", req.HoverMarkAsRead)
		}

		if req.ImageGalleryEnabled != "" {
			h.DB.SetSetting("image_gallery_enabled", req.ImageGalleryEnabled)
		}

		if req.Language != "" {
			h.DB.SetSetting("language", req.Language)
		}

		if req.LastGlobalRefresh != "" {
			h.DB.SetSetting("last_global_refresh", req.LastGlobalRefresh)
		}

		if req.LastNetworkTest != "" {
			h.DB.SetSetting("last_network_test", req.LastNetworkTest)
		}

		if req.LayoutMode != "" {
			h.DB.SetSetting("layout_mode", req.LayoutMode)
		}

		if req.MaxArticleAgeDays != "" {
			h.DB.SetSetting("max_article_age_days", req.MaxArticleAgeDays)
		}

		if req.MaxCacheSizeMb != "" {
			h.DB.SetSetting("max_cache_size_mb", req.MaxCacheSizeMb)
		}

		if req.MaxConcurrentRefreshes != "" {
			h.DB.SetSetting("max_concurrent_refreshes", req.MaxConcurrentRefreshes)
		}

		if req.MediaCacheEnabled != "" {
			h.DB.SetSetting("media_cache_enabled", req.MediaCacheEnabled)
		}

		if req.MediaCacheMaxAgeDays != "" {
			h.DB.SetSetting("media_cache_max_age_days", req.MediaCacheMaxAgeDays)
		}

		if req.MediaCacheMaxSizeMb != "" {
			h.DB.SetSetting("media_cache_max_size_mb", req.MediaCacheMaxSizeMb)
		}

		if req.MediaProxyFallback != "" {
			h.DB.SetSetting("media_proxy_fallback", req.MediaProxyFallback)
		}

		if req.NetworkBandwidthMbps != "" {
			h.DB.SetSetting("network_bandwidth_mbps", req.NetworkBandwidthMbps)
		}

		if req.NetworkLatencyMs != "" {
			h.DB.SetSetting("network_latency_ms", req.NetworkLatencyMs)
		}

		if req.NetworkSpeed != "" {
			h.DB.SetSetting("network_speed", req.NetworkSpeed)
		}

		if req.ObsidianEnabled != "" {
			h.DB.SetSetting("obsidian_enabled", req.ObsidianEnabled)
		}

		if req.ObsidianVault != "" {
			h.DB.SetSetting("obsidian_vault", req.ObsidianVault)
		}

		if req.ObsidianVaultPath != "" {
			h.DB.SetSetting("obsidian_vault_path", req.ObsidianVaultPath)
		}

		if req.ProxyEnabled != "" {
			h.DB.SetSetting("proxy_enabled", req.ProxyEnabled)
		}

		if req.ProxyHost != "" {
			h.DB.SetSetting("proxy_host", req.ProxyHost)
		}

		if err := h.DB.SetEncryptedSetting("proxy_password", req.ProxyPassword); err != nil {
			log.Printf("Failed to save proxy_password: %v", err)
			http.Error(w, "Failed to save proxy_password", http.StatusInternalServerError)
			return
		}

		if req.ProxyPort != "" {
			h.DB.SetSetting("proxy_port", req.ProxyPort)
		}

		if req.ProxyType != "" {
			h.DB.SetSetting("proxy_type", req.ProxyType)
		}

		if err := h.DB.SetEncryptedSetting("proxy_username", req.ProxyUsername); err != nil {
			log.Printf("Failed to save proxy_username: %v", err)
			http.Error(w, "Failed to save proxy_username", http.StatusInternalServerError)
			return
		}

		if req.RefreshMode != "" {
			h.DB.SetSetting("refresh_mode", req.RefreshMode)
		}

		if req.RetryTimeoutSeconds != "" {
			h.DB.SetSetting("retry_timeout_seconds", req.RetryTimeoutSeconds)
		}

		if err := h.DB.SetEncryptedSetting("rsshub_api_key", req.RsshubAPIKey); err != nil {
			log.Printf("Failed to save rsshub_api_key: %v", err)
			http.Error(w, "Failed to save rsshub_api_key", http.StatusInternalServerError)
			return
		}

		if req.RsshubEnabled != "" {
			h.DB.SetSetting("rsshub_enabled", req.RsshubEnabled)
		}

		if req.RsshubEndpoint != "" {
			h.DB.SetSetting("rsshub_endpoint", req.RsshubEndpoint)
		}

		if req.Rules != "" {
			h.DB.SetSetting("rules", req.Rules)
		}

		if req.Shortcuts != "" {
			h.DB.SetSetting("shortcuts", req.Shortcuts)
		}

		if req.ShortcutsEnabled != "" {
			h.DB.SetSetting("shortcuts_enabled", req.ShortcutsEnabled)
		}

		if req.ShowArticlePreviewImages != "" {
			h.DB.SetSetting("show_article_preview_images", req.ShowArticlePreviewImages)
		}

		if req.ShowHiddenArticles != "" {
			h.DB.SetSetting("show_hidden_articles", req.ShowHiddenArticles)
		}

		if req.StartupOnBoot != "" {
			h.DB.SetSetting("startup_on_boot", req.StartupOnBoot)
		}

		if req.SummaryEnabled != "" {
			h.DB.SetSetting("summary_enabled", req.SummaryEnabled)
		}

		if req.SummaryLength != "" {
			h.DB.SetSetting("summary_length", req.SummaryLength)
		}

		if req.SummaryProvider != "" {
			h.DB.SetSetting("summary_provider", req.SummaryProvider)
		}

		if req.SummaryTriggerMode != "" {
			h.DB.SetSetting("summary_trigger_mode", req.SummaryTriggerMode)
		}

		if req.TargetLanguage != "" {
			h.DB.SetSetting("target_language", req.TargetLanguage)
		}

		if req.Theme != "" {
			h.DB.SetSetting("theme", req.Theme)
		}

		if req.TranslationEnabled != "" {
			h.DB.SetSetting("translation_enabled", req.TranslationEnabled)
		}

		if req.TranslationOnlyMode != "" {
			h.DB.SetSetting("translation_only_mode", req.TranslationOnlyMode)
		}

		if req.TranslationProvider != "" {
			h.DB.SetSetting("translation_provider", req.TranslationProvider)
		}

		if req.UpdateInterval != "" {
			h.DB.SetSetting("update_interval", req.UpdateInterval)
		}

		if req.WindowHeight != "" {
			h.DB.SetSetting("window_height", req.WindowHeight)
		}

		if req.WindowMaximized != "" {
			h.DB.SetSetting("window_maximized", req.WindowMaximized)
		}

		if req.WindowWidth != "" {
			h.DB.SetSetting("window_width", req.WindowWidth)
		}

		if req.WindowX != "" {
			h.DB.SetSetting("window_x", req.WindowX)
		}

		if req.WindowY != "" {
			h.DB.SetSetting("window_y", req.WindowY)
		}
		// Re-fetch all settings after save to return updated values
		aiApiKey := safeGetEncryptedSetting(h, "ai_api_key")
		aiChatEnabled := safeGetSetting(h, "ai_chat_enabled")
		aiCustomHeaders := safeGetSetting(h, "ai_custom_headers")
		aiEndpoint := safeGetSetting(h, "ai_endpoint")
		aiModel := safeGetSetting(h, "ai_model")
		aiSearchEnabled := safeGetSetting(h, "ai_search_enabled")
		aiSummaryPrompt := safeGetSetting(h, "ai_summary_prompt")
		aiTranslationPrompt := safeGetSetting(h, "ai_translation_prompt")
		aiUsageLimit := safeGetSetting(h, "ai_usage_limit")
		aiUsageTokens := safeGetSetting(h, "ai_usage_tokens")
		autoCleanupEnabled := safeGetSetting(h, "auto_cleanup_enabled")
		autoShowAllContent := safeGetSetting(h, "auto_show_all_content")
		baiduAppId := safeGetSetting(h, "baidu_app_id")
		baiduSecretKey := safeGetEncryptedSetting(h, "baidu_secret_key")
		closeToTray := safeGetSetting(h, "close_to_tray")
		contentFontFamily := safeGetSetting(h, "content_font_family")
		contentFontSize := safeGetSetting(h, "content_font_size")
		contentLineHeight := safeGetSetting(h, "content_line_height")
		customCssFile := safeGetSetting(h, "custom_css_file")
		customTranslationBodyTemplate := safeGetSetting(h, "custom_translation_body_template")
		customTranslationEnabled := safeGetSetting(h, "custom_translation_enabled")
		customTranslationEndpoint := safeGetSetting(h, "custom_translation_endpoint")
		customTranslationHeaders := safeGetSetting(h, "custom_translation_headers")
		customTranslationLangMapping := safeGetSetting(h, "custom_translation_lang_mapping")
		customTranslationMethod := safeGetSetting(h, "custom_translation_method")
		customTranslationName := safeGetSetting(h, "custom_translation_name")
		customTranslationResponsePath := safeGetSetting(h, "custom_translation_response_path")
		customTranslationTimeout := safeGetSetting(h, "custom_translation_timeout")
		deeplApiKey := safeGetEncryptedSetting(h, "deepl_api_key")
		deeplEndpoint := safeGetSetting(h, "deepl_endpoint")
		defaultViewMode := safeGetSetting(h, "default_view_mode")
		feedDrawerExpanded := safeGetSetting(h, "feed_drawer_expanded")
		feedDrawerPinned := safeGetSetting(h, "feed_drawer_pinned")
		freshrssApiPassword := safeGetEncryptedSetting(h, "freshrss_api_password")
		freshrssAutoSyncInterval := safeGetSetting(h, "freshrss_auto_sync_interval")
		freshrssEnabled := safeGetSetting(h, "freshrss_enabled")
		freshrssLastSyncTime := safeGetSetting(h, "freshrss_last_sync_time")
		freshrssServerUrl := safeGetSetting(h, "freshrss_server_url")
		freshrssSyncOnStartup := safeGetSetting(h, "freshrss_sync_on_startup")
		freshrssUsername := safeGetSetting(h, "freshrss_username")
		fullTextFetchEnabled := safeGetSetting(h, "full_text_fetch_enabled")
		googleTranslateEndpoint := safeGetSetting(h, "google_translate_endpoint")
		hoverMarkAsRead := safeGetSetting(h, "hover_mark_as_read")
		imageGalleryEnabled := safeGetSetting(h, "image_gallery_enabled")
		language := safeGetSetting(h, "language")
		lastGlobalRefresh := safeGetSetting(h, "last_global_refresh")
		lastNetworkTest := safeGetSetting(h, "last_network_test")
		layoutMode := safeGetSetting(h, "layout_mode")
		maxArticleAgeDays := safeGetSetting(h, "max_article_age_days")
		maxCacheSizeMb := safeGetSetting(h, "max_cache_size_mb")
		maxConcurrentRefreshes := safeGetSetting(h, "max_concurrent_refreshes")
		mediaCacheEnabled := safeGetSetting(h, "media_cache_enabled")
		mediaCacheMaxAgeDays := safeGetSetting(h, "media_cache_max_age_days")
		mediaCacheMaxSizeMb := safeGetSetting(h, "media_cache_max_size_mb")
		mediaProxyFallback := safeGetSetting(h, "media_proxy_fallback")
		networkBandwidthMbps := safeGetSetting(h, "network_bandwidth_mbps")
		networkLatencyMs := safeGetSetting(h, "network_latency_ms")
		networkSpeed := safeGetSetting(h, "network_speed")
		obsidianEnabled := safeGetSetting(h, "obsidian_enabled")
		obsidianVault := safeGetSetting(h, "obsidian_vault")
		obsidianVaultPath := safeGetSetting(h, "obsidian_vault_path")
		proxyEnabled := safeGetSetting(h, "proxy_enabled")
		proxyHost := safeGetSetting(h, "proxy_host")
		proxyPassword := safeGetEncryptedSetting(h, "proxy_password")
		proxyPort := safeGetSetting(h, "proxy_port")
		proxyType := safeGetSetting(h, "proxy_type")
		proxyUsername := safeGetEncryptedSetting(h, "proxy_username")
		refreshMode := safeGetSetting(h, "refresh_mode")
		retryTimeoutSeconds := safeGetSetting(h, "retry_timeout_seconds")
		rsshubApiKey := safeGetEncryptedSetting(h, "rsshub_api_key")
		rsshubEnabled := safeGetSetting(h, "rsshub_enabled")
		rsshubEndpoint := safeGetSetting(h, "rsshub_endpoint")
		rules := safeGetSetting(h, "rules")
		shortcuts := safeGetSetting(h, "shortcuts")
		shortcutsEnabled := safeGetSetting(h, "shortcuts_enabled")
		showArticlePreviewImages := safeGetSetting(h, "show_article_preview_images")
		showHiddenArticles := safeGetSetting(h, "show_hidden_articles")
		startupOnBoot := safeGetSetting(h, "startup_on_boot")
		summaryEnabled := safeGetSetting(h, "summary_enabled")
		summaryLength := safeGetSetting(h, "summary_length")
		summaryProvider := safeGetSetting(h, "summary_provider")
		summaryTriggerMode := safeGetSetting(h, "summary_trigger_mode")
		targetLanguage := safeGetSetting(h, "target_language")
		theme := safeGetSetting(h, "theme")
		translationEnabled := safeGetSetting(h, "translation_enabled")
		translationOnlyMode := safeGetSetting(h, "translation_only_mode")
		translationProvider := safeGetSetting(h, "translation_provider")
		updateInterval := safeGetSetting(h, "update_interval")
		windowHeight := safeGetSetting(h, "window_height")
		windowMaximized := safeGetSetting(h, "window_maximized")
		windowWidth := safeGetSetting(h, "window_width")
		windowX := safeGetSetting(h, "window_x")
		windowY := safeGetSetting(h, "window_y")
		json.NewEncoder(w).Encode(map[string]string{
			"ai_api_key":                       aiApiKey,
			"ai_chat_enabled":                  aiChatEnabled,
			"ai_custom_headers":                aiCustomHeaders,
			"ai_endpoint":                      aiEndpoint,
			"ai_model":                         aiModel,
			"ai_search_enabled":                aiSearchEnabled,
			"ai_summary_prompt":                aiSummaryPrompt,
			"ai_translation_prompt":            aiTranslationPrompt,
			"ai_usage_limit":                   aiUsageLimit,
			"ai_usage_tokens":                  aiUsageTokens,
			"auto_cleanup_enabled":             autoCleanupEnabled,
			"auto_show_all_content":            autoShowAllContent,
			"baidu_app_id":                     baiduAppId,
			"baidu_secret_key":                 baiduSecretKey,
			"close_to_tray":                    closeToTray,
			"content_font_family":              contentFontFamily,
			"content_font_size":                contentFontSize,
			"content_line_height":              contentLineHeight,
			"custom_css_file":                  customCssFile,
			"custom_translation_body_template": customTranslationBodyTemplate,
			"custom_translation_enabled":       customTranslationEnabled,
			"custom_translation_endpoint":      customTranslationEndpoint,
			"custom_translation_headers":       customTranslationHeaders,
			"custom_translation_lang_mapping":  customTranslationLangMapping,
			"custom_translation_method":        customTranslationMethod,
			"custom_translation_name":          customTranslationName,
			"custom_translation_response_path": customTranslationResponsePath,
			"custom_translation_timeout":       customTranslationTimeout,
			"deepl_api_key":                    deeplApiKey,
			"deepl_endpoint":                   deeplEndpoint,
			"default_view_mode":                defaultViewMode,
			"feed_drawer_expanded":             feedDrawerExpanded,
			"feed_drawer_pinned":               feedDrawerPinned,
			"freshrss_api_password":            freshrssApiPassword,
			"freshrss_auto_sync_interval":      freshrssAutoSyncInterval,
			"freshrss_enabled":                 freshrssEnabled,
			"freshrss_last_sync_time":          freshrssLastSyncTime,
			"freshrss_server_url":              freshrssServerUrl,
			"freshrss_sync_on_startup":         freshrssSyncOnStartup,
			"freshrss_username":                freshrssUsername,
			"full_text_fetch_enabled":          fullTextFetchEnabled,
			"google_translate_endpoint":        googleTranslateEndpoint,
			"hover_mark_as_read":               hoverMarkAsRead,
			"image_gallery_enabled":            imageGalleryEnabled,
			"language":                         language,
			"last_global_refresh":              lastGlobalRefresh,
			"last_network_test":                lastNetworkTest,
			"layout_mode":                      layoutMode,
			"max_article_age_days":             maxArticleAgeDays,
			"max_cache_size_mb":                maxCacheSizeMb,
			"max_concurrent_refreshes":         maxConcurrentRefreshes,
			"media_cache_enabled":              mediaCacheEnabled,
			"media_cache_max_age_days":         mediaCacheMaxAgeDays,
			"media_cache_max_size_mb":          mediaCacheMaxSizeMb,
			"media_proxy_fallback":             mediaProxyFallback,
			"network_bandwidth_mbps":           networkBandwidthMbps,
			"network_latency_ms":               networkLatencyMs,
			"network_speed":                    networkSpeed,
			"obsidian_enabled":                 obsidianEnabled,
			"obsidian_vault":                   obsidianVault,
			"obsidian_vault_path":              obsidianVaultPath,
			"proxy_enabled":                    proxyEnabled,
			"proxy_host":                       proxyHost,
			"proxy_password":                   proxyPassword,
			"proxy_port":                       proxyPort,
			"proxy_type":                       proxyType,
			"proxy_username":                   proxyUsername,
			"refresh_mode":                     refreshMode,
			"retry_timeout_seconds":            retryTimeoutSeconds,
			"rsshub_api_key":                   rsshubApiKey,
			"rsshub_enabled":                   rsshubEnabled,
			"rsshub_endpoint":                  rsshubEndpoint,
			"rules":                            rules,
			"shortcuts":                        shortcuts,
			"shortcuts_enabled":                shortcutsEnabled,
			"show_article_preview_images":      showArticlePreviewImages,
			"show_hidden_articles":             showHiddenArticles,
			"startup_on_boot":                  startupOnBoot,
			"summary_enabled":                  summaryEnabled,
			"summary_length":                   summaryLength,
			"summary_provider":                 summaryProvider,
			"summary_trigger_mode":             summaryTriggerMode,
			"target_language":                  targetLanguage,
			"theme":                            theme,
			"translation_enabled":              translationEnabled,
			"translation_only_mode":            translationOnlyMode,
			"translation_provider":             translationProvider,
			"update_interval":                  updateInterval,
			"window_height":                    windowHeight,
			"window_maximized":                 windowMaximized,
			"window_width":                     windowWidth,
			"window_x":                         windowX,
			"window_y":                         windowY,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
