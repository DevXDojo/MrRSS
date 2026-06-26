String htmlToReadableText(String html) {
  var text = html;

  text = text.replaceAll(
    RegExp(r'<script[^>]*>.*?</script>', caseSensitive: false, dotAll: true),
    ' ',
  );
  text = text.replaceAll(
    RegExp(r'<style[^>]*>.*?</style>', caseSensitive: false, dotAll: true),
    ' ',
  );
  text = text.replaceAll(RegExp(r'<br\s*/?>', caseSensitive: false), '\n');
  text = text.replaceAll(RegExp(r'</p\s*>', caseSensitive: false), '\n\n');
  text = text.replaceAll(RegExp(r'</h[1-6]\s*>', caseSensitive: false), '\n\n');
  text = text.replaceAll(RegExp(r'</li\s*>', caseSensitive: false), '\n');
  text = text.replaceAll(RegExp(r'<li[^>]*>', caseSensitive: false), '- ');
  text = text.replaceAll(
    RegExp(r'</(div|section|article|blockquote)\s*>', caseSensitive: false),
    '\n\n',
  );
  text = text.replaceAll(RegExp(r'<[^>]+>'), ' ');
  text = _decodeHtmlEntities(text);
  text = text.replaceAll(RegExp(r'[ \t\f\r]+'), ' ');
  text = text.replaceAll(RegExp(r' *\n *'), '\n');
  text = text.replaceAll(RegExp(r'\n{3,}'), '\n\n');

  return text.trim();
}

String _decodeHtmlEntities(String text) {
  return text
      .replaceAll('&nbsp;', ' ')
      .replaceAll('&amp;', '&')
      .replaceAll('&lt;', '<')
      .replaceAll('&gt;', '>')
      .replaceAll('&quot;', '"')
      .replaceAll('&#39;', "'")
      .replaceAll('&apos;', "'");
}
