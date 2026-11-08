import { onMounted, onUnmounted } from 'vue';
import { Events, System } from '@wailsio/runtime';
import type { Article } from '@/types/models';
import i18n from '@/i18n';

interface ReaderNavigation {
  currentArticleId: number | null;
  currentFeedId: number | null;
  setArticleNavigationContext: (articles: Article[] | null) => void;
  selectFeedInArticleList: (feedId: number, articleId: number) => void;
}

export function useTickerNavigation(
  store: ReaderNavigation,
  openSettings?: () => void | Promise<void>
) {
  const disposers: Array<() => void> = [];
  let controller: AbortController | undefined;
  let disposed = false;
  onMounted(() => {
    if (!System.IsDesktop()) return;
    disposers.push(
      Events.On('ticker:open-settings', () => {
        controller?.abort();
        void openSettings?.();
      })
    );
    disposers.push(
      Events.On('ticker:open-article', async (event) => {
        const data = event.data as { article_id?: unknown; feed_id?: unknown } | null;
        if (
          !data ||
          typeof data.article_id !== 'number' ||
          typeof data.feed_id !== 'number' ||
          !Number.isSafeInteger(data.article_id) ||
          !Number.isSafeInteger(data.feed_id) ||
          data.article_id <= 0 ||
          data.feed_id <= 0
        )
          return;
        controller?.abort();
        const current = new AbortController();
        controller = current;
        const beforeArticle = store.currentArticleId,
          beforeFeed = store.currentFeedId;
        const isCurrent = () =>
          !disposed &&
          !current.signal.aborted &&
          store.currentArticleId === beforeArticle &&
          store.currentFeedId === beforeFeed;
        try {
          const response = await fetch(`/api/articles?id=${data.article_id}`, {
            signal: current.signal,
          });
          if (!response.ok) throw new Error('Could not load article');
          const result: Article[] = await response.json();
          const selected = Array.isArray(result)
            ? result.find((item) => item.id === data.article_id && item.feed_id === data.feed_id)
            : undefined;
          if (!isCurrent()) return;
          if (!selected) throw new Error('Article is no longer available');
          // Seed the selection before loading its feed. This also works when the
          // main list is sorted oldest first and the headline is outside its page.
          store.setArticleNavigationContext([selected]);
          store.selectFeedInArticleList(data.feed_id, data.article_id);
        } catch {
          if (isCurrent())
            window.showToast?.(i18n.global.t('common.errors.fetchingArticleContent'), 'error');
        }
      })
    );
  });
  onUnmounted(() => {
    disposed = true;
    controller?.abort();
    disposers.forEach((dispose) => dispose());
  });
}
