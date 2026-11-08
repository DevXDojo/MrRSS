import { afterEach, expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import { ToggleControl } from '@/components/settings';
import { generateInitialSettings } from '@/composables/core/useSettings.generated';
import ApplicationSettings from './ApplicationSettings.vue';
import TickerSettings from './TickerSettings.vue';
import type { SettingsData } from '@/types/settings';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));
vi.mock('@wailsio/runtime', () => ({
  System: { IsDesktop: () => true },
  Call: { ByName: vi.fn() },
  Events: { On: () => () => {} },
}));
afterEach(() => vi.clearAllMocks());

it('expands ticker subsettings with the switch and preserves them when turned off', async () => {
  const original = { ...generateInitialSettings(), ticker_enabled: false, ticker_speed: 84 };
  const wrapper = shallowMount(ApplicationSettings, {
    props: { settings: original },
    global: { renderStubDefaultSlot: true },
  });
  try {
    expect(wrapper.findComponent(TickerSettings).exists()).toBe(false);
    const tickerGroup = wrapper.get('#desktop-ticker-settings').element;
    expect(tickerGroup.parentElement?.lastElementChild).toBe(tickerGroup);
    expect(wrapper.find('button').exists()).toBe(false);
    const toggle = wrapper
      .findAllComponents(ToggleControl)
      .find((component) => component.attributes('aria-label') === 'setting.ticker.enabled')!;
    toggle.vm.$emit('update:modelValue', true);
    const enabled = wrapper.emitted('update:settings')![0][0] as SettingsData;
    await wrapper.setProps({ settings: enabled });
    expect(wrapper.findComponent(TickerSettings).exists()).toBe(true);
    toggle.vm.$emit('update:modelValue', false);
    const disabled = wrapper.emitted('update:settings')![1][0] as SettingsData;
    await wrapper.setProps({ settings: disabled });
    expect(wrapper.findComponent(TickerSettings).exists()).toBe(false);
    expect(disabled.ticker_speed).toBe(84);
  } finally {
    wrapper.unmount();
  }
});

it('passes subsetting edits through the standard settings update flow', async () => {
  const settings = { ...generateInitialSettings(), ticker_enabled: true };
  const wrapper = shallowMount(ApplicationSettings, {
    props: { settings },
    global: { renderStubDefaultSlot: true },
  });
  try {
    wrapper
      .getComponent(TickerSettings)
      .vm.$emit('update:settings', { ...settings, ticker_font_size: 22 });
    expect(wrapper.emitted('update:settings')?.[0][0]).toMatchObject({
      ticker_font_size: 22,
      ticker_enabled: true,
    });
  } finally {
    wrapper.unmount();
  }
});
