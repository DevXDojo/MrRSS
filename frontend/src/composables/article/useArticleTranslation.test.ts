import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Article } from '@/types/models';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));

describe('automatic title translation during refresh', () => {
  let translation: ReturnType<typeof import('./useArticleTranslation').useArticleTranslation>;
  let callback: IntersectionObserverCallback;
  let observe: ReturnType<typeof vi.fn>;
  let root: HTMLElement;
  let row: HTMLElement;
  beforeEach(async () => {
    vi.resetModules();
    vi.useFakeTimers();
    window.showToast = vi.fn();
    observe = vi.fn();
    vi.stubGlobal(
      'IntersectionObserver',
      class {
        observe = observe;
        unobserve = vi.fn();
        disconnect = vi.fn();
        constructor(cb: IntersectionObserverCallback) {
          callback = cb;
        }
      }
    );
    translation = (await import('./useArticleTranslation')).useArticleTranslation();
    translation.handleTranslationSettingsChange(true, 'zh', 'auto');
    root = document.createElement('div');
    row = document.createElement('div');
    row.dataset.articleId = '1';
    root.appendChild(row);
  });
  afterEach(() => {
    translation.cleanup();
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });
  const article = () => ({ id: 1, title: 'Title' }) as Article;
  function enter() {
    callback(
      [{ target: row, isIntersecting: true } as IntersectionObserverEntry],
      {} as IntersectionObserver
    );
  }

  it('keeps skipped results across refreshed article objects without rebuilding observers', async () => {
    const fetchMock = vi.fn(
      async () => new Response(JSON.stringify({ translated_title: 'Title', skipped: true }))
    );
    vi.stubGlobal('fetch', fetchMock);
    translation.setupIntersectionObserver(root, [article()]);
    await vi.advanceTimersByTimeAsync(0);
    enter();
    await vi.advanceTimersByTimeAsync(0);
    const refreshed = article();
    translation.setupIntersectionObserver(root, [refreshed]);
    await vi.advanceTimersByTimeAsync(0);
    expect(observe).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(refreshed.translated_title).toBe('Title');
  });

  it('backs off failed visible titles and resumes automatically after the delay', async () => {
    const fetchMock = vi.fn(async () => new Response('{}', { status: 500 }));
    vi.stubGlobal('fetch', fetchMock);
    translation.setupIntersectionObserver(root, [article()]);
    await vi.advanceTimersByTimeAsync(0);
    enter();
    await vi.advanceTimersByTimeAsync(0);
    for (let i = 0; i < 10; i++) {
      translation.setupIntersectionObserver(root, [article()]);
      await vi.advanceTimersByTimeAsync(500);
    }
    expect(fetchMock).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(25000);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    translation.cleanup();
    await vi.advanceTimersByTimeAsync(60000);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('allows explicit retry and treats a changed title as new work', async () => {
    const fetchMock = vi.fn(async () => new Response('{}', { status: 500 }));
    vi.stubGlobal('fetch', fetchMock);
    await translation.translateArticle(article(), true);
    await translation.translateArticle(article());
    await translation.translateArticle({ id: 1, title: 'Updated title' } as Article, true);
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });

  it('keeps a rate-limited title pending across refreshes and fills the latest article object', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(null, { status: 429, headers: { 'Retry-After': '2' } }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ translated_title: '标题' })));
    vi.stubGlobal('fetch', fetchMock);
    translation.setupIntersectionObserver(root, [article()]);
    await vi.advanceTimersByTimeAsync(0);
    enter();
    await vi.advanceTimersByTimeAsync(0);
    const refreshed = article();
    translation.setupIntersectionObserver(root, [refreshed]);
    await vi.advanceTimersByTimeAsync(1000);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(translation.translatingArticles.value.has(1)).toBe(true);
    expect(window.showToast).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(1000);
    expect(refreshed.translated_title).toBe('标题');
    expect(translation.translatingArticles.value.size).toBe(0);
  });

  it('releases a waiting title offscreen and translates it when it becomes visible again', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(null, { status: 429, headers: { 'Retry-After': '2' } }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ translated_title: '标题' })));
    vi.stubGlobal('fetch', fetchMock);
    const item = article();
    translation.setupIntersectionObserver(root, [item]);
    await vi.advanceTimersByTimeAsync(0);
    enter();
    await vi.advanceTimersByTimeAsync(0);
    callback(
      [{ target: row, isIntersecting: false } as IntersectionObserverEntry],
      {} as IntersectionObserver
    );
    await vi.advanceTimersByTimeAsync(1000);
    expect(translation.translatingArticles.value.size).toBe(0);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    enter();
    await vi.advanceTimersByTimeAsync(1000);
    expect(item.translated_title).toBe('标题');
  });
});
