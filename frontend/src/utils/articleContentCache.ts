/**
 * Small LRU cache for article bodies.
 *
 * The reader asks the backend for the body on every selection, so moving back
 * and forth between articles re-downloads and re-parses the same HTML. Keeping
 * the raw response for the most recent articles makes repeated selections cheap.
 *
 * Entries are only dropped on eviction or when callers invalidate them, which
 * they must do whenever the stored body can change (reload content, full-text
 * fetch, bulk content cleanup).
 */

export interface ArticleContentResponse {
  content: string;
  feedUrl: string;
  cached: boolean;
}

const MAX_CACHED_ARTICLES = 8;

const cachedContent = new Map<number, ArticleContentResponse>();
const inFlightRequests = new Map<number, Promise<ArticleContentResponse>>();

/** Returns the cached body and refreshes its LRU position. */
export function getCachedArticleContent(articleId: number): ArticleContentResponse | undefined {
  const entry = cachedContent.get(articleId);
  if (!entry) return undefined;
  cachedContent.delete(articleId);
  cachedContent.set(articleId, entry);
  return entry;
}

export function invalidateArticleContent(articleId: number): void {
  cachedContent.delete(articleId);
}

export function clearArticleContentCache(): void {
  cachedContent.clear();
}

export function getArticleContentCacheSize(): number {
  return cachedContent.size;
}

/**
 * Loads an article body, reusing the cache and any request that is already in
 * flight for the same article.
 */
export async function loadArticleContent(
  articleId: number,
  signal?: AbortSignal
): Promise<ArticleContentResponse> {
  const hit = getCachedArticleContent(articleId);
  if (hit) return hit;

  const pending = inFlightRequests.get(articleId);
  if (pending) return pending;

  const request = (async (): Promise<ArticleContentResponse> => {
    const response = await fetch(`/api/articles/content?id=${articleId}`, { signal });
    if (!response.ok) {
      throw new Error(`Failed to load article content: ${response.status}`);
    }
    const data = await response.json();
    const entry: ArticleContentResponse = {
      content: typeof data.content === 'string' ? data.content : '',
      feedUrl: typeof data.feed_url === 'string' ? data.feed_url : '',
      cached: data.cached === true,
    };

    // An empty body means the backend could not produce the article this time
    // (it fetches the source on demand and returns "cached: false" with nothing
    // when that fails). Caching it would leave the reader stuck on "no content"
    // until an unrelated reload, so only non-empty bodies are remembered.
    if (entry.content.trim() !== '') {
      cachedContent.set(articleId, entry);
      while (cachedContent.size > MAX_CACHED_ARTICLES) {
        const oldest = cachedContent.keys().next();
        if (oldest.done) break;
        cachedContent.delete(oldest.value);
      }
    }
    return entry;
  })();

  inFlightRequests.set(articleId, request);
  try {
    return await request;
  } finally {
    inFlightRequests.delete(articleId);
  }
}
