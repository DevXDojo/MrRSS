import i18n from '@/i18n';

/**
 * Opens a URL in the user's default web browser using Wails v3 Browser API.
 * This function calls the backend /api/browser/open endpoint which uses
 * app.Browser.OpenURL() to open URLs securely.
 *
 * @param url - The URL to open (must be http or https)
 * @returns Promise that resolves after reporting success or failure to the user
 */
export async function openInBrowser(url: string): Promise<void> {
  if (!url) {
    console.error('openInBrowser: URL is required');
    return;
  }

  try {
    const response = await fetch('/api/browser/open', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ url }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to open URL: ${errorText}`);
    }

    // Check for redirect instruction (server mode)
    const data = await response.json();
    if (data.redirect) {
      const target = new URL(data.redirect);
      if (target.protocol !== 'http:' && target.protocol !== 'https:') {
        throw new Error('Unsupported browser URL');
      }
      const opened = window.open('about:blank', '_blank');
      if (!opened) throw new Error('Browser popup was blocked');
      opened.opener = null;
      opened.location.href = target.href;
    }

    window.showToast?.(i18n.global.t('common.toast.openedInBrowser'), 'success');
  } catch (error) {
    console.error('Error opening URL in browser:', error);
    // Show user-friendly error message
    if (window.showToast) {
      window.showToast(i18n.global.t('common.errors.failedToOpenLink'), 'error');
    }
  }
}
