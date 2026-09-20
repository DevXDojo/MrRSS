import { mount } from '@vue/test-utils';
import { createI18n } from 'vue-i18n';
import { expect, it, vi } from 'vitest';
import XPathConfig from './XPathConfig.vue';
import XPathPicker from './XPathPicker.vue';
vi.mock('@/utils/browser', () => ({ openInBrowser: vi.fn() }));
it.each(['custom-id', 'old-link', ''])(
  'preserves manual metadata when applying a visual selection (UID %s)',
  async (uid) => {
    const wrapper = mount(XPathConfig, {
      props: {
        mode: 'edit',
        url: 'https://example.org',
        xpathType: 'HTML+XPath',
        xpathItem: 'old',
        xpathItemTitle: 'old-title',
        xpathItemContent: '',
        xpathItemUri: 'old-link',
        xpathItemAuthor: './/author',
        xpathItemTimestamp: '',
        xpathItemTimeFormat: '2006-01-02',
        xpathItemThumbnail: '',
        xpathItemCategories: './/category',
        xpathItemUid: uid,
      },
      global: {
        plugins: [
          createI18n({
            legacy: false,
            locale: 'en',
            missingWarn: false,
            fallbackWarn: false,
            messages: { en: {} },
          }),
        ],
        stubs: { XPathPicker: true, BaseSelect: true },
      },
    });
    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'modal.feed.picker.title')!
      .trigger('click');
    wrapper
      .findComponent(XPathPicker)
      .vm.$emit('apply', {
        item: 'new',
        title: 'new-title',
        uri: 'new-link',
        content: '',
        timestamp: '',
        thumbnail: '',
      });
    expect(wrapper.emitted('update:xpath-item-author')).toBeUndefined();
    expect(wrapper.emitted('update:xpath-item-categories')).toBeUndefined();
    expect(wrapper.emitted('update:xpath-item-time-format')).toBeUndefined();
    expect(wrapper.emitted('update:xpath-item-uid')).toEqual(
      uid === 'custom-id' ? undefined : [['new-link']]
    );
    wrapper.unmount();
  }
);
