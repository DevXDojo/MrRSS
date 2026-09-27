import { afterEach, describe, expect, it, vi } from 'vitest';
import { shallowMount, flushPromises } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { createI18n } from 'vue-i18n';
import { useAppStore } from '@/stores/app';
import { setSettingsFromRawData } from '@/composables/core/useSettings';
import en from '@/i18n/locales/en';
import type { Article } from '@/types/models';
import ArticleList from './ArticleList.vue';

vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({
    locale: { value: 'en' },
    t: (key: string) =>
      key
        .split('.')
        .reduce<unknown>(
          (value, part) =>
            value && typeof value === 'object'
              ? (value as Record<string, unknown>)[part]
              : undefined,
          en
        ) ?? key,
  }),
}));

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  vi.useRealTimers();
});

async function setup() {
  const pinia = createPinia();
  setActivePinia(pinia);
  const store = useAppStore();
  store.showOnlyUnread = false;
  store.currentFeedId = 7;
  store.articles = [{ id: 1, feed_id: 7, is_read: false, published_at: '2026-09-27' } as Article];
  store.hasMore = false;
  const fetch = vi.fn().mockResolvedValue(new Response('{}'));
  // Each request needs its own readable Response body.
  fetch.mockImplementation(async () => new Response('{}'));
  vi.stubGlobal('fetch', fetch);
  const wrapper = shallowMount(ArticleList, {
    global: { plugins: [pinia, createI18n({ legacy: false, locale: 'en', messages: { en } })] },
  });
  await flushPromises();
  setSettingsFromRawData({ confirm_mark_as_read: 'false' });
  return { store, wrapper, fetch };
}

describe('article list bulk read and navigation', () => {
  it('keeps favorites bulk-read within the displayed result set', async () => {
    const f = await setup();
    f.store.currentFilter = 'favorites';
    await flushPromises();
    const mark = vi.spyOn(f.store, 'markAllAsRead');
    await f.wrapper.get(`button[title^="${en.article.action.markAllRead}"]`).trigger('click');
    await flushPromises();
    expect(mark).not.toHaveBeenCalled();
    expect(f.fetch).toHaveBeenCalledWith(
      '/api/articles/read-batch',
      expect.objectContaining({
        body: JSON.stringify({ ids: [1], read: true }),
      })
    );
    f.wrapper.unmount();
  });

  it('marks seen rows only when they leave above the viewport and ignores old observers', async () => {
    const observers: {
      callback: IntersectionObserverCallback;
      options?: IntersectionObserverInit;
    }[] = [];
    vi.stubGlobal(
      'IntersectionObserver',
      class {
        constructor(callback: IntersectionObserverCallback, options?: IntersectionObserverInit) {
          observers.push({ callback, options });
        }
        observe() {}
        unobserve() {}
        disconnect() {}
      }
    );
    const f = await setup();
    setSettingsFromRawData({ scroll_mark_as_read: 'true' });
    const observer = observers.findLast((item) => item.options?.threshold?.toString() === '0,0.6');
    expect(observer).toBeDefined();
    const target = document.createElement('div');
    target.dataset.articleId = '1';
    function emit(visible: boolean, bottom: number) {
      observer!.callback(
        [
          {
            target,
            isIntersecting: visible,
            intersectionRatio: visible ? 1 : 0,
            rootBounds: { top: 0 },
            boundingClientRect: { bottom },
          } as IntersectionObserverEntry,
        ],
        {} as IntersectionObserver
      );
    }
    const reads = () =>
      f.fetch.mock.calls.filter(([url]) => String(url).startsWith('/api/articles/read?'));
    emit(true, 100);
    emit(false, 900);
    expect(reads()).toHaveLength(0);
    emit(true, 100);
    emit(false, -1);
    await flushPromises();
    expect(reads()).toHaveLength(1);
    f.store.articles[0].is_read = false;
    emit(true, 100);
    f.store.currentFeedId = 8;
    emit(false, -1);
    await flushPromises();
    expect(reads()).toHaveLength(1);
    f.wrapper.unmount();
  });

  it('uses the same complete feed scope for the footer and toolbar', async () => {
    const f = await setup();
    const mark = vi.spyOn(f.store, 'markAllAsRead').mockResolvedValue(true);
    vi.useFakeTimers();
    await f.wrapper.get('.article-list-scroll').trigger('scroll');
    await vi.advanceTimersByTimeAsync(210);
    const buttons = f.wrapper
      .findAll('button')
      .filter((button) => button.text().includes(en.article.action.markAllRead));
    expect(buttons).toHaveLength(1);
    await buttons[0].trigger('click');
    expect(mark).toHaveBeenLastCalledWith(7, undefined);
    await f.wrapper.get(`button[title^="${en.article.action.markAllRead}"]`).trigger('click');
    expect(mark).toHaveBeenCalledTimes(2);
    f.wrapper.unmount();
  });

  it('resets a retained list scroll offset when changing feeds', async () => {
    const f = await setup();
    const list = f.wrapper.get('.article-list-scroll').element;
    list.scrollTop = 1200;
    f.store.currentFeedId = 8;
    f.store.articles = [];
    f.store.isLoading = true;
    expect(list.scrollTop).toBe(0);
    await flushPromises();
    list.scrollTop = 500;
    f.store.articles = [{ id: 2, feed_id: 8, published_at: '2026-09-27' } as Article];
    f.store.isLoading = false;
    await flushPromises();
    expect(list.scrollTop).toBe(0);
    expect(f.fetch.mock.calls.some(([url]) => String(url).startsWith('/api/articles/read?'))).toBe(
      false
    );
    f.wrapper.unmount();
  });
});
