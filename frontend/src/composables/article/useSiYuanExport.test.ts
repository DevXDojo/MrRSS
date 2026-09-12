import { mount, flushPromises } from '@vue/test-utils';
import { defineComponent } from 'vue';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useSiYuanExport } from './useSiYuanExport';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));
afterEach(() => {
  vi.unstubAllGlobals();
});

describe('SiYuan export', () => {
  it('prevents duplicate submissions and reports HTTP failures', async () => {
    let resolve!: (response: Response) => void;
    const fetchMock = vi.fn(
      (_input: RequestInfo | URL, _init?: RequestInit) =>
        new Promise<Response>((done) => {
          resolve = done;
        })
    );
    vi.stubGlobal('fetch', fetchMock);
    window.showToast = vi.fn();
    let exporter!: ReturnType<typeof useSiYuanExport>;
    const wrapper = mount(
      defineComponent({
        setup() {
          exporter = useSiYuanExport();
          return () => null;
        },
      })
    );
    const request = exporter.exportToSiYuan(42);
    await exporter.exportToSiYuan(42);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    resolve(new Response('{}', { status: 400 }));
    await request;
    expect(window.showToast).toHaveBeenLastCalledWith(
      'setting.plugins.siyuan.configurationError',
      'error'
    );
    expect(exporter.isExporting.value).toBe(false);
    wrapper.unmount();
  });

  it('aborts on unmount and suppresses a late success toast', async () => {
    let resolve!: (response: Response) => void;
    const fetchMock = vi.fn(
      (_input: RequestInfo | URL, _init?: RequestInit) =>
        new Promise<Response>((done) => {
          resolve = done;
        })
    );
    vi.stubGlobal('fetch', fetchMock);
    window.showToast = vi.fn();
    let exporter!: ReturnType<typeof useSiYuanExport>;
    const wrapper = mount(
      defineComponent({
        setup() {
          exporter = useSiYuanExport();
          return () => null;
        },
      })
    );
    const request = exporter.exportToSiYuan(42);
    const options = fetchMock.mock.calls[0][1];
    wrapper.unmount();
    expect(options?.signal?.aborted).toBe(true);
    resolve(new Response('{}', { status: 200 }));
    await request;
    await flushPromises();
    expect(window.showToast).toHaveBeenCalledTimes(1);
  });
});
