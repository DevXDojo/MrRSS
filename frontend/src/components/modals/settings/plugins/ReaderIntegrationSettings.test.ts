import { mount } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { settingsDefaults } from '@/config/defaults';
import { InputControl } from '@/components/settings';
import type { SettingsData } from '@/types/settings';
import ReaderIntegrationSettings from './ReaderIntegrationSettings.vue';
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ theme: 'light', startFreshRSSStatusPolling: vi.fn() }),
}));
afterEach(() => vi.unstubAllGlobals());
describe('independent reader settings', () => {
  it('edits Miniflux credentials without overwriting the configured FreshRSS account', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Response.json({ last_sync_time: null }))
    );
    const settings = {
      ...settingsDefaults,
      freshrss_enabled: true,
      miniflux_enabled: true,
      freshrss_server_url: 'https://fresh.example',
      miniflux_server_url: 'https://mini.example',
    } as SettingsData;
    const wrapper = mount(ReaderIntegrationSettings, { props: { settings, provider: 'miniflux' } });
    const url = wrapper.findAllComponents(InputControl)[0];
    expect(url.props('modelValue')).toBe('https://mini.example');
    url.vm.$emit('update:modelValue', 'https://new-mini.example');
    const updated = wrapper.emitted('update:settings')?.[0][0] as SettingsData;
    expect(updated.miniflux_server_url).toBe('https://new-mini.example');
    expect(updated.freshrss_server_url).toBe('https://fresh.example');
    expect(updated.freshrss_enabled).toBe(true);
    wrapper.unmount();
  });
});
