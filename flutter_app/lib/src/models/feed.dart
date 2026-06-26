class Feed {
  const Feed({
    required this.id,
    required this.title,
    required this.url,
    required this.link,
    required this.description,
    required this.category,
    required this.imageUrl,
    required this.position,
    required this.lastUpdated,
    required this.hideFromTimeline,
    required this.proxyEnabled,
    required this.refreshInterval,
    required this.isImageMode,
    required this.emailAddress,
    required this.tags,
  });

  final int id;
  final String title;
  final String url;
  final String link;
  final String description;
  final String category;
  final String imageUrl;
  final int position;
  final DateTime lastUpdated;
  final bool hideFromTimeline;
  final bool proxyEnabled;
  final int refreshInterval;
  final bool isImageMode;
  final String emailAddress;
  final List<FeedTag> tags;

  factory Feed.fromJson(Map<String, dynamic> json) {
    return Feed(
      id: _requiredInt(json, 'id'),
      title: _requiredString(json, 'title'),
      url: _requiredString(json, 'url'),
      link: _optionalString(json, 'link'),
      description: _optionalString(json, 'description'),
      category: _optionalString(json, 'category'),
      imageUrl: _optionalString(json, 'image_url'),
      position: _requiredInt(json, 'position'),
      lastUpdated: DateTime.parse(_requiredString(json, 'last_updated')),
      hideFromTimeline: _requiredBool(json, 'hide_from_timeline'),
      proxyEnabled: _requiredBool(json, 'proxy_enabled'),
      refreshInterval: _requiredInt(json, 'refresh_interval'),
      isImageMode: _requiredBool(json, 'is_image_mode'),
      emailAddress: _optionalString(json, 'email_address'),
      tags: _optionalList(json, 'tags').map(FeedTag.fromJson).toList(),
    );
  }
}

class FeedTag {
  const FeedTag({
    required this.id,
    required this.name,
    required this.color,
    required this.position,
  });

  final int id;
  final String name;
  final String color;
  final int position;

  factory FeedTag.fromJson(Map<String, dynamic> json) {
    return FeedTag(
      id: _requiredInt(json, 'id'),
      name: _requiredString(json, 'name'),
      color: _requiredString(json, 'color'),
      position: _requiredInt(json, 'position'),
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

List<Map<String, dynamic>> _optionalList(
  Map<String, dynamic> json,
  String key,
) {
  final value = json[key];
  if (value == null) {
    return const [];
  }
  if (value is! List<dynamic>) {
    throw FormatException('$key must be a list when present');
  }
  return value.cast<Map<String, dynamic>>();
}
