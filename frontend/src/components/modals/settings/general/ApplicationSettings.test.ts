import { afterEach, expect, it, vi } from 'vitest';
import { flushPromises, shallowMount } from '@vue/test-utils';
import { Call } from '@wailsio/runtime';
import { ToggleControl } from '@/components/settings';
import { generateInitialSettings } from '@/composables/core/useSettings.generated';
import ApplicationSettings from './ApplicationSettings.vue';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));
vi.mock('@wailsio/runtime', () => ({
  System: { IsDesktop: () => true },
  Call: { ByName: vi.fn(async () => undefined) },
  Events: { On: () => () => {} },
}));
afterEach(() => {
  vi.unstubAllGlobals();
  vi.clearAllMocks();
});
it.each([true, false])(
  'sets ticker enabled=%s through the switch and applies the native state',
  async (enabled) => {
    const fetchMock = vi.fn(async (_url: string, _options?: RequestInit) => Response.json({}));
    vi.stubGlobal('fetch', fetchMock);
    const wrapper = shallowMount(ApplicationSettings, {
      props: { settings: { ...generateInitialSettings(), ticker_enabled: !enabled } },
      global: { renderStubDefaultSlot: true },
    });
    try {
      const toggle = wrapper
        .findAllComponents(ToggleControl)
        .find((component) => component.attributes('aria-label') === 'setting.ticker.enabled')!;
      toggle.vm.$emit('update:modelValue', enabled);
      await flushPromises();
      expect(JSON.parse(String(fetchMock.mock.calls[0][1]?.body))).toEqual({
        ticker_enabled: String(enabled),
      });
      expect(wrapper.emitted('update:settings')?.[0][0]).toMatchObject({ ticker_enabled: enabled });
      expect(Call.ByName).toHaveBeenCalledWith('main.DesktopCapabilities.ApplyTickerSettings');
    } finally {
      wrapper.unmount();
    }
  }
);
