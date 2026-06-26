import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/version_info.dart';
import 'api_config.dart';
import 'api_exception.dart';

class VersionApi {
  VersionApi({
    required ApiConfig config,
    http.Client? client,
  })  : _config = config,
        _client = client ?? http.Client();

  final ApiConfig _config;
  final http.Client _client;

  Future<VersionInfo> getVersion() async {
    final response = await _client.get(_config.resolve('/api/version'));
    if (response.statusCode != 200) {
      throw ApiException(
        message: 'Failed to load application version',
        statusCode: response.statusCode,
        body: response.body,
      );
    }

    final body = jsonDecode(response.body) as Map<String, dynamic>;
    return VersionInfo.fromJson(body);
  }
}
