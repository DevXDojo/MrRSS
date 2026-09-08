# Agent Guidelines for MrRSS

## Purpose and source of truth

This file records durable project conventions. It may remain unchanged across many releases; do not treat it as a snapshot of the latest release or a complete feature inventory.

- Read dependency and toolchain versions from `go.mod`, package manifests, lockfiles, and CI configuration. Do not pin versions or minimum versions in this guide.
- Read commands, build flags, platform requirements, and test setup from the relevant Taskfiles, package scripts, and workflows before running them.
- For implementation details, defaults, limits, provider support, and platform behavior, inspect the relevant source and tests. If descriptive documentation disagrees with executable configuration or code, use the latter to establish current behavior; preserve the conventions in this guide when making changes.
- Update this file when a lasting workflow or architectural rule changes, not for routine releases, dependency upgrades, or individual features. Keep release history in `CHANGELOG.md` and detailed explanations in `docs/`.

## Project principles

MrRSS is a privacy-focused RSS reader with a Go backend, a Wails desktop shell, a Vue and TypeScript frontend, and SQLite storage. It also supports a server build.

- Preserve local storage and offline behavior where supported. External services must follow the user's configuration; do not introduce analytics or unnecessary data transmission.
- Consider Windows, macOS, and Linux when changing paths, process execution, startup, window behavior, or packaging.
- Preserve existing subscriptions, articles, reading state, favorites, settings, and credentials when changing persistence or cleanup behavior.
- Keep business logic usable across desktop and server modes. Use the HTTP API for data operations and reserve Wails integration for desktop capabilities.

## Repository navigation

Use this map as a starting point, then search for the implementation and neighboring tests. It is intentionally not an exhaustive directory listing.

| Area | Starting points |
| --- | --- |
| Application startup and mode selection | `main.go`, `main-core.go`, build tags in related Go files |
| HTTP routes and request handling | `internal/routes/`, `internal/handlers/`, `internal/middleware/` |
| Desktop integration API | `internal/desktopapi/` |
| Persistence and data models | `internal/database/`, `internal/models/` |
| Settings definitions and generation | `internal/config/settings_schema.json`, `tools/settings-generator/` |
| Backend features and shared utilities | Feature packages under `internal/`, helpers under `internal/utils/` |
| Reader interface and client state | `frontend/src/components/`, `frontend/src/composables/`, `frontend/src/stores/` |
| Frontend contracts and translations | `frontend/src/types/`, `frontend/src/i18n/` |
| Build, checks, and automation | `Taskfile.yml`, `build/`, `scripts/`, `.github/workflows/` |
| Documentation, website, and integration skill | `docs/`, `website/`, `skills/` |

## Working on a change

1. Inspect `git status` and any applicable directory-specific instructions. Preserve unrelated changes already in the working tree.
2. Read the affected implementation, callers, and tests. For a data feature, trace the frontend, route registration, handler, and persistence or service layer before editing.
3. Follow neighboring patterns and reuse existing utilities. Keep changes focused; split components or extract helpers when responsibilities become difficult to understand, rather than enforcing arbitrary file or folder size limits.
4. Keep API payloads, Go models, TypeScript types, settings, and translations consistent across the layers a change touches.
5. Edit generation sources and regenerate their outputs. Check generated-file headers and generator code instead of maintaining a duplicate list of generated files here.
6. Validate the affected behavior, inspect the final diff for accidental changes, and report what changed, what was checked, and any remaining limitations.

Avoid unrelated dependency upgrades, lockfile churn, broad formatting, and release metadata changes. Build and formatting tools may modify files; review their output against the initial working tree.

## Backend conventions

- Keep handlers focused on request validation, service calls, and HTTP responses. Put reusable business logic in feature packages and database operations in `internal/database/`.
- Propagate `context.Context` through operations that perform I/O or can block. Use request contexts, cancellation, and bounded timeouts for network requests, background work, and child processes.
- Wrap errors with useful operation context using `%w` when callers need the underlying error. Return meaningful HTTP statuses and avoid exposing credentials or internal details in responses.
- Use parameterized SQL for values and prepared statements where appropriate. Validate dynamic identifiers against an allowlist. Use transactions for changes that must succeed together.
- Keep initial schema creation and migrations consistent. Verify both fresh databases and upgrades when changing persisted data; make migrations safe to rerun and preserve user data.
- Close database rows, statements, response bodies, and files, and release timers and cancellation functions. Check errors from iteration and transaction completion.
- Bound concurrent work, synchronize shared state, and ensure goroutines terminate on cancellation or shutdown. Preserve progress reporting and cache invalidation when changing background operations.
- Reuse existing network, proxy, path, and logging utilities. Format modified Go files with `gofmt`.

## Frontend conventions

- Use Vue Composition API with `<script setup lang="ts">`. Type props, emits, API data, and shared state; avoid introducing `any` to bypass a type mismatch.
- Reuse Pinia stores and composables for shared behavior. Use `useAppStore()` for application state and keep component-local state local.
- Use the HTTP API for data access. Preserve loading, error, cancellation, and empty states; prevent stale asynchronous responses from overwriting newer selections.
- Use `t()` for user-facing strings and update the corresponding English and Chinese translations under `frontend/src/i18n/`. Reuse existing translation namespaces.
- Show actionable failures through the existing toast mechanism (`window.showToast()`) where appropriate, rather than only logging them.
- Reuse shared components, Phosphor icons, semantic Tailwind classes, and theme variables. Keep light, dark, and responsive layouts working; avoid hardcoded presentation values when existing styles cover the need.
- Debounce frequent operations such as search and autosave. Clean up listeners, observers, timers, and pending work on unmount.

## Settings and generated code

Settings are schema-driven. Do not add a setting only to generated types, defaults, or handlers.

1. Edit `internal/config/settings_schema.json`. Keep setting keys and frontend properties in `snake_case`, and mark sensitive settings as encrypted.
2. From the repository root, run:

   ```sh
   go run tools/settings-generator/main.go
   ```

3. Review all generated changes. If generated behavior needs to change, update the generator; put feature-specific behavior in the surrounding handwritten code.
4. Add the consuming logic, UI where needed, and matching English and Chinese translations. Follow the existing settings and autosave flow.
5. Verify defaults, save/load round trips, and compatibility with existing stored values. Treat renaming or removing a persisted key as a migration, not just a type change.

See [Settings Guide](docs/SETTINGS.md) for examples; the schema and generator determine the actual output files and behavior.

## Security boundaries

- Treat feed content, imported files, external responses, and user-supplied URLs and paths as untrusted. Validate inputs at the boundary, including allowed URL schemes and filesystem containment.
- Never interpolate untrusted values into SQL or shell command strings. Use filesystem APIs for file operations and argument-based process execution with timeouts for scripts.
- Custom scripts execute code on the user's machine. Preserve path and execution restrictions; do not describe timeouts or a working directory as a complete sandbox.
- Never render raw untrusted content through `v-html` or equivalent DOM APIs. Article rendering needs HTML: route it through the appropriate content preparation and sanitization helpers, and check the full rendering path when changing that code.
- Use the existing encryption helpers for secrets. Do not put credentials, private article content, or user databases in logs, fixtures, or source control.
- Preserve desktop loopback and origin protections and the server mode's access boundaries. Do not widen listeners or CORS policies as an incidental fix.

## Development and validation

Resolve toolchain requirements from the repository configuration before installing tools. Keep the Wails CLI compatible with the module dependency instead of assuming the latest CLI is appropriate. Platform dependencies and CGO requirements belong in build configuration and CI, not a fixed list here.

Common entry points are below. Confirm they still exist in the current checkout; package scripts and Taskfiles are authoritative.

| Purpose | Working directory | Command |
| --- | --- | --- |
| Download Go dependencies | Repository root | `go mod download` |
| Install locked frontend dependencies | `frontend/` | `npm ci` |
| Discover build tasks | Repository root | `task --list` |
| Desktop development | Repository root | `task dev` |
| Backend tests | Repository root | `go test -timeout=5m ./internal/...` |
| Go static checks | Repository root | `go vet ./...` |
| Frontend unit tests, single run | `frontend/` | `npm run test:unit` |
| Frontend lint without autofix | `frontend/` | `npx --no-install eslint .` |
| Frontend build | `frontend/` | `npm run build` |
| End-to-end tests | `frontend/` | `npm run test:e2e` |
| Desktop build | Repository root | `wails3 build` or `task build` |

- Start with tests for the affected packages or files, then run broader checks when the scope warrants them. Add regression tests for bug fixes and behavior changes; cover relevant failures and boundary conditions without relying on live services or real credentials.
- Use a non-watching test command for automated verification. Inspect scripts before running them: lint and format scripts may apply fixes across the tree.
- Go commands that compile the entry packages need the embedded `frontend/dist` assets. Build the frontend first if they are absent or stale.
- Run `wails3 build` (or the equivalent current build task) before completing application code or build changes. For shared backend or startup changes, also verify the server build using the current [server instructions](docs/SERVER_MODE/README.md).
- For documentation-only changes, verify referenced paths, links, and commands, and run `git diff --check`. Application builds are only needed if the change also affects build behavior or requires testing runnable examples.
- If a required check cannot run, state the exact command and blocker. Distinguish environmental limitations and pre-existing failures from failures introduced by the change; do not claim unrun checks passed or change dependencies merely to make local validation pass.

## Further documentation

- [Architecture](docs/ARCHITECTURE.md): component responsibilities and data flow.
- [Code Patterns](docs/CODE_PATTERNS.md): examples for backend, frontend, and styling.
- [Testing](docs/TESTING.md): testing patterns and fixtures.
- [Build Requirements](docs/BUILD_REQUIREMENTS.md): platform setup guidance; cross-check with current build tasks and CI.
- [Settings](docs/SETTINGS.md): schema and generator workflow.
- [Server Mode](docs/SERVER_MODE/README.md): server build and deployment behavior.
- [Custom Script Mode](docs/CUSTOM_SCRIPT_MODE.md): script integration and examples.
- [Contributing](CONTRIBUTING.md): contribution process.

When updating this guide, keep rules concise and actionable, verify links, and prefer references to authoritative files over copying details that will drift.
