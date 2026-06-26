import 'package:flutter_test/flutter_test.dart';
import 'package:mrrss_flutter/src/models/article_content.dart';

void main() {
  test('parses article content JSON contract', () {
    final content = ArticleContent.fromJson({
      'content': '<p>Hello reader.</p>',
      'feed_url': 'https://example.com/feed.xml',
      'cached': true,
    });

    expect(content.content, '<p>Hello reader.</p>');
    expect(content.feedUrl, 'https://example.com/feed.xml');
    expect(content.cached, isTrue);
  });

  test('rejects content JSON without cached flag', () {
    expect(
      () => ArticleContent.fromJson({
        'content': '<p>Hello reader.</p>',
        'feed_url': 'https://example.com/feed.xml',
      }),
      throwsFormatException,
    );
  });
}
