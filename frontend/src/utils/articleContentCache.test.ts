import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  clearArticleContentCache,
  loadArticleContent,
  getArticleContentCacheSize,
  getCachedArticleContent,
  invalidateArticleContent,
} from './articleContentCache';

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('article content cache', () => {
  beforeEach(() => {
    clearArticleContentCache();
    vi.unstubAllGlobals();
  });

  it('serves a repeated selection from the cache without another request', async () => {
    const fetchMock = vi.fn(() => Promise.resolve(jsonResponse({ content: 'body-1' })));
    vi.stubGlobal('fetch', fetchMock);

    const first = await loadArticleContent(1);
    const second = await loadArticleContent(1);

    expect(first.content).toBe('body-1');
    expect(second.content).toBe('body-1');
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenCalledWith('/api/articles/content?id=1', { signal: undefined });
  });

  it('maps the response fields and remembers them', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() =>
        Promise.resolve(jsonResponse({ content: 'c', feed_url: 'https://feed', cached: true }))
      )
    );

    const entry = await loadArticleContent(7);
    expect(entry).toEqual({ content: 'c', feedUrl: 'https://feed', cached: true });
    expect(getCachedArticleContent(7)).toEqual(entry);
  });

  it('evicts the least recently used article once the cache is full', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn((url: string) => Promise.resolve(jsonResponse({ content: url })))
    );

    for (let id = 1; id <= 8; id += 1) await loadArticleContent(id);
    // refresh article 1 so article 2 becomes the least recently used
    await loadArticleContent(1);
    await loadArticleContent(9);

    expect(getArticleContentCacheSize()).toBe(8);
    expect(getCachedArticleContent(1)).toBeDefined();
    expect(getCachedArticleContent(2)).toBeUndefined();
    expect(getCachedArticleContent(9)).toBeDefined();
  });

  it('refetches after invalidation', async () => {
    const fetchMock = vi.fn(() => Promise.resolve(jsonResponse({ content: 'body' })));
    vi.stubGlobal('fetch', fetchMock);

    await loadArticleContent(3);
    invalidateArticleContent(3);
    await loadArticleContent(3);

    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('shares one request between concurrent selections', async () => {
    let resolveFetch!: (response: Response) => void;
    const fetchMock = vi.fn(
      () =>
        new Promise<Response>((resolve) => {
          resolveFetch = resolve;
        })
    );
    vi.stubGlobal('fetch', fetchMock);

    const first = loadArticleContent(5);
    const second = loadArticleContent(5);
    resolveFetch(jsonResponse({ content: 'shared' }));

    await expect(first).resolves.toMatchObject({ content: 'shared' });
    await expect(second).resolves.toMatchObject({ content: 'shared' });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it('does not cache a failed request', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.resolve(new Response('nope', { status: 500 })))
    );

    await expect(loadArticleContent(11)).rejects.toThrow('500');
    expect(getCachedArticleContent(11)).toBeUndefined();

    // the failed attempt must not block a later one
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.resolve(jsonResponse({ content: 'ok' })))
    );
    await expect(loadArticleContent(11)).resolves.toMatchObject({ content: 'ok' });
  });

  it('does not cache an empty body so the reader can retry', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse({ content: '', cached: false }))
      .mockResolvedValueOnce(jsonResponse({ content: 'second try', cached: false }));
    vi.stubGlobal('fetch', fetchMock);

    await expect(loadArticleContent(21)).resolves.toMatchObject({ content: '' });
    expect(getCachedArticleContent(21)).toBeUndefined();

    await expect(loadArticleContent(21)).resolves.toMatchObject({ content: 'second try' });
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(getCachedArticleContent(21)).toMatchObject({ content: 'second try' });
  });

  it('does not cache a whitespace-only body', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.resolve(jsonResponse({ content: '   ' })))
    );

    await loadArticleContent(23);
    expect(getCachedArticleContent(23)).toBeUndefined();
  });

  it('does not cache an aborted request', async () => {
    const controller = new AbortController();
    vi.stubGlobal(
      'fetch',
      vi.fn((_url: string, init?: RequestInit) => {
        return new Promise<Response>((_resolve, reject) => {
          init?.signal?.addEventListener('abort', () => reject(new Error('AbortError')));
          controller.abort();
        });
      })
    );

    await expect(loadArticleContent(13, controller.signal)).rejects.toThrow('AbortError');
    expect(getCachedArticleContent(13)).toBeUndefined();
  });
});
