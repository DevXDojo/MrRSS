/**
 * Bilibili utility functions for video URL handling
 */
import { safeVideoUrl } from './video';

const watchPath = /^\/video\/(BV[A-Za-z0-9]+)\/?$/i;
const mobilePlayerPath = '/blackboard/html5mobileplayer.html';

function parseBilibiliUrl(raw: string | undefined): URL | undefined {
  const safe = safeVideoUrl(raw);
  if (!safe) return;
  const url = new URL(safe);
  if (url.hostname === 'player.bilibili.com' && url.pathname === '/player.html') return url;
  if (
    (url.hostname === 'bilibili.com' || url.hostname === 'www.bilibili.com') &&
    (url.pathname === mobilePlayerPath || watchPath.test(url.pathname))
  ) {
    return url;
  }
}

/**
 * Regular expression to extract BVID from Bilibili URLs
 */
const BVID_EXTRACT_PATTERNS = [/bilibili\.com\/video\/(BV[\w]+)/i, /bvid=([\w]+)/i];

/**
 * Check if a URL is a Bilibili video URL
 * @param url - The URL to check
 * @returns True if the URL is a Bilibili video URL
 */
export function isBilibiliUrl(url: string | undefined): boolean {
  return !!parseBilibiliUrl(url);
}

/**
 * Check if an article has a Bilibili video
 * @param article - The article to check
 * @returns True if the article has a Bilibili video
 */
export function isBilibiliArticle(article: { video_url?: string } | undefined): boolean {
  return article ? isBilibiliUrl(article.video_url) : false;
}

/**
 * Extract BVID from various Bilibili URL formats
 * @param url - The Bilibili URL
 * @returns The BVID or empty string if not found
 */
export function extractBilibiliBVID(url: string): string {
  if (!url) return '';

  for (const pattern of BVID_EXTRACT_PATTERNS) {
    const match = url.match(pattern);
    if (match && match[1]) {
      return match[1];
    }
  }

  return '';
}

/**
 * Get Bilibili embed URL from a video URL
 * @param url - The Bilibili video URL or iframe URL
 * @returns The embed URL or original URL if not a Bilibili URL
 */
export function getBilibiliEmbedUrl(url: string): string {
  let player = parseBilibiliUrl(url);
  if (!player) return url;
  const bvid = watchPath.exec(player.pathname)?.[1];
  if (bvid) {
    const part = player.searchParams.get('p');
    player = new URL('https://player.bilibili.com/player.html');
    player.searchParams.set('bvid', bvid);
    if (part) player.searchParams.set('p', part);
  }
  player.searchParams.set('autoplay', '0');
  return player.href;
}

export function bilibiliVideoKey(raw: string | undefined): string | undefined {
  const url = parseBilibiliUrl(raw);
  if (!url) return;
  const params = url.searchParams;
  const episode = params.get('episodeId');
  if (episode) return `episode:${episode}`;
  const id = watchPath.exec(url.pathname)?.[1] || params.get('bvid') || params.get('aid');
  if (!id) return;
  const cid = params.get('cid');
  return cid && /^\d+$/.test(cid) ? `${id}:cid:${cid}` : `${id}:p:${params.get('p') || '1'}`;
}

/**
 * Get Bilibili thumbnail URL from a BVID
 * Note: Bilibili thumbnails are extracted from RSS content, not from BVID alone
 * This function is a placeholder for future implementation
 * @param bvid - The Bilibili video BVID
 * @returns Empty string (thumbnails should be extracted from RSS content)
 */
export function getBilibiliThumbnailUrl(bvid: string): string {
  // Bilibili doesn't provide a simple thumbnail URL pattern like YouTube
  // Thumbnails should be extracted from the RSS feed's image content
  return '';
}

/**
 * Extract Bilibili video URL from RSS content/description
 * @param content - The RSS content or description HTML
 * @returns The Bilibili iframe URL or empty string if not found
 */
export function extractBilibiliVideoFromHTML(content: string): string {
  if (!content) return '';

  // Look for Bilibili iframe in the content
  const iframeRegex =
    /<iframe[^>]+src=["']([^"']*bilibili\.com\/blackboard\/html5mobileplayer\.html[^"']*)["'][^>]*>/i;
  const match = content.match(iframeRegex);

  if (match && match[1]) {
    // Unescape HTML entities
    const unescapedUrl = match[1].replace(/&amp;/g, '&');
    return unescapedUrl;
  }

  return '';
}

/**
 * Get the Bilibili watch page URL from a BVID
 * @param bvid - The Bilibili video BVID
 * @returns The watch page URL
 */
export function getBilibiliWatchUrl(bvid: string): string {
  if (!bvid) return '';
  return `https://www.bilibili.com/video/${bvid}`;
}
