---
name: mrrss-assistant
description: Control and analyze a running MrRSS desktop or server-mode instance through its local REST API. Use when a user asks an AI agent to inspect feeds, search or summarize articles, diagnose feed health, manage read/favorite/read-later state, refresh feeds, export data, review statistics, or configure MrRSS using its API.
---

# MrRSS Assistant

## Overview

Use the MrRSS HTTP API as the source of truth for user data. Prefer read-only inspection first, ask before destructive operations, and keep outputs concise enough for a user to act on.

## Connection Workflow

1. Determine the base URL. Default to `http://localhost:1234` for desktop and server mode unless the user provides another host or port. Released desktop builds listen on the loopback interface only while MrRSS is running.
2. Set the API root to `{base_url}/api`.
3. Verify connectivity with `GET {api_root}/version`, then `GET {api_root}/feeds`.
4. If both fail, explain that MrRSS or `mrrss-server` must be running. For the desktop app, also note that another process may be occupying port `1234`. Stop before attempting data operations.
5. Use `references/api.md` for supported endpoints and request shapes. If the repo is available and API details may have changed, regenerate it with `scripts/generate_api_reference.py`.

## Safety

- Treat `GET` requests as safe.
- Ask for explicit confirmation before deleting feeds, clearing caches, resetting statistics, clearing summaries/translations, or bulk-marking articles.
- For settings changes, first read `GET /settings`, show the exact keys that will change, then submit only after confirmation.
- Never expose API keys, passwords, custom headers, IMAP credentials, or FreshRSS credentials in final answers. Redact sensitive settings when reporting configuration.
- Avoid direct SQLite access unless the API is unavailable and the user explicitly asks for offline inspection.

## Common Tasks

### Inspect Feeds

Use `GET /feeds` to list feeds. Report feed titles, categories, unread counts, last error fields when present, and stale feeds. For health checks, follow with targeted refresh only after user approval if many feeds are involved.

### Find Articles

Use `GET /articles` with `filter`, `feed_id`, `category`, `page`, and `limit`. For user-facing summaries, fetch only the pages required and cite article titles, feeds, and dates. Use `GET /articles/content?id={id}` when the list response lacks enough text.

### Summarize Or Translate

Prefer MrRSS endpoints so user settings and cached results are respected:

- `POST /summarize` for article summaries.
- `POST /translate/text` for ad hoc text translation.
- `POST /translate/article` for article translation.

If an AI endpoint fails, check `GET /ai/test/info` and available profiles before asking the user to reconfigure credentials.

### Manage Article State

Use the dedicated article endpoints for read, favorite, hidden, and read-later state. Confirm before bulk operations such as mark-all-read or clearing read-later.

### Analyze Reading Activity

Use `GET /api/statistics`, `/api/statistics/all-time`, and `/api/statistics/available-months`. Present trends as plain language, not only raw counts.

## API Reference

Read `references/api.md` when you need endpoint details. It is generated from `docs/SERVER_MODE/swagger.json` in the MrRSS repository.

To regenerate:

```bash
python skills/mrrss-assistant/scripts/generate_api_reference.py docs/SERVER_MODE/swagger.json skills/mrrss-assistant/references/api.md
```
