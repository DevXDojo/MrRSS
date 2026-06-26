import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:mrrss_flutter/src/api/api_config.dart';
import 'package:mrrss_flutter/src/api/api_exception.dart';
import 'package:mrrss_flutter/src/api/version_api.dart';
import 'package:mrrss_flutter/src/models/version_info.dart';

void main() {
  test('parses version response', () {
    final info = VersionInfo.fromJson({'version': '1.3.23'});

    expect(info.version, '1.3.23');
  });

  test('rejects malformed version response', () {
    expect(
      () => VersionInfo.fromJson({'version': ''}),
      throwsFormatException,
    );
  });

  test('loads version from backend', () async {
    final api = VersionApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        expect(request.url.path, '/api/version');
        return http.Response('{"version":"1.3.23"}', 200);
      }),
    );

    final info = await api.getVersion();

    expect(info.version, '1.3.23');
  });

  test('throws ApiException for non-success responses', () async {
    final api = VersionApi(
      config: ApiConfig.local(),
      client: MockClient((request) async {
        return http.Response('not found', 404);
      }),
    );

    expect(api.getVersion(), throwsA(isA<ApiException>()));
  });
}
