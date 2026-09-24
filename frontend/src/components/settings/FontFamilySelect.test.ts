import { describe, expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import { createI18n } from 'vue-i18n';
import FontFamilySelect from './FontFamilySelect.vue';
import BaseSelect from '@/components/common/BaseSelect.vue';
import en from '@/i18n/locales/en';

vi.mock('@/utils/fontDetector', async (original) => ({
  ...await original<typeof import('@/utils/fontDetector')>(),
  getRecommendedFonts: () => ({ serif: [], sansSerif: [], monospace: [] }),
}));

describe('custom local font selection', () => {
  it('keeps a saved custom font visible and passes new names to settings', () => {
    const wrapper = shallowMount(FontFamilySelect, {
      props: { modelValue: '本机中文字体' },
      global: { plugins: [createI18n({ legacy: false, locale: 'en', messages: { en } })] },
    });
    const select = wrapper.findComponent(BaseSelect);
    expect(select.props('allowCustomInput')).toBe(true);
    expect(JSON.stringify(select.props('options'))).toContain('本机中文字体');
    select.vm.$emit('custom-input', '另一字体');
    expect(wrapper.emitted('update:modelValue')).toEqual([['另一字体']]);
    wrapper.unmount();
  });
});
