import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mrrss_flutter/src/api/api_config.dart';
import 'package:mrrss_flutter/src/api/api_exception.dart';
import 'package:mrrss_flutter/src/api/article_status_api.dart';

void main() {
  test('marks article read through backend', () async {
    final api = ArticleStatusApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        expect(request.method, 'POST');
        expect(request.url.path, '/api/articles/read');
        expect(request.url.queryParameters['id'], '42');
        expect(request.url.queryParameters['read'], 'true');
        return http.Response('', 200);
      }),
    );

    await api.markRead(articleId: 42, read: true);
  });

  test('toggles favorite through backend', () async {
    final api = ArticleStatusApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        expect(request.method, 'POST');
        expect(request.url.path, '/api/articles/favorite');
        expect(request.url.queryParameters['id'], '42');
        return http.Response('', 200);
      }),
    );

    await api.toggleFavorite(42);
  });

  test('throws ApiException for non-success responses', () async {
    final api = ArticleStatusApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        return http.Response('server error', 500);
      }),
    );

    expect(api.toggleFavorite(42), throwsA(isA<ApiException>()));
  });
}
