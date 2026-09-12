import { describe, expect, it } from 'vitest';
import { parseArticleTableColumns } from './articleTable';

describe('article table columns', () => {
  it('keeps the article title available when every optional column is hidden', () => {
    expect(parseArticleTableColumns('[]')).toEqual(['title']);
    expect(parseArticleTableColumns('["date"]')).toEqual(['title', 'date']);
  });
  it('discards stale, duplicate, and non-string column preferences', () => {
    expect(parseArticleTableColumns('["author","author","removed",null,42,"feed"]')).toEqual([
      'title',
      'feed',
      'author',
    ]);
  });
  it('uses defaults for invalid stored preferences', () => {
    for (const value of ['', 'not json', '{}', 'null']) {
      expect(parseArticleTableColumns(value)).toEqual(['title', 'feed', 'date', 'status']);
    }
  });
});
