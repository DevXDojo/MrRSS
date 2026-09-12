import { describe, expect, it } from 'vitest';
import type { Article } from '@/types/models';
import {
  articleGroupKey,
  articleGroupStarts,
  orderGroupedArticles,
  parseArticleGroupBy,
} from './articleGrouping';

function article(id: number, feed: number, date: Date): Article {
  return {
    id,
    feed_id: feed,
    feed_title: 'Same title',
    published_at: date.toISOString(),
  } as Article;
}

describe('article grouping', () => {
  it('keeps feeds together across appended pages, including feeds with the same title', () => {
    const early = new Date(2026, 8, 12, 9);
    const late = new Date(2026, 8, 12, 10);
    const items = [
      article(1, 2, late),
      article(2, 1, early),
      article(3, 1, late),
      article(4, 2, early),
    ];
    const grouped = orderGroupedArticles(items, 'feed', 'newest');
    expect(grouped.map((item) => item.id)).toEqual([3, 2, 1, 4]);
    expect([...articleGroupStarts(grouped, 'feed')]).toEqual([3, 1]);
    expect(orderGroupedArticles(items, 'feed', 'oldest').map((item) => item.id)).toEqual([
      2, 3, 4, 1,
    ]);
    expect(items.map((item) => item.id)).toEqual([1, 2, 3, 4]);
  });

  it('uses local calendar days at midnight, independently of the UTC date prefix', () => {
    const items = [
      article(1, 1, new Date(2026, 8, 12, 0, 1)),
      article(2, 2, new Date(2026, 8, 12, 23, 59)),
      article(3, 1, new Date(2026, 8, 13, 0, 1)),
    ];
    expect(articleGroupKey(items[0], 'date')).toBe('date:2026-9-12');
    expect(articleGroupKey(items[1], 'date')).toBe('date:2026-9-12');
    const grouped = orderGroupedArticles(items, 'date', 'oldest');
    expect([...articleGroupStarts(grouped, 'date')]).toEqual([1, 3]);
  });

  it('falls back for stale preferences and missing dates without changing ungrouped order', () => {
    expect(parseArticleGroupBy('unexpected')).toBe('none');
    expect(parseArticleGroupBy(null)).toBe('none');
    const items = [{ id: 1, published_at: '' } as Article];
    expect(articleGroupKey(items[0], 'date')).toBe('date:unknown');
    expect(orderGroupedArticles(items, 'none', 'newest')).toBe(items);
    expect(articleGroupStarts(items, 'none').size).toBe(0);
  });
});
