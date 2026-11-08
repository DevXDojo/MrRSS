import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, ref } from 'vue';
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils';
import { generateInitialSettings } from '@/composables/core/useSettings.generated';
import { useTickerArticles } from './useTickerArticles';

describe('ticker article refreshes', () => {
  let wrapper: VueWrapper;
  let api: ReturnType<typeof useTickerArticles>;
  const settings = ref(generateInitialSettings());
  const visible = ref(true);
  const article = { id: 1, title: 'Original headline', url: 'https://example.org/1' };
  beforeEach(() => {
    vi.useFakeTimers();
    settings.value = generateInitialSettings();
    visible.value = true;
  });
  afterEach(() => {
    wrapper?.unmount();
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });
  function start() {
    wrapper = mount(
      defineComponent({
        setup() {
          api = useTickerArticles(settings, visible);
        },
        template: '<div />',
      })
    );
  }
  it('retains articles on temporary errors and does no polling while hidden', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(Response.json([article]))
      .mockResolvedValueOnce(new Response('', { status: 500 }))
      .mockImplementation(async () => Response.json([article]));
    vi.stubGlobal('fetch', fetchMock);
    start();
    await flushPromises();
    const original = api.articles.value;
    await vi.advanceTimersByTimeAsync(60000);
    expect(api.failed.value).toBe(true);
    expect(api.articles.value).toBe(original);
    visible.value = false;
    await vi.advanceTimersByTimeAsync(180000);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    visible.value = true;
    await flushPromises();
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(api.failed.value).toBe(false);
    expect(api.articles.value).toBe(original);
  });
  it('ignores an old response after changing sources and applies the unread filter', async () => {
    let finishOld!: (response: Response) => void;
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise<Response>((resolve) => {
            finishOld = resolve;
          })
      )
      .mockResolvedValueOnce(Response.json([{ ...article, id: 2 }]));
    vi.stubGlobal('fetch', fetchMock);
    start();
    settings.value.ticker_feed_id = 7;
    settings.value.ticker_only_unread = true;
    await flushPromises();
    finishOld(Response.json([article]));
    await flushPromises();
    expect(api.articles.value.map((item) => item.id)).toEqual([2]);
    const query = new URL(fetchMock.mock.calls[1][0], 'http://localhost').searchParams;
    expect(query.get('feed_id')).toBe('7');
    expect(query.get('filter')).toBe('unread');
    expect(fetchMock.mock.calls[0][1].signal.aborted).toBe(true);
  });
});
