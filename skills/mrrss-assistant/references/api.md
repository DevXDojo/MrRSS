# MrRSS API Reference

Generated from `docs/SERVER_MODE/swagger.json`. Regenerate with:

```bash
python skills/mrrss-assistant/scripts/generate_api_reference.py docs/SERVER_MODE/swagger.json skills/mrrss-assistant/references/api.md
```

- API version: `1.3.26`
- API root: `{base_url}/api`
- Endpoint paths below are relative to the API root unless they already start with `/api/`.

## Usage Notes

- Use `GET` endpoints freely for inspection.
- Ask before `DELETE`, bulk updates, cache clearing, or settings changes.
- Send JSON request bodies with `Content-Type: application/json` unless an endpoint describes file upload.
- Redact credentials and API keys from user-facing output.

## Ai

### `GET /ai/profiles`

List AI profiles

### `POST /ai/profiles`

Create AI profile

Parameters:
  - `request` (body, required): Profile data

Request body: see the Swagger schema for full field details.

### `POST /ai/profiles/test-all`

Test all AI profiles

### `POST /ai/profiles/test-config`

Test AI configuration without saving

Parameters:
  - `request` (body, required): Configuration to test

Request body: see the Swagger schema for full field details.

### `DELETE /ai/profiles/{id}`

Delete AI profile

Parameters:
  - `id` (path, required): Profile ID

### `GET /ai/profiles/{id}`

Get AI profile

Parameters:
  - `id` (path, required): Profile ID

### `PUT /ai/profiles/{id}`

Update AI profile

Parameters:
  - `id` (path, required): Profile ID
  - `request` (body, required): Profile data

Request body: see the Swagger schema for full field details.

### `POST /ai/profiles/{id}/default`

Set default AI profile

Parameters:
  - `id` (path, required): Profile ID

### `POST /ai/profiles/{id}/test`

Test AI profile

Parameters:
  - `id` (path, required): Profile ID

### `POST /ai/search`

AI-powered article search

Parameters:
  - `request` (body, required): Search query

Request body: see the Swagger schema for full field details.

### `POST /ai/test`

Test AI configuration

### `GET /ai/test/info`

Get AI test info

### `GET /ai/usage`

Get AI usage statistics

### `POST /ai/usage/reset`

Reset AI usage counter

## Articles

### `GET /articles`

Get articles with filtering

Parameters:
  - `filter` (query, optional): Filter: 'all', 'unread', 'favorite', 'read_later'
  - `feed_id` (query, optional): Filter by feed ID
  - `category` (query, optional): Filter by category name
  - `page` (query, optional): Page number (default: 1)
  - `limit` (query, optional): Items per page (default: 50, max: 500)

### `POST /articles/cleanup`

Cleanup all articles

### `POST /articles/cleanup-content`

Cleanup article content cache

### `POST /articles/clear-read-later`

Clear read-later list

### `GET /articles/content`

Get article content

Parameters:
  - `id` (query, required): Article ID

### `GET /articles/content-cache-info`

Get article content cache info

### `POST /articles/export/notion`

Export article to Notion

Parameters:
  - `request` (body, required): Article export request

Request body: see the Swagger schema for full field details.

### `POST /articles/export/obsidian`

Export article to Obsidian

Parameters:
  - `request` (body, required): Article export request

Request body: see the Swagger schema for full field details.

### `POST /articles/export/zotero`

Export article to Zotero

Parameters:
  - `request` (body, required): Article export request

Request body: see the Swagger schema for full field details.

### `GET /articles/extract-images`

Extract all images from article

Parameters:
  - `id` (query, required): Article ID

### `POST /articles/fetch-full`

Fetch full article content

Parameters:
  - `id` (query, required): Article ID

### `POST /articles/filter`

Get filtered articles

Parameters:
  - `request` (body, required): Filter criteria

Request body: see the Swagger schema for full field details.

### `GET /articles/filter-counts`

Get filter-specific feed counts

### `GET /articles/image-gallery`

Get image gallery articles

Parameters:
  - `feed_id` (query, optional): Filter by feed ID
  - `category` (query, optional): Filter by category name
  - `only_unread` (query, optional): Filter for only unread articles
  - `page` (query, optional): Page number (default: 1)
  - `limit` (query, optional): Items per page (default: 50)

### `POST /articles/mark-all-read`

Mark all articles as read

Parameters:
  - `feed_id` (query, optional): Mark all as read for specific feed ID
  - `category` (query, optional): Mark all as read for specific category

### `POST /articles/mark-read-sync`

Mark article as read/unread with immediate FreshRSS sync

Parameters:
  - `id` (query, required): Article ID
  - `read` (query, required): Read status: 'true', '1', 'false', or '0'

### `POST /articles/mark-relative`

Mark articles relative to reference article

Parameters:
  - `id` (query, required): Reference article ID
  - `direction` (query, required): Direction: 'above' for newer articles, 'below' for older articles
  - `feed_id` (query, optional): Optional: only mark articles from this feed
  - `category` (query, optional): Optional: only mark articles from this category

### `POST /articles/refresh`

Refresh all feeds

### `POST /articles/toggle-favorite-sync`

Toggle article favorite status with immediate FreshRSS sync

Parameters:
  - `id` (query, required): Article ID

### `POST /articles/toggle-hide`

Toggle article hidden status

Parameters:
  - `id` (query, required): Article ID

### `POST /articles/toggle-read-later`

Toggle article read-later status

Parameters:
  - `id` (query, required): Article ID

### `GET /articles/unread-counts`

Get unread counts

## Chat

### `POST /chat`

AI chat with article

Parameters:
  - `request` (body, required): Chat request (messages, article info)

Request body: see the Swagger schema for full field details.

### `DELETE /chat/message`

Delete chat message

Parameters:
  - `message_id` (query, required): Message ID

### `GET /chat/messages`

List chat messages

Parameters:
  - `session_id` (query, required): Session ID

### `DELETE /chat/session`

Delete chat session

Parameters:
  - `session_id` (query, required): Session ID

### `GET /chat/session`

Get chat session

Parameters:
  - `session_id` (query, required): Session ID

### `PUT /chat/session`

Update chat session

Parameters:
  - `session_id` (query, required): Session ID
  - `request` (body, required): Update request (title)

Request body: see the Swagger schema for full field details.

### `GET /chat/sessions`

List chat sessions

Parameters:
  - `article_id` (query, required): Article ID

### `POST /chat/sessions`

Create chat session

Parameters:
  - `request` (body, required): Session creation request (article_id, title)

Request body: see the Swagger schema for full field details.

### `DELETE /chat/sessions/all`

Delete all chat sessions

## Discovery

### `POST /discovery/all`

Discover feeds from all subscriptions

### `POST /discovery/batch/clear`

Clear batch discovery state

### `GET /discovery/batch/progress`

Get batch discovery progress

### `POST /discovery/batch/start`

Start batch discovery

### `POST /discovery/blogs`

Discover blogs from feed

Parameters:
  - `request` (body, required): Discovery request (feed_id)

Request body: see the Swagger schema for full field details.

### `POST /discovery/single/clear`

Clear single discovery state

### `GET /discovery/single/progress`

Get single discovery progress

### `POST /discovery/single/start`

Start single feed discovery

Parameters:
  - `request` (body, required): Discovery request (feed_id)

Request body: see the Swagger schema for full field details.

## Feeds

### `GET /feeds`

Get all feeds

### `POST /feeds/add`

Add a new feed

Parameters:
  - `request` (body, required): Feed details

Request body: see the Swagger schema for full field details.

### `POST /feeds/delete`

Delete a feed

Parameters:
  - `id` (query, required): Feed ID

### `POST /feeds/refresh`

Refresh a single feed

Parameters:
  - `id` (query, required): Feed ID

### `POST /feeds/reorder`

Reorder a feed

Parameters:
  - `request` (body, required): Reorder details (feed_id, category, position)

Request body: see the Swagger schema for full field details.

### `POST /feeds/update`

Update a feed

Parameters:
  - `request` (body, required): Feed update details

Request body: see the Swagger schema for full field details.

## Rules

### `POST /rules/apply`

Apply rule to articles

Parameters:
  - `rule` (body, required): Rule definition (conditions and actions)

Request body: see the Swagger schema for full field details.

## Saved Filters

### `GET /saved-filters`

Create a new saved filter

Parameters:
  - `request` (body, required): Filter details

Request body: see the Swagger schema for full field details.

### `POST /saved-filters`

Create a new saved filter

Parameters:
  - `request` (body, required): Filter details

Request body: see the Swagger schema for full field details.

### `DELETE /saved-filters/filter`

Delete a saved filter

Parameters:
  - `id` (query, required): Filter ID
  - `request` (body, required): Updated filter details
  - `id` (query, required): Filter ID

Request body: see the Swagger schema for full field details.

### `PUT /saved-filters/filter`

Delete a saved filter

Parameters:
  - `id` (query, required): Filter ID
  - `request` (body, required): Updated filter details
  - `id` (query, required): Filter ID

Request body: see the Swagger schema for full field details.

### `POST /saved-filters/reorder`

Reorder saved filters

Parameters:
  - `request` (body, required): Filters with updated positions

Request body: see the Swagger schema for full field details.

## Statistics

### `DELETE /api/statistics`

Reset all statistics

### `GET /api/statistics`

Get statistics

Parameters:
  - `period` (query, optional): Time period
  - `offset` (query, optional): Period offset for navigation (e.g., -1 for previous, 1 for next)
  - `start_date` (query, optional): Start date (YYYY-MM-DD format, required for period=custom)
  - `end_date` (query, optional): End date (YYYY-MM-DD format, required for period=custom)

### `GET /api/statistics/all-time`

Get all-time statistics

### `GET /api/statistics/available-months`

Get available months

## Tags

### `GET /tags`

List or create tags

Parameters:
  - `request` (body, optional): Tag details (for POST)

Request body: see the Swagger schema for full field details.

### `POST /tags`

List or create tags

Parameters:
  - `request` (body, optional): Tag details (for POST)

Request body: see the Swagger schema for full field details.

### `POST /tags/delete`

Delete a tag

Parameters:
  - `request` (body, required): Tag ID to delete

Request body: see the Swagger schema for full field details.

### `POST /tags/reorder`

Reorder a tag

Parameters:
  - `request` (body, required): Tag ID and new position

Request body: see the Swagger schema for full field details.

### `POST /tags/update`

Update a tag

Parameters:
  - `request` (body, required): Tag update details (id, name, color, position)

Request body: see the Swagger schema for full field details.

## Translation

### `POST /translation/test-custom`

Test custom translation

Parameters:
  - `request` (body, required): Test request

Request body: see the Swagger schema for full field details.

## Browser

### `GET /browser/open`

Open URL in browser

Parameters:
  - `url` (query, optional): URL to open (for GET requests)
  - `request` (body, required): Open URL request (url) (for POST requests)

Request body: see the Swagger schema for full field details.

### `POST /browser/open`

Open URL in browser (server mode)

Parameters:
  - `request` (body, required): Open URL request (url)

Request body: see the Swagger schema for full field details.

## Custom Css

### `DELETE /custom-css`

Delete custom CSS

### `GET /custom-css`

Get custom CSS

### `POST /custom-css/dialog`

Upload CSS dialog (not available)

### `POST /custom-css/upload`

Upload custom CSS file

Parameters:
  - `file` (formData, required): CSS file to upload

## Email

### `POST /email/imap/test`

Test IMAP connection

Parameters:
  - `request` (body, required): IMAP connection test (email_imap_server, email_imap_port, email_username, email_password, email_folder)

Request body: see the Swagger schema for full field details.

## Freshrss

### `GET /freshrss/status`

Get FreshRSS sync status

### `POST /freshrss/sync`

Sync with FreshRSS

### `POST /freshrss/sync-feed`

Sync single FreshRSS feed

Parameters:
  - `stream_id` (query, required): FreshRSS stream ID

## Media

### `POST /media/cache/cleanup`

Cleanup media cache

Parameters:
  - `all` (query, optional): Clean all files (ignores age/size settings)

### `GET /media/cache/info`

Get media cache info

### `GET /media/proxy`

Proxy media file

Parameters:
  - `url` (query, required): Media URL to proxy
  - `referer` (query, optional): Referer URL for hotlink protection
  - `force_cache` (query, optional): Force caching even if globally disabled

### `GET /media/proxy-webpage`

Proxy webpage content

Parameters:
  - `url` (query, required): Webpage URL to proxy

## Network

### `POST /network/detect`

Detect network speed

### `GET /network/info`

Get network info

## Opml

### `GET /opml/export`

Export OPML file

### `POST /opml/export/dialog`

Export dialog (not available in server mode)

### `POST /opml/import`

Import OPML file

Parameters:
  - `file` (formData, required): OPML file

### `POST /opml/import/dialog`

Import dialog (not available in server mode)

## Progress

### `GET /progress`

Get fetch progress

### `GET /progress/task-details`

Get task details

## Rsshub

### `POST /api/rsshub/add`

Add RSSHub feed

Parameters:
  - `request` (body, required): RSSHub feed details

Request body: see the Swagger schema for full field details.

### `POST /api/rsshub/test-connection`

Test RSSHub connection

Parameters:
  - `request` (body, required): RSSHub connection details

Request body: see the Swagger schema for full field details.

### `POST /api/rsshub/transform-url`

Transform RSSHub URL

Parameters:
  - `request` (body, required): RSSHub URL to transform (rsshub:// protocol)

Request body: see the Swagger schema for full field details.

### `POST /api/rsshub/validate-route`

Validate RSSHub route

Parameters:
  - `request` (body, required): Route to validate

Request body: see the Swagger schema for full field details.

## Scripts

### `GET /scripts/dir`

Get scripts directory path

### `POST /scripts/dir/open`

Open scripts directory

### `GET /scripts/list`

List available scripts

## Summaries

### `DELETE /summaries/clear`

Clear all summaries

## Summarize

### `POST /summarize`

Summarize article

Parameters:
  - `request` (body, required): Summarize request (article_id, length, content)

Request body: see the Swagger schema for full field details.

## Translate

### `POST /translate/article`

Translate article title

Parameters:
  - `request` (body, required): Translation request (article_id, title, target_language)

Request body: see the Swagger schema for full field details.

### `POST /translate/text`

Translate text

Parameters:
  - `request` (body, required): Translation request (text, target_language)

Request body: see the Swagger schema for full field details.

## Translations

### `POST /translations/clear`

Clear all translations

## Update

### `GET /update/check`

Check for updates

### `POST /update/download`

Download update

Parameters:
  - `request` (body, required): Download request (download_url, asset_name)

Request body: see the Swagger schema for full field details.

### `POST /update/install`

Install update

Parameters:
  - `request` (body, required): Install request (file_path)

Request body: see the Swagger schema for full field details.

## Version

### `GET /version`

Get application version

## Webpage

### `GET /webpage/resource`

Proxy webpage resource

Parameters:
  - `url` (query, required): Resource URL to proxy
  - `referer` (query, required): Referer URL for the webpage

## Window

### `GET /window/state`

Get window state

### `POST /window/state`

Save window state

Parameters:
  - `state` (body, required): Window state to save

Request body: see the Swagger schema for full field details.
