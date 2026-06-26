class ApiException implements Exception {
  const ApiException({
    required this.message,
    required this.statusCode,
    this.body,
  });

  final String message;
  final int statusCode;
  final String? body;

  @override
  String toString() {
    return 'ApiException($statusCode): $message';
  }
}
