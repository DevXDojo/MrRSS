import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { computed, nextTick, ref } from 'vue';
import { createPinia, setActivePinia } from 'pinia';
import { useAppStore } from '@/stores/app';
import { useArticleListTransition } from './useArticleListTransition';
import type { Article } from '@/types/models';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ locale: { value: 'en' } }) }));
beforeEach(() => {
  setActivePinia(createPinia());
  vi.useFakeTimers();
});
afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

function mountTransition(articles = ref<Article[]>([{ id: 1 } as Article]), loading = ref(false)) {
  let transition!: ReturnType<typeof useArticleListTransition>;
  const wrapper = mount({
    setup() {
      transition = useArticleListTransition(articles, loading);
      return transition;
    },
    template:
      '<div :inert="showingPrevious || undefined"><span v-for="article in displayedArticles" :key="article.id">{{ article.id }}</span></div>',
  });
  return { articles, loading, transition, wrapper };
}

describe('article navigation presentation', () => {
  it('releases a retained list after the current request fails', async () => {
    const store = useAppStore();
    store.articles = [{ id: 1 } as Article];
    const { transition, wrapper } = mountTransition(
      computed(() => store.articles),
      computed(() => store.isLoading)
    );
    let finish!: (response: Response) => void;
    vi.stubGlobal(
      'fetch',
      vi.fn(
        () =>
          new Promise<Response>((resolve) => {
            finish = resolve;
          })
      )
    );
    const request = store.fetchArticles();
    await nextTick();
    expect(wrapper.text()).toBe('1');
    finish(new Response('{}', { status: 500 }));
    await request;
    await nextTick();
    expect(transition.showingPrevious.value).toBe(false);
    expect(wrapper.text()).toBe('');
    expect(store.navigableArticles).toEqual([]);
    wrapper.unmount();
  });
  it('retains rows through a clear/loading batch, then replaces them with the actual empty result', async () => {
    const { articles, loading, transition, wrapper } = mountTransition();
    // Matches fetchArticles: clear the data before setting its loading flag.
    articles.value = [];
    loading.value = true;
    await nextTick();
    expect(wrapper.text()).toBe('1');
    expect(wrapper.attributes('inert')).toBeDefined();
    expect(articles.value).toEqual([]);
    await vi.advanceTimersByTimeAsync(100);
    expect(transition.showLoadingIndicator.value).toBe(false);
    await vi.advanceTimersByTimeAsync(50);
    expect(transition.showLoadingIndicator.value).toBe(true);
    loading.value = false;
    await nextTick();
    expect(wrapper.text()).toBe('');
    expect(wrapper.attributes('inert')).toBeUndefined();
    expect(transition.showLoadingIndicator.value).toBe(false);
    wrapper.unmount();
  });

  it('does not flash a spinner for fast results or insert snapshots into live keyboard navigation', async () => {
    const store = useAppStore();
    store.articles = [{ id: 1 } as Article];
    const { transition, wrapper } = mountTransition(
      computed(() => store.articles),
      computed(() => store.isLoading)
    );
    let finishFirst!: (response: Response) => void;
    let finishSecond!: (response: Response) => void;
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockImplementationOnce(
          () =>
            new Promise<Response>((resolve) => {
              finishFirst = resolve;
            })
        )
        .mockImplementationOnce(
          () =>
            new Promise<Response>((resolve) => {
              finishSecond = resolve;
            })
        )
    );
    store.currentFeedId = 2;
    const first = store.fetchArticles();
    await nextTick();
    store.currentFeedId = 3;
    const second = store.fetchArticles();
    await nextTick();
    expect(wrapper.text()).toBe('1');
    expect(store.navigableArticles).toEqual([]);
    finishSecond(new Response(JSON.stringify([{ id: 3 }])));
    await second;
    await nextTick();
    expect(wrapper.text()).toBe('3');
    finishFirst(new Response(JSON.stringify([{ id: 2 }])));
    await first;
    await nextTick();
    expect(wrapper.text()).toBe('3');
    expect(store.navigableArticles.map((article) => article.id)).toEqual([3]);
    await vi.advanceTimersByTimeAsync(200);
    expect(transition.showLoadingIndicator.value).toBe(false);
    wrapper.unmount();
  });
});
