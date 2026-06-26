import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mrrss_flutter/src/api/api_config.dart';
import 'package:mrrss_flutter/src/api/api_exception.dart';
import 'package:mrrss_flutter/src/api/articles_api.dart';

void main() {
  test('loads articles from backend with query parameters', () async {
    final api = ArticlesApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        expect(request.url.path, '/api/articles');
        expect(request.url.queryParameters['filter'], 'all');
        expect(request.url.queryParameters['feed_id'], '12');
        expect(request.url.queryParameters['page'], '2');
        expect(request.url.queryParameters['limit'], '20');
        return http.Response(
          '''
[
  {
    "id": 42,
    "feed_id": 12,
    "title": "Contract Article",
    "url": "https://example.com/article",
    "image_url": "",
    "audio_url": "",
    "video_url": "",
    "published_at": "2026-06-25T09:45:00Z",
    "is_read": false,
    "is_favorite": false,
    "is_hidden": false,
    "is_read_later": false,
    "feed_title": "Contract Feed",
    "translated_title": "",
    "summary": "",
    "freshrss_item_id": ""
  }
]
''',
          200,
        );
      }),
    );

    final articles = await api.listArticles(
      filter: 'all',
      feedId: 12,
      page: 2,
      limit: 20,
    );

    expect(articles, hasLength(1));
    expect(articles.single.title, 'Contract Article');
    expect(articles.single.feedTitle, 'Contract Feed');
  });

  test('throws ApiException for non-success responses', () async {
    final api = ArticlesApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        return http.Response('server error', 500);
      }),
    );

    expect(api.listArticles(), throwsA(isA<ApiException>()));
  });
}
