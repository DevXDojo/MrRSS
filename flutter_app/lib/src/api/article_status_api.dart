import 'package:http/http.dart' as http;

import 'api_config.dart';
import 'api_exception.dart';

class ArticleStatusApi {
  ArticleStatusApi({
    required ApiConfig config,
    http.Client? client,
  })  : _config = config,
        _client = client ?? http.Client();

  final ApiConfig _config;
  final http.Client _client;

  Future<void> markRead({
    required int articleId,
    required bool read,
  }) async {
    final response = await _client.post(
      _config.resolve('/api/articles/read', {
        'id': articleId.toString(),
        'read': read.toString(),
      }),
    );
    if (response.statusCode != 200) {
      throw ApiException(
        message: 'Failed to update article read state',
        statusCode: response.statusCode,
        body: response.body,
      );
    }
  }

  Future<void> toggleFavorite(int articleId) async {
    final response = await _client.post(
      _config.resolve('/api/articles/favorite', {'id': articleId.toString()}),
    );
    if (response.statusCode != 200) {
      throw ApiException(
        message: 'Failed to toggle article favorite state',
        statusCode: response.statusCode,
        body: response.body,
      );
    }
  }
}
