class VersionInfo {
  const VersionInfo({
    required this.version,
  });

  final String version;

  factory VersionInfo.fromJson(Map<String, dynamic> json) {
    final version = json['version'];
    if (version is! String || version.isEmpty) {
      throw const FormatException('Version response requires a version string');
    }

    return VersionInfo(version: version);
  }
}
