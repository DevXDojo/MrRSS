import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { advertisementText, useAdvertisementsStore } from './advertisements';

const ad = {
  id: 'partner',
  title: { en: 'AI APIs' },
  description: { en: 'Offer' },
  url: 'https://example.com',
};

describe('partner advertisement preferences', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-10-08T00:00:00Z'));
  });
  afterEach(() => {
    useAdvertisementsStore().stop();
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('shares a persisted snooze across popup/cards and expires after a week', async () => {
    const until = Date.now() / 1000 + 7 * 24 * 60 * 60;
    const fetchMock = vi.fn(
      async (url: string) =>
        new Response(
          JSON.stringify(
            url.endsWith('/snooze') ? { snoozed_until: until } : { ads: [ad], snoozed_until: 0 }
          )
        )
    );
    vi.stubGlobal('fetch', fetchMock);
    const store = useAdvertisementsStore();
    await store.load();
    expect(store.visibleAds).toEqual([ad]);
    expect(await store.snooze()).toBe(true);
    expect(store.visibleAds).toEqual([]);
    expect(useAdvertisementsStore().visibleAds).toEqual([]);
    fetchMock.mockImplementation(
      async () => new Response(JSON.stringify({ ads: [ad], snoozed_until: until }))
    );
    store.start();
    await vi.advanceTimersByTimeAsync(0);
    vi.setSystemTime(new Date(until * 1000));
    await vi.advanceTimersByTimeAsync(60 * 1000);
    expect(store.visibleAds).toEqual([ad]);
    store.stop();
    expect(vi.getTimerCount()).toBe(0);
  });

  it('does not hide an offer when saving fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async (url: string) =>
        url.endsWith('/snooze')
          ? new Response('', { status: 500 })
          : new Response(JSON.stringify({ ads: [ad], snoozed_until: 0 }))
      )
    );
    const store = useAdvertisementsStore();
    await store.load();
    expect(await store.snooze()).toBe(false);
    expect(store.visibleAds).toEqual([ad]);
    expect(store.saving).toBe(false);
  });

  it('prevents an older configuration response from undoing a snooze', async () => {
    let resolveLoad!: (response: Response) => void;
    vi.stubGlobal(
      'fetch',
      vi.fn((url: string) =>
        url.endsWith('/snooze')
          ? Promise.resolve(
              new Response(JSON.stringify({ snoozed_until: Date.now() / 1000 + 604800 }))
            )
          : new Promise<Response>((resolve) => {
              resolveLoad = resolve;
            })
      )
    );
    const store = useAdvertisementsStore();
    const loading = store.load();
    await store.snooze();
    resolveLoad(new Response(JSON.stringify({ ads: [ad], snoozed_until: 0 })));
    await loading;
    expect(store.visibleAds).toEqual([]);
  });

  it('selects exact locale, language, and fallback text', () => {
    expect(advertisementText({ 'zh-CN': '简体', zh: '中文', en: 'English' }, 'zh-CN')).toBe('简体');
    expect(advertisementText({ zh: '中文', en: 'English' }, 'zh-TW')).toBe('中文');
    expect(advertisementText({ en: 'English' }, 'fr')).toBe('English');
    expect(advertisementText(undefined, 'en')).toBe('');
  });
});
