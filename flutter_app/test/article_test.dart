import 'package:flutter_test/flutter_test.dart';
import 'package:mrrss_flutter/src/models/article.dart';

void main() {
  test('parses article JSON contract used by /api/articles', () {
    final article = Article.fromJson({
      'id': 42,
      'feed_id': 12,
      'title': 'Contract Article',
      'url': 'https://example.com/article',
      'image_url': 'https://example.com/image.png',
      'audio_url': 'https://example.com/audio.mp3',
      'video_url': 'https://example.com/video.mp4',
      'published_at': '2026-06-25T09:45:00Z',
      'is_read': true,
      'is_favorite': true,
      'is_hidden': false,
      'is_read_later': true,
      'feed_title': 'Contract Feed',
      'author': 'Writer',
      'translated_title': 'Translated Contract Article',
      'summary': 'Short summary',
      'freshrss_item_id': 'fresh-item-1',
    });

    expect(article.id, 42);
    expect(article.feedId, 12);
    expect(article.title, 'Contract Article');
    expect(article.imageUrl, 'https://example.com/image.png');
    expect(article.isRead, isTrue);
    expect(article.isFavorite, isTrue);
    expect(article.isHidden, isFalse);
    expect(article.isReadLater, isTrue);
    expect(article.feedTitle, 'Contract Feed');
    expect(article.author, 'Writer');
    expect(article.translatedTitle, 'Translated Contract Article');
    expect(article.summary, 'Short summary');
    expect(article.freshRssItemId, 'fresh-item-1');
    expect(
      article.publishedAt.toUtc().toIso8601String(),
      '2026-06-25T09:45:00.000Z',
    );
  });

  test('rejects article JSON without required fields', () {
    expect(
      () => Article.fromJson({
        'id': 42,
        'title': 'Contract Article',
        'url': 'https://example.com/article',
      }),
      throwsFormatException,
    );
  });
}
