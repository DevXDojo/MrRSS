// Copyright 2026 Ch3nyang & MrRSS Team. All rights reserved.
//
// Auto-generated settings composable helpers
// CODE GENERATED - DO NOT EDIT MANUALLY
// To add new settings, edit internal/config/settings_schema.json and run: go run tools/settings-generator/main.go
import { type Ref } from 'vue';
import type { SettingsData } from '@/types/settings.generated';
import { settingsDefaults } from '@/config/defaults';

/**
 * Generate the initial settings object with defaults
 * This should be used in useSettings() to initialize the settings ref
 */
export function generateInitialSettings(): SettingsData {
  return {
    ai_api_key: settingsDefaults.ai_api_key,
    ai_chat_enabled: settingsDefaults.ai_chat_enabled,
    ai_custom_headers: settingsDefaults.ai_custom_headers,
    ai_endpoint: settingsDefaults.ai_endpoint,
    ai_model: settingsDefaults.ai_model,
    ai_search_enabled: settingsDefaults.ai_search_enabled,
    ai_summary_prompt: settingsDefaults.ai_summary_prompt,
    ai_translation_prompt: settingsDefaults.ai_translation_prompt,
    ai_usage_limit: settingsDefaults.ai_usage_limit,
    ai_usage_tokens: settingsDefaults.ai_usage_tokens,
    auto_cleanup_enabled: settingsDefaults.auto_cleanup_enabled,
    auto_show_all_content: settingsDefaults.auto_show_all_content,
    baidu_app_id: settingsDefaults.baidu_app_id,
    baidu_secret_key: settingsDefaults.baidu_secret_key,
    close_to_tray: settingsDefaults.close_to_tray,
    content_font_family: settingsDefaults.content_font_family,
    content_font_size: settingsDefaults.content_font_size,
    content_line_height: settingsDefaults.content_line_height,
    custom_css_file: settingsDefaults.custom_css_file,
    custom_translation_body_template: settingsDefaults.custom_translation_body_template,
    custom_translation_enabled: settingsDefaults.custom_translation_enabled,
    custom_translation_endpoint: settingsDefaults.custom_translation_endpoint,
    custom_translation_headers: settingsDefaults.custom_translation_headers,
    custom_translation_lang_mapping: settingsDefaults.custom_translation_lang_mapping,
    custom_translation_method: settingsDefaults.custom_translation_method,
    custom_translation_name: settingsDefaults.custom_translation_name,
    custom_translation_response_path: settingsDefaults.custom_translation_response_path,
    custom_translation_timeout: settingsDefaults.custom_translation_timeout,
    deepl_api_key: settingsDefaults.deepl_api_key,
    deepl_endpoint: settingsDefaults.deepl_endpoint,
    default_view_mode: settingsDefaults.default_view_mode,
    feed_drawer_expanded: settingsDefaults.feed_drawer_expanded,
    feed_drawer_pinned: settingsDefaults.feed_drawer_pinned,
    freshrss_api_password: settingsDefaults.freshrss_api_password,
    freshrss_auto_sync_interval: settingsDefaults.freshrss_auto_sync_interval,
    freshrss_enabled: settingsDefaults.freshrss_enabled,
    freshrss_last_sync_time: settingsDefaults.freshrss_last_sync_time,
    freshrss_server_url: settingsDefaults.freshrss_server_url,
    freshrss_sync_on_startup: settingsDefaults.freshrss_sync_on_startup,
    freshrss_username: settingsDefaults.freshrss_username,
    full_text_fetch_enabled: settingsDefaults.full_text_fetch_enabled,
    google_translate_endpoint: settingsDefaults.google_translate_endpoint,
    hover_mark_as_read: settingsDefaults.hover_mark_as_read,
    image_gallery_enabled: settingsDefaults.image_gallery_enabled,
    language: settingsDefaults.language,
    last_global_refresh: settingsDefaults.last_global_refresh,
    last_network_test: settingsDefaults.last_network_test,
    layout_mode: settingsDefaults.layout_mode,
    max_article_age_days: settingsDefaults.max_article_age_days,
    max_cache_size_mb: settingsDefaults.max_cache_size_mb,
    max_concurrent_refreshes: settingsDefaults.max_concurrent_refreshes,
    media_cache_enabled: settingsDefaults.media_cache_enabled,
    media_cache_max_age_days: settingsDefaults.media_cache_max_age_days,
    media_cache_max_size_mb: settingsDefaults.media_cache_max_size_mb,
    media_proxy_fallback: settingsDefaults.media_proxy_fallback,
    network_bandwidth_mbps: settingsDefaults.network_bandwidth_mbps,
    network_latency_ms: settingsDefaults.network_latency_ms,
    network_speed: settingsDefaults.network_speed,
    obsidian_enabled: settingsDefaults.obsidian_enabled,
    obsidian_vault: settingsDefaults.obsidian_vault,
    obsidian_vault_path: settingsDefaults.obsidian_vault_path,
    proxy_enabled: settingsDefaults.proxy_enabled,
    proxy_host: settingsDefaults.proxy_host,
    proxy_password: settingsDefaults.proxy_password,
    proxy_port: settingsDefaults.proxy_port,
    proxy_type: settingsDefaults.proxy_type,
    proxy_username: settingsDefaults.proxy_username,
    refresh_mode: settingsDefaults.refresh_mode,
    retry_timeout_seconds: settingsDefaults.retry_timeout_seconds,
    rsshub_api_key: settingsDefaults.rsshub_api_key,
    rsshub_enabled: settingsDefaults.rsshub_enabled,
    rsshub_endpoint: settingsDefaults.rsshub_endpoint,
    rules: settingsDefaults.rules,
    shortcuts: settingsDefaults.shortcuts,
    shortcuts_enabled: settingsDefaults.shortcuts_enabled,
    show_article_preview_images: settingsDefaults.show_article_preview_images,
    show_hidden_articles: settingsDefaults.show_hidden_articles,
    startup_on_boot: settingsDefaults.startup_on_boot,
    summary_enabled: settingsDefaults.summary_enabled,
    summary_length: settingsDefaults.summary_length,
    summary_provider: settingsDefaults.summary_provider,
    summary_trigger_mode: settingsDefaults.summary_trigger_mode,
    target_language: settingsDefaults.target_language,
    theme: settingsDefaults.theme,
    translation_enabled: settingsDefaults.translation_enabled,
    translation_only_mode: settingsDefaults.translation_only_mode,
    translation_provider: settingsDefaults.translation_provider,
    update_interval: settingsDefaults.update_interval,
    window_height: settingsDefaults.window_height,
    window_maximized: settingsDefaults.window_maximized,
    window_width: settingsDefaults.window_width,
    window_x: settingsDefaults.window_x,
    window_y: settingsDefaults.window_y,
  } as SettingsData;
}

/**
 * Generate the fetchSettings response parser
 * This should be used in useSettings() fetchSettings() to parse backend data
 */
export function parseSettingsData(data: Record<string, string>): SettingsData {
  return {
    ai_api_key: data.ai_api_key || settingsDefaults.ai_api_key,
    ai_chat_enabled: data.ai_chat_enabled === 'true',
    ai_custom_headers: data.ai_custom_headers || settingsDefaults.ai_custom_headers,
    ai_endpoint: data.ai_endpoint || settingsDefaults.ai_endpoint,
    ai_model: data.ai_model || settingsDefaults.ai_model,
    ai_search_enabled: data.ai_search_enabled === 'true',
    ai_summary_prompt: data.ai_summary_prompt || settingsDefaults.ai_summary_prompt,
    ai_translation_prompt: data.ai_translation_prompt || settingsDefaults.ai_translation_prompt,
    ai_usage_limit: data.ai_usage_limit || settingsDefaults.ai_usage_limit,
    ai_usage_tokens: data.ai_usage_tokens || settingsDefaults.ai_usage_tokens,
    auto_cleanup_enabled: data.auto_cleanup_enabled === 'true',
    auto_show_all_content: data.auto_show_all_content === 'true',
    baidu_app_id: data.baidu_app_id || settingsDefaults.baidu_app_id,
    baidu_secret_key: data.baidu_secret_key || settingsDefaults.baidu_secret_key,
    close_to_tray: data.close_to_tray === 'true',
    content_font_family: data.content_font_family || settingsDefaults.content_font_family,
    content_font_size: parseInt(data.content_font_size) || settingsDefaults.content_font_size,
    content_line_height: data.content_line_height || settingsDefaults.content_line_height,
    custom_css_file: data.custom_css_file || settingsDefaults.custom_css_file,
    custom_translation_body_template:
      data.custom_translation_body_template || settingsDefaults.custom_translation_body_template,
    custom_translation_enabled: data.custom_translation_enabled === 'true',
    custom_translation_endpoint:
      data.custom_translation_endpoint || settingsDefaults.custom_translation_endpoint,
    custom_translation_headers:
      data.custom_translation_headers || settingsDefaults.custom_translation_headers,
    custom_translation_lang_mapping:
      data.custom_translation_lang_mapping || settingsDefaults.custom_translation_lang_mapping,
    custom_translation_method:
      data.custom_translation_method || settingsDefaults.custom_translation_method,
    custom_translation_name:
      data.custom_translation_name || settingsDefaults.custom_translation_name,
    custom_translation_response_path:
      data.custom_translation_response_path || settingsDefaults.custom_translation_response_path,
    custom_translation_timeout:
      parseInt(data.custom_translation_timeout) || settingsDefaults.custom_translation_timeout,
    deepl_api_key: data.deepl_api_key || settingsDefaults.deepl_api_key,
    deepl_endpoint: data.deepl_endpoint || settingsDefaults.deepl_endpoint,
    default_view_mode: data.default_view_mode || settingsDefaults.default_view_mode,
    feed_drawer_expanded: data.feed_drawer_expanded === 'true',
    feed_drawer_pinned: data.feed_drawer_pinned === 'true',
    freshrss_api_password: data.freshrss_api_password || settingsDefaults.freshrss_api_password,
    freshrss_auto_sync_interval:
      parseInt(data.freshrss_auto_sync_interval) || settingsDefaults.freshrss_auto_sync_interval,
    freshrss_enabled: data.freshrss_enabled === 'true',
    freshrss_last_sync_time:
      data.freshrss_last_sync_time || settingsDefaults.freshrss_last_sync_time,
    freshrss_server_url: data.freshrss_server_url || settingsDefaults.freshrss_server_url,
    freshrss_sync_on_startup: data.freshrss_sync_on_startup === 'true',
    freshrss_username: data.freshrss_username || settingsDefaults.freshrss_username,
    full_text_fetch_enabled: data.full_text_fetch_enabled === 'true',
    google_translate_endpoint:
      data.google_translate_endpoint || settingsDefaults.google_translate_endpoint,
    hover_mark_as_read: data.hover_mark_as_read === 'true',
    image_gallery_enabled: data.image_gallery_enabled === 'true',
    language: data.language || settingsDefaults.language,
    last_global_refresh: data.last_global_refresh || settingsDefaults.last_global_refresh,
    last_network_test: data.last_network_test || settingsDefaults.last_network_test,
    layout_mode: data.layout_mode || settingsDefaults.layout_mode,
    max_article_age_days:
      parseInt(data.max_article_age_days) || settingsDefaults.max_article_age_days,
    max_cache_size_mb: parseInt(data.max_cache_size_mb) || settingsDefaults.max_cache_size_mb,
    max_concurrent_refreshes:
      data.max_concurrent_refreshes || settingsDefaults.max_concurrent_refreshes,
    media_cache_enabled: data.media_cache_enabled === 'true',
    media_cache_max_age_days:
      parseInt(data.media_cache_max_age_days) || settingsDefaults.media_cache_max_age_days,
    media_cache_max_size_mb:
      parseInt(data.media_cache_max_size_mb) || settingsDefaults.media_cache_max_size_mb,
    media_proxy_fallback: data.media_proxy_fallback === 'true',
    network_bandwidth_mbps: data.network_bandwidth_mbps || settingsDefaults.network_bandwidth_mbps,
    network_latency_ms: data.network_latency_ms || settingsDefaults.network_latency_ms,
    network_speed: data.network_speed || settingsDefaults.network_speed,
    obsidian_enabled: data.obsidian_enabled === 'true',
    obsidian_vault: data.obsidian_vault || settingsDefaults.obsidian_vault,
    obsidian_vault_path: data.obsidian_vault_path || settingsDefaults.obsidian_vault_path,
    proxy_enabled: data.proxy_enabled === 'true',
    proxy_host: data.proxy_host || settingsDefaults.proxy_host,
    proxy_password: data.proxy_password || settingsDefaults.proxy_password,
    proxy_port: data.proxy_port || settingsDefaults.proxy_port,
    proxy_type: data.proxy_type || settingsDefaults.proxy_type,
    proxy_username: data.proxy_username || settingsDefaults.proxy_username,
    refresh_mode: data.refresh_mode || settingsDefaults.refresh_mode,
    retry_timeout_seconds:
      parseInt(data.retry_timeout_seconds) || settingsDefaults.retry_timeout_seconds,
    rsshub_api_key: data.rsshub_api_key || settingsDefaults.rsshub_api_key,
    rsshub_enabled: data.rsshub_enabled === 'true',
    rsshub_endpoint: data.rsshub_endpoint || settingsDefaults.rsshub_endpoint,
    rules: data.rules || settingsDefaults.rules,
    shortcuts: data.shortcuts || settingsDefaults.shortcuts,
    shortcuts_enabled: data.shortcuts_enabled === 'true',
    show_article_preview_images: data.show_article_preview_images === 'true',
    show_hidden_articles: data.show_hidden_articles === 'true',
    startup_on_boot: data.startup_on_boot === 'true',
    summary_enabled: data.summary_enabled === 'true',
    summary_length: data.summary_length || settingsDefaults.summary_length,
    summary_provider: data.summary_provider || settingsDefaults.summary_provider,
    summary_trigger_mode: data.summary_trigger_mode || settingsDefaults.summary_trigger_mode,
    target_language: data.target_language || settingsDefaults.target_language,
    theme: data.theme || settingsDefaults.theme,
    translation_enabled: data.translation_enabled === 'true',
    translation_only_mode: data.translation_only_mode === 'true',
    translation_provider: data.translation_provider || settingsDefaults.translation_provider,
    update_interval: parseInt(data.update_interval) || settingsDefaults.update_interval,
    window_height: data.window_height || settingsDefaults.window_height,
    window_maximized: data.window_maximized || settingsDefaults.window_maximized,
    window_width: data.window_width || settingsDefaults.window_width,
    window_x: data.window_x || settingsDefaults.window_x,
    window_y: data.window_y || settingsDefaults.window_y,
  } as SettingsData;
}

/**
 * Generate the auto-save payload
 * This should be used in useSettingsAutoSave.ts to build the save payload
 */
export function buildAutoSavePayload(settingsRef: Ref<SettingsData>): Record<string, string> {
  return {
    ai_api_key: settingsRef.value.ai_api_key ?? settingsDefaults.ai_api_key,
    ai_chat_enabled: (
      settingsRef.value.ai_chat_enabled ?? settingsDefaults.ai_chat_enabled
    ).toString(),
    ai_custom_headers: settingsRef.value.ai_custom_headers ?? settingsDefaults.ai_custom_headers,
    ai_endpoint: settingsRef.value.ai_endpoint ?? settingsDefaults.ai_endpoint,
    ai_model: settingsRef.value.ai_model ?? settingsDefaults.ai_model,
    ai_search_enabled: (
      settingsRef.value.ai_search_enabled ?? settingsDefaults.ai_search_enabled
    ).toString(),
    ai_summary_prompt: settingsRef.value.ai_summary_prompt ?? settingsDefaults.ai_summary_prompt,
    ai_translation_prompt:
      settingsRef.value.ai_translation_prompt ?? settingsDefaults.ai_translation_prompt,
    ai_usage_limit: settingsRef.value.ai_usage_limit ?? settingsDefaults.ai_usage_limit,
    ai_usage_tokens: settingsRef.value.ai_usage_tokens ?? settingsDefaults.ai_usage_tokens,
    auto_cleanup_enabled: (
      settingsRef.value.auto_cleanup_enabled ?? settingsDefaults.auto_cleanup_enabled
    ).toString(),
    auto_show_all_content: (
      settingsRef.value.auto_show_all_content ?? settingsDefaults.auto_show_all_content
    ).toString(),
    baidu_app_id: settingsRef.value.baidu_app_id ?? settingsDefaults.baidu_app_id,
    baidu_secret_key: settingsRef.value.baidu_secret_key ?? settingsDefaults.baidu_secret_key,
    close_to_tray: (settingsRef.value.close_to_tray ?? settingsDefaults.close_to_tray).toString(),
    content_font_family:
      settingsRef.value.content_font_family ?? settingsDefaults.content_font_family,
    content_font_size: (
      settingsRef.value.content_font_size ?? settingsDefaults.content_font_size
    ).toString(),
    content_line_height:
      settingsRef.value.content_line_height ?? settingsDefaults.content_line_height,
    custom_css_file: settingsRef.value.custom_css_file ?? settingsDefaults.custom_css_file,
    custom_translation_body_template:
      settingsRef.value.custom_translation_body_template ??
      settingsDefaults.custom_translation_body_template,
    custom_translation_enabled: (
      settingsRef.value.custom_translation_enabled ?? settingsDefaults.custom_translation_enabled
    ).toString(),
    custom_translation_endpoint:
      settingsRef.value.custom_translation_endpoint ?? settingsDefaults.custom_translation_endpoint,
    custom_translation_headers:
      settingsRef.value.custom_translation_headers ?? settingsDefaults.custom_translation_headers,
    custom_translation_lang_mapping:
      settingsRef.value.custom_translation_lang_mapping ??
      settingsDefaults.custom_translation_lang_mapping,
    custom_translation_method:
      settingsRef.value.custom_translation_method ?? settingsDefaults.custom_translation_method,
    custom_translation_name:
      settingsRef.value.custom_translation_name ?? settingsDefaults.custom_translation_name,
    custom_translation_response_path:
      settingsRef.value.custom_translation_response_path ??
      settingsDefaults.custom_translation_response_path,
    custom_translation_timeout: (
      settingsRef.value.custom_translation_timeout ?? settingsDefaults.custom_translation_timeout
    ).toString(),
    deepl_api_key: settingsRef.value.deepl_api_key ?? settingsDefaults.deepl_api_key,
    deepl_endpoint: settingsRef.value.deepl_endpoint ?? settingsDefaults.deepl_endpoint,
    default_view_mode: settingsRef.value.default_view_mode ?? settingsDefaults.default_view_mode,
    freshrss_api_password:
      settingsRef.value.freshrss_api_password ?? settingsDefaults.freshrss_api_password,
    freshrss_auto_sync_interval: (
      settingsRef.value.freshrss_auto_sync_interval ?? settingsDefaults.freshrss_auto_sync_interval
    ).toString(),
    freshrss_enabled: (
      settingsRef.value.freshrss_enabled ?? settingsDefaults.freshrss_enabled
    ).toString(),
    freshrss_last_sync_time:
      settingsRef.value.freshrss_last_sync_time ?? settingsDefaults.freshrss_last_sync_time,
    freshrss_server_url:
      settingsRef.value.freshrss_server_url ?? settingsDefaults.freshrss_server_url,
    freshrss_sync_on_startup: (
      settingsRef.value.freshrss_sync_on_startup ?? settingsDefaults.freshrss_sync_on_startup
    ).toString(),
    freshrss_username: settingsRef.value.freshrss_username ?? settingsDefaults.freshrss_username,
    full_text_fetch_enabled: (
      settingsRef.value.full_text_fetch_enabled ?? settingsDefaults.full_text_fetch_enabled
    ).toString(),
    google_translate_endpoint:
      settingsRef.value.google_translate_endpoint ?? settingsDefaults.google_translate_endpoint,
    hover_mark_as_read: (
      settingsRef.value.hover_mark_as_read ?? settingsDefaults.hover_mark_as_read
    ).toString(),
    image_gallery_enabled: (
      settingsRef.value.image_gallery_enabled ?? settingsDefaults.image_gallery_enabled
    ).toString(),
    language: settingsRef.value.language ?? settingsDefaults.language,
    last_network_test: settingsRef.value.last_network_test ?? settingsDefaults.last_network_test,
    layout_mode: settingsRef.value.layout_mode ?? settingsDefaults.layout_mode,
    max_article_age_days: (
      settingsRef.value.max_article_age_days ?? settingsDefaults.max_article_age_days
    ).toString(),
    max_cache_size_mb: (
      settingsRef.value.max_cache_size_mb ?? settingsDefaults.max_cache_size_mb
    ).toString(),
    max_concurrent_refreshes:
      settingsRef.value.max_concurrent_refreshes ?? settingsDefaults.max_concurrent_refreshes,
    media_cache_enabled: (
      settingsRef.value.media_cache_enabled ?? settingsDefaults.media_cache_enabled
    ).toString(),
    media_cache_max_age_days: (
      settingsRef.value.media_cache_max_age_days ?? settingsDefaults.media_cache_max_age_days
    ).toString(),
    media_cache_max_size_mb: (
      settingsRef.value.media_cache_max_size_mb ?? settingsDefaults.media_cache_max_size_mb
    ).toString(),
    media_proxy_fallback: (
      settingsRef.value.media_proxy_fallback ?? settingsDefaults.media_proxy_fallback
    ).toString(),
    network_bandwidth_mbps:
      settingsRef.value.network_bandwidth_mbps ?? settingsDefaults.network_bandwidth_mbps,
    network_latency_ms: settingsRef.value.network_latency_ms ?? settingsDefaults.network_latency_ms,
    network_speed: settingsRef.value.network_speed ?? settingsDefaults.network_speed,
    obsidian_enabled: (
      settingsRef.value.obsidian_enabled ?? settingsDefaults.obsidian_enabled
    ).toString(),
    obsidian_vault: settingsRef.value.obsidian_vault ?? settingsDefaults.obsidian_vault,
    obsidian_vault_path:
      settingsRef.value.obsidian_vault_path ?? settingsDefaults.obsidian_vault_path,
    proxy_enabled: (settingsRef.value.proxy_enabled ?? settingsDefaults.proxy_enabled).toString(),
    proxy_host: settingsRef.value.proxy_host ?? settingsDefaults.proxy_host,
    proxy_password: settingsRef.value.proxy_password ?? settingsDefaults.proxy_password,
    proxy_port: settingsRef.value.proxy_port ?? settingsDefaults.proxy_port,
    proxy_type: settingsRef.value.proxy_type ?? settingsDefaults.proxy_type,
    proxy_username: settingsRef.value.proxy_username ?? settingsDefaults.proxy_username,
    refresh_mode: settingsRef.value.refresh_mode ?? settingsDefaults.refresh_mode,
    retry_timeout_seconds: (
      settingsRef.value.retry_timeout_seconds ?? settingsDefaults.retry_timeout_seconds
    ).toString(),
    rsshub_api_key: settingsRef.value.rsshub_api_key ?? settingsDefaults.rsshub_api_key,
    rsshub_enabled: (
      settingsRef.value.rsshub_enabled ?? settingsDefaults.rsshub_enabled
    ).toString(),
    rsshub_endpoint: settingsRef.value.rsshub_endpoint ?? settingsDefaults.rsshub_endpoint,
    rules: settingsRef.value.rules ?? settingsDefaults.rules,
    shortcuts: settingsRef.value.shortcuts ?? settingsDefaults.shortcuts,
    shortcuts_enabled: (
      settingsRef.value.shortcuts_enabled ?? settingsDefaults.shortcuts_enabled
    ).toString(),
    show_article_preview_images: (
      settingsRef.value.show_article_preview_images ?? settingsDefaults.show_article_preview_images
    ).toString(),
    show_hidden_articles: (
      settingsRef.value.show_hidden_articles ?? settingsDefaults.show_hidden_articles
    ).toString(),
    startup_on_boot: (
      settingsRef.value.startup_on_boot ?? settingsDefaults.startup_on_boot
    ).toString(),
    summary_enabled: (
      settingsRef.value.summary_enabled ?? settingsDefaults.summary_enabled
    ).toString(),
    summary_length: settingsRef.value.summary_length ?? settingsDefaults.summary_length,
    summary_provider: settingsRef.value.summary_provider ?? settingsDefaults.summary_provider,
    summary_trigger_mode:
      settingsRef.value.summary_trigger_mode ?? settingsDefaults.summary_trigger_mode,
    target_language: settingsRef.value.target_language ?? settingsDefaults.target_language,
    theme: settingsRef.value.theme ?? settingsDefaults.theme,
    translation_enabled: (
      settingsRef.value.translation_enabled ?? settingsDefaults.translation_enabled
    ).toString(),
    translation_only_mode: (
      settingsRef.value.translation_only_mode ?? settingsDefaults.translation_only_mode
    ).toString(),
    translation_provider:
      settingsRef.value.translation_provider ?? settingsDefaults.translation_provider,
    update_interval: (
      settingsRef.value.update_interval ?? settingsDefaults.update_interval
    ).toString(),
  };
}
