// Copyright 2026 Ch3nyang & MrRSS Team. All rights reserved.
//
// Package config provides centralized default values for settings.
// The defaults are loaded from config/defaults.json which is shared between
// frontend and backend to ensure consistency.
// CODE GENERATED - DO NOT EDIT MANUALLY
// To add new settings, edit internal/config/settings_schema.json and run: go run tools/settings-generator/main.go
package config

import (
	_ "embed"
	"encoding/json"
	"strconv"
)

//go:embed defaults.json
var defaultsJSON []byte

// Defaults holds all default settings values
type Defaults struct {
	AIAPIKey                      string `json:"ai_api_key"`
	AIChatEnabled                 bool   `json:"ai_chat_enabled"`
	AICustomHeaders               string `json:"ai_custom_headers"`
	AIEndpoint                    string `json:"ai_endpoint"`
	AIModel                       string `json:"ai_model"`
	AISearchEnabled               bool   `json:"ai_search_enabled"`
	AISummaryPrompt               string `json:"ai_summary_prompt"`
	AITranslationPrompt           string `json:"ai_translation_prompt"`
	AIUsageLimit                  string `json:"ai_usage_limit"`
	AIUsageTokens                 string `json:"ai_usage_tokens"`
	AutoCleanupEnabled            bool   `json:"auto_cleanup_enabled"`
	AutoShowAllContent            bool   `json:"auto_show_all_content"`
	BaiduAppId                    string `json:"baidu_app_id"`
	BaiduSecretKey                string `json:"baidu_secret_key"`
	CloseToTray                   bool   `json:"close_to_tray"`
	ContentFontFamily             string `json:"content_font_family"`
	ContentFontSize               int    `json:"content_font_size"`
	ContentLineHeight             string `json:"content_line_height"`
	CustomCssFile                 string `json:"custom_css_file"`
	CustomTranslationBodyTemplate string `json:"custom_translation_body_template"`
	CustomTranslationEnabled      bool   `json:"custom_translation_enabled"`
	CustomTranslationEndpoint     string `json:"custom_translation_endpoint"`
	CustomTranslationHeaders      string `json:"custom_translation_headers"`
	CustomTranslationLangMapping  string `json:"custom_translation_lang_mapping"`
	CustomTranslationMethod       string `json:"custom_translation_method"`
	CustomTranslationName         string `json:"custom_translation_name"`
	CustomTranslationResponsePath string `json:"custom_translation_response_path"`
	CustomTranslationTimeout      int    `json:"custom_translation_timeout"`
	DeeplAPIKey                   string `json:"deepl_api_key"`
	DeeplEndpoint                 string `json:"deepl_endpoint"`
	DefaultViewMode               string `json:"default_view_mode"`
	FeedDrawerExpanded            bool   `json:"feed_drawer_expanded"`
	FeedDrawerPinned              bool   `json:"feed_drawer_pinned"`
	FreshRSSAPIPassword           string `json:"freshrss_api_password"`
	FreshRSSAutoSyncInterval      int    `json:"freshrss_auto_sync_interval"`
	FreshRSSEnabled               bool   `json:"freshrss_enabled"`
	FreshRSSLastSyncTime          string `json:"freshrss_last_sync_time"`
	FreshRSSServerUrl             string `json:"freshrss_server_url"`
	FreshRSSSyncOnStartup         bool   `json:"freshrss_sync_on_startup"`
	FreshRSSUsername              string `json:"freshrss_username"`
	FullTextFetchEnabled          bool   `json:"full_text_fetch_enabled"`
	GoogleTranslateEndpoint       string `json:"google_translate_endpoint"`
	HoverMarkAsRead               bool   `json:"hover_mark_as_read"`
	ImageGalleryEnabled           bool   `json:"image_gallery_enabled"`
	Language                      string `json:"language"`
	LastGlobalRefresh             string `json:"last_global_refresh"`
	LastNetworkTest               string `json:"last_network_test"`
	LayoutMode                    string `json:"layout_mode"`
	MaxArticleAgeDays             int    `json:"max_article_age_days"`
	MaxCacheSizeMb                int    `json:"max_cache_size_mb"`
	MaxConcurrentRefreshes        string `json:"max_concurrent_refreshes"`
	MediaCacheEnabled             bool   `json:"media_cache_enabled"`
	MediaCacheMaxAgeDays          int    `json:"media_cache_max_age_days"`
	MediaCacheMaxSizeMb           int    `json:"media_cache_max_size_mb"`
	MediaProxyFallback            bool   `json:"media_proxy_fallback"`
	NetworkBandwidthMbps          string `json:"network_bandwidth_mbps"`
	NetworkLatencyMs              string `json:"network_latency_ms"`
	NetworkSpeed                  string `json:"network_speed"`
	ObsidianEnabled               bool   `json:"obsidian_enabled"`
	ObsidianVault                 string `json:"obsidian_vault"`
	ObsidianVaultPath             string `json:"obsidian_vault_path"`
	ProxyEnabled                  bool   `json:"proxy_enabled"`
	ProxyHost                     string `json:"proxy_host"`
	ProxyPassword                 string `json:"proxy_password"`
	ProxyPort                     string `json:"proxy_port"`
	ProxyType                     string `json:"proxy_type"`
	ProxyUsername                 string `json:"proxy_username"`
	RefreshMode                   string `json:"refresh_mode"`
	RetryTimeoutSeconds           int    `json:"retry_timeout_seconds"`
	RsshubAPIKey                  string `json:"rsshub_api_key"`
	RsshubEnabled                 bool   `json:"rsshub_enabled"`
	RsshubEndpoint                string `json:"rsshub_endpoint"`
	Rules                         string `json:"rules"`
	Shortcuts                     string `json:"shortcuts"`
	ShortcutsEnabled              bool   `json:"shortcuts_enabled"`
	ShowArticlePreviewImages      bool   `json:"show_article_preview_images"`
	ShowHiddenArticles            bool   `json:"show_hidden_articles"`
	StartupOnBoot                 bool   `json:"startup_on_boot"`
	SummaryEnabled                bool   `json:"summary_enabled"`
	SummaryLength                 string `json:"summary_length"`
	SummaryProvider               string `json:"summary_provider"`
	SummaryTriggerMode            string `json:"summary_trigger_mode"`
	TargetLanguage                string `json:"target_language"`
	Theme                         string `json:"theme"`
	TranslationEnabled            bool   `json:"translation_enabled"`
	TranslationOnlyMode           bool   `json:"translation_only_mode"`
	TranslationProvider           string `json:"translation_provider"`
	UpdateInterval                int    `json:"update_interval"`
	WindowHeight                  string `json:"window_height"`
	WindowMaximized               string `json:"window_maximized"`
	WindowWidth                   string `json:"window_width"`
	WindowX                       string `json:"window_x"`
	WindowY                       string `json:"window_y"`
}

var defaults Defaults

func init() {
	if err := json.Unmarshal(defaultsJSON, &defaults); err != nil {
		panic("failed to parse defaults.json: " + err.Error())
	}
}

// Get returns the loaded defaults
func Get() Defaults {
	return defaults
}

// GetString returns a setting default as a string
func GetString(key string) string {
	switch key {
	case "ai_api_key":
		return defaults.AIAPIKey
	case "ai_chat_enabled":
		return strconv.FormatBool(defaults.AIChatEnabled)
	case "ai_custom_headers":
		return defaults.AICustomHeaders
	case "ai_endpoint":
		return defaults.AIEndpoint
	case "ai_model":
		return defaults.AIModel
	case "ai_search_enabled":
		return strconv.FormatBool(defaults.AISearchEnabled)
	case "ai_summary_prompt":
		return defaults.AISummaryPrompt
	case "ai_translation_prompt":
		return defaults.AITranslationPrompt
	case "ai_usage_limit":
		return defaults.AIUsageLimit
	case "ai_usage_tokens":
		return defaults.AIUsageTokens
	case "auto_cleanup_enabled":
		return strconv.FormatBool(defaults.AutoCleanupEnabled)
	case "auto_show_all_content":
		return strconv.FormatBool(defaults.AutoShowAllContent)
	case "baidu_app_id":
		return defaults.BaiduAppId
	case "baidu_secret_key":
		return defaults.BaiduSecretKey
	case "close_to_tray":
		return strconv.FormatBool(defaults.CloseToTray)
	case "content_font_family":
		return defaults.ContentFontFamily
	case "content_font_size":
		return strconv.Itoa(defaults.ContentFontSize)
	case "content_line_height":
		return defaults.ContentLineHeight
	case "custom_css_file":
		return defaults.CustomCssFile
	case "custom_translation_body_template":
		return defaults.CustomTranslationBodyTemplate
	case "custom_translation_enabled":
		return strconv.FormatBool(defaults.CustomTranslationEnabled)
	case "custom_translation_endpoint":
		return defaults.CustomTranslationEndpoint
	case "custom_translation_headers":
		return defaults.CustomTranslationHeaders
	case "custom_translation_lang_mapping":
		return defaults.CustomTranslationLangMapping
	case "custom_translation_method":
		return defaults.CustomTranslationMethod
	case "custom_translation_name":
		return defaults.CustomTranslationName
	case "custom_translation_response_path":
		return defaults.CustomTranslationResponsePath
	case "custom_translation_timeout":
		return strconv.Itoa(defaults.CustomTranslationTimeout)
	case "deepl_api_key":
		return defaults.DeeplAPIKey
	case "deepl_endpoint":
		return defaults.DeeplEndpoint
	case "default_view_mode":
		return defaults.DefaultViewMode
	case "feed_drawer_expanded":
		return strconv.FormatBool(defaults.FeedDrawerExpanded)
	case "feed_drawer_pinned":
		return strconv.FormatBool(defaults.FeedDrawerPinned)
	case "freshrss_api_password":
		return defaults.FreshRSSAPIPassword
	case "freshrss_auto_sync_interval":
		return strconv.Itoa(defaults.FreshRSSAutoSyncInterval)
	case "freshrss_enabled":
		return strconv.FormatBool(defaults.FreshRSSEnabled)
	case "freshrss_last_sync_time":
		return defaults.FreshRSSLastSyncTime
	case "freshrss_server_url":
		return defaults.FreshRSSServerUrl
	case "freshrss_sync_on_startup":
		return strconv.FormatBool(defaults.FreshRSSSyncOnStartup)
	case "freshrss_username":
		return defaults.FreshRSSUsername
	case "full_text_fetch_enabled":
		return strconv.FormatBool(defaults.FullTextFetchEnabled)
	case "google_translate_endpoint":
		return defaults.GoogleTranslateEndpoint
	case "hover_mark_as_read":
		return strconv.FormatBool(defaults.HoverMarkAsRead)
	case "image_gallery_enabled":
		return strconv.FormatBool(defaults.ImageGalleryEnabled)
	case "language":
		return defaults.Language
	case "last_global_refresh":
		return defaults.LastGlobalRefresh
	case "last_network_test":
		return defaults.LastNetworkTest
	case "layout_mode":
		return defaults.LayoutMode
	case "max_article_age_days":
		return strconv.Itoa(defaults.MaxArticleAgeDays)
	case "max_cache_size_mb":
		return strconv.Itoa(defaults.MaxCacheSizeMb)
	case "max_concurrent_refreshes":
		return defaults.MaxConcurrentRefreshes
	case "media_cache_enabled":
		return strconv.FormatBool(defaults.MediaCacheEnabled)
	case "media_cache_max_age_days":
		return strconv.Itoa(defaults.MediaCacheMaxAgeDays)
	case "media_cache_max_size_mb":
		return strconv.Itoa(defaults.MediaCacheMaxSizeMb)
	case "media_proxy_fallback":
		return strconv.FormatBool(defaults.MediaProxyFallback)
	case "network_bandwidth_mbps":
		return defaults.NetworkBandwidthMbps
	case "network_latency_ms":
		return defaults.NetworkLatencyMs
	case "network_speed":
		return defaults.NetworkSpeed
	case "obsidian_enabled":
		return strconv.FormatBool(defaults.ObsidianEnabled)
	case "obsidian_vault":
		return defaults.ObsidianVault
	case "obsidian_vault_path":
		return defaults.ObsidianVaultPath
	case "proxy_enabled":
		return strconv.FormatBool(defaults.ProxyEnabled)
	case "proxy_host":
		return defaults.ProxyHost
	case "proxy_password":
		return defaults.ProxyPassword
	case "proxy_port":
		return defaults.ProxyPort
	case "proxy_type":
		return defaults.ProxyType
	case "proxy_username":
		return defaults.ProxyUsername
	case "refresh_mode":
		return defaults.RefreshMode
	case "retry_timeout_seconds":
		return strconv.Itoa(defaults.RetryTimeoutSeconds)
	case "rsshub_api_key":
		return defaults.RsshubAPIKey
	case "rsshub_enabled":
		return strconv.FormatBool(defaults.RsshubEnabled)
	case "rsshub_endpoint":
		return defaults.RsshubEndpoint
	case "rules":
		return defaults.Rules
	case "shortcuts":
		return defaults.Shortcuts
	case "shortcuts_enabled":
		return strconv.FormatBool(defaults.ShortcutsEnabled)
	case "show_article_preview_images":
		return strconv.FormatBool(defaults.ShowArticlePreviewImages)
	case "show_hidden_articles":
		return strconv.FormatBool(defaults.ShowHiddenArticles)
	case "startup_on_boot":
		return strconv.FormatBool(defaults.StartupOnBoot)
	case "summary_enabled":
		return strconv.FormatBool(defaults.SummaryEnabled)
	case "summary_length":
		return defaults.SummaryLength
	case "summary_provider":
		return defaults.SummaryProvider
	case "summary_trigger_mode":
		return defaults.SummaryTriggerMode
	case "target_language":
		return defaults.TargetLanguage
	case "theme":
		return defaults.Theme
	case "translation_enabled":
		return strconv.FormatBool(defaults.TranslationEnabled)
	case "translation_only_mode":
		return strconv.FormatBool(defaults.TranslationOnlyMode)
	case "translation_provider":
		return defaults.TranslationProvider
	case "update_interval":
		return strconv.Itoa(defaults.UpdateInterval)
	case "window_height":
		return defaults.WindowHeight
	case "window_maximized":
		return defaults.WindowMaximized
	case "window_width":
		return defaults.WindowWidth
	case "window_x":
		return defaults.WindowX
	case "window_y":
		return defaults.WindowY
	default:
		return ""
	}
}
