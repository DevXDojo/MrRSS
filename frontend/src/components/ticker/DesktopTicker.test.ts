import { afterEach, describe, expect, it, vi } from 'vitest';
import { flushPromises, mount } from '@vue/test-utils';
import DesktopTicker from './DesktopTicker.vue';
import { setSettingsFromRawData } from '@/composables/core/useSettings';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));
vi.mock('@/utils/browser', () => ({ openInBrowser: vi.fn() }));

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

describe('desktop ticker data and preferences', () => {
  it('filters articles via HTTP and saves only ticker preferences', async () => {
    vi.useFakeTimers();
    vi.stubGlobal('matchMedia', () => ({ matches: false }));
    vi.stubGlobal(
      'ResizeObserver',
      class {
        observe() {}
        disconnect() {}
      }
    );
    setSettingsFromRawData({});
    const fetchMock = vi.fn(
      async (url: string) =>
        new Response(
          JSON.stringify(
            url === '/api/feeds'
              ? [{ id: 2, title: 'Selected feed' }]
              : url.startsWith('/api/articles?')
                ? [{ id: 1, title: 'A post', url: 'https://example.org/1' }]
                : {}
          )
        )
    );
    vi.stubGlobal('fetch', fetchMock);
    const wrapper = mount(DesktopTicker);
    try {
      await flushPromises();
      await wrapper.get('select').setValue('2');
      await flushPromises();
      expect(fetchMock.mock.calls.some(([url]) => url.includes('feed_id=2'))).toBe(true);
      await vi.advanceTimersByTimeAsync(500);
      await flushPromises();
      const saved = fetchMock.mock.calls.find(([url]) => url === '/api/settings');
      expect(saved).toBeDefined();
      const options = (saved as unknown as [string, RequestInit])[1];
      expect(JSON.parse(String(options.body))).toEqual({
        ticker_feed_id: '2',
        ticker_speed: '60',
        ticker_font_size: '18',
        ticker_text_color: '#f8fafc',
        ticker_background_color: '#1e293b',
        ticker_background_color_end: '#0f172a',
      });
    } finally {
      wrapper.unmount();
    }
  });
});
