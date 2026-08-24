import { ref, onMounted, onBeforeUnmount } from 'vue';
import { useAppStore } from '@/stores/app';
import { openInBrowser } from '@/utils/browser';

interface KeyboardShortcuts {
  nextArticle: string;
  previousArticle: string;
  nextArticleArrow: string;
  previousArticleArrow: string;
  openArticle: string;
  closeArticle: string;
  toggleReadStatus: string;
  toggleFavoriteStatus: string;
  toggleReadLaterStatus: string;
  openInBrowser: string;
  toggleContentView: string;
  refreshFeeds: string;
  markAllRead: string;
  openSettings: string;
  addFeed: string;
  focusSearch: string;
  toggleFilter: string;
  toggleUnreadFilter: string;
  toggleFavoritesFilter: string;
  toggleReadLaterFilter: string;
  goToAllArticles: string;
  goToUnread: string;
  goToFavorites: string;
  goToReadLater: string;
}

interface KeyboardShortcutCallbacks {
  onOpenSettings: () => void;
  onAddFeed: () => void;
  onMarkAllRead: () => Promise<void>;
}

export function useKeyboardShortcuts(callbacks: KeyboardShortcutCallbacks) {
  const store = useAppStore();

  const shortcutsEnabled = ref(true);
  const shortcuts = ref<KeyboardShortcuts>({
    nextArticle: 'j',
    previousArticle: 'k',
    nextArticleArrow: 'ArrowRight',
    previousArticleArrow: 'ArrowLeft',
    openArticle: 'Enter',
    closeArticle: 'Escape',
    toggleReadStatus: 'r',
    toggleFavoriteStatus: 's',
    toggleReadLaterStatus: 'l',
    openInBrowser: 'o',
    toggleContentView: 'v',
    refreshFeeds: 'Shift+r',
    markAllRead: 'Shift+a',
    openSettings: ',',
    addFeed: 'a',
    focusSearch: '/',
    toggleFilter: 'f',
    toggleUnreadFilter: 'Alt+r',
    toggleFavoritesFilter: 'Alt+s',
    toggleReadLaterFilter: 'Alt+l',
    goToAllArticles: '1',
    goToUnread: '2',
    goToFavorites: '3',
    goToReadLater: '4',
  });

  // Helper functions
  function buildKeyCombo(e: KeyboardEvent): string {
    let key = '';
    if (e.ctrlKey) key += 'Ctrl+';
    if (e.altKey) key += 'Alt+';
    if (e.shiftKey) key += 'Shift+';
    if (e.metaKey) key += 'Meta+';

    let actualKey = e.key;
    if (actualKey === ' ') actualKey = 'Space';
    else if (actualKey.length === 1) actualKey = actualKey.toLowerCase();

    key += actualKey;
    return key;
  }

  function navigateArticle(direction: number): void {
    const articles = store.articles;
    if (!articles || articles.length === 0) return;

    const currentIndex = store.currentArticleId
      ? articles.findIndex((a) => a.id === store.currentArticleId)
      : -1;

    let newIndex: number;
    if (currentIndex === -1) {
      newIndex = direction > 0 ? 0 : articles.length - 1;
    } else {
      newIndex = currentIndex + direction;
      if (newIndex < 0) newIndex = 0;
      if (newIndex >= articles.length) newIndex = articles.length - 1;
    }

    selectArticleByIndex(newIndex);
  }

  function selectArticleByIndex(index: number): void {
    const article = store.articles[index];
    if (!article) return;

    store.currentArticleId = article.id;

    // Mark as read
    if (!article.is_read) {
      article.is_read = true;
      fetch(`/api/articles/read?id=${article.id}&read=true`, { method: 'POST' })
        .then(() => store.fetchUnreadCounts())
        .catch((e) => console.error('Error marking as read:', e));
    }

    // Scroll the article into view
    setTimeout(() => {
      const articleEl = document.querySelector(`[data-article-id="${article.id}"]`);
      if (articleEl) {
        articleEl.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
      }
    }, 50);
  }

  function toggleCurrentArticleRead(): void {
    const article = store.articles.find((a) => a.id === store.currentArticleId);
    if (!article) return;

    const newState = !article.is_read;
    article.is_read = newState;
    fetch(`/api/articles/read?id=${article.id}&read=${newState}`, { method: 'POST' })
      .then(() => store.fetchUnreadCounts())
      .catch((e) => {
        console.error('Error toggling read:', e);
        article.is_read = !newState;
      });
  }

  function toggleCurrentArticleFavorite(): void {
    const article = store.articles.find((a) => a.id === store.currentArticleId);
    if (!article) return;

    const newState = !article.is_favorite;
    article.is_favorite = newState;
    fetch(`/api/articles/favorite?id=${article.id}`, { method: 'POST' }).catch((e) => {
      console.error('Error toggling favorite:', e);
      article.is_favorite = !newState;
    });
  }

  function toggleCurrentArticleReadLater(): void {
    const article = store.articles.find((a) => a.id === store.currentArticleId);
    if (!article) return;

    const newState = !article.is_read_later;
    article.is_read_later = newState;
    // When adding to read later, also mark as unread
    if (newState) {
      article.is_read = false;
    }
    fetch(`/api/articles/toggle-read-later?id=${article.id}`, { method: 'POST' })
      .then(() => store.fetchUnreadCounts())
      .catch((e) => {
        console.error('Error toggling read later:', e);
        article.is_read_later = !newState;
      });
  }

  function openCurrentArticleInBrowser(): void {
    const article = store.articles.find((a) => a.id === store.currentArticleId);
    if (article && article.url) {
      openInBrowser(article.url);
    }
  }

  function focusSearchInput(): void {
    const searchInput = document.querySelector('[data-search-input]') as HTMLInputElement;
    if (searchInput) {
      searchInput.focus();
    }
  }

  function toggleListFilter(filter: 'unread' | 'favorites' | 'readLater'): void {
    store.setFilter(store.currentFilter === filter ? 'all' : filter);
  }

  // Check if an article detail panel is open and scrollable
  function isArticleDetailOpen(): boolean {
    // Check if there's a current article selected
    if (!store.currentArticleId) return false;

    // Check if the article detail panel is visible
    const articleDetail = document.querySelector('main[class*="flex-1 bg-bg-primary"]');
    if (!articleDetail) return false;

    // Check if the article detail has scrollable content
    // Check for both overflow-y-auto and overflow-y-scroll
    const scrollableContent = articleDetail.querySelector('.overflow-y-auto, .overflow-y-scroll');
    if (!scrollableContent) return false;

    return true;
  }

  // Check if currently viewing original webpage (iframe mode)
  function isWebpageViewMode(): boolean {
    const iframe = document.querySelector('iframe[src*="/api/webpage/proxy"]');
    return iframe !== null;
  }

  // Scroll the article detail panel
  function scrollArticleDetail(direction: 'up' | 'down' | 'pageDown' | 'pageUp'): void {
    const articleDetail = document.querySelector('main[class*="flex-1 bg-bg-primary"]');
    if (!articleDetail) return;

    // Check for both overflow-y-auto and overflow-y-scroll
    const scrollableContent = articleDetail.querySelector(
      '.overflow-y-auto, .overflow-y-scroll'
    ) as HTMLElement;
    if (!scrollableContent) return;

    const scrollAmount =
      direction === 'pageDown' || direction === 'pageUp'
        ? scrollableContent.clientHeight * 0.9
        : 100; // For arrow keys

    const newScrollTop =
      direction === 'down' || direction === 'pageDown'
        ? scrollableContent.scrollTop + scrollAmount
        : scrollableContent.scrollTop - scrollAmount;

    scrollableContent.scrollTo({
      top: newScrollTop,
      behavior: 'smooth',
    });
  }

  // Keyboard event handler
  function handleKeyboardShortcut(e: KeyboardEvent): void {
    // Skip if shortcuts are disabled
    if (!shortcutsEnabled.value) {
      return;
    }

    // Check if image viewer is open - if so, let it handle arrow keys
    const imageViewerOpen = document.querySelector('[data-image-viewer="true"]') !== null;
    if (imageViewerOpen) {
      // Image viewer handles its own keyboard events
      // Only ESC key should be handled here to close the viewer
      const key = buildKeyCombo(e);
      if (key === shortcuts.value.closeArticle) {
        // Let the image viewer's ESC handler close it
        return;
      }
      // Block all other shortcuts when image viewer is open
      return;
    }

    // Check if settings modal is open
    const settingsModalOpen = document.querySelector('[data-settings-modal="true"]') !== null;

    // If settings modal is open, only allow ESC key
    if (settingsModalOpen) {
      const key = buildKeyCombo(e);
      if (key === shortcuts.value.closeArticle) {
        // Let the modal's own ESC handler deal with it
        return;
      }
      // Block all other shortcuts when settings modal is open
      return;
    }

    // Skip if we're in an input field, textarea, or contenteditable
    const target = e.target as HTMLElement;
    const tagName = target.tagName.toLowerCase();
    const isEditable = target.isContentEditable;
    const isInput = tagName === 'input' || tagName === 'textarea' || tagName === 'select';

    const key = buildKeyCombo(e);

    // Handle article detail scrolling when article is open
    // Only in RSS content view mode, not in webpage (iframe) view mode
    if (isArticleDetailOpen() && !isWebpageViewMode()) {
      // Space key - scroll page down
      if (key === 'Space') {
        // Prevent default only if not in input field
        if (!isInput && !isEditable) {
          e.preventDefault();
          scrollArticleDetail('pageDown');
          return;
        }
      }

      // ArrowDown - scroll down
      if (key === 'ArrowDown') {
        if (!isInput && !isEditable) {
          e.preventDefault();
          scrollArticleDetail('down');
          return;
        }
      }

      // ArrowUp - scroll up
      if (key === 'ArrowUp') {
        if (!isInput && !isEditable) {
          e.preventDefault();
          scrollArticleDetail('up');
          return;
        }
      }

      // Shift+Space - scroll page up
      if (key === 'Shift+Space') {
        if (!isInput && !isEditable) {
          e.preventDefault();
          scrollArticleDetail('pageUp');
          return;
        }
      }
    }

    // Check for escape key to close modals first (always allow, even when shortcuts disabled)
    if (key === shortcuts.value.closeArticle) {
      // Check if the find in page search input is focused
      const findInputFocused = document.activeElement?.classList.contains('find-input');

      // If find input is focused, don't handle ESC here - let FindInPage component handle it
      if (findInputFocused) {
        return;
      }

      // Check if there are any open modals
      const hasOpenModal = document.querySelector('[data-modal-open="true"]') !== null;

      if (!hasOpenModal) {
        // No modals open, handle article close
        if (store.currentArticleId) {
          store.currentArticleId = null;
          e.preventDefault();
        }
      }
      // If modals are open, let them handle ESC themselves
      return;
    }

    // Skip shortcuts if in input field (except escape)
    if (isInput || isEditable) {
      return;
    }

    // Match the key combination to a shortcut action
    const action = Object.entries(shortcuts.value).find(([, shortcut]) => shortcut === key)?.[0];

    if (!action) return;

    e.preventDefault();

    // Execute the action
    switch (action) {
      case 'nextArticle':
        navigateArticle(1);
        break;
      case 'previousArticle':
        navigateArticle(-1);
        break;
      case 'nextArticleArrow':
        navigateArticle(1);
        break;
      case 'previousArticleArrow':
        navigateArticle(-1);
        break;
      case 'openArticle':
        if (store.articles.length > 0 && !store.currentArticleId) {
          selectArticleByIndex(0);
        }
        break;
      case 'toggleReadStatus':
        toggleCurrentArticleRead();
        break;
      case 'toggleFavoriteStatus':
        toggleCurrentArticleFavorite();
        break;
      case 'toggleReadLaterStatus':
        toggleCurrentArticleReadLater();
        break;
      case 'openInBrowser':
        openCurrentArticleInBrowser();
        break;
      case 'toggleContentView':
        window.dispatchEvent(new CustomEvent('toggle-content-view'));
        break;
      case 'refreshFeeds':
        store.refreshFeeds();
        break;
      case 'markAllRead':
        callbacks.onMarkAllRead();
        break;
      case 'openSettings':
        callbacks.onOpenSettings();
        break;
      case 'addFeed':
        callbacks.onAddFeed();
        break;
      case 'focusSearch':
        focusSearchInput();
        break;
      case 'toggleFilter':
        window.dispatchEvent(new CustomEvent('toggle-filter'));
        break;
      case 'toggleUnreadFilter':
        toggleListFilter('unread');
        break;
      case 'toggleFavoritesFilter':
        toggleListFilter('favorites');
        break;
      case 'toggleReadLaterFilter':
        toggleListFilter('readLater');
        break;
      case 'goToAllArticles':
        store.setFilter('all');
        break;
      case 'goToUnread':
        store.setFilter('unread');
        break;
      case 'goToFavorites':
        store.setFilter('favorites');
        break;
      case 'goToReadLater':
        store.setFilter('readLater');
        break;
    }
  }

  // Handle shortcuts changed event
  function handleShortcutsChanged(e: Event): void {
    const customEvent = e as CustomEvent;
    if (customEvent.detail && customEvent.detail.shortcuts) {
      shortcuts.value = { ...shortcuts.value, ...customEvent.detail.shortcuts };
    }
  }

  // Handle shortcuts enabled changed event
  function handleShortcutsEnabledChanged(e: Event): void {
    const customEvent = e as CustomEvent;
    if (customEvent.detail && typeof customEvent.detail.enabled === 'boolean') {
      shortcutsEnabled.value = customEvent.detail.enabled;
    }
  }

  // Initialize shortcuts enabled state from settings
  function initializeShortcutsEnabled(): void {
    // Note: store.settings is not available in the store
    // The shortcuts_enabled state is initialized via the shortcuts-enabled-changed event
    // Default is true (enabled)
  }

  // Lifecycle
  onMounted(() => {
    initializeShortcutsEnabled();
    window.addEventListener('keydown', handleKeyboardShortcut);
    window.addEventListener('shortcuts-changed', handleShortcutsChanged);
    window.addEventListener('shortcuts-enabled-changed', handleShortcutsEnabledChanged);
  });

  onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleKeyboardShortcut);
    window.removeEventListener('shortcuts-changed', handleShortcutsChanged);
    window.removeEventListener('shortcuts-enabled-changed', handleShortcutsEnabledChanged);
  });

  return {
    shortcuts,
  };
}
