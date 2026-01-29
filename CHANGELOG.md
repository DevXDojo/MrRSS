# Changelog

All notable changes to MrRSS will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.18] - 2026-01-29

### Refactored

- Refactored all popup windows and context menus for improved consistency and maintainability. (#582)

### Added

- Added AI-powered article search functionality. (#248)
- Added support for saving custom filters. (#223)
- Implemented feed tagging for better organization. (#545)
- Added batch operations to the feed list for efficient management. (#593)
- Added card layout view option for the article list. (#592)

### Changed

- Added confirmation dialog when bulk-marking articles as read to prevent accidental actions. (#560)
- Thumbnail previews now display in compact mode when enabled. (#589)
- Update checks detect firewall-related connectivity issues for users in mainland China. (#621)

### Fixed

- Fixed missing default title assignment when articles lack a title. (#566)
- Fixed multiple minor styling inconsistencies. (#569, #584, #579)
- Fixed styling issues in the image gallery view. (#571, #581)
- Fixed layout shift in list width after navigating to the settings page. (#575)
- Fixed HTTP headers being blocked by Cloudflare for some requests. (#620)
- Fixed broken images in article content caused by incorrect referrer headers. (#597)
- Fixed issue preventing the article summary from being closed. (#591)
- Fixed synchronization errors with FreshRSS. (#598, #600)
- Fixed HTML character encoding issues in the image gallery view. (#596)
- Improved filter-by-category performance by adding a missing database index. (#570)
- Implemented IMAP ID command support for enhanced client identification. (#602)

## [1.3.17] - 2026-01-24

### Refactored

- Refactored the settings page and i18n system for improved maintainability and extensibility.
- Upgraded the Wails version and corresponding Go dependencies.

### Added

- Added support for thumbnail images in the gallery view for easier navigation. (#495)
- Added the ability to filter images by category in the gallery view. (#487, #490)
- Added support for translation between Traditional and Simplified Chinese. (#511)
- Added the ability to copy images to the clipboard. (#515)
- Enhanced the styling and user experience of the gallery mode. (#520)
- Added support for customizing typography styles. (#488)
- Added the ability to mark items above or below as read. (#390, #524)
- Changed the default behavior to open links in an external browser. (#551)
- Added the ability to jump to a specific feed by clicking on it in the settings page. (#548)
- Added support for displaying multiple authors in a single feed. (#554)

### Changed

- Changed the checkbox checked indicator from an asterisk to a checkmark. (#507)
- Improved the feed list in the settings page for better usability. (#498)
- Made the protocol optional when adding or editing feeds. (#502)
- Improved the performance of article content search. (#509)
- Optimized the styling of compact mode. (#488, #504)

### Fixed

- Fixed multiple minor styling issues. (#492, #493, #494, #496, #503, #505, #506, #510, #516, #517, #519, #521, #522, #523, #550)
- Fixed an issue where plain text could not be translated correctly. (#511, #514)
- Fixed an issue where the reading status did not update correctly in gallery mode.
- Fixed an issue where translation occurred even when the feature was disabled. (#541)
- Fixed an issue where custom headers could accept non-ASCII characters. (#549)

## [1.3.16] - 2026-01-15

### Added

- Added compact mode for article list to reduce visual clutter. (#403)
- Enhanced image gallery with multi‑image support and improved navigation. (#457)
- Added support for Anthropic and DeepSeek AI services.
- Added option to hide text overlay in image gallery view. (#486)
- Added indication for feeds using image gallery mode in feed list. (#485)
- Added ability to customize translation service endpoint. (#383)
- Added option to disable automatic feed refresh. (#448)
- Added option to display translated text only (hide original). (#464)

### Changed

- Documents now open in the default browser with added multi‑language support. (#458)
- Import/export no longer shows error messages when no file is selected. (#483)
- Articles with >60% target language content are no longer translated to reduce API usage.

### Fixed

- Fixed conflict between left/right arrow shortcuts and input fields. (#454)
- Fixed article list not scrolling automatically when switching articles. (#451)
- Fixed minor styling issues. (#452, #456, #453, #484)
- Fixed display of future publish times for some articles.
- Fixed summary generation not respecting language settings. (#480)
- Fixed Gemini API integration. (#459)
- Fixed automatic application updates occurring without user confirmation. (#479)
- Fixed intermittent FreshRSS synchronization failures. (#460)
- Fixed view mode reset when switching between articles and images. (#432)
- Fixed XPath feed parsing in certain cases. (#479)

### Refactored

- Refactored sidebar, settings, and summary components for improved maintainability and performance. (#461, #466)

## [1.3.15] - 2026-01-11

### Changed

- Reduced the size of binary files by optimizing lingua-go import. (#450)

### Fixed

- Fixed an issue where the old database can not be migrated correctly in some cases.

## [1.3.14] - 2026-01-10

### Added

- Supported better reverse proxy for website display. (#414)
- Supported RSSHub feed type for better integration with RSSHub instances. (#176, #302) (@cry0404)
- Supported a statistics tab in the settings modal to view usage statistics over time.
- Supported manually sorting rules for advanced users. (#398)
- Supported Gemini service API. (#437)
- Supported language detection to reduce unnecessary translation requests. (#410)
- Added error messages for feeds that fail to refresh. (#429)
- Supported buttons to switch to previous/next articles in the article detail view. (#357)
- Supported -10s and +10s skip buttons in the audio player. (#395)

### Changed

- Cached thumbnail images in the article list to avoid disappearing after restarting the application. (#423)
- Improved the performance of article list rendering.
- Prevented the article content viewer from closing when clicking the same article again. (#434)

### Fixed

- Fixed an issue where URLs were not trimmed correctly when adding or editing feeds. (#413)
- Fixed an issue where the summary could not be regenerated after the article content changed. (#412)
- Fixed some minor style issues. (#396, #397, #402, #407, #425, #428, #430, #443, #449)
- Fixed an issue where the image gallery view showed only 2 columns. (#399)
- Fixed an issue where left and right click actions did not work correctly in the feeds list. (#394)
- Fixed an issue where feeds could not be dragged into collapsed categories. (#394)
- Fixed an issue where links could not be opened in the default browser after extracting the full article content. (#409)
- Fixed an issue where duplicate feeds could be added. (#401)
- Fixed an issue where the article list got stuck in some cases. (#422)
- Fixed an issue where the sidebar width would shrink when feed titles were short. (#433)
- Fixed an issue where the image viewer could not be closed automatically after switching articles or feeds. (#431)
- Fixed an issue where AI summaries were always regenerated in English or were not accurate enough. (#424, #438)
- Fixed an issue where FreshRSS synchronization failed for feeds in some cases. (#440)
- Fixed an issue where translation failures caused many toast notifications. (#436)
- Fixed an issue where FreshRSS articles could not display thumbnail images correctly. (#446)

**Special Thanks** to @EnterMan123 for carefully testing and reporting many of these issues!

## [1.3.13] - 2026-01-03

**BREAKING**: The FreshRSS synchronization feature has been significantly enhanced, offering more options and improved reliability. You may need to remove and re-add your FreshRSS feeds after upgrading.

### Added

- Enhanced FreshRSS synchronization with additional configuration options. (#333, #376)
- Added support for building and publishing multi-architecture Docker images to GHCR. (#349) (@czyt)
- Added support for email newsletter feeds via IMAP. (#313)
- Added more filter conditions, including title regex, FreshRSS feed status, image gallery status, and feed mode. (#372)
- Added display and sorting by last updated time, refresh status, and update frequency in the feed list. (#374)
- Added playback speed and volume controls for audio in article content rendering mode. (#354)

### Changed

- Increased default concurrency and timeout settings for feed fetching based on network speed detection. (#375)
- Disabled auto-close when clicking the background of popup windows to prevent accidental closures. (#355)
- Optimized the display of article publish times in the article list. (#373)
- Applied consistent scrollbar styling across all scrollable areas. (#389)

### Fixed

- Fixed "python command not found" error. (#364)
- Fixed broken links in article content rendering mode that failed to open in the default browser. (#330)
- Fixed incorrect rendering of some images in article content rendering mode. (#327)
- Fixed incorrect translation application in nested structures.
- Fixed search result highlight styling in article content. (#361)
- Fixed incorrect application name display (`{{.info.ProductName}}`) on Windows. (#351)
- Fixed repeated macOS privacy permission dialogs when opening articles. (#337)
- Fixed white screen flash when opening or closing windows. (#384)
- Fixed visual glitches caused by scrollbar thumb and article item borders. (#387, #388)
- Fixed inconsistent height in date input fields. (#391)
- Fixed image gallery not loading more images correctly on scroll. (#385)

### Removed

- Removed HTTPS requirement for feed URLs and API endpoints. (#251)
- Removed automatic article translation during feed refresh.
- Disabled hiding of image feeds in the "All Articles" view. (#386)

## [1.3.12] - 2025-12-29

**BREAKING**: The core system (including the feed fetcher, scheduler, and database cleaner) has been re-architected to improve performance and maintainability. (#350, #366)

The following changes may affect existing configurations:

- *Feeds that are not set to "Use global refresh settings" may no longer be refreshed when fetching all feeds.*
- *All article contents are now cached; enabling "Auto Cleanup" is recommended to prevent excessive database growth.*
- *The maximum refresh interval for intelligent scheduling has been increased from 3 hours to 24 hours.*
- *Feed refresh operations now time out after 5 seconds, followed by one retry with a 10-second timeout.*

### Added

- Added visual indicators for feeds that are currently refreshing or queued to refresh in the feed list.
- Added support for creating new chat sessions and viewing chat history in the AI Chat feature. (#340)
- Added support for rendering chat messages in Markdown format in the AI Chat feature. (#338, #346)
- Added the ability to search within article content. (#361)

### Changed

- Added a user setting to enable or disable automatic installation of updates after download. (#336)
- Keyboard shortcuts can now be enabled or disabled via settings.
- All article contents are now cached to improve loading speed when switching between articles. (#344)
- Improved error messages when adding or editing feeds in XPath mode for better user experience. (#345, #364)

### Fixed

- Fixed repeated macOS privacy permission dialogs when opening articles. (#337)
- Fixed high GPU usage when opening the settings page. (#339)
- Fixed feed refresh failures caused by certain invalid feeds. (#341)
- Fixed incorrect rendering of some images in article content. (#327)
- Fixed server startup failure due to the newly added custom CSS file upload feature. (#343)
- Fixed translation issues in confirmation pop-up windows.
- Fixed keyboard shortcut conflicts when the settings page is open. (#355)
- Fixed the "Mark all as read" button in the article list not working correctly. (#318, #353, #363)
- Fixed articles disappearing from the article list when opened while a filter is applied. (#318, #353, #362)
- Fixed incorrect application name display (`{{.info.ProductName}}`) on Windows. (#351)
- Fixed removal of advanced settings when moving a feed. (#356)
- Fixed incorrect summary display when switching articles before summary generation completes. (#365)
- Fixed error messages caused by NULL DATETIME values. (#347)
- Fixed inability to cancel text selection by clicking on blank areas in articles. (#360)

## [1.3.11] - 2025-12-26

### Fixed

- Fixed the issue of some incorrect styles in settings page.
- Fixed the issue where sidebar disappears.
- Fixed the issue drag-and-drop will not work correctly in some cases.

## [1.3.10] - 2025-12-26

### Added

- Supported import and export feeds in JSON format. (#317)
- Supported choosing auto expand content for each feed. (#306)
- Supported uploading CSS files for customized styling of articles. (#324)
- Supported showing only unread articles in article list. (#318)

### Changed

- Improved I18n translations, icons, and descriptions in settings page for better clarity and user experience.
- Improved UX of feed adding/editing modal. (#317)
- Expand status of categories in sidebar is now persisted across application restarts. (#315)

### Fixed

- Fixed the issue where length limit for AI-generated summaries was not applied correctly. (#323)
- Fixed the issue where the last time of network detect displays 739609 days ago if never detected before. (#314)
- Fixed the issue where multi-layer categories in sidebar do not display correctly. (#322)
- Fixed the issue of incorrect folder path in server mode. (#321)

## [1.3.9] - 2025-12-25

### Added

- Supported customized request headers for AI services. (#301)
- Supported enable automatically extracting full article content from original website. (#306)
- Supported choosing article view mode for each feed. (#309)

### Changed

- Reorganized settings page layout for better user experience.

### Fixed

- Fixed the issue where some articles failed to open when filter is applied. (#304)
- Fixed the issue where folders are not synchronized correctly and articles are duplicated when syncing with FreshRSS. (#305)

## [1.3.8] - 2025-12-24

### Added

- Supported AI setting tests in settings page to verify connectivity and credentials. (#297)
- Supported fetching feeds which require JavaScript rendering using headless browser. (#298)

### Changed

- AI generated summaries are now stored in the database to avoid redundant requests and improve performance. (#295)
- Reduce the frequency of automatic record of window status to improve performance.
- Improved the conversion from HTML to Markdown when exporting articles to Obsidian. (#299)

### Fixed

- Fixed the issue where docker image failed to access local files due to permission issues. (#296)
- Fixed the issue where articles failed to open in default browser. (#294)
- Fixed the issue where importing and exporting OPML files did not work correctly. (#271)
- Fixed the issue of CSP blocking some external resources.

## [1.3.7] - 2025-12-23

### Added

- Supported server mode for self-hosted web application deployment. (#267) (@caoli5288)
- Supported drag-and-drop to reorder feeds or change feed categories. (#288)
- Supported AI Chat on article content. And of course **it's disabled by default**! (#286)
- Supported exporting articles to Obsidian. (#289)
- Supported extracting full article content from original website when RSS feed only provides summary. (#266)

### Changed

- AI summarization is now triggered manually on default to avoid excessive API usage. Users can enable automatic summarization in settings if desired. (#287)
- Added Plugin setting tab in settings page and moved FreshRSS synchronization settings there.
- Improved icons and translations for better user experience.
- Enhanced the conversion from HTML to Markdown when exporting articles to Obsidian. (#299)

### Fixed

- Fixed the issue where concurrent feed refreshes exceed network capacity limit. (#262)

## [1.3.6] - 2025-12-22

### Added

- Supported importing feeds with HTML+XPath / XML+XPath type from OPML files. (#264)
- Supported FreshRSS synchronization. (#245)

### Changed

- Improved error display for customized scripts when adding/editing feeds. (#264)
- Network connection test now supports proxy settings. (#256)

### Fixed

- Fixed the issue where different articles display the same content due to incorrect URL matching. (#257)
- Fixed the issue where import and export of OPML files did not work correctly on macOS. (#263)
- Fixed the issue where localhost cannot be processed correctly. (#257)

### Removed

- Removed single instance lock on Linux platform to avoid D-Bus related issues. (#246)

## [1.3.5] - 2025-12-20

**BREAKING**: AI-based summarization and translation now need a full path instead of just endpoint URL.

> e.g. for OpenAI services, use `https://api.openai.com/v1/chat/completions`. for Ollama, use `http://localhost:11434/api/generate`.

### Added

- Supported ollama and other local LLMs for AI-based translation and summarization. (#251)
- Supported limits and quotas for AI services to control usage and costs. (#252)
- Supported hover to mark articles as read in article list. (#250)
- Supported deelpx translation service. (#247)

### Changed

- Improved AI settings UI/UX for better user experience.
- Refactored docs and workflows to improve maintainability and clarity.
- AI translation and summarization are now cached to reduce redundant requests and improve performance.
- Recent articles are now cached to improve loading speed.
- When AI functionality gets errors, fallback to local summarization/translation automatically.

### Fixed

- Fixed the issue where some opml files cannot be imported and outported correctly. (#249)
- Fixed the issue where proxy settings were not applied correctly for feed fetching. (#256)
- Fixed the issue where software print too much debug logs in production builds.
- Fixed the issue where network connection test fails when some test endpoints are unreachable. (#256)
- Fixed the issue where summarization failures will affect article content rendering. (#242)
- Fixed the issue where article content fetching blocked by feed refreshes.
- Fixed the issue of dark mode styles on Linux platform.

## [1.3.4] - 2025-12-18

### Fixed

- Fixed the issue where window title bar buttons on MacOS overlapping with content area.
- Fixed the issue where window cannot be dragged on MacOS. (#242)

## [1.3.3] - 2025-12-18

### Added

- Supported copying article link and title to clipboard from article actions menu. (#155)

### Changed

- Replace the following functionality with a native implementation using wails3 (#242)
  - Open link in default browser
  - Window events handling (minimize, maximize, close) and management
  - Native window context menu and title bar on MacOS

### Fixed

- Fixed the issue where super loooooooong article titles causing layout breaking in article list.
- Fixed the issue where cutting long article titles in chinese does not work correctly.

## [1.3.2] - 2025-12-17

### Fixed

- Fixed the issue where MacOS window cannot be closed correctly after maximizing. (#221)
- Fixed the issue where images in article content rendering mode cannot be displayed correctly. (#222)
- Fixed the issue where windows app cannot be packaged correctly due to wrong version number format.

## [1.3.0] - 2025-12-17

### Changed

- **BREAKING**: Upgraded from Wails v2 to Wails v3 (alpha) framework (#234)
  - Migrated to new API
  - Replaced external systray library with Wails v3 built-in system tray
  - Updated single instance handling to use v3 API
  - Updated event handling to use v3 hooks
  - Updated build system to use Taskfile and Wails v3 CLI
  - Updated dependencies to work with WebKit2GTK 4.1 and libsoup 3.0
- Changed GitHub Actions workflows compatibility with Wails v3

## [1.2.20] - 2025-12-16

### Changed

- Added more tests for backend and frontend components to improve code reliability.
- Updated dependencies to latest versions for security and performance improvements.

### Fixed

- Fixed issues related to MacOS platform (#212)
  - Updated icons for better appearance.
  - Added more white space on top of the main window for better visual balance.
  - Disabled icon name on tray.
  - Fixed the issue where window cannot be dragged.
  - Fixed the issue where application not closing correctly after maximizing.

## [1.2.19] - 2025-12-15

### Fixed

- Fixed the issue where some settings were not saved and applied correctly. (#201)
- Fixed the issue where macOS application failing to launch after installation.

## [1.2.18] - 2025-12-14

### Added

- Supported image gallery for browsing all images in articles. (#190)
- Supported network latency and bandwidth testing in settings. (#194)

### Changed

- Added number of concurrent feed refreshes according to network situation.

### Fixed

- Fixed the issue where software can open multiple instances. (#198)
- Fixed the issue where number of feeds left to refresh is not accurately displayed during feed refresh. (#194)

## [1.2.17] - 2025-12-13

### Added

- Supported upgrade in portable mode. (#191)

### Fixed

- Fixed the issue where settings cannot be saved and applied by downgrade TailwindCSS version.

## [1.2.16] - 2025-12-13

### Added

- Add toggle button to hide/show article content translations. (#186)

### Changed

- Updated all dependencies to latest versions for security and performance improvements.

### Fixed

- Fixed the issue where MacOS cannot complile correctly for system tray support. (#181)
- Fixed the issue where Linux-ARM64 AppImage cannot run correctly.

## [1.2.15] - 2025-12-13

### Changed

- Supported alpha, beta, and pre-release version tags. (#182)
- Enhanced credential encryption mechanism to improve security during database migration and storage. (#160)

## [1.2.14] - 2025-12-12

### Added

- Supported portable mode for running MrRSS from USB drives with all data stored in a single folder. (#167)
- Supported minimizing to system tray on close action.
- Supported hiding preview images in article list for a more compact view. (#157)

### Fixed

- Fixed the issue where some images wrapped in links cannot be operated correctly.
- Fixed the issue where single-line link cannot be translated correctly.
- Fixed the issue where some links cannot be opened in the default browser.
- Fixed the issue where icons on MacOS were not displayed correctly. (#173)
- Fixed the issue where the window size and position were not restored correctly. (#173)

## [1.2.13] - 2025-12-11

### Added

- Supported media cache system to bypass anti-hotlinking restrictions and cache images/videos locally. (#152)
- Supported proxy settings for network requests.
- Supported intelligent refresh scheduling based on feed update frequency. (#151)
- Supported customizing proxy and refresh settings per feed. (#151)
- Supported read all articles for a specific feed or category. (#156)

### Changed

- Google Translate endpoint is now customizable in settings. (#158)

### Fixed

- Fixed the issue where title and summary cannot be selected and copied in article content rendering mode. (#155)
- Fixed the issue where some articles are rendered with incorrect formatting in article content rendering mode.

## [1.2.12] - 2025-12-10

### Changed

- Settings now support validation and show error messages for invalid inputs. (#147)

### Fixed

- Links in article content rendering mode can now be translated correctly. (#148)
- Fixed the issue where some images were not displayed in article content rendering mode. (#148)

## [1.2.11] - 2025-12-08

### Added

- Supported selecting existing categories when adding or editing a new feed.
- When playing audio or video in article content rendering mode, playback controls are now available.
- Supported customizing the AI prompt for article summarization and translation.

### Changed

- Improved styles for article content rendering mode.

### Fixed

- Fixed the issue where some feeds cannot be handled due to invalid styles in RSS XML.
- Fixed the issue where inline elements (e.g. code, formulas) were not handled correctly in translation.
- Fixed the issue where toast notifications not supporting dark mode caused visibility problems.
- Fixed the issue related to importing OPML files.

## [1.2.10] - 2025-12-07

### Added

- Supported audio and video embedding in article content rendering mode.

### Changed

- Enhanced styling of article content for better readability.

## [1.2.9] - 2025-12-05

### Added

- Supported Baidu Translation and AI-based translation.
- Supported AI-based article summarization using OpenAI-compatible APIs.

### Changed

- Errors from translation services are now logged and displayed to users for better troubleshooting.

### Fixed

- Fixed the issue where the default settings were not applied correctly on first launch.
- Fixed the issue where `PubMed` feed parsing failed.

## [1.2.8] - 2025-12-04

### Added

- Implemented Read Later functionality, articles marked for read later are automatically set to unread.

### Changed

- Last update time now displayed as inline sub-text instead of separate row.
- Added toggle filter shortcut (default: `f`).
- Nav icons use fill style when active for stronger visual feedback.
- Category headers are now sticky for scroll context retention.
- Feed refresh now skipped on startup if last article update interval is within set threshold.
- After each article refresh completes, the app now checks for updates. If a new version is detected, it automatically downloads and installs in the background.
- Changed some default settings.

### Fixed

- Fixed styling issues, including incorrect icon colors in dark mode, inconsistent font sizes, and misaligned elements.

## [1.2.7] - 2025-12-03

### Added

- Supported hiding feeds from timeline.

### Fixed

- Fixed initialization problem by adding progress tracking for single feed and OPML import.

## [1.2.6] - 2025-12-02

### Added

- Added TF-IDF and TextRank algorithms for generating article summaries.
- Added auto-translation support for summary, title, and content in rendering.
- Enhanced multimedia support in content rendering mode.

### Changed

- Improved image viewer with drag support and better zooming.
- Refactored both frontend and backend code for better maintainability.

### Fixed

- Fixed the issue where searching box scrolls with the feed list.

## [1.2.5] - 2025-11-27

### Added

- Supported for user-defined scripts to fetch and parse non-standard RSS feeds.
- Improved shortcuts for popup window actions.
- Supported sorting articles list in settings by various criteria.
- Supported for refreshing individual feeds via right-click context menu.
- Supported for searching feeds in the feed list.

### Changed

- Article list will not refresh during feed refresh, fixing a bug causing the article list to occasionally crash.
- Generate article titles from content when RSS feed items are missing titles.

### Fixed

- Fixed issue where some UI elements did not scale properly.
- Fixed bug causing view mode performe incorrectly when switching articles rapidly.

### Removed

- Removed search box for article list because the filter function covers the same use case.

## [1.2.4] - 2025-11-27

### Changed

- Refactored frontend codebase and landing page for better maintainability and user experience.
- Added tests for critical components to improve code reliability.
- Updated dependencies to latest versions for security and performance improvements.
- Better documentation for developers and contributors.
- Improved CI/CD pipeline for faster and more reliable builds and deployments.

## [1.2.2] - 2025-11-26

### Added

- Added keyboard shortcuts for common actions and corresponding settings in the Settings tab.
- Supported customizing rules with "If [condition], then [action]" format for advanced users.

### Changed

- Improved landing page UI/UX for better user engagement.
- Improved documentation for new users.

## [1.2.1] - 2025-11-26

### Added

- Adds advanced article filtering via a modal accessible from a filter button next to the search box.

## [1.2.0] - 2025-11-25

### Added

- Implements automated feed discovery from friend links with intelligent batch scanning, comprehensive deduplication, real-time progress tracking

### Changed

- Major restructuring of backend code for improved maintainability

## [1.1.8] - 2025-11-24

### Added

- Feed icons now display in the Settings > Feeds tab feed list for better visual identification
- Website homepage link stored for each feed (from RSS feed metadata)

### Changed

- "Open Website" context menu action now opens the feed's website homepage (if available) instead of RSS feed URL, with fallback to RSS URL
- All hardcoded strings now properly use i18n translations for better internationalization support

### Fixed

- Replaced native `prompt()` with custom `showInput()` dialog for consistent UI

## [1.1.7] - 2025-11-24

### Added

- Unread count badge displayed on each feed in the sidebar and "All Articles" button
- "Mark All as Read" button next to the refresh button in article list and feed context menu
- When feeds fail to load, display error message in feed list instead of silently failing
- Implemented input dialog for moving feeds to a new category

### Changed

- Fixed unfavorite icon for better visibility

## [1.1.6] - 2025-11-23

### Added

- "Open Website" option in feed right-click menu
- Startup on boot setting in General settings tab (default off)

### Changed

- "Hide Article" context menu item now shows as a danger button
- Improved settings tab switching style with hover effects and animations
- Fixed unfavorite icon visibility in article detail view

### Fixed

- Fixed software update installation process - updates can now be properly installed

## [1.1.5] - 2025-11-23

### Added

- Switch between viewing the original webpage and RSS content within the app
- Article hiding functionality
- Last article update time display in settings

### Changed

- Improved UI text and image selection prevention

### Fixed

- Fixed issue where translation settings changes didn't clear existing translations

## [1.1.4] - 2025-11-23

### Added

- Auto cleanup sub-settings:
  - Max cache size setting (default 20MB) - controls maximum database size before cleanup
  - Max article age setting (default 30 days) - automatically delete articles older than specified days (except favorites)
- Download progress bar during update download
- App automatically closes after starting installer to prevent conflicts during update
- Automatic cleanup of installation packages after update installation

### Changed

- Settings now auto-save immediately when changed (no need to click save button)

### Removed

- "Save Settings" button at bottom of settings page (replaced with auto-save)

## [1.1.3] - 2025-11-22

### Added

- Automatically detects user's operating system and CPU architecture and downloads appropriate installer from GitHub releases. Then launches installer and prepares for update
- Multi-Platform Support:
  - Windows: x64 (amd64), ARM64
  - Linux: x64 (amd64), ARM64 (aarch64)
  - macOS: Universal (Intel & Apple Silicon)
- Visual feedback during update download and installation

## [1.1.2] - 2025-11-22

### Added

- Initial release preparation
- OPML import/export functionality
- Feed category organization
- Automatically detect and apply system theme preference
- Better defaults for translation settings
- Version check functionality in Settings → About tab

### Changed

- Simplified update check UI
- Improved theme switching mechanism
- Better handling of translation provider selection

### Fixed

- Various bug fixes and stability improvements
- UI refinements for better user experience
- Theme switching issues between light and dark modes
- Translation default language selection
- Update notification display

## [1.1.0] - 2025-11-20

### Added

- **Initial Public Release** of MrRSS
- **Cross-Platform Support**: Native desktop app for Windows, macOS, and Linux
- **RSS Feed Management**: Add, edit, and delete RSS feeds
- **Article Reading**: Clean, distraction-free reading interface
- **Smart Organization**: Organize feeds into categories
- **Favorites & Reading Tracking**: Save articles and track read/unread status
- **Modern UI**: Clean, responsive interface with dark mode support
- **Auto-Translation**: Translate article titles using translation services or AI-based translation
- **OPML Support**: Import and export feed subscriptions
- **Auto-Update**: Configurable interval for fetching new articles
- **Database Cleanup**: Automatic removal of old articles
- **Multi-Language Support**: English and Chinese interface
- **Theme Support**: Light, dark, and auto (system) themes

---

## Release Notes

### Version Numbering

MrRSS follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

### Download

Downloads for all platforms are available on the [GitHub Releases](https://github.com/WCY-dt/MrRSS/releases) page.

### Upgrade Notes

When upgrading from a previous version:

1. Your data (feeds, articles, settings) is preserved in platform-specific directories
2. Database migrations are applied automatically on first launch
3. For major version upgrades, please review the changelog for breaking changes

### Support

- Report bugs: [GitHub Issues](https://github.com/WCY-dt/MrRSS/issues)
- Feature requests: [GitHub Issues](https://github.com/WCY-dt/MrRSS/issues)
- Documentation: [README](README.md)
