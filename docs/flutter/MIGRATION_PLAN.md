# Flutter Migration Plan

This document controls the incremental migration from the current Wails + Vue desktop client to a Flutter client that can target desktop and mobile.

## Migration Rules

1. Move in small, reviewable slices. Each slice must have a clear scope, rollback path, and test evidence.
2. Keep the Go HTTP API as the stable backend boundary until a later slice explicitly changes that contract.
3. Do not remove the existing Vue/Wails frontend until the Flutter client reaches feature parity for the affected workflow.
4. Every slice must update `docs/flutter/MIGRATION_LOG.md`.
5. Every route or payload shape used by Flutter must be covered by a contract, unit, widget, or integration test.

## Target Architecture

- Backend: keep the existing Go packages, SQLite storage, feed processing, summarization, translation, and HTTP REST API.
- Desktop shell: migrate from Wails-hosted Vue to Flutter desktop.
- Mobile shell: use Flutter mobile against the same backend contract. Local-device backend packaging is intentionally deferred until the first Flutter API client slice is green.
- API boundary: `docs/flutter/API_ROUTE_INVENTORY.json` is the first baseline for routes registered by `internal/routes`.

## Slices

### Slice 0: Migration Control Plane

Status: completed

Scope:
- Add this plan.
- Add migration log.
- Add route inventory for the Flutter API boundary.
- Add a Go test that fails when registered API routes and the inventory diverge.

Validation:
- `go test ./internal/routes`

Rollback:
- Remove `docs/flutter/*` and `internal/routes/route_inventory_test.go`.

### Slice 1: Flutter Workspace Bootstrap

Status: completed for shared workspace, tests, and generated platform directories

Prerequisite:
- Flutter SDK available. Current local SDK is `E:\flutter\flutter`; prepend `E:\flutter\flutter\bin` to `PATH` in shells that do not already include it.
- Android SDK is still required before Android builds can be validated.

Scope:
- Create a Flutter app workspace under `flutter_app/`.
- Enable generated Windows, macOS, Linux, Android, and iOS target directories.
- Add a minimal app shell with routing, theme state, localization scaffolding, and test harness.
- Mobile SDK setup and native mobile build validation remain part of later packaging work.

Validation:
- `flutter analyze`
- `flutter test`

### Slice 2: API Client and Models

Status: completed for initial read-only API set

Scope:
- Implement a typed Dart API client for a small read-only route group first.
- Start with version, settings read, feeds list, and article list only.
- Add fixture-backed tests for JSON decoding and HTTP error handling.

Completed in this slice:
- Version read API.
- Feeds list API.
- Settings read API bootstrap fields.
- Article list API.

Validation:
- `flutter test`
- `go test ./internal/routes`
- `go test ./internal/handlers/feed`
- `go test ./internal/handlers/settings`
- `go test ./internal/handlers/article`
- `go test ./tools/flutter_migration`

### Slice 3: Read-Only Reader Workflow

Status: in progress

Scope:
- Build sidebar, article list, and article detail in Flutter using the API client from Slice 2.
- Keep mutating actions disabled until write routes are covered.

Completed in this slice:
- Reader repository abstraction for test injection.
- HTTP repository that composes version, settings, feeds, and articles read APIs.
- Read-only Flutter screen with feed sidebar, article list, article detail, loading, empty, error, and retry states.
- Widget test file covering loading, empty, error, populated, and article selection states.
- Article content read API and detail panel content loading.
- Basic HTML-to-readable-text conversion for article detail content.
- Compact-width article list to detail navigation with a return action for mobile layouts.

Remaining in this slice:
- Replace basic text conversion with richer HTML/media rendering in a later small step.

Validation:
- Flutter widget tests for empty, loading, error, and populated states.
- Flutter widget tests for compact-width list/detail navigation.
- `go test ./internal/handlers/article`
- `go test ./tools/flutter_migration`

### Slice 4: Mutating Workflows

Status: in progress

Scope:
- Add mark-read, favorite, feed refresh, OPML, settings write, and translation workflows in small route groups.
- Add contract tests for request/response payloads before wiring UI actions.

Completed in this slice:
- Article mark-read API.
- Article favorite API.
- Reader detail action buttons for mark-read and favorite.
- Feed refresh API.
- Reader feed sidebar refresh action.
- Settings write API for bootstrap settings.
- Minimal settings sheet for theme, language, and refresh interval.
- Translation text API.
- Reader detail translation action.
- OPML text export API.
- OPML text import API.
- Settings-sheet OPML text import/export controls.
- Platform-neutral OPML file import/export wiring through `file_picker` for desktop and mobile, with text fallback still available.
- Full article-content translation action using the existing text translation API contract.
- Translation settings controls for enablement, translation-only mode, provider selection, and target language.
- Translation cache clearing through `/api/articles/clear-translations`.
- DeepL provider API key and endpoint controls.
- Baidu provider app ID and secret key controls.
- Microsoft provider API key, endpoint, and region controls.
- Tencent provider secret ID, secret key, and region controls.
- AI provider translation profile ID and prompt controls.
- Custom provider core controls: enablement, name, endpoint, method, response path, and timeout.
- Custom provider headers JSON and POST body-template controls.
- Custom provider language-mapping control.

Remaining in this slice:
- Richer translation workflows such as richer AI profile selection.

Validation:
- Go handler tests for touched routes.
- Flutter unit/widget tests for each workflow.
- `go test ./internal/handlers/article`
- `go test ./internal/handlers/feed`
- `go test ./internal/handlers/settings`
- `go test ./internal/handlers/translation`
- `go test ./internal/handlers/opml`
- `go test ./tools/flutter_migration`
- `flutter analyze`
- `flutter test`

### Slice 5: Packaging and Platform Integration

Status: pending

Scope:
- Define desktop packaging.
- Define mobile backend strategy.
- Replace Wails-only window/browser integration with Flutter plugins or platform channels.

Validation:
- Platform builds on every supported target available in CI.
