import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { openInBrowser } from '@/utils/browser';
import type { Article } from '@/types/models';
import { proxyImagesInHtml, isMediaCacheEnabled } from '@/utils/mediaProxy';

type ViewMode = 'original' | 'rendered' | 'external';
type RenderAction = 'showContent' | 'showOriginal' | null;

interface ViewModeChangeEvent extends Event {
  detail: {
    mode: ViewMode;
  };
}

interface RenderActionEvent extends Event {
  detail: {
    action: RenderAction;
  };
}

export function useArticleDetail() {
  const store = useAppStore();
  const { t, locale } = useI18n();

  const article = computed<Article | undefined>(() =>
    store.articles.find((a) => a.id === store.currentArticleId)
  );

  // Get current article index in the filtered list
  const currentArticleIndex = computed(() => {
    if (!store.currentArticleId) return -1;
    return store.articles.findIndex((a) => a.id === store.currentArticleId);
  });

  // Check if there's a previous article
  const hasPreviousArticle = computed(() => currentArticleIndex.value > 0);

  // Check if there's a next article
  const hasNextArticle = computed(
    () => currentArticleIndex.value >= 0 && currentArticleIndex.value < store.articles.length - 1
  );

  // Navigate to previous article
  function goToPreviousArticle() {
    if (hasPreviousArticle.value) {
      const prevArticle = store.articles[currentArticleIndex.value - 1];
      store.currentArticleId = prevArticle.id;
      markAsReadIfNeeded(prevArticle);
      scrollArticleIntoView(prevArticle.id);
    }
  }

  // Navigate to next article
  function goToNextArticle() {
    if (hasNextArticle.value) {
      const nextArticle = store.articles[currentArticleIndex.value + 1];
      store.currentArticleId = nextArticle.id;
      markAsReadIfNeeded(nextArticle);
      scrollArticleIntoView(nextArticle.id);
    }
  }

  // Scroll article into view in the article list
  function scrollArticleIntoView(articleId: number) {
    setTimeout(() => {
      const articleEl = document.querySelector(`[data-article-id="${articleId}"]`);
      if (articleEl) {
        articleEl.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
      }
    }, 50);
  }

  // Mark article as read if it's not already read
  async function markAsReadIfNeeded(article: Article) {
    if (!article.is_read) {
      article.is_read = true;
      try {
        await fetch(`/api/articles/read?id=${article.id}&read=true`, {
          method: 'POST',
        });
        store.fetchUnreadCounts();
      } catch (e) {
        console.error('Error marking as read:', e);
      }
    }
  }

  // Expose articles list and index for UI display
  const articles = computed(() => store.articles);
  const currentArticleIndexForDisplay = computed(() => currentArticleIndex.value + 1);

  // Get effective view mode based on feed settings and global settings
  function getEffectiveViewMode(): ViewMode {
    if (!article.value) return defaultViewMode.value;

    // Find the feed for this article
    const feed = store.feeds.find((f) => f.id === article.value!.feed_id);
    if (!feed) return defaultViewMode.value;

    // Check feed's article_view_mode
    if (feed.article_view_mode === 'webpage') {
      return 'original';
    } else if (feed.article_view_mode === 'rendered') {
      return 'rendered';
    } else if (feed.article_view_mode === 'external') {
      return 'external';
    } else {
      // 'global' or undefined - use global setting
      return defaultViewMode.value;
    }
  }

  const showContent = ref(false);
  const articleContent = ref('');
  const isLoadingContent = ref(false);
  const currentArticleId = ref<number | null>(null);
  const defaultViewMode = ref<ViewMode>('original');
  const pendingRenderAction = ref<RenderAction>(null);
  const imageViewerSrc = ref<string | null>(null);
  const imageViewerAlt = ref('');
  const imageViewerImages = ref<string[]>([]);
  const imageViewerInitialIndex = ref(0);

  // Watch for article changes and apply view mode
  watch(
    () => store.currentArticleId,
    async (newId, oldId) => {
      if (newId && newId !== oldId) {
        // Close image viewer when switching articles
        imageViewerSrc.value = null;
        imageViewerAlt.value = '';
        imageViewerImages.value = [];
        imageViewerInitialIndex.value = 0;

        // Reset content when switching articles
        articleContent.value = '';
        currentArticleId.value = null;

        // Always fetch article content for AI chat and translation features
        await fetchArticleContent();

        // Check if there's a pending render action from context menu
        if (pendingRenderAction.value) {
          // Apply the explicit action instead of default
          // Don't save user preference for context menu actions - they're one-time actions
          if (pendingRenderAction.value === 'showContent') {
            showContent.value = true;
          } else if (pendingRenderAction.value === 'showOriginal') {
            showContent.value = false;
          }
          pendingRenderAction.value = null; // Clear the pending action
        } else {
          // Apply user's preferred mode for this article from store, or determine from feed/global settings
          const storedPreference = store.articleViewModePreferences.get(newId);
          const effectiveMode = getEffectiveViewMode();
          const preferredMode = storedPreference || effectiveMode;
          showContent.value = preferredMode === 'rendered';

          // Save to localStorage if this is the first time viewing this article
          // and there's no stored preference (we're using the default mode)
          if (!storedPreference) {
            const mode = showContent.value ? 'rendered' : 'original';
            store.articleViewModePreferences.set(newId, mode);
            try {
              const preferences = Object.fromEntries(store.articleViewModePreferences.entries());
              localStorage.setItem('articleViewModePreferences', JSON.stringify(preferences));
            } catch (e) {
              console.error('Failed to save article view mode to localStorage:', e);
            }
          }
        }
      }
    }
  );

  // Watch for feed/filter changes and close image viewer
  watch(
    () => [store.currentFeedId, store.currentFilter, store.currentCategory],
    () => {
      // Close image viewer when switching feeds, filters, or categories
      imageViewerSrc.value = null;
      imageViewerAlt.value = '';
      imageViewerImages.value = [];
      imageViewerInitialIndex.value = 0;
    }
  );

  // Listen for default view mode changes from settings
  window.addEventListener('default-view-mode-changed', (e: Event) => {
    const event = e as ViewModeChangeEvent;
    defaultViewMode.value = event.detail.mode;
    // Clear all user preferences when default changes
    store.articleViewModePreferences.clear();
  });

  function close() {
    store.currentArticleId = null;
    showContent.value = false;
    articleContent.value = '';
    currentArticleId.value = null;
  }

  function toggleRead() {
    if (!article.value) return;
    const newState = !article.value.is_read;
    article.value.is_read = newState;
    fetch(`/api/articles/read?id=${article.value.id}&read=${newState}`, {
      method: 'POST',
    });
  }

  function toggleFavorite() {
    if (!article.value) return;
    const newState = !article.value.is_favorite;
    article.value.is_favorite = newState;
    fetch(`/api/articles/favorite?id=${article.value.id}`, { method: 'POST' });
  }

  async function toggleReadLater() {
    if (!article.value) return;
    const newState = !article.value.is_read_later;
    article.value.is_read_later = newState;
    // When adding to read later, also mark as unread
    if (newState) {
      article.value.is_read = false;
    }
    try {
      await fetch(`/api/articles/toggle-read-later?id=${article.value.id}`, { method: 'POST' });
      store.fetchUnreadCounts();
    } catch (e) {
      console.error('Error toggling read later:', e);
      // Revert on error
      article.value.is_read_later = !newState;
    }
  }

  function openOriginal() {
    if (article.value) openInBrowser(article.value.url);
  }

  async function toggleContentView() {
    if (!showContent.value) {
      // Switching to content view - fetch content if needed
      if (!article.value) return;
      // Check if we need to fetch content (different article or no content yet)
      if (currentArticleId.value !== article.value.id) {
        await fetchArticleContent();
      }
    }
    showContent.value = !showContent.value;
    // Remember user's preference for this specific article
    if (article.value) {
      const mode = showContent.value ? 'rendered' : 'original';

      // Save to both store and localStorage for persistence
      store.articleViewModePreferences.set(article.value.id, mode);

      // Also save to localStorage as backup
      try {
        const preferences = Object.fromEntries(store.articleViewModePreferences.entries());
        localStorage.setItem('articleViewModePreferences', JSON.stringify(preferences));
      } catch (e) {
        console.error('Failed to save article view mode to localStorage:', e);
      }
    }
  }

  async function fetchArticleContent() {
    if (!article.value) return;

    const loadingArticleId = article.value.id;
    currentArticleId.value = loadingArticleId; // Track which article we're loading
    isLoadingContent.value = true;

    try {
      const res = await fetch(`/api/articles/content?id=${loadingArticleId}`);
      if (currentArticleId.value !== loadingArticleId) return;

      if (res.ok) {
        const data = await res.json();
        if (currentArticleId.value !== loadingArticleId) return;

        let content = data.content || '';

        // Proxy images if media cache is enabled
        const cacheEnabled = await isMediaCacheEnabled();
        if (currentArticleId.value !== loadingArticleId) return;

        if (cacheEnabled && content) {
          // Use feed URL as referer for anti-hotlinking (more reliable than article URL)
          const feedUrl = data.feed_url || article.value.url;
          content = proxyImagesInHtml(content, feedUrl);
        }

        articleContent.value = content;

        // Only show loading animation for non-cached content
        if (!data.cached) {
          // Content was fetched from feed, show loading and trigger watch
          await nextTick(); // Ensure content is rendered first
        }
      } else {
        console.error('Failed to fetch article content');
        articleContent.value = '';
      }
    } catch (e) {
      console.error('Error fetching article content:', e);
      articleContent.value = '';
    } finally {
      if (currentArticleId.value === loadingArticleId) {
        isLoadingContent.value = false;
      }
    }
  }

  // Handle retry loading content
  function handleRetryLoadContent() {
    if (article.value && showContent.value) {
      fetchArticleContent();
    }
  }

  async function reloadArticleContent() {
    if (!article.value) return;

    const reloadingArticleId = article.value.id;
    articleContent.value = '';
    currentArticleId.value = null;
    isLoadingContent.value = true;

    try {
      const res = await fetch(`/api/articles/reload-content?id=${reloadingArticleId}`, {
        method: 'POST',
      });
      if (!res.ok) {
        throw new Error(`Reload content failed: ${res.status}`);
      }
      if (store.currentArticleId === reloadingArticleId) {
        await fetchArticleContent();
      }
    } catch (e) {
      console.error('Error reloading article content:', e);
      window.showToast(t('common.errors.fetchingArticleContent'), 'error');
      if (store.currentArticleId === reloadingArticleId) {
        isLoadingContent.value = false;
      }
    }
  }

  // Unwrap images from hyperlinks
  // This ensures images can be clicked directly without triggering link navigation
  // Works on both main content and translated content
  function unwrapImagesFromLinks() {
    // Process all links in prose content (both main content and translations)
    const links = document.querySelectorAll<HTMLAnchorElement>('.prose-content a, .prose a');
    const linksToProcess: HTMLAnchorElement[] = [];

    // Collect links that contain images (check both direct children and nested)
    links.forEach((link) => {
      const images = link.querySelectorAll('img');
      if (images.length > 0) {
        linksToProcess.push(link);
      }
    });

    // Process collected links
    linksToProcess.forEach((link) => {
      try {
        // Check if parent exists and is valid
        if (!link.parentNode) {
          return;
        }

        // Extract all child nodes from the link
        const fragment = document.createDocumentFragment();
        while (link.firstChild) {
          fragment.appendChild(link.firstChild);
        }

        // Replace the link with its contents
        link.parentNode.replaceChild(fragment, link);
      } catch (error) {
        console.error('Error unwrapping image from link:', error);
      }
    });
  }

  // Attach event listeners to images in rendered content
  // Can be called multiple times (e.g., after translations modify the DOM)
  function attachImageEventListeners() {
    // First, unwrap any images that are inside hyperlinks
    unwrapImagesFromLinks();

    // Get all images in prose content (use more specific selector)
    const proseContainers = document.querySelectorAll('.prose-content, .prose');

    if (proseContainers.length === 0) {
      return;
    }

    const images = document.querySelectorAll<HTMLImageElement>('.prose-content img, .prose img');

    // Process images if there are any
    if (images.length > 0) {
      images.forEach((img) => {
        try {
          // Verify the image has a valid parent
          if (!img.parentNode) {
            return;
          }

          // Skip images that are very small (likely icons/emojis)
          const isSmallIcon = img.height <= 24 && img.height > 0;
          if (isSmallIcon) {
            return;
          }

          // Ensure the image and its parents can receive pointer events
          img.style.cursor = 'pointer';
          img.style.pointerEvents = 'auto';

          // Remove old listeners by replacing with clone
          const newImg = img.cloneNode(true) as HTMLImageElement;
          img.parentNode.replaceChild(newImg, img);

          // Ensure cloned image maintains pointer interaction styles
          newImg.style.cursor = 'pointer';
          newImg.style.pointerEvents = 'auto';

          // Left click - open image viewer with all images from article
          newImg.addEventListener(
            'click',
            (e: Event) => {
              e.preventDefault();
              e.stopPropagation(); // Prevent event bubbling to parent elements

              // Verify image src exists
              if (!newImg.src) {
                return;
              }

              // Collect all images from the article content
              const allImages = Array.from(
                document.querySelectorAll<HTMLImageElement>('.prose-content img, .prose img')
              )
                .filter((img) => {
                  // Filter out small icons
                  return !(img.height <= 24 && img.height > 0);
                })
                .map((img) => img.src);

              // Find the index of the clicked image
              const clickedIndex = allImages.findIndex((src) => src === newImg.src);

              // Set up image viewer with all images
              imageViewerSrc.value = newImg.src;
              imageViewerAlt.value = newImg.alt || '';
              imageViewerImages.value = allImages;
              imageViewerInitialIndex.value = clickedIndex >= 0 ? clickedIndex : 0;
            },
            { capture: true }
          ); // Use capture phase to ensure we get the event first

          // Right click - show context menu for saving
          newImg.addEventListener(
            'contextmenu',
            (e: MouseEvent) => {
              e.preventDefault();
              e.stopPropagation(); // Prevent event bubbling to parent elements

              // Verify image src exists
              if (!newImg.src) {
                return;
              }

              // Use global context menu system
              window.dispatchEvent(
                new CustomEvent('open-context-menu', {
                  detail: {
                    x: e.clientX,
                    y: e.clientY,
                    items: [
                      {
                        label: t('common.contextMenu.copyImage'),
                        action: 'copy',
                        icon: 'PhCopy',
                      },
                      {
                        label: t('article.action.viewImage'),
                        action: 'view',
                        icon: 'PhMagnifyingGlassPlus',
                      },
                      {
                        label: t('common.contextMenu.downloadImage'),
                        action: 'download',
                        icon: 'PhDownloadSimple',
                      },
                    ],
                    data: { src: newImg.src },
                    callback: (action: string, data: { src: string }) => {
                      if (action === 'copy') {
                        copyImage(data.src);
                      } else if (action === 'view') {
                        imageViewerSrc.value = data.src;
                        imageViewerAlt.value = '';
                      } else if (action === 'download') {
                        downloadImage(data.src);
                      }
                    },
                  },
                })
              );
            },
            { capture: true }
          ); // Use capture phase to ensure we get the event first
        } catch (error) {
          console.error('Error attaching event listeners to image:', error);
        }
      });
    }

    // Always attach link event listeners (even if there are no images)
    attachLinkEventListeners();
  }

  // Attach event listeners to links in rendered content
  // Uses the same strategy as image handling for consistency
  // Works for dynamically added content (e.g., translations)
  function attachLinkEventListeners() {
    // Get all text-only links (no images) in prose content
    const links = document.querySelectorAll<HTMLAnchorElement>('.prose-content a, .prose a');

    links.forEach((link) => {
      try {
        // Verify the link has a valid parent
        if (!link.parentNode) {
          return;
        }

        // Skip if the link contains an image (handled separately by image handlers)
        if (link.querySelector('img')) {
          return;
        }

        // Check if already processed (has our marker)
        if (link.dataset.linkHandlerAttached === 'true') {
          return;
        }

        // Mark as processed
        link.dataset.linkHandlerAttached = 'true';

        // Add event listener using capture phase
        link.addEventListener(
          'click',
          (e: Event) => {
            // Prevent default behavior and stop propagation
            e.preventDefault();
            e.stopPropagation();
            e.stopImmediatePropagation();

            // Get the href
            let href = link.getAttribute('href');
            if (href) {
              // Convert relative URLs to absolute URLs
              // If it starts with / or is a relative path, convert to absolute using article URL
              if (
                href.startsWith('/') ||
                (!href.startsWith('http://') &&
                  !href.startsWith('https://') &&
                  !href.startsWith('mailto:') &&
                  !href.startsWith('#'))
              ) {
                if (article.value?.url) {
                  try {
                    const articleUrl = new URL(article.value.url);
                    // For relative paths like "path/to/page"
                    if (!href.startsWith('/')) {
                      href = new URL(href, article.value.url).href;
                    } else {
                      // For absolute paths like "/path/to/page"
                      href = `${articleUrl.origin}${href}`;
                    }
                  } catch (error) {
                    console.error('Error converting relative URL:', error);
                  }
                }
              }

              try {
                openInBrowser(href);
              } catch (error) {
                console.error('Error opening URL:', href, error);
              }
            }
          },
          { capture: true } // Use capture phase to intercept clicks early
        );
      } catch (error) {
        console.error('Error attaching event listeners to link:', error);
      }
    });
  }

  function closeImageViewer() {
    imageViewerSrc.value = null;
    imageViewerAlt.value = '';
    imageViewerImages.value = [];
    imageViewerInitialIndex.value = 0;
  }

  // Copy image to clipboard
  async function copyImage(src: string) {
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

  // Download image from URL
  async function downloadImage(src: string) {
    try {
      const response = await fetch(src);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const blob = await response.blob();

      // Extract and sanitize filename from URL
      let filename = 'image';
      try {
        const url = new URL(src);
        const pathname = url.pathname;
        const pathSegments = pathname.split('/').filter((segment) => segment.length > 0);
        if (pathSegments.length > 0) {
          const lastSegment = pathSegments[pathSegments.length - 1];
          filename = lastSegment.split('?')[0].replace(/[^a-zA-Z0-9._-]/g, '_') || 'image';
        }
      } catch {
        filename = 'image';
      }

      // Ensure it has a valid extension based on MIME type
      if (!filename.match(/\.(jpg|jpeg|png|gif|webp|svg|bmp)$/i)) {
        const mimeType = blob.type;
        const ext = mimeType.split('/')[1]?.replace('jpeg', 'jpg') || 'png';
        filename = `${filename}.${ext}`;
      }

      // Create download link
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to download image:', error);
      window.open(src, '_blank');
    }
  }

  // Export article to Obsidian
  async function exportToObsidian() {
    if (!article.value) return;

    try {
      window.showToast(t('setting.plugins.obsidian.exporting'), 'info');

      const response = await fetch('/api/articles/export/obsidian', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          article_id: article.value.id,
        }),
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(error);
      }

      const data = await response.json();

      // Show success message with file path
      const message = data.message || t('setting.plugins.obsidian.exported');
      const filePath = data.file_path ? ` (${data.file_path})` : '';
      window.showToast(message + filePath, 'success');
    } catch (error) {
      console.error('Failed to export to Obsidian:', error);
      const message =
        error instanceof Error ? error.message : t('setting.plugins.obsidian.exportFailed');
      window.showToast(message, 'error');
    }
  }

  // Export article to Notion
  async function exportToNotion() {
    if (!article.value) return;

    try {
      window.showToast(t('setting.plugins.notion.exporting'), 'info');

      const response = await fetch('/api/articles/export/notion', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          article_id: article.value.id,
        }),
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(error);
      }

      const data = await response.json();

      // Show success message with page URL
      const message = data.message || t('setting.plugins.notion.exported');
      window.showToast(message, 'success');

      // Open the Notion page in external browser
      if (data.page_url) {
        openInBrowser(data.page_url);
      }
    } catch (error) {
      console.error('Failed to export to Notion:', error);
      const message =
        error instanceof Error ? error.message : t('setting.plugins.notion.exportFailed');
      window.showToast(message, 'error');
    }
  }

  // Export article to Zotero
  async function exportToZotero() {
    if (!article.value) return;

    try {
      window.showToast(t('setting.plugins.zotero.exporting'), 'info');

      const response = await fetch('/api/articles/export/zotero', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          article_id: article.value.id,
        }),
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(error);
      }

      const data = await response.json();

      // Show success message
      const message = data.message || t('setting.plugins.zotero.exported');
      window.showToast(message, 'success');
    } catch (error) {
      console.error('Failed to export to Zotero:', error);
      const message =
        error instanceof Error ? error.message : t('setting.plugins.zotero.exportFailed');
      window.showToast(message, 'error');
    }
  }

  // Listen for render content event from context menu
  async function handleRenderContent(e: Event) {
    const event = e as RenderActionEvent;
    if (!article.value) return;

    const action = event.detail?.action || 'showContent';

    // Mark as read when rendering content
    if (!article.value.is_read) {
      article.value.is_read = true;
      fetch(`/api/articles/read?id=${article.value.id}&read=true`, { method: 'POST' });
    }

    if (action === 'showContent') {
      // Check if we need to fetch content for this article
      if (currentArticleId.value !== article.value.id) {
        await fetchArticleContent();
      }
      showContent.value = true;
      // Don't set userPreferredMode for context menu actions
    } else if (action === 'showOriginal') {
      showContent.value = false;
      // Don't set userPreferredMode for context menu actions
    }
  }

  // Listen for explicit render action from context menu (before article selection)
  function handleExplicitRenderAction(e: Event) {
    const event = e as RenderActionEvent;
    pendingRenderAction.value = event.detail?.action;
  }

  // Handle toggle content view from keyboard shortcut
  function handleToggleContentView() {
    if (article.value) {
      toggleContentView();
    }
  }

  onMounted(async () => {
    // Restore preferences from localStorage if store is empty
    if (store.articleViewModePreferences.size === 0) {
      try {
        const saved = localStorage.getItem('articleViewModePreferences');
        if (saved) {
          const preferences = JSON.parse(saved) as Record<number, 'original' | 'rendered'>;
          Object.entries(preferences).forEach(([articleId, mode]) => {
            store.articleViewModePreferences.set(Number(articleId), mode);
          });
        }
      } catch (e) {
        console.error('Failed to restore article view mode preferences from localStorage:', e);
      }
    }

    // If there's already a current article selected (e.g., after switching back from image gallery),
    // apply the saved preference and fetch content if needed
    if (store.currentArticleId) {
      const storedPreference = store.articleViewModePreferences.get(store.currentArticleId);
      if (storedPreference) {
        showContent.value = storedPreference === 'rendered';
      }

      // Fetch article content if not already loaded
      if (currentArticleId.value !== store.currentArticleId || !articleContent.value) {
        await fetchArticleContent();
      }
    }

    window.addEventListener('render-article-content', handleRenderContent);
    window.addEventListener('explicit-render-action', handleExplicitRenderAction);
    window.addEventListener('toggle-content-view', handleToggleContentView);

    // Load default view mode from settings
    try {
      const res = await fetch('/api/settings');
      const data = await res.json();
      defaultViewMode.value = data.default_view_mode || 'original';
    } catch (e) {
      console.error('Error loading settings:', e);
    }
  });

  onBeforeUnmount(() => {
    window.removeEventListener('render-article-content', handleRenderContent);
    window.removeEventListener('explicit-render-action', handleExplicitRenderAction);
    window.removeEventListener('toggle-content-view', handleToggleContentView);
  });

  return {
    // Reactive state
    article,
    showContent,
    articleContent,
    isLoadingContent,
    imageViewerSrc,
    imageViewerAlt,
    imageViewerImages,
    imageViewerInitialIndex,
    locale,
    hasPreviousArticle,
    hasNextArticle,
    articles,
    currentArticleIndex: currentArticleIndexForDisplay,

    // Functions
    close,
    toggleRead,
    toggleFavorite,
    toggleReadLater,
    openOriginal,
    toggleContentView,
    reloadArticleContent,
    closeImageViewer,
    copyImage,
    downloadImage,
    exportToObsidian,
    exportToNotion,
    exportToZotero,
    attachImageEventListeners, // Expose for re-attaching after content modifications
    handleRetryLoadContent,
    goToPreviousArticle,
    goToNextArticle,

    // Translations
    t,
  };
}
