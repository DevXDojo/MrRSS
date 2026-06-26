import 'package:flutter_test/flutter_test.dart';
import 'package:mrrss_flutter/src/models/translation_result.dart';

void main() {
  test('parses translation JSON contract', () {
    final result = TranslationResult.fromJson({
      'translated_text': '[DE] Hello',
      'html': '<p>[DE] Hello</p>',
      'skipped': 'false',
    });

    expect(result.translatedText, '[DE] Hello');
    expect(result.html, '<p>[DE] Hello</p>');
    expect(result.skipped, isFalse);
  });

  test('parses skipped translation JSON contract', () {
    final result = TranslationResult.fromJson({
      'translated_text': 'Hallo',
      'html': '<p>Hallo</p>',
      'skipped': 'true',
      'reason': 'already_target_language',
    });

    expect(result.skipped, isTrue);
    expect(result.reason, 'already_target_language');
  });
}
