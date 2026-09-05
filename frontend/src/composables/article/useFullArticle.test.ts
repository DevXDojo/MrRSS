import { describe, it, expect, vi, afterEach } from 'vitest';
import { defineComponent, ref, nextTick } from 'vue';
import { mount, flushPromises } from '@vue/test-utils';
import type { Article } from '@/types/models';
import { useFullArticle } from './useFullArticle';
vi.mock('@/utils/mediaProxy', () => ({
  isMediaCacheEnabled: async () => false,
  proxyImagesInHtml: (s: string) => s,
}));
afterEach(() => vi.unstubAllGlobals());
function fixture(auto = true) {
  const article = ref({ id: 1, url: 'https://example.org/1' } as Article);
  const enabled = ref(true);
  const loading = ref(false);
  const onError = vi.fn();
  const onContent = vi.fn();
  let result!: ReturnType<typeof useFullArticle>;
  const wrapper = mount(
    defineComponent({
      setup() {
        result = useFullArticle({
          article: () => article.value,
          enabled: () => enabled.value,
          automatic: () => auto,
          loading: () => loading.value,
          onError,
          onContent,
          onSuccess: vi.fn(),
        });
        return () => null;
      },
    })
  );
  return {
    article,
    enabled,
    loading,
    onError,
    onContent,
    wrapper,
    get result() {
      return result;
    },
  };
}
function response(content: string) {
  return { ok: true, json: async () => ({ content }) };
}
describe('full article loading', () => {
  it('loads empty RSS articles once and keeps full text when RSS loading changes', async () => {
    const fetch = vi.fn().mockResolvedValue(response('<p>Full article</p>'));
    vi.stubGlobal('fetch', fetch);
    const f = fixture();
    await flushPromises();
    expect(f.result.fullArticleContent.value).toContain('Full article');
    f.loading.value = true;
    await nextTick();
    f.loading.value = false;
    await flushPromises();
    expect(fetch).toHaveBeenCalledTimes(1);
    f.wrapper.unmount();
  });
  it('rejects old full-text responses after switching A to B to A', async () => {
    const pending: Array<(value: unknown) => void> = [];
    vi.stubGlobal(
      'fetch',
      vi.fn(() => new Promise((resolve) => pending.push(resolve)))
    );
    const f = fixture();
    await nextTick();
    f.article.value = { id: 2 } as Article;
    await nextTick();
    f.article.value = { id: 1 } as Article;
    await nextTick();
    pending[2](response('Fresh A'));
    await flushPromises();
    pending[0](response('Old A'));
    pending[1](response('Old B'));
    await flushPromises();
    expect(f.result.fullArticleContent.value).toBe('Fresh A');
    expect(f.onContent).toHaveBeenCalledTimes(1);
    f.wrapper.unmount();
  });
  it('allows manual retry after automatic failure without a retry loop', async () => {
    const fetch = vi
      .fn()
      .mockResolvedValueOnce({ ok: false })
      .mockResolvedValue(response('Recovered'));
    vi.stubGlobal('fetch', fetch);
    const f = fixture();
    await flushPromises();
    expect(f.onError).not.toHaveBeenCalled();
    await f.result.fetchFullArticle();
    expect(f.result.fullArticleContent.value).toBe('Recovered');
    expect(fetch).toHaveBeenCalledTimes(2);
    f.wrapper.unmount();
  });
  it('honors disabled auto expansion and aborts on unmount', async () => {
    const fetch = vi.fn().mockResolvedValue(response('Manual'));
    vi.stubGlobal('fetch', fetch);
    const f = fixture(false);
    await flushPromises();
    expect(fetch).not.toHaveBeenCalled();
    f.enabled.value = false;
    await f.result.fetchFullArticle();
    expect(fetch).not.toHaveBeenCalled();
    f.enabled.value = true;
    await f.result.fetchFullArticle();
    expect(fetch).toHaveBeenCalledTimes(1);
    const signal = fetch.mock.calls[0][1].signal;
    f.wrapper.unmount();
    expect(signal.aborted).toBe(true);
  });
});
