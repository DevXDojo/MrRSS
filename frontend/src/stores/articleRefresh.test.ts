import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useAppStore } from './app';
import { useArticleFilter } from '@/composables/article/useArticleFilter';
import type { Article } from '@/types/models';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ locale: { value: 'en' } }) }));

const article = (id: number, title = `Article ${id}`) => ({ id, title }) as Article;
const response = (articles: Article[]) => new Response(JSON.stringify(articles));

describe('article refresh integration', () => {
  beforeEach(() => setActivePinia(createPinia()));
  afterEach(() => vi.unstubAllGlobals());

  it('keeps the reader populated and restores the retained article to its page without duplicates', async () => {
    const store = useAppStore();
    store.articles = [article(75)];
    store.currentArticleId = 75;
    let finishRefresh!: (value: Response) => void;
    const fetchMock = vi.fn().mockImplementationOnce(
      () =>
        new Promise<Response>((resolve) => {
          finishRefresh = resolve;
        })
    );
    vi.stubGlobal('fetch', fetchMock);

    const refresh = store.fetchArticles(false, true);
    expect(store.navigableArticles.find((item) => item.id === store.currentArticleId)?.id).toBe(75);
    finishRefresh(response(Array.from({ length: 50 }, (_, i) => article(i + 1))));
    await refresh;
    expect(store.navigableArticles.find((item) => item.id === store.currentArticleId)?.id).toBe(75);

    fetchMock.mockResolvedValueOnce(
      response(Array.from({ length: 50 }, (_, i) => article(i + 51, 'Fresh')))
    );
    await store.loadMore();
    expect(store.articles.map((item) => item.id)).toEqual(
      Array.from({ length: 100 }, (_, i) => i + 1)
    );
    expect(store.navigableArticles.find((item) => item.id === store.currentArticleId)?.title).toBe(
      'Fresh'
    );
  });

  it('lets navigation supersede an in-flight background refresh', async () => {
    const store = useAppStore();
    store.articles = [article(75)];
    store.currentArticleId = 75;
    let finishRefresh!: (value: Response) => void;
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockImplementationOnce(
          () =>
            new Promise<Response>((resolve) => {
              finishRefresh = resolve;
            })
        )
        .mockResolvedValueOnce(response([article(200)]))
    );
    const refresh = store.fetchArticles(false, true);
    store.currentFeedId = 2;
    await store.fetchArticles();
    finishRefresh(response([article(1)]));
    await refresh;
    expect(store.articles.map((item) => item.id)).toEqual([200]);
    expect(store.isLoading).toBe(false);
  });

  it('keeps the selected article when background refresh fails', async () => {
    const store = useAppStore();
    store.articles = [article(75)];
    store.currentArticleId = 75;
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{}', { status: 500 })));
    await store.fetchArticles(false, true);
    expect(store.navigableArticles.find((item) => item.id === store.currentArticleId)?.id).toBe(75);
    expect(store.isLoading).toBe(false);
  });
  it('returns to the source feed with filters cleared and the selected older article retained', async () => {
    const store = useAppStore();
    const selected = { ...article(75), feed_id: 2, is_read: true };
    store.articles = [article(1)];
    store.articleNavigationContext = [selected];
    store.currentArticleId = 75;
    store.currentFilter = 'favorites';
    store.showOnlyUnread = true;
    store.searchQuery = 'old search';
    store.activeFilters = [{ field: 'title', operator: 'contains', value: 'other' }] as any;
    const fetchMock = vi
      .fn()
      .mockImplementation((url: string) =>
        Promise.resolve(
          new Response(
            JSON.stringify(url.includes('filter-counts') ? {} : [{ ...article(2), feed_id: 2 }])
          )
        )
      );
    vi.stubGlobal('fetch', fetchMock);
    store.selectFeedInArticleList(2, 75);
    await vi.waitFor(() => expect(store.isLoading).toBe(false));
    expect(store.currentFilter).toBe('all');
    expect(store.currentFeedId).toBe(2);
    expect(store.searchQuery).toBe('');
    expect(store.activeFilters).toEqual([]);
    expect(store.showOnlyUnread).toBe(false);
    expect(store.navigableArticles.find((item) => item.id === 75)).toEqual(selected);
    const request = fetchMock.mock.calls.find(([url]) => url.startsWith('/api/articles?'))?.[0];
    expect(request).toContain('feed_id=2');
    expect(request).not.toContain('only_unread=true');
  });
  it('ignores a late filter result after returning to the feed', async () => {
    const store = useAppStore();
    store.activeFilters = [{ field: 'title', operator: 'contains', value: 'old' }] as any;
    const filter = useArticleFilter();
    let finishFilter!: (value: Response) => void;
    vi.stubGlobal(
      'fetch',
      vi.fn().mockImplementation((url: string) => {
        if (url === '/api/articles/filter')
          return new Promise<Response>((resolve) => {
            finishFilter = resolve;
          });
        return Promise.resolve(
          new Response(JSON.stringify(url.includes('filter-counts') ? {} : []))
        );
      })
    );
    const pending = filter.fetchFilteredArticles(store.activeFilters);
    store.selectFeedInArticleList(2);
    finishFilter(new Response(JSON.stringify({ articles: [article(99)], total: 1 })));
    await pending;
    expect(store.filteredArticlesFromServer).toEqual([]);
    expect(store.articles.find((item) => item.id === 99)).toBeUndefined();
    expect(store.isFilterLoading).toBe(false);
  });
});
