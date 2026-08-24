import { computed, ref, watch, type Ref } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { openInBrowser } from '@/utils/browser';
import type { Feed } from '@/types/models';

interface TreeNode {
  _feeds: Feed[];
  _children: Record<string, TreeNode>;
  isOpen: boolean;
}

interface TreeData {
  tree: Record<string, TreeNode>;
  uncategorized: Feed[];
  categories: Set<string>;
}

export function useSidebar() {
  const store = useAppStore();
  const { t } = useI18n();

  // Load saved category state from localStorage
  const savedCategories = localStorage.getItem('openCategories');
  const openCategories: Ref<Set<string>> = ref(
    savedCategories ? new Set(JSON.parse(savedCategories)) : new Set()
  );

  const searchQuery: Ref<string> = ref('');

  // Build category tree with search filtering and filter-specific filtering
  const tree = computed<TreeData>(() => {
    const t: Record<string, TreeNode> = {};
    const uncategorized: Feed[] = [];
    const categories = new Set<string>();

    if (!store.feeds) return { tree: {}, uncategorized: [], categories };

    const query = searchQuery.value.toLowerCase().trim();

    // Determine which filter counts to use based on currentFilter
    const currentFilterType = store.currentFilter;
    const filterTypeMap: Record<string, string> = {
      unread: 'unread',
      favorites: 'favorites',
      readLater: 'read_later',
      imageGallery: 'images',
    };
    const filterKey = filterTypeMap[currentFilterType] || '';

    store.feeds.forEach((feed: Feed) => {
      const matchesSearch =
        query === '' ||
        feed.title.toLowerCase().includes(query) ||
        feed.url.toLowerCase().includes(query);

      if (!matchesSearch) return;

      // Filter by currentFilter if applicable
      if (currentFilterType && filterKey) {
        const feedCount = store.filterCounts[filterKey]?.[feed.id] || 0;
        if (feedCount === 0) return;
      }

      if (feed.category) {
        const parts = feed.category.split('/');
        let currentLevel = t;
        parts.forEach((part, index) => {
          if (!currentLevel[part]) {
            currentLevel[part] = { _feeds: [], _children: {}, isOpen: false };
          }
          if (index === parts.length - 1) {
            currentLevel[part]._feeds.push(feed);
            categories.add(feed.category);
          } else {
            currentLevel = currentLevel[part]._children;
          }
        });
      } else {
        uncategorized.push(feed);
      }
    });
    if (uncategorized.length > 0) {
      categories.add('uncategorized');
    }
    return { tree: t, uncategorized, categories };
  });

  // Compute unread counts for categories based on current filter
  const categoryUnreadCounts = computed<Record<string, number>>(() => {
    const counts: Record<string, number> = {};
    if (!store.feeds || !store.unreadCounts.feedCounts) return counts;

    // Determine which counts to use based on current filter
    let countsSource: Record<number | string, number>;
    switch (store.currentFilter) {
      case 'favorites':
        countsSource = store.filterCounts.favorites_unread;
        break;
      case 'readLater':
        countsSource = store.filterCounts.read_later_unread;
        break;
      case 'unread':
        countsSource = store.filterCounts.unread;
        break;
      case 'imageGallery':
        countsSource = store.filterCounts.images_unread;
        break;
      default:
        // For 'all' or empty filter, use regular unread counts
        countsSource = store.unreadCounts.feedCounts;
        break;
    }

    store.feeds.forEach((feed: Feed) => {
      if (feed.category) {
        const unreadCount = countsSource[feed.id] || 0;
        if (unreadCount > 0) {
          counts[feed.category] = (counts[feed.category] || 0) + unreadCount;
        }
      }
    });

    // Calculate uncategorized count
    const uncategorizedFeeds = store.feeds.filter((f) => !f.category);
    counts['uncategorized'] = uncategorizedFeeds.reduce((sum, feed) => {
      return sum + (countsSource[feed.id] || 0);
    }, 0);

    return counts;
  });

  // Compute feed unread counts based on current filter (for displaying on individual feeds)
  const feedUnreadCounts = computed<Record<number, number>>(() => {
    if (!store.feeds) return {};

    // Determine which counts to use based on current filter
    switch (store.currentFilter) {
      case 'favorites':
        return store.filterCounts.favorites_unread;
      case 'readLater':
        return store.filterCounts.read_later_unread;
      case 'unread':
        return store.filterCounts.unread;
      case 'imageGallery':
        return store.filterCounts.images_unread;
      default:
        // For 'all' or empty filter, use regular unread counts
        return store.unreadCounts.feedCounts;
    }
  });

  // Auto-expand new categories only if no saved state exists
  watch(
    () => tree.value.categories,
    (newCategories) => {
      if (newCategories) {
        const hasSavedState = localStorage.getItem('openCategories') !== null;
        newCategories.forEach((cat) => {
          // Always auto-expand 'uncategorized' category
          if (cat === 'uncategorized' && !openCategories.value.has(cat)) {
            openCategories.value.add(cat);
            return;
          }
          // Only auto-expand if this is a new category and no saved state exists
          if (!openCategories.value.has(cat) && !hasSavedState) {
            openCategories.value.add(cat);
          }
        });

        // Also auto-expand parent categories for multi-level
        // For example, if "Tech/Blogs" exists, also expand "Tech"
        const parentCategories = new Set<string>();
        newCategories.forEach((cat) => {
          const parts = cat.split('/');
          for (let i = 1; i < parts.length; i++) {
            const parentPath = parts.slice(0, i).join('/');
            parentCategories.add(parentPath);
          }
        });

        parentCategories.forEach((parentCat) => {
          if (!openCategories.value.has(parentCat) && !hasSavedState) {
            openCategories.value.add(parentCat);
          }
        });
      }
    },
    { immediate: true }
  );

  function toggleCategory(path: string): void {
    if (openCategories.value.has(path)) {
      openCategories.value.delete(path);
    } else {
      openCategories.value.add(path);
    }
    // Save to localStorage
    localStorage.setItem('openCategories', JSON.stringify([...openCategories.value]));
  }

  function isCategoryOpen(path: string): boolean {
    return openCategories.value.has(path);
  }

  // Feed actions
  async function handleFeedAction(action: string, feed: Feed): Promise<void> {
    if (action === 'markAllRead') {
      await store.markAllAsRead(feed.id);
      window.showToast(t('article.action.markedAllAsRead'), 'success');
    } else if (action === 'refreshFeed') {
      await fetch(`/api/feeds/refresh?id=${feed.id}`, { method: 'POST' });
      window.showToast(t('modal.feed.feedRefreshStarted'), 'success');
      // Start polling for progress as the backend is now fetching articles for this feed
      store.pollProgress();
    } else if (action === 'syncFeed') {
      // Sync individual FreshRSS feed
      await fetch(`/api/freshrss/sync-feed?stream_id=${feed.freshrss_stream_id}`, {
        method: 'POST',
      });
      window.showToast(t('modal.feed.syncFeedStarted'), 'success');
      // Start polling for progress
      store.pollProgress();
    } else if (action === 'delete') {
      const confirmed = await window.showConfirm({
        title: t('modal.feed.unsubscribeTitle'),
        message: t('modal.feed.unsubscribeMessage', { name: feed.title }),
        confirmText: t('common.action.unsubscribe'),
        cancelText: t('common.action.cancel'),
        isDanger: true,
      });
      if (confirmed) {
        await fetch(`/api/feeds/delete?id=${feed.id}`, { method: 'POST' });
        store.fetchFeeds();
        window.showToast(t('modal.feed.unsubscribedSuccess'), 'success');
      }
    } else if (action === 'edit') {
      window.dispatchEvent(new CustomEvent('show-edit-feed', { detail: feed }));
    } else if (action === 'openWebsite') {
      // Handle RSSHub URLs - need to transform rsshub:// to full URL
      let urlToOpen = feed.website_url || feed.url;
      if (urlToOpen.startsWith('rsshub://')) {
        try {
          const response = await fetch('/api/rsshub/transform-url', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({ url: urlToOpen }),
          });
          if (response.ok) {
            const data = await response.json();
            if (data.url && (data.url.startsWith('http://') || data.url.startsWith('https://'))) {
              urlToOpen = data.url;
            } else {
              // Invalid transformed URL
              window.showToast(
                t('common.errors.failedToTransformRSSHubURL') || 'Failed to transform RSSHub URL',
                'error'
              );
              return;
            }
          } else {
            // Transformation failed - try to get error message from response
            let errorMessage =
              t('common.errors.failedToTransformRSSHubURL') || 'Failed to transform RSSHub URL';
            try {
              const errorText = await response.text();
              if (errorText) {
                errorMessage = errorText;
              }
            } catch (e) {
              // Ignore error reading response
            }
            window.showToast(errorMessage, 'error');
            return;
          }
        } catch (error) {
          window.showToast(
            t('common.errors.failedToTransformRSSHubURL') || 'Failed to transform RSSHub URL',
            'error'
          );
          return;
        }
      }
      // Only open if URL is http/https (not rsshub://)
      if (urlToOpen.startsWith('http://') || urlToOpen.startsWith('https://')) {
        openInBrowser(urlToOpen);
      } else {
        window.showToast(t('common.errors.invalidURLScheme') || 'Invalid URL scheme', 'error');
      }
    } else if (action === 'discover') {
      window.dispatchEvent(new CustomEvent('show-discover-blogs', { detail: feed }));
    }
  }

  function onFeedContextMenu(e: MouseEvent, feed: Feed): void {
    e.preventDefault();
    e.stopPropagation();

    // Build menu items dynamically based on whether this is a FreshRSS feed
    const items: Array<{
      label?: string;
      action?: string;
      icon?: string;
      separator?: boolean;
      danger?: boolean;
    }> = [];

    // For FreshRSS feeds, show "Sync Feed" instead of "Refresh Feed"
    if (feed.is_freshrss_source) {
      items.push({
        label: t('modal.feed.syncFeed'),
        action: 'syncFeed',
        icon: 'PhArrowsClockwise',
      });
    } else {
      items.push({
        label: t('article.action.refreshFeed'),
        action: 'refreshFeed',
        icon: 'PhArrowsClockwise',
      });
    }

    items.push({
      label: t('article.action.markAllAsReadFeed'),
      action: 'markAllRead',
      icon: 'PhCheckCircle',
    });
    items.push({ separator: true });
    items.push({ label: t('common.action.openWebsite'), action: 'openWebsite', icon: 'PhGlobe' });

    // Only add discover for non-FreshRSS feeds
    if (!feed.is_freshrss_source) {
      items.push({
        label: t('modal.discovery.discoverFeeds'),
        action: 'discover',
        icon: 'PhBinoculars',
      });
    }

    // Only add edit and delete options for non-FreshRSS feeds
    if (!feed.is_freshrss_source) {
      items.push({ separator: true });
      items.push({ label: t('modal.feed.editSubscription'), action: 'edit', icon: 'PhPencil' });
      items.push({
        label: t('common.action.unsubscribe'),
        action: 'delete',
        icon: 'PhTrash',
        danger: true,
      });
    }

    window.dispatchEvent(
      new CustomEvent('open-context-menu', {
        detail: {
          x: e.clientX,
          y: e.clientY,
          items,
          data: feed,
          callback: handleFeedAction,
        },
      })
    );
  }

  // Category actions
  async function handleCategoryAction(action: string, categoryName: string): Promise<void> {
    if (action === 'markAllRead') {
      // Use the category parameter for the API call
      const category = categoryName === 'uncategorized' ? '' : categoryName;
      await fetch(`/api/articles/mark-all-read?category=${encodeURIComponent(category)}`, {
        method: 'POST',
      });
      store.fetchUnreadCounts();
      window.showToast(t('article.action.markedAllAsRead'), 'success');
    } else if (action === 'rename') {
      const newName = await window.showInput({
        title: t('modal.feed.renameCategory'),
        message: t('modal.feed.enterCategoryName'),
        defaultValue: categoryName,
        confirmText: t('common.action.confirm'),
        cancelText: t('common.action.cancel'),
      });
      if (newName && newName !== categoryName) {
        const feedsToUpdate = store.feeds.filter(
          (f) => f.category === categoryName || f.category.startsWith(categoryName + '/')
        );

        const promises = feedsToUpdate.map((feed) => {
          let newCategory = feed.category;
          if (feed.category === categoryName) {
            newCategory = newName;
          } else if (feed.category.startsWith(categoryName + '/')) {
            newCategory = newName + feed.category.substring(categoryName.length);
          }

          return fetch('/api/feeds/update', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              id: feed.id,
              title: feed.title,
              url: feed.url,
              category: newCategory,
            }),
          });
        });

        await Promise.all(promises);
        store.fetchFeeds();
      }
    }
  }

  function onCategoryContextMenu(e: MouseEvent, categoryName: string): void {
    e.preventDefault();
    e.stopPropagation();

    const items: Array<{ label?: string; action?: string; icon?: string; separator?: boolean }> = [
      {
        label: t('article.action.markAllAsReadFeed'),
        action: 'markAllRead',
        icon: 'ph-check-circle',
      },
    ];

    if (categoryName !== 'uncategorized') {
      items.push({ separator: true });
      items.push({ label: t('modal.feed.renameCategory'), action: 'rename', icon: 'ph-pencil' });
    }

    window.dispatchEvent(
      new CustomEvent('open-context-menu', {
        detail: {
          x: e.clientX,
          y: e.clientY,
          items: items,
          data: categoryName,
          callback: handleCategoryAction,
        },
      })
    );
  }

  // Expand category containing a specific feed
  function expandCategoryForFeed(feedId: number): void {
    const feed = store.feeds?.find((f) => f.id === feedId);
    if (feed?.category) {
      // Expand the feed's category
      if (!openCategories.value.has(feed.category)) {
        openCategories.value.add(feed.category);
      }
      // Also expand all parent categories
      const parts = feed.category.split('/');
      for (let i = 1; i < parts.length; i++) {
        const parentPath = parts.slice(0, i).join('/');
        if (!openCategories.value.has(parentPath)) {
          openCategories.value.add(parentPath);
        }
      }
      // Trigger reactivity by creating a new Set
      const newSet = new Set(openCategories.value);
      openCategories.value = newSet;
      // Save to localStorage
      localStorage.setItem('openCategories', JSON.stringify([...openCategories.value]));
      // Dispatch event to notify other components
      window.dispatchEvent(new CustomEvent('categories-expanded'));
    } else if (feed && !feed.category) {
      // Expand uncategorized category
      if (!openCategories.value.has('uncategorized')) {
        openCategories.value.add('uncategorized');
        // Trigger reactivity by creating a new Set
        const newSet = new Set(openCategories.value);
        openCategories.value = newSet;
        // Save to localStorage
        localStorage.setItem('openCategories', JSON.stringify([...openCategories.value]));
        // Dispatch event to notify other components
        window.dispatchEvent(new CustomEvent('categories-expanded'));
      }
    }
  }

  return {
    tree,
    categoryUnreadCounts,
    feedUnreadCounts,
    openCategories,
    searchQuery,
    toggleCategory,
    isCategoryOpen,
    onFeedContextMenu,
    onCategoryContextMenu,
    expandCategoryForFeed,
  };
}
