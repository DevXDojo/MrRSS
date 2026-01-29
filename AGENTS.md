# AI Agent Guidelines for MrRSS

> **Quick Links**: [Architecture](docs/ARCHITECTURE.md) | [Code Patterns](docs/CODE_PATTERNS.md) | [Testing](docs/TESTING.md) | [Build Requirements](docs/BUILD_REQUIREMENTS.md) | [Settings Guide](docs/SETTINGS.md)

## Project Overview

**MrRSS** is a modern, privacy-focused, cross-platform desktop RSS reader with AI-powered features.

### Core Philosophy

- **Privacy-First**: No external analytics, all data stored locally
- **Cross-Platform**: Native desktop experience on Windows, macOS, and Linux
- **Modern Tech Stack**: Go 1.25+ + Wails v3 + Vue 3.5+ with TypeScript
- **AI-Enhanced**: Local algorithms (TF-IDF + TextRank) and cloud AI integration
- **Performance-Optimized**: Concurrent processing, intelligent caching, WAL mode SQLite

### Tech Stack

- **Backend**: Go 1.25+ with Wails v3 (alpha) framework, SQLite with `modernc.org/sqlite`
- **Frontend**: Vue 3.5+ Composition API, Pinia, Tailwind CSS 3.3+, Vite 5+, TypeScript
- **Communication**: HTTP REST API (not Wails bindings) for data operations
- **Icons**: Phosphor Icons | **I18n**: vue-i18n (English/Chinese)

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
├── main.go                      # Desktop application entry point
├── main-core.go                 # Headless server entry point
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
│   ├── version/                 # Version constant
│   └── webview/                 # Webview utilities
├── frontend/src/
│   ├── components/              # Vue components (article/, sidebar/, modals/, common/)
│   ├── composables/             # Reusable logic (article/, feed/, discovery/, rules/, ui/)
│   ├── stores/                  # Pinia state management
│   ├── types/                   # TypeScript definitions
│   ├── i18n/                    # Translations (en, zh)
│   └── utils/                   # Frontend utilities
├── docs/                        # Comprehensive documentation
├── build/                       # Platform-specific build configurations
├── tools/                       # Development tools (settings generator)
└── scripts/                     # Automation scripts (check, pre-release)
```

📚 **Detailed Structure**: See [ARCHITECTURE.md](docs/ARCHITECTURE.md)

## Key Technologies & Patterns

### Backend Architecture (Go 1.25+)

#### Framework & Communication
- **Wails v3**: Desktop application framework with HTTP API
- **HTTP REST API**: Primary communication (not Wails bindings)
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

### Frontend Architecture (Vue 3.5+)

#### Core Technologies
- **Vue 3.5+**: Composition API with `<script setup>`
- **TypeScript**: Full type safety
- **Pinia**: State management
- **Tailwind CSS 3.3+**: Utility-first styling
- **Vite 5+**: Fast build tooling
- **vue-i18n**: Internationalization (en/zh)
- **Phosphor Icons**: Iconography

#### Component Organization
- **`components/article/`**: Article display and rendering
  - `ArticleList.vue` - Virtual scrolling list
  - `ArticleDetail.vue` - Article reader
  - `ArticleContent.vue` - Content rendering with multimedia support
  - `parts/` - Reusable content components

- **`components/sidebar/`**: Navigation sidebar
- **`components/modals/`**: Modal dialogs
  - `settings/` - Settings tabs (general, feeds, rules, shortcuts, about)
  - `feed/` - Add/Edit feed
  - `filter/` - Article filters
  - `rules/` - Rules editor
  - `discovery/` - Feed discovery

- **`components/common/`**: Reusable components
  - `Toast.vue` - Notifications
  - `ContextMenu.vue` - Right-click menus
  - `ImageViewer.vue` - Image gallery

#### Composables Architecture
- **`composables/article/`**: Article operations
- **`composables/feed/`**: Feed management
- **`composables/discovery/`**: Feed discovery
- **`composables/filter/`**: Article filtering
- **`composables/rules/`**: Filtering rules
- **`composables/ui/`**: UI utilities (context menu, keyboard shortcuts, toast)

#### State Management (Pinia)
The main store (`stores/app.ts`) manages:
- Articles and feeds state
- Filter states (all, unread, favorites, imageGallery)
- Theme management (light/dark/auto)
- Refresh progress tracking
- Unread counts
- Search functionality

#### Multimedia Support
- **Images**: Clickable viewer, context menu, download support
- **Audio**: Full-width player with podcast styling
- **Video**: Responsive player with aspect ratio
- **Iframes**: 16:9 aspect ratio for embeds (YouTube/Vimeo)
- **Rich Text**: Tables, blockquotes, code blocks, math (KaTeX)

📚 **Detailed Patterns**: See [CODE_PATTERNS.md](docs/CODE_PATTERNS.md)

## Development Workflow

### Getting Started

1. **Prerequisites**: Go 1.25+, Node.js 18+, Wails CLI v3
2. **Setup**: `go mod download && cd frontend && npm install`
3. **Development**: `wails3 dev` (hot reload enabled)
4. **Build**: Use `task build` or `wails3 build`

### Development Tools

#### Task Runner (Recommended)
```bash
task --list      # Show all tasks
task dev         # Development mode
task build       # Build for current platform
task package     # Package with installer
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
- **Build Verification**: Always run `wails3 build` before completing changes

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

### Vue/TypeScript Standards

1. **Composition API**: Use `<script setup>` syntax
2. **Type Safety**: Proper TypeScript typing for all props and data
3. **Internationalization**: Always use `t()` for user-facing strings
4. **State Management**: Access store via `useAppStore()`
5. **Error Handling**: Show toast notifications with `window.showToast()`
6. **Styling**: Use Tailwind semantic classes (no inline styles)
7. **Performance**: Debounce frequent operations (search, auto-save)
8. **Lifecycle**: Proper cleanup on component unmount (timers, listeners)

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

### Frontend Tests

- **Unit tests**: `npm test` (uses Vitest)
- **E2E tests**: `npm run test:e2e` (uses Cypress)
- **Test UI**: `npm run test:ui`
- **Coverage goals**: Components 70%+, Composables 80%+, Utilities 90%+

📚 **Testing Guide**: See [TESTING.md](docs/TESTING.md)

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

### Hybrid Communication Pattern

MrRSS uses a **hybrid approach**:

1. **HTTP API** (`/api/*`) - Primary communication for data operations
2. **Wails Bindings** - System integration (browser, window management)
3. **Static Files** - Frontend assets served from embedded `frontend/dist`

This provides better control and flexibility compared to pure Wails bindings.

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

1. **No Wails Bindings for Data**: The application primarily uses HTTP API, not Wails bindings for data operations
2. **Privacy-First**: No external analytics, all data stored locally
3. **Cross-Platform**: Build for Windows, macOS, and Linux
4. **Portable Mode**: Supports portable deployment with `portable.txt`
5. **Single Instance**: Enforced on Windows/macOS, disabled on Linux due to D-Bus issues
6. **Concurrent Processing**: Feed fetching uses goroutines with configurable limits
7. **Frontend snake_case**: All frontend settings use snake_case (not camelCase)
8. **Settings Generation**: Always use the schema-driven generator for new settings

## Common Issues

1. **Linux D-Bus Issues**: Single instance mode disabled on Linux
2. **Build Requirements**: Ensure platform-specific dependencies are installed
3. **Frontend Hot Reload**: Use `wails3 dev` for development with hot reload
4. **Database Migrations**: Handle schema changes carefully with proper versioning
5. **Settings Not Working**: Ensure you ran the settings generator after editing the schema

## Related Documentation

- [Architecture Overview](docs/ARCHITECTURE.md) - Detailed system architecture
- [Code Patterns](docs/CODE_PATTERNS.md) - Common patterns and examples
- [Testing Guide](docs/TESTING.md) - Testing strategies
- [Settings Guide](docs/SETTINGS.md) - Settings system documentation
- [Build Requirements](docs/BUILD_REQUIREMENTS.md) - Platform-specific dependencies
- [Custom Scripts](docs/CUSTOM_SCRIPTS.md) - Writing custom feed scripts
