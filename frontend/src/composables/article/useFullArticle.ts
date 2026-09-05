import { ref, watch, onBeforeUnmount } from 'vue';
import type { Article } from '@/types/models';
import { isMediaCacheEnabled, proxyImagesInHtml } from '@/utils/mediaProxy';

interface Options {
  article: () => Article;
  enabled: () => boolean;
  automatic: () => boolean;
  loading: () => boolean;
  onContent: (content: string) => void | Promise<void>;
  onError: () => void;
  onSuccess: () => void;
}

export function useFullArticle(options: Options) {
  const content = ref('');
  const loading = ref(false);
  let generation = 0;
  let controller: AbortController | null = null;
  let attempted: number | null = null;

  function reset() {
    generation += 1;
    controller?.abort();
    controller = null;
    content.value = '';
    loading.value = false;
    attempted = null;
  }

  async function fetchFullArticle(showErrors = true) {
    const article = options.article();
    if (!article?.id || !options.enabled() || loading.value) return;
    const requestId = ++generation;
    controller?.abort();
    controller = new AbortController();
    attempted = article.id;
    loading.value = true;
    const current = () => requestId === generation && options.article()?.id === article.id;
    try {
      const response = await fetch(`/api/articles/fetch-full?id=${article.id}`, {
        method: 'POST',
        signal: controller.signal,
      });
      if (!response.ok) throw new Error(`Full article: ${response.status}`);
      const data = await response.json();
      const cacheEnabled = await isMediaCacheEnabled();
      if (!current()) return;
      if (typeof data.content !== 'string' || !data.content.trim())
        throw new Error('Empty article');
      content.value = cacheEnabled
        ? proxyImagesInHtml(data.content, data.feed_url || article.url)
        : data.content;
      if (showErrors) options.onSuccess();
      await options.onContent(content.value);
    } catch (error) {
      if (current() && showErrors) options.onError();
    } finally {
      if (current()) loading.value = false;
    }
  }

  watch(() => options.article()?.id, reset, { flush: 'sync' });
  watch(
    () => [options.article()?.id, options.enabled(), options.automatic(), options.loading()],
    () => {
      if (
        options.enabled() &&
        options.automatic() &&
        !options.loading() &&
        attempted !== options.article()?.id &&
        !content.value
      )
        void fetchFullArticle(false);
    },
    { immediate: true, flush: 'post' }
  );
  function onReload(event: Event) {
    if ((event as CustomEvent<number>).detail === options.article()?.id) reset();
  }
  window.addEventListener('article-content-reloaded', onReload);
  onBeforeUnmount(() => {
    reset();
    window.removeEventListener('article-content-reloaded', onReload);
  });
  return { fullArticleContent: content, isFetchingFullArticle: loading, fetchFullArticle };
}
