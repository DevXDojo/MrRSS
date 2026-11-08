import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, ref } from 'vue';
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils';
import { Call } from '@wailsio/runtime';
import { generateInitialSettings } from '@/composables/core/useSettings.generated';
import { useTickerPreferences } from './useTickerPreferences';

const nativeListeners = vi.hoisted(() => new Map<string, () => void>());
vi.mock('@wailsio/runtime', () => ({
  Call: { ByName: vi.fn(async () => undefined) },
  Events: {
    On: (name: string, callback: () => void) => {
      nativeListeners.set(name, callback);
      return () => nativeListeners.delete(name);
    },
  },
}));

describe('independent ticker preference writes', () => {
  let wrapper: VueWrapper;
  let api: ReturnType<typeof useTickerPreferences>;
  const settings = ref(generateInitialSettings());
  beforeEach(() => {
    vi.useFakeTimers();
    vi.clearAllMocks();
    settings.value = generateInitialSettings();
  });
  afterEach(() => {
    wrapper?.unmount();
    vi.useRealTimers();
    vi.unstubAllGlobals();
    nativeListeners.clear();
  });
  function start() {
    wrapper = mount(
      defineComponent({
        setup() {
          api = useTickerPreferences(settings);
        },
        template: '<div />',
      })
    );
  }
  it('serializes changes and merges other windows without overwriting pending input', async () => {
    let finishFirst!: (response: Response) => void;
    const writes: Array<Record<string, string>> = [];
    const fetchMock = vi.fn(async (_url: string, options?: RequestInit): Promise<Response> => {
      if (options?.method !== 'POST')
        return Response.json({
          ticker_speed: '20',
          ticker_font_size: '12',
          ticker_always_on_top: 'false',
        });
      writes.push(JSON.parse(String(options.body)));
      if (writes.length === 1)
        return new Promise((resolve) => {
          finishFirst = resolve;
        });
      return Response.json({});
    });
    vi.stubGlobal('fetch', fetchMock);
    start();
    settings.value.ticker_speed = 70;
    const first = api.saveNow();
    settings.value.ticker_font_size = 22;
    nativeListeners.get('ticker:settings-changed')?.();
    await flushPromises();
    expect(settings.value.ticker_speed).toBe(70);
    expect(settings.value.ticker_font_size).toBe(22);
    expect(settings.value.ticker_always_on_top).toBe(false);
    finishFirst(Response.json({}));
    await first;
    expect(writes).toEqual([{ ticker_speed: '70' }, { ticker_font_size: '22' }]);
    expect(api.status.value).toBe('saved');
    expect(Call.ByName).toHaveBeenCalledWith('main.DesktopCapabilities.ApplyTickerSettings');
  });
  it('keeps failed changes for an explicit retry', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response('', { status: 500 }))
      .mockResolvedValueOnce(Response.json({}));
    vi.stubGlobal('fetch', fetchMock);
    start();
    settings.value.ticker_enabled = true;
    expect(await api.saveNow()).toBe(false);
    expect(api.status.value).toBe('error');
    expect(Call.ByName).not.toHaveBeenCalled();
    expect(await api.saveNow()).toBe(true);
    expect(JSON.parse(fetchMock.mock.calls[1][1].body)).toEqual({ ticker_enabled: 'true' });
  });
});
