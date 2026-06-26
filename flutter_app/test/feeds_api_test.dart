import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mrrss_flutter/src/api/api_config.dart';
import 'package:mrrss_flutter/src/api/api_exception.dart';
import 'package:mrrss_flutter/src/api/feeds_api.dart';

void main() {
  test('loads feeds from backend', () async {
    final api = FeedsApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        expect(request.url.path, '/api/feeds');
        return http.Response(
          '''
[
  {
    "id": 12,
    "title": "Contract Feed",
    "url": "https://example.com/feed.xml",
    "link": "https://example.com",
    "description": "A feed",
    "category": "Tech",
    "image_url": "https://example.com/icon.png",
    "position": 7,
    "last_updated": "2026-06-25T09:30:00Z",
    "hide_from_timeline": false,
    "proxy_enabled": false,
    "refresh_interval": 0,
    "is_image_mode": false,
    "email_address": "",
    "tags": []
  }
]
''',
          200,
        );
      }),
    );

    final feeds = await api.listFeeds();

    expect(feeds, hasLength(1));
    expect(feeds.single.title, 'Contract Feed');
  });

  test('throws ApiException for non-success responses', () async {
    final api = FeedsApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        return http.Response('server error', 500);
      }),
    );

    expect(api.listFeeds(), throwsA(isA<ApiException>()));
  });
}
