import { mount } from '@vue/test-utils';
import { defineComponent } from 'vue';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useFeedPreview } from './useFeedPreview';

afterEach(() => vi.unstubAllGlobals());

describe('feed preview requests', () => {
  it('ignores late results after switching URLs and cancels on reset', async () => {
    const replies: Array<(response: Response) => void> = [];
    const fetchMock = vi.fn(
      (_url: RequestInfo | URL, _options?: RequestInit) =>
        new Promise<Response>((resolve) => replies.push(resolve))
    );
    vi.stubGlobal('fetch', fetchMock);
    let preview!: ReturnType<typeof useFeedPreview>;
    const wrapper = mount(
      defineComponent({
        setup() {
          preview = useFeedPreview();
          return () => null;
        },
      })
    );
    const first = preview.load({
      url: 'https://first.example',
      proxy_enabled: false,
      proxy_url: '',
    });
    const second = preview.load({
      url: 'https://second.example',
      proxy_enabled: false,
      proxy_url: '',
    });
    expect(fetchMock.mock.calls[0][1]?.signal?.aborted).toBe(true);
    replies[1](Response.json({ title: 'Second', description: '', total: 0, articles: [] }));
    await second;
    replies[0](Response.json({ title: 'First', description: '', total: 0, articles: [] }));
    await first;
    expect(preview.preview.value?.title).toBe('Second');
    const pending = preview.load({
      url: 'https://third.example',
      proxy_enabled: true,
      proxy_url: '',
    });
    wrapper.unmount();
    expect(fetchMock.mock.calls[2][1]?.signal?.aborted).toBe(true);
    replies[2](Response.json({ title: 'Third', articles: [] }));
    await pending;
    expect(preview.preview.value).toBeNull();
    expect(preview.failed.value).toBe(false);
  });

  it('shows failed requests instead of an empty successful preview', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response('{}', { status: 502 }))
    );
    let preview!: ReturnType<typeof useFeedPreview>;
    const wrapper = mount(
      defineComponent({
        setup() {
          preview = useFeedPreview();
          return () => null;
        },
      })
    );
    await preview.load({ url: 'https://failed.example', proxy_enabled: false, proxy_url: '' });
    expect(preview.failed.value).toBe(true);
    expect(preview.preview.value).toBeNull();
    expect(preview.isLoading.value).toBe(false);
    wrapper.unmount();
  });
});
