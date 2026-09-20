import { mount, flushPromises } from '@vue/test-utils';
import { afterEach, expect, it, vi } from 'vitest';
import { createI18n } from 'vue-i18n';
import AdFilterSettings from './AdFilterSettings.vue';
import { enAdFilter, zhAdFilter } from '@/i18n/locales/adFilter';
import { defaultAdFilterConfig } from '@/types/adFilter';
import type { SettingsData } from '@/types/settings';
afterEach(() => vi.unstubAllGlobals());
it.each(['en', 'zh'])(
  'renders real %s messages, validates draft state and saves',
  async (locale) => {
    const initial = defaultAdFilterConfig();
    const fetch = vi
      .fn()
      .mockImplementation(async (_url: string, init?: RequestInit) => ({
        ok: true,
        json: async () =>
          init?.method === 'PUT' ? { ...JSON.parse(String(init.body)), revision: 1 } : initial,
      }));
    vi.stubGlobal('fetch', fetch);
    window.showToast = vi.fn();
    const wrapper = mount(AdFilterSettings, {
      props: { settings: { ai_ad_filter_profile_id: '0' } as SettingsData },
      global: {
        plugins: [
          createI18n({
            legacy: false,
            locale,
            messages: {
              en: { setting: { adFilter: enAdFilter } },
              zh: { setting: { adFilter: zhAdFilter } },
            },
          }),
        ],
        stubs: { AIProfileSelector: true },
      },
    });
    await flushPromises();
    expect(wrapper.text()).toContain(locale === 'en' ? 'Ad filtering' : '广告过滤');
    await wrapper.find('textarea').setValue('example.org');
    expect(wrapper.emitted('editing')?.at(-1)).toEqual([true]);
    const save = wrapper
      .findAll('button')
      .find((b) => b.text() === (locale === 'en' ? 'Save filtering options' : '保存过滤选项'))!;
    await save.trigger('click');
    await flushPromises();
    expect(wrapper.emitted('editing')?.at(-1)).toEqual([false]);
    expect(fetch).toHaveBeenCalledTimes(2);
    wrapper.unmount();
  }
);
it('keeps the draft and shows an actionable save conflict', async () => {
  vi.stubGlobal(
    'fetch',
    vi
      .fn()
      .mockResolvedValueOnce({ ok: true, json: async () => defaultAdFilterConfig() })
      .mockResolvedValue({ ok: false, json: async () => ({ error: { message: 'conflict' } }) })
  );
  const wrapper = mount(AdFilterSettings, {
    props: { settings: {} as SettingsData },
    global: {
      plugins: [
        createI18n({
          legacy: false,
          locale: 'en',
          messages: { en: { setting: { adFilter: enAdFilter } } },
        }),
      ],
      stubs: { AIProfileSelector: true },
    },
  });
  await flushPromises();
  await wrapper.find('textarea').setValue('example.org');
  await wrapper
    .findAll('button')
    .find((b) => b.text() === 'Save filtering options')!
    .trigger('click');
  await flushPromises();
  expect(wrapper.get('[role="alert"]').text()).toContain('another window');
  expect(wrapper.find('textarea').element.value).toBe('example.org');
  expect(wrapper.emitted('editing')?.at(-1)).toEqual([true]);
  wrapper.unmount();
});
