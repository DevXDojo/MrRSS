# Papr Desktop Experience Migration Plan

## Current MrRSS Architecture

- **App shell:** `frontend/src/App.vue` already composes the sidebar, virtualized article list,
  reader, modal layer, keyboard shortcuts, and two desktop panel widths.
- **Sidebar:** `Sidebar.vue` coordinates the existing Activity Bar and Feed List. Feed categories,
  saved filters, context menus, refresh, discovery, and settings remain owned by the existing
  sidebar components.
- **Article list:** `ArticleList.vue` owns querying, pagination, selection and visibility tracking;
  `ArticleItem.vue` and `ArticleCardItem.vue` are view layers and must not bypass that pipeline.
- **Reader:** `ArticleDetail.vue` selects webpage/RSS mode. `ArticleContent.vue` preserves scroll,
  full-text, translation, summary, media, chat, and safe rendering; `ArticleBody.vue` applies prose
  styling without changing the content sanitization boundary.
- **Settings:** the existing reading tabs update one `SettingsData` model and use the debounced
  `useSettingsAutoSave` persistence path.
- **Stores:** Pinia remains authoritative for feeds/articles/filter state. Reader preferences remain
  in the shared settings composable rather than a second UI store.
- **Persistence:** settings are declared in `internal/config/settings_schema.json`; the generator
  produces Go handlers/defaults and frontend types/composable defaults.

## Papr to MrRSS Mapping

| Papr reference               | Native MrRSS implementation                                              |
| ---------------------------- | ------------------------------------------------------------------------ |
| `App.tsx` three-column shell | `App.vue` and `useResizablePanels.ts`                                    |
| Sidebar/feed rows            | `Sidebar.vue`, `FeedList.vue`, `SidebarCategory.vue`, `SidebarFeed.vue`  |
| `ArticleList.tsx`            | `ArticleList.vue`, `ArticleItem.vue`, `ArticleCardItem.vue`              |
| `Reader.tsx`                 | `ArticleDetail.vue`, `ArticleContent.vue`, `ArticleToolbar.vue`          |
| Reader typography            | `ArticleBody.vue`, `ArticleContent.css`, shared CSS reader tokens        |
| Reading settings section     | existing `modals/settings/reading/*` components                          |
| Zustand reader state         | generated settings model plus existing settings composables              |
| `ResizeHandle.tsx`           | existing `useResizablePanels.ts` resizer behavior                        |
| `styles.css` paper system    | concise semantic tokens in `frontend/src/style.css` and component styles |

Papr is a design/interaction reference only. React, Zustand, Tauri, and its application state are
not dependencies. The reference repository could not be cloned in the build environment (the
GitHub CONNECT tunnel returned HTTP 403), so the supplied screenshots and explicitly documented
Papr values are used alongside the existing MrRSS implementation.

## Delivery Phases

1. Introduce warm paper/surface/selection tokens for light and dark themes.
2. Refine the existing desktop shell and resize constraints while retaining independent scrollers.
3. Restyle existing sidebar and article rows without changing feed or virtual-list logic.
4. Center and constrain the existing reader, keeping media, sanitizer, translation, AI, and custom
   CSS behavior intact.
5. Add accessible live typography controls, reading-time metadata, and generated persistence.
6. Validate semantic global/per-feed view-mode mapping already present in MrRSS.
7. Add unit tests, build checks, and desktop screenshots.

## Risks and Mitigations

- **WebView/CSP:** no proxy, iframe sandbox, or sanitizer changes.
- **Article rendering:** keep `ArticleContent`/`ArticleBody` enhancement and media paths intact;
  presentation changes are CSS and derived metadata only.
- **Settings migration:** additive schema keys with defaults; no destructive database migration.
- **Virtual list:** do not replace item observation, pagination, caching, or list scroll ownership.
- **Dark mode:** define every new semantic token in both themes.
- **Custom CSS:** keep the current injection boundary and use reader variables as defaults.
- **AI toolbar density:** preserve current toolbar actions and use its existing menus.
- **Responsive layout:** retain current overlay behavior below the desktop breakpoint.
- **Reading time:** strip markup once in a computed value, use linear text counting, and test CJK,
  Latin, mixed, empty, and short input.
