import { onBeforeUnmount, ref } from 'vue';

export interface FeedPreviewArticle {
  title: string;
  url: string;
  published_at?: string;
  // Prepared and sanitized by the feed preview endpoint.
  content_html: string;
  truncated: boolean;
}
export interface FeedPreview {
  title: string;
  description: string;
  total: number;
  articles: FeedPreviewArticle[];
}
export interface FeedPreviewRequest {
  url: string;
  proxy_enabled: boolean;
  proxy_url: string;
}

export function useFeedPreview() {
  const preview = ref<FeedPreview | null>(null);
  const isLoading = ref(false);
  const failed = ref(false);
  let controller: AbortController | null = null;

  function reset() {
    controller?.abort();
    controller = null;
    preview.value = null;
    failed.value = false;
    isLoading.value = false;
  }

  async function load(input: FeedPreviewRequest): Promise<void> {
    reset();
    const request = new AbortController();
    controller = request;
    isLoading.value = true;
    try {
      const response = await fetch('/api/feeds/preview', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input),
        signal: request.signal,
      });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const data: FeedPreview = await response.json();
      if (!Array.isArray(data.articles)) throw new Error('Invalid preview');
      if (controller === request && !request.signal.aborted) preview.value = data;
    } catch {
      if (controller === request && !request.signal.aborted) failed.value = true;
    } finally {
      if (controller === request) {
        isLoading.value = false;
        controller = null;
      }
    }
  }

  onBeforeUnmount(reset);
  return { preview, isLoading, failed, load, reset };
}
