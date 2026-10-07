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
  it('coalesces bulk refresh updates while keeping progress responsive and flushes completion', async () => {
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
    store.pollProgress();
    await vi.advanceTimersByTimeAsync(3000);
    expect(fetchMock.mock.calls.filter(([url]) => url === '/api/progress')).toHaveLength(7);
    expect(fetchMock.mock.calls.filter(([url]) => url.startsWith('/api/articles?'))).toHaveLength(
      2
    );
    expect(
      fetchMock.mock.calls.filter(([url]) => url === '/api/articles/unread-counts')
    ).toHaveLength(2);
    expect(settingsEvent).not.toHaveBeenCalled();
    expect(contentEvent.mock.calls.every(([event]) => event.detail.recoveryOnly)).toBe(true);
    running = false;
    await vi.advanceTimersByTimeAsync(500);
    expect(fetchMock.mock.calls.filter(([url]) => url.startsWith('/api/articles?'))).toHaveLength(
      3
    );
    expect(settingsEvent).toHaveBeenCalledTimes(1);
    expect(contentEvent.mock.lastCall?.[0].detail.recoveryOnly).toBe(false);
    window.removeEventListener('settings-updated', settingsEvent);
    window.removeEventListener('article-content-updated', contentEvent);
  });
});
