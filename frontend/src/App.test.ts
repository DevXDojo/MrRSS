import { describe, it, expect, vi } from 'vitest';
import { readFileSync } from 'node:fs';
import { nextTick } from 'vue';
import { mount, shallowMount, flushPromises } from '@vue/test-utils';
import { createPinia } from 'pinia';
import { createI18n } from 'vue-i18n';
import en from './i18n/locales/en';
import zh from './i18n/locales/zh';
import RuleLogicConnector from './components/modals/rules/RuleLogicConnector.vue';
import ArticleList from './components/article/ArticleList.vue';
import type { Feed } from './types/models';
import App from './App.vue';
import ArticleChatPanel from './components/article/ArticleChatPanel.vue';
import AIFeatureSettings from './components/modals/settings/ai/AIFeatureSettings.vue';
import type { SettingsData } from './types/settings';
import { useArticleActions } from './composables/article/useArticleActions';
import type { Article } from './types/models';
import { setSettingsFromRawData } from './composables/core/useSettings';
import { getRecommendedFonts } from './utils/fontDetector';
import {
  useAppStore,
  createAutoRefreshScheduler,
  getAutoRefreshInterval,
  preserveSelectedArticle,
} from './stores/app';

// Create stub components for complex child components
const createStub = (name: string) => ({
  name,
  template: '<div class="stub-component"><slot /></div>',
});

describe('App', () => {
  it('schedules normal and very long automatic refresh intervals without overflowing', async () => {
    vi.useFakeTimers();
    const refresh = vi.fn();
    const scheduler = createAutoRefreshScheduler(refresh);

    scheduler.start(30);
    await vi.advanceTimersByTimeAsync(30 * 60 * 1000 - 1);
    expect(refresh).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(1);
    expect(refresh).toHaveBeenCalledTimes(1);

    refresh.mockClear();
    scheduler.start(46_080);
    await vi.advanceTimersByTimeAsync(46_080 * 60 * 1000 - 1);
    expect(refresh).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(1);
    expect(refresh).toHaveBeenCalledTimes(1);

    scheduler.stop();
    vi.useRealTimers();
  });

  it('replaces or disables an existing automatic refresh schedule', async () => {
    vi.useFakeTimers();
    const refresh = vi.fn();
    const scheduler = createAutoRefreshScheduler(refresh);

    scheduler.start(30);
    await vi.advanceTimersByTimeAsync(15 * 60 * 1000);
    scheduler.start(60);
    await vi.advanceTimersByTimeAsync(45 * 60 * 1000);
    expect(refresh).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(15 * 60 * 1000);
    expect(refresh).toHaveBeenCalledTimes(1);

    refresh.mockClear();
    scheduler.start(30);
    scheduler.start(0);
    await vi.advanceTimersByTimeAsync(60 * 60 * 1000);
    expect(refresh).not.toHaveBeenCalled();

    scheduler.stop();
    vi.useRealTimers();
  });

  it('disables the frontend timer outside fixed refresh mode', () => {
    expect(getAutoRefreshInterval('fixed', 30)).toBe(30);
    expect(getAutoRefreshInterval('intelligent', 30)).toBe(0);
    expect(getAutoRefreshInterval('never', 30)).toBe(0);
  });

  it('skips overlapping refreshes and catches up only once after sleep', async () => {
    vi.useFakeTimers();
    const refresh = vi.fn();
    let refreshing = true;
    let currentTime = 0;
    const scheduler = createAutoRefreshScheduler(
      refresh,
      () => !refreshing,
      () => currentTime
    );

    scheduler.start(30);
    currentTime = 30 * 60 * 1000;
    await vi.advanceTimersByTimeAsync(30 * 60 * 1000);
    expect(refresh).not.toHaveBeenCalled();

    refreshing = false;
    currentTime = 4 * 30 * 60 * 1000;
    await vi.advanceTimersByTimeAsync(30 * 60 * 1000);
    expect(refresh).toHaveBeenCalledTimes(1);

    await vi.advanceTimersByTimeAsync(1);
    expect(refresh).toHaveBeenCalledTimes(1);

    currentTime = 5 * 30 * 60 * 1000;
    await vi.advanceTimersByTimeAsync(30 * 60 * 1000);
    expect(refresh).toHaveBeenCalledTimes(2);

    scheduler.stop();
    vi.useRealTimers();
  });

  it('preserves the selected article during a background refresh', () => {
    const selected = { id: 75, title: 'Selected article' };
    const fresh = [{ id: 1, title: 'Fresh article' }];

    expect(preserveSelectedArticle(fresh, [selected], 75)).toEqual([fresh[0], selected]);
    expect(preserveSelectedArticle([selected], [selected], 75)).toEqual([selected]);
    expect(preserveSelectedArticle(fresh, [selected], null)).toEqual(fresh);
  });

  it('keeps long toast messages inside narrow viewports', () => {
    const toast = readFileSync('src/components/common/Toast.vue', 'utf8');
    expect(toast).toContain('calc(100vw - 2rem)');
    expect(toast).toContain('overflow-wrap: anywhere');
    expect(toast).toContain('min-w-0 flex-1');
    expect(toast).toContain('shrink-0');
  });

  it('renders and reacts to interface typography settings', async () => {
    setSettingsFromRawData({});
    const pinia = createPinia();
    const i18n = createI18n({
      legacy: false,
      locale: 'en',
      messages: { en },
    });

    // Mock store methods
    const mockFetchFeeds = vi.fn();
    const mockFetchArticles = vi.fn();
    const mockInitTheme = vi.fn();

    const wrapper = mount(App, {
      global: {
        plugins: [pinia, i18n],
        stubs: {
          Sidebar: createStub('Sidebar'),
          ArticleList: createStub('ArticleList'),
          ArticleDetail: createStub('ArticleDetail'),
          ImageGalleryView: createStub('ImageGalleryView'),
          AddFeedModal: createStub('AddFeedModal'),
          EditFeedModal: createStub('EditFeedModal'),
          SettingsModal: createStub('SettingsModal'),
          DiscoverFeedsModal: createStub('DiscoverFeedsModal'),
          UpdateAvailableDialog: createStub('UpdateAvailableDialog'),
          ContextMenu: createStub('ContextMenu'),
          ConfirmDialog: createStub('ConfirmDialog'),
          InputDialog: createStub('InputDialog'),
          MultiSelectDialog: createStub('MultiSelectDialog'),
          Toast: createStub('Toast'),
        },
        mocks: {
          $window: {
            showToast: vi.fn(),
            showConfirm: vi.fn(() => Promise.resolve(true)),
          },
        },
      },
    });

    // Check that the app container is rendered
    expect(wrapper.find('.app-container').exists()).toBe(true);
    expect(document.documentElement.style.getPropertyValue('--ui-font-family')).toContain('Inter');
    expect(document.documentElement.style.getPropertyValue('--ui-font-size')).toBe('16px');
    expect(document.documentElement.style.getPropertyValue('--ui-font-scale')).toBe('1');

    setSettingsFromRawData({
      ui_font_family: 'serif',
      ui_font_size: '8',
    });
    await nextTick();

    expect(document.documentElement.style.getPropertyValue('--ui-font-family')).toContain(
      'Georgia'
    );
    expect(document.documentElement.style.getPropertyValue('--ui-font-size')).toBe('12px');
    expect(document.documentElement.style.getPropertyValue('--ui-font-scale')).toBe('0.75');

    setSettingsFromRawData({
      ui_font_family: 'serif',
      ui_font_size: '24',
    });
    await nextTick();

    expect(document.documentElement.style.getPropertyValue('--ui-font-size')).toBe('20px');
    expect(document.documentElement.style.getPropertyValue('--ui-font-scale')).toBe('1.25');

    setSettingsFromRawData({
      ui_font_family: 'Noto Sans SC',
      ui_font_size: '16',
    });
    await nextTick();

    expect(document.documentElement.style.getPropertyValue('--ui-font-family')).toContain(
      '"Noto Sans SC"'
    );
    expect(document.documentElement.style.getPropertyValue('--ui-font-family')).toContain(
      'system-ui'
    );

    wrapper.unmount();
    expect(document.documentElement.style.getPropertyValue('--ui-font-family')).toBe('');
  });

  it('detects the expanded Chinese font catalog in the expected groups', () => {
    let currentFont = '';
    const context = {
      get font() {
        return currentFont;
      },
      set font(value: string) {
        currentFont = value;
      },
      measureText: () => ({ width: currentFont === '100px sans-serif' ? 100 : 200 }),
    };
    const getContextSpy = vi
      .spyOn(HTMLCanvasElement.prototype, 'getContext')
      .mockReturnValue(context as unknown as CanvasRenderingContext2D);

    const fonts = getRecommendedFonts();

    expect(fonts.sansSerif).toEqual(
      expect.arrayContaining([
        'Noto Sans SC',
        'Source Han Sans CN',
        'Sarasa Gothic SC',
        'Sarasa UI SC',
      ])
    );
    expect(fonts.serif).toEqual(
      expect.arrayContaining([
        'Noto Serif SC',
        'Source Han Serif CN',
        'LXGW WenKai',
        'LXGW WenKai GB',
        'LXGW WenKai Lite',
        'LXGW WenKai Screen',
      ])
    );

    getContextSpy.mockRestore();
  });
});

describe('AI chat panel size', () => {
  it('restores a saved size, fits a smaller viewport, and keeps the preferred size', async () => {
    const ChatPanel = (await import('./components/article/ArticleChatPanel.vue')).default;
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => ({ ok: true, json: async () => [] }))
    );
    vi.stubGlobal('innerWidth', 1000);
    vi.stubGlobal('innerHeight', 800);
    localStorage.setItem('mrrssChatPanelSize', JSON.stringify({ width: 650, height: 680 }));
    const wrapper = mount(ChatPanel, {
      props: {
        article: { id: 1, title: 'Article', url: 'https://example.com' } as any,
        articleContent: 'Article body',
        settings: { ai_chat_enabled: true, ai_chat_profile_id: '', ai_chat_quick_prompts: '' },
      },
      global: {
        plugins: [createI18n({ legacy: false, locale: 'en', messages: { en } })],
        stubs: { Teleport: true },
      },
    });
    try {
      await nextTick();
      const panel = wrapper.get('.chat-panel').element as HTMLElement;
      expect(panel.style.width).toBe('650px');
      expect(panel.style.height).toBe('680px');
      vi.stubGlobal('innerWidth', 500);
      vi.stubGlobal('innerHeight', 400);
      window.dispatchEvent(new Event('resize'));
      expect(panel.style.width).toBe('468px');
      expect(panel.style.height).toBe('344px');
      vi.stubGlobal('innerWidth', 1000);
      vi.stubGlobal('innerHeight', 800);
      window.dispatchEvent(new Event('resize'));
      expect(panel.style.width).toBe('650px');
      expect(JSON.parse(localStorage.getItem('mrrssChatPanelSize')!).width).toBe(650);
      vi.spyOn(panel, 'getBoundingClientRect').mockReturnValue({ width: 650, height: 680 } as DOMRect);
      await wrapper.get('.cursor-nw-resize').trigger('mousedown', { clientX: 100, clientY: 100 });
      document.dispatchEvent(new MouseEvent('mousemove', { clientX: 50, clientY: 60 }));
      document.dispatchEvent(new MouseEvent('mouseup'));
      expect(JSON.parse(localStorage.getItem('mrrssChatPanelSize')!)).toEqual({
        width: 700,
        height: 720,
      });
    } finally {
      wrapper.unmount();
      localStorage.removeItem('mrrssChatPanelSize');
      vi.unstubAllGlobals();
    }
  });
});


describe('Rule logic localization', () => {
  it('localizes AND/OR labels while preserving emitted rule operators', async () => {
    const i18n = createI18n({ legacy: false, locale: 'zh', messages: { en, zh } });
    const wrapper = mount(RuleLogicConnector, {
      props: { logic: 'and' },
      global: { plugins: [i18n] },
    });
    const buttons = wrapper.findAll('button');
    expect(buttons.map((button) => button.text())).toEqual(['且', '或']);
    await buttons[1].trigger('click');
    await buttons[0].trigger('click');
    expect(wrapper.emitted('update')).toEqual([['or'], ['and']]);

    i18n.global.locale.value = 'en';
    await nextTick();
    expect(buttons.map((button) => button.text())).toEqual(['AND', 'OR']);
    wrapper.unmount();
  });
});

describe('Chat response preferences', () => {
  it('edits and clears the shared preference without changing model selection', async () => {
    const settings = {
      ai_chat_enabled: true,
      ai_chat_profile_id: '7',
      ai_chat_quick_prompts: '[]',
      ai_chat_response_preferences: '',
    } as SettingsData;
    const wrapper = mount(AIFeatureSettings, {
      props: { settings },
      global: {
        plugins: [createI18n({ legacy: false, locale: 'en', messages: { en } })],
        stubs: { AIProfileSelector: true, AIChatQuickPromptsSettings: true },
      },
    });
    const input = wrapper.get('textarea');
    await input.setValue('用中文回答，保持简洁。');
    const updated = wrapper.emitted('update:settings')?.[0]?.[0] as SettingsData;
    expect(updated.ai_chat_response_preferences).toBe('用中文回答，保持简洁。');
    expect(updated.ai_chat_profile_id).toBe('7');
    await wrapper.setProps({ settings: updated });
    await input.setValue('');
    const cleared = wrapper.emitted('update:settings')?.[1]?.[0] as SettingsData;
    expect(cleared.ai_chat_response_preferences).toBe('');
    wrapper.unmount();
  });
});


describe('Selected source titles', () => {
  it('omits only the all filter and localizes uncategorized selections', async () => {
    const pinia = createPinia();
    const i18n = createI18n({ legacy: false, locale: 'en', messages: { en, zh } });
    const wrapper = shallowMount(ArticleList, { global: { plugins: [pinia, i18n] } });
    const store = useAppStore(pinia);
    store.feeds = [{ id: 7, title: 'Example Feed' } as Feed];
    store.currentFilter = 'all';
    await nextTick();
    const title = () => wrapper.get('h3').text();
    expect(title()).toBe('All Articles');

    store.tempSelection = { feedId: 7, category: null };
    await nextTick();
    expect(title()).toBe('Example Feed');
    for (const [filter, label] of [
      ['unread', 'Unread Articles'],
      ['favorites', 'Favorites'],
      ['readLater', 'Read Later'],
    ] as const) {
      store.currentFilter = filter;
      await nextTick();
      expect(title()).toBe(`Example Feed - ${label}`);
    }
    store.currentFilter = 'all';
    store.tempSelection = { feedId: null, category: 'Technology/News' };
    await nextTick();
    expect(title()).toBe('Technology/News');

    store.tempSelection = { feedId: null, category: 'uncategorized' };
    await nextTick();
    expect(title()).toBe('Uncategorized');
    i18n.global.locale.value = 'zh';
    await nextTick();
    expect(title()).toBe('未分类');
    store.currentFilter = 'favorites';
    await nextTick();
    expect(title()).toBe(`未分类 - ${zh.sidebar.activity.favorites}`);
    wrapper.unmount();
  });
});
describe('Article context menu read status', () => {
  it('uses neutral read icons with distinct shapes while preserving status actions and colors', () => {
    const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } });
    let actions: ReturnType<typeof useArticleActions>;
    const wrapper = mount({
      setup() {
        actions = useArticleActions(i18n.global.t, { value: 'rendered' });
        return {};
      },
      template: '<div />',
    }, { global: { plugins: [createPinia(), i18n] } });
    const openMenu = vi.fn();
    window.addEventListener('open-context-menu', openMenu);
    try {
      for (const isRead of [false, true]) {
        const article: Article = {
          id: 1, feed_id: 1, title: 'Article', url: 'https://example.com/article',
          published_at: '2026-01-01T00:00:00Z', is_read: isRead,
          is_favorite: true, is_read_later: true, is_hidden: false,
        };
        actions!.showArticleContextMenu(new MouseEvent('contextmenu'), article);
        const event = openMenu.mock.calls.at(-1)![0] as CustomEvent;
        expect(event.detail.data).toBe(article);
        expect(event.detail.items[0]).toMatchObject({
          action: 'toggleRead', icon: 'ph-circle', iconColor: 'text-text-secondary',
          iconWeight: isRead ? 'regular' : 'fill',
          label: i18n.global.t(isRead ? 'article.action.markAsUnread' : 'article.action.markAsRead'),
        });
        expect(event.detail.items).toEqual(expect.arrayContaining([
          expect.objectContaining({ action: 'toggleFavorite', iconColor: 'text-yellow-500' }),
          expect.objectContaining({ action: 'toggleReadLater', iconColor: 'text-blue-500' }),
        ]));
      }
    } finally {
      window.removeEventListener('open-context-menu', openMenu);
      wrapper.unmount();
    }
  });
});

describe('Chat generation cancellation', () => {
  function deferred<T>() {
    let resolve!: (value: T) => void;
    const promise = new Promise<T>((done) => {
      resolve = done;
    });
    return { promise, resolve };
  }

  function mountChat() {
    return mount(ArticleChatPanel, {
      props: {
        article: { id: 12, title: 'Article', url: 'https://example.com/article' } as Article,
        articleContent: 'Article content',
        settings: { ai_chat_enabled: true, ai_chat_profile_id: '', ai_chat_quick_prompts: '[]' },
      },
      global: {
        plugins: [createI18n({ legacy: false, locale: 'en', messages: { en } })],
        stubs: { teleport: true },
      },
    });
  }

  it.each([false, true])('reuses stopped creation (%s)', async (resolvedBeforeRetry) => {
    const creation = deferred<Response>();
    const generated = deferred<Response>();
    const toast = vi.fn();
    const previousToast = window.showToast;
    window.showToast = toast;
    const fetchMock = vi.fn((url: string, options?: RequestInit) => {
      if (url === '/api/ai/chat/session/create') return creation.promise;
      if (url === '/api/ai-chat') return generated.promise;
      return Promise.resolve(
        new Response(JSON.stringify(url === '/api/ai-chat/cancel' ? { success: true } : []))
      );
    });
    vi.stubGlobal('fetch', fetchMock);
    const wrapper = mountChat();
    try {
      await flushPromises();
      await wrapper.get('input').setValue('first');
      await wrapper.get('[data-testid="chat-send-message"]').trigger('click');
      await flushPromises();
      await wrapper.get('[data-testid="chat-stop-generation"]').trigger('click');
      expect((wrapper.get('input').element as HTMLInputElement).value).toBe('first');
      expect(wrapper.text()).not.toContain('first');
      if (resolvedBeforeRetry) {
        creation.resolve(new Response(JSON.stringify({ id: 55, article_id: 12, title: 'first' })));
        await flushPromises();
        expect((wrapper.get('input').element as HTMLInputElement).value).toBe('first');
        expect(fetchMock.mock.calls.filter(([url]) => url === '/api/ai-chat')).toHaveLength(0);
      }
      await wrapper.get('input').setValue('second');
      await wrapper.get('[data-testid="chat-send-message"]').trigger('click');
      await flushPromises();
      expect(
        fetchMock.mock.calls.filter(([url]) => url === '/api/ai/chat/session/create')
      ).toHaveLength(1);
      if (!resolvedBeforeRetry) {
        creation.resolve(new Response(JSON.stringify({ id: 55, article_id: 12, title: 'first' })));
      }
      await flushPromises();
      const sends = fetchMock.mock.calls.filter(([url]) => url === '/api/ai-chat');
      expect(sends).toHaveLength(1);
      const request = JSON.parse(sends[0][1]?.body as string);
      expect(request.session_id).toBe(55);
      expect(request.request_id).toBeTruthy();
      expect(request.messages).toHaveLength(1);
      expect(request.messages.at(-1).content).toBe('second');
      await wrapper.get('[data-testid="chat-stop-generation"]').trigger('click');
      await flushPromises();
      expect(sends[0][1]?.signal?.aborted).toBe(true);
      generated.resolve(new Response(JSON.stringify({ response: 'late answer', session_id: 55 })));
      await flushPromises();
      expect(wrapper.text()).not.toContain('late answer');
      expect(toast).not.toHaveBeenCalled();
    } finally {
      if (wrapper.exists()) wrapper.unmount();
      window.showToast = previousToast;
      vi.unstubAllGlobals();
    }
  });

  it('preserves the first question when creating a session fails and retries it once', async () => {
    let createCount = 0;
    const fetchMock = vi.fn((url: string, options?: RequestInit) => {
      if (url === '/api/ai/chat/session/create') {
        createCount++;
        return Promise.resolve(
          createCount === 1
            ? new Response('', { status: 503 })
            : new Response(JSON.stringify({ id: 55, article_id: 12, title: 'retry me' }))
        );
      }
      if (url === '/api/ai-chat') {
        return Promise.resolve(
          new Response(JSON.stringify({ response: 'answer', session_id: 55 }))
        );
      }
      return Promise.resolve(new Response('[]'));
    });
    const previousToast = window.showToast;
    const toast = vi.fn();
    window.showToast = toast;
    vi.stubGlobal('fetch', fetchMock);
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    const wrapper = mountChat();
    try {
      await flushPromises();
      await wrapper.get('input').setValue('retry me');
      await wrapper.get('[data-testid="chat-send-message"]').trigger('click');
      await flushPromises();
      expect((wrapper.get('input').element as HTMLInputElement).value).toBe('retry me');
      expect(wrapper.text()).not.toContain('retry me');
      expect(fetchMock.mock.calls.filter(([url]) => url === '/api/ai-chat')).toHaveLength(0);
      expect(toast).toHaveBeenCalledTimes(1);
      await wrapper.get('[data-testid="chat-send-message"]').trigger('click');
      await flushPromises();
      const sends = fetchMock.mock.calls.filter(([url]) => url === '/api/ai-chat');
      expect(createCount).toBe(2);
      expect(sends).toHaveLength(1);
      const request = JSON.parse(sends[0][1]?.body as string);
      expect(request.messages).toHaveLength(1);
      expect(request.messages[0].content).toBe('retry me');
      expect(request.is_first_message).toBe(true);
      expect((wrapper.get('input').element as HTMLInputElement).value).toBe('');
    } finally {
      wrapper.unmount();
      errorSpy.mockRestore();
      window.showToast = previousToast;
      vi.unstubAllGlobals();
    }
  });

  it.each([
    { action: 'close', hasAssistant: false },
    { action: 'unmount', hasAssistant: false },
    { action: 'close', hasAssistant: true },
  ])('isolates retries: $action, assistant $hasAssistant', async ({ action, hasAssistant }) => {
    const first = deferred<Response>();
    const second = deferred<Response>();
    let sendCount = 0;
    const fetchMock = vi.fn((url: string, options?: RequestInit) => {
      if (url === '/api/ai/chat/session/create')
        return Promise.resolve(
          new Response(JSON.stringify({ id: 55, article_id: 12, title: 'first' }))
        );
      if (url === '/api/ai-chat') return ++sendCount === 1 ? first.promise : second.promise;
      if (url.startsWith('/api/ai/chat/messages?')) {
        const history = [{ role: 'user', content: 'stored first question' }];
        if (hasAssistant) history.push({ role: 'assistant', content: 'stored first answer' });
        return Promise.resolve(new Response(JSON.stringify(history)));
      }
      return Promise.resolve(
        new Response(JSON.stringify(url === '/api/ai-chat/cancel' ? { success: true } : []))
      );
    });
    const toast = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    const previousToast = window.showToast;
    window.showToast = toast;
    const wrapper = mountChat();
    try {
      await flushPromises();
      await wrapper.get('input').setValue('first');
      await wrapper.get('[data-testid="chat-send-message"]').trigger('click');
      await flushPromises();
      await wrapper.get('[data-testid="chat-stop-generation"]').trigger('click');
      await flushPromises();
      await wrapper.get('input').setValue('second');
      await wrapper.get('[data-testid="chat-send-message"]').trigger('click');
      await flushPromises();
      first.resolve(new Response(JSON.stringify({ response: 'old result', session_id: 55 })));
      await flushPromises();
      expect(wrapper.find('[data-testid="chat-stop-generation"]').exists()).toBe(true);
      expect(wrapper.text()).not.toContain('old result');
      if (action === 'close') {
        await wrapper.get('[data-testid="chat-close"]').trigger('click');
        expect(wrapper.emitted('close')).toHaveLength(1);
      } else {
        wrapper.unmount();
      }
      await flushPromises();
      const sends = fetchMock.mock.calls.filter(([url]) => url === '/api/ai-chat');
      const cancels = fetchMock.mock.calls.filter(([url]) => url === '/api/ai-chat/cancel');
      const nextRequest = JSON.parse(sends[1][1]?.body as string);
      expect(nextRequest.is_first_message).toBe(!hasAssistant);
      expect(nextRequest.article_content).toBe('Article content');
      expect(cancels).toHaveLength(2);
      expect(JSON.parse(cancels[0][1]?.body as string).request_id).not.toBe(
        JSON.parse(cancels[1][1]?.body as string).request_id
      );
      expect(sends[1][1]?.signal?.aborted).toBe(true);
      second.resolve(new Response(JSON.stringify({ response: 'closed result', session_id: 55 })));
      await flushPromises();
      expect(wrapper.text()).not.toContain('closed result');
      expect(toast).not.toHaveBeenCalled();
    } finally {
      if (wrapper.exists()) wrapper.unmount();
      window.showToast = previousToast;
      vi.unstubAllGlobals();
    }
  });
});
