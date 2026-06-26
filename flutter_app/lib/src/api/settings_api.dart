import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/app_settings.dart';
import 'api_config.dart';
import 'api_exception.dart';

class SettingsApi {
  SettingsApi({
    required ApiConfig config,
    http.Client? client,
  })  : _config = config,
        _client = client ?? http.Client();

  final ApiConfig _config;
  final http.Client _client;

  Future<AppSettings> getSettings() async {
    final response = await _client.get(_config.resolve('/api/settings'));
    if (response.statusCode != 200) {
      throw ApiException(
        message: 'Failed to load settings',
        statusCode: response.statusCode,
        body: response.body,
      );
    }

    final body = jsonDecode(response.body) as Map<String, dynamic>;
    return AppSettings.fromJson(body);
  }

  Future<AppSettings> updateSettings(Map<String, String> settings) async {
    final response = await _client.post(
      _config.resolve('/api/settings'),
      headers: const {'Content-Type': 'application/json'},
      body: jsonEncode(settings),
    );
    if (response.statusCode != 200) {
      throw ApiException(
        message: 'Failed to update settings',
        statusCode: response.statusCode,
        body: response.body,
      );
    }

    final body = jsonDecode(response.body) as Map<String, dynamic>;
    return AppSettings.fromJson(body);
  }
}
