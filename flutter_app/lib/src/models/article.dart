class Article {
  const Article({
    required this.id,
    required this.feedId,
    required this.title,
    required this.url,
    required this.imageUrl,
    required this.audioUrl,
    required this.videoUrl,
    required this.publishedAt,
    required this.isRead,
    required this.isFavorite,
    required this.isHidden,
    required this.isReadLater,
    required this.feedTitle,
    required this.author,
    required this.translatedTitle,
    required this.summary,
    required this.freshRssItemId,
  });

  final int id;
  final int feedId;
  final String title;
  final String url;
  final String imageUrl;
  final String audioUrl;
  final String videoUrl;
  final DateTime publishedAt;
  final bool isRead;
  final bool isFavorite;
  final bool isHidden;
  final bool isReadLater;
  final String feedTitle;
  final String author;
  final String translatedTitle;
  final String summary;
  final String freshRssItemId;

  factory Article.fromJson(Map<String, dynamic> json) {
    return Article(
      id: _requiredInt(json, 'id'),
      feedId: _requiredInt(json, 'feed_id'),
      title: _requiredString(json, 'title'),
      url: _requiredString(json, 'url'),
      imageUrl: _optionalString(json, 'image_url'),
      audioUrl: _optionalString(json, 'audio_url'),
      videoUrl: _optionalString(json, 'video_url'),
      publishedAt: DateTime.parse(_requiredString(json, 'published_at')),
      isRead: _requiredBool(json, 'is_read'),
      isFavorite: _requiredBool(json, 'is_favorite'),
      isHidden: _requiredBool(json, 'is_hidden'),
      isReadLater: _requiredBool(json, 'is_read_later'),
      feedTitle: _optionalString(json, 'feed_title'),
      author: _optionalString(json, 'author'),
      translatedTitle: _optionalString(json, 'translated_title'),
      summary: _optionalString(json, 'summary'),
      freshRssItemId: _optionalString(json, 'freshrss_item_id'),
    );
  }
}

String _requiredString(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is! String || value.isEmpty) {
    throw FormatException('$key must be a non-empty string');
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

int _requiredInt(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is int) {
    return value;
  }
  throw FormatException('$key must be an integer');
}

bool _requiredBool(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is bool) {
    return value;
  }
  throw FormatException('$key must be a bool');
}
