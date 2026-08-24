# Issue Fix Verification

Date: 2026-06-09

This document records the urgent and low-effort open issue fixes handled in this pass, plus the verification status for items that cannot be fully reproduced in the current Windows development environment.

## Verified By Automated Checks

- #920 LibreTranslate custom API template error
  - Fixed quoted `{{text}}` placeholder handling and preset request bodies.
  - Covered by `TestCustomTranslatorReplacePlaceholdersAcceptsQuotedTextPlaceholder` and `TestCustomTranslatorReplacePlaceholdersAcceptsRawTextPlaceholder`.

- #902 Turkish translation target option
  - Added `tr` to the translation target selector and English/Chinese labels.
  - Covered by frontend build.

- #911 LAN Ollama/local AI with self-signed TLS
  - AI HTTP clients now use the shared HTTP client and can opt into `MRRSS_INSECURE_SKIP_TLS_VERIFY=true`.
  - Covered by `TestCreateHTTPClientHonorsInsecureTLSVerifyEnv` and `TestCreateHTTPClientKeepsTLSVerificationByDefault`.

- #875 full-text fetch does not use proxy
  - Full-text fetching now uses the same feed/global proxy selection semantics as feed refresh.
  - Covered by `TestCreateArticleHTTPClientUsesFeedProxy` and `TestCreateArticleHTTPClientUsesGlobalProxyWhenFeedRequestsIt`.

- #912 newsletter sender filter does not work
  - IMAP search now includes a `From` header criterion and applies a local sender match fallback.
  - Covered by `TestEmailMatchesSenderFilter`.

- #896 feed management search then select-all selects all feeds
  - Select-all now operates on the currently filtered/sorted visible feed list.
  - Covered by frontend build.

- #802 old deleted articles reappear when cache threshold is small
  - Size cleanup now removes cached article content first and preserves unread/read-state metadata when possible.
  - Covered by `TestCleanupBySizePreservesUnreadMetadataAndDeletesContentFirst`.

- #669 unread view can be empty or fail to load more
  - Normal article pages now support server-side `only_unread=true` pagination. The unread activity page no longer applies an extra client-side unread filter on top of server results.
  - Covered by `TestGetArticlesWithUnreadFilterCombinesWithFavorites`, frontend unit tests, and frontend build.

## Code-Level Fixes With Partial Verification

- #913, #770, #716, #643, #320 macOS/window restore and maximize behavior
  - Window maximize state is now persisted and reapplied on startup, tray reopen, second instance activation, and macOS dock reopen.
  - Verified by Go compilation and existing window handler tests.
  - Not fully verified locally because this environment is Windows and cannot reproduce macOS dock/window-manager behavior.

- #917, #826, #873 old/read articles appearing again or in unread views
  - The size-based cleanup fix preserves article metadata and dedupe keys instead of deleting unread article rows. The unread list now uses server-side unread pagination.
  - Verified by database tests and frontend build.
  - Not fully verified against the reporter feeds because they depend on live RSSHub/GitHub feed update behavior and user-specific local state.

- #918, #909, #772, #767, #750 translation and AI failures
  - Fixed custom translation template quoting, added safer HTTP client creation for AI and translation clients, and added optional TLS bypass for local/self-signed AI endpoints.
  - Verified by unit tests and build.
  - Not fully verified against Tencent Cloud, every custom translation endpoint, or reporter API credentials because those external services and credentials are not available locally.

- #467, #601, #795, #876, #894 full-text extraction, subscription, and encoding edge cases
  - Full-text fetching now sends browser-like headers and uses the configured proxy instead of `readability.FromURL`'s internal default client.
  - Verified by proxy client unit tests and Go compilation.
  - Not fully verified against each reported website/feed because outcomes depend on live site behavior, access policy, encoding, and anti-scraping responses.

## Verification Commands Run

```powershell
go test -v ./internal/utils/httputil ./internal/translation ./internal/handlers/core ./internal/database ./internal/feed
go test -v ./internal/database ./internal/handlers/article
go test -v -timeout=5m ./...
cd frontend; npm run test:unit
cd frontend; npm run build
git diff --check
```

Notes:

- `go test` emits existing SQLite migration warnings about an unavailable `MD5` function in tests, but the test suites pass.
- `npm run build` emits the existing Vite large chunk warning, but the production build succeeds.
