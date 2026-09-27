import { afterEach, describe, expect, it, vi } from 'vitest';
import i18n from '@/i18n';
import { openInBrowser } from './browser';

describe('openInBrowser', () => {
  const originalToast = window.showToast;
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
    window.showToast = originalToast;
  });

  it('does not report success when the server-mode popup is blocked', async () => {
    window.showToast = vi.fn();
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ redirect: 'https://example.com' }),
      })
    );
    vi.spyOn(window, 'open').mockReturnValue(null);
    await openInBrowser('https://example.com');
    expect(window.showToast).toHaveBeenCalledExactlyOnceWith(
      i18n.global.t('common.errors.failedToOpenLink'),
      'error'
    );
  });

  it('isolates the opener before navigating a server-mode popup', async () => {
    window.showToast = vi.fn();
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ redirect: 'https://example.com/' }),
      })
    );
    const popup = { opener: window, location: { href: '' } };
    vi.spyOn(window, 'open').mockReturnValue(popup as unknown as Window);
    await openInBrowser('https://example.com/');
    expect(popup.opener).toBeNull();
    expect(popup.location.href).toBe('https://example.com/');
    expect(window.showToast).toHaveBeenCalledWith(
      i18n.global.t('common.toast.openedInBrowser'),
      'success'
    );
  });

  it('rejects unsafe server redirect schemes', async () => {
    window.showToast = vi.fn();
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ redirect: 'javascript:alert(1)' }),
      })
    );
    const open = vi.spyOn(window, 'open');
    await openInBrowser('https://example.com/');
    expect(open).not.toHaveBeenCalled();
    expect(window.showToast).toHaveBeenCalledWith(
      i18n.global.t('common.errors.failedToOpenLink'),
      'error'
    );
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
