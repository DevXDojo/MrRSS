import 'package:http/http.dart' as http;

import 'api_config.dart';
import 'api_exception.dart';

class OpmlApi {
  OpmlApi({
    required ApiConfig config,
    http.Client? client,
  })  : _config = config,
        _client = client ?? http.Client();

  final ApiConfig _config;
  final http.Client _client;

  Future<String> exportOpml() async {
    final response = await _client.get(_config.resolve('/api/opml/export'));
    if (response.statusCode != 200) {
      throw ApiException(
        message: 'Failed to export OPML',
        statusCode: response.statusCode,
        body: response.body,
      );
    }

    return response.body;
  }

  Future<void> importOpmlText(String opmlText) async {
    final response = await _client.post(
      _config.resolve('/api/opml/import'),
      headers: const {'Content-Type': 'text/xml'},
      body: opmlText,
    );
    if (response.statusCode != 200) {
      throw ApiException(
        message: 'Failed to import OPML',
        statusCode: response.statusCode,
        body: response.body,
      );
    }
  }
}
