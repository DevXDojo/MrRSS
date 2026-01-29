# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MrRSS is a modern, cross-platform desktop RSS reader built with **Wails v3** (Go backend + Vue.js frontend). It prioritizes privacy, user experience, and performance with features like auto-translation, smart feed discovery, and AI-powered summarization.

## Development Commands

### Frontend (Vue.js + TypeScript)

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Development server with hot reload
npm run dev

# Build for production
npm run build

# Lint and fix code
npm run lint

# Format code
npm run format

# Run unit tests
npm test

# Run tests with UI
npm run test:ui

# Run E2E tests
npm run test:e2e
```

### Backend (Go)

```bash
# Install dependencies
go mod download

# Run backend tests
go test ./...

# Run tests with coverage
go test -v -timeout=5m -coverprofile=coverage.out -covermode=atomic ./internal/...
go tool cover -html=coverage.out -o coverage.html

# Run tests for specific package
go test -v ./internal/database

# Lint backend
go vet ./...

# Format backend code
gofmt -w .
goimports -w .
```

### Wails Application

```bash
# Development mode with hot reload (recommended)
wails3 dev

# Or using Task runner
task dev

# Build for current platform
wails3 build
# Or
task build

# Package with installer
task package

# Run the built application
task run

# List all available tasks
task --list
```

### Using Makefile (Cross-platform)

```bash
# Show all available commands
make help

# Full development check (lint + test + build)
make check

# Setup development environment
make setup

# Clean build artifacts
make clean

# Platform-specific builds
make build-windows
make build-linux
make build-darwin
```

### Pre-commit Hooks

```bash
# Install hooks
pre-commit install

# Run on all files
pre-commit run --all-files
```

## Architecture Overview

### High-Level Structure

- **Backend**: Go 1.25+ with Wails v3, SQLite database
- **Frontend**: Vue 3.5+ Composition API, Pinia state management, TypeScript
- **Communication**: HTTP API (primary) + Wails bindings (system integration)
- **Build System**: Wails v3 + Task runner + Vite

### Key Directories

```plaintext
MrRSS/
├── main.go                    # Desktop application entry point
├── main-core.go               # Headless server entry point
├── internal/                  # Go backend packages
│   ├── ai/                   # AI configuration and utilities
│   ├── aiusage/              # AI usage tracking and limits
│   ├── cache/                # Media cache management
│   ├── config/               # Configuration with schema-driven generation
│   ├── database/             # SQLite operations with WAL mode
│   ├── discovery/            # Feed discovery engine
│   ├── feed/                 # RSS fetching and processing
│   ├── handlers/             # HTTP API handlers by feature
│   ├── models/               # Core data structures
│   ├── summary/              # TF-IDF + TextRank + AI summarization
│   ├── translation/          # Multi-service translation
│   └── utils/                # Platform utilities
├── frontend/                  # Vue.js frontend
│   ├── src/
│   │   ├── components/       # UI components (article/, sidebar/, modals/)
│   │   ├── composables/      # Vue composables (article/, feed/, ui/)
│   │   ├── stores/           # Pinia state management
│   │   ├── types/            # TypeScript definitions
│   │   └── i18n/             # Internationalization (en, zh)
│   └── dist/                 # Built assets (embedded)
├── docs/                      # Comprehensive documentation
├── tools/                     # Development tools (settings generator)
└── scripts/                   # Automation scripts
```

### Communication Pattern

The application uses a **hybrid approach**:

1. **HTTP API** (`/api/*`) - Primary communication for data operations
2. **Wails Bindings** - System integration (browser, window management)
3. **Static Files** - Frontend assets served from embedded `frontend/dist`

### Database Schema

Core tables:

- `feeds` - RSS subscriptions with metadata
- `articles` - Individual feed items with read/favorite status
- `settings` - Key-value configuration storage
- `translation_cache` - Cached translations for performance

Important: The database uses SQLite with WAL mode for better concurrency.

## Code Patterns and Guidelines

### Backend Patterns

1. **Context Usage**: Always use `context.Context` for exported methods
2. **Error Handling**: Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
3. **Database Operations**: Use prepared statements for all queries
4. **Input Validation**: Validate URLs, file paths, and user inputs
5. **Resource Cleanup**: Use `defer` for proper cleanup

### Frontend Patterns

1. **Vue 3 Composition API**: Use `<script setup>` syntax
2. **State Management**: Access store via `useAppStore()`
3. **Internationalization**: Always use `t()` for user-facing strings
4. **Error Handling**: Show toast notifications with `window.showToast()`
5. **Type Safety**: Use TypeScript with proper type annotations

### Settings Management

**IMPORTANT:** The settings system has been optimized! Adding new settings is now much simpler.

**Instead of manually editing 11+ files, you only need to edit 1 file:**

1. Edit `internal/config/settings_schema.json` to add your setting (5 lines)
2. Run `go run tools/settings-generator/main.go` to generate all boilerplate code
3. Add translations and UI (optional but recommended)

**See [docs/SETTINGS.md](docs/SETTINGS.md) for complete guide.**

**Key Points:**
- **Frontend uses snake_case** (e.g., `settings.ai_api_key`, `settings.update_interval`)
- All generated files are sorted alphabetically for stable diffs
- The generator handles all boilerplate automatically

## Testing

### Backend Tests

- Run tests with timeout: `go test -v -timeout=5m ./...`
- Coverage report: `go test -coverprofile=coverage.out ./...`
- Single test: `go test -v ./internal/database -run TestSpecificFunction`

### Frontend Tests

- Unit tests: `npm test` (uses Vitest)
- E2E tests: `npm run test:e2e` (uses Cypress)
- Test UI: `npm run test:ui`

## Important Notes

1. **No Wails Bindings for Data**: The application primarily uses HTTP API, not Wails bindings for data operations
2. **Privacy-First**: No external analytics, all data stored locally
3. **Cross-Platform**: Build for Windows, macOS, and Linux
4. **Portable Mode**: Supports portable deployment with `portable.txt`
5. **Single Instance**: Enforced on Windows/macOS, disabled on Linux due to D-Bus issues
6. **Concurrent Processing**: Feed fetching uses goroutines with configurable limits

## Common Issues

1. **Linux D-Bus Issues**: Single instance mode disabled on Linux
2. **Build Requirements**: Ensure platform-specific dependencies are installed
3. **Frontend Hot Reload**: Use `wails3 dev` for development with hot reload
4. **Database Migrations**: Handle schema changes carefully with proper versioning

## Quick Reference

### Build Commands
- Development: `wails3 dev`
- Production Build: `wails3 build`
- Important: MrRSS uses HTTP API, not Wails bindings

### Store Access
- `const store = useAppStore()`
- `const { t } = useI18n()`
- Theme: `store.theme` returns `'light'` or `'dark'`
- Language: `store.i18n.locale.value` returns `'en'` or `'zh'`

### UI Helpers
- Toast: `window.showToast(message, type)`
- Confirm: `await window.showConfirm(title, message, isDanger)`

### API Endpoints
- Settings: `GET/POST /api/settings`
- Articles: `GET /api/articles` with query params
- Progress: `GET /api/progress` for async operations

## Related Documentation

- [AGENTS.md](AGENTS.md) - Comprehensive AI agent guidelines
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - System architecture
- [docs/CODE_PATTERNS.md](docs/CODE_PATTERNS.md) - Coding patterns
- [docs/SETTINGS.md](docs/SETTINGS.md) - Settings system
- [docs/TESTING.md](docs/TESTING.md) - Testing guide
- [docs/BUILD_REQUIREMENTS.md](docs/BUILD_REQUIREMENTS.md) - Build dependencies
