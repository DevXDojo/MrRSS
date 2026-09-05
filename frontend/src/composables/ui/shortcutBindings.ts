import { ref } from 'vue';

export interface KeyboardShortcuts {
  nextArticle: string;
  previousArticle: string;
  nextArticleArrow: string;
  previousArticleArrow: string;
  openArticle: string;
  closeArticle: string;
  toggleReadStatus: string;
  toggleFavoriteStatus: string;
  toggleReadLaterStatus: string;
  openInBrowser: string;
  toggleContentView: string;
  refreshFeeds: string;
  markAllRead: string;
  openSettings: string;
  addFeed: string;
  focusSearch: string;
  toggleFilter: string;
  toggleUnreadFilter: string;
  toggleFavoritesFilter: string;
  toggleReadLaterFilter: string;
  goToAllArticles: string;
  goToUnread: string;
  goToFavorites: string;
  goToReadLater: string;
}

export const shortcutsEnabled = ref(true);
export const shortcuts = ref<KeyboardShortcuts>({
  nextArticle: 'j',
  previousArticle: 'k',
  nextArticleArrow: 'ArrowRight',
  previousArticleArrow: 'ArrowLeft',
  openArticle: 'Enter',
  closeArticle: 'Escape',
  toggleReadStatus: 'r',
  toggleFavoriteStatus: 's',
  toggleReadLaterStatus: 'l',
  openInBrowser: 'o',
  toggleContentView: 'v',
  refreshFeeds: 'Shift+r',
  markAllRead: 'Shift+a',
  openSettings: ',',
  addFeed: 'a',
  focusSearch: '/',
  toggleFilter: 'f',
  toggleUnreadFilter: 'Alt+r',
  toggleFavoritesFilter: 'Alt+s',
  toggleReadLaterFilter: 'Alt+l',
  goToAllArticles: '1',
  goToUnread: '2',
  goToFavorites: '3',
  goToReadLater: '4',
});

export function withShortcut(label: string, actions: keyof KeyboardShortcuts | (keyof KeyboardShortcuts)[]): string {
  if (!shortcutsEnabled.value) return label;
  const keys = [...new Set((Array.isArray(actions) ? actions : [actions])
    .map((action) => shortcuts.value[action]).filter(Boolean))];
  return keys.length ? `${label} (${keys.join(' / ')})` : label;
}
