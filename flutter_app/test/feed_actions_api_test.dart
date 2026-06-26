import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mrrss_flutter/src/api/api_config.dart';
import 'package:mrrss_flutter/src/api/api_exception.dart';
import 'package:mrrss_flutter/src/api/feed_actions_api.dart';

void main() {
  test('refreshes feed through backend', () async {
    final api = FeedActionsApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        expect(request.method, 'POST');
        expect(request.url.path, '/api/feeds/refresh');
        expect(request.url.queryParameters['id'], '7');
        return http.Response('{"status":"refreshing"}', 200);
      }),
    );

    final status = await api.refreshFeed(7);

    expect(status, 'refreshing');
  });

  test('throws ApiException for non-success responses', () async {
    final api = FeedActionsApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        return http.Response('not found', 404);
      }),
    );

    expect(api.refreshFeed(7), throwsA(isA<ApiException>()));
  });
}
