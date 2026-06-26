import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mrrss_flutter/src/api/api_config.dart';
import 'package:mrrss_flutter/src/api/api_exception.dart';
import 'package:mrrss_flutter/src/api/translation_api.dart';

void main() {
  test('translates text through backend', () async {
    final api = TranslationApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        expect(request.method, 'POST');
        expect(request.url.path, '/api/articles/translate-text');
        expect(request.body, contains('"target_language":"de"'));
        return http.Response(
          '{"translated_text":"[DE] Hello","html":"<p>[DE] Hello</p>","skipped":"false"}',
          200,
        );
      }),
    );

    final result = await api.translateText(
      text: 'Hello',
      targetLanguage: 'de',
      force: true,
    );

    expect(result.translatedText, '[DE] Hello');
  });

  test('throws ApiException for non-success responses', () async {
    final api = TranslationApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        return http.Response('bad request', 400);
      }),
    );

    expect(
      api.translateText(text: 'Hello', targetLanguage: 'de'),
      throwsA(isA<ApiException>()),
    );
  });

  test('clears translation cache through backend', () async {
    final api = TranslationApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        expect(request.method, 'POST');
        expect(request.url.path, '/api/articles/clear-translations');
        return http.Response('{"success":true}', 200);
      }),
    );

    await api.clearTranslations();
  });

  test('throws ApiException when clearing translations fails', () async {
    final api = TranslationApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        return http.Response('server error', 500);
      }),
    );

    expect(api.clearTranslations(), throwsA(isA<ApiException>()));
  });
}
