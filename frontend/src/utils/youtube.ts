/**
 * YouTube utility functions for video URL handling
 */

/**
 * Check if a URL is a YouTube video URL
 * @param url - The URL to check
 * @returns True if the URL is a YouTube video URL
 */
export function isYouTubeUrl(url: string | undefined): boolean {
  if (!url) return false;

  const patterns = [/youtube\.com\/watch\?v=/i, /youtu\.be\//i, /youtube\.com\/embed\//i];

  return patterns.some((pattern) => pattern.test(url));
}

/**
 * Check if an article has a YouTube video
 * @param article - The article to check
 * @returns True if the article has a YouTube video
 */
export function isYouTubeArticle(article: { video_url?: string } | undefined): boolean {
  return article ? isYouTubeUrl(article.video_url) : false;
}

/**
 * Extract YouTube video ID from various YouTube URL formats
 * @param url - The YouTube URL
 * @returns The YouTube video ID or empty string if not found
 */
export function extractYouTubeVideoId(url: string): string {
  if (!url) return '';

  const patterns = [
    /youtube\.com\/watch\?v=([^&]+)/i,
    /youtu\.be\/([^?&]+)/i,
    /youtube\.com\/embed\/([^?&]+)/i,
  ];

  for (const pattern of patterns) {
    const match = url.match(pattern);
    if (match && match[1]) {
      return match[1];
    }
  }

  return '';
}

/**
 * Get YouTube thumbnail URL for a video
 * @param videoId - The YouTube video ID
 * @param quality - The thumbnail quality ('default', 'medium', 'high', 'maxres')
 * @returns The thumbnail URL
 */
export function getYouTubeThumbnailUrl(
  videoId: string,
  quality: 'default' | 'medium' | 'high' | 'maxres' = 'high'
): string {
  const qualityMap = {
    default: 'default',
    medium: 'mqdefault',
    high: 'hqdefault',
    maxres: 'maxresdefault',
  };

  return `https://img.youtube.com/vi/${videoId}/${qualityMap[quality]}.jpg`;
}

/**
 * Get YouTube embed URL from a video URL
 * @param url - The YouTube video URL
 * @returns The embed URL or original URL if not a YouTube URL
 */
export function getYouTubeEmbedUrl(url: string): string {
  const videoId = extractYouTubeVideoId(url);
  if (videoId) {
    return `https://www.youtube.com/embed/${videoId}`;
  }
  return url;
}
