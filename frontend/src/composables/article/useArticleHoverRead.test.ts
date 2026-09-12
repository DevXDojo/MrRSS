import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { ref } from 'vue';
import { useArticleHoverRead } from './useArticleHoverRead';
import type { Article } from '@/types/models';

vi.mock('@/composables/core/useSettings', () => ({
  useSettings: () => ({ settings: ref({ hover_mark_as_read: true }) }),
}));
vi.mock('vue-i18n', () => ({ useI18n: () => ({ locale: ref('en') }) }));

beforeEach(() => {
  setActivePinia(createPinia());
  vi.useFakeTimers();
});
afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

function setup(article: Article) {
  let hover!: ReturnType<typeof useArticleHoverRead>;
  const onRead = vi.fn();
  const wrapper = mount({
    setup() {
      hover = useArticleHoverRead(() => article, onRead);
      return {};
    },
    template: '<div />',
  });
  return { hover, onRead, wrapper };
}

describe('shared article hover reading', () => {
  it('preserves read-later articles', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    const { hover, wrapper } = setup({ id: 1, is_read: false, is_read_later: true } as Article);
    hover.enter();
    await vi.advanceTimersByTimeAsync(500);
    expect(fetchMock).not.toHaveBeenCalled();
    wrapper.unmount();
  });
  it('does not mark a failed request as read', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{}', { status: 500 })));
    const { hover, onRead, wrapper } = setup({
      id: 1,
      is_read: false,
      is_read_later: false,
    } as Article);
    hover.enter();
    await vi.advanceTimersByTimeAsync(500);
    expect(onRead).not.toHaveBeenCalled();
    wrapper.unmount();
  });
  it('cancels delayed reading when a row disappears', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    const { hover, wrapper } = setup({ id: 1, is_read: false, is_read_later: false } as Article);
    hover.enter();
    wrapper.unmount();
    await vi.advanceTimersByTimeAsync(500);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
