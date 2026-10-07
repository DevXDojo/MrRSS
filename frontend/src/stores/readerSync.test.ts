import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useAppStore } from './app';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ locale: { value: 'en' } }) }));

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
          return new Response(JSON.stringify({ last_sync_time: null, is_syncing: states[provider] }));
        }
      }
      return new Response(JSON.stringify(url.startsWith('/api/articles?') || url === '/api/feeds' ? [] : {}));
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
});
