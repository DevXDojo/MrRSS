import { afterEach, describe, expect, it, vi } from 'vitest';
import { useArticleSummary } from './useArticleSummary';
import type { Article } from '@/types/models';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));

describe('summary recovery', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('retries a temporary fallback instead of treating it as the completed AI summary', async () => {
    window.showToast = vi.fn();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ summary: 'Local fallback', used_fallback: true }))
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ summary: 'AI summary' })));
    vi.stubGlobal('fetch', fetchMock);
    const summary = useArticleSummary();
    summary.summarySettings.value.enabled = true;
    const article = { id: 1 } as Article;
    await summary.generateSummary(article);
    expect((await summary.generateSummary(article))?.summary).toBe('AI summary');
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('does not let a cancelled response replace a newer request for the same article', async () => {
    let finishOld!: (response: Response) => void;
    let finishNew!: (response: Response) => void;
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockImplementationOnce(
          () =>
            new Promise<Response>((resolve) => {
              finishOld = resolve;
            })
        )
        .mockImplementationOnce(
          () =>
            new Promise<Response>((resolve) => {
              finishNew = resolve;
            })
        )
    );
    const summary = useArticleSummary();
    summary.summarySettings.value.enabled = true;
    const article = { id: 1 } as Article;
    const oldRequest = summary.generateSummary(article);
    summary.cancelSummaryGeneration(1);
    const newRequest = summary.generateSummary(article);
    finishOld(new Response(JSON.stringify({ summary: 'Old' })));
    expect(await oldRequest).toBeNull();
    expect(summary.isSummaryLoading(1)).toBe(true);
    finishNew(new Response(JSON.stringify({ summary: 'New' })));
    await newRequest;
    expect(summary.getCachedSummary(1)?.summary).toBe('New');
  });
});
