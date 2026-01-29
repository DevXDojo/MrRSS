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

- **Backend**: Go 1.25+ with Wails v3 (alpha) framework
- **Frontend**: Vue 3.5+ Composition API with TypeScript
- **Database**: SQLite with pure Go implementation (`modernc.org/sqlite`)
- **Communication**: HTTP REST API (not Wails bindings)

### Key Design Principles

1. **Privacy-First**: All data stored locally, no external analytics
2. **Performance-Optimized**: Concurrent processing, intelligent caching, WAL mode SQLite
3. **Modular Architecture**: Feature-based organization, clear separation of concerns
4. **Schema-Driven Configuration**: JSON schema-driven settings system with code generation
5. **Hybrid Communication**: HTTP API for data, Wails bindings for system integration

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

## Frontend Architecture

### Component Organization

Components are organized by feature in `frontend/src/components/`:

```plaintext
components/
├── article/       # Article display and rendering
│   ├── ArticleList.vue
│   ├── ArticleItem.vue
│   ├── ArticleDetail.vue
│   ├── ArticleContent.vue
│   ├── ArticleDetailToolbar.vue
│   ├── ArticleToolbar.vue
│   └── parts/     # Content rendering parts
│       ├── ArticleTitle.vue
│       ├── ArticleSummary.vue
│       ├── ArticleBody.vue
│       ├── ArticleLoading.vue
│       ├── AudioPlayer.vue
│       └── VideoPlayer.vue
├── sidebar/       # Feed list sidebar
│   ├── Sidebar.vue
│   ├── SidebarFeed.vue
│   ├── SidebarCategory.vue
│   └── SidebarNavItem.vue
├── common/        # Reusable components
│   ├── Toast.vue
│   ├── ContextMenu.vue
│   └── ImageViewer.vue
└── modals/        # Modal dialogs
    ├── SettingsModal.vue
    ├── settings/  # Settings tabs
    ├── feed/      # Feed modals
    ├── filter/    # Filter modals
    ├── rules/     # Rules editor
    ├── discovery/ # Discovery modal
    └── common/    # Common modals
```

### Composables Organization

Composables provide reusable logic in `frontend/src/composables/`:

```plaintext
composables/
├── article/       # Article-related logic
│   ├── useArticleDetail.ts
│   ├── useArticleList.ts
│   ├── useArticleContent.ts
│   └── useArticleSummary.ts
├── feed/          # Feed management
│   ├── useFeedManagement.ts
│   └── useFeedRefresh.ts
├── discovery/     # Feed discovery
│   └── useFeedDiscovery.ts
├── filter/        # Article filtering
│   └── useArticleFilter.ts
├── rules/         # Filtering rules
│   └── useRules.ts
├── ui/            # UI utilities
│   ├── useContextMenu.ts
│   ├── useKeyboardShortcuts.ts
│   └── useToast.ts
└── core/          # Core utilities
    └── useSettings.ts
```

### State Management

Pinia store (`frontend/src/stores/app.ts`) manages global state:

- Articles list and selection
- Feeds and categories
- Filter states
- Theme and locale
- Refresh progress
- Unread counts

### Multimedia Support

Enhanced content rendering (`ArticleContent.vue` + `ArticleContent.css`):

- **Images**: Clickable for viewer, right-click context menu, download support
- **Audio**: Full-width player with podcast container styling (`AudioPlayer.vue`)
- **Video**: Responsive player with proper aspect ratio (`VideoPlayer.vue`)
- **Iframes**: 16:9 aspect ratio for YouTube/Vimeo embeds
- **Rich Text**: Tables, blockquotes, code blocks, definition lists

### Translation Integration

Auto-translation features:

- Title translation (on-demand)
- Content paragraph translation (inline display)
- Summary translation
- Supports Google Translate, DeepL, Baidu Translation, and AI-based translation

## Communication Flow

### HTTP API Pattern

Frontend uses direct HTTP fetch (not Wails bindings):

```javascript
// GET request
const response = await fetch('/api/articles');
const articles = await response.json();

// POST request
await fetch('/api/settings', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(data)
});
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

### Frontend

- Virtual scrolling for large article lists
- Debounced operations (search, auto-save)
- Lazy loading of article content
- Efficient state updates with Pinia

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

## Frontend Architecture Details

### Component Communication Patterns

#### Props vs Events

**Best Practice**: Use props for data down, events for actions up

```vue
<!-- Parent -->
<ChildComponent
  :data="parentData"
  @update="handleUpdate"
/>

<!-- Child -->
<script setup>
defineProps<{ data: Type }>()
const emit = defineEmits<{ update: [value: Type] }>()
</script>
```

#### Store Communication

**For Cross-Component State**: Use Pinia store

```typescript
// Access store
const store = useAppStore()

// Read state
const articles = computed(() => store.articles)

// Update state
store.loadArticles()
```

### Multimedia Support

#### Image Handling

- **Lazy Loading**: Images load as needed
- **Click to View**: Opens in image viewer
- **Context Menu**: Right-click for options
- **Download Support**: Save images locally
- **Proxy Caching**: Cached media proxy

#### Audio/Video Support

- **HTML5 Players**: Native browser support
- **Responsive Sizing**: Adapts to content
- **Custom Styling**: Branded player appearance
- **Keyboard Controls**: Space to play/pause

#### Math Rendering

- **KaTeX Integration**: Fast math rendering
- **Inline Math**: `$E = mc^2$`
- **Block Math**: `$$ \int_0^\infty e^{-x^2} dx $$`

#### Code Highlighting

- **highlight.js**: Syntax highlighting
- **Auto-Detection**: Language detection
- **Dark Mode**: Theme-aware highlighting
- **Copy Button**: Easy code copying
