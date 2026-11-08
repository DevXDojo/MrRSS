import { afterEach, expect, it, vi } from 'vitest';
import { shallowMount, flushPromises } from '@vue/test-utils';
import { generateInitialSettings } from '@/composables/core/useSettings.generated';
import { NumberControl, SelectControl } from '@/components/settings';
import BaseSelect from '@/components/common/BaseSelect.vue';
import TickerSettings from './TickerSettings.vue';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));
vi.mock('@wailsio/runtime', () => ({ Call: { ByName: vi.fn(async () => []) } }));
afterEach(() => {
  vi.unstubAllGlobals();
  vi.clearAllMocks();
});

it('loads searchable local subscriptions and emits bounded settings edits', async () => {
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => Response.json([{ id: 2, title: 'Local feed' }]))
  );
  const settings = { ...generateInitialSettings(), ticker_enabled: true };
  const wrapper = shallowMount(TickerSettings, {
    props: { settings },
    global: { renderStubDefaultSlot: true, stubs: { SubSettingItem: false } },
  });
  try {
    await flushPromises();
    const sources = wrapper.getComponent(BaseSelect);
    expect(sources.props('searchable')).toBe(true);
    expect(sources.props('options')).toContainEqual({ value: 2, label: 'Local feed' });
    sources.vm.$emit('update:modelValue', 2);
    expect(wrapper.emitted('update:settings')?.[0][0]).toMatchObject({
      ticker_feed_id: 2,
      ticker_enabled: true,
    });
    wrapper
      .findAllComponents(NumberControl)
      .find((control) => control.props('min') === 10)!
      .vm.$emit('update:modelValue', 900);
    expect(wrapper.emitted('update:settings')?.[1][0]).toMatchObject({ ticker_speed: 200 });
    wrapper
      .findAllComponents(SelectControl)
      .find((control) => control.props('options').some((option) => option.value === 'midnight'))!
      .vm.$emit('update:modelValue', 'midnight');
    expect(wrapper.emitted('update:settings')?.[2][0]).toMatchObject({
      ticker_text_color: '#f8fafc',
      ticker_background_color: '#1e293b',
      ticker_background_color_end: '#0f172a',
    });
  } finally {
    wrapper.unmount();
  }
});

it('allows retrying a subscription loading error without discarding preferences', async () => {
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce(new Response('', { status: 500 }))
    .mockResolvedValueOnce(Response.json([{ id: 7, title: 'Recovered' }]));
  vi.stubGlobal('fetch', fetchMock);
  const wrapper = shallowMount(TickerSettings, {
    props: { settings: generateInitialSettings() },
    global: { renderStubDefaultSlot: true, stubs: { SubSettingItem: false } },
  });
  try {
    await flushPromises();
    expect(wrapper.text()).toContain('setting.ticker.loadFailed');
    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'setting.ticker.retry')!
      .trigger('click');
    await flushPromises();
    expect(wrapper.getComponent(BaseSelect).props('options')).toContainEqual({
      value: 7,
      label: 'Recovered',
    });
    expect(wrapper.emitted('update:settings')).toBeUndefined();
  } finally {
    wrapper.unmount();
  }
});
