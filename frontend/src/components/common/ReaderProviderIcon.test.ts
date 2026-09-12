import { mount } from '@vue/test-utils';
import { reactive } from 'vue';
import { describe, expect, it, vi } from 'vitest';
import ReaderProviderIcon from './ReaderProviderIcon.vue';
const state = vi.hoisted(() => ({ theme: 'light' }));
vi.mock('@/stores/app', () => ({ useAppStore: () => reactive(state) }));
describe('reader icons', () => {
  it('uses the matching Miniflux icon when the application theme changes', async () => {
    const wrapper = mount(ReaderProviderIcon, { props: { provider: 'miniflux' } });
    expect(wrapper.attributes('src')).toBe('/assets/plugin_icons/miniflux.svg');
    reactive(state).theme = 'dark';
    await wrapper.vm.$nextTick();
    expect(wrapper.attributes('src')).toBe('/assets/plugin_icons/miniflux-dark.svg');
    await wrapper.setProps({ provider: 'freshrss' });
    expect(wrapper.attributes('src')).toBe('/assets/plugin_icons/freshrss.svg');
    wrapper.unmount();
  });
});
