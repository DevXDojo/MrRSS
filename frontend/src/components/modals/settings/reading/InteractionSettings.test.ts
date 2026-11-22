import { shallowMount } from '@vue/test-utils';
import { expect, it, vi } from 'vitest';
import { NumberControl, SettingWithSelect } from '@/components/settings';
import { generateInitialSettings } from '@/composables/core/useSettings.generated';
import InteractionSettings from './InteractionSettings.vue';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));

it('opens custom unread retention from a preset without changing the saved days', async () => {
  const settings = {
    ...generateInitialSettings(),
    auto_mark_read_enabled: true,
    auto_mark_read_days: 7,
  };
  const wrapper = shallowMount(InteractionSettings, {
    props: { settings },
    global: { renderStubDefaultSlot: true },
  });
  const select = wrapper.getComponent(SettingWithSelect);

  await select.setValue('custom');
  expect(wrapper.emitted('update:settings')).toBeUndefined();
  expect(wrapper.getComponent(NumberControl).props('modelValue')).toBe(7);

  await wrapper.getComponent(NumberControl).setValue(3);
  expect(wrapper.emitted('update:settings')?.[0]).toEqual([
    { ...settings, auto_mark_read_days: 3 },
  ]);
  await wrapper.setProps({ settings: { ...settings, auto_mark_read_days: 3 } });
  expect(select.props('modelValue')).toBe('custom');
  expect(wrapper.getComponent(NumberControl).props('modelValue')).toBe(3);

  await select.setValue(3);
  expect(select.props('modelValue')).toBe(3);
  expect(wrapper.findComponent(NumberControl).exists()).toBe(false);
  wrapper.unmount();
});
