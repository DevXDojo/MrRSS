import { computed, onMounted, onUnmounted, ref, watch, type Ref } from 'vue';
import type { SettingsData } from '@/types/settings';
import type { Article } from '@/types/models';
import { tickerFeedIDs } from './tickerPreferences';

export function useTickerArticles(settings: Ref<SettingsData>, visible: Ref<boolean>) {
  const articles = ref<Article[]>([]);
  const loading = ref(false);
  const failed = ref(false);
  const refreshedAt = ref<Date | null>(null);
  const newOnResume = ref(0);
  const sources = computed(() => tickerFeedIDs(settings.value).join(','));
  let controller: AbortController | undefined;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let disposed = false;
  let resumePending = false;
  const fingerprint = (items: Article[]) =>
    JSON.stringify(
      items.map((a) => [
        a.id,
        a.title,
        a.translated_title,
        a.feed_title,
        a.published_at,
        a.is_read,
        a.original_summary,
        a.summary,
      ])
    );

  async function refresh() {
    clearTimeout(timer);
    controller?.abort();
    if (disposed || !visible.value) return;
    const current = new AbortController();
    controller = current;
    loading.value = true;
    try {
      const query = new URLSearchParams({
        filter: settings.value.ticker_only_unread ? 'unread' : 'all',
        limit: '50',
        sort_order: 'newest',
        include_original_summary: 'true',
      });
      if (sources.value) query.set('feed_ids', sources.value);
      const response = await fetch(`/api/articles?${query}`, { signal: current.signal });
      if (!response.ok) throw new Error('Could not load ticker articles');
      const result: Article[] | null = await response.json();
      if (result !== null && !Array.isArray(result)) throw new Error('Invalid article response');
      if (disposed || current.signal.aborted) return;
      const items = (result ?? []).filter((article) => article.title?.trim());
      if (resumePending) {
        const previousIDs = new Set(articles.value.map((article) => article.id));
        newOnResume.value = items.filter((article) => !previousIDs.has(article.id)).length;
        resumePending = false;
      }
      if (fingerprint(items) !== fingerprint(articles.value)) articles.value = items;
      failed.value = false;
      refreshedAt.value = new Date();
    } catch {
      if (!disposed && !current.signal.aborted) failed.value = true;
    } finally {
      if (!disposed && !current.signal.aborted) {
        loading.value = false;
        if (visible.value) timer = setTimeout(refresh, 60000);
      }
    }
  }
  watch([sources, () => settings.value.ticker_only_unread], () => {
    articles.value = [];
    failed.value = false;
    resumePending = false;
    newOnResume.value = 0;
    refreshedAt.value = null;
    void refresh();
  });
  watch(visible, (shown) => {
    if (shown) {
      resumePending = refreshedAt.value !== null;
      void refresh();
    } else {
      clearTimeout(timer);
      controller?.abort();
      loading.value = false;
      newOnResume.value = 0;
    }
  });
  onMounted(() => {
    void refresh();
  });
  onUnmounted(() => {
    disposed = true;
    clearTimeout(timer);
    controller?.abort();
  });
  return { articles, loading, failed, refreshedAt, newOnResume, refresh };
}
