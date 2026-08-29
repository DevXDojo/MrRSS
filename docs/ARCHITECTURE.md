# MrRSS Architecture Documentation

## Table of Contents

- [Overview](#overview)
- [Backend Architecture](#backend-architecture)
- [Frontend Architecture](#frontend-architecture)
- [Communication Flow](#communication-flow)
- [Data Flow](#data-flow)
- [Security Considerations](#security-considerations)
- [Performance Optimizations](#performance-optimizations)
- [Related Documentation](#related-documentation)

## Overview

MrRSS is built with a modern, modular architecture using:

- **Backend**: Go 1.27+ serving an HTTP API
- **Client**: SwiftUI for macOS 14+, a SwiftPM package
- **Database**: SQLite with pure Go implementation (`modernc.org/sqlite`)
- **Communication**: the HTTP REST API alone

### Key Design Principles

1. **Privacy-First**: All data stored locally, no external analytics
2. **Performance-Optimized**: Concurrent processing, intelligent caching, WAL mode SQLite
3. **Modular Architecture**: Feature-based organization, clear separation of concerns
4. **Schema-Driven Configuration**: JSON schema-driven settings system with code generation
5. **One Channel**: everything goes over the HTTP API; system integration is native

## Backend Architecture

### Handler Organization

Handlers are organized by feature domains in `internal/handlers/`:

```plaintext
handlers/
├── ai/            # AI configuration test and AI search
├── article/       # Article CRUD and filtering
├── browser/       # URL opening in browser
├── chat/          # AI chat sessions and messages
├── core/          # Core handler initialization and scheduling
├── custom_css/    # Custom CSS management
├── discovery/     # Feed discovery
├── feed/          # Feed management
├── filter_category/  # Saved filters management
├── freshrss/      # FreshRSS synchronization
├── media/         # Media handling (images, audio, video)
├── network/       # Network detection
├── opml/          # OPML import/export
├── rsshub/        # RSSHub integration
├── rules/         # Filtering rules
├── script/        # Custom script execution
├── settings/      # Settings management
├── statistics/    # Usage statistics
├── summary/       # Article summarization
├── tags/          # Tag management
├── translation/   # Translation services
├── update/        # Application updates
└── window/        # Window management
```

### Core Components

#### Database Layer (`internal/database/`)

- `db.go` - Database initialization and core operations
- `article_db.go` - Article CRUD operations
- `feed_db.go` - Feed CRUD operations
- `settings_db.go` - Key-value settings store
- `cleanup_db.go` - Auto-cleanup logic (preserves favorites)

**Key Features**:

- SQLite with WAL mode for better concurrency
- Prepared statements for all queries
- Indexed queries for performance
- Automatic cleanup with favorite preservation

#### Feed Processing (`internal/feed/`)

- `fetcher.go` - RSS/Atom parsing with `gofeed`, concurrent fetching
- `script_executor.go` - Custom script execution for non-standard feeds
- `article_processor.go` - Article content processing and extraction
- `content_extraction.go` - HTML content extraction utilities
- `http_client.go` - HTTP client with timeout and retry logic
- `intelligent_refresh.go` - Smart feed refresh scheduling
- `progress.go` - Progress tracking for feed operations
- `subscription.go` - Feed subscription management

**Supported Scripts**:

- Python (`.py`)
- Shell (`.sh`)
- PowerShell (`.ps1`)
- Node.js (`.js`)
- Ruby (`.rb`)

#### Discovery System (`internal/discovery/`)

- `feed_discovery.go` - Main discovery orchestration
- `html_parser.go` - HTML parsing for RSS links
- `rss_detector.go` - RSS feed detection logic
- `service.go` - Discovery service with progress tracking

**Features**:

- Discover feeds from URLs
- Batch discovery from friend links
- Real-time progress tracking
- Comprehensive deduplication

#### Summarization (`internal/summary/`)

- `summarizer.go` - TF-IDF and TextRank-based summarization
- `ai_summarizer.go` - AI-based summarization using OpenAI-compatible APIs
- `scoring.go` - Sentence scoring algorithms
- `text_utils.go` - Text processing utilities
- `types.go` - Type definitions for summarization
- `utils.go` - Utility functions for summarization

**Local Algorithms**:

- TF-IDF for term importance
- TextRank for sentence ranking
- Combined scoring (0.5 TF-IDF + 0.5 TextRank)
- Smart sentence selection preserving narrative flow

**AI Summarization**:

- Supports OpenAI-compatible APIs (GPT, Claude, etc.)
- Configurable API endpoint and model
- Token-efficient prompts

#### Translation (`internal/translation/`)

- `translator.go` - Translation interface and factory
- `google.go` - Google Translate (free, no API key)
- `deepl.go` - DeepL API integration
- `baidu.go` - Baidu Translation API integration
- `ai.go` - AI-based translation integration
- `dynamic.go` - Dynamic translation service selection

## Client Architecture

The macOS client lives in `frontend` and is a SwiftPM package with no
third-party dependencies.

### Organisation

```plaintext
frontend/Sources/
├── MrRSSApp.swift            # Scene, menu commands, settings window
├── Localization/
│   ├── LocalizationTables.swift   # The catalogue ported from the previous frontend
│   ├── ClientStrings.swift        # Wording only this client needs
│   └── Localization.swift         # Lookup, fallback, placeholder substitution
├── Models/                   # Codable mirrors of the API payloads
│   ├── Feed.swift, Article.swift, Organization.swift, AI.swift, System.swift
│   ├── FilterFields.swift    # The fields and operators saved filters can use
│   └── SettingsCatalog*.swift     # Generated from the backend schema
├── Services/
│   ├── API/                  # Transport plus one extension per domain
│   ├── KeyboardShortcuts.swift    # Bindings, read from the stored settings
│   └── AppDelegate.swift     # Starts the bundled backend when one is packaged
├── ViewModels/
│   ├── AppViewModel.swift    # Connection, feeds, folders, selection, settings
│   ├── AppViewModel+Articles.swift    # List state, article actions, AI search
│   ├── AppViewModel+Feeds.swift       # Feed, tag, saved-filter and OPML actions
│   └── AppViewModel+Shortcuts.swift   # What each key press does
└── Views/
    ├── Sidebar/              # An NSOutlineView, so dragging behaves natively
    ├── ArticleListView.swift, ArticleRowView.swift
    ├── ArticleDetailView.swift, WebView.swift, ImageGalleryView.swift
    ├── FeedEditorView.swift, SavedFilterEditorView.swift, DiscoveryView.swift
    ├── ArticleChatView.swift
    └── Settings/             # The panes that need more than a generated list
```

### The sidebar is an outline view

`List` cannot reproduce what source lists do while something is dragged over
them: the gap that opens between rows, the way a long list follows the pointer
past its edges, and the folder that lights up underneath. `SidebarOutline`
wraps `NSOutlineView` so all of that comes for free.

### Reading

`WebView` renders either the article text the backend supplied or the original
page, loaded live. Rendered text is stripped of scripts, frames and inline
handlers, and is typeset from the reading settings; a live page keeps its own
scripts because it needs them to display.

### Settings

`SettingsCatalog.generated.swift` is produced from
`internal/config/settings_schema.json` by `tools/settings-swift/generate.py`,
and paired with the wording the previous frontend used. Panes that need more
than a list of controls add their own section.

### Translations

Every interface string is looked up by key through `t(_:)`. The language
follows the `language` setting the backend stores, so a reader who switches it
on one machine sees the same language on another. Dates are formatted with the
matching locale.

## Communication Flow

### HTTP API Pattern

The client talks to the backend through `APIService`:

```swift
// GET request
let articles: [Article] = try await get("articles", queryItems: [
    URLQueryItem(name: "filter", value: "unread")
])

// POST request
try await sendJSONReturningData("settings", body: settings)
```

### Backend Handler Pattern

```go
func (h *Handler) GetArticles(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    feedID := r.URL.Query().Get("feed_id")

    // Database operation
    articles, err := h.DB.GetArticles(feedID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // JSON response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(articles)
}
```

## Data Flow

### Feed Refresh Flow

1. User triggers refresh (manual or scheduled)
2. Backend starts concurrent feed fetching
3. For each feed:
   - Execute script (if custom script) OR fetch RSS/Atom
   - Parse feed with `gofeed`
   - Extract articles
   - Store new articles in database
4. Update progress tracking
5. Frontend polls progress endpoint
6. UI updates with new articles

### Article Display Flow

1. User selects feed/category/filter
2. Frontend fetches articles from API
3. Article list displays with virtual scrolling
4. User selects article
5. Content loaded based on view mode:
   - **Rendered**: Parse and display HTML content
   - **Webpage**: Load original URL in iframe
6. Optional: Generate summary, translate content

### Discovery Flow

1. User initiates discovery (single URL or batch)
2. Backend creates discovery session
3. For each source:
   - Fetch HTML content
   - Parse for RSS links
   - Detect common RSS patterns
   - Verify feeds
4. Progress tracking via polling
5. Frontend displays discovered feeds
6. User selects feeds to import

## Security Considerations

### Input Validation

- URL validation for feeds and websites
- File path validation to prevent directory traversal
- Script path sandboxing within scripts directory

### Safe Operations

- Use `os.Remove()` instead of shell commands
- Prepared SQL statements prevent injection
- No shell command concatenation
- XSS prevention in content rendering

### Script Execution

- Timeout enforcement (30 seconds)
- Working directory restricted to scripts folder
- Path traversal prevention
- Separate execution context per script

## Performance Optimizations

### Database

- SQLite WAL mode for concurrent access
- Indexed columns for frequent queries
- Prepared statement caching
- Periodic VACUUM for space reclamation

### Client

- `LazyVStack` and list reuse for long article lists
- Paged loading, with the next page requested as the last row appears
- Thumbnails decoded at the size they are drawn and cached in memory
- Ordering applied on this Mac, so switching it does not refetch

### Concurrency

- Goroutines for parallel feed fetching
- Background task scheduling
- Progress tracking without blocking
- Graceful timeout handling

## Related Documentation

- [Code Patterns](CODE_PATTERNS.md) - Common coding patterns and examples
- [Testing Guide](TESTING.md) - Testing strategies and examples
- [Custom Scripts](CUSTOM_SCRIPTS.md) - Guide for writing custom feed scripts
- [Settings Guide](SETTINGS.md) - Settings system documentation
- [Build Requirements](BUILD_REQUIREMENTS.md) - Platform-specific build dependencies

## Advanced Features

### AI-Powered Summarization

MrRSS supports two types of summarization:

#### Local Summarization (Offline)

- **TF-IDF Algorithm**: Calculates term importance across articles
- **TextRank Algorithm**: Ranks sentences based on importance
- **Combined Scoring**: 0.5 TF-IDF + 0.5 TextRank for balanced results
- **No API Required**: Works completely offline
- **Smart Sentence Selection**: Preserves narrative flow and coherence

#### AI Summarization (Cloud)

- **OpenAI-Compatible APIs**: Supports GPT, Claude, Gemini, etc.
- **Configurable Endpoint**: Self-hosted or commercial APIs
- **Token-Efficient Prompts**: Optimized for cost-effectiveness
- **Smart Caching**: Avoids redundant API calls

### Smart Translation System

#### Translation Services

1. **Google Translate** (free, no API key required)
2. **DeepL API** (high quality, requires API key)
3. **Baidu Translation** (Chinese language optimized)
4. **AI-Based Translation** (uses configured AI endpoint)

#### Caching Strategy

- **Translation Cache**: Stores all translations in database
- **Automatic Cache Invalidation**: Smart cache management
- **Performance**: Significant speed improvement for repeated content

### Feed Discovery Engine

#### Discovery Methods

1. **URL-Based Discovery**: Extract feeds from any website URL
2. **Batch Discovery**: Process multiple URLs concurrently
3. **Friend Links Discovery**: Crawl friend links for feed discovery
4. **HTML Parsing**: Intelligent RSS link detection

#### Discovery Features

- **Real-Time Progress**: Track discovery status
- **Deduplication**: Automatic duplicate detection
- **Validation**: Verify feeds before adding
- **Concurrent Processing**: Fast batch operations

### Custom Script System

#### Supported Script Types

- **Python** (`.py`) - Cross-platform
- **Shell** (`.sh`) - Linux/macOS only
- **PowerShell** (`.ps1`) - Windows only
- **Node.js** (`.js`) - Cross-platform
- **Ruby** (`.rb`) - Cross-platform

#### Script Execution Security

- **Path Validation**: Prevents directory traversal
- **Timeout Enforcement**: 30-second limit per script
- **Working Directory Restriction**: Scripts run in isolated folder
- **No Shell Concatenation**: Safe command execution
- **Error Capture**: Comprehensive stderr logging

### Filtering Rules Engine

#### Rule Structure

```
IF [condition] THEN [action]
```

#### Conditions

- Feed matches
- Title contains
- Content contains
- Author matches
- Tag matches

#### Actions

- Mark as read/unread
- Mark as favorite/unfavorite
- Hide/show
- Set tag
- Apply label

### Email Newsletter Integration

#### IMAP Support

- **Connection**: Secure IMAP connections
- **Folder Selection**: Choose specific folders to monitor
- **Conversion**: Emails converted to feed articles
- **Attachments**: Handles email attachments

### XPath Scraping

For websites without RSS feeds:

- **XPath Selector**: Target specific content elements
- **Content Extraction**: Clean article content
- **Automatic Detection**: Smart content area detection
- **Fallback Strategies**: Multiple extraction methods

### Image Gallery Mode

#### Visual Browsing

- **Image Extraction**: Pulls images from articles
- **Thumbnail Generation**: Creates optimized thumbnails
- **Gallery View**: Grid-based visual interface
- **Full-Screen Viewer**: Immersive image viewing

### FreshRSS Synchronization

#### Sync Features

- **Bidirectional Sync**: Articles and subscriptions
- **Conflict Resolution**: Intelligent merge strategies
- **Progress Tracking**: Monitor sync status
- **Error Recovery**: Handles network failures gracefully

## Database Optimization

### Performance Features

#### WAL Mode

- **Write-Ahead Logging**: Better concurrent access
- **Read Performance**: Unblocked reads during writes
- **Crash Recovery**: Automatic recovery from crashes

#### Indexing Strategy

```sql
-- Performance indexes
CREATE INDEX idx_articles_feed_id ON articles(feed_id);
CREATE INDEX idx_articles_published_at ON articles(published_at DESC);
CREATE INDEX idx_articles_is_read ON articles(is_read);
CREATE INDEX idx_articles_is_favorite ON articles(is_favorite);
CREATE INDEX idx_articles_hidden ON articles(is_hidden);
```

#### Prepared Statements

- **Query Caching**: Prepared statements cached and reused
- **SQL Injection Prevention**: Automatic parameter escaping
- **Performance**: Faster query execution

#### Connection Pooling

- **Single Connection**: Efficient connection management
- **Mutex Protection**: Thread-safe access
- **WAL Mode**: Enables concurrent reads

### Cleanup Strategy

#### Smart Article Retention

- **Favorites Preservation**: Never deletes favorited articles
- **Per-Feed Limits**: Configurable article limits (default: 15,000)
- **Age-Based Cleanup**: Remove articles older than X days
- **Automatic VACUUM**: Reclaim disk space

## Client Details

### How state reaches the views

`AppViewModel` is the single observable object. Views bind to it directly, and
the feature extensions beside it hold the behaviour, so no view owns state that
another view needs.

For state that belongs to one screen — a sheet's draft, a search field, an
expanded disclosure — use `@State` in that view. Anything the rest of the
interface reacts to belongs on the view model.

### Multimedia Support

#### Image Handling

Images in rendered articles are drawn by the web view, which keeps the original
layout. Every image the backend can extract is also available in a gallery
sheet, with a larger preview and links to open or copy the original.

#### Audio and Video

An article carrying audio or video shows a row above the body offering it in the
system player, so playback uses the machine's own controls and output device.

#### Math Rendering

- **KaTeX Integration**: Fast math rendering
- **Inline Math**: `$E = mc^2$`
- **Block Math**: `$$ \int_0^\infty e^{-x^2} dx $$`

#### Code Highlighting

- **highlight.js**: Syntax highlighting
- **Auto-Detection**: Language detection
- **Dark Mode**: Theme-aware highlighting
- **Copy Button**: Easy code copying
