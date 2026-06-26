# MrRSS Flutter App

This directory is the incremental Flutter client migration workspace.

The first checked-in slice contains only cross-platform Flutter/Dart source and tests that can be completed without generated platform folders. When the Flutter SDK is available, run `flutter create --platforms=windows,macos,linux,android,ios .` from this directory to add platform scaffolding without changing the shared app structure.

Current scope:
- Application entry point.
- API configuration.
- Version API client.
- Unit tests for the first API client slice.

Deferred until Flutter CLI is available:
- Generated desktop and mobile platform folders.
- `pubspec.lock`.
- `flutter analyze` and `flutter test` execution.
