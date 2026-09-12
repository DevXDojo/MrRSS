import type { Article } from '@/types/models';

export type ArticleGroupBy = 'none' | 'date' | 'feed';

export function parseArticleGroupBy(value: string | null): ArticleGroupBy {
  return value === 'date' || value === 'feed' ? value : 'none';
}

export function articleGroupKey(article: Article, groupBy: ArticleGroupBy): string {
  if (groupBy === 'feed') return `feed:${article.feed_id}`;
  if (groupBy === 'date') {
    const date = new Date(article.published_at);
    if (Number.isNaN(date.getTime())) return 'date:unknown';
    // Calendar days follow the reader's local timezone, not the UTC date prefix.
    return `date:${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`;
  }
  return '';
}

export function orderGroupedArticles(
  articles: Article[],
  groupBy: ArticleGroupBy,
  sortOrder: 'newest' | 'oldest'
): Article[] {
  if (groupBy === 'none') return articles;
  return [...articles].sort((left, right) => {
    if (groupBy === 'feed' && left.feed_id !== right.feed_id) return left.feed_id - right.feed_id;
    const leftTime = new Date(left.published_at).getTime();
    const rightTime = new Date(right.published_at).getTime();
    const delta =
      (Number.isFinite(leftTime) ? leftTime : 0) - (Number.isFinite(rightTime) ? rightTime : 0);
    return sortOrder === 'oldest' ? delta || left.id - right.id : -delta || right.id - left.id;
  });
}

export function articleGroupStarts(articles: Article[], groupBy: ArticleGroupBy): Set<number> {
  const starts = new Set<number>();
  if (groupBy === 'none') return starts;
  let previous: string | null = null;
  for (const article of articles) {
    const key = articleGroupKey(article, groupBy);
    if (key !== previous) starts.add(article.id);
    previous = key;
  }
  return starts;
}
