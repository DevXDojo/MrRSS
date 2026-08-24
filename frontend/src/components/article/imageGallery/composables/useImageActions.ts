import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { openInBrowser } from '@/utils/browser';
import type { Article } from '@/types/models';
import type { ImageActionsReturn } from '../types';

/**
 * Composable for image and article actions (download, copy, favorite, etc.)
 * @returns Image action methods
 */
export function useImageActions(): ImageActionsReturn {
  const store = useAppStore();
  const { t } = useI18n();

  /**
   * Toggle favorite status of an article
   * @param article - The article to toggle favorite for
   * @param event - Optional event to stop propagation
   */
  async function toggleFavorite(article: Article, event?: Event): Promise<void> {
    if (event) {
      event.stopPropagation();
    }
    try {
      const res = await fetch(`/api/articles/favorite?id=${article.id}`, {
        method: 'POST',
      });
      if (res.ok) {
        article.is_favorite = !article.is_favorite;
        // Update filter counts after toggling favorite status
        await store.fetchFilterCounts();
      }
    } catch (e) {
      console.error('Failed to toggle favorite:', e);
    }
  }

  /**
   * Mark article as read
   * @param article - The article to mark as read
   */
  async function markAsRead(article: Article): Promise<void> {
    try {
      const res = await fetch(`/api/articles/read?id=${article.id}&read=true`, {
        method: 'POST',
      });
      if (res.ok) {
        article.is_read = true;
        // Update unread counts after marking as read
        await store.fetchUnreadCounts();
        await store.fetchFilterCounts();
      }
    } catch (e) {
      console.error('Failed to mark as read:', e);
    }
  }

  /**
   * Toggle article read status
   * @param article - The article to toggle read status for
   */
  async function toggleReadStatus(article: Article): Promise<void> {
    const newState = !article.is_read;
    article.is_read = newState;
    try {
      await fetch(`/api/articles/read?id=${article.id}&read=${newState}`, {
        method: 'POST',
      });
      // Update unread counts after toggling read status
      await store.fetchUnreadCounts();
      await store.fetchFilterCounts();
    } catch (e) {
      console.error('Error toggling read status:', e);
      // Revert the state change on error
      article.is_read = !newState;
      window.showToast(t('common.errors.savingSettings'), 'error');
    }
  }

  /**
   * Download image
   * @param src - Image URL to download
   */
  async function downloadImage(src: string): Promise<void> {
    try {
      const response = await fetch(src);
      const blob = await response.blob();

      // Extract and sanitize filename from URL
      let filename = 'image';
      try {
        const url = new URL(src);
        const pathname = url.pathname;
        const pathSegments = pathname.split('/').filter((segment) => segment.length > 0);
        if (pathSegments.length > 0) {
          const lastSegment = pathSegments[pathSegments.length - 1];
          // Remove query params and sanitize filename
          filename = lastSegment.split('?')[0].replace(/[^a-zA-Z0-9._-]/g, '_') || 'image';
        }
      } catch {
        // If URL parsing fails, use default filename
        filename = 'image';
      }

      // Ensure it has a valid extension based on MIME type
      if (!filename.match(/\.(jpg|jpeg|png|gif|webp|svg|bmp)$/i)) {
        const mimeType = blob.type;
        const ext = mimeType.split('/')[1]?.replace('jpeg', 'jpg') || 'png';
        filename = `${filename}.${ext}`;
      }

      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
    } catch (e) {
      console.error('Failed to download image:', e);
      window.showToast(t('common.toast.downloadFailed'), 'error');
    }
  }

  /**
   * Copy image to clipboard (converts to PNG for maximum compatibility)
   * @param src - Image URL to copy
   */
  async function copyImage(src: string): Promise<void> {
    try {
      const response = await fetch(src);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const blob = await response.blob();

      // Convert to PNG for maximum clipboard compatibility
      const pngBlob = await new Promise<Blob>((resolve, reject) => {
        const img = new Image();
        img.crossOrigin = 'anonymous';

        img.onload = () => {
          const canvas = document.createElement('canvas');
          canvas.width = img.width;
          canvas.height = img.height;
          const ctx = canvas.getContext('2d');
          if (ctx) {
            ctx.drawImage(img, 0, 0);
            canvas.toBlob((convertedBlob) => {
              if (convertedBlob) {
                resolve(convertedBlob);
              } else {
                reject(new Error('Failed to convert image to PNG'));
              }
            }, 'image/png');
          } else {
            reject(new Error('Failed to get canvas context'));
          }
        };

        img.onerror = () => {
          reject(new Error('Failed to load image for conversion'));
        };

        img.src = URL.createObjectURL(blob);
      });

      // Copy to clipboard using only PNG format (widely supported)
      await navigator.clipboard.write([
        new ClipboardItem({
          'image/png': pngBlob,
        }),
      ]);

      window.showToast(t('common.toast.copiedToClipboard'), 'success');
    } catch (error) {
      console.error('Failed to copy image:', error);
      window.showToast(t('common.errors.failedToCopy'), 'error');
    }
  }

  /**
   * Open original article in browser
   * @param article - The article to open
   */
  function openOriginal(article: Article): void {
    openInBrowser(article.url);
  }

  /**
   * Copy article title to clipboard
   * @param article - The article to copy title from
   */
  async function copyArticleTitle(article: Article): Promise<void> {
    try {
      await navigator.clipboard.writeText(article.title);
      window.showToast(t('common.toast.copiedToClipboard'), 'success');
    } catch (error) {
      console.error('Failed to copy title:', error);
      window.showToast(t('common.errors.failedToCopy'), 'error');
    }
  }

  /**
   * Copy article link to clipboard
   * @param article - The article to copy link from
   */
  async function copyArticleLink(article: Article): Promise<void> {
    try {
      await navigator.clipboard.writeText(article.url);
      window.showToast(t('common.toast.copiedToClipboard'), 'success');
    } catch (error) {
      console.error('Failed to copy link:', error);
      window.showToast(t('common.errors.failedToCopy'), 'error');
    }
  }

  return {
    toggleFavorite,
    markAsRead,
    toggleReadStatus,
    downloadImage,
    copyImage,
    openOriginal,
    copyArticleTitle,
    copyArticleLink,
  };
}
