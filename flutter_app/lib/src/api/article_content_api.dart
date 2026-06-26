import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/article_content.dart';
import 'api_config.dart';
import 'api_exception.dart';

class ArticleContentApi {
  ArticleContentApi({
    required ApiConfig config,
    http.Client? client,
  })  : _config = config,
        _client = client ?? http.Client();

  final ApiConfig _config;
  final http.Client _client;

  Future<ArticleContent> getContent(int articleId) async {
    final response = await _client.get(
      _config.resolve('/api/articles/content', {'id': articleId.toString()}),
    );
    if (response.statusCode != 200) {
      throw ApiException(
        message: 'Failed to load article content',
        statusCode: response.statusCode,
        body: response.body,
      );
    }

    final body = jsonDecode(response.body) as Map<String, dynamic>;
    return ArticleContent.fromJson(body);
  }
}
