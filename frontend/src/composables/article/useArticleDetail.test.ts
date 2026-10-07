import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { flushPromises, mount } from '@vue/test-utils';
import { useArticleDetail } from './useArticleDetail';
import { useAppStore } from '@/stores/app';
import { clearArticleContentCache } from '@/utils/articleContentCache';
import type { Article } from '@/types/models';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ locale: { value: 'en' }, t: (key: string) => key }) }));
vi.mock('@/utils/mediaProxy', () => ({ isMediaCacheEnabled: async () => false, proxyImagesInHtml: (html: string) => html }));

beforeEach(() => { setActivePinia(createPinia()); clearArticleContentCache(); });
afterEach(() => vi.unstubAllGlobals());

describe('reader content after feed refresh', () => {
  it('recovers empty content and preserves a successful body if a later request fails', async () => {
    const store = useAppStore();
    store.articles = [{ id: 1, feed_id: 1, url: 'https://example.org/a' } as Article];
    store.currentArticleId = 1;
    let body = '';
    let fail = false;
    vi.stubGlobal('fetch', vi.fn(async (url: string) => url.startsWith('/api/articles/content')
      ? new Response(JSON.stringify({ content: body }), { status: fail ? 503 : 200 })
      : new Response('{}')));
    let detail!: ReturnType<typeof useArticleDetail>;
    const wrapper = mount({ setup() { detail = useArticleDetail(); return {}; }, template: '<div />' });
    await flushPromises();
    expect(detail.articleContent.value).toBe('');
    body = '<p>Recovered</p>';
    window.dispatchEvent(new CustomEvent('article-content-updated'));
    await flushPromises();
    expect(detail.articleContent.value).toBe(body);
    clearArticleContentCache();
    fail = true;
    window.dispatchEvent(new CustomEvent('article-content-updated'));
    await flushPromises();
    expect(detail.articleContent.value).toBe(body);
    expect(detail.isLoadingContent.value).toBe(false);
    wrapper.unmount();
  });
});
