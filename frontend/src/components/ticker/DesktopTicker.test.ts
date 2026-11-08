import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils';
import { Call } from '@wailsio/runtime';
import DesktopTicker from './DesktopTicker.vue';
import { setSettingsFromRawData } from '@/composables/core/useSettings';
import { generateInitialSettings } from '@/composables/core/useSettings.generated';
import { openInBrowser } from '@/utils/browser';

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key, locale: { value: 'en' } }),
}));
vi.mock('@/utils/browser', () => ({ openInBrowser: vi.fn() }));
vi.mock('@wailsio/runtime', () => ({
  Call: { ByName: vi.fn(async () => undefined) },
  Events: { On: () => () => {} },
}));

describe('compact desktop ticker', () => {
  let wrapper: VueWrapper;
  let fetchMock: ReturnType<typeof vi.fn>;
  beforeEach(() => {
    vi.useFakeTimers();
    vi.clearAllMocks();
    vi.stubGlobal('matchMedia', () => ({ matches: false }));
    vi.stubGlobal(
      'ResizeObserver',
      class {
        observe() {}
        disconnect() {}
      }
    );
    vi.stubGlobal('requestAnimationFrame', () => 1);
    vi.stubGlobal('cancelAnimationFrame', () => {});
    setSettingsFromRawData(
      Object.fromEntries(
        Object.entries({ ...generateInitialSettings(), ticker_enabled: true }).map(
          ([key, value]) => [key, String(value)]
        )
      )
    );
    fetchMock = vi.fn(async (url: string) =>
      Response.json(
        url.startsWith('/api/articles?')
          ? [
              {
                id: 1,
                feed_id: 2,
                title: 'A quiet headline',
                feed_title: 'A source',
                url: 'https://example.org/1',
                published_at: '2026-11-08T00:00:00Z',
              },
            ]
          : {}
      )
    );
    vi.stubGlobal('fetch', fetchMock);
  });
  afterEach(() => {
    wrapper?.unmount();
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });
  it('opens headlines in the reader, supports browser modifiers and keyboard pause', async () => {
    wrapper = mount(DesktopTicker);
    await flushPromises();
    await wrapper.get('.ticker-headline').trigger('click');
    expect(Call.ByName).toHaveBeenCalledWith('main.DesktopCapabilities.OpenTickerArticle', 1, 2);
    await wrapper.get('.ticker-headline').trigger('click', { ctrlKey: true });
    expect(openInBrowser).toHaveBeenCalledWith('https://example.org/1');
    document.dispatchEvent(
      new KeyboardEvent('keydown', { code: 'Space', key: ' ', bubbles: true })
    );
    await wrapper.vm.$nextTick();
    expect(wrapper.classes()).toContain('is-paused');
  });
  it('turns the feature off without changing other preferences when dismissed', async () => {
    wrapper = mount(DesktopTicker);
    await flushPromises();
    await wrapper.get('button[aria-label="setting.ticker.disable"]').trigger('click');
    await flushPromises();
    const request = fetchMock.mock.calls.find(([, options]) => options?.method === 'POST');
    expect(JSON.parse(String(request?.[1]?.body))).toEqual({ ticker_enabled: 'false' });
    expect(Call.ByName).toHaveBeenCalledWith('main.DesktopCapabilities.ApplyTickerSettings');
  });
});
