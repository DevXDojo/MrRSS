import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useAppStore } from './app';
import { setSettingsFromRawData } from '@/composables/core/useSettings';

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ locale: { value: 'en' }, t: (key: string) => key }),
}));

describe('reader sync polling', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.useFakeTimers();
  });
  afterEach(() => {
    useAppStore().stopFreshRSSStatusPolling();
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('refreshes after either provider finishes, even when its timestamp stays the same', async () => {
    const store = useAppStore();
    const states = { freshrss: true, miniflux: true };
    const fetchMock = vi.fn(async (url: string) => {
      if (url === '/api/settings') {
        return new Response(JSON.stringify({ freshrss_enabled: 'true', miniflux_enabled: 'true' }));
      }
      for (const provider of ['freshrss', 'miniflux'] as const) {
        if (url === `/api/${provider}/status`) {
          return new Response(
            JSON.stringify({ last_sync_time: null, is_syncing: states[provider] })
          );
        }
      }
      return new Response(
        JSON.stringify(url.startsWith('/api/articles?') || url === '/api/feeds' ? [] : {})
      );
    });
    vi.stubGlobal('fetch', fetchMock);
    await store.startFreshRSSStatusPolling();
    expect(fetchMock.mock.calls.filter(([url]) => url === '/api/feeds')).toHaveLength(0);
    states.miniflux = false;
    await vi.advanceTimersByTimeAsync(5000);
    expect(fetchMock.mock.calls.filter(([url]) => url === '/api/feeds')).toHaveLength(1);
    states.freshrss = false;
    await vi.advanceTimersByTimeAsync(5000);
    expect(fetchMock.mock.calls.filter(([url]) => url === '/api/feeds')).toHaveLength(2);
    await vi.advanceTimersByTimeAsync(5000);
    expect(fetchMock.mock.calls.filter(([url]) => url === '/api/feeds')).toHaveLength(2);
  });

  it('manually syncs both readers when a feed is selected, accepting an active sync', async () => {
    setSettingsFromRawData({ freshrss_enabled: 'true', miniflux_enabled: 'true' });
    const store = useAppStore();
    store.currentFeedId = 1;
    const fetchMock = vi.fn(async (url: string) => {
      if (url === '/api/settings') {
        return new Response(JSON.stringify({ freshrss_enabled: 'true', miniflux_enabled: 'true' }));
      }
      if (url === '/api/freshrss/sync') return new Response('', { status: 409 });
      return new Response(JSON.stringify({ is_running: false }));
    });
    vi.stubGlobal('fetch', fetchMock);
    const refresh = store.refreshFeeds();
    await vi.advanceTimersByTimeAsync(500);
    await refresh;
    expect(fetchMock).toHaveBeenCalledWith('/api/freshrss/sync', { method: 'POST' });
    expect(fetchMock).toHaveBeenCalledWith('/api/miniflux/sync', { method: 'POST' });
    expect(fetchMock).not.toHaveBeenCalledWith('/api/refresh', { method: 'POST' });
    store.stopProgressPolling();
  });
});
