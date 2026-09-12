import { describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { createI18n } from 'vue-i18n';
import { ref } from 'vue';
import ArticleTableRow from './ArticleTableRow.vue';
import type { Article } from '@/types/models';

vi.mock('@/composables/core/useSettings', () => ({
  useSettings: () => ({ settings: ref({ translation_only_mode: true }) }),
}));
vi.mock('@/composables/article/useArticleHoverRead', () => ({
  useArticleHoverRead: () => ({ enter: vi.fn(), leave: vi.fn() }),
}));
vi.mock('@/composables/article/useArticleDateFormat', () => ({
  useArticleDateFormat: () => ({
    formatArticleDate: () => '2026-09-12',
    formatArticleDateTime: () => '2026-09-12 09:30',
  }),
}));

describe('article table rows', () => {
  it('renders only selected columns, escapes feed content, and keeps article actions available', async () => {
    const article: Article = {
      id: 5,
      feed_id: 2,
      title: '<img src=x onerror=alert(1)>',
      translated_title: 'Translated title',
      author: 'Author',
      feed_title: 'Feed',
      url: 'https://example.com/article',
      published_at: '2026-09-12T09:30:00Z',
      is_read: false,
      is_favorite: false,
      is_read_later: false,
      is_hidden: false,
    };
    const wrapper = mount(ArticleTableRow, {
      props: { article, columns: ['title', 'author'], isActive: true },
      global: { plugins: [createI18n({ legacy: false, locale: 'en', messages: { en: {} } })] },
    });
    expect(wrapper.findAll('td')).toHaveLength(2);
    expect(wrapper.get('button').text()).toBe('Translated title');
    expect(wrapper.get('button').attributes('aria-current')).toBe('true');
    expect(wrapper.find('img').exists()).toBe(false);
    await wrapper.get('button').trigger('click');
    expect(wrapper.emitted('click')).toHaveLength(1);
    await wrapper.trigger('contextmenu');
    expect(wrapper.emitted('contextmenu')).toHaveLength(1);
    await wrapper.setProps({ columns: ['title', 'date'], isActive: false });
    expect(wrapper.find('time').text()).toBe('2026-09-12');
    expect(wrapper.text()).not.toContain('Author');
    expect(wrapper.get('button').attributes('aria-current')).toBeUndefined();
    wrapper.unmount();
  });
});
