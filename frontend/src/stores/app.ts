import { defineStore } from 'pinia';
import { ref, computed, type Ref } from 'vue';
import type { Article, Feed, Tag, UnreadCounts, RefreshProgress } from '@/types/models';
import type { FilterCondition } from '@/types/filter';
import { useSettings } from '@/composables/core/useSettings';

export type Filter = 'all' | 'unread' | 'favorites' | 'readLater' | 'imageGallery' | '';
export type ThemePreference = 'light' | 'dark' | 'auto';
export type Theme = 'light' | 'dark';

// Temporary selection state for feed drawer selections
export interface TempSelection {
  feedId: number | null;
  category: string | null;
}

export interface AppState {
  articles: Ref<Article[]>;
  feeds: Ref<Feed[]>;
  unreadCounts: Ref<UnreadCounts>;
  currentFilter: Ref<Filter>;
  currentFeedId: Ref<number | null>;
  currentCategory: Ref<string | null>;
  currentArticleId: Ref<number | null>;
  tempSelection: Ref<TempSelection>;
  isLoading: Ref<boolean>;
  page: Ref<number>;
  hasMore: Ref<boolean>;
  searchQuery: Ref<string>;
  themePreference: Ref<ThemePreference>;
  theme: Ref<Theme>;
  refreshProgress: Ref<RefreshProgress>;
  showOnlyUnread: Ref<boolean>;
  activeFilters: Ref<FilterCondition[]>;
  filteredArticlesFromServer: Ref<Article[]>;
  isFilterLoading: Ref<boolean>;
}

export interface AppActions {
  setFilter: (filter: Filter) => void;
  setFeed: (feedId: number) => void;
  setCategory: (category: string) => void;
  fetchArticles: (append?: boolean) => Promise<void>;
  loadMore: () => Promise<void>;
  fetchFeeds: () => Promise<void>;
  fetchUnreadCounts: () => Promise<void>;
  markAllAsRead: (feedId?: number) => Promise<void>;
  updateArticleSummary: (articleId: number, summary: string) => void;
  toggleTheme: () => void;
  setTheme: (preference: ThemePreference) => void;
  applyTheme: () => void;
  initTheme: () => void;
  refreshFeeds: () => Promise<void>;
  pollProgress: () => void;
  checkForAppUpdates: () => Promise<void>;
  startAutoRefresh: (minutes: number) => void;
  toggleShowOnlyUnread: () => void;
  setActiveFilters: (filters: FilterCondition[]) => void;
}

export const useAppStore = defineStore('app', () => {
  // Get settings composable once at store initialization
  const { settings: settingsRef } = useSettings();

  // State
  const articles = ref<Article[]>([]);
  const feeds = ref<Feed[]>([]);
  // Feed map for O(1) lookups - computed from feeds array
  const feedMap = computed(() => {
    const map = new Map<number, Feed>();
    feeds.value.forEach((feed) => map.set(feed.id, feed));
    return map;
  });
  const tags = ref<Tag[]>([]);
  // Tag map for O(1) lookups - computed from tags array
  const tagMap = computed(() => {
    const map = new Map<number, Tag>();
    tags.value.forEach((tag) => map.set(tag.id, tag));
    return map;
  });
  const unreadCounts = ref<UnreadCounts>({
    total: 0,
    feedCounts: {},
  });
  const currentFilter = ref<Filter>('all');
  const currentFeedId = ref<number | null>(null);
  const currentCategory = ref<string | null>(null);
  const currentArticleId = ref<number | null>(null);
  const tempSelection = ref<TempSelection>({ feedId: null, category: null });
  const isLoading = ref<boolean>(false);
  const page = ref<number>(1);
  const hasMore = ref<boolean>(true);
  const searchQuery = ref<string>('');
  const themePreference = ref<ThemePreference>(
    (localStorage.getItem('themePreference') as ThemePreference) || 'auto'
  );
  const theme = ref<Theme>('light');
  const showOnlyUnread = ref<boolean>(localStorage.getItem('showOnlyUnread') === 'true');
  const activeFilters = ref<FilterCondition[]>([]);
  const filteredArticlesFromServer = ref<Article[]>([]);
  const isFilterLoading = ref(false);

  // Article view mode preferences (persisted across component mounts)
  const articleViewModePreferences = ref<Map<number, 'original' | 'rendered'>>(new Map());

  // Refresh progress
  const refreshProgress = ref<RefreshProgress>({ isRunning: false });
  let refreshInterval: ReturnType<typeof setInterval> | null = null;

  // Actions - Article Management
  async function setFilter(filter: Filter): Promise<void> {
    currentFilter.value = filter;
    currentFeedId.value = null;
    currentCategory.value = null;
    tempSelection.value = { feedId: null, category: null };
    // Refresh filter counts to ensure sidebar shows correct feeds
    await fetchFilterCounts();
    // Clear and reset will be handled by fetchArticles
    fetchArticles();
  }

  function setFeed(feedId: number): void {
    // Check if this feed is an image mode feed
    const feed = feeds.value.find((f) => f.id === feedId);
    if (feed?.is_image_mode) {
      // For image mode feeds, switch filter to image gallery
      currentFilter.value = 'imageGallery';
      currentFeedId.value = feedId;
      currentCategory.value = null;
      tempSelection.value = { feedId, category: null };
      // Clear and reset will be handled by fetchArticles
    } else {
      // For regular feeds, keep currentFilter and set tempSelection
      currentFeedId.value = feedId;
      currentCategory.value = null;
      tempSelection.value = { feedId, category: null };
      fetchArticles();
    }
  }

  function setCategory(category: string): void {
    // Check if this category contains only image mode feeds
    const categoryFeeds = feeds.value.filter((f) => {
      // Handle uncategorized category (empty string)
      if (category === '') {
        return !f.category || f.category === '';
      }

      // Handle nested categories by checking if the feed's category starts with the selected path
      // For example, if category is "Tech", it should match "Tech", "Tech/AI", "Tech/AI/ML", etc.
      const feedCategory = f.category || '';
      return feedCategory === category || feedCategory.startsWith(category + '/');
    });

    const allImageMode = categoryFeeds.length > 0 && categoryFeeds.every((f) => f.is_image_mode);

    // If all feeds in this category are image mode, switch to image gallery filter
    if (allImageMode) {
      currentFilter.value = 'imageGallery';
      currentFeedId.value = null;
      currentCategory.value = category;
      tempSelection.value = { feedId: null, category };
      // Don't call fetchArticles here - ImageGalleryView will handle fetching
    } else {
      // For regular categories, keep currentFilter and set tempSelection
      currentFeedId.value = null;
      currentCategory.value = category;
      tempSelection.value = { feedId: null, category };
      fetchArticles();
    }
  }

  async function fetchArticles(append: boolean = false): Promise<void> {
    if (isLoading.value) return;

    // If not appending, reset to page 1 and clear articles
    if (!append) {
      page.value = 1;
      articles.value = [];
      hasMore.value = true;
    }

    isLoading.value = true;
    const limit = 50;

    let url = `/api/articles?page=${page.value}&limit=${limit}`;
    if (currentFilter.value) url += `&filter=${currentFilter.value}`;
    if (currentFeedId.value) url += `&feed_id=${currentFeedId.value}`;
    if (currentCategory.value !== null)
      url += `&category=${encodeURIComponent(currentCategory.value)}`;

    try {
      const res = await fetch(url);
      const data: Article[] = (await res.json()) || [];

      if (data.length < limit) {
        hasMore.value = false;
      }

      if (append) {
        articles.value = [...articles.value, ...data];
      } else {
        articles.value = data;
      }
    } catch {
      // Error handled silently
    } finally {
      isLoading.value = false;
    }
  }

  async function loadMore(): Promise<void> {
    if (hasMore.value && !isLoading.value) {
      page.value++;
      await fetchArticles(true);
    }
  }

  async function fetchFeeds(): Promise<void> {
    try {
      const res = await fetch('/api/feeds');

      const text = await res.text();

      let data;
      try {
        data = JSON.parse(text) || [];
      } catch (e) {
        console.error('[App Store] JSON parse error:', e);
        console.error('[App Store] Response text (first 500 chars):', text.substring(0, 500));
        throw e;
      }

      feeds.value = data;

      // Fetch unread counts and filter counts after fetching feeds
      await fetchUnreadCounts();
      await fetchFilterCounts();
      // Fetch tags after fetching feeds
      await fetchTags();
    } catch (e) {
      console.error('[App Store] Fetch feeds error:', e);
      feeds.value = [];
    }
  }

  async function fetchTags(): Promise<void> {
    try {
      const res = await fetch('/api/tags');
      const data = await res.json();
      tags.value = data || [];
    } catch (e) {
      console.error('[App Store] Fetch tags error:', e);
      tags.value = [];
    }
  }

  async function fetchUnreadCounts(): Promise<void> {
    try {
      const res = await fetch('/api/articles/unread-counts');
      const data = await res.json();
      unreadCounts.value = {
        total: data.total || 0,
        feedCounts: data.feed_counts || {},
      };
    } catch {
      unreadCounts.value = { total: 0, feedCounts: {} };
    }
  }

  // Filter-specific counts for sidebar filtering
  const filterCounts = ref<Record<string, Record<number | string, number>>>({
    unread: {},
    favorites: {},
    favorites_unread: {},
    read_later: {},
    read_later_unread: {},
    images: {},
    images_unread: {},
  });

  async function fetchFilterCounts(): Promise<void> {
    try {
      const res = await fetch('/api/articles/filter-counts');
      const data = await res.json();
      filterCounts.value = {
        unread: data.unread || {},
        favorites: data.favorites || {},
        favorites_unread: data.favorites_unread || {},
        read_later: data.read_later || {},
        read_later_unread: data.read_later_unread || {},
        images: data.images || {},
        images_unread: data.images_unread || {},
      };
    } catch (e) {
      console.error('[App Store] Fetch filter counts error:', e);
      filterCounts.value = {
        unread: {},
        favorites: {},
        favorites_unread: {},
        read_later: {},
        read_later_unread: {},
        images: {},
        images_unread: {},
      };
    }
  }

  async function markAllAsRead(feedId?: number, category?: string): Promise<void> {
    try {
      const params = new URLSearchParams();
      if (feedId) params.append('feed_id', String(feedId));
      if (category) params.append('category', category);

      const url = params.toString()
        ? `/api/articles/mark-all-read?${params.toString()}`
        : '/api/articles/mark-all-read';
      await fetch(url, { method: 'POST' });
      // Refresh articles and unread counts
      await fetchArticles();
      await fetchUnreadCounts();
    } catch {
      // Error handled silently
    }
  }

  // Update article summary in store
  function updateArticleSummary(articleId: number, summary: string): void {
    const articleIndex = articles.value.findIndex((a) => a.id === articleId);
    if (articleIndex !== -1) {
      articles.value[articleIndex] = {
        ...articles.value[articleIndex],
        summary,
      };
    }
  }

  // Theme Management
  function toggleTheme(): void {
    // Cycle through: light -> dark -> auto -> light
    if (themePreference.value === 'light') {
      themePreference.value = 'dark';
    } else if (themePreference.value === 'dark') {
      themePreference.value = 'auto';
    } else {
      themePreference.value = 'light';
    }
    localStorage.setItem('themePreference', themePreference.value);
    applyTheme();
  }

  function setTheme(preference: ThemePreference): void {
    themePreference.value = preference;
    localStorage.setItem('themePreference', preference);
    applyTheme();
  }

  function applyTheme(): void {
    let actualTheme: Theme = themePreference.value as Theme;

    // If auto, detect system preference
    if (themePreference.value === 'auto') {
      actualTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }

    theme.value = actualTheme;

    // Apply to both html and body for consistency
    const htmlElement = document.documentElement;
    if (actualTheme === 'dark') {
      htmlElement.classList.add('dark-mode');
      document.body.classList.add('dark-mode');
    } else {
      htmlElement.classList.remove('dark-mode');
      document.body.classList.remove('dark-mode');
    }
  }

  function initTheme(): void {
    // Listen for system theme changes
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQuery.addEventListener('change', () => {
      if (themePreference.value === 'auto') {
        applyTheme();
      }
    });

    // Apply initial theme
    applyTheme();
  }

  // Auto Refresh
  async function refreshFeeds(): Promise<void> {
    refreshProgress.value.isRunning = true;
    try {
      // First, trigger standard refresh
      const refreshRes = await fetch('/api/refresh', { method: 'POST' });
      if (!refreshRes.ok) {
        throw new Error(`Refresh API returned ${refreshRes.status}: ${refreshRes.statusText}`);
      }
      // Verify the response is valid JSON by consuming it
      try {
        await refreshRes.json();
      } catch (e) {
        console.error('Invalid JSON response from /api/refresh:', e);
        throw new Error(`Invalid JSON response from refresh API: ${e}`);
      }

      // Also trigger FreshRSS sync if enabled
      if (settingsRef.value.freshrss_enabled === true) {
        try {
          await fetch('/api/freshrss/sync', { method: 'POST' });
        } catch (e) {
          // If FreshRSS sync fails, it's okay - just log it
          console.log('FreshRSS sync failed:', e);
        }
      }

      // Wait a moment to check if refresh is actually running
      await new Promise((resolve) => setTimeout(resolve, 200));

      // Check progress to see if there are actually any tasks
      const progressRes = await fetch('/api/progress');
      if (!progressRes.ok) {
        throw new Error(`Progress API returned ${progressRes.status}: ${progressRes.statusText}`);
      }
      const progressData = await progressRes.json();

      // If no tasks are running, mark as completed immediately
      if (!progressData.is_running) {
        refreshProgress.value.isRunning = false;

        // Still refresh feeds and articles to get any updates from FreshRSS sync
        fetchFeeds();
        fetchArticles();
        fetchUnreadCounts();

        // Notify components that settings have been updated
        window.dispatchEvent(new CustomEvent('settings-updated'));
        return;
      }

      // If tasks are running, proceed with normal progress polling
      await fetchProgressOnce();
      pollProgress();
    } catch (e) {
      console.error('Error refreshing feeds:', e);
      refreshProgress.value.isRunning = false;
    }
  }

  async function fetchProgressOnce(): Promise<void> {
    try {
      // Wait a bit for the backend to start processing
      await new Promise((resolve) => setTimeout(resolve, 100));

      const res = await fetch('/api/progress');
      if (!res.ok) {
        throw new Error(`Progress API returned ${res.status}: ${res.statusText}`);
      }
      const data = await res.json();
      console.log('Initial progress update:', data);
      refreshProgress.value = {
        ...refreshProgress.value,
        isRunning: data.is_running,
        errors: data.errors,
        pool_task_count: data.pool_task_count,
        article_click_count: data.article_click_count,
        queue_task_count: data.queue_task_count,
      };
      console.log('Initial refreshProgress:', refreshProgress.value);
    } catch (e) {
      console.error('Error fetching initial progress:', e);
    }
  }

  function pollProgress(): void {
    // Track previous pool/queue counts to detect task completion
    let previousPoolCount = 0;
    let previousQueueCount = 0;

    const interval = setInterval(async () => {
      try {
        const res = await fetch('/api/progress');
        if (!res.ok) {
          throw new Error(`Progress API returned ${res.status}: ${res.statusText}`);
        }
        const data = await res.json();
        refreshProgress.value = {
          ...refreshProgress.value, // Preserve existing pool_tasks and queue_tasks
          isRunning: data.is_running,
          errors: data.errors,
          pool_task_count: data.pool_task_count ?? 0,
          article_click_count: data.article_click_count ?? 0,
          queue_task_count: data.queue_task_count ?? 0,
        };

        // Fetch task details if refresh is running
        if (data.is_running && (data.pool_task_count > 0 || data.queue_task_count > 0)) {
          await fetchTaskDetails();
        }

        // Detect task completion and update unread counts immediately
        const currentPoolCount = data.pool_task_count ?? 0;
        const currentQueueCount = data.queue_task_count ?? 0;
        const totalTasks = currentPoolCount + currentQueueCount;
        const previousTotal = previousPoolCount + previousQueueCount;

        // If task count decreased, tasks completed - update unread counts
        if (totalTasks < previousTotal && previousTotal > 0) {
          fetchUnreadCounts();
          fetchFeeds(); // Also update feeds to refresh error marks
        }

        // Update previous counts
        previousPoolCount = currentPoolCount;
        previousQueueCount = currentQueueCount;

        if (!data.is_running) {
          clearInterval(interval);
          fetchFeeds();
          fetchArticles();
          fetchUnreadCounts();

          // Notify components that settings have been updated (e.g., last_article_update)
          // This triggers components using useSettings() to refresh their settings
          window.dispatchEvent(new CustomEvent('settings-updated'));

          // Note: We no longer show error toasts for failed feeds
          // Users can see error status in the feed list sidebar

          // Check for app updates after initial refresh completes

          checkForAppUpdates();
        }
      } catch {
        clearInterval(interval);
        refreshProgress.value.isRunning = false;
      }
    }, 500);
  }

  // FreshRSS sync status monitoring
  let freshrssPollInterval: ReturnType<typeof setInterval> | null = null;
  let lastKnownFreshRSSSyncTime: string | null = null;

  async function startFreshRSSStatusPolling(): Promise<void> {
    // Stop any existing polling
    if (freshrssPollInterval) {
      clearInterval(freshrssPollInterval);
    }

    // Check if FreshRSS is enabled
    try {
      const res = await fetch('/api/settings');
      if (!res.ok) return;
      const settings = await res.json();

      if (settings.freshrss_enabled !== 'true') {
        return; // FreshRSS not enabled, don't start polling
      }

      // Initialize last known sync time
      const statusRes = await fetch('/api/freshrss/status');
      if (statusRes.ok) {
        const statusData = await statusRes.json();
        lastKnownFreshRSSSyncTime = statusData.last_sync_time;
      }
    } catch (e) {
      console.error('[FreshRSS] Error checking status:', e);
      return;
    }

    // Start polling every 5 seconds
    freshrssPollInterval = setInterval(async () => {
      try {
        const res = await fetch('/api/freshrss/status');
        if (!res.ok) return;

        const data = await res.json();

        // Check if sync time has updated (sync completed)
        if (
          lastKnownFreshRSSSyncTime !== null &&
          data.last_sync_time !== lastKnownFreshRSSSyncTime
        ) {
          console.log('[FreshRSS] Sync completed detected, refreshing data...');
          // Refresh all data
          await fetchFeeds();
          await fetchArticles();
          await fetchUnreadCounts();
        }

        // Update known sync time
        lastKnownFreshRSSSyncTime = data.last_sync_time;
      } catch (e) {
        console.error('[FreshRSS] Error polling status:', e);
      }
    }, 5000); // Poll every 5 seconds
  }

  function stopFreshRSSStatusPolling(): void {
    if (freshrssPollInterval) {
      clearInterval(freshrssPollInterval);
      freshrssPollInterval = null;
    }
  }

  async function checkForAppUpdates(): Promise<void> {
    try {
      const res = await fetch('/api/check-updates');
      if (res.ok) {
        const data = await res.json();

        // Only proceed if there's an update available and a download URL
        if (data.has_update && data.download_url) {
          // Check if auto-update is enabled before downloading
          const { settings } = useSettings();

          console.log('[DEBUG] Update found, auto_update =', settings.value.auto_update);
          if (settings.value.auto_update) {
            console.log('[DEBUG] Auto-downloading update...');
            // Auto download and install in background
            autoDownloadAndInstall(data.download_url, data.asset_name);
          } else {
            console.log('[DEBUG] Auto-update disabled, showing notification only');
            // Just show notification that update is available
            if (window.showToast) {
              window.showToast(`Update available: v${data.latest_version}`, 'info', 5000);
            }
          }
        }
      }
    } catch {
      console.error('Auto-update check failed');
      // Silently fail - don't disrupt user experience
    }
  }

  async function autoDownloadAndInstall(downloadUrl: string, assetName?: string): Promise<void> {
    try {
      // Download the update in background
      const downloadRes = await fetch('/api/download-update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          download_url: downloadUrl,
          asset_name: assetName,
        }),
      });

      if (!downloadRes.ok) {
        console.error('Auto-download failed');
        return;
      }

      const downloadData = await downloadRes.json();
      if (!downloadData.success || !downloadData.file_path) {
        console.error('Auto-download failed: Invalid response');
        return;
      }

      // Wait a moment to ensure file is fully written
      await new Promise((resolve) => setTimeout(resolve, 500));

      // Install the update
      const installRes = await fetch('/api/install-update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          file_path: downloadData.file_path,
        }),
      });

      if (!installRes.ok) {
        console.error('Auto-install failed');
        return;
      }

      const installData = await installRes.json();
      if (installData.success && window.showToast) {
        window.showToast('Update installed. Restart to apply.', 'success');
      }
    } catch {
      console.error('Auto-update failed');
      // Silently fail - don't disrupt user experience
    }
  }

  function startAutoRefresh(minutes: number): void {
    if (refreshInterval) clearInterval(refreshInterval);
    if (minutes > 0) {
      refreshInterval = setInterval(
        () => {
          refreshFeeds();
        },
        minutes * 60 * 1000
      );
    }
  }

  function toggleShowOnlyUnread(): void {
    showOnlyUnread.value = !showOnlyUnread.value;
    localStorage.setItem('showOnlyUnread', String(showOnlyUnread.value));
  }

  function setActiveFilters(filters: FilterCondition[]): void {
    activeFilters.value = filters;
  }

  function setFilteredArticlesFromServer(articles: Article[]): void {
    filteredArticlesFromServer.value = articles;
  }

  function setIsFilterLoading(loading: boolean): void {
    isFilterLoading.value = loading;
  }

  async function fetchTaskDetails(): Promise<void> {
    try {
      const res = await fetch('/api/progress/task-details');
      if (res.ok) {
        const data = await res.json();
        refreshProgress.value = {
          ...refreshProgress.value,
          pool_tasks: data.pool_tasks,
          queue_tasks: data.queue_tasks,
        };
      }
    } catch (e) {
      console.error('Error fetching task details:', e);
    }
  }

  return {
    // State
    articles,
    feeds,
    feedMap,
    tags,
    tagMap,
    unreadCounts,
    filterCounts,
    currentFilter,
    currentFeedId,
    currentCategory,
    currentArticleId,
    tempSelection,
    isLoading,
    page,
    hasMore,
    searchQuery,
    themePreference,
    theme,
    refreshProgress,
    showOnlyUnread,
    activeFilters,
    filteredArticlesFromServer,
    isFilterLoading,
    articleViewModePreferences,

    // Actions
    setFilter,
    setFeed,
    setCategory,
    fetchArticles,
    loadMore,
    fetchFeeds,
    fetchTags,
    fetchUnreadCounts,
    fetchFilterCounts,
    markAllAsRead,
    updateArticleSummary,
    toggleTheme,
    setTheme,
    applyTheme,
    initTheme,
    refreshFeeds,
    pollProgress,
    startFreshRSSStatusPolling,
    stopFreshRSSStatusPolling,
    checkForAppUpdates,
    startAutoRefresh,
    toggleShowOnlyUnread,
    setActiveFilters,
    setFilteredArticlesFromServer,
    setIsFilterLoading,
    fetchTaskDetails,
  };
});
