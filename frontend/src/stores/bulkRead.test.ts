import { afterEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useAppStore } from './app';
import type { Article, Feed } from '@/types/models';

vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ locale: { value: 'en' }, t: (key: string) => key }),
}));

afterEach(() => vi.unstubAllGlobals());

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
    expect(await store.markAllAsRead(undefined, 'Tech')).toBe(true);
    expect(store.articles.map((article) => article.is_read)).toEqual([true, false]);
    expect(fetch.mock.calls.map(([url]) => url)).toContain('/api/articles/filter-counts');
  });
});
