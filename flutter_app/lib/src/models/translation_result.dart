class TranslationResult {
  const TranslationResult({
    required this.translatedText,
    required this.html,
    required this.skipped,
    required this.reason,
  });

  final String translatedText;
  final String html;
  final bool skipped;
  final String reason;

  factory TranslationResult.fromJson(Map<String, dynamic> json) {
    return TranslationResult(
      translatedText: _requiredString(json, 'translated_text'),
      html: _optionalString(json, 'html'),
      skipped: _optionalBoolString(json, 'skipped'),
      reason: _optionalString(json, 'reason'),
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

bool _optionalBoolString(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value == null || value == 'false') {
    return false;
  }
  if (value == 'true') {
    return true;
  }
  throw FormatException('$key must be a bool string when present');
}
