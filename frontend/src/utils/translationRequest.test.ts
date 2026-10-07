import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

describe('shared translation request budget', () => {
  beforeEach(() => {
    vi.resetModules();
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });
  it('retains limited titles and paragraphs, then resumes without a request burst', async () => {
    const { requestTranslation, translationQueue } = await import('./translationRequest');
    const pending: Array<(res: Response) => void> = [];
    const fetchMock = vi.fn(() => new Promise<Response>((resolve) => pending.push(resolve)));
    vi.stubGlobal('fetch', fetchMock);
    const requests = Array.from({ length: 20 }, (_, i) =>
      requestTranslation(i % 2 ? '/titles' : '/paragraphs', {})
    );
    expect(fetchMock).toHaveBeenCalledTimes(3);
    pending[0](new Response(null, { status: 429, headers: { 'Retry-After': '120' } }));
    pending[1](new Response('{}'));
    pending[2](new Response('{}'));
    fetchMock.mockImplementation(async () => new Response('{}'));
    await vi.advanceTimersByTimeAsync(0);
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(translationQueue.value.pending).toBe(18);
    await vi.advanceTimersByTimeAsync(119999);
    expect(fetchMock).toHaveBeenCalledTimes(3);
    await vi.advanceTimersByTimeAsync(1);
    expect(fetchMock).toHaveBeenCalledTimes(4);
    await vi.advanceTimersByTimeAsync(17000);
    const results = await Promise.all(requests);
    expect(results.every((res) => res.ok)).toBe(true);
    expect(fetchMock).toHaveBeenCalledTimes(21);
    expect(translationQueue.value.pending).toBe(0);
    expect(vi.getTimerCount()).toBe(0);
  });
  it('drops obsolete queued work before making a network request', async () => {
    const { requestTranslation } = await import('./translationRequest');
    const pending: Array<(res: Response) => void> = [];
    const fetchMock = vi.fn(() => new Promise<Response>((resolve) => pending.push(resolve)));
    vi.stubGlobal('fetch', fetchMock);
    const running = Array.from({ length: 3 }, () => requestTranslation('/titles', {}));
    let current = true;
    const stale = requestTranslation('/titles', {}, () => current);
    current = false;
    pending.forEach((resolve) => resolve(new Response('{}')));
    await Promise.all(running);
    expect((await stale).status).toBe(204);
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });
  it('allows translation with a newly selected provider during the previous cooldown', async () => {
    const { requestTranslation, resetTranslationCooldown } = await import('./translationRequest');
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(null, { status: 429 }))
      .mockResolvedValue(new Response('{}'));
    vi.stubGlobal('fetch', fetchMock);
    const oldRequest = requestTranslation('/titles', {});
    await vi.advanceTimersByTimeAsync(0);
    resetTranslationCooldown();
    expect((await oldRequest).status).toBe(204);
    expect((await requestTranslation('/titles', {})).ok).toBe(true);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('prioritizes the current article and waits for one successful recovery probe', async () => {
    const { requestTranslation } = await import('./translationRequest');
    const pending: Array<(res: Response) => void> = [];
    const fetchMock = vi.fn(() => new Promise<Response>((resolve) => pending.push(resolve)));
    vi.stubGlobal('fetch', fetchMock);
    const titles = Array.from({ length: 6 }, () =>
      requestTranslation('/titles', {}, () => true, 'background')
    );
    pending[0](new Response(null, { status: 429, headers: { 'Retry-After': '2' } }));
    pending[1](new Response('{}'));
    pending[2](new Response('{}'));
    await vi.advanceTimersByTimeAsync(0);
    const content = requestTranslation('/paragraphs', {});
    await vi.advanceTimersByTimeAsync(2000);
    expect(fetchMock).toHaveBeenLastCalledWith('/paragraphs', expect.any(Object));
    await vi.advanceTimersByTimeAsync(5000);
    expect(fetchMock).toHaveBeenCalledTimes(4);
    fetchMock.mockImplementation(async () => new Response('{}'));
    pending[3](new Response('{}'));
    expect((await content).ok).toBe(true);
    await vi.advanceTimersByTimeAsync(5000);
    expect((await Promise.all(titles)).every((res) => res.ok)).toBe(true);
  });

  it('cancels obsolete work during cooldown instead of waiting for the provider', async () => {
    const { requestTranslation, translationQueue } = await import('./translationRequest');
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(null, { status: 429 }))
    );
    let current = true;
    const request = requestTranslation('/paragraphs', {}, () => current);
    await vi.advanceTimersByTimeAsync(0);
    current = false;
    await vi.advanceTimersByTimeAsync(1000);
    expect((await request).status).toBe(204);
    expect(translationQueue.value.pending).toBe(0);
    expect(vi.getTimerCount()).toBe(0);
  });

  it('aborts in-flight requests after leaving the article', async () => {
    const { requestTranslation } = await import('./translationRequest');
    const fetchMock = vi.fn(
      (_url: string, options: RequestInit) =>
        new Promise<Response>((_resolve, reject) => {
          options.signal?.addEventListener('abort', () =>
            reject(new DOMException('', 'AbortError'))
          );
        })
    );
    vi.stubGlobal('fetch', fetchMock);
    let current = true;
    const request = requestTranslation('/paragraphs', {}, () => current);
    current = false;
    await vi.advanceTimersByTimeAsync(1000);
    expect((await request).status).toBe(204);
    expect(fetchMock.mock.calls[0][1].signal?.aborted).toBe(true);
    expect(vi.getTimerCount()).toBe(0);
  });

  it('backs off repeated limits without discarding the paragraph', async () => {
    const { requestTranslation } = await import('./translationRequest');
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(null, { status: 429 }))
      .mockResolvedValueOnce(new Response(null, { status: 429 }))
      .mockResolvedValue(new Response('{}'));
    vi.stubGlobal('fetch', fetchMock);
    const request = requestTranslation('/paragraphs', {});
    await vi.advanceTimersByTimeAsync(60000);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    await vi.advanceTimersByTimeAsync(119999);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    await vi.advanceTimersByTimeAsync(1);
    expect((await request).ok).toBe(true);
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });

  it('honors HTTP-date retry headers and leaves permanent errors to the caller', async () => {
    const { requestTranslation } = await import('./translationRequest');
    vi.setSystemTime(new Date('2026-10-07T00:00:00Z'));
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(null, {
          status: 429,
          headers: { 'Retry-After': 'Wed, 07 Oct 2026 00:00:10 GMT' },
        })
      )
      .mockResolvedValueOnce(new Response(null, { status: 401 }));
    vi.stubGlobal('fetch', fetchMock);
    const request = requestTranslation('/paragraphs', {});
    await vi.advanceTimersByTimeAsync(9999);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(1);
    expect((await request).status).toBe(401);
    expect(vi.getTimerCount()).toBe(0);
  });

  it('does not let an old provider response throttle the newly selected provider', async () => {
    const { requestTranslation, resetTranslationCooldown, translationQueue } =
      await import('./translationRequest');
    let finish!: (res: Response) => void;
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise<Response>((resolve) => {
            finish = resolve;
          })
      )
      .mockResolvedValue(new Response('{}'));
    vi.stubGlobal('fetch', fetchMock);
    const oldRequest = requestTranslation('/titles', {});
    resetTranslationCooldown();
    const newRequest = requestTranslation('/paragraphs', {});
    finish(new Response(null, { status: 429, headers: { 'Retry-After': '120' } }));
    expect((await oldRequest).status).toBe(204);
    expect((await newRequest).ok).toBe(true);
    expect(translationQueue.value.recovering).toBe(false);
    expect(vi.getTimerCount()).toBe(0);
  });
});
