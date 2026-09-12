export const toolbarActions = [
  { id: 'view', label: 'article.action.viewOriginal' },
  { id: 'translation', label: 'setting.reading.showTranslations' },
  { id: 'read', label: 'article.action.markAsRead' },
  { id: 'favorite', label: 'article.toolbar.addToFavorite' },
  { id: 'readLater', label: 'article.toolbar.addToReadLater' },
  { id: 'browser', label: 'article.action.openInBrowser' },
  { id: 'copyTitle', label: 'common.contextMenu.copyTitle' },
  { id: 'copyLink', label: 'common.contextMenu.copyLink' },
  { id: 'reload', label: 'article.action.reloadContent' },
  { id: 'obsidian', label: 'setting.plugins.obsidian.exportTo' },
  { id: 'notion', label: 'setting.plugins.notion.exportTo' },
  { id: 'zotero', label: 'setting.plugins.zotero.exportTo' },
] as const;

export type ToolbarActionID = (typeof toolbarActions)[number]['id'];
export interface ToolbarItem {
  id: ToolbarActionID;
  visible: boolean;
}

// Ignore unknown/duplicate entries and append newly introduced actions. An
// invalid saved value must never make the toolbar impossible to recover.
export function parseToolbarLayout(raw: string): ToolbarItem[] {
  const result: ToolbarItem[] = [];
  const known = new Set<string>(toolbarActions.map((action) => action.id));
  const seen = new Set<string>();
  try {
    const parsed: unknown = JSON.parse(raw);
    if (Array.isArray(parsed)) {
      for (const item of parsed) {
        if (
          typeof item !== 'object' ||
          item === null ||
          !('id' in item) ||
          typeof item.id !== 'string' ||
          !known.has(item.id) ||
          seen.has(item.id)
        )
          continue;
        seen.add(item.id);
        result.push({
          id: item.id as ToolbarActionID,
          visible: !('visible' in item) || item.visible !== false,
        });
      }
    }
  } catch {
    /* Missing or malformed settings use the default layout. */
  }
  for (const action of toolbarActions) {
    if (!seen.has(action.id)) result.push({ id: action.id, visible: true });
  }
  return result;
}
