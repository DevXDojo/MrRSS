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
  it('limits concurrency and stops queued titles and paragraphs after a 429', async () => {
    const { requestTranslation } = await import('./translationRequest');
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
    const results = await Promise.all(requests);
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(results.filter((res) => res.status === 429)).toHaveLength(18);
    await vi.advanceTimersByTimeAsync(120000);
    const resumed = requestTranslation('/titles', {});
    expect(fetchMock).toHaveBeenCalledTimes(4);
    pending[3](new Response('{}'));
    expect((await resumed).ok).toBe(true);
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
    await requestTranslation('/titles', {});
    resetTranslationCooldown();
    expect((await requestTranslation('/titles', {})).ok).toBe(true);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
});
