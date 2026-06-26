class ArticleContent {
  const ArticleContent({
    required this.content,
    required this.feedUrl,
    required this.cached,
  });

  final String content;
  final String feedUrl;
  final bool cached;

  factory ArticleContent.fromJson(Map<String, dynamic> json) {
    return ArticleContent(
      content: _requiredString(json, 'content'),
      feedUrl: _optionalString(json, 'feed_url'),
      cached: _requiredBool(json, 'cached'),
    );
  }
}

String _requiredString(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is! String) {
    throw FormatException('$key must be a string');
  }
  return value;
}

String _optionalString(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value == null) {
    return '';
  }
  if (value is! String) {
    throw FormatException('$key must be a string when present');
  }
  return value;
}

bool _requiredBool(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is bool) {
    return value;
  }
  throw FormatException('$key must be a bool');
}
