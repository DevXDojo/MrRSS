import 'package:flutter_test/flutter_test.dart';
import 'package:mrrss_flutter/src/models/feed.dart';

void main() {
  test('parses feed JSON contract used by /api/feeds', () {
    final feed = Feed.fromJson({
      'id': 12,
      'title': 'Contract Feed',
      'url': 'https://example.com/feed.xml',
      'link': 'https://example.com',
      'description': 'A feed',
      'category': 'Tech',
      'image_url': 'https://example.com/icon.png',
      'position': 7,
      'last_updated': '2026-06-25T09:30:00Z',
      'hide_from_timeline': true,
      'proxy_enabled': true,
      'refresh_interval': 30,
      'is_image_mode': true,
      'email_address': 'newsletter@example.com',
      'tags': [
        {
          'id': 3,
          'name': 'Important',
          'color': '#2563eb',
          'position': 1,
        },
      ],
    });

    expect(feed.id, 12);
    expect(feed.title, 'Contract Feed');
    expect(feed.category, 'Tech');
    expect(feed.imageUrl, 'https://example.com/icon.png');
    expect(feed.hideFromTimeline, isTrue);
    expect(feed.proxyEnabled, isTrue);
    expect(feed.refreshInterval, 30);
    expect(feed.isImageMode, isTrue);
    expect(feed.emailAddress, 'newsletter@example.com');
    expect(
      feed.lastUpdated.toUtc().toIso8601String(),
      '2026-06-25T09:30:00.000Z',
    );
    expect(feed.tags.single.name, 'Important');
  });

  test('rejects feed JSON without required list fields', () {
    expect(
      () => Feed.fromJson({
        'id': 12,
        'title': 'Contract Feed',
        'url': 'https://example.com/feed.xml',
        'last_updated': '2026-06-25T09:30:00Z',
      }),
      throwsFormatException,
    );
  });
}
