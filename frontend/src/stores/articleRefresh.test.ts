import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useAppStore } from './app';
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
});
