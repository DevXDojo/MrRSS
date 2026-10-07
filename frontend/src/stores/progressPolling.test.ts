import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useAppStore } from './app';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ locale: { value: 'en' } }) }));

describe('background refresh observation', () => {
  beforeEach(() => { setActivePinia(createPinia()); vi.useFakeTimers(); });
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
    const fetchMock = vi.fn(async (url: string) => new Response(JSON.stringify(
      url === '/api/progress' ? { is_running: false, article_revision: revision } :
      url === '/api/feeds' || url.startsWith('/api/articles?') ? [] : {}
    )));
    vi.stubGlobal('fetch', fetchMock);
    store.pollProgress();
    store.pollProgress();
    await vi.advanceTimersByTimeAsync(0);
    expect(fetchMock.mock.calls.filter(([url]) => url === '/api/progress')).toHaveLength(1);
    revision++;
    await vi.advanceTimersByTimeAsync(5000);
    expect(fetchMock.mock.calls.filter(([url]) => url.startsWith('/api/articles?'))).toHaveLength(1);
    expect(store.currentFeedId).toBe(2);
    expect(store.navigableArticles.find((a) => a.id === store.currentArticleId)?.title).toBe('Selected');
    await vi.advanceTimersByTimeAsync(5000);
    expect(fetchMock.mock.calls.filter(([url]) => url.startsWith('/api/articles?'))).toHaveLength(1);
    store.stopProgressPolling();
    await vi.advanceTimersByTimeAsync(5000);
    expect(fetchMock.mock.calls.filter(([url]) => url === '/api/progress')).toHaveLength(3);
  });
  it('retries after a temporary status failure', async () => {
    const fetchMock = vi.fn().mockRejectedValueOnce(new Error('offline')).mockResolvedValue(new Response('{}'));
    vi.stubGlobal('fetch', fetchMock);
    useAppStore().pollProgress();
    await vi.advanceTimersByTimeAsync(5000);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
});
