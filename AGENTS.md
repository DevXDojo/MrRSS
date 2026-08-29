# AI Agent Guidelines for MrRSS

> **Quick Links**: [Architecture](docs/ARCHITECTURE.md) | [Code Patterns](docs/CODE_PATTERNS.md) | [Testing](docs/TESTING.md) | [Build Requirements](docs/BUILD_REQUIREMENTS.md) | [Settings Guide](docs/SETTINGS.md)

## Project Overview

**MrRSS** is a modern, privacy-focused, cross-platform desktop RSS reader with AI-powered features.

### Core Philosophy

- **Privacy-First**: No external analytics, all data stored locally
- **Cross-Platform**: Native desktop experience on Windows, macOS, and Linux
- **Modern Tech Stack**: Go 1.27+ backend with a native SwiftUI client for macOS
- **AI-Enhanced**: Local algorithms (TF-IDF + TextRank) and cloud AI integration
- **Performance-Optimized**: Concurrent processing, intelligent caching, WAL mode SQLite

### Tech Stack

- **Backend**: Go 1.27+, SQLite with `modernc.org/sqlite`, serving only `/api`
- **Client**: SwiftUI for macOS 14+, a SwiftPM package with no third-party dependencies
- **Communication**: the HTTP REST API alone; there are no native bindings
- **Translations**: a catalogue ported from the previous frontend (English/Chinese)

> **This branch builds the macOS client.** The Vue frontend and the Wails shell
> were removed, so the Go binary has no embedded web assets.


### Key Features

#### Core Functionality
- **Feed Management**: RSS/Atom subscription with custom script support (Python, Shell, Node.js, Ruby, PowerShell)
- **OPML Support**: Import/export subscriptions from other RSS readers
- **Concurrent Fetching**: Intelligent feed refresh with configurable limits

#### AI-Powered Features
- **Local Summarization**: TF-IDF + TextRank algorithms work offline without API keys
- **AI Summarization**: Integration with OpenAI-compatible APIs (GPT, Claude, Gemini, etc.)
- **Smart Translation**: Google Translate, DeepL, Baidu Translation, and AI-based translation
- **Translation Caching**: Intelligent cache management for performance

#### Advanced Features
- **Smart Discovery**: Auto-discover RSS feeds from websites and friend links with progress tracking
- **Custom Scripts**: Execute any script type for non-standard feeds
- **Email Integration**: IMAP support to convert newsletters to feeds
- **XPath Scraping**: Extract content from HTML-based websites
- **Filtering Rules**: "If-then" automation for intelligent article organization
- **Image Gallery Mode**: Visual browsing for image-heavy feeds
- **FreshRSS Sync**: Synchronize with FreshRSS instances
- **Custom CSS**: UI customization and personalization
- **Proxy Support**: Per-feed or global proxy configuration

#### User Experience
- **Modern UI**: Dark/Light/Auto themes with responsive design
- **Keyboard Shortcuts**: Fully customizable keybindings
- **Multi-Language**: English and Chinese (simplified) support
- **Statistics Tracking**: Monitor reading habits and patterns

📚 **Detailed Feature Documentation**: See [ARCHITECTURE.md](docs/ARCHITECTURE.md)

## Project Structure

### High-Level Organization

```plaintext
MrRSS/
├── main.go                      # API server entry point
├── internal/                    # Backend Go code
│   ├── ai/                      # AI configuration and utilities
│   ├── aiusage/                 # AI usage tracking and limits
│   ├── cache/                   # Media cache management
│   ├── config/                  # Configuration with schema-driven generation
│   ├── crypto/                  # Encryption for sensitive settings
│   ├── database/                # SQLite operations with WAL mode
│   ├── discovery/               # Smart feed discovery system
│   ├── feed/                    # RSS/Atom parsing and script execution
│   ├── freshrss/                # FreshRSS synchronization logic
│   ├── handlers/                # HTTP API handlers (organized by feature)
│   ├── jsonimport/              # JSON feed import utilities
│   ├── models/                  # Core data structures
│   ├── monitor/                 # System monitoring utilities
│   ├── network/                 # Network detection and utilities
│   ├── opml/                    # OPML import/export
│   ├── rsshub/                  # RSSHub integration
│   ├── rules/                   # Filtering rules engine
│   ├── statistics/              # Usage analytics
│   ├── summary/                 # TF-IDF + TextRank + AI summarization
│   ├── translation/             # Multi-service translation
│   ├── utils/                   # Platform utilities
│   └── version/                 # Version constant
├── frontend-swift/
│   ├── Sources/
│   │   ├── Models/              # Codable mirrors of the API payloads
│   │   ├── Services/API/        # Transport plus one extension per domain
│   │   ├── Localization/        # Translation catalogue and client-only wording
│   │   ├── ViewModels/          # AppViewModel and its feature extensions
│   │   └── Views/               # SwiftUI views, including the sidebar outline
│   ├── Tests/                   # XCTest suite
│   ├── build-app.sh             # Bundle and DMG packaging
│   └── run.sh                   # Backend plus client launcher
├── docs/                        # Documentation
├── build/                       # Icon assets used when packaging
├── tools/                       # Settings generators (Go and Swift)
└── scripts/                     # Automation scripts (check, pre-release)
```


📚 **Detailed Structure**: See [ARCHITECTURE.md](docs/ARCHITECTURE.md)

## Key Technologies & Patterns

### Backend Architecture (Go 1.27+)

#### Framework & Communication
- **HTTP REST API**: the only channel between the client and the backend
- **SQLite**: Pure Go implementation (`modernc.org/sqlite`) with WAL mode

#### Core Packages
- **`internal/handlers/`**: Feature-based API handlers
  - `ai/` - AI configuration test and AI search
  - `article/` - CRUD operations, filtering, export
  - `browser/` - URL opening in browser
  - `chat/` - AI chat sessions and messages
  - `core/` - Core handler initialization
  - `custom_css/` - Custom CSS management
  - `discovery/` - Feed discovery engine
  - `feed/` - Feed management, metadata
  - `filter_category/` - Saved filters management
  - `freshrss/` - FreshRSS synchronization
  - `media/` - Image/audio/video processing and proxy
  - `network/` - Network detection
  - `opml/` - OPML import/export
  - `rsshub/` - RSSHub integration
  - `rules/` - Filtering automation
  - `script/` - Custom script management
  - `settings/` - Configuration management
  - `statistics/` - Usage statistics
  - `summary/` - Local + AI summarization
  - `tags/` - Tag management
  - `translation/` - Multi-service translation
  - `update/` - Application update handling
  - `window/` - Window state management

- **`internal/database/`**: SQLite operations
  - WAL mode for concurrency
  - Prepared statements for performance
  - Automatic cleanup with favorites preservation
  - 15,000+ articles cache per feed

- **`internal/feed/`**: RSS/Atom processing
  - Concurrent fetching with configurable limits
  - Custom script execution (Python, Shell, Node.js, Ruby, PowerShell)
  - Content extraction and processing
  - Intelligent refresh scheduling

- **`internal/summary/`**: Summarization algorithms
  - Local: TF-IDF + TextRank (offline, no API)
  - AI: OpenAI-compatible APIs (GPT, Claude, Gemini)
  - Smart caching for performance

- **`internal/translation/`**: Translation services
  - Google Translate (free, no API key)
  - DeepL API integration
  - Baidu Translation API
  - AI-based translation
  - Dynamic service selection

#### Concurrency & Performance
- **Goroutines**: Parallel feed fetching
- **Semaphores**: Configurable concurrency limits
- **Context with Timeout**: Graceful cancellation
- **Sync Mutex**: Safe shared state access

#### Security Best Practices
- **Input Validation**: URL, file path validation
- **Path Sanitization**: Directory traversal prevention
- **Prepared Statements**: SQL injection prevention
- **Safe File Operations**: No shell command concatenation
- **Script Sandboxing**: Restricted execution context

### Client Architecture (SwiftUI)

**Layering**

- `Models/` mirror the API payloads. Every decoder is written by hand and
  tolerates missing fields, because older backends omit some of them.
- `Services/API/` holds the transport in `APIService.swift` and one extension
  per domain. `APIClient` is the full protocol, so the compiler catches a call
  the service forgot to implement.
- `ViewModels/AppViewModel.swift` owns the state; the feature extensions beside
  it own the behaviour. Views stay thin.
- `Views/` are SwiftUI, except the sidebar, which is an `NSOutlineView` so that
  dragging behaves the way source lists do on macOS.

**Conventions**

- Interface strings come from the catalogue: `t("some.key")`, never a literal.
- The language follows the `language` setting the backend stores, and dates are
  formatted with the matching locale.
- Changes are applied locally first and rolled back if the request fails.
- Reading preferences a feed overrides (view mode, auto-expand) win over the
  global setting, as they did before.

## Development Workflow

### Getting Started

1. **Prerequisites**: Go 1.27+, macOS 14+, Xcode 15+
2. **Setup**: `go mod download`
3. **Development**: `./frontend-swift/run.sh` starts the backend and the client
4. **Build**: `make build-app VERSION=1.3.28` produces the bundle and the DMG

### Development Tools

#### Make
```bash
make help        # Show all targets
make dev         # Backend and client together
make build       # Build both halves
make build-app   # Package the .app and the DMG
```

#### Make (Alternative)
```bash
make help        # Show available commands
make check       # Run lint + test + build
make clean       # Clean artifacts
```

#### Cross-Platform Scripts
**Linux/macOS:**
```bash
./scripts/check.sh            # Run all checks
./scripts/pre-release.sh      # Pre-release validation
```

**Windows (PowerShell):**
```powershell
.\scripts\check.ps1           # Run all checks
.\scripts\pre-release.ps1     # Pre-release validation
```

### Code Organization Guidelines

- **File Size**: Refactor files over 300-400 lines
- **Folder Organization**: Create subfolders when folders have 10-15+ files
- **Backend**: Extract related functions into separate files within the same package
- **Frontend**: Split large components or extract logic into composables
- **Build Verification**: run `make build` and both test suites before completing changes

### Settings Management (OPTIMIZED)

✅ **The settings system has been optimized!** Adding new settings is now much simpler:

**Quick Method (3 steps):**

1. Edit `internal/config/settings_schema.json` (add 5 lines)
2. Run `go run tools/settings-generator/main.go` (generates all code)
3. Add UI and translations (optional)

**What Gets Generated Automatically:**

- Backend types and handlers
- Frontend types and composables
- Database initialization keys
- Default values

📚 **Complete Guide**: See [docs/SETTINGS.md](docs/SETTINGS.md)

**Important Notes:**
- **Frontend uses snake_case** (e.g., `settings.ai_api_key`, `settings.update_interval`)
- All generated files are sorted alphabetically for stable diffs
- The generator handles all boilerplate automatically

## Coding Standards

### Go Standards

1. **Context Usage**: Always use `context.Context` for exported methods
2. **Error Handling**: Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
3. **Database Operations**: Use prepared statements for all queries
4. **Input Validation**: Validate URLs, file paths, and user inputs
5. **Resource Cleanup**: Use `defer` for proper cleanup
6. **No Shell Commands**: Never use shell command concatenation (security risk)
7. **Concurrency**: Use goroutines with proper synchronization (WaitGroup, Mutex, semaphores)

### Swift Standards

- SwiftUI with `@MainActor` view models; async work through structured concurrency
- Named types over tuples once a value carries more than two fields
- No force unwrapping outside tests
- Comments explain why, not what; the surrounding code sets the density

### Security Practices

1. **Input Validation**: URLs, file paths, user data
2. **Safe File Operations**: `os.Remove()` instead of shell commands
3. **SQL Injection**: Prepared statements for all queries
4. **XSS Prevention**: No `v-html` for user content
5. **Script Execution**: Timeout enforcement, path sandboxing
6. **Path Traversal**: Validate and sanitize file paths
7. **Sensitive Data**: Encrypt API keys and passwords

📚 **Code Examples**: See [CODE_PATTERNS.md](docs/CODE_PATTERNS.md)

## Testing Guidelines

### Backend Tests

- **Run with timeout**: `go test -v -timeout=5m ./...`
- **Coverage report**: `go test -coverprofile=coverage.out ./...`
- **Single test**: `go test -v ./internal/database -run TestSpecificFunction`
- **Coverage goals**: Database 80%+, Handlers 70%+, Business Logic 80%+, Utilities 90%+

### Client Tests

```bash
swift test --package-path frontend-swift
swift test --package-path frontend-swift --filter LocalizationTests
```

Tests use `StubAPIClient`, which fails any call the test did not prepare, so an
unexpected request is reported rather than quietly returning nothing.

## Architecture Highlights

### Schema-Driven Settings System

The settings system uses a JSON schema to automatically generate:
- Backend Go code (types, handlers, database operations)
- Frontend TypeScript code (types, composables)
- Default values and initialization

**Benefits:**
- 90% reduction in development time for new settings
- Zero copy-paste errors
- Guaranteed consistency between frontend and backend
- Automatic type safety

### One Communication Channel

The client speaks to the backend only over `/api/*`. System integration that a
desktop shell used to provide is handled natively instead: links open through
`NSWorkspace`, file import and export use `NSOpenPanel` and `NSSavePanel`, and
window geometry is saved through `/api/window/save`.

### Intelligent Caching System

Multiple caching layers optimize performance:
- **Translation Cache**: Avoid redundant API calls
- **Media Cache**: Images, audio, video caching
- **Article Cache**: 15,000+ articles per feed with automatic cleanup

### Concurrent Processing

Feed fetching uses sophisticated concurrency control:
- **Goroutines**: Parallel feed processing
- **Semaphores**: Configurable concurrency limits
- **Context Timeout**: Graceful cancellation
- **Progress Tracking**: Real-time status updates
- **Error Collection**: Collect all errors without stopping

## Important Notes

1. **No web assets**: the Go binary serves only `/api`; nothing is embedded
2. **Privacy-First**: No external analytics, all data stored locally
3. **macOS client**: the interface targets macOS 14+; the backend stays portable
4. **Portable Mode**: Supports portable deployment with `portable.txt`
5. **Concurrent Processing**: Feed fetching uses goroutines with configurable limits
6. **snake_case settings**: the client reads and writes settings by their schema keys
7. **Settings Generation**: run both generators after changing the schema
8. **Upstream merges**: `frontend/` is deleted here, so upstream changes to it
   report "deleted by us"; keep them deleted

## Common Issues

1. **Build Requirements**: Xcode 15+ for the Swift toolchain, Go 1.27+ for the backend
2. **Slow type checking**: keep SwiftUI bodies small; a long expression can stall the build
3. **Running the client**: `./frontend-swift/run.sh` starts the backend and the client together
4. **Database Migrations**: Handle schema changes carefully with proper versioning
5. **Settings Not Working**: Ensure you ran the settings generator after editing the schema

## Related Documentation

- [Architecture Overview](docs/ARCHITECTURE.md) - Detailed system architecture
- [Code Patterns](docs/CODE_PATTERNS.md) - Common patterns and examples
- [Testing Guide](docs/TESTING.md) - Testing strategies
- [Settings Guide](docs/SETTINGS.md) - Settings system documentation
- [Build Requirements](docs/BUILD_REQUIREMENTS.md) - Platform-specific dependencies
- [Custom Scripts](docs/CUSTOM_SCRIPTS.md) - Writing custom feed scripts
