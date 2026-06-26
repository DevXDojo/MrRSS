# Flutter Migration Log

Record every migration slice here. Keep entries small and factual so regressions can be traced quickly.

## 2026-06-25: Slice 0 Started

Scope:
- Established Flutter migration plan.
- Captured the current HTTP API route inventory as the initial Flutter API boundary.
- Added a route inventory test to detect drift between `internal/routes` and `docs/flutter/API_ROUTE_INVENTORY.json`.

Environment:
- Current branch: `refactor-flutter`.
- Flutter CLI: not available on PATH, so Flutter workspace generation and Flutter tests are deferred.

Validation:
- Passed: `go test ./internal/routes`

Notes:
- `docs/SERVER_MODE/swagger.json` is not currently a complete source of truth for registered runtime routes. The route inventory is intentionally based on `internal/routes` for this first migration control slice.

## 2026-06-25: Flutter SDK Available Locally

Scope:
- Verified the locally installed Flutter SDK.
- Re-ran Flutter dependency resolution, static analysis, and test suite against the shared Flutter workspace.

Environment:
- Flutter SDK: `E:\flutter\flutter`
- Flutter CLI: available when `E:\flutter\flutter\bin` is prepended to `PATH`.
- Flutter: 3.44.4 stable.
- Dart: 3.12.2.
- Desktop/web targets: Windows, Chrome, and Edge available.
- Android target: not ready because Android SDK is not installed or configured.

Validation:
- Passed: `flutter pub get`
- Passed: `flutter analyze`
- Passed: `flutter test` with 46 tests.
- Passed: `flutter doctor -v` except Android toolchain.

Notes:
- Add `E:\flutter\flutter\bin` to the user or system `PATH` to avoid per-shell path injection.
- Install Android Studio/Android SDK and run `flutter config --android-sdk <sdk-path>` before Android builds can be validated.

## 2026-06-25: Slice 1 Shared Flutter Scaffold Started

Scope:
- Added `flutter_app/` as the shared Flutter client workspace.
- Added `pubspec.yaml`, analysis options, app entry point, minimal Material shell, API configuration, `/api/version` client, and fixture-backed Flutter test files.
- Added `.gitignore` rules for Flutter generated artifacts.
- Added `tools/flutter_migration` Go tests to validate the scaffold and keep the first Flutter API client aligned with `docs/flutter/API_ROUTE_INVENTORY.json`.

Environment:
- Flutter CLI: still not available on PATH.
- Dart CLI: still not available on PATH.

Validation:
- Passed: `go test ./tools/flutter_migration`
- Passed: `go test ./internal/routes`
- Not run: `flutter analyze` because Flutter CLI is unavailable.
- Not run: `flutter test` because Flutter CLI is unavailable.

Notes:
- Generated platform directories are intentionally deferred. Run `flutter create --platforms=windows,macos,linux,android,ios .` inside `flutter_app/` once the SDK is available, then run Flutter native validation before migrating more UI.

## 2026-06-25: Slice 2 Feeds Read API Started

Scope:
- Added a Go JSON contract test for `/api/feeds` covering the first Flutter feed-list fields.
- Verified `/api/feeds` does not expose `email_password`; the field may be absent because the backend model uses `omitempty`.
- Added Flutter `Feed` and `FeedTag` models.
- Added Flutter `FeedsApi` for `GET /api/feeds`.
- Added Flutter test files for feed JSON decoding and `FeedsApi` success/error paths.
- Extended `tools/flutter_migration` tests to keep `FeedsApi` aligned with `docs/flutter/API_ROUTE_INVENTORY.json`.

Environment:
- Flutter CLI: still not available on PATH.
- Dart CLI: still not available on PATH.

Validation:
- Passed: `go test ./internal/handlers/feed`
- Passed: `go test ./tools/flutter_migration ./internal/routes`
- Not run: `flutter analyze` because Flutter CLI is unavailable.
- Not run: `flutter test` because Flutter CLI is unavailable.

## 2026-06-25: Slice 4 Settings Write Added

Scope:
- Extended Flutter `SettingsApi` with `updateSettings`, preserving the existing snake_case string settings contract.
- Added Flutter API test coverage for POST `/api/settings`.
- Extended `ReaderRepository` and `HttpReaderRepository` with `updateSettings`.
- Added a minimal `SettingsSheet` for theme, language, and refresh interval.
- Added a Settings action to the Reader app bar and reloads reader data after save.
- Extended `reader_screen_test.dart` with widget coverage for opening settings and saving bootstrap settings.
- Extended `tools/flutter_migration` tests to require settings write API, settings sheet, repository method, and widget coverage markers.

Environment:
- Flutter CLI: still not available on PATH.
- Dart CLI: still not available on PATH.

Validation:
- Passed: `go test ./internal/handlers/settings`
- Passed: `go test ./tools/flutter_migration ./internal/routes`
- Not run: `flutter analyze` because Flutter CLI is unavailable.
- Not run: `flutter test` because Flutter CLI is unavailable.

## 2026-06-25: Slice 4 Translation Action Added

Scope:
- Added Go JSON contract coverage for `/api/articles/translate-text`.
- Added Flutter `TranslationResult` model.
- Added Flutter `TranslationApi` for text translation requests.
- Extended `ReaderRepository` and `HttpReaderRepository` with `translateText`.
- Added a detail-panel `Translate summary` action using the configured target language.
- Extended Flutter widget coverage for invoking translation from the reader detail panel.
- Extended `tools/flutter_migration` tests to require translation API files, route alignment, repository wiring, and widget coverage markers.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so mobile native builds remain unverified.

Validation:
- Passed: `go test ./internal/handlers/translation`
- Passed: `go test ./internal/handlers/article ./internal/handlers/feed ./internal/handlers/settings ./internal/handlers/translation ./tools/flutter_migration ./internal/routes`
- Passed: `flutter analyze`
- Passed: `flutter test` with 46 tests.

## 2026-06-26: Slice 4 OPML Text Import and Export Added

Scope:
- Added Flutter `OpmlApi` for direct `GET /api/opml/export` and `POST /api/opml/import`.
- Kept this slice platform-neutral by importing/exporting OPML text instead of adding a desktop/mobile file picker.
- Extended `ReaderRepository` and `HttpReaderRepository` with OPML import/export methods.
- Added OPML controls to the Flutter settings sheet for exporting current subscriptions and importing pasted OPML text.
- Added Go contract coverage for `/api/opml/import` raw OPML uploads and `/api/opml/export` text/xml responses on the `/api/*` paths used by Flutter.
- Extended `tools/flutter_migration` tests to require OPML API files, route alignment, repository wiring, and widget coverage markers.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `go test ./internal/handlers/opml ./tools/flutter_migration ./internal/routes`
- Passed: `flutter analyze`
- Passed: `flutter test` with 50 tests.

## 2026-06-26: Slice 4 OPML File Import and Export Added

Scope:
- Added `file_picker` to the Flutter workspace for desktop/mobile OPML file selection.
- Added `OpmlFileService` to isolate platform file dialogs and UTF-8 OPML file reading/writing from UI widgets.
- Extended `MrRssApp`, `ReaderScreen`, and `SettingsSheet` to support exporting OPML to a file and importing OPML from a picked file while preserving the existing text fallback.
- Extended Flutter widget coverage for OPML file export/import using an injected fake file service.
- Extended `tools/flutter_migration` tests to require the OPML file service, dependency declaration, UI controls, and widget coverage markers.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `flutter analyze`
- Passed: `flutter test` with 51 tests.
- Passed: `go test ./tools/flutter_migration ./internal/routes ./internal/handlers/opml`

## 2026-06-26: Slice 1 Generated Platform Directories Added

Scope:
- Ran `flutter create --platforms=windows,linux,android,ios,macos --project-name mrrss_flutter .` inside `flutter_app/`.
- Added generated Flutter platform directories for Windows, Linux, Android, iOS, and macOS.
- Removed the generated counter template widget test because the migrated app already has repository-backed widget coverage.
- Extended `tools/flutter_migration` tests to require representative platform entry files so generated target support is tracked.
- Updated Slice 1 status to reflect that shared source, tests, and generated platform directories are complete.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `flutter pub get`
- Passed: `flutter analyze`
- Passed: `flutter test` with 51 tests.
- Passed: `flutter build windows`
- Passed: `go test ./tools/flutter_migration ./internal/routes ./internal/handlers/article ./internal/handlers/feed ./internal/handlers/settings ./internal/handlers/translation ./internal/handlers/opml`

## 2026-06-26: Slice 4 Article Content Translation Added

Scope:
- Added a `Translate content` action to the Flutter article detail panel.
- Reused the existing `/api/articles/translate-text` API client and repository contract for full readable article text.
- Cached article content loading in `ArticleDetailPanel` state so translation uses the same loaded content and resets cleanly when selecting another article.
- Extended widget coverage to verify content translation requests, rendered translated content, and reset behavior after article selection changes.
- Extended `tools/flutter_migration` checks to require content translation UI and coverage markers.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `flutter analyze`
- Passed: `flutter test` with 52 tests.
- Passed: `go test ./tools/flutter_migration`

## 2026-06-26: Slice 4 Translation Settings Controls Added

Scope:
- Extended Flutter `AppSettings` with `translation_only_mode` and `translation_provider`, alongside existing translation enablement and target language fields.
- Added translation controls to `SettingsSheet`: enable translation, translation-only mode, provider selection, and target language.
- Extended settings save payloads to persist translation settings through the existing `/api/settings` contract.
- Extended Flutter unit/widget tests for settings parsing, API responses, and saving translation settings.
- Extended Go settings handler contract coverage for translation-only mode, provider, and target language persistence.
- Extended `tools/flutter_migration` checks to require translation settings controls and save coverage.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `flutter analyze`
- Passed: `flutter test` with 52 tests.
- Passed: `go test ./tools/flutter_migration ./internal/handlers/settings`

## 2026-06-26: Slice 4 Translation Cache Clear Added

Scope:
- Added Flutter `TranslationApi.clearTranslations` for `POST /api/articles/clear-translations`.
- Extended `ReaderRepository` and `HttpReaderRepository` with translation cache clearing.
- Added a `Clear translation cache` action to the Flutter settings sheet with progress and success states.
- Extended Flutter API and widget tests for cache clearing.
- Extended Go handler coverage to assert the `/api/articles/clear-translations` response contract.
- Extended `tools/flutter_migration` checks to require the API client, repository wiring, and widget coverage markers.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `flutter analyze`
- Passed: `flutter test` with 55 tests.
- Passed: `go test ./tools/flutter_migration ./internal/handlers/translation ./internal/routes`

## 2026-06-26: Slice 4 DeepL Translation Settings Added

Scope:
- Extended Flutter `AppSettings` with `deepl_api_key` and `deepl_endpoint` fields from the existing settings contract.
- Added conditional DeepL API key and endpoint controls to the Flutter settings sheet when the DeepL provider is selected.
- Included DeepL fields in settings save payloads so desktop and mobile Flutter clients can configure the provider.
- Extended Flutter settings model, settings API, and widget tests for DeepL provider configuration.
- Extended Go settings handler coverage for `deepl_endpoint` persistence while preserving encrypted `deepl_api_key` coverage.
- Extended `tools/flutter_migration` checks to require DeepL settings controls and widget coverage markers.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `flutter analyze`
- Passed: `flutter test` with 55 tests.
- Passed: `flutter build windows`
- Passed: `go test ./tools/flutter_migration ./internal/handlers/settings`

## 2026-06-26: Slice 4 Baidu Translation Settings Added

Scope:
- Extended Flutter `AppSettings` with `baidu_app_id` and `baidu_secret_key`.
- Added conditional Baidu app ID and secret key controls to the Flutter settings sheet when the Baidu provider is selected.
- Included Baidu fields in settings save payloads.
- Extended Flutter settings model, API fixture, and widget tests for Baidu provider configuration.
- Extended Go settings handler coverage for Baidu app ID and encrypted secret key persistence.
- Extended `tools/flutter_migration` checks to require Baidu settings controls and widget coverage markers.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `flutter analyze`
- Passed: `flutter test` with 56 tests.
- Passed: `go test ./tools/flutter_migration ./internal/handlers/settings`

## 2026-06-26: Slice 4 Microsoft Translation Settings Added

Scope:
- Extended Flutter `AppSettings` with `microsoft_api_key`, `microsoft_endpoint`, and `microsoft_region`.
- Added conditional Microsoft API key, endpoint, and region controls to the Flutter settings sheet when the Microsoft provider is selected.
- Included Microsoft fields in settings save payloads.
- Extended Flutter settings model, API fixture, and widget tests for Microsoft provider configuration.
- Extended Go settings handler coverage for encrypted Microsoft API key and non-secret endpoint/region persistence.
- Extended `tools/flutter_migration` checks to require Microsoft settings controls and widget coverage markers.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `flutter analyze`
- Passed: `flutter test` with 57 tests.
- Passed: `flutter build windows`
- Passed: `go test ./tools/flutter_migration ./internal/handlers/settings`

## 2026-06-26: Slice 4 Tencent Translation Settings Added

Scope:
- Extended Flutter `AppSettings` with `tencent_secret_id`, `tencent_secret_key`, and `tencent_region`.
- Added conditional Tencent secret ID, secret key, and region controls to the Flutter settings sheet when the Tencent provider is selected.
- Included Tencent fields in settings save payloads.
- Extended Flutter settings model, API fixture, and widget tests for Tencent provider configuration.
- Extended Go settings handler coverage for Tencent secret ID, encrypted secret key, and region persistence.
- Extended `tools/flutter_migration` checks to require Tencent settings controls and widget coverage markers.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `flutter analyze`
- Passed: `flutter test` with 58 tests.
- Passed: `flutter build windows`
- Passed: `go test ./tools/flutter_migration ./internal/handlers/settings`

## 2026-06-26: Slice 4 AI Translation Settings Added

Scope:
- Extended Flutter `AppSettings` with `ai_translation_profile_id` and `ai_translation_prompt`.
- Added conditional AI translation profile ID and prompt controls to the Flutter settings sheet when the AI provider is selected.
- Included AI translation fields in settings save payloads.
- Extended Flutter settings model, API fixture, and widget tests for AI translation provider configuration.
- Extended Go settings handler coverage for AI translation profile and prompt persistence.
- Extended `tools/flutter_migration` checks to require AI translation settings controls and widget coverage markers.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `flutter analyze`
- Passed: `flutter test` with 59 tests.
- Passed: `flutter build windows`
- Passed: `go test ./tools/flutter_migration ./internal/handlers/settings`

## 2026-06-26: Slice 4 Custom Translation Core Settings Added

Scope:
- Extended Flutter `AppSettings` with custom translation enablement, name, endpoint, method, response path, and timeout fields.
- Added conditional custom translation controls to the Flutter settings sheet when the custom provider is selected.
- Included custom translation core fields in settings save payloads.
- Added positive-number validation for custom translation timeout.
- Extended Flutter settings model, API fixture, and widget tests for custom provider core configuration.
- Extended Go settings handler coverage for custom translation core field persistence.
- Extended `tools/flutter_migration` checks to require custom core controls and widget coverage markers.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `flutter analyze`
- Passed: `flutter test` with 60 tests.
- Passed: `flutter build windows`
- Passed: `go test ./tools/flutter_migration ./internal/handlers/settings`

## 2026-06-26: Slice 4 Custom Translation Headers and Body Template Added

Scope:
- Extended Flutter `AppSettings` with `custom_translation_headers` and `custom_translation_body_template`.
- Added custom provider settings controls for headers JSON and POST body templates.
- Included headers and body template fields in settings save payloads.
- Moved the settings save action after provider-specific fields so long provider forms can be saved without scrolling back to the top of the sheet.
- Extended Flutter settings fixtures and widget coverage for custom provider headers/body persistence.
- Extended Go settings handler contract coverage for custom provider headers and body template persistence.
- Extended `tools/flutter_migration` checks to require custom headers/body controls and coverage markers.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `flutter analyze`
- Passed: `flutter test` with 60 tests.
- Passed: `flutter build windows`
- Passed: `go test ./tools/flutter_migration ./internal/handlers/settings`

## 2026-06-26: Slice 4 Custom Translation Language Mapping Added

Scope:
- Extended Flutter `AppSettings` with `custom_translation_lang_mapping`.
- Added a custom provider language-mapping control to the Flutter settings sheet.
- Included the language-mapping field in settings save payloads.
- Extended Flutter settings model, settings API fixture, and widget coverage for custom language mapping persistence.
- Extended Go settings handler contract coverage for custom language mapping.
- Extended `tools/flutter_migration` checks to require the language-mapping setting and UI label.

Environment:
- Flutter CLI available through `E:\flutter\flutter\bin`.
- Android SDK still missing, so Android native builds remain unverified.

Validation:
- Passed: `flutter analyze`
- Passed: `flutter test test\app_settings_test.dart test\settings_api_test.dart test\reader_screen_test.dart --plain-name "saves custom translation provider core settings"`
- Passed: `flutter test` with 60 tests.
- Passed: `flutter build windows`
- Passed: `go test ./tools/flutter_migration ./internal/handlers/settings`
- Passed: `go test ./tools/flutter_migration ./internal/routes ./internal/handlers/article ./internal/handlers/feed ./internal/handlers/settings ./internal/handlers/translation ./internal/handlers/opml`

## 2026-06-25: Slice 2 Settings Read API Added

Scope:
- Added a Go JSON contract test for `/api/settings` covering Flutter bootstrap fields: language, theme, update interval, refresh mode, view mode, translation enabled, and target language.
- Confirmed `/api/settings` returns string values for numeric and boolean settings, matching the existing frontend setting convention.
- Added Flutter `AppSettings` model with strong parsing for integer and boolean string settings.
- Added Flutter `SettingsApi` for `GET /api/settings`.
- Added Flutter test files for settings JSON decoding and `SettingsApi` success/error paths.
- Extended `tools/flutter_migration` tests to keep `SettingsApi` aligned with `docs/flutter/API_ROUTE_INVENTORY.json`.

Environment:
- Flutter CLI: still not available on PATH.
- Dart CLI: still not available on PATH.

Validation:
- Passed: `go test ./internal/handlers/settings`
- Passed: `go test ./tools/flutter_migration ./internal/routes`
- Not run: `flutter analyze` because Flutter CLI is unavailable.
- Not run: `flutter test` because Flutter CLI is unavailable.

## 2026-06-25: Slice 2 Articles List API Added

Scope:
- Added a Go JSON contract test for `/api/articles` covering the first Flutter article-list fields and pagination query usage.
- Confirmed `published_at` is RFC3339-compatible for Flutter `DateTime.parse`.
- Confirmed `freshrss_item_id` is present as an empty string for locally saved articles in this path.
- Added Flutter `Article` model.
- Added Flutter `ArticlesApi` for `GET /api/articles` with filter, feed ID, category, page, and limit parameters.
- Added Flutter test files for article JSON decoding and `ArticlesApi` success/error paths.
- Extended `tools/flutter_migration` tests to keep `ArticlesApi` aligned with `docs/flutter/API_ROUTE_INVENTORY.json`.

Environment:
- Flutter CLI: still not available on PATH.
- Dart CLI: still not available on PATH.

Validation:
- Passed: `go test ./internal/handlers/article`
- Passed: `go test ./tools/flutter_migration ./internal/routes`
- Not run: `flutter analyze` because Flutter CLI is unavailable.
- Not run: `flutter test` because Flutter CLI is unavailable.

## 2026-06-25: Slice 4 Feed Refresh Action Added

Scope:
- Added Go contract coverage for `/api/feeds/refresh`, verifying HTTP 200 and `{"status":"refreshing"}` response.
- Added Flutter `FeedActionsApi` for feed refresh.
- Added Flutter API test file for feed refresh success/error paths.
- Extended `ReaderRepository` and `HttpReaderRepository` with `refreshFeed`.
- Added a refresh button for each feed in the desktop sidebar.
- Extended `reader_screen_test.dart` with widget coverage for invoking feed refresh.
- Extended `tools/flutter_migration` tests to require feed action API files, route alignment, repository method, sidebar button, and widget coverage markers.

Environment:
- Flutter CLI: still not available on PATH.
- Dart CLI: still not available on PATH.

Validation:
- Passed: `go test ./internal/handlers/feed`
- Passed: `go test ./tools/flutter_migration ./internal/routes`
- Not run: `flutter analyze` because Flutter CLI is unavailable.
- Not run: `flutter test` because Flutter CLI is unavailable.

## 2026-06-25: Slice 3 Basic Article HTML Text Rendering Added

Scope:
- Added `htmlToReadableText` to convert common article HTML into readable text without adding a Flutter package dependency yet.
- Handles script/style removal, paragraph and list spacing, basic tag stripping, and common HTML entities.
- Updated `ArticleDetailPanel` to render readable text instead of raw HTML tags.
- Added `html_text_test.dart` for Dart-side conversion coverage once Flutter/Dart tests can run.
- Updated `reader_screen_test.dart` expectations from raw HTML to readable text.
- Extended `tools/flutter_migration` tests to require the HTML conversion utility, test file, and detail panel integration.

Environment:
- Flutter CLI: still not available on PATH.
- Dart CLI: still not available on PATH.

Validation:
- Passed: `go test ./tools/flutter_migration ./internal/routes`
- Passed: `go test ./internal/handlers/article`
- Not run: `flutter analyze` because Flutter CLI is unavailable.
- Not run: `flutter test` because Flutter CLI is unavailable.

## 2026-06-25: Slice 4 Article Read and Favorite Actions Started

Scope:
- Added Go contract coverage for `/api/articles/read` and `/api/articles/favorite`, verifying HTTP 200 responses and database state changes.
- Added Flutter `ArticleStatusApi` for marking articles read/unread and toggling favorite.
- Added Flutter API test file for read and favorite status calls.
- Extended `ReaderRepository` and `HttpReaderRepository` with article status methods.
- Added mark-read and favorite action buttons to `ArticleDetailPanel`.
- Extended `reader_screen_test.dart` with widget coverage for invoking article status actions.
- Extended `tools/flutter_migration` tests to require status API files, route alignment, repository methods, and detail action coverage markers.

Environment:
- Flutter CLI: still not available on PATH.
- Dart CLI: still not available on PATH.

Validation:
- Passed: `go test ./internal/handlers/article`
- Passed: `go test ./tools/flutter_migration ./internal/routes`
- Not run: `flutter analyze` because Flutter CLI is unavailable.
- Not run: `flutter test` because Flutter CLI is unavailable.

## 2026-06-25: Slice 3 Compact Reader Navigation Added

Scope:
- Added compact-width Reader behavior for mobile layouts.
- Tapping an article on narrow screens now opens the article detail view instead of staying on the list.
- Added a `Back to articles` action to return from detail to list on compact layouts.
- Kept wide desktop behavior as the existing feed/sidebar/list/detail multi-column layout.
- Added `reader_screen_test.dart` coverage for compact-width article open and return behavior.
- Extended `tools/flutter_migration` tests to require compact navigation state and widget coverage markers.

Environment:
- Flutter CLI: still not available on PATH.
- Dart CLI: still not available on PATH.

Validation:
- Passed: `go test ./tools/flutter_migration ./internal/routes`
- Passed: `go test ./internal/handlers/article`
- Not run: `flutter analyze` because Flutter CLI is unavailable.
- Not run: `flutter test` because Flutter CLI is unavailable.

## 2026-06-25: Slice 3 Read-Only Reader UI Started

Scope:
- Replaced the Flutter migration placeholder home with `ReaderScreen`.
- Added `ReaderRepository` and `ReaderSnapshot` for injecting reader data into widget tests.
- Added `HttpReaderRepository` that composes `VersionApi`, `SettingsApi`, `FeedsApi`, and `ArticlesApi`.
- Added read-only Reader UI widgets: feed sidebar, article list panel, and article detail panel.
- Added loading, empty, error, retry, responsive desktop layout, and mobile-width list behavior in the shared Flutter code.
- Added `reader_screen_test.dart` covering loading, empty, error/retry, populated state, and article selection.
- Extended `tools/flutter_migration` tests to require the Reader UI files and widget coverage markers.

Environment:
- Flutter CLI: still not available on PATH.
- Dart CLI: still not available on PATH.

Validation:
- Passed: `go test ./tools/flutter_migration ./internal/routes`
- Passed: `go test ./internal/handlers/update ./internal/handlers/feed ./internal/handlers/settings ./internal/handlers/article`
- Not run: `flutter analyze` because Flutter CLI is unavailable.
- Not run: `flutter test` because Flutter CLI is unavailable.

## 2026-06-25: Slice 3 Article Content Read Added

Scope:
- Added a Go JSON contract test for `/api/articles/content` using pre-cached article content to avoid network-dependent tests.
- Added Flutter `ArticleContent` model.
- Added Flutter `ArticleContentApi` for `GET /api/articles/content?id=...`.
- Added Flutter test files for article content JSON decoding and API success/error paths.
- Extended `ReaderRepository` with `loadArticleContent`.
- Extended `HttpReaderRepository` to compose `ArticleContentApi`.
- Updated `ArticleDetailPanel` to load selected article content with loading, error, empty, and content states.
- Extended `reader_screen_test.dart` with fake article content and content rendering assertions.
- Extended `tools/flutter_migration` tests to keep `ArticleContentApi` aligned with `docs/flutter/API_ROUTE_INVENTORY.json` and require Reader detail content loading.

Environment:
- Flutter CLI: still not available on PATH.
- Dart CLI: still not available on PATH.

Validation:
- Passed: `go test ./internal/handlers/article`
- Passed: `go test ./tools/flutter_migration ./internal/routes`
- Not run: `flutter analyze` because Flutter CLI is unavailable.
- Not run: `flutter test` because Flutter CLI is unavailable.
