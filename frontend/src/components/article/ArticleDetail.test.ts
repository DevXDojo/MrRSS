import { describe, expect, it, vi } from 'vitest';
import { flushPromises, shallowMount } from '@vue/test-utils';
import { defineComponent, ref } from 'vue';
import ArticleDetail from './ArticleDetail.vue';
import ArticleToolbar from './ArticleToolbar.vue';

vi.mock('@/composables/article/useArticleDetail', () => ({
  useArticleDetail: () => ({
    article: ref({ id: 1, url: 'https://example.org/article' }),
    showContent: ref(false), articleContent: ref('<p>Feed content</p>'),
    isLoadingContent: ref(false), hasPreviousArticle: ref(false), hasNextArticle: ref(false),
    imageViewerSrc: ref(null), t: (key: string) => key,
  }),
}));

describe('reading mode entry', () => {
  it('replaces the webpage with the reader before requesting clean text', async () => {
    const enterReadingMode = vi.fn();
    const wrapper = shallowMount(ArticleDetail, {
      global: {
        stubs: {
          ArticleContent: defineComponent({
            setup(_, { expose }) {
              expose({ enterReadingMode, isFetchingFullArticle: false });
              return () => null;
            },
          }),
        },
      },
    });
    expect(wrapper.find('iframe').exists()).toBe(true);
    wrapper.findComponent(ArticleToolbar).vm.$emit('readingMode');
    await flushPromises();
    expect(wrapper.find('iframe').exists()).toBe(false);
    expect(enterReadingMode).toHaveBeenCalledTimes(1);
    wrapper.unmount();
  });
});
