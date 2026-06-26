import 'package:flutter/material.dart';

import '../api/api_config.dart';
import '../platform/opml_file_service.dart';
import '../reader/http_reader_repository.dart';
import '../reader/reader_repository.dart';
import '../reader/reader_screen.dart';

class MrRssApp extends StatelessWidget {
  const MrRssApp({
    super.key,
    this.readerRepository,
    this.opmlFileService = const FilePickerOpmlFileService(),
  });

  final ReaderRepository? readerRepository;
  final OpmlFileService opmlFileService;

  @override
  Widget build(BuildContext context) {
    final repository = readerRepository ??
        HttpReaderRepository(
          config: ApiConfig.local(),
        );

    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'MrRSS',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF2563EB)),
        useMaterial3: true,
      ),
      darkTheme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF60A5FA),
          brightness: Brightness.dark,
        ),
        useMaterial3: true,
      ),
      home: ReaderScreen(
        repository: repository,
        opmlFileService: opmlFileService,
      ),
    );
  }
}
