import i18n from '@/i18n';

/**
 * Opens a URL in the user's default web browser using Wails v3 Browser API.
 * This function calls the backend /api/browser/open endpoint which uses
 * app.Browser.OpenURL() to open URLs securely.
 *
 * @param url - The URL to open (must be http or https)
 * @returns Promise that resolves when URL is successfully opened
 * @throws Error if URL is invalid or browser operation fails
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
      window.open(data.redirect, '_blank');
    }

    window.showToast(i18n.global.t('common.toast.openedInBrowser'), 'success');
  } catch (error) {
    console.error('Error opening URL in browser:', error);
    // Show user-friendly error message
    if (window.showToast) {
      window.showToast(i18n.global.t('common.errors.failedToOpenLink'), 'error');
    }
  }
}
