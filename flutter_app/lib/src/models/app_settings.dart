class AppSettings {
  const AppSettings({
    required this.language,
    required this.theme,
    required this.updateInterval,
    required this.refreshMode,
    required this.defaultViewMode,
    required this.translationEnabled,
    required this.translationOnlyMode,
    required this.targetLanguage,
    required this.translationProvider,
    required this.deeplApiKey,
    required this.deeplEndpoint,
    required this.baiduAppId,
    required this.baiduSecretKey,
    required this.microsoftApiKey,
    required this.microsoftEndpoint,
    required this.microsoftRegion,
    required this.tencentSecretId,
    required this.tencentSecretKey,
    required this.tencentRegion,
    required this.aiTranslationProfileId,
    required this.aiTranslationPrompt,
    required this.customTranslationEnabled,
    required this.customTranslationName,
    required this.customTranslationEndpoint,
    required this.customTranslationMethod,
    required this.customTranslationHeaders,
    required this.customTranslationBodyTemplate,
    required this.customTranslationResponsePath,
    required this.customTranslationLangMapping,
    required this.customTranslationTimeout,
    required this.raw,
  });

  final String language;
  final String theme;
  final int updateInterval;
  final String refreshMode;
  final String defaultViewMode;
  final bool translationEnabled;
  final bool translationOnlyMode;
  final String targetLanguage;
  final String translationProvider;
  final String deeplApiKey;
  final String deeplEndpoint;
  final String baiduAppId;
  final String baiduSecretKey;
  final String microsoftApiKey;
  final String microsoftEndpoint;
  final String microsoftRegion;
  final String tencentSecretId;
  final String tencentSecretKey;
  final String tencentRegion;
  final String aiTranslationProfileId;
  final String aiTranslationPrompt;
  final bool customTranslationEnabled;
  final String customTranslationName;
  final String customTranslationEndpoint;
  final String customTranslationMethod;
  final String customTranslationHeaders;
  final String customTranslationBodyTemplate;
  final String customTranslationResponsePath;
  final String customTranslationLangMapping;
  final int customTranslationTimeout;
  final Map<String, String> raw;

  factory AppSettings.fromJson(Map<String, dynamic> json) {
    final raw = json.map((key, value) {
      if (value is! String) {
        throw FormatException('$key must be a string');
      }
      return MapEntry(key, value);
    });

    return AppSettings(
      language: _requiredString(raw, 'language'),
      theme: _requiredString(raw, 'theme'),
      updateInterval: _requiredInt(raw, 'update_interval'),
      refreshMode: _requiredString(raw, 'refresh_mode'),
      defaultViewMode: _requiredString(raw, 'default_view_mode'),
      translationEnabled: _requiredBool(raw, 'translation_enabled'),
      translationOnlyMode: _requiredBool(raw, 'translation_only_mode'),
      targetLanguage: _requiredString(raw, 'target_language'),
      translationProvider: _requiredString(raw, 'translation_provider'),
      deeplApiKey: raw['deepl_api_key'] ?? '',
      deeplEndpoint: raw['deepl_endpoint'] ?? '',
      baiduAppId: raw['baidu_app_id'] ?? '',
      baiduSecretKey: raw['baidu_secret_key'] ?? '',
      microsoftApiKey: raw['microsoft_api_key'] ?? '',
      microsoftEndpoint: raw['microsoft_endpoint'] ?? '',
      microsoftRegion: raw['microsoft_region'] ?? '',
      tencentSecretId: raw['tencent_secret_id'] ?? '',
      tencentSecretKey: raw['tencent_secret_key'] ?? '',
      tencentRegion: raw['tencent_region'] ?? 'ap-guangzhou',
      aiTranslationProfileId: raw['ai_translation_profile_id'] ?? '',
      aiTranslationPrompt: raw['ai_translation_prompt'] ?? '',
      customTranslationEnabled: _optionalBool(
        raw,
        'custom_translation_enabled',
        defaultValue: false,
      ),
      customTranslationName: raw['custom_translation_name'] ?? '',
      customTranslationEndpoint: raw['custom_translation_endpoint'] ?? '',
      customTranslationMethod: raw['custom_translation_method'] ?? 'POST',
      customTranslationHeaders: raw['custom_translation_headers'] ?? '',
      customTranslationBodyTemplate:
          raw['custom_translation_body_template'] ?? '',
      customTranslationResponsePath:
          raw['custom_translation_response_path'] ?? '',
      customTranslationLangMapping:
          raw['custom_translation_lang_mapping'] ?? '',
      customTranslationTimeout: _optionalInt(
        raw,
        'custom_translation_timeout',
        defaultValue: 10,
      ),
      raw: Map.unmodifiable(raw),
    );
  }
}

String _requiredString(Map<String, String> settings, String key) {
  final value = settings[key];
  if (value == null || value.isEmpty) {
    throw FormatException('$key must be a non-empty string');
  }
  return value;
}

int _requiredInt(Map<String, String> settings, String key) {
  final value = _requiredString(settings, key);
  final parsed = int.tryParse(value);
  if (parsed == null) {
    throw FormatException('$key must be an integer string');
  }
  return parsed;
}

bool _requiredBool(Map<String, String> settings, String key) {
  final value = _requiredString(settings, key);
  if (value == 'true') {
    return true;
  }
  if (value == 'false') {
    return false;
  }
  throw FormatException('$key must be a bool string');
}

int _optionalInt(
  Map<String, String> settings,
  String key, {
  required int defaultValue,
}) {
  final value = settings[key];
  if (value == null || value.isEmpty) {
    return defaultValue;
  }
  final parsed = int.tryParse(value);
  if (parsed == null) {
    throw FormatException('$key must be an integer string');
  }
  return parsed;
}

bool _optionalBool(
  Map<String, String> settings,
  String key, {
  required bool defaultValue,
}) {
  final value = settings[key];
  if (value == null || value.isEmpty) {
    return defaultValue;
  }
  if (value == 'true') {
    return true;
  }
  if (value == 'false') {
    return false;
  }
  throw FormatException('$key must be a bool string');
}
