import { afterEach, describe, expect, it, vi } from 'vitest';
import { Clipboard } from '@wailsio/runtime';
import { copyArticleLink } from './clipboard';

vi.mock('@wailsio/runtime', () => ({ Clipboard: { SetText: vi.fn() } }));
afterEach(() => {
  vi.restoreAllMocks();
  vi.clearAllMocks();
  vi.unstubAllGlobals();
});

describe('copy article link', () => {
  it('uses the browser clipboard when available', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal('navigator', { clipboard: { writeText } });
    expect(await copyArticleLink('https://example.com/a')).toBe(true);
    expect(writeText).toHaveBeenCalledWith('https://example.com/a');
    expect(Clipboard.SetText).not.toHaveBeenCalled();
  });

  it('falls back to native clipboard when browser permission is denied', async () => {
    const writeText = vi.fn().mockRejectedValue(new Error('denied'));
    vi.stubGlobal('navigator', { clipboard: { writeText } });
    vi.mocked(Clipboard.SetText).mockResolvedValue(undefined);
    expect(await copyArticleLink('https://example.com/b')).toBe(true);
    expect(Clipboard.SetText).toHaveBeenCalledWith('https://example.com/b');
  });
});
