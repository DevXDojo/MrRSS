import { afterEach, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { createI18n } from 'vue-i18n';
import ArticleSummary from './ArticleSummary.vue';

afterEach(() => vi.useRealTimers());
it.each(['local', 'ai'])(
  'shows the translated %s summary beside the original',
  (summaryProvider) => {
    vi.useFakeTimers();
    const wrapper = mount(ArticleSummary, {
      props: {
        summaryProvider,
        translationEnabled: true,
        isLoadingSummary: false,
        summaryResult: {
          summary: 'Original English summary',
          sentence_count: 1,
          is_too_short: false,
        },
        translatedSummary: { text: '摘要译文', html: '<p>摘要译文</p>' },
      },
      global: {
        plugins: [
          createI18n({
            legacy: false,
            locale: 'en',
            messages: { en: {} },
            missingWarn: false,
            fallbackWarn: false,
          }),
        ],
      },
    });
    try {
      expect(wrapper.text()).toContain('摘要译文');
      expect(wrapper.text()).toContain('Original English summary');
    } finally {
      wrapper.unmount();
    }
  }
);
