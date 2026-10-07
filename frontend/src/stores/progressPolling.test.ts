import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useAppStore } from './app';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ locale: { value: 'en' } }) }));

describe('background refresh observation', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.useFakeTimers();
  });
  afterEach(() => {
    useAppStore().stopProgressPolling();
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });
  it('detects a refresh completed between polls without clearing selection or duplicating timers', async () => {
    const store = useAppStore();
    store.currentFeedId = 2;
    store.currentArticleId = 5;
    store.articles = [{ id: 5, title: 'Selected' } as import('@/types/models').Article];
    let revision = 0;
    const fetchMock = vi.fn(
      async (url: string) =>
        new Response(
          JSON.stringify(
            url === '/api/progress'
              ? { is_running: false, article_revision: revision }
              : url === '/api/feeds' || url.startsWith('/api/articles?')
                ? []
                : {}
          )
        )
    );
    vi.stubGlobal('fetch', fetchMock);
    store.pollProgress();
    store.pollProgress();
    await vi.advanceTimersByTimeAsync(0);
    expect(fetchMock.mock.calls.filter(([url]) => url === '/api/progress')).toHaveLength(1);
    revision++;
    await vi.advanceTimersByTimeAsync(5000);
    expect(fetchMock.mock.calls.filter(([url]) => url.startsWith('/api/articles?'))).toHaveLength(
      1
    );
    expect(store.currentFeedId).toBe(2);
    expect(store.navigableArticles.find((a) => a.id === store.currentArticleId)?.title).toBe(
      'Selected'
    );
    await vi.advanceTimersByTimeAsync(5000);
    expect(fetchMock.mock.calls.filter(([url]) => url.startsWith('/api/articles?'))).toHaveLength(
      1
    );
    store.stopProgressPolling();
    await vi.advanceTimersByTimeAsync(5000);
    expect(fetchMock.mock.calls.filter(([url]) => url === '/api/progress')).toHaveLength(3);
  });
  it('retries after a temporary status failure', async () => {
    const fetchMock = vi
      .fn()
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValue(new Response('{}'));
    vi.stubGlobal('fetch', fetchMock);
    useAppStore().pollProgress();
    await vi.advanceTimersByTimeAsync(5000);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
  it('keeps the reading viewport stable during bulk refresh and publishes the finished batch once', async () => {
    let revision = 0;
    let running = true;
    const settingsEvent = vi.fn();
    const contentEvent = vi.fn();
    window.addEventListener('settings-updated', settingsEvent);
    window.addEventListener('article-content-updated', contentEvent);
    const fetchMock = vi.fn(
      async (url: string) =>
        new Response(
          JSON.stringify(
            url === '/api/progress'
              ? { is_running: running, article_revision: revision++ }
              : url === '/api/feeds' || url.startsWith('/api/articles?')
                ? []
                : {}
          )
        )
    );
    vi.stubGlobal('fetch', fetchMock);
    const store = useAppStore();
    store.articles = [
      { id: 5, title: 'Selected', translated_title: '已翻译' } as import('@/types/models').Article,
    ];
    store.currentArticleId = 5;
    const visibleArticles = store.articles;
    store.pollProgress();
    await vi.advanceTimersByTimeAsync(30000);
    expect(fetchMock.mock.calls.filter(([url]) => url === '/api/progress')).toHaveLength(61);
    expect(fetchMock.mock.calls.filter(([url]) => url.startsWith('/api/articles?'))).toHaveLength(
      0
    );
    expect(store.articles).toBe(visibleArticles);
    expect(store.articles[0].translated_title).toBe('已翻译');
    expect(
      fetchMock.mock.calls.filter(([url]) => url === '/api/articles/unread-counts')
    ).toHaveLength(0);
    expect(settingsEvent).not.toHaveBeenCalled();
    expect(contentEvent).not.toHaveBeenCalled();
    running = false;
    await vi.advanceTimersByTimeAsync(500);
    expect(fetchMock.mock.calls.filter(([url]) => url.startsWith('/api/articles?'))).toHaveLength(
      1
    );
    expect(store.currentArticleId).toBe(5);
    expect(settingsEvent).toHaveBeenCalledTimes(1);
    expect(contentEvent.mock.lastCall?.[0].detail.recoveryOnly).toBe(false);
    window.removeEventListener('settings-updated', settingsEvent);
    window.removeEventListener('article-content-updated', contentEvent);
  });

  it('updates completed-feed counts without replacing the article viewport', async () => {
    let tasks = 10;
    const store = useAppStore();
    store.articles = [{ id: 5, title: 'Reading' } as import('@/types/models').Article];
    const visibleArticles = store.articles;
    const fetchMock = vi.fn(
      async (url: string) =>
        new Response(
          JSON.stringify(
            url === '/api/progress'
              ? { is_running: true, article_revision: 1, pool_task_count: tasks }
              : url === '/api/feeds'
                ? []
                : {}
          )
        )
    );
    vi.stubGlobal('fetch', fetchMock);
    store.pollProgress();
    await vi.advanceTimersByTimeAsync(0);
    tasks = 9;
    await vi.advanceTimersByTimeAsync(500);
    expect(
      fetchMock.mock.calls.filter(([url]) => url === '/api/articles/unread-counts')
    ).toHaveLength(1);
    expect(fetchMock.mock.calls.filter(([url]) => url.startsWith('/api/articles?'))).toHaveLength(
      0
    );
    expect(store.articles).toBe(visibleArticles);
  });
});
