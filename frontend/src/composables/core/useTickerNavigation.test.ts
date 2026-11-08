import { afterEach, expect, it, vi } from 'vitest';
import { defineComponent } from 'vue';
import { flushPromises, mount } from '@vue/test-utils';
import { System } from '@wailsio/runtime';
import { useTickerNavigation } from './useTickerNavigation';

const native = vi.hoisted(() => ({
  listener: undefined as undefined | ((event: { data: unknown }) => void),
  settingsListener: undefined as undefined | ((event: { data: unknown }) => void),
  dispose: vi.fn(),
}));
vi.mock('@wailsio/runtime', () => ({
  System: { IsDesktop: vi.fn(() => true) },
  Events: {
    On: (_name: string, callback: (event: { data: unknown }) => void) => {
      if (_name === 'ticker:open-settings') native.settingsListener = callback;
      else native.listener = callback;
      return native.dispose;
    },
  },
}));
afterEach(() => {
  vi.unstubAllGlobals();
  vi.clearAllMocks();
});
it('does not install desktop navigation in server mode', () => {
  vi.mocked(System.IsDesktop).mockReturnValueOnce(false);
  native.listener = undefined;
  const store = state();
  const wrapper = mount(
    defineComponent({
      setup() {
        useTickerNavigation(store);
      },
      template: '<div />',
    })
  );
  expect(native.listener).toBeUndefined();
  wrapper.unmount();
});
const item = { id: 8, feed_id: 3, title: 'Selected headline' };
const state = () => ({
  currentArticleId: null as number | null,
  currentFeedId: null as number | null,
  setArticleNavigationContext: vi.fn(),
  selectFeedInArticleList: vi.fn(),
});
it('loads the headline outside the main page before selecting its source', async () => {
  const store = state();
  const fetchMock = vi.fn(async () => Response.json([item]));
  vi.stubGlobal('fetch', fetchMock);
  const wrapper = mount(
    defineComponent({
      setup() {
        useTickerNavigation(store);
      },
      template: '<div />',
    })
  );
  try {
    native.listener?.({ data: { article_id: 8, feed_id: 3 } });
    native.listener?.({ data: { article_id: '8', feed_id: 3 } });
    native.listener?.({ data: { article_id: -1, feed_id: 3 } });
    await flushPromises();
    expect(fetchMock).toHaveBeenCalledOnce();
    expect(store.setArticleNavigationContext).toHaveBeenCalledExactlyOnceWith([item]);
    expect(store.selectFeedInArticleList).toHaveBeenCalledExactlyOnceWith(3, 8);
  } finally {
    wrapper.unmount();
  }
  expect(native.dispose).toHaveBeenCalledTimes(2);
});
it('does not replace a newer user selection after a slow headline lookup', async () => {
  let finish!: (response: Response) => void;
  vi.stubGlobal(
    'fetch',
    vi.fn(
      () =>
        new Promise<Response>((resolve) => {
          finish = resolve;
        })
    )
  );
  const store = state();
  const wrapper = mount(
    defineComponent({
      setup() {
        useTickerNavigation(store);
      },
      template: '<div />',
    })
  );
  try {
    native.listener?.({ data: { article_id: 8, feed_id: 3 } });
    store.currentArticleId = 99;
    finish(Response.json([item]));
    await flushPromises();
    expect(store.selectFeedInArticleList).not.toHaveBeenCalled();
  } finally {
    wrapper.unmount();
  }
});

it('opens the main settings through its navigation callback without selecting an article', async () => {
  const store = state();
  const openSettings = vi.fn();
  const wrapper = mount(
    defineComponent({
      setup() {
        useTickerNavigation(store, openSettings);
      },
      template: '<div />',
    })
  );
  try {
    native.settingsListener?.({ data: undefined });
    expect(openSettings).toHaveBeenCalledOnce();
    expect(store.selectFeedInArticleList).not.toHaveBeenCalled();
  } finally {
    wrapper.unmount();
  }
});
