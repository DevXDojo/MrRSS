import { describe, it, expect } from 'vitest';
import {
  isBilibiliUrl,
  isBilibiliArticle,
  extractBilibiliBVID,
  getBilibiliEmbedUrl,
  getBilibiliThumbnailUrl,
  extractBilibiliVideoFromHTML,
  getBilibiliWatchUrl,
} from './bilibili';

describe('Bilibili Utils', () => {
  describe('isBilibiliUrl', () => {
    it('should return true for Bilibili video URLs', () => {
      expect(isBilibiliUrl('https://www.bilibili.com/video/BV11bAUzBEqG')).toBe(true);
      expect(isBilibiliUrl('https://www.bilibili.com/video/BV11bAUzBEqG?param=value')).toBe(true);
      expect(isBilibiliUrl('https://bilibili.com/video/BV11bAUzBEqG')).toBe(true);
    });

    it('should return true for Bilibili player URLs', () => {
      expect(
        isBilibiliUrl(
          'https://www.bilibili.com/blackboard/html5mobileplayer.html?aid=123&bvid=BV11bAUzBEqG'
        )
      ).toBe(true);
    });

    it('should return false for non-Bilibili URLs', () => {
      expect(isBilibiliUrl('https://www.youtube.com/watch?v=dQw4w9WgXcQ')).toBe(false);
      expect(isBilibiliUrl('https://vimeo.com/123456789')).toBe(false);
      expect(isBilibiliUrl('https://example.com/video')).toBe(false);
      expect(isBilibiliUrl('')).toBe(false);
      expect(isBilibiliUrl(undefined as any)).toBe(false);
    });
  });

  describe('isBilibiliArticle', () => {
    it('should return true if article has Bilibili video_url', () => {
      expect(
        isBilibiliArticle({
          video_url:
            'https://www.bilibili.com/blackboard/html5mobileplayer.html?aid=123&bvid=BV11bAUzBEqG',
        })
      ).toBe(true);
      expect(
        isBilibiliArticle({
          video_url: 'https://www.bilibili.com/video/BV11bAUzBEqG',
        })
      ).toBe(true);
    });

    it('should return false if article has non-Bilibili video_url', () => {
      expect(
        isBilibiliArticle({
          video_url: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
        })
      ).toBe(false);
      expect(isBilibiliArticle({ video_url: '' })).toBe(false);
    });

    it('should return false for undefined article', () => {
      expect(isBilibiliArticle(undefined)).toBe(false);
    });

    it('should return false for article without video_url', () => {
      expect(isBilibiliArticle({})).toBe(false);
      expect(isBilibiliArticle({ video_url: undefined })).toBe(false);
    });
  });

  describe('extractBilibiliBVID', () => {
    it('should extract BVID from video URLs', () => {
      expect(extractBilibiliBVID('https://www.bilibili.com/video/BV11bAUzBEqG')).toBe(
        'BV11bAUzBEqG'
      );
      expect(extractBilibiliBVID('https://www.bilibili.com/video/BV11bAUzBEqG?param=value')).toBe(
        'BV11bAUzBEqG'
      );
    });

    it('should extract BVID from player URLs', () => {
      expect(
        extractBilibiliBVID(
          'https://www.bilibili.com/blackboard/html5mobileplayer.html?aid=123&bvid=BV11bAUzBEqG'
        )
      ).toBe('BV11bAUzBEqG');
    });

    it('should return empty string for non-Bilibili URLs', () => {
      expect(extractBilibiliBVID('https://www.youtube.com/watch?v=dQw4w9WgXcQ')).toBe('');
      expect(extractBilibiliBVID('')).toBe('');
    });
  });

  describe('getBilibiliEmbedUrl', () => {
    it('should return original URL if already an html5mobileplayer URL', () => {
      const playerUrl =
        'https://www.bilibili.com/blackboard/html5mobileplayer.html?aid=123&bvid=BV11bAUzBEqG';
      expect(getBilibiliEmbedUrl(playerUrl)).toBe(playerUrl);
    });

    it('should convert video URL to embed URL', () => {
      expect(getBilibiliEmbedUrl('https://www.bilibili.com/video/BV11bAUzBEqG')).toContain(
        'html5mobileplayer.html'
      );
      expect(getBilibiliEmbedUrl('https://www.bilibili.com/video/BV11bAUzBEqG')).toContain(
        'bvid=BV11bAUzBEqG'
      );
    });

    it('should return original URL for non-Bilibili URLs', () => {
      const nonBilibiliUrl = 'https://www.youtube.com/watch?v=dQw4w9WgXcQ';
      expect(getBilibiliEmbedUrl(nonBilibiliUrl)).toBe(nonBilibiliUrl);
    });

    it('should return empty string for empty input', () => {
      expect(getBilibiliEmbedUrl('')).toBe('');
    });
  });

  describe('getBilibiliThumbnailUrl', () => {
    it('should return empty string (thumbnails from RSS)', () => {
      expect(getBilibiliThumbnailUrl('BV11bAUzBEqG')).toBe('');
    });
  });

  describe('extractBilibiliVideoFromHTML', () => {
    it('should extract iframe URL from RSS content', () => {
      const content =
        '<iframe width="640" height="360" src="https://www.bilibili.com/blackboard/html5mobileplayer.html?aid=116142274314455&amp;cid=undefined&amp;bvid=BV11bAUzBEqG" frameborder="0" allowfullscreen=""></iframe>';
      const result = extractBilibiliVideoFromHTML(content);
      expect(result).toContain('html5mobileplayer.html');
      expect(result).toContain('bvid=BV11bAUzBEqG');
      expect(result).not.toContain('&amp;');
    });

    it('should return empty string for content without iframe', () => {
      const content = '<p>Some text without iframe</p>';
      expect(extractBilibiliVideoFromHTML(content)).toBe('');
    });

    it('should return empty string for empty content', () => {
      expect(extractBilibiliVideoFromHTML('')).toBe('');
    });
  });

  describe('getBilibiliWatchUrl', () => {
    it('should generate watch URL from BVID', () => {
      expect(getBilibiliWatchUrl('BV11bAUzBEqG')).toBe(
        'https://www.bilibili.com/video/BV11bAUzBEqG'
      );
    });

    it('should return empty string for empty BVID', () => {
      expect(getBilibiliWatchUrl('')).toBe('');
    });
  });
});
