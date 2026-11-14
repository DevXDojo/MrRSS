import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils';
import { Call } from '@wailsio/runtime';
import DesktopTicker from './DesktopTicker.vue';
import { setSettingsFromRawData } from '@/composables/core/useSettings';
import { generateInitialSettings } from '@/composables/core/useSettings.generated';
import { openInBrowser } from '@/utils/browser';

const marquee = vi.hoisted(() => ({ paused: undefined as undefined | (() => boolean) }));
vi.mock('./useTickerMarquee', () => ({
  useTickerMarquee: (
    _viewport: unknown,
    _group: unknown,
    _track: unknown,
    options: { paused: () => boolean }
  ) => {
    marquee.paused = options.paused;
    return { copies: 2, step: vi.fn() };
  },
}));

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
  it('opens headlines using the configured destination and pauses through its toolbar', async () => {
    wrapper = mount(DesktopTicker);
    await flushPromises();
    await wrapper.get('.ticker-headline').trigger('click');
    expect(Call.ByName).toHaveBeenCalledWith(
      'MrRSS/internal/desktopapi.DesktopCapabilities.OpenTickerArticle',
      1,
      2
    );
    await wrapper.get('.ticker-headline').trigger('auxclick', { button: 1 });
    expect(openInBrowser).toHaveBeenCalledWith('https://example.org/1');
    await wrapper.get('button[aria-label="setting.ticker.pause"]').trigger('click');
    expect(wrapper.classes()).toContain('is-paused');
  });
  it('turns the feature off without changing other preferences when dismissed', async () => {
    wrapper = mount(DesktopTicker);
    await flushPromises();
    await wrapper.get('button[aria-label="setting.ticker.disable"]').trigger('click');
    await flushPromises();
    const request = fetchMock.mock.calls.find(([, options]) => options?.method === 'POST');
    expect(JSON.parse(String(request?.[1]?.body))).toEqual({ ticker_enabled: 'false' });
    expect(Call.ByName).toHaveBeenCalledWith(
      'MrRSS/internal/desktopapi.DesktopCapabilities.ApplyTickerSettings'
    );
  });
  it('resumes after mouse playback controls while pausing for keyboard focus', async () => {
    wrapper = mount(DesktopTicker);
    await flushPromises();
    const headline = wrapper.get('.ticker-headline');
    const matches = vi.spyOn(headline.element, 'matches');
    matches.mockReturnValue(false);
    await headline.trigger('focusin');
    await wrapper.trigger('mouseenter');
    expect(marquee.paused?.()).toBe(true);
    await wrapper.trigger('mouseleave');
    expect(marquee.paused?.()).toBe(false);
    matches.mockReturnValue(true);
    await headline.trigger('focusin');
    expect(marquee.paused?.()).toBe(true);
    await headline.trigger('focusout');
    expect(marquee.paused?.()).toBe(false);
    matches.mockRestore();
  });
});
