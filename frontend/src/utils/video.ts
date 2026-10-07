export function safeVideoUrl(raw: string | undefined): string {
  if (!raw) return '';
  try {
    const url = new URL(raw);
    return ['http:', 'https:'].includes(url.protocol) && !url.username && !url.password
      ? url.href
      : '';
  } catch {
    return '';
  }
}

export function isVideoArticle(article: { video_url?: string } | null | undefined): boolean {
  return !!safeVideoUrl(article?.video_url);
}
