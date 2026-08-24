import { ref, computed, onUnmounted, type Ref } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';

export interface DiscoveredFeed {
  name: string;
  homepage: string;
  rss_feed: string;
  icon_url?: string;
  recent_articles?: Array<{
    title: string;
    date?: string;
  }>;
}

interface ProgressCounts {
  current: number;
  total: number;
  found: number;
}

interface StartResult {
  status: string;
  message?: string;
  total?: number;
}

interface ProgressState {
  is_complete: boolean;
  error?: string;
  feeds?: DiscoveredFeed[];
  progress?: {
    stage: string;
    message?: string;
    detail?: string;
    current?: number;
    total?: number;
    found_count?: number;
    feed_name?: string;
  };
}

export function useDiscoverAllFeeds() {
  const store = useAppStore();
  const { t } = useI18n();

  const isDiscovering = ref(false);
  const discoveredFeeds: Ref<DiscoveredFeed[]> = ref([]);
  const selectedFeeds: Ref<Set<number>> = ref(new Set());
  const errorMessage = ref('');
  const progressMessage = ref('');
  const progressDetail = ref('');
  const progressCounts: Ref<ProgressCounts> = ref({ current: 0, total: 0, found: 0 });
  const isSubscribing = ref(false);
  let pollInterval: ReturnType<typeof setInterval> | null = null;

  function getHostname(url: string): string {
    try {
      return new URL(url).hostname;
    } catch {
      return url;
    }
  }

  async function startDiscovery() {
    isDiscovering.value = true;
    errorMessage.value = '';
    discoveredFeeds.value = [];
    selectedFeeds.value.clear();
    progressMessage.value = t('modal.discovery.preparingDiscovery');
    progressDetail.value = '';
    progressCounts.value = { current: 0, total: 0, found: 0 };

    // Clear any existing poll interval
    if (pollInterval) {
      clearInterval(pollInterval);
      pollInterval = null;
    }

    try {
      // Clear any previous discovery state
      await fetch('/api/feeds/discover-all/clear', { method: 'POST' });

      // Start batch discovery in background
      const startResponse = await fetch('/api/feeds/discover-all/start', {
        method: 'POST',
      });

      if (!startResponse.ok) {
        const errorText = await startResponse.text();
        throw new Error(errorText || 'Failed to start batch discovery');
      }

      const startResult = (await startResponse.json()) as StartResult;

      // Check if already complete (all feeds discovered)
      if (startResult.status === 'complete') {
        errorMessage.value = startResult.message || t('modal.discovery.noFriendLinksFound');
        isDiscovering.value = false;
        return;
      }

      progressCounts.value.total = startResult.total || 0;

      // Start polling for progress
      pollInterval = setInterval(async () => {
        try {
          const progressResponse = await fetch('/api/feeds/discover-all/progress');
          if (!progressResponse.ok) {
            throw new Error('Failed to get progress');
          }

          const state = (await progressResponse.json()) as ProgressState;

          // Update progress display
          if (state.progress) {
            const progress = state.progress;
            switch (progress.stage) {
              case 'starting':
                progressMessage.value = t('modal.discovery.preparingDiscovery');
                progressDetail.value = '';
                break;
              case 'processing_feed':
                progressMessage.value = t('modal.discovery.processingFeed', {
                  current: progress.current || 0,
                  total: progress.total || 0,
                });
                progressDetail.value = progress.feed_name || '';
                break;
              case 'fetching_homepage':
                progressMessage.value = t('modal.discovery.fetchingHomepage');
                progressDetail.value = progress.feed_name ? `${progress.feed_name}` : '';
                break;
              case 'finding_friend_links':
                progressMessage.value = t('modal.discovery.searchingFriendLinks');
                progressDetail.value = progress.feed_name || '';
                break;
              case 'fetching_friend_page':
                progressMessage.value = t('modal.discovery.fetchingFriendPage');
                progressDetail.value = progress.feed_name || '';
                break;
              case 'checking_rss':
                progressMessage.value = t('modal.discovery.checkingRssFeed');
                progressDetail.value =
                  progress.feed_name +
                  (progress.detail ? ' - ' + getHostname(progress.detail) : '');
                break;
              default:
                progressMessage.value = progress.message || t('modal.discovery.discovering');
                progressDetail.value =
                  progress.feed_name || (progress.detail ? getHostname(progress.detail) : '');
            }
            progressCounts.value.current = progress.current || 0;
            progressCounts.value.total = progress.total || 0;
            progressCounts.value.found = progress.found_count || 0;
          }

          // Check if complete
          if (state.is_complete) {
            if (pollInterval !== null) {
              clearInterval(pollInterval);
              pollInterval = null;
            }

            if (state.error) {
              errorMessage.value = state.error;
            } else {
              discoveredFeeds.value = state.feeds || [];
              if (discoveredFeeds.value.length === 0) {
                errorMessage.value = t('modal.discovery.noFriendLinksFound');
              }
            }

            isDiscovering.value = false;
            progressMessage.value = '';
            progressDetail.value = '';

            // Refresh feeds to show updated discovery status
            await store.fetchFeeds();

            // Clear the discovery state
            await fetch('/api/feeds/discover-all/clear', { method: 'POST' });
          }
        } catch (pollError) {
          console.error('Polling error:', pollError);
          // Don't stop polling on transient errors
        }
      }, 500); // Poll every 500ms
    } catch (error) {
      console.error('Batch discovery error:', error);
      errorMessage.value = t('modal.discovery.discoveryFailed') + ': ' + (error as Error).message;
      isDiscovering.value = false;
      progressMessage.value = '';
      progressDetail.value = '';
      if (pollInterval) {
        clearInterval(pollInterval);
        pollInterval = null;
      }
    }
  }

  function toggleFeedSelection(index: number) {
    if (selectedFeeds.value.has(index)) {
      selectedFeeds.value.delete(index);
    } else {
      selectedFeeds.value.add(index);
    }
  }

  function selectAll() {
    if (selectedFeeds.value.size === discoveredFeeds.value.length) {
      selectedFeeds.value.clear();
    } else {
      discoveredFeeds.value.forEach((_, index) => selectedFeeds.value.add(index));
    }
  }

  const hasSelection = computed(() => selectedFeeds.value.size > 0);
  const allSelected = computed(
    () =>
      discoveredFeeds.value.length > 0 && selectedFeeds.value.size === discoveredFeeds.value.length
  );

  async function subscribeSelected() {
    if (!hasSelection.value) return;

    isSubscribing.value = true;
    const subscribePromises = [];

    for (const index of selectedFeeds.value) {
      const feed = discoveredFeeds.value[index];
      const promise = fetch('/api/feeds/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          url: feed.rss_feed,
          category: '',
          title: feed.name,
        }),
      });
      subscribePromises.push(promise);
    }

    try {
      const results = await Promise.allSettled(subscribePromises);
      const successful = results.filter((r) => r.status === 'fulfilled').length;
      const failed = results.filter((r) => r.status === 'rejected').length;

      await store.fetchFeeds();

      if (failed === 0) {
        window.showToast(t('modal.feed.feedsSubscribedSuccess', { count: successful }), 'success');
      } else {
        window.showToast(t('modal.feed.feedsSubscribedPartial', { successful, failed }), 'warning');
      }
    } catch (error) {
      console.error('Subscription error:', error);
      window.showToast(t('common.errors.subscribingFeeds'), 'error');
    } finally {
      isSubscribing.value = false;
    }
  }

  function cleanup() {
    // Clear polling interval if active
    if (pollInterval) {
      clearInterval(pollInterval);
      pollInterval = null;
    }
    // Clear discovery state on server
    fetch('/api/feeds/discover-all/clear', { method: 'POST' }).catch(() => {});
  }

  // Cleanup on unmount
  onUnmounted(() => {
    cleanup();
  });

  return {
    // State
    isDiscovering,
    discoveredFeeds,
    selectedFeeds,
    errorMessage,
    progressMessage,
    progressDetail,
    progressCounts,
    isSubscribing,
    hasSelection,
    allSelected,

    // Functions
    startDiscovery,
    toggleFeedSelection,
    selectAll,
    subscribeSelected,
    cleanup,
  };
}
