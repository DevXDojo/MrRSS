import 'dart:convert';

import 'package:http/http.dart' as http;

import 'api_config.dart';
import 'api_exception.dart';

class FeedActionsApi {
  FeedActionsApi({
    required ApiConfig config,
    http.Client? client,
  })  : _config = config,
        _client = client ?? http.Client();

  final ApiConfig _config;
  final http.Client _client;

  Future<String> refreshFeed(int feedId) async {
    final response = await _client.post(
      _config.resolve('/api/feeds/refresh', {'id': feedId.toString()}),
    );
    if (response.statusCode != 200) {
      throw ApiException(
        message: 'Failed to refresh feed',
        statusCode: response.statusCode,
        body: response.body,
      );
    }

    final body = jsonDecode(response.body) as Map<String, dynamic>;
    final status = body['status'];
    if (status is! String || status.isEmpty) {
      throw const FormatException(
        'Feed refresh response requires a status string',
      );
    }
    return status;
  }
}
