class ApiConfig {
  const ApiConfig({
    required this.baseUrl,
  });

  final Uri baseUrl;

  Uri resolve(String path, [Map<String, String>? queryParameters]) {
    final normalizedPath = path.startsWith('/') ? path : '/$path';
    return baseUrl.replace(
      path: '${baseUrl.path}$normalizedPath',
      queryParameters: queryParameters,
    );
  }

  static ApiConfig local({int port = 1234}) {
    return ApiConfig(baseUrl: Uri.parse('http://localhost:$port'));
  }
}
