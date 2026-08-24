import { ref, computed, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAppStore } from '@/stores/app';
import type { Feed } from '@/types/models';
import type { DiscoveredFeed } from '@/types/discovery';

export function useFeedSubscription(feed: Feed, discoveredFeeds: Ref<DiscoveredFeed[]>) {
  const { t } = useI18n();
  const store = useAppStore();

  const selectedFeeds: Ref<Set<number>> = ref(new Set());
  const isSubscribing = ref(false);

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
      const discoveredFeed = discoveredFeeds.value[index];
      const promise = fetch('/api/feeds/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          url: discoveredFeed.rss_feed,
          category: feed.category || '',
          title: discoveredFeed.name,
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

  return {
    selectedFeeds,
    isSubscribing,
    hasSelection,
    allSelected,
    toggleFeedSelection,
    selectAll,
    subscribeSelected,
  };
}
