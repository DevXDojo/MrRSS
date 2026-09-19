import { mount, flushPromises } from '@vue/test-utils';
import { afterEach, expect, it, vi } from 'vitest';
import XPathPicker from './XPathPicker.vue';
import type { XPathPreviewNode } from '@/utils/xpathPicker';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));
afterEach(() => vi.unstubAllGlobals());

it('guides a title click through list confirmation and applies inferred fields', async () => {
  const rows: XPathPreviewNode[] = [1, 2, 3].map((index) => {
    const path = `/html[1]/body[1]/ul[1]/li[${index}]`;
    return {
      path,
      tag: 'li',
      group: '/html[1]/body[1]/ul[1]/li',
      children: [
        {
          path: path + '/a[1]',
          tag: 'a',
          link: `https://example.com/${index}`,
          text: `Article ${index}`,
        },
      ],
    };
  });
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        html: '<html><head></head><body></body></html>',
        base_url: 'https://example.com/',
        children: rows,
      }),
    })
  );
  const wrapper = mount(XPathPicker, {
    props: { url: 'https://example.com/' },
    attachTo: document.body,
    global: {
      stubs: { Teleport: true, BaseModal: { template: '<div><slot/><slot name="footer"/></div>' } },
    },
  });
  try {
    await flushPromises();
    const frame = wrapper.find('iframe').element;
    const token = frame.srcdoc.match(/nonce="([^"]+)"/)![1];
    const pick = (candidateToken: string, row = 1) =>
      window.dispatchEvent(
        new MessageEvent('message', {
          source: frame.contentWindow,
          data: { type: 'mrrss-xpath-pick', token: candidateToken, path: rows[row].path + '/a[1]' },
        })
      );
    pick('wrong-token');
    await flushPromises();
    expect(wrapper.text()).not.toContain('modal.feed.picker.confirmGroup');
    pick(token);
    await flushPromises();
    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'modal.feed.picker.calibrate')!
      .trigger('click');
    pick(token, 2);
    await flushPromises();
    const confirm = wrapper
      .findAll('button')
      .find((button) => button.text() === 'modal.feed.picker.confirmGroup')!;
    expect(confirm.exists()).toBe(true);
    await confirm.trigger('click');
    await flushPromises();
    expect(wrapper.text()).toContain('Article 1');
    expect(wrapper.text()).toContain('https://example.com/3');
    await wrapper
      .findAll('button')
      .find((button) => button.text().startsWith('modal.feed.picker.titleField'))!
      .trigger('click');
    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'modal.feed.picker.calibrateField')!
      .trigger('click');
    pick(token); // Same article cannot provide an independent example.
    await flushPromises();
    expect(wrapper.text()).toContain('modal.feed.picker.incompatibleExamples');
    pick(token, 0);
    await flushPromises();
    const apply = wrapper
      .findAll('button')
      .find((button) => button.text() === 'modal.feed.picker.apply')!;
    expect(apply.attributes('disabled')).toBeUndefined();
    await apply.trigger('click');
    expect(wrapper.emitted('apply')?.[0]).toEqual([
      {
        item: rows[1].group,
        title: './a[1]',
        uri: './a[1]/@href',
        timestamp: '',
        content: '',
        thumbnail: '',
      },
    ]);
  } finally {
    wrapper.unmount();
  }
});
