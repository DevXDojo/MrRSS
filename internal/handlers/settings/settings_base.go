// Package settings provides handlers for application settings management.
// This file contains the base types and utilities for the definition-driven settings system.
// CODE GENERATED - DO NOT EDIT MANUALLY
// To add new settings, edit internal/config/settings_schema.json and run: go run tools/settings-generator/main.go
package settings

import (
	"MrRSS/internal/handlers/core"
)

// SettingDef defines a single setting's metadata
type SettingDef struct {
	Key       string // Database key (snake_case)
	Encrypted bool   // Whether the value should be encrypted in the database
}

// AllSettings returns all setting definitions in alphabetical order by key.
// This is the single source of truth for all settings.
var AllSettings = []SettingDef{
	{Key: "ai_api_key", Encrypted: true},
	{Key: "ai_chat_enabled", Encrypted: false},
	{Key: "ai_chat_profile_id", Encrypted: false},
	{Key: "ai_custom_headers", Encrypted: false},
	{Key: "ai_endpoint", Encrypted: false},
	{Key: "ai_model", Encrypted: false},
	{Key: "ai_search_enabled", Encrypted: false},
	{Key: "ai_search_profile_id", Encrypted: false},
	{Key: "ai_summary_profile_id", Encrypted: false},
	{Key: "ai_summary_prompt", Encrypted: false},
	{Key: "ai_translation_profile_id", Encrypted: false},
	{Key: "ai_translation_prompt", Encrypted: false},
	{Key: "ai_usage_limit", Encrypted: false},
	{Key: "ai_usage_tokens", Encrypted: false},
	{Key: "auto_cleanup_enabled", Encrypted: false},
	{Key: "auto_show_all_content", Encrypted: false},
	{Key: "baidu_app_id", Encrypted: false},
	{Key: "baidu_secret_key", Encrypted: true},
	{Key: "close_to_tray", Encrypted: false},
	{Key: "content_font_family", Encrypted: false},
	{Key: "content_font_size", Encrypted: false},
	{Key: "content_line_height", Encrypted: false},
	{Key: "custom_css_file", Encrypted: false},
	{Key: "custom_translation_body_template", Encrypted: false},
	{Key: "custom_translation_enabled", Encrypted: false},
	{Key: "custom_translation_endpoint", Encrypted: false},
	{Key: "custom_translation_headers", Encrypted: false},
	{Key: "custom_translation_lang_mapping", Encrypted: false},
	{Key: "custom_translation_method", Encrypted: false},
	{Key: "custom_translation_name", Encrypted: false},
	{Key: "custom_translation_response_path", Encrypted: false},
	{Key: "custom_translation_timeout", Encrypted: false},
	{Key: "deepl_api_key", Encrypted: true},
	{Key: "deepl_endpoint", Encrypted: false},
	{Key: "default_view_mode", Encrypted: false},
	{Key: "feed_drawer_expanded", Encrypted: false},
	{Key: "feed_drawer_pinned", Encrypted: false},
	{Key: "freshrss_api_password", Encrypted: true},
	{Key: "freshrss_auto_sync_interval", Encrypted: false},
	{Key: "freshrss_enabled", Encrypted: false},
	{Key: "freshrss_last_sync_time", Encrypted: false},
	{Key: "freshrss_server_url", Encrypted: false},
	{Key: "freshrss_sync_on_startup", Encrypted: false},
	{Key: "freshrss_username", Encrypted: false},
	{Key: "full_text_fetch_enabled", Encrypted: false},
	{Key: "google_translate_endpoint", Encrypted: false},
	{Key: "hover_mark_as_read", Encrypted: false},
	{Key: "image_gallery_enabled", Encrypted: false},
	{Key: "language", Encrypted: false},
	{Key: "last_global_refresh", Encrypted: false},
	{Key: "last_network_test", Encrypted: false},
	{Key: "layout_mode", Encrypted: false},
	{Key: "max_article_age_days", Encrypted: false},
	{Key: "max_cache_size_mb", Encrypted: false},
	{Key: "max_concurrent_refreshes", Encrypted: false},
	{Key: "media_cache_enabled", Encrypted: false},
	{Key: "media_cache_max_age_days", Encrypted: false},
	{Key: "media_cache_max_size_mb", Encrypted: false},
	{Key: "media_proxy_fallback", Encrypted: false},
	{Key: "microsoft_api_key", Encrypted: true},
	{Key: "microsoft_endpoint", Encrypted: false},
	{Key: "microsoft_region", Encrypted: false},
	{Key: "network_bandwidth_mbps", Encrypted: false},
	{Key: "network_latency_ms", Encrypted: false},
	{Key: "network_speed", Encrypted: false},
	{Key: "notion_api_key", Encrypted: true},
	{Key: "notion_enabled", Encrypted: false},
	{Key: "notion_page_id", Encrypted: false},
	{Key: "obsidian_enabled", Encrypted: false},
	{Key: "obsidian_vault", Encrypted: false},
	{Key: "obsidian_vault_path", Encrypted: false},
	{Key: "proxy_enabled", Encrypted: false},
	{Key: "proxy_host", Encrypted: false},
	{Key: "proxy_password", Encrypted: true},
	{Key: "proxy_port", Encrypted: false},
	{Key: "proxy_type", Encrypted: false},
	{Key: "proxy_username", Encrypted: true},
	{Key: "refresh_mode", Encrypted: false},
	{Key: "retry_timeout_seconds", Encrypted: false},
	{Key: "rsshub_api_key", Encrypted: true},
	{Key: "rsshub_enabled", Encrypted: false},
	{Key: "rsshub_endpoint", Encrypted: false},
	{Key: "rules", Encrypted: false},
	{Key: "shortcuts", Encrypted: false},
	{Key: "shortcuts_enabled", Encrypted: false},
	{Key: "show_article_preview_images", Encrypted: false},
	{Key: "show_floating_toc", Encrypted: false},
	{Key: "show_hidden_articles", Encrypted: false},
	{Key: "startup_on_boot", Encrypted: false},
	{Key: "summary_enabled", Encrypted: false},
	{Key: "summary_length", Encrypted: false},
	{Key: "summary_provider", Encrypted: false},
	{Key: "summary_trigger_mode", Encrypted: false},
	{Key: "target_language", Encrypted: false},
	{Key: "tencent_region", Encrypted: false},
	{Key: "tencent_secret_id", Encrypted: false},
	{Key: "tencent_secret_key", Encrypted: true},
	{Key: "theme", Encrypted: false},
	{Key: "translation_enabled", Encrypted: false},
	{Key: "translation_only_mode", Encrypted: false},
	{Key: "translation_provider", Encrypted: false},
	{Key: "ui_font_family", Encrypted: false},
	{Key: "ui_font_size", Encrypted: false},
	{Key: "update_check_enabled", Encrypted: false},
	{Key: "update_interval", Encrypted: false},
	{Key: "window_height", Encrypted: false},
	{Key: "window_maximized", Encrypted: false},
	{Key: "window_width", Encrypted: false},
	{Key: "window_x", Encrypted: false},
	{Key: "window_y", Encrypted: false},
	{Key: "zotero_api_key", Encrypted: true},
	{Key: "zotero_enabled", Encrypted: false},
	{Key: "zotero_user_id", Encrypted: false},
}

// GetAllSettings reads all settings from the database and returns them as a map.
// Encrypted settings are automatically decrypted.
func GetAllSettings(h *core.Handler) map[string]string {
	result := make(map[string]string, len(AllSettings))

	for _, def := range AllSettings {
		var value string
		if def.Encrypted {
			value = safeGetEncryptedSetting(h, def.Key)
		} else {
			value = safeGetSetting(h, def.Key)
		}
		result[def.Key] = value
	}

	return result
}

// SaveSettings saves settings from a map to the database.
// Empty string values are skipped (to allow partial updates).
// Encrypted settings are automatically encrypted.
func SaveSettings(h *core.Handler, settings map[string]string) error {
	// Create a lookup for encrypted keys
	encryptedKeys := make(map[string]bool, len(AllSettings))
	for _, def := range AllSettings {
		if def.Encrypted {
			encryptedKeys[def.Key] = true
		}
	}

	// Save each setting
	for key, value := range settings {
		if encryptedKeys[key] {
			if err := h.DB.SetEncryptedSetting(key, value); err != nil {
				return err
			}
		} else if value != "" {
			h.DB.SetSetting(key, value)
		}
	}

	return nil
}

// IsEncryptedSetting returns true if the given key is an encrypted setting.
func IsEncryptedSetting(key string) bool {
	for _, def := range AllSettings {
		if def.Key == key {
			return def.Encrypted
		}
	}
	return false
}
