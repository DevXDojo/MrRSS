import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mrrss_flutter/src/api/api_config.dart';
import 'package:mrrss_flutter/src/api/api_exception.dart';
import 'package:mrrss_flutter/src/api/article_content_api.dart';

void main() {
  test('loads article content from backend', () async {
    final api = ArticleContentApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        expect(request.url.path, '/api/articles/content');
        expect(request.url.queryParameters['id'], '42');
        return http.Response(
          '{"content":"<p>Hello reader.</p>","feed_url":"https://example.com/feed.xml","cached":true}',
          200,
        );
      }),
    );

    final content = await api.getContent(42);

    expect(content.content, '<p>Hello reader.</p>');
    expect(content.cached, isTrue);
  });

  test('throws ApiException for non-success responses', () async {
    final api = ArticleContentApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        return http.Response('server error', 500);
      }),
    );

    expect(api.getContent(42), throwsA(isA<ApiException>()));
  });
}
