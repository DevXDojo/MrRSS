# Testing Guide

This document covers testing strategies and patterns for MrRSS.

## Backend Testing (Go)

### Unit Test Pattern

```go
func TestDatabaseOperations(t *testing.T) {
    // Setup test database
    db, cleanup := setupTestDB(t)
    defer cleanup()

    // Test data
    feed := models.Feed{
        Title: "Test Feed",
        URL:   "https://example.com/feed.xml",
    }

    // Execute
    id, err := db.AddFeed(feed)

    // Assert
    if err != nil {
        t.Fatalf("AddFeed failed: %v", err)
    }
    if id == 0 {
        t.Error("Expected non-zero ID")
    }

    // Verify
    retrieved, err := db.GetFeed(id)
    if err != nil {
        t.Fatalf("GetFeed failed: %v", err)
    }
    if retrieved.Title != feed.Title {
        t.Errorf("Expected title %q, got %q", feed.Title, retrieved.Title)
    }
}
```

### Test Helper Functions

```go
func setupTestDB(t *testing.T) (*database.DB, func()) {
    t.Helper()

    // Create temporary database
    tmpFile, err := os.CreateTemp("", "test-*.db")
    if err != nil {
        t.Fatal(err)
    }
    tmpFile.Close()

    // Initialize database
    db, err := database.New(tmpFile.Name())
    if err != nil {
        os.Remove(tmpFile.Name())
        t.Fatal(err)
    }

    // Return cleanup function
    cleanup := func() {
        db.Close()
        os.Remove(tmpFile.Name())
    }

    return db, cleanup
}
```

### Table-Driven Tests

```go
func TestValidateURL(t *testing.T) {
    tests := []struct {
        name    string
        url     string
        wantErr bool
    }{
        {"valid http", "http://example.com", false},
        {"valid https", "https://example.com", false},
        {"missing scheme", "example.com", true},
        {"invalid scheme", "ftp://example.com", true},
        {"empty url", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateURL(tt.url)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Client Testing (XCTest)

### The stub client

Tests drive the view model through `StubAPIClient`, which fails any call the
test did not prepare. An unexpected request is therefore reported rather than
quietly returning nothing.

```swift
final class ListAPIClient: StubAPIClient {
    var articles: [Article] = []

    override func fetchArticles(
        feedID: Int?,
        category: String?,
        filter: String,
        page: Int,
        limit: Int
    ) async throws -> [Article] { articles }
}
```

### View model tests

```swift
@MainActor
final class ArticleListStateTests: XCTestCase {
    override func setUp() {
        super.setUp()
        // Titles are translated, so pin the language the assertions expect.
        Localization.shared.setLanguage(.english)
    }

    func testUnreadFirstKeepsUnreadAtTheTop() {
        let viewModel = AppViewModel(api: ListAPIClient(), autoLoad: false)
        viewModel.articles = [read, unread]

        viewModel.sortOrder = .unreadFirst

        XCTAssertEqual(viewModel.displayedArticles.map(\.id), [unread.id, read.id])
    }
}
```

### Waiting for asynchronous work

A fixed sleep has to guess how long the work takes, and a guess that holds on a
developer's machine fails on a loaded runner. Poll instead:

```swift
try await waitUntil("the multimedia listing to drop read items") {
    viewModel.articles.map(\.id) == [2]
}
```

### Request tests

`MockURLProtocol` serves canned responses, so the API tests never touch the
network and can assert on the request that was built:

```swift
respond("{}") { request, components, query in
    XCTAssertEqual(components.path, "/api/articles/mark-all-read")
    XCTAssertEqual(query["category"], "Tech")
}

try await service.markAllRead(feedID: nil, category: "Tech")
```

### Layout tests

Views that have to lay out correctly are hosted in an `NSHostingView` and
measured, rather than compared against a stored image:

```swift
let hosting = NSHostingView(rootView: FeedEditorView(mode: .add, viewModel: viewModel))
hosting.frame = NSRect(x: 0, y: 0, width: 620, height: 640)
```

## Running Tests

### Backend Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestDatabaseOperations ./internal/database

# Verbose output
go test -v ./...
```

### Client Tests

```bash
# Run all tests
swift test --package-path frontend-swift

# Run with coverage
swift test --package-path frontend-swift --enable-code-coverage

# Run one suite
swift test --package-path frontend-swift --filter ArticleListStateTests
```

## Test Coverage

### Backend Coverage Goals

- Database operations: 80%+
- Handler functions: 70%+
- Business logic: 80%+
- Utility functions: 90%+

### Client Coverage Goals

- Models and decoding: 90%+
- API request building: 80%+
- View model behaviour: 80%+
- Views: layout and structure where it matters, not pixel comparisons

## Continue Reading

- [Architecture Overview](ARCHITECTURE.md)
- [Code Patterns](CODE_PATTERNS.md)
