import 'package:flutter_test/flutter_test.dart';
import 'package:mrrss_flutter/src/api/api_config.dart';

void main() {
  test('resolves API paths against the configured backend URL', () {
    final config = ApiConfig(baseUrl: Uri.parse('http://localhost:1234'));

    expect(
      config.resolve('/api/version').toString(),
      'http://localhost:1234/api/version',
    );
  });

  test('supports custom local backend ports', () {
    final config = ApiConfig.local(port: 4321);

    expect(
      config.resolve('/api/settings').toString(),
      'http://localhost:4321/api/settings',
    );
  });
}
