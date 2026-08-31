# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MrRSS is a cross-platform RSS reader. **On this branch the interface is a native
macOS client**: a SwiftUI application in `frontend`, talking to the Go
backend over its HTTP API. The Vue frontend and the Wails shell are not part of
this branch, so the Go binary is a plain API server with no embedded web assets.

## Development Commands

### macOS client (Swift)

```bash
# Build and test the client
swift build --package-path frontend
swift test --package-path frontend

# Run the backend and the client together
./frontend/run.sh

# Run the client against a backend already running elsewhere
MRRSS_API_BASE_URL=http://127.0.0.1:1234/api swift run --package-path frontend MrRSS

# Build the .app bundle and the DMG
make build-app VERSION=1.3.28
```

### Backend (Go)

```bash
go mod download
go test ./...
go test -v -timeout=5m -coverprofile=coverage.out -covermode=atomic ./internal/...
go vet ./...
gofmt -w . && goimports -w .

# Run the API server
go run . -host 127.0.0.1 -port 1234
```

### Makefile

```bash
make help          # every target
make check         # lint, test and build
make test          # backend and client tests
make clean         # remove build artifacts
```

## Architecture Overview

- **Backend**: Go 1.27+, SQLite in WAL mode, serving only `/api`
- **Client**: SwiftUI for macOS 14+, SwiftPM package, no third-party dependencies
- **Communication**: the HTTP API alone; there are no native bindings

### Key directories

```plaintext
MrRSS/
├── main.go                       # API server entry point
├── internal/                     # Go backend packages
│   ├── database/                 # Data layer, models, migrations
│   ├── handlers/                 # HTTP handlers by feature
│   ├── feed/                     # Fetching and processing
│   ├── translation/              # Multi-language support
│   ├── discovery/                # Feed discovery engine
│   └── config/settings_schema.json  # The single source of truth for settings
├── frontend/               # The macOS client
│   ├── Sources/
│   │   ├── Models/               # Codable mirrors of the API payloads
│   │   ├── Services/API/         # Transport plus one extension per domain
│   │   ├── Localization/         # The translation catalogue
│   │   ├── ViewModels/           # AppViewModel and its feature extensions
│   │   └── Views/                # SwiftUI views
│   ├── Tests/
│   └── build-app.sh              # Bundle and DMG packaging
└── tools/settings-swift/         # Generates the client settings catalogue
```

## Code Patterns and Guidelines

### Backend

1. **Context usage**: always take a `context.Context` in exported methods
2. **Error handling**: wrap with context, `fmt.Errorf("operation failed: %w", err)`
3. **Database**: prepared statements for every query
4. **Validation**: check URLs, file paths and user input
5. **Cleanup**: `defer` for resources

### macOS client

1. **One place per concern**: models mirror the API, `APIService` extensions hold
   the calls, `AppViewModel` extensions hold the behaviour, views stay thin
2. **Interface strings** come from the catalogue: call `t("some.key")`, never a
   literal. Keys ported from the previous frontend live in
   `Localization/LocalizationTables.swift`; wording only this client needs lives
   in `Localization/ClientStrings.swift` under a `client.` prefix
3. **Language** follows the `language` setting the backend stores, not the
   system locale, so dates are formatted with that locale too
4. **Optimistic updates**: change local state first, roll back if the request
   fails (see `mutateArticle`)
5. **Tests** use `StubAPIClient`, which fails any call the test did not prepare,
   so an unexpected request is reported rather than silently returning nothing

### Settings

`internal/config/settings_schema.json` is the single source of truth.

1. Add the setting there
2. Run `go run tools/settings-generator/main.go` for the Go side
3. Run `python3 tools/settings-swift/generate.py` for the client side
4. Add wording to the translation catalogue if the generator reports a fallback

The client reads and writes settings by their snake_case keys.

## Testing

```bash
go test -v -timeout=5m ./internal/...          # backend
go test -v ./internal/database -run TestName   # one backend test
swift test --package-path frontend       # client
swift test --package-path frontend --filter LocalizationTests
```

## Important Notes

1. **No web assets**: the Go binary serves only `/api` and embeds nothing; `frontend/` is the Swift package, and `frontend/dist` holds the packaged `.app` and DMG
2. **Privacy-first**: no external analytics, all data stored locally
3. **Portable mode**: supported through `portable.txt`
4. **Concurrent processing**: feed fetching uses goroutines with configurable limits
5. **Upstream merges**: this branch deletes `frontend/`, so merging upstream
   changes to those files reports "deleted by us"; keep them deleted
