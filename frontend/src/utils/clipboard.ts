/**
 * Clipboard utilities for MrRSS
 * Uses Wails v3 native Clipboard API
 */

import { Clipboard } from '@wailsio/runtime';

/**
 * Copy text to clipboard using Wails v3 native API
 * @param text Text to copy
 * @returns Promise that resolves to true if successful, false otherwise
 */
async function copyToClipboard(text: string): Promise<boolean> {
  if (!text) {
    console.warn('copyToClipboard: text is empty');
    return false;
  }

  try {
    await Clipboard.SetText(text);
    return true;
  } catch (error) {
    console.error('Failed to copy to clipboard:', error);
    return false;
  }
}

/**
 * Copy article URL to clipboard
 * @param url Article URL
 * @returns Promise that resolves to true if successful
 */
export async function copyArticleLink(url: string): Promise<boolean> {
  return copyToClipboard(url);
}

/**
 * Copy article title to clipboard
 * @param title Article title
 * @returns Promise that resolves to true if successful
 */
export async function copyArticleTitle(title: string): Promise<boolean> {
  return copyToClipboard(title);
}

/**
 * Copy feed URL to clipboard
 * @param url Feed URL
 * @returns Promise that resolves to true if successful
 */
export async function copyFeedURL(url: string): Promise<boolean> {
  return copyToClipboard(url);
}
