import 'package:flutter_test/flutter_test.dart';
import 'package:mrrss_flutter/src/utils/html_text.dart';

void main() {
  test('converts common article HTML to readable text', () {
    final text = htmlToReadableText(
      '<h1>Title</h1><p>Hello <strong>Flutter</strong> &amp; RSS.</p><ul><li>One</li><li>Two</li></ul>',
    );

    expect(text, 'Title\n\nHello Flutter & RSS.\n\n- One\n- Two');
  });

  test('removes script and style blocks', () {
    final text = htmlToReadableText(
      '<style>.x{}</style><p>Visible</p><script>alert(1)</script>',
    );

    expect(text, 'Visible');
  });

  test('returns empty string for tag-only content', () {
    expect(htmlToReadableText('<div><span></span></div>'), '');
  });
}
