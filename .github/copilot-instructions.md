# GitHub Copilot Instructions for MrRSS

> **Documentation**: [AGENTS.md](../AGENTS.md) | [Architecture](../docs/ARCHITECTURE.md) | [Code Patterns](../docs/CODE_PATTERNS.md) | [Testing](../docs/TESTING.md) | [Build Requirements](../docs/BUILD_REQUIREMENTS.md)

## Project Context

MrRSS is a privacy-focused RSS reader. On this branch the interface is a native macOS client in SwiftUI, talking to a Go backend over its HTTP API.

**Core Principles**: Privacy-first, cross-platform, modern UI, high performance, accessible

## Tech Stack

- **Backend**: Go 1.27+, SQLite with `modernc.org/sqlite`, serving only `/api`
- **Client**: SwiftUI for macOS 14+, a SwiftPM package with no third-party dependencies
- **Tools**: Swift toolchain from Xcode, Go modules
- **Icons**: Phosphor Icons | **I18n**: vue-i18n (English/Chinese)

## Quick Patterns Reference

### Backend (Go)

**Key Principles**:

- Always use `context.Context` for exported methods
- Error wrapping with `fmt.Errorf("operation failed: %w", err)`
- Prepared statements for all database queries
- Proper cleanup with `defer`
- Input validation before processing

📚 **Full Patterns**: See [CODE_PATTERNS.md](../docs/CODE_PATTERNS.md#backend-patterns-go)

### Client (SwiftUI)

When writing a view, follow this shape:

```swift
struct ArticleRowView: View {
    let article: Article
    @ObservedObject var viewModel: AppViewModel

    var body: some View {
        HStack(alignment: .top, spacing: 10) {
            unreadDot
            VStack(alignment: .leading, spacing: 4) {
                // Every interface string comes from the catalogue.
                Text(article.displayTitle(preferTranslation: showsTranslation))
                    .fontWeight(article.isRead ? .regular : .semibold)
                metadata
            }
        }
        .contextMenu { contextMenu }
    }

    private var unreadDot: some View { /* ... */ }
    private var metadata: some View { /* ... */ }

    @ViewBuilder
    private var contextMenu: some View {
        Button {
            viewModel.setArticleRead(article, read: !article.isRead)
        } label: {
            Label(t("article.action.markAsRead"), systemImage: "checkmark.circle")
        }
    }
}
```

Behaviour belongs on `AppViewModel` or one of its extensions, not in the view:

```swift
extension AppViewModel {
    func toggleReadLater(_ article: Article) {
        mutateArticle(article.id, apply: { $0.isReadLater.toggle() }) { [weak self] in
            try await self?.api.toggleReadLater(id: article.id)
        }
    }
}
```

## Don'ts

❌ **Don't**:

- Hardcode user-facing strings (always use `t("some.key")`)
- Write long SwiftUI bodies (type checking stalls; split into computed properties)
- Force unwrap outside tests
- Forget error handling in async operations
- Commit API keys, secrets, or sensitive data
- Render untrusted markup without sanitising it (see `HTMLDocument.sanitize`)
- Make breaking changes without a migration path
- Use shell command concatenation (security risk)
- Let a superseded request overwrite newer state (check the request identifier)
- Forget to remove time observers and event monitors when a view disappears
- Delete favorited articles during cleanup operations
- Block the main actor with long-running work

## Do's

✅ **Do**:

- Use TypeScript with proper type annotations
- Follow existing code patterns and conventions
- Write comprehensive tests for new features
- Keep functions small and focused (single responsibility)
- Use meaningful variable and function names
- Handle all edge cases and error conditions
- Validate inputs thoroughly (URLs, file paths, user data)
- Log errors with appropriate context for debugging
- Use semantic HTML with proper ARIA attributes
- Debounce frequent operations (auto-save, search, etc.)
- Use `os.Remove()` instead of shell commands for file operations
- Clean up resources (timers, goroutines, event listeners) properly
- Preserve favorited articles during any cleanup operation
- Use prepared statements for all database queries
- Implement proper loading states and progress indicators
- Follow semantic versioning (MAJOR.MINOR.PATCH)
- Document exported functions and complex logic
- Use goroutines for concurrent operations
- Implement graceful degradation for network failures

## Quick Reference

**Build Commands**:

- Development: `./frontend-swift/run.sh`
- Client only: `swift build --package-path frontend-swift`
- Release bundle: `make build-app VERSION=1.3.28`

**State Access**:

- `@ObservedObject var viewModel: AppViewModel`
- Strings: `t("some.key")`, `t("some.key", ["count": n])`
- Settings: `viewModel.setting("key")`, `viewModel.boolSetting("key")`
- Theme follows the `theme` setting through `viewModel.preferredColorScheme`

**UI Helpers**:

- Confirmation: `viewModel.statusMessage = t("...")` shows a short message
- Errors: `viewModel.errorMessage = ...` raises an alert
- Confirmations use `.confirmationDialog`

**API Endpoints**:

- Settings: `GET/POST /api/settings`
- Articles: `GET /api/articles` with query params
- Progress: `GET /api/progress` for async operations

---

When generating code, prioritize:

1. **Correctness**: Code that works and handles errors properly
2. **Consistency**: Follow existing patterns in the codebase
3. **Clarity**: Easy to understand and maintain
4. **Performance**: Efficient queries, minimal re-renders, proper cleanup
5. **Security**: Input validation, safe file operations, no injection vulnerabilities
6. **User Experience**: Loading states, progress indicators, error messages
