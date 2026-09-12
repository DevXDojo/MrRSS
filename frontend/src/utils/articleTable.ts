export const articleTableColumns = ['title', 'feed', 'author', 'date', 'status'] as const;
export type ArticleTableColumn = (typeof articleTableColumns)[number];
const defaults: ArticleTableColumn[] = ['title', 'feed', 'date', 'status'];

export function parseArticleTableColumns(value: string): ArticleTableColumn[] {
  try {
    const parsed: unknown = JSON.parse(value);
    if (!Array.isArray(parsed)) return [...defaults];
    // Keep the title available for opening articles, and discard unknown keys.
    return articleTableColumns.filter((column) => column === 'title' || parsed.includes(column));
  } catch {
    return [...defaults];
  }
}
