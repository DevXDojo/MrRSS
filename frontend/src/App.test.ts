import { describe, it, expect, vi } from 'vitest';
import { readFileSync } from 'node:fs';
import { nextTick } from 'vue';
import { mount } from '@vue/test-utils';
import { createPinia } from 'pinia';
import { createI18n } from 'vue-i18n';
import en from './i18n/locales/en';
import App from './App.vue';
import { setSettingsFromRawData } from './composables/core/useSettings';
import { getRecommendedFonts } from './utils/fontDetector';
import {
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
    expect(toast).toContain('calc(100vw-2rem)');
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
