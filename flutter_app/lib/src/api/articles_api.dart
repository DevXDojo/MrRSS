import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/article.dart';
import 'api_config.dart';
import 'api_exception.dart';

class ArticlesApi {
  ArticlesApi({
    required ApiConfig config,
    http.Client? client,
  })  : _config = config,
        _client = client ?? http.Client();

  final ApiConfig _config;
  final http.Client _client;

  Future<List<Article>> listArticles({
    String? filter,
    int? feedId,
    String? category,
    int page = 1,
    int limit = 50,
  }) async {
    final query = <String, String>{
      'page': page.toString(),
      'limit': limit.toString(),
      if (filter != null && filter.isNotEmpty) 'filter': filter,
      if (feedId != null) 'feed_id': feedId.toString(),
      if (category != null) 'category': category,
    };

    final response = await _client.get(_config.resolve('/api/articles', query));
    if (response.statusCode != 200) {
      throw ApiException(
        message: 'Failed to load articles',
        statusCode: response.statusCode,
        body: response.body,
      );
    }

    final body = jsonDecode(response.body) as List<dynamic>;
    return body
        .cast<Map<String, dynamic>>()
        .map(Article.fromJson)
        .toList(growable: false);
  }
}
