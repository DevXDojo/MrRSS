import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/feed.dart';
import 'api_config.dart';
import 'api_exception.dart';

class FeedsApi {
  FeedsApi({
    required ApiConfig config,
    http.Client? client,
  })  : _config = config,
        _client = client ?? http.Client();

  final ApiConfig _config;
  final http.Client _client;

  Future<List<Feed>> listFeeds() async {
    final response = await _client.get(_config.resolve('/api/feeds'));
    if (response.statusCode != 200) {
      throw ApiException(
        message: 'Failed to load feeds',
        statusCode: response.statusCode,
        body: response.body,
      );
    }

    final body = jsonDecode(response.body) as List<dynamic>;
    return body
        .cast<Map<String, dynamic>>()
        .map(Feed.fromJson)
        .toList(growable: false);
  }
}
