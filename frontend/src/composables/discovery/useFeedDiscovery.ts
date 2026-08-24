import { ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Feed } from '@/types/models';
import type { DiscoveredFeed, ProgressCounts, ProgressState } from '@/types/discovery';

export function useFeedDiscovery(feed: Feed) {
  const { t } = useI18n();

  const isDiscovering = ref(false);
  const discoveredFeeds: Ref<DiscoveredFeed[]> = ref([]);
  const errorMessage = ref('');
  const progressMessage = ref('');
  const progressDetail = ref('');
  const progressCounts: Ref<ProgressCounts> = ref({ current: 0, total: 0, found: 0 });
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
    progressMessage.value = t('modal.discovery.fetchingHomepage');
    progressDetail.value = '';
    progressCounts.value = { current: 0, total: 0, found: 0 };

    // Clear any existing poll interval
    if (pollInterval) {
      clearInterval(pollInterval);
      pollInterval = null;
    }

    try {
      // Validate feed ID
      if (!feed?.id) {
        throw new Error('Invalid feed ID');
      }

      // Clear any previous discovery state
      await fetch('/api/feeds/discover/clear', { method: 'POST' });

      // Start discovery in background
      const startResponse = await fetch('/api/feeds/discover/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ feed_id: feed.id }),
      });

      if (!startResponse.ok) {
        const errorText = await startResponse.text();
        throw new Error(errorText || 'Failed to start discovery');
      }

      // Start polling for progress
      pollInterval = setInterval(async () => {
        try {
          const progressResponse = await fetch('/api/feeds/discover/progress');
          if (!progressResponse.ok) {
            throw new Error('Failed to get progress');
          }

          const state = (await progressResponse.json()) as ProgressState;

          // Update progress display
          if (state.progress) {
            const progress = state.progress;
            switch (progress.stage) {
              case 'fetching_homepage':
                progressMessage.value = t('modal.discovery.fetchingHomepage');
                progressDetail.value = progress.detail ? getHostname(progress.detail) : '';
                break;
              case 'finding_friend_links':
                progressMessage.value = t('modal.discovery.searchingFriendLinks');
                progressDetail.value = progress.detail ? getHostname(progress.detail) : '';
                break;
              case 'fetching_friend_page':
                progressMessage.value = t('modal.discovery.fetchingFriendPage');
                progressDetail.value = progress.detail ? getHostname(progress.detail) : '';
                break;
              case 'found_links':
                progressMessage.value = t('modal.discovery.foundPotentialLinks', {
                  count: progress.total,
                });
                progressDetail.value = '';
                progressCounts.value.total = progress.total || 0;
                break;
              case 'checking_rss':
                progressMessage.value = t('modal.discovery.checkingRssFeed');
                progressDetail.value = progress.detail ? getHostname(progress.detail) : '';
                progressCounts.value.current = progress.current || 0;
                progressCounts.value.total = progress.total || 0;
                progressCounts.value.found = progress.found_count || 0;
                break;
              default:
                progressMessage.value = progress.message || t('modal.discovery.discovering');
                progressDetail.value = progress.detail ? getHostname(progress.detail) : '';
            }
          }

          // Check if complete
          if (state.is_complete) {
            if (pollInterval !== null) {
              clearInterval(pollInterval);
              pollInterval = null;
            }

            if (state.error) {
              errorMessage.value = t('modal.discovery.discoveryFailed') + ': ' + state.error;
            } else {
              discoveredFeeds.value = state.feeds || [];
              if (discoveredFeeds.value.length === 0) {
                errorMessage.value = t('modal.discovery.noFriendLinksFound');
              }
            }

            isDiscovering.value = false;
            progressMessage.value = '';
            progressDetail.value = '';

            // Clear the discovery state
            await fetch('/api/feeds/discover/clear', { method: 'POST' });
          }
        } catch (pollError) {
          console.error('Polling error:', pollError);
          // Don't stop polling on transient errors
        }
      }, 500); // Poll every 500ms
    } catch (error) {
      console.error('Discovery error:', error);
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

  function cleanup() {
    if (pollInterval) {
      clearInterval(pollInterval);
      pollInterval = null;
    }
    // Clear discovery state on server
    fetch('/api/feeds/discover/clear', { method: 'POST' }).catch(() => {});
  }

  return {
    isDiscovering,
    discoveredFeeds,
    errorMessage,
    progressMessage,
    progressDetail,
    progressCounts,
    startDiscovery,
    cleanup,
  };
}
