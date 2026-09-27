import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { createI18n } from 'vue-i18n';
import { useAppStore } from '@/stores/app';
import { setSettingsFromRawData, useSettings } from '@/composables/core/useSettings';
import { useCategoryOrder } from '@/composables/ui/useCategoryOrder';
import SidebarCategory from './SidebarCategory.vue';
import en from '@/i18n/locales/en';
import type { Feed } from '@/types/models';

vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ locale: { value: 'en' }, t: (key: string) => key }),
}));

afterEach(() => vi.unstubAllGlobals());

describe('folder header dragging', () => {
  it('reorders whole folder headers without changing subscription paths', async () => {
    const pinia = createPinia();
    setActivePinia(pinia);
    const store = useAppStore();
    store.feeds = [
      { id: 1, category: 'Alpha' },
      { id: 2, category: 'Beta/Child' },
    ] as Feed[];
    setSettingsFromRawData({ sidebar_sort_mode: 'manual' });
    const fetch = vi.fn(async () => new Response('{}'));
    vi.stubGlobal('fetch', fetch);
    const wrapper = mount(
      {
        components: { SidebarCategory },
        setup() {
          useCategoryOrder();
          return { names: ['Alpha', 'Beta'] };
        },
        template:
          '<div><SidebarCategory v-for="name in names" :key="name" :name="name" :feeds="[]" :is-open="false" :is-active="false" :unread-count="0" :current-feed-id="null" :feed-unread-counts="{}" :is-edit-mode="true" /></div>',
      },
      {
        global: { plugins: [pinia, createI18n({ legacy: false, locale: 'en', messages: { en } })] },
      }
    );
    const headers = wrapper.findAll('.category-header');
    expect(headers[1].attributes('draggable')).toBe('true');
    const dataTransfer = { setData: vi.fn(), effectAllowed: '', dropEffect: '' };
    await headers[1].trigger('dragstart', { dataTransfer });
    await headers[0].trigger('dragover', { dataTransfer, clientY: -1 });
    await headers[0].trigger('drop', { dataTransfer });
    await flushPromises();
    expect(useSettings().settings.value.sidebar_category_order).toBe('["Beta","Alpha"]');
    expect(store.feeds.map((feed) => feed.category)).toEqual(['Alpha', 'Beta/Child']);
    expect(fetch).toHaveBeenCalledOnce();
    wrapper.unmount();
  });
});
