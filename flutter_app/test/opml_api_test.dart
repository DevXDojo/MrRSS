import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mrrss_flutter/src/api/api_config.dart';
import 'package:mrrss_flutter/src/api/api_exception.dart';
import 'package:mrrss_flutter/src/api/opml_api.dart';

void main() {
  test('exports OPML through backend', () async {
    final api = OpmlApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        expect(request.method, 'GET');
        expect(request.url.path, '/api/opml/export');
        return http.Response(
          '<?xml version="1.0"?><opml version="2.0"></opml>',
          200,
          headers: {'Content-Type': 'text/xml'},
        );
      }),
    );

    final body = await api.exportOpml();

    expect(body, contains('<opml'));
  });

  test('imports OPML text through backend', () async {
    const opml = '<?xml version="1.0"?><opml version="2.0"></opml>';
    final api = OpmlApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        expect(request.method, 'POST');
        expect(request.url.path, '/api/opml/import');
        expect(request.headers['Content-Type'], startsWith('text/xml'));
        expect(request.body, opml);
        return http.Response('', 200);
      }),
    );

    await api.importOpmlText(opml);
  });

  test('throws ApiException for non-success responses', () async {
    final api = OpmlApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        return http.Response('bad opml', 400);
      }),
    );

    expect(api.exportOpml(), throwsA(isA<ApiException>()));
    expect(api.importOpmlText('<opml />'), throwsA(isA<ApiException>()));
  });
}
