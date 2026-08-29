# Code Patterns for MrRSS

This document provides common coding patterns and best practices for the MrRSS project.

## Table of Contents

- [Code Organization Guidelines](#code-organization-guidelines)
- [Settings Management](#settings-management)
- [Backend Patterns (Go)](#backend-patterns-go)
- [Client Patterns (SwiftUI)](#client-patterns-swiftui)
- [Styling Patterns](#styling-patterns)
- [API Communication](#api-communication)

## Code Organization Guidelines

### File Size

When a file becomes too long (typically over 300-400 lines), consider refactoring:

- **Go**: Extract related functions into separate files within the same package
- **SwiftUI**: Split the body into computed properties, or move behaviour onto the view model
- **TypeScript**: Extract utilities into separate modules

### Folder Organization

When a folder contains too many files (typically over 10-15 files), create subfolders:

- Group by feature or domain (e.g., `handlers/article/`, `handlers/feed/`)
- Keep related tests alongside their source files
- Use index files for clean exports when appropriate

### Build Verification

Before completing any significant change, verify both halves build and their
tests pass:

```bash
make build
make test
```

## Settings Management

**✅ UPDATED**: The settings system has been optimized with schema-driven code generation!

### Quick Method (Recommended)

Adding a new setting now requires **only 1-3 steps**:

#### Step 1: Edit Schema

Edit `internal/config/settings_schema.json`:

```json
"new_feature_enabled": {
  "type": "bool",
  "default": false,
  "category": "general",
  "encrypted": false,
  "frontend_key": "newFeatureEnabled"
}
```

#### Step 2: Generate Code

```bash
go run tools/settings-generator/main.go
```

This automatically generates:

- Backend types and handlers
- Frontend types and composables
- Database initialization keys
- Default values

#### Step 3: Add UI (Optional)

Add to your settings component:

```vue
<SettingItem :title="t('newFeatureEnabled')">
  <Toggle
    :model-value="settings.newFeatureEnabled"
    @update:model-value="updateSetting('newFeatureEnabled', $event)"
  />
</SettingItem>
```

### Complete Documentation

See **[docs/SETTINGS.md](SETTINGS.md)** for:

- Complete step-by-step guide
- Detailed examples
- Troubleshooting tips
- Best practices

### Verification

After running the generator, verify:

```bash
# Backend
go build ./...
go test ./internal/config

# Client, after running python3 tools/settings-swift/generate.py
swift build --package-path frontend-swift
swift test --package-path frontend-swift --filter SettingsCatalogTests
```

### Legacy Method (Deprecated)

The old manual method of editing 8+ files is **deprecated**. All new settings should use the schema-driven approach above.

If you need to maintain old manually-written settings, refer to the git history for the old checklist, but note that it's highly error-prone and should be avoided.

## Backend Patterns (Go)

### Handler Method Pattern

Always use `context.Context` for exported methods and proper error wrapping:

```go
func (h *Handler) GetArticles(ctx context.Context, feedID int) ([]models.Article, error) {
    if feedID < 0 {
        return nil, errors.New("invalid feed ID")
    }

    articles, err := h.DB.GetArticles(ctx, feedID)
    if err != nil {
        return nil, fmt.Errorf("failed to get articles: %w", err)
    }

    return articles, nil
}
```

**Key Points**:

- Use `context.Context` as first parameter
- Validate inputs early
- Wrap errors with `fmt.Errorf` and `%w`
- Return zero values and errors, not panics

### Database Query Pattern

Always use prepared statements with proper cleanup:

```go
func (db *DB) GetArticlesByFeed(feedID int, isRead bool) ([]models.Article, error) {
    // Prepare statement
    stmt, err := db.conn.Prepare(`
        SELECT id, title, url, content, published_at, is_read, is_favorite
        FROM articles
        WHERE feed_id = ? AND is_read = ?
        ORDER BY published_at DESC
    `)
    if err != nil {
        return nil, fmt.Errorf("prepare statement: %w", err)
    }
    defer stmt.Close()

    // Execute query
    rows, err := stmt.Query(feedID, isRead)
    if err != nil {
        return nil, fmt.Errorf("execute query: %w", err)
    }
    defer rows.Close()

    // Scan results
    var articles []models.Article
    for rows.Next() {
        var article models.Article
        err := rows.Scan(
            &article.ID,
            &article.Title,
            &article.URL,
            &article.Content,
            &article.PublishedAt,
            &article.IsRead,
            &article.IsFavorite,
        )
        if err != nil {
            return nil, fmt.Errorf("scan row: %w", err)
        }
        articles = append(articles, article)
    }

    // Check for iteration errors
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate rows: %w", err)
    }

    return articles, nil
}
```

**Key Points**:

- Use prepared statements for all queries
- Always `defer Close()` on statements and rows
- Scan into proper types
- Check `rows.Err()` after iteration
- Use proper error wrapping

### Settings Management Pattern

Settings are stored as key-value strings in the database:

```go
// Get setting with default value
func (db *DB) GetSetting(key string, defaultValue string) string {
    var value string
    err := db.conn.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
    if err == sql.ErrNoRows {
        return defaultValue
    }
    if err != nil {
        log.Printf("Error getting setting %s: %v", key, err)
        return defaultValue
    }
    return value
}

// Set setting
func (db *DB) SetSetting(key, value string) error {
    _, err := db.conn.Exec(`
        INSERT INTO settings (key, value)
        VALUES (?, ?)
        ON CONFLICT(key) DO UPDATE SET value = excluded.value
    `, key, value)
    if err != nil {
        return fmt.Errorf("set setting %s: %w", key, err)
    }
    return nil
}

// Get boolean setting
func (db *DB) GetBoolSetting(key string, defaultValue bool) bool {
    value := db.GetSetting(key, "")
    if value == "" {
        return defaultValue
    }
    return value == "true" || value == "1"
}

// Get integer setting
func (db *DB) GetIntSetting(key string, defaultValue int) int {
    value := db.GetSetting(key, "")
    if value == "" {
        return defaultValue
    }
    intValue, err := strconv.Atoi(value)
    if err != nil {
        return defaultValue
    }
    return intValue
}
```

### Cleanup Logic Pattern

Auto-cleanup preserves favorites:

```go
func (db *DB) CleanupOldArticles(maxAgeDays int) (int64, error) {
    cutoffDate := time.Now().AddDate(0, 0, -maxAgeDays)

    // IMPORTANT: Delete old articles EXCEPT favorites
    result, err := db.conn.Exec(`
        DELETE FROM articles
        WHERE published_at < ? AND is_favorite = 0
    `, cutoffDate)

    if err != nil {
        return 0, fmt.Errorf("cleanup articles: %w", err)
    }

    // Run VACUUM to reclaim space
    _, _ = db.conn.Exec("VACUUM")

    return result.RowsAffected()
}
```

**Critical**: Always exclude favorites (`is_favorite = 0`) in cleanup queries.

### Concurrent Processing Pattern

Use goroutines for parallel operations with proper error handling:

```go
func (f *Fetcher) FetchFeeds(feeds []models.Feed) map[int]error {
    errors := make(map[int]error)
    var mu sync.Mutex
    var wg sync.WaitGroup

    // Limit concurrent fetches
    semaphore := make(chan struct{}, 5) // Max 5 concurrent

    for _, feed := range feeds {
        wg.Add(1)
        go func(feed models.Feed) {
            defer wg.Done()

            // Acquire semaphore
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            // Fetch with timeout
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()

            err := f.FetchFeed(ctx, feed)
            if err != nil {
                mu.Lock()
                errors[feed.ID] = err
                mu.Unlock()
            }
        }(feed)
    }

    wg.Wait()
    return errors
}
```

**Key Points**:

- Use `sync.WaitGroup` to wait for goroutines
- Use semaphore to limit concurrency
- Use `sync.Mutex` for shared state
- Always use context with timeout
- Capture loop variables properly

### Script Execution Pattern

Execute scripts safely with timeout and path validation:

```go
func (e *ScriptExecutor) ExecuteScript(ctx context.Context, scriptPath string) (*gofeed.Feed, error) {
    // Construct full path
    fullPath := filepath.Join(e.scriptsDir, scriptPath)
    fullPath = filepath.Clean(fullPath)
    cleanScriptsDir := filepath.Clean(e.scriptsDir)

    // Security: prevent directory traversal
    relPath, err := filepath.Rel(cleanScriptsDir, fullPath)
    if err != nil || strings.HasPrefix(relPath, "..") {
        return nil, fmt.Errorf("invalid script path: must be within scripts directory")
    }

    // Create context with timeout
    execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Determine command based on extension
    var cmd *exec.Cmd
    ext := strings.ToLower(filepath.Ext(fullPath))

    switch ext {
    case ".py":
        pythonCmd := "python3"
        if runtime.GOOS == "windows" {
            pythonCmd = "python"
        }
        cmd = exec.CommandContext(execCtx, pythonCmd, fullPath)
    case ".sh":
        if runtime.GOOS == "windows" {
            return nil, fmt.Errorf("shell scripts not supported on Windows")
        }
        cmd = exec.CommandContext(execCtx, "bash", fullPath)
    case ".js":
        cmd = exec.CommandContext(execCtx, "node", fullPath)
    default:
        cmd = exec.CommandContext(execCtx, fullPath)
    }

    // Set working directory
    cmd.Dir = e.scriptsDir

    // Capture output
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    // Execute
    if err := cmd.Run(); err != nil {
        if stderr.Len() > 0 {
            return nil, fmt.Errorf("script failed: %v, stderr: %s", err, stderr.String())
        }
        return nil, fmt.Errorf("script failed: %v", err)
    }

    // Parse output as RSS feed
    parser := gofeed.NewParser()
    feed, err := parser.ParseString(stdout.String())
    if err != nil {
        return nil, fmt.Errorf("parse feed output: %w", err)
    }

    return feed, nil
}
```

**Security Checklist**:

- ✅ Path validation to prevent directory traversal
- ✅ Timeout enforcement (30 seconds)
- ✅ Working directory restriction
- ✅ No shell command concatenation
- ✅ Proper error handling with stderr capture

### Input Validation Pattern

Always validate user inputs, especially URLs and file paths:

```go
// Validate feed URL
func validateFeedURL(urlStr string) error {
    u, err := url.Parse(urlStr)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }

    if u.Scheme != "http" && u.Scheme != "https" {
        return errors.New("URL must use HTTP or HTTPS")
    }

    if u.Host == "" {
        return errors.New("URL must have a host")
    }

    return nil
}

// Validate file path to prevent directory traversal
func validateFilePath(baseDir, filePath string) error {
    cleanPath := filepath.Clean(filePath)
    cleanBase := filepath.Clean(baseDir)

    if !strings.HasPrefix(cleanPath, cleanBase) {
        return errors.New("invalid file path: path traversal detected")
    }

    return nil
}
```

### HTTP Handler Pattern

Standard HTTP handler with JSON response:

```go
func (h *Handler) HandleGetArticles(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    feedIDStr := r.URL.Query().Get("feed_id")
    feedID, err := strconv.Atoi(feedIDStr)
    if err != nil {
        http.Error(w, "invalid feed_id", http.StatusBadRequest)
        return
    }

    // Call service layer
    articles, err := h.DB.GetArticles(feedID)
    if err != nil {
        log.Printf("Error getting articles: %v", err)
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }

    // Return JSON response
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(articles); err != nil {
        log.Printf("Error encoding response: %v", err)
    }
}
```

**Key Points**:

- Validate inputs from query params/body
- Return appropriate HTTP status codes
- Set proper Content-Type header
- Log errors (don't expose to client)
- Use `http.Error` for error responses

## Client Patterns (SwiftUI)

### View structure

Keep bodies small. SwiftUI type-checks a body as one expression, and a long one
can stall the build for minutes. Split it into computed properties named after
what they show.

```swift
struct ArticleListView: View {
    @ObservedObject var viewModel: AppViewModel
    @State private var isConfirmingMarkAllRead = false

    var body: some View {
        VStack(spacing: 0) {
            if isShowingSearch {
                searchBar
                Divider()
            }
            list
        }
        .navigationTitle(viewModel.articleListTitle)
        .toolbar { toolbarContent }
    }

    private var searchBar: some View { /* ... */ }
    private var list: some View { /* ... */ }

    @ToolbarContentBuilder
    private var toolbarContent: some ToolbarContent { /* ... */ }
}
```

### Where state lives

`AppViewModel` is the one observable object. State the rest of the interface
reacts to belongs there; state a single screen owns belongs in `@State`.

```swift
// On the view model: the list and the reading pane both read this
@Published var selectedArticleID: Int?

// In the view: only this sheet cares
@State private var isConfirmingClearReadLater = false
```

The view model is split by feature across files:

- `AppViewModel.swift` — connection, feeds, folders, selection, settings
- `AppViewModel+Articles.swift` — list state, article actions, AI search
- `AppViewModel+Feeds.swift` — feed, tag, saved-filter and OPML actions
- `AppViewModel+Shortcuts.swift` — what each key press does

Properties an extension writes to are declared without `private(set)`, because
Swift scopes that to the file.

### Optimistic updates

Apply the change locally, then roll it back if the request fails:

```swift
private func mutateArticle(
    _ id: Int,
    apply change: (inout Article) -> Void,
    request: @escaping () async throws -> Void
) {
    guard let index = articles.firstIndex(where: { $0.id == id }) else { return }
    let previous = articles[index]
    change(&articles[index])

    Task { [weak self] in
        guard let self else { return }
        do {
            try await request()
        } catch {
            if let currentIndex = articles.firstIndex(where: { $0.id == id }) {
                articles[currentIndex] = previous
            }
            errorMessage = error.localizedDescription
        }
    }
}
```

### Cancelling superseded work

A request whose result is no longer wanted must not overwrite newer state. Each
load carries an identifier that is checked before anything is assigned:

```swift
articleTask?.cancel()
articleRequestID = UUID()
let requestID = articleRequestID

articleTask = Task { [weak self] in
    let loaded = try await load(query: query, page: page)
    try Task.checkCancellation()
    guard requestID == self?.articleRequestID else { return }
    self?.articles = loaded
}
```

### Interface strings

Every string goes through `t(_:)`, never a literal:

```swift
Label(t("article.action.markAllRead"), systemImage: "checkmark.circle")
Text(t("article.action.markedNArticlesAsRead", ["count": affected]))
```

Keys ported from the previous frontend live in `LocalizationTables.swift`;
wording only this client needs lives in `ClientStrings.swift` under a `client.`
prefix. The language follows the `language` setting the backend stores.

### Decoding

Every model decodes leniently, because an older backend omits fields a newer one
sends:

```swift
init(from decoder: Decoder) throws {
    let container = try decoder.container(keyedBy: CodingKeys.self)
    id = try container.decode(Int.self, forKey: .id)
    title = try container.decodeIfPresent(String.self, forKey: .title) ?? ""
    isRead = try container.decodeIfPresent(Bool.self, forKey: .isRead) ?? false
}
```

Several endpoints answer `null` instead of `[]`, which `APIService.decode`
turns into an empty collection.

### When AppKit is the right answer

Reach for AppKit where SwiftUI cannot reproduce platform behaviour:

- **The sidebar** is an `NSOutlineView`, because `List` cannot show the gap that
  opens between rows during a drag, the edge scrolling, or the folder lighting
  up underneath.
- **Article content** is a `WKWebView`, so the original markup renders.
- **File import and export** use `NSOpenPanel` and `NSSavePanel`.
- **Links** open through `NSWorkspace`.

Wrap these in `NSViewRepresentable` and keep the SwiftUI surface small.

## API Communication

### Frontend API Calls

MrRSS uses direct HTTP fetch (not Wails bindings) for better control.

#### GET request

```swift
// Simple
let feeds: [Feed] = try await get("feeds")

// With query parameters
let articles: [Article] = try await get("articles", queryItems: [
    URLQueryItem(name: "feed_id", value: "123"),
    URLQueryItem(name: "filter", value: "unread"),
    URLQueryItem(name: "limit", value: "50")
])
```

#### POST request

```swift
// With a JSON body built from a dictionary
try await post("feeds/add", jsonBody: draft.jsonBody)

// With an Encodable body and a decoded response
let result: SummaryResult = try await postJSON(
    "articles/summarize",
    body: Request(articleID: id, length: "medium", content: nil)
)

// With query parameters only
try await post("articles/favorite", queryItems: [
    URLQueryItem(name: "id", value: String(id))
])
```

#### Error handling

`APIService.data(for:)` turns a non-2xx response into `APIError.server`,
preferring the `error` field of a JSON body over its raw text. Callers surface
the message and, where the change was optimistic, roll it back:

```swift
do {
    try await api.refreshFeed(id: feed.id)
    statusMessage = t("modal.feed.feedRefreshStarted")
} catch {
    errorMessage = error.localizedDescription
}
```

### Progress Tracking

Long-running operations report progress by polling, so a run started elsewhere
is still followed:

```swift
try await api.refreshAllFeeds()

for _ in 0..<180 {
    try Task.checkCancellation()
    let progress = try await api.fetchRefreshProgress()
    refreshProgress = progress
    if !progress.isRunning { break }
    try await Task.sleep(for: .seconds(1))
}
```

Discovery works the same way, through `fetchDiscoveryProgress()`.

### Backend HTTP Handlers

Standard pattern for HTTP handlers:

```go
func (h *Handler) HandleGetArticles(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    feedIDStr := r.URL.Query().Get("feed_id")
    feedID, err := strconv.Atoi(feedIDStr)
    if err != nil {
        http.Error(w, "invalid feed_id", http.StatusBadRequest)
        return
    }

    // Call service/database
    articles, err := h.DB.GetArticles(feedID)
    if err != nil {
        log.Printf("Error: %v", err)
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    // Return JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(articles)
}
```

### API Endpoints

#### Articles

- `GET /api/articles` - List articles (with filters)
- `GET /api/articles/:id` - Get single article
- `POST /api/articles/:id/read` - Mark as read
- `POST /api/articles/:id/favorite` - Toggle favorite

#### Feeds

- `GET /api/feeds` - List all feeds
- `POST /api/feeds` - Add new feed
- `PUT /api/feeds/:id` - Update feed
- `DELETE /api/feeds/:id` - Delete feed
- `POST /api/feeds/:id/refresh` - Refresh single feed
- `POST /api/refresh` - Refresh all feeds

#### Settings

- `GET /api/settings` - Get all settings
- `POST /api/settings` - Save settings

#### Discovery

- `POST /api/discovery/single` - Discover from URL
- `POST /api/discovery/batch` - Batch discovery
- `GET /api/discovery/progress` - Get discovery progress

#### Media

- `GET /api/media/proxy` - Proxy cached media content

#### Window

- `GET /api/window/state` - Get saved window state
- `POST /api/window/state` - Save window state

#### Summary

- `POST /api/summary` - Generate article summary

#### Translation

- `POST /api/translate` - Translate text
