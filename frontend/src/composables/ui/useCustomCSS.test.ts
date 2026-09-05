import { afterEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, ref, nextTick } from 'vue';
import { mount, flushPromises } from '@vue/test-utils';
import { useCustomCSS } from './useCustomCSS';

const response = (css: string) => ({ ok: true, text: async () => css });
const styles = () => document.head.querySelectorAll('style[data-custom-css="application"]');
const wrappers: Array<ReturnType<typeof mount>> = [];
function fixture() {
  const file = ref('theme.css');
  const wrapper = mount(
    defineComponent({
      setup() {
        useCustomCSS(() => file.value);
        return () => null;
      },
    })
  );
  wrappers.push(wrapper);
  return { file, wrapper };
}
afterEach(() => {
  wrappers.splice(0).forEach((wrapper) => wrapper.unmount());
  vi.unstubAllGlobals();
});

describe('application custom CSS', () => {
  it('loads without a reader and replaces the same file without duplicate styles', async () => {
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(response(':root { --accent-color: purple; }'))
      .mockResolvedValueOnce(response(':root { --accent-color: orange; }'));
    vi.stubGlobal('fetch', fetch);
    fixture();
    await flushPromises();
    expect(styles()).toHaveLength(1);
    expect(styles()[0].textContent).toContain('purple');
    window.dispatchEvent(new Event('custom-css-changed'));
    await flushPromises();
    expect(styles()).toHaveLength(1);
    expect(styles()[0].textContent).toContain('orange');
  });
  it('does not let a slow old load replace a newer theme', async () => {
    let finish!: (value: ReturnType<typeof response>) => void;
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockImplementationOnce(
          () =>
            new Promise((resolve) => {
              finish = resolve;
            })
        )
        .mockResolvedValueOnce(response('body { color: blue; }'))
    );
    const f = fixture();
    f.file.value = 'new.css';
    await flushPromises();
    finish(response('body { color: red; }'));
    await flushPromises();
    expect(styles()).toHaveLength(1);
    expect(styles()[0].textContent).toContain('blue');
  });
  it('removes deleted themes and prevents in-flight requests from restoring them', async () => {
    let finish!: (value: ReturnType<typeof response>) => void;
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(response('body { color: red; }'))
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            finish = resolve;
          })
      );
    vi.stubGlobal('fetch', fetch);
    const f = fixture();
    await flushPromises();
    window.dispatchEvent(new Event('custom-css-changed'));
    f.file.value = '';
    await nextTick();
    expect(styles()).toHaveLength(0);
    finish(response('body { color: blue; }'));
    await flushPromises();
    expect(styles()).toHaveLength(0);
  });
  it('cancels loading and removes event listeners when the app unmounts', async () => {
    let signal!: AbortSignal;
    const fetch = vi.fn().mockImplementation((_url, options) => {
      signal = options.signal;
      return new Promise(() => {});
    });
    vi.stubGlobal('fetch', fetch);
    const f = fixture();
    f.wrapper.unmount();
    expect(signal.aborted).toBe(true);
    window.dispatchEvent(new Event('custom-css-changed'));
    await flushPromises();
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(styles()).toHaveLength(0);
  });
});
