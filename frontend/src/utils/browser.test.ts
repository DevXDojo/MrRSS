import { afterEach, describe, expect, it, vi } from 'vitest';
import i18n from '@/i18n';
import { openInBrowser } from './browser';

describe('openInBrowser', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it('shows a success toast after the browser open request succeeds', async () => {
    const showToast = vi.fn();
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ status: 'success' }),
      })
    );
    window.showToast = showToast;

    await openInBrowser('https://example.com');

    expect(showToast).toHaveBeenCalledWith(
      i18n.global.t('common.toast.openedInBrowser'),
      'success'
    );
  });

  it('does not show a success toast when opening the browser fails', async () => {
    const showToast = vi.fn();
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        text: vi.fn().mockResolvedValue('browser unavailable'),
      })
    );
    window.showToast = showToast;

    await openInBrowser('https://example.com');

    expect(showToast).toHaveBeenCalledWith(
      i18n.global.t('common.errors.failedToOpenLink'),
      'error'
    );
    expect(showToast).not.toHaveBeenCalledWith(
      i18n.global.t('common.toast.openedInBrowser'),
      'success'
    );
  });
});
