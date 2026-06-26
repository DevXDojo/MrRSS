import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';

abstract class OpmlFileService {
  Future<String?> exportText(
    String opmlText, {
    String suggestedFileName = 'mrrss-subscriptions.opml',
  });

  Future<String?> pickText();
}

class FilePickerOpmlFileService implements OpmlFileService {
  const FilePickerOpmlFileService();

  @override
  Future<String?> exportText(
    String opmlText, {
    String suggestedFileName = 'mrrss-subscriptions.opml',
  }) {
    return FilePicker.saveFile(
      dialogTitle: 'Export OPML',
      fileName: suggestedFileName,
      type: FileType.custom,
      allowedExtensions: const ['opml', 'xml'],
      bytes: Uint8List.fromList(utf8.encode(opmlText)),
      lockParentWindow: true,
    );
  }

  @override
  Future<String?> pickText() async {
    final result = await FilePicker.pickFiles(
      dialogTitle: 'Import OPML',
      type: FileType.custom,
      allowedExtensions: const ['opml', 'xml'],
      allowMultiple: false,
      withData: true,
      lockParentWindow: true,
    );
    if (result == null || result.files.isEmpty) {
      return null;
    }

    final file = result.files.single;
    final bytes = file.bytes;
    if (bytes != null) {
      return utf8.decode(bytes);
    }

    final path = file.path;
    if (path != null) {
      return File(path).readAsString();
    }

    throw StateError('Selected OPML file cannot be read.');
  }
}
