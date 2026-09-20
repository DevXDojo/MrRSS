import { mount, flushPromises } from '@vue/test-utils';
import { defineComponent, ref } from 'vue';
import { afterEach, expect, it, vi } from 'vitest';
import { useArticleAdFilter } from './useArticleAdFilter';
import { defaultAdFilterConfig } from '@/types/adFilter';
vi.mock('@/utils/mediaProxy', () => ({ proxyImagesInHtml: (html: string) => html }));
afterEach(() => vi.unstubAllGlobals());
const answer = {
  content: '<p>Editorial</p>',
  decisions: [{ id: 'b2', category: 'advertisement', confidence: 0.99, evidence: 'Buy now' }],
  scanned: 2,
  total: 2,
  truncated: false,
  cached: false,
  guarded: 0,
};
const response = (data: unknown) => ({ ok: true, json: async () => data });
function fixture() {
  const id = ref(1);
  const html = ref('<p>Editorial</p><p>Buy now</p>');
  let result!: ReturnType<typeof useArticleAdFilter>;
  const wrapper = mount(
    defineComponent({
      setup() {
        result = useArticleAdFilter({
          id: () => id.value,
          url: () => 'https://example.org/story',
          content: () => html.value,
          info: () => ({
            originalContent: html.value,
            report: { enabled: true, removed: 0, reasons: {}, guarded: 0 },
          }),
          mediaCache: () => false,
        });
        return () => null;
      },
    })
  );
  return { id, html, result, wrapper };
}
it('does not send article text until requested in review mode, and restores the original', async () => {
  const fetch = vi
    .fn()
    .mockResolvedValueOnce(response({ ...defaultAdFilterConfig(), ai_enabled: true }))
    .mockResolvedValue(response(answer));
  vi.stubGlobal('fetch', fetch);
  const f = fixture();
  await flushPromises();
  expect(fetch).toHaveBeenCalledTimes(1);
  await f.result.analyze();
  expect(f.result.content.value).toContain('Buy now');
  f.result.apply();
  expect(f.result.content.value).not.toContain('Buy now');
  f.result.toggleOriginal();
  expect(f.result.content.value).toContain('Buy now');
  f.wrapper.unmount();
});
it('ignores stale AI results after navigation and cancellation', async () => {
  let finish!: (value: unknown) => void;
  const fetch = vi
    .fn()
    .mockResolvedValueOnce(response({ ...defaultAdFilterConfig(), ai_enabled: true }))
    .mockImplementation(
      () =>
        new Promise((resolve) => {
          finish = resolve;
        })
    );
  vi.stubGlobal('fetch', fetch);
  const f = fixture();
  await flushPromises();
  const pending = f.result.analyze();
  f.id.value = 2;
  finish(response(answer));
  await pending;
  expect(f.result.analysis.value).toBeNull();
  const next = f.result.analyze();
  f.result.cancel();
  finish(response(answer));
  await next;
  expect(f.result.analysis.value).toBeNull();
  expect(f.result.busy.value).toBe(false);
  f.wrapper.unmount();
});
it('automatically applies only opted-in results and skips site exceptions', async () => {
  const fetch = vi
    .fn()
    .mockResolvedValueOnce(
      response({ ...defaultAdFilterConfig(), ai_enabled: true, ai_mode: 'automatic' })
    )
    .mockResolvedValue(response(answer));
  vi.stubGlobal('fetch', fetch);
  const f = fixture();
  await flushPromises();
  expect(fetch).toHaveBeenCalledTimes(2);
  expect(f.result.applied.value).toBe(true);
  window.dispatchEvent(
    new CustomEvent('ad-filter-changed', {
      detail: {
        ...defaultAdFilterConfig(),
        ai_enabled: true,
        ai_mode: 'automatic',
        allowlist: 'example.org',
      },
    })
  );
  await flushPromises();
  expect(fetch).toHaveBeenCalledTimes(2);
  expect(f.result.content.value).toContain('Buy now');
  f.wrapper.unmount();
});
