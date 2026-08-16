<script setup lang="ts">
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick, type Ref } from 'vue';
import {
  PhArrowClockwise,
  PhList,
  PhSpinner,
  PhFunnel,
  PhTrash,
  PhCheckCircle,
  PhEye,
  PhEyeSlash,
  PhCircle,
  PhClock,
  PhLightning,
} from '@phosphor-icons/vue';
import ArticleFilterModal from '../modals/filter/ArticleFilterModal.vue';
import ArticleItem from './ArticleItem.vue';
import ArticleCardItem from './ArticleCardItem.vue';
import ArticleDetailModal from './ArticleDetailModal.vue';
import AISearchBar from './AISearchBar.vue';
import { useArticleTranslation } from '@/composables/article/useArticleTranslation';
import { useArticleFilter } from '@/composables/article/useArticleFilter';
import { useArticleActions } from '@/composables/article/useArticleActions';
import { useShowPreviewImages } from '@/composables/ui/useShowPreviewImages';
import { useSettings } from '@/composables/core/useSettings';
import { parseSettingsData } from '@/composables/core/useSettings.generated';
import { openInBrowser } from '@/utils/browser';
import { proxyImagesInHtml, isMediaCacheEnabled } from '@/utils/mediaProxy';
import type { Article } from '@/types/models';

const store = useAppStore();
const { t } = useI18n();
const { settings } = useSettings();

const listRef: Ref<HTMLDivElement | null> = ref(null);
const defaultViewMode = ref<'original' | 'rendered' | 'external'>('original');
const showFilterModal = ref(false);
const isRefreshing = ref(false);
const savedScrollTop = ref(0);
const showRefreshTooltip = ref(false);
// Track articles that should be temporarily kept in list even if read
const temporarilyKeepArticles = ref<Set<number>>(new Set());
// Flag to control when scroll position should be restored
const shouldRestoreScroll = ref(false);

// Card mode modal state
const showCardModal = ref(false);
const cardModalArticle = ref<Article | null>(null);
const cardModalContent = ref('');
const isCardModalLoading = ref(false);

// Track if user has scrolled to bottom
const hasScrolledToBottom = ref(false);

// Layout mode computed
const layoutMode = computed(() => settings.value.layout_mode || 'normal');
const isCardMode = computed(() => layoutMode.value === 'card');

interface Props {
  isSidebarOpen?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  toggleSidebar: [];
}>();

// Use composables
const {
  translationSettings,
  loadTranslationSettings,
  setupIntersectionObserver,
  observeArticle,
  handleTranslationSettingsChange,
  cleanup: cleanupTranslation,
} = useArticleTranslation();

const { activeFilters, resetFilterState, fetchFilteredArticles, loadMoreFilteredArticles } =
  useArticleFilter();

// AI Search state
const aiSearchResults = ref<Article[]>([]);
const isAISearchActive = ref(false);

// AI Search enabled from settings
const isAISearchEnabled = computed(() => settings.value.ai_search_enabled);

// Use store's filtered articles and loading state directly
const filteredArticlesFromServer = computed(() => store.filteredArticlesFromServer);
const isFilterLoading = computed(() => store.isFilterLoading);

// Computed filtered articles - optimized to avoid excessive recomputation
const filteredArticles = computed(() => {
  const usesClientSideUnreadFilter =
    activeFilters.value.length > 0 || (isAISearchActive.value && aiSearchResults.value.length > 0);

  // If AI search is active, use AI search results
  if (isAISearchActive.value && aiSearchResults.value.length > 0) {
    let articles = aiSearchResults.value;
    if (store.showOnlyUnread) {
      articles = articles.filter((article) => !article.is_read);
    }
    return articles;
  }

  let articles = activeFilters.value.length > 0 ? filteredArticlesFromServer.value : store.articles;

  // Normal article pages apply showOnlyUnread on the server so pagination does
  // not return a page of read articles that then disappears client-side.
  // Using a simpler filter that avoids Set.has() calls when possible
  if (
    usesClientSideUnreadFilter &&
    store.showOnlyUnread &&
    temporarilyKeepArticles.value.size > 0
  ) {
    articles = articles.filter(
      (article) => !article.is_read || temporarilyKeepArticles.value.has(article.id)
    );
  } else if (usesClientSideUnreadFilter && store.showOnlyUnread) {
    // Fast path when no temporarily kept articles
    articles = articles.filter((article) => !article.is_read);
  }

  return articles;
});

// AI Search handlers
function handleAISearchResults(articles: Article[]) {
  aiSearchResults.value = articles;
  isAISearchActive.value = true;
}

function handleAISearchClear() {
  aiSearchResults.value = [];
  isAISearchActive.value = false;
}

const { showArticleContextMenu } = useArticleActions(t, defaultViewMode, async () => {
  await store.fetchUnreadCounts();
  await store.fetchFilterCounts();
});

// Virtual rendering: only render visible articles + buffer
const visibleArticles = computed(() => {
  // For now, render all articles but could be optimized for virtual scrolling
  // Keeping it simple to avoid complexity
  return filteredArticles.value;
});

// Helper to truncate text to max length
function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength - 1) + '…';
}

// Dynamic title based on current filter and temporary selection
const articleListTitle = computed(() => {
  // If there's a temporary selection from feed drawer, show feed/category name with filter
  if (store.tempSelection.feedId) {
    const feed = store.feeds?.find((f) => f.id === store.tempSelection.feedId);
    const feedName = feed?.title || '';
    const filterText = getFilterText();

    // Truncate feed name if it's too long (leave room for " - filterText")
    const maxFeedNameLength = filterText ? 40 : 50;
    const truncatedFeedName = truncateText(feedName, maxFeedNameLength);

    return filterText ? `${truncatedFeedName} - ${filterText}` : truncatedFeedName;
  }

  if (store.tempSelection.category) {
    const categoryName = store.tempSelection.category;
    const filterText = getFilterText();

    // Truncate category name if it's too long
    const maxCategoryLength = filterText ? 40 : 50;
    const truncatedCategory = truncateText(categoryName, maxCategoryLength);

    return filterText ? `${truncatedCategory} - ${filterText}` : truncatedCategory;
  }

  // No temporary selection, show filter only
  return getFilterText() || t('sidebar.feedList.articles');
});

// Helper to get filter text
function getFilterText(): string {
  switch (store.currentFilter) {
    case 'all':
      return t('sidebar.activity.allArticles');
    case 'unread':
      return t('sidebar.activity.unreadArticles');
    case 'favorites':
      return t('sidebar.activity.favorites');
    case 'readLater':
      return t('sidebar.activity.readLater');
    case 'imageGallery':
      return t('sidebar.activity.imageGallery');
    default:
      return '';
  }
}

// Initialize show preview images setting
const { initialize: initializeShowPreviewImages } = useShowPreviewImages();

// Load settings and setup
onMounted(async () => {
  await loadTranslationSettings();
  await initializeShowPreviewImages();

  try {
    const res = await fetch('/api/settings');
    const data = await res.json();
    defaultViewMode.value = data.default_view_mode || 'original';

    // Parse and apply settings including layout_mode
    settings.value = parseSettingsData(data);
    console.log('ArticleList settings loaded on mount:', settings.value.layout_mode);

    // Set up intersection observer for auto-translation
    if (translationSettings.value.enabled && listRef.value) {
      setupIntersectionObserver(listRef.value, store.articles);
    }
  } catch (e) {
    console.error('Error loading settings:', e);
  }

  // Listen for translation settings changes
  window.addEventListener(
    'translation-settings-changed',
    onTranslationSettingsChanged as EventListener
  );
  // Listen for default view mode changes
  window.addEventListener('default-view-mode-changed', onDefaultViewModeChanged as EventListener);
  // Listen for show preview images changes
  window.addEventListener(
    'show-preview-images-changed',
    onShowPreviewImagesChanged as EventListener
  );
  // Listen for layout mode changes
  window.addEventListener('layout-mode-changed', onLayoutModeChanged as EventListener);
  // Listen for settings loaded event (from App.vue on startup)
  window.addEventListener('settings-loaded', onSettingsLoaded as EventListener);
  // Listen for refresh articles events
  window.addEventListener('refresh-articles', onRefreshArticles);
  // Listen for toggle filter events (from keyboard shortcut)
  window.addEventListener('toggle-filter', onToggleFilter);
  // Listen for mark-all-read events (from keyboard shortcut)
  window.addEventListener('mark-all-as-read', onMarkAllAsRead);
});

// Watch for articles array length changes (list content changes)
watch(
  () => store.articles.length,
  async () => {
    // Only restore scroll position when explicitly needed (e.g., during refresh)
    if (shouldRestoreScroll.value && listRef.value) {
      const currentScroll = listRef.value.scrollTop;
      await nextTick();
      listRef.value.scrollTop = currentScroll;
      shouldRestoreScroll.value = false;
    }
  }
);

// Watch for articles array changes to re-observe new articles for translation
// Use shallow watch to avoid triggering on property changes (like is_read)
watch(
  () => store.articles,
  async () => {
    // Re-setup observer to observe newly added articles
    if (translationSettings.value.enabled && listRef.value) {
      await nextTick();
      setupIntersectionObserver(listRef.value, store.articles);
    }
  }
);

// Watch for refresh completion to scroll to top
watch(
  () => store.refreshProgress.isRunning,
  (isRunning) => {
    if (!isRunning && isRefreshing.value) {
      // Refresh completed, scroll to top and reset state
      isRefreshing.value = false;
      shouldRestoreScroll.value = false; // Disable scroll restoration after refresh
      if (listRef.value) {
        listRef.value.scrollTop = 0;
      }
    }
  }
);

// Watch for filtered articles length changes to re-observe new articles
// Changed from deep watch to length watch for better performance
watch(
  () => filteredArticlesFromServer.value.length,
  async () => {
    // Re-setup observer to observe newly added filtered articles
    if (translationSettings.value.enabled && listRef.value) {
      await nextTick();
      setupIntersectionObserver(listRef.value, filteredArticlesFromServer.value);
    }
  }
);

onBeforeUnmount(() => {
  cleanupTranslation();
  // Clear scroll throttle timer
  if (scrollThrottleTimer) {
    clearTimeout(scrollThrottleTimer);
    scrollThrottleTimer = null;
  }
  window.removeEventListener(
    'translation-settings-changed',
    onTranslationSettingsChanged as EventListener
  );
  window.removeEventListener(
    'default-view-mode-changed',
    onDefaultViewModeChanged as EventListener
  );
  window.removeEventListener(
    'show-preview-images-changed',
    onShowPreviewImagesChanged as EventListener
  );
  window.removeEventListener('layout-mode-changed', onLayoutModeChanged as EventListener);
  window.removeEventListener('settings-loaded', onSettingsLoaded as EventListener);
  window.removeEventListener('refresh-articles', onRefreshArticles);
  window.removeEventListener('toggle-filter', onToggleFilter);
  window.removeEventListener('mark-all-as-read', onMarkAllAsRead);
});

interface CustomEventDetail {
  mode?: string;
  enabled?: boolean;
  targetLang?: string;
}

// Event handlers
function onDefaultViewModeChanged(e: Event): void {
  const customEvent = e as CustomEvent<CustomEventDetail>;
  if (customEvent.detail.mode) {
    defaultViewMode.value = customEvent.detail.mode as 'original' | 'rendered';
  }
}

function onTranslationSettingsChanged(e: Event): void {
  const customEvent = e as CustomEvent<CustomEventDetail>;
  const { enabled, targetLang } = customEvent.detail;
  if (enabled !== undefined && targetLang) {
    handleTranslationSettingsChange(enabled, targetLang);

    // Re-setup observer if needed
    if (enabled && listRef.value) {
      setupIntersectionObserver(listRef.value, store.articles);
    }
  }
}

function onMarkAllAsRead(): void {
  void markAllAsRead();
}

function onShowPreviewImagesChanged(e: Event): void {
  const customEvent = e as CustomEvent<{ value: boolean }>;
  const { updateValue } = useShowPreviewImages();
  updateValue(customEvent.detail.value);
}

function onLayoutModeChanged(): void {
  // Force a re-fetch of settings to update the reactive settings object
  fetch('/api/settings')
    .then((res) => res.json())
    .then((data) => {
      settings.value = parseSettingsData(data);
    })
    .catch((err) => console.error('Error refreshing settings after layout mode change:', err));
}

function onSettingsLoaded(): void {
  // Load initial settings when App.vue has loaded them
  fetch('/api/settings')
    .then((res) => res.json())
    .then((data) => {
      settings.value = parseSettingsData(data);
      console.log('ArticleList settings loaded on startup:', settings.value.layout_mode);
    })
    .catch((err) => console.error('Error loading initial settings in ArticleList:', err));
}

function onRefreshArticles(): void {
  store.fetchArticles();
}

function onToggleFilter(): void {
  showFilterModal.value = !showFilterModal.value;
}

// Show tooltip when hovering over refresh button
function onRefreshTooltipShow(): void {
  showRefreshTooltip.value = true;
  // Task details are automatically updated via pollProgress()
}

function onRefreshTooltipHide(): void {
  showRefreshTooltip.value = false;
}

// Article selection and interaction
function selectArticle(article: Article): void {
  // Check if we should open in browser based on feed or global settings
  const feed = store.feeds.find((f) => f.id === article.feed_id);
  let openInBrowserMode = false;

  if (feed?.article_view_mode === 'external') {
    openInBrowserMode = true;
  } else if (feed?.article_view_mode === 'global' || !feed?.article_view_mode) {
    // Check global setting
    if (defaultViewMode.value === 'external') {
      openInBrowserMode = true;
    }
  }

  // If external mode is selected, open in browser and mark as read
  if (openInBrowserMode) {
    // Mark as read if not already read
    if (!article.is_read) {
      article.is_read = true;
      fetch(`/api/articles/read?id=${article.id}&read=true`, { method: 'POST' })
        .then(async () => {
          await store.fetchUnreadCounts();
          await store.fetchFilterCounts();
        })
        .catch((e) => {
          console.error('Error marking as read:', e);
        });
    }
    // Open article URL in browser
    openInBrowser(article.url);
    return;
  }

  // Card mode: open in modal instead of side panel
  if (isCardMode.value) {
    openCardModal(article);
    return;
  }

  // Normal article selection - show in app
  // If switching from one article to another, remove the previous one from temp list
  if (store.currentArticleId) {
    temporarilyKeepArticles.value.delete(store.currentArticleId);
  }

  store.currentArticleId = article.id;
  if (!article.is_read) {
    article.is_read = true;
    // Add to temporarily keep list so it doesn't disappear immediately
    temporarilyKeepArticles.value.add(article.id);
    fetch(`/api/articles/read?id=${article.id}&read=true`, { method: 'POST' })
      .then(async () => {
        await store.fetchUnreadCounts();
        await store.fetchFilterCounts();
      })
      .catch((e) => {
        console.error('Error marking as read:', e);
      });
  }
}

// Scrolling handler with throttling to improve performance
let scrollThrottleTimer: ReturnType<typeof setTimeout> | null = null;
const SCROLL_THROTTLE_DELAY = 200; // 200ms throttle
const SCROLL_THRESHOLD = 400; // Increased from 200 to 400 for better UX

function handleScroll(e: Event): void {
  // Throttle scroll events to improve performance
  if (scrollThrottleTimer) return;

  scrollThrottleTimer = setTimeout(() => {
    scrollThrottleTimer = null;

    const target = e.target as HTMLElement;
    const { scrollTop, clientHeight, scrollHeight } = target;

    // Check if scrolled to bottom (within small threshold)
    const isAtBottom = scrollTop + clientHeight >= scrollHeight - 10;
    hasScrolledToBottom.value = isAtBottom;

    // Load more when user is within threshold distance from bottom
    if (scrollTop + clientHeight >= scrollHeight - SCROLL_THRESHOLD) {
      if (activeFilters.value.length > 0) {
        loadMoreFilteredArticles();
      } else {
        store.loadMore();
      }
    }
  }, SCROLL_THROTTLE_DELAY);
}

// Filter handlers
async function handleApplyFilters(filters: typeof activeFilters.value): Promise<void> {
  activeFilters.value = filters;
  if (filters.length === 0) {
    resetFilterState();
    store.page = 1;
    shouldRestoreScroll.value = false; // Don't restore scroll when clearing filters
    await store.fetchArticles(false);
  } else {
    shouldRestoreScroll.value = false; // Don't restore scroll when applying filters
    await fetchFilteredArticles(filters, false);
  }
}

// Actions
async function refreshArticles(): Promise<void> {
  // Save current scroll position and set refreshing state
  if (listRef.value) {
    savedScrollTop.value = listRef.value.scrollTop;
  }
  isRefreshing.value = true;
  shouldRestoreScroll.value = true; // Enable scroll restoration during refresh

  await store.refreshFeeds();
  // Note: Scrolling to top is now handled by the watch on refreshProgress.isRunning
}

async function markAllAsRead(): Promise<void> {
  // Show confirmation dialog
  const confirmed = await window.showConfirm({
    title: t('article.action.markAllReadConfirmTitle'),
    message: t('article.action.markAllReadConfirmMessage'),
    confirmText: t('common.confirm'),
    cancelText: t('common.cancel'),
    isDanger: false,
  });

  if (!confirmed) {
    return;
  }

  // If filters are active, mark only filtered articles as read
  if (activeFilters.value.length > 0) {
    try {
      // Get IDs of filtered articles
      const articleIds = filteredArticlesFromServer.value.map((a) => a.id);
      if (articleIds.length === 0) {
        window.showToast(t('article.action.noArticlesToMark'), 'info');
        return;
      }

      // Mark all filtered articles as read
      await Promise.all(
        articleIds.map((id) => fetch(`/api/articles/read?id=${id}&read=true`, { method: 'POST' }))
      );

      articleIds.forEach((id) => temporarilyKeepArticles.value.add(id));
      store.setFilteredArticlesFromServer(
        filteredArticlesFromServer.value.map((article) => ({ ...article, is_read: true }))
      );
      store.articles = store.articles.map((article) =>
        articleIds.includes(article.id) ? { ...article, is_read: true } : article
      );
      await store.fetchUnreadCounts();
      await store.fetchFilterCounts();
      window.showToast(t('article.action.markedAllAsRead'), 'success');
    } catch (e) {
      console.error('Error marking filtered articles as read:', e);
    }
  } else {
    // Use store's markAllAsRead which handles feed and category
    const params: { feed_id?: number; category?: string } = {};

    if (store.currentFeedId !== null) {
      params.feed_id = store.currentFeedId;
    } else if (store.currentCategory !== null) {
      params.category = store.currentCategory;
    }

    await store.markAllAsRead(params.feed_id, params.category);
    window.showToast(t('article.action.markedAllAsRead'), 'success');
  }
}

async function clearReadLater(): Promise<void> {
  try {
    const res = await fetch('/api/articles/clear-read-later', { method: 'POST' });
    if (res.ok) {
      await store.fetchArticles();
      await store.fetchFilterCounts();
      window.showToast(t('common.toast.clearedReadLater'), 'success');
    }
  } catch (e) {
    console.error('Error clearing read later:', e);
  }
}

// Handle hover mark as read event from ArticleItem
function handleHoverMarkAsRead(articleId: number): void {
  // Find and update the article in the store
  const article = store.articles.find((a) => a.id === articleId);
  if (article) {
    article.is_read = true;
  }
  // Also update in filtered articles if applicable
  const filteredArticle = filteredArticlesFromServer.value.find((a) => a.id === articleId);
  if (filteredArticle) {
    filteredArticle.is_read = true;
  }
}

// Card mode functions
async function openCardModal(article: Article): Promise<void> {
  cardModalArticle.value = article;
  showCardModal.value = true;
  isCardModalLoading.value = true;
  cardModalContent.value = '';

  // Mark as read
  if (!article.is_read) {
    article.is_read = true;
    temporarilyKeepArticles.value.add(article.id);
    fetch(`/api/articles/read?id=${article.id}&read=true`, { method: 'POST' })
      .then(async () => {
        await store.fetchUnreadCounts();
        await store.fetchFilterCounts();
      })
      .catch((e) => console.error('Error marking as read:', e));
  }

  // Load article content
  try {
    const mediaCacheEnabled = await isMediaCacheEnabled();
    const res = await fetch(`/api/articles/content?id=${article.id}`);
    if (res.ok) {
      const data = await res.json();
      let content = data.content || '';
      if (mediaCacheEnabled && content) {
        content = proxyImagesInHtml(content, article.url);
      }
      cardModalContent.value = content;
    } else {
      cardModalContent.value = '';
    }
  } catch (e) {
    console.error('Error loading article content:', e);
    cardModalContent.value = '';
  } finally {
    isCardModalLoading.value = false;
  }
}

function closeCardModal(): void {
  showCardModal.value = false;
  cardModalArticle.value = null;
  cardModalContent.value = '';
}

function cardModalPrevious(): void {
  if (!cardModalArticle.value) return;
  const currentIndex = filteredArticles.value.findIndex((a) => a.id === cardModalArticle.value!.id);
  if (currentIndex > 0) {
    openCardModal(filteredArticles.value[currentIndex - 1]);
  }
}

function cardModalNext(): void {
  if (!cardModalArticle.value) return;
  const currentIndex = filteredArticles.value.findIndex((a) => a.id === cardModalArticle.value!.id);
  if (currentIndex >= 0 && currentIndex < filteredArticles.value.length - 1) {
    openCardModal(filteredArticles.value[currentIndex + 1]);
  }
}

async function cardModalToggleRead(): Promise<void> {
  if (!cardModalArticle.value) return;
  const article = cardModalArticle.value;
  const newReadState = !article.is_read;

  try {
    await fetch(`/api/articles/read?id=${article.id}&read=${newReadState}`, { method: 'POST' });
    article.is_read = newReadState;
    await store.fetchUnreadCounts();
    await store.fetchFilterCounts();
  } catch (e) {
    console.error('Error toggling read state:', e);
  }
}

async function cardModalToggleFavorite(): Promise<void> {
  if (!cardModalArticle.value) return;
  const article = cardModalArticle.value;
  const newFavoriteState = !article.is_favorite;

  try {
    await fetch(`/api/articles/favorite?id=${article.id}&favorite=${newFavoriteState}`, {
      method: 'POST',
    });
    article.is_favorite = newFavoriteState;
    await store.fetchFilterCounts();
  } catch (e) {
    console.error('Error toggling favorite:', e);
  }
}

async function cardModalToggleReadLater(): Promise<void> {
  if (!cardModalArticle.value) return;
  const article = cardModalArticle.value;

  try {
    await fetch(`/api/articles/toggle-read-later?id=${article.id}`, {
      method: 'POST',
    });
    article.is_read_later = !article.is_read_later;
    await store.fetchFilterCounts();
  } catch (e) {
    console.error('Error toggling read later:', e);
  }
}

function cardModalRetryLoadContent(): void {
  if (cardModalArticle.value) {
    openCardModal(cardModalArticle.value);
  }
}

async function cardModalReloadContent(): Promise<void> {
  if (!cardModalArticle.value) return;

  const article = cardModalArticle.value;
  isCardModalLoading.value = true;
  cardModalContent.value = '';

  try {
    const res = await fetch(`/api/articles/reload-content?id=${article.id}`, { method: 'POST' });
    if (!res.ok) {
      throw new Error(t('common.errors.fetchingArticleContent'));
    }
    await openCardModal(article);
  } catch (e) {
    console.error('Error reloading article content:', e);
    window.showToast(t('common.errors.fetchingArticleContent'), 'error');
    isCardModalLoading.value = false;
  }
}

// Show "Mark All Visible as Read" button at bottom
const shouldShowBottomMarkAllRead = computed(() => {
  return (
    hasScrolledToBottom.value &&
    !store.hasMore &&
    !store.isLoading &&
    !isFilterLoading.value &&
    filteredArticles.value.length > 0
  );
});

// Mark all currently visible articles as read
async function markAllVisibleAsRead(): Promise<void> {
  const articleIds = filteredArticles.value.map((a) => a.id);

  if (articleIds.length === 0) {
    window.showToast(t('article.action.noArticlesToMark'), 'info');
    return;
  }

  try {
    await Promise.all(
      articleIds.map((id) => fetch(`/api/articles/read?id=${id}&read=true`, { method: 'POST' }))
    );

    // Update local article states
    filteredArticles.value.forEach((article) => {
      article.is_read = true;
    });

    // Refresh counts
    await store.fetchUnreadCounts();
    await store.fetchFilterCounts();

    // Show success message with count
    const message = t('article.action.markedNArticlesAsRead', { count: articleIds.length });
    window.showToast(message, 'success');
  } catch (e) {
    console.error('Error marking visible articles as read:', e);
  }
}
</script>

<template>
  <section
    :class="[
      'article-list flex flex-col w-full border-r border-border bg-bg-primary shrink-0 h-full',
      { 'card-mode': isCardMode },
    ]"
  >
    <div class="p-2 sm:p-4 border-b border-border bg-bg-primary">
      <div class="flex items-center justify-between">
        <h3
          class="m-0 text-base sm:text-lg font-semibold truncate flex-1"
          :title="articleListTitle"
        >
          {{ articleListTitle }}
        </h3>
        <div class="flex items-center gap-1 sm:gap-2">
          <!-- Clear Read Later button - only shown when viewing Read Later list -->
          <button
            v-if="store.currentFilter === 'readLater'"
            class="text-text-secondary hover:text-red-500 hover:bg-bg-tertiary p-1 sm:p-1.5 rounded transition-colors"
            :title="t('common.clearReadLater')"
            @click="clearReadLater"
          >
            <PhTrash :size="18" class="sm:w-5 sm:h-5" />
          </button>
          <button
            class="text-text-secondary hover:text-text-primary hover:bg-bg-tertiary p-1 sm:p-1.5 rounded transition-colors"
            :title="t('article.action.markAllRead')"
            @click="markAllAsRead"
          >
            <PhCheckCircle :size="18" class="sm:w-5 sm:h-5" />
          </button>
          <button
            class="text-text-secondary hover:text-text-primary hover:bg-bg-tertiary p-1 sm:p-1.5 rounded transition-colors"
            :class="store.showOnlyUnread ? 'text-accent' : ''"
            :title="
              store.showOnlyUnread
                ? t('setting.reading.showAllArticles')
                : t('setting.reading.showOnlyUnread')
            "
            @click="store.toggleShowOnlyUnread()"
          >
            <component
              :is="store.showOnlyUnread ? PhEyeSlash : PhEye"
              :size="18"
              class="sm:w-5 sm:h-5"
            />
          </button>
          <div class="relative">
            <button
              class="text-text-secondary hover:text-text-primary hover:bg-bg-tertiary p-1 sm:p-1.5 rounded transition-colors"
              :class="activeFilters.length > 0 ? 'filter-active' : ''"
              :title="t('modal.filter.filter')"
              @click="showFilterModal = true"
            >
              <PhFunnel :size="18" class="sm:w-5 sm:h-5" />
            </button>
            <div
              v-if="activeFilters.length > 0"
              class="absolute -top-1 -right-1 bg-accent text-white text-[9px] sm:text-[10px] font-bold rounded-full min-w-[14px] sm:min-w-[16px] h-3.5 sm:h-4 px-0.5 sm:px-1 flex items-center justify-center"
            >
              {{ activeFilters.length }}
            </div>
          </div>
          <div
            class="relative"
            @mouseenter="onRefreshTooltipShow"
            @mouseleave="onRefreshTooltipHide"
          >
            <button
              class="text-text-secondary hover:text-text-primary hover:bg-bg-tertiary p-1 sm:p-1.5 rounded transition-colors"
              :title="t('article.action.refresh')"
              @click="refreshArticles"
            >
              <PhArrowClockwise
                :size="18"
                class="sm:w-5 sm:h-5"
                :class="store.refreshProgress.isRunning ? 'animate-spin' : ''"
              />
            </button>
            <div
              v-if="
                store.refreshProgress.isRunning &&
                (store.refreshProgress.queue_task_count || 0) +
                  (store.refreshProgress.pool_task_count || 0) >
                  0
              "
              class="absolute -top-1 -right-1 bg-accent text-white text-[9px] sm:text-[10px] font-bold rounded-full min-w-[14px] sm:min-w-[16px] h-3.5 sm:h-4 px-0.5 sm:px-1 flex items-center justify-center"
            >
              {{
                (store.refreshProgress.queue_task_count || 0) +
                (store.refreshProgress.pool_task_count || 0)
              }}
            </div>

            <!-- Task Pool Tooltip -->
            <Transition
              enter-active-class="transition ease-out duration-200"
              enter-from-class="opacity-0 scale-95"
              enter-to-class="opacity-100 scale-100"
              leave-active-class="transition ease-in duration-150"
              leave-from-class="opacity-100 scale-100"
              leave-to-class="opacity-0 scale-95"
            >
              <div
                v-if="
                  showRefreshTooltip &&
                  ((store.refreshProgress.pool_task_count || 0) > 0 ||
                    (store.refreshProgress.queue_task_count || 0) > 0 ||
                    (store.refreshProgress.article_click_count || 0) > 0)
                "
                class="absolute right-0 top-full mt-2 z-50 w-72 bg-bg-secondary rounded-lg shadow-xl overflow-hidden"
              >
                <div class="px-3 py-2">
                  <div class="text-xs font-semibold text-text-primary mb-2 flex items-center gap-2">
                    <PhArrowClockwise :size="12" class="animate-spin-slow" />
                    {{ t('article.action.refreshing') }}
                  </div>

                  <!-- Pool Tasks - Show all tasks sorted alphabetically -->
                  <div v-if="(store.refreshProgress.pool_task_count || 0) > 0" class="mb-2">
                    <div
                      class="text-[10px] text-text-secondary mb-1.5 font-medium flex items-center gap-1"
                    >
                      <PhCircle :size="10" class="text-accent" />
                      {{ t('article.progress.activeTasks') }} ({{
                        store.refreshProgress.pool_task_count || 0
                      }})
                    </div>
                    <div class="space-y-0.5">
                      <div
                        v-for="(task, index) in store.refreshProgress.pool_tasks || []"
                        :key="'pool-' + index"
                        class="text-xs text-text-primary bg-accent/10 px-2.5 py-1.5 rounded truncate"
                        :title="task.feed_title"
                      >
                        <div class="flex items-center gap-2">
                          <PhCircle :size="10" class="text-accent animate-pulse flex-shrink-0" />
                          <span class="truncate flex-1">{{ task.feed_title }}</span>
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- Queue Tasks - Show first 3 -->
                  <div v-if="(store.refreshProgress.queue_task_count || 0) > 0">
                    <div
                      class="text-[10px] text-text-secondary mb-1.5 font-medium flex items-center gap-1"
                    >
                      <PhClock :size="10" />
                      {{ t('sidebar.activity.queuedTasks') }} ({{
                        store.refreshProgress.queue_task_count || 0
                      }})
                    </div>
                    <div class="space-y-0.5">
                      <div
                        v-for="(task, index) in store.refreshProgress.queue_tasks || []"
                        :key="'queue-' + index"
                        class="text-xs text-text-secondary bg-bg-tertiary/50 px-2.5 py-1.5 rounded truncate"
                        :title="task.feed_title"
                      >
                        <div class="flex items-center gap-2">
                          <PhClock :size="10" class="flex-shrink-0" />
                          <span class="truncate flex-1">{{ task.feed_title }}</span>
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- Article Click Tasks -->
                  <div
                    v-if="(store.refreshProgress.article_click_count || 0) > 0"
                    class="mt-2 pt-2 border-t border-border/50"
                  >
                    <div
                      class="text-[10px] text-text-secondary mb-1.5 font-medium flex items-center gap-1"
                    >
                      <PhLightning :size="10" class="text-accent" />
                      {{ t('sidebar.activity.immediateTasks') }} ({{
                        store.refreshProgress.article_click_count || 0
                      }})
                    </div>
                    <div class="text-xs text-accent bg-accent/10 px-2.5 py-1.5 rounded truncate">
                      <div class="flex items-center gap-2">
                        <PhLightning :size="10" class="flex-shrink-0" />
                        <span class="truncate">{{
                          t('article.content.fetchingArticleContent')
                        }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </Transition>
          </div>
          <button class="md:hidden text-xl sm:text-2xl p-1" @click="emit('toggleSidebar')">
            <PhList :size="18" class="sm:w-5 sm:h-5" />
          </button>
        </div>
      </div>
    </div>

    <!-- AI Search Bar -->
    <AISearchBar
      v-if="isAISearchEnabled"
      @search="handleAISearchResults"
      @clear="handleAISearchClear"
    />

    <div ref="listRef" class="flex-1 overflow-y-scroll article-list-scroll" @scroll="handleScroll">
      <div
        v-if="
          filteredArticles.length === 0 && !store.isLoading && !isFilterLoading && !isAISearchActive
        "
        class="p-4 sm:p-5 text-center text-text-secondary text-sm sm:text-base"
      >
        {{ t('article.content.noArticles') }}
      </div>

      <!-- AI Search no results message -->
      <div
        v-if="isAISearchActive && filteredArticles.length === 0 && !store.isLoading"
        class="p-4 sm:p-5 text-center text-text-secondary text-sm sm:text-base"
      >
        {{ t('aiSearch.noResults') }}
      </div>

      <!-- Article list with content-visibility for performance -->
      <!-- Card mode: grid layout -->
      <div v-if="isCardMode" class="card-grid-container">
        <ArticleCardItem
          v-for="article in visibleArticles"
          :key="article.id"
          :article="article"
          :is-active="cardModalArticle?.id === article.id"
          @click="selectArticle(article)"
          @contextmenu="(e) => showArticleContextMenu(e, article)"
        />
      </div>
      <!-- Normal/Compact mode: list layout -->
      <div v-else class="article-list-container">
        <ArticleItem
          v-for="article in visibleArticles"
          :key="article.id"
          :article="article"
          :is-active="store.currentArticleId === article.id"
          @click="selectArticle(article)"
          @contextmenu="(e) => showArticleContextMenu(e, article)"
          @observe-element="observeArticle"
          @hover-mark-as-read="handleHoverMarkAsRead"
        />
      </div>

      <!-- Bottom: Mark All Visible as Read button (inserted at end of list) -->
      <Transition
        enter-active-class="transition ease-out duration-200"
        enter-from-class="opacity-0 translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition ease-in duration-150"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 translate-y-2"
      >
        <div v-if="shouldShowBottomMarkAllRead" class="mx-3 mb-3 pt-6 pb-3 text-center">
          <button
            class="inline-flex items-center gap-2 px-4 py-2 bg-accent hover:bg-accent/80 text-white rounded-lg transition-colors text-sm font-medium"
            @click="markAllVisibleAsRead"
          >
            <PhCheckCircle :size="18" />
            <span>{{ t('article.list.markAllVisibleAsRead') }}</span>
          </button>
          <div class="text-xs text-text-secondary mt-2">
            {{ t('article.list.allArticlesLoaded') }}
          </div>
        </div>
      </Transition>

      <div
        v-if="store.isLoading || isFilterLoading"
        class="p-3 sm:p-4 text-center text-text-secondary"
      >
        <PhSpinner :size="20" class="animate-spin sm:w-6 sm:h-6" />
      </div>
    </div>
  </section>

  <!-- Card Mode Article Modal -->
  <ArticleDetailModal
    v-if="showCardModal && cardModalArticle"
    :article="cardModalArticle"
    :article-content="cardModalContent"
    :is-loading-content="isCardModalLoading"
    @close="closeCardModal"
    @previous="cardModalPrevious"
    @next="cardModalNext"
    @toggle-read="cardModalToggleRead"
    @toggle-favorite="cardModalToggleFavorite"
    @toggle-read-later="cardModalToggleReadLater"
    @retry-load-content="cardModalRetryLoadContent"
    @reload-content="cardModalReloadContent"
  />

  <!-- Filter Modal - Teleported to body to avoid positioning constraints -->
  <Teleport to="body">
    <ArticleFilterModal
      :show="showFilterModal"
      :current-filters="activeFilters"
      @close="showFilterModal = false"
      @apply="handleApplyFilters"
    />
  </Teleport>
</template>

<style scoped>
@reference "../../style.css";
@media (min-width: 768px) {
  .article-list {
    width: var(--article-list-width, 400px);
  }
}

/* Responsive width for article list on medium screens */
@media (max-width: 1400px) and (min-width: 768px) {
  .article-list {
    width: min(var(--article-list-width, 400px), 320px) !important;
  }
}

/* Card mode: full width, no max-width restriction */
.article-list.card-mode {
  @apply flex-1;
  width: auto !important;
  max-width: none !important;
  border-right: none;
}

@media (min-width: 768px) {
  .article-list.card-mode {
    width: auto !important;
    max-width: none !important;
  }
}

/* Card grid layout - narrower cards */
.card-grid-container {
  @apply grid gap-3 p-3;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
}

/* Responsive adjustments for card grid */
@media (min-width: 640px) {
  .card-grid-container {
    grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
  }
}

@media (min-width: 1024px) {
  .card-grid-container {
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  }
}

@media (min-width: 1400px) {
  .card-grid-container {
    grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
  }
}

.filter-active {
  @apply text-accent border-accent;
  background-color: rgba(59, 130, 246, 0.1);
}

.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* Performance optimization: content-visibility for article list */
.article-list-container {
  content-visibility: auto;
  contain-intrinsic-size: auto 200px;
}

/* Optimize scrolling performance */
.article-list-scroll {
  /* Enable GPU acceleration for smooth scrolling */
  transform: translateZ(0);
  -webkit-transform: translateZ(0);
  /* Optimize scroll performance */
  overflow-anchor: none;
  /* Smooth scrolling behavior */
  scroll-behavior: auto;
}

.article-list {
  /* Enable GPU acceleration for smooth scrolling */
  transform: translateZ(0);
  -webkit-transform: translateZ(0);
}

/* Optimize article card rendering */
.article-card {
  /* Only use will-change when actually animating */
  will-change: auto;
  /* Isolate compositing layers for better performance */
  contain: layout style paint;
  /* Smooth hover transitions */
  transition: background-color 0.15s ease;
}

.article-card:hover {
  /* Enable GPU acceleration during hover */
  transform: translateZ(0);
  -webkit-transform: translateZ(0);
}
</style>
