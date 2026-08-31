#!/usr/bin/env python3
"""Generate the macOS client's settings catalogue from the backend schema.

The schema in internal/config/settings_schema.json is the single source of
truth for what settings exist. This script pairs each one with the wording in
the client's translation catalogue and writes SettingsCatalog.generated.swift.

Run it after adding a setting to the schema:

    python3 tools/settings-swift/generate.py
"""

from __future__ import annotations

import json
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parents[2]
SCHEMA = ROOT / "internal/config/settings_schema.json"
TABLES = ROOT / "frontend/Sources/Localization/LocalizationTables.swift"
OUTPUT = ROOT / "frontend/Sources/Models/SettingsCatalog.generated.swift"

# Settings the interface never shows: window geometry, cached measurements and
# values the application maintains for itself.
HIDDEN = {
    "rules",
    "last_global_refresh",
    "window_x",
    "window_y",
    "window_width",
    "window_height",
    "window_maximized",
    "network_speed",
    "network_bandwidth_mbps",
    "network_latency_ms",
    "last_network_test",
    "freshrss_last_sync_time",
    "ai_usage_count",
    "ai_usage_reset_at",
}

# Where the schema's category does not match the tab the previous interface
# showed the setting on.
PANE_OVERRIDES = {
    "custom_css_file": "customization",
    "layout_mode": "customization",
    "ui_font_family": "typography",
    "ui_font_size": "typography",
    "content_font_family": "typography",
    "content_font_size": "typography",
    "content_line_height": "typography",
}

# Translation keys that cannot be derived from the setting name.
LABEL_OVERRIDES = {
    "update_interval": "setting.feed.refreshInterval",
    "ai_usage_limit": "setting.ai.setUsageLimit",
    "auto_cleanup_enabled": "setting.database.autoCleanup",
    "max_cache_size_mb": "setting.database.maxCacheSize",
    "max_article_age_days": "setting.database.maxArticleAge",
    "media_proxy_fallback": "setting.database.mediaProxyFallback",
    "media_cache_max_size_mb": "setting.database.mediaCacheMaxSize",
    "media_cache_max_age_days": "setting.database.mediaCacheMaxAge",
    "max_concurrent_refreshes": "setting.network.maxConcurrent",
    "retry_timeout_seconds": "setting.feed.retryTimeout",
    "custom_css_file": "setting.customization.customCss",
    "obsidian_enabled": "setting.plugins.obsidian.enabled",
    "obsidian_vault": "setting.plugins.obsidian.vaultName",
    "obsidian_vault_path": "setting.plugins.obsidian.vaultPath",
    "notion_enabled": "setting.plugins.notion.enabled",
    "notion_api_key": "setting.plugins.notion.apiKey",
    "notion_page_id": "setting.plugins.notion.pageId",
    "zotero_enabled": "setting.plugins.zotero.enabled",
    "zotero_api_key": "setting.plugins.zotero.apiKey",
    "zotero_user_id": "setting.plugins.zotero.userId",
    "freshrss_enabled": "setting.freshrss.enabled",
    "freshrss_server_url": "setting.freshrss.serverUrl",
    "freshrss_username": "setting.freshrss.username",
    "freshrss_api_password": "setting.freshrss.apiPassword",
    "freshrss_auto_sync_interval": "setting.freshrss.autoSyncInterval",
    "freshrss_sync_on_startup": "setting.freshrss.syncOnStartup",
    "rsshub_enabled": "setting.rsshub.enabled",
    "rsshub_endpoint": "setting.rsshub.endpoint",
    "rsshub_api_key": "setting.rsshub.apiKey",
    "custom_translation_enabled": "setting.translation.custom.enabled",
    "custom_translation_name": "setting.translation.custom.name",
    "custom_translation_endpoint": "setting.translation.custom.endpoint",
    "custom_translation_method": "setting.translation.custom.method",
    "custom_translation_headers": "setting.translation.custom.headers",
    "custom_translation_body_template": "setting.translation.custom.bodyTemplate",
    "custom_translation_response_path": "setting.translation.custom.responsePath",
    "custom_translation_lang_mapping": "setting.translation.custom.langMapping",
    "custom_translation_timeout": "setting.translation.custom.timeout",
    "ai_translation_profile_id": "setting.ai.selectProfileForTranslation",
    "ai_summary_profile_id": "setting.ai.selectProfileForSummary",
    "ai_chat_profile_id": "setting.ai.selectProfileForChat",
    "ai_search_profile_id": "setting.ai.selectProfileForSearch",
}

# Settings whose value is chosen from a fixed list.
CHOICES = {
    "refresh_mode": [("fixed", "setting.feed.fixedInterval"),
                     ("smart", "setting.feed.intelligentInterval")],
    "language": [("en-US", None), ("zh-CN", None)],
    "theme": [("auto", "setting.general.themeAuto"),
              ("light", "setting.general.themeLight"),
              ("dark", "setting.general.themeDark")],
    "default_view_mode": [("rendered", "setting.reading.viewAsRendered"),
                          ("webpage", "setting.reading.viewAsWebpage")],
    "translation_provider": [("google", "setting.content.googleTranslate"),
                             ("deepl", "setting.content.deeplTranslate"),
                             ("baidu", "setting.content.baiduTranslate"),
                             ("microsoft", "setting.content.microsoftTranslate"),
                             ("tencent", "setting.content.tencentTranslate"),
                             ("ai", "setting.content.aiTranslation"),
                             ("custom", "setting.translation.custom.name")],
    "summary_provider": [("local", "setting.content.localSummary"),
                         ("ai", "setting.content.aiSummary")],
    "summary_length": [("short", "setting.content.summaryShort"),
                       ("medium", "setting.content.summaryMedium"),
                       ("long", "setting.content.summaryLong")],
    "summary_trigger_mode": [("manual", "setting.content.summaryManual"),
                             ("auto", "setting.content.summaryAuto")],
    "proxy_type": [("http", None), ("https", None), ("socks5", None)],
    "content_font_family": [("system", "setting.typography.fontSystem"),
                            ("serif", "setting.typography.fontSerif"),
                            ("monospace", "setting.typography.fontMonospace")],
    "ui_font_family": [("system", "setting.typography.fontSystem"),
                       ("serif", "setting.typography.fontSerif"),
                       ("monospace", "setting.typography.fontMonospace")],
    "layout_mode": [("normal", "setting.customization.layoutNormal"),
                    ("compact", "setting.customization.layoutCompact"),
                    ("wide", "setting.customization.layoutWide")],
    "custom_translation_method": [("POST", None), ("GET", None)],
}

TABS = [
    "general", "reading", "content", "translation", "ai", "network", "database",
    "feed", "freshrss", "plugins", "rsshub", "customization", "typography",
    "update", "statistic", "about", "rule", "shortcut",
]


def load_translations() -> dict[str, str]:
    source = TABLES.read_text()
    match = re.search(r'static let english = #"""\n(.*?)\n"""#', source, re.DOTALL)
    if not match:
        raise SystemExit("could not read the English table from LocalizationTables.swift")
    return json.loads(match.group(1))


def camel(key: str) -> str:
    parts = key.split("_")
    return parts[0] + "".join(part.capitalize() for part in parts[1:])


def label_key(key: str, meta: dict, translations: dict[str, str]) -> str | None:
    if key in LABEL_OVERRIDES:
        return LABEL_OVERRIDES[key] if LABEL_OVERRIDES[key] in translations else None

    candidates = [meta.get("frontend_key") or "", camel(key)]
    if key.endswith("_enabled"):
        base = key[: -len("_enabled")].split("_")
        candidates.append("enable" + "".join(part.capitalize() for part in base))

    for candidate in candidates:
        if not candidate:
            continue
        for tab in TABS:
            full = f"setting.{tab}.{candidate}"
            if full in translations:
                return full
    return None


def humanise(key: str) -> str:
    special = {"ai": "AI", "api": "API", "rss": "RSS", "url": "URL", "css": "CSS",
               "mb": "MB", "ms": "ms", "id": "ID", "imap": "IMAP", "ui": "UI"}
    words = []
    for word in key.split("_"):
        words.append(special.get(word.lower(), word.capitalize()))
    return " ".join(words)


def swift_string(value: str) -> str:
    escaped = value.replace("\\", "\\\\").replace('"', '\\"')
    return f'"{escaped}"'


def main() -> int:
    schema = json.loads(SCHEMA.read_text())["settings"]
    translations = load_translations()

    entries = []
    for key in sorted(schema):
        meta = schema[key]
        if key in HIDDEN or meta["category"] == "internal":
            continue

        pane = PANE_OVERRIDES.get(key, meta["category"])
        label = label_key(key, meta, translations)
        description = None
        if label and f"{label}Desc" in translations:
            description = f"{label}Desc"

        kind = meta["type"]
        if key in CHOICES:
            control = "choice"
        elif kind == "bool":
            control = "toggle"
        elif kind == "int":
            control = "number"
        elif meta.get("encrypted") or key.endswith(("_key", "_password", "_secret", "_secret_key")):
            control = "secret"
        else:
            control = "text"

        choices = ""
        if key in CHOICES:
            pairs = []
            for value, translation in CHOICES[key]:
                title = swift_string(translation) if translation and translation in translations else "nil"
                pairs.append(f"SettingChoice(value: {swift_string(value)}, titleKey: {title})")
            choices = ", ".join(pairs)

        entries.append(
            "        SettingDefinition(\n"
            f"            key: {swift_string(key)},\n"
            f"            pane: .{pane},\n"
            f"            control: .{control},\n"
            f"            defaultValue: {swift_string(str(meta['default']).lower() if kind == 'bool' else str(meta['default']))},\n"
            f"            titleKey: {swift_string(label) if label else 'nil'},\n"
            f"            fallbackTitle: {swift_string(humanise(key))},\n"
            f"            detailKey: {swift_string(description) if description else 'nil'},\n"
            f"            choices: [{choices}]\n"
            "        )"
        )

    body = (
        "// Generated from internal/config/settings_schema.json.\n"
        "// Regenerate with: python3 tools/settings-swift/generate.py\n"
        "\n"
        "import Foundation\n"
        "\n"
        "extension SettingsCatalog {\n"
        "    /// Every setting the backend stores, paired with the wording the\n"
        "    /// previous interface used for it.\n"
        "    static let generated: [SettingDefinition] = [\n"
        + ",\n".join(entries)
        + "\n    ]\n"
        "}\n"
    )

    OUTPUT.write_text(body)
    print(f"wrote {OUTPUT.relative_to(ROOT)} with {len(entries)} settings")
    return 0


if __name__ == "__main__":
    sys.exit(main())
