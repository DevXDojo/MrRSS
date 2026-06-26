import 'package:flutter_test/flutter_test.dart';
import 'package:mrrss_flutter/src/models/app_settings.dart';

void main() {
  test('parses settings JSON contract used by /api/settings', () {
    final settings = AppSettings.fromJson({
      'language': 'zh-CN',
      'theme': 'dark',
      'update_interval': '45',
      'refresh_mode': 'intelligent',
      'default_view_mode': 'webpage',
      'translation_enabled': 'true',
      'translation_only_mode': 'false',
      'target_language': 'en',
      'translation_provider': 'deepl',
      'deepl_api_key': 'secret-key',
      'deepl_endpoint': 'https://api-free.deepl.com/v2/translate',
      'baidu_app_id': 'baidu-app',
      'baidu_secret_key': 'baidu-secret',
      'microsoft_api_key': 'ms-key',
      'microsoft_endpoint': 'https://api.cognitive.microsofttranslator.com',
      'microsoft_region': 'eastasia',
      'tencent_secret_id': 'tencent-id',
      'tencent_secret_key': 'tencent-key',
      'tencent_region': 'ap-shanghai',
      'ai_translation_profile_id': 'profile-1',
      'ai_translation_prompt': 'Translate precisely.',
      'custom_translation_enabled': 'true',
      'custom_translation_name': 'Acme Translate',
      'custom_translation_endpoint': 'https://translate.example.com',
      'custom_translation_method': 'POST',
      'custom_translation_headers': '{"X-API-Key":"secret"}',
      'custom_translation_body_template': '{"text":"{{text}}"}',
      'custom_translation_response_path': 'data.text',
      'custom_translation_lang_mapping': '{"zh":"zh-CN","en":"en-US"}',
      'custom_translation_timeout': '15',
    });

    expect(settings.language, 'zh-CN');
    expect(settings.theme, 'dark');
    expect(settings.updateInterval, 45);
    expect(settings.refreshMode, 'intelligent');
    expect(settings.defaultViewMode, 'webpage');
    expect(settings.translationEnabled, isTrue);
    expect(settings.translationOnlyMode, isFalse);
    expect(settings.targetLanguage, 'en');
    expect(settings.translationProvider, 'deepl');
    expect(settings.deeplApiKey, 'secret-key');
    expect(settings.deeplEndpoint, 'https://api-free.deepl.com/v2/translate');
    expect(settings.baiduAppId, 'baidu-app');
    expect(settings.baiduSecretKey, 'baidu-secret');
    expect(settings.microsoftApiKey, 'ms-key');
    expect(
      settings.microsoftEndpoint,
      'https://api.cognitive.microsofttranslator.com',
    );
    expect(settings.microsoftRegion, 'eastasia');
    expect(settings.tencentSecretId, 'tencent-id');
    expect(settings.tencentSecretKey, 'tencent-key');
    expect(settings.tencentRegion, 'ap-shanghai');
    expect(settings.aiTranslationProfileId, 'profile-1');
    expect(settings.aiTranslationPrompt, 'Translate precisely.');
    expect(settings.customTranslationEnabled, isTrue);
    expect(settings.customTranslationName, 'Acme Translate');
    expect(settings.customTranslationEndpoint, 'https://translate.example.com');
    expect(settings.customTranslationMethod, 'POST');
    expect(settings.customTranslationHeaders, '{"X-API-Key":"secret"}');
    expect(settings.customTranslationBodyTemplate, '{"text":"{{text}}"}');
    expect(settings.customTranslationResponsePath, 'data.text');
    expect(
      settings.customTranslationLangMapping,
      '{"zh":"zh-CN","en":"en-US"}',
    );
    expect(settings.customTranslationTimeout, 15);
    expect(settings.raw['deepl_api_key'], 'secret-key');
  });

  test('rejects non-string settings values', () {
    expect(
      () => AppSettings.fromJson({
        'language': 'en-US',
        'theme': 'auto',
        'update_interval': 30,
        'refresh_mode': 'fixed',
        'default_view_mode': 'rendered',
        'translation_enabled': 'false',
        'translation_only_mode': 'false',
        'target_language': 'zh',
        'translation_provider': 'google',
      }),
      throwsFormatException,
    );
  });

  test('rejects invalid bool strings', () {
    expect(
      () => AppSettings.fromJson({
        'language': 'en-US',
        'theme': 'auto',
        'update_interval': '30',
        'refresh_mode': 'fixed',
        'default_view_mode': 'rendered',
        'translation_enabled': 'yes',
        'translation_only_mode': 'false',
        'target_language': 'zh',
        'translation_provider': 'google',
      }),
      throwsFormatException,
    );
  });
}
