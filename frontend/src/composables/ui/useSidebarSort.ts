import { computed } from 'vue';
import { useSettings } from '@/composables/core/useSettings';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import type { Feed } from '@/types/models';

export const sidebarSortModes = ['manual', 'name_asc', 'name_desc', 'count_asc', 'count_desc', 'latest'] as const;
export type SidebarSortMode = typeof sidebarSortModes[number];
interface SortRow { name: string; count: number; latest: number; position: number; pinned: boolean }
export function compareSidebarRows(a: SortRow, b: SortRow, mode: SidebarSortMode) {
  if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
  let result = 0;
  switch (mode) {
    case 'manual': result = a.position - b.position; break;
    case 'name_asc': result = a.name.localeCompare(b.name); break;
    case 'name_desc': result = b.name.localeCompare(a.name); break;
    case 'count_asc': result = a.count - b.count; break;
    case 'count_desc': result = b.count - a.count; break;
    case 'latest': result = b.latest - a.latest; break;
  }
  return result || a.name.localeCompare(b.name);
}
function readArray(value: string): string[] {
  try { const data = JSON.parse(value); return Array.isArray(data) ? data.filter((item) => typeof item === 'string') : []; }
  catch { return []; }
}
let saveQueue = Promise.resolve();

export function useSidebarSort() {
  const { settings } = useSettings();
  const store = useAppStore();
  const { t } = useI18n();
  const mode = computed(() => sidebarSortModes.includes(settings.value.sidebar_sort_mode as SidebarSortMode) ? settings.value.sidebar_sort_mode as SidebarSortMode : 'manual');
  const pinned = computed(() => new Set(readArray(settings.value.sidebar_pinned_items)));
  const isPinned = (key: string) => pinned.value.has(key);
  const counts = computed(() => {
    const key = { favorites: 'favorites', readLater: 'read_later_unread', unread: 'unread', imageGallery: 'images_unread' }[store.currentFilter];
    return key ? store.filterCounts[key] || {} : store.unreadCounts.feedCounts;
  });
  const categoryStats = computed(() => {
    const stats = new Map<string, { count: number; latest: number }>();
    for (const feed of store.feeds) {
      const parts = (feed.category || '').split('/');
      for (let i = 1; i <= parts.length; i++) {
        const path = parts.slice(0, i).join('/');
        const value = stats.get(path) || { count: 0, latest: 0 };
        value.count += counts.value[feed.id] || 0;
        value.latest = Math.max(value.latest, Date.parse(feed.latest_article_time || '') || 0);
        stats.set(path, value);
      }
    }
    return stats;
  });
  function compareCategories(a: string, b: string, order: string[]) {
    const row = (path: string): SortRow => ({
      name: path.split('/').at(-1) || path, ...(categoryStats.value.get(path) || { count: 0, latest: 0 }),
      pinned: isPinned(`category:${path}`), position: order.includes(path) ? order.indexOf(path) : Number.MAX_SAFE_INTEGER,
    });
    return compareSidebarRows(row(a), row(b), mode.value);
  }
  function compareFeeds(a: Feed, b: Feed) {
    const row = (feed: Feed): SortRow => ({ name: feed.title, count: counts.value[feed.id] || 0,
      latest: Date.parse(feed.latest_article_time || '') || 0, position: feed.position || 0, pinned: isPinned(`feed:${feed.id}`) });
    return compareSidebarRows(row(a), row(b), mode.value) || a.id - b.id;
  }
  function save(makeUpdate: () => { sidebar_sort_mode?: string; sidebar_pinned_items?: string }) {
    const task = saveQueue.then(async () => {
      const values = makeUpdate();
      const response = await fetch('/api/settings', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(values) });
      if (!response.ok) throw new Error('Sidebar preference save failed');
      Object.assign(settings.value, values);
    });
    saveQueue = task.catch(() => { window.showToast(t('common.errors.savingSettings'), 'error'); });
    return saveQueue;
  }
  const setMode = (value: string) => sidebarSortModes.includes(value as SidebarSortMode) ? save(() => ({ sidebar_sort_mode: value })) : Promise.resolve();
  const togglePin = (key: string) => save(() => {
    const next = new Set(pinned.value);
    if (next.has(key)) next.delete(key); else next.add(key);
    return { sidebar_pinned_items: JSON.stringify([...next]) };
  });
  return { mode, isPinned, setMode, togglePin, compareCategories, compareFeeds };
}
