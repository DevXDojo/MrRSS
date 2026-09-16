export interface ReportSource {
  id: number;
  article_id: number;
  title: string;
  url: string;
  feed: string;
  kind: 'cached' | 'rss_excerpt' | 'missing';
  truncated: boolean;
  characters: number;
  excerpt: string;
}

export interface ReadingReportResult {
  model: string;
  sources: ReportSource[];
  report: {
    overview: string;
    topics: { title: string; summary: string; source_ids: number[] }[];
    reading_order: { source_id: number; reason: string }[];
    caveats: string[];
  };
}
