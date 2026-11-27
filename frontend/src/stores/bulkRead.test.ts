import { afterEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useSidebar } from '@/composables/core/useSidebar';
import { setSettingsFromRawData } from '@/composables/core/useSettings';
import { useAppStore } from './app';
import type { Article, Feed } from '@/types/models';

vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ locale: { value: 'en' }, t: (key: string) => key }),
}));

afterEach(() => vi.unstubAllGlobals());

function markFolderAsRead(category: string) {
  setSettingsFromRawData({ confirm_mark_as_read: 'false' });
  return new Promise<void>((resolve, reject) => {
    window.addEventListener(
      'open-context-menu',
      (event) => {
        const { callback, data } = (event as CustomEvent<{
          callback: (action: string, categoryName: string) => Promise<void>;
          data: string;
        }>).detail;
        callback('markAllRead', data).then(resolve, reject);
      },
      { once: true }
    );
    useSidebar().onCategoryContextMenu(new MouseEvent('contextmenu'), category);
  });
}

describe('mark all as read', () => {
  it('leaves local reading state unchanged on HTTP errors', async () => {
    setActivePinia(createPinia());
    const store = useAppStore();
    store.articles = [{ id: 1, feed_id: 7, is_read: false } as Article];
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 500 })));
    expect(await store.markAllAsRead(7)).toBe(false);
    expect(store.articles[0].is_read).toBe(false);
  });

  it('updates nested folders and both sidebar count sets after success', async () => {
    setActivePinia(createPinia());
    const store = useAppStore();
    store.feeds = [
      { id: 1, category: 'Tech/Go' },
      { id: 2, category: 'Technology' },
    ] as Feed[];
    store.articles = [
      { id: 1, feed_id: 1, is_read: false },
      { id: 2, feed_id: 2, is_read: false },
    ] as Article[];
    const fetch = vi.fn(async (_input: RequestInfo | URL) => new Response('{}'));
    vi.stubGlobal('fetch', fetch);
    const toast = vi.fn();
    vi.stubGlobal('showToast', toast);
    await markFolderAsRead('Tech');
    expect(store.articles.map((article) => article.is_read)).toEqual([true, false]);
    expect(fetch).toHaveBeenCalledWith('/api/articles/mark-all-read?category=Tech', {
      method: 'POST',
    });
    expect(fetch.mock.calls.map(([url]) => url)).toContain('/api/articles/unread-counts');
    expect(fetch.mock.calls.map(([url]) => url)).toContain('/api/articles/filter-counts');
    expect(toast).toHaveBeenCalledWith('article.action.markedAllAsRead', 'success');
  });

  it('keeps uncategorized articles unchanged and reports an error if the request fails', async () => {
    setActivePinia(createPinia());
    const store = useAppStore();
    store.feeds = [{ id: 1, category: '' }] as Feed[];
    store.articles = [{ id: 1, feed_id: 1, is_read: false }] as Article[];
    const fetch = vi.fn().mockResolvedValue(new Response('', { status: 500 }));
    const toast = vi.fn();
    vi.stubGlobal('fetch', fetch);
    vi.stubGlobal('showToast', toast);

    await markFolderAsRead('uncategorized');

    expect(fetch).toHaveBeenCalledExactlyOnceWith('/api/articles/mark-all-read?category=', {
      method: 'POST',
    });
    expect(store.articles[0].is_read).toBe(false);
    expect(toast).toHaveBeenCalledWith('article.action.batchReadUpdateFailed', 'error');
  });
});
