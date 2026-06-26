import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/translation_result.dart';
import 'api_config.dart';
import 'api_exception.dart';

class TranslationApi {
  TranslationApi({
    required ApiConfig config,
    http.Client? client,
  })  : _config = config,
        _client = client ?? http.Client();

  final ApiConfig _config;
  final http.Client _client;

  Future<TranslationResult> translateText({
    required String text,
    required String targetLanguage,
    bool force = false,
  }) async {
    final response = await _client.post(
      _config.resolve('/api/articles/translate-text'),
      headers: const {'Content-Type': 'application/json'},
      body: jsonEncode({
        'text': text,
        'target_language': targetLanguage,
        'force': force,
      }),
    );
    if (response.statusCode != 200) {
      throw ApiException(
        message: 'Failed to translate text',
        statusCode: response.statusCode,
        body: response.body,
      );
    }

    final body = jsonDecode(response.body) as Map<String, dynamic>;
    return TranslationResult.fromJson(body);
  }

  Future<void> clearTranslations() async {
    final response = await _client.post(
      _config.resolve('/api/articles/clear-translations'),
    );
    if (response.statusCode != 200) {
      throw ApiException(
        message: 'Failed to clear translations',
        statusCode: response.statusCode,
        body: response.body,
      );
    }

    final body = jsonDecode(response.body) as Map<String, dynamic>;
    if (body['success'] != true) {
      throw const FormatException(
        'clear translations response missing success',
      );
    }
  }
}
