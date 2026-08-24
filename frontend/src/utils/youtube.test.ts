import { describe, it, expect } from 'vitest';
import {
  isYouTubeUrl,
  isYouTubeArticle,
  extractYouTubeVideoId,
  getYouTubeThumbnailUrl,
  getYouTubeEmbedUrl,
} from './youtube';

describe('YouTube Utils', () => {
  describe('isYouTubeUrl', () => {
    it('should return true for YouTube watch URLs', () => {
      expect(isYouTubeUrl('https://www.youtube.com/watch?v=dQw4w9WgXcQ')).toBe(true);
      expect(isYouTubeUrl('http://youtube.com/watch?v=dQw4w9WgXcQ')).toBe(true);
      expect(isYouTubeUrl('https://m.youtube.com/watch?v=dQw4w9WgXcQ')).toBe(true);
    });

    it('should return true for YouTube short URLs', () => {
      expect(isYouTubeUrl('https://youtu.be/dQw4w9WgXcQ')).toBe(true);
      expect(isYouTubeUrl('http://youtu.be/dQw4w9WgXcQ')).toBe(true);
    });

    it('should return true for YouTube embed URLs', () => {
      expect(isYouTubeUrl('https://www.youtube.com/embed/dQw4w9WgXcQ')).toBe(true);
      expect(isYouTubeUrl('https://www.youtube.com/embed/dQw4w9WgXcQ?autoplay=1')).toBe(true);
    });

    it('should return false for non-YouTube URLs', () => {
      expect(isYouTubeUrl('https://vimeo.com/123456789')).toBe(false);
      expect(isYouTubeUrl('https://example.com/video')).toBe(false);
      expect(isYouTubeUrl('')).toBe(false);
      expect(isYouTubeUrl(undefined as any)).toBe(false);
    });
  });

  describe('isYouTubeArticle', () => {
    it('should return true if article has YouTube video_url', () => {
      expect(isYouTubeArticle({ video_url: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' })).toBe(
        true
      );
      expect(isYouTubeArticle({ video_url: 'https://youtu.be/dQw4w9WgXcQ' })).toBe(true);
      expect(isYouTubeArticle({ video_url: 'https://www.youtube.com/embed/dQw4w9WgXcQ' })).toBe(
        true
      );
    });

    it('should return false if article has non-YouTube video_url', () => {
      expect(isYouTubeArticle({ video_url: 'https://vimeo.com/123456789' })).toBe(false);
      expect(isYouTubeArticle({ video_url: '' })).toBe(false);
    });

    it('should return false for undefined article', () => {
      expect(isYouTubeArticle(undefined)).toBe(false);
    });

    it('should return false for article without video_url', () => {
      expect(isYouTubeArticle({})).toBe(false);
      expect(isYouTubeArticle({ video_url: undefined })).toBe(false);
    });
  });

  describe('extractYouTubeVideoId', () => {
    it('should extract video ID from watch URLs', () => {
      expect(extractYouTubeVideoId('https://www.youtube.com/watch?v=dQw4w9WgXcQ')).toBe(
        'dQw4w9WgXcQ'
      );
      expect(extractYouTubeVideoId('https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=10s')).toBe(
        'dQw4w9WgXcQ'
      );
    });

    it('should extract video ID from short URLs', () => {
      expect(extractYouTubeVideoId('https://youtu.be/dQw4w9WgXcQ')).toBe('dQw4w9WgXcQ');
      expect(extractYouTubeVideoId('https://youtu.be/dQw4w9WgXcQ?t=10')).toBe('dQw4w9WgXcQ');
    });

    it('should extract video ID from embed URLs', () => {
      expect(extractYouTubeVideoId('https://www.youtube.com/embed/dQw4w9WgXcQ')).toBe(
        'dQw4w9WgXcQ'
      );
      expect(extractYouTubeVideoId('https://www.youtube.com/embed/dQw4w9WgXcQ?autoplay=1')).toBe(
        'dQw4w9WgXcQ'
      );
    });

    it('should return empty string for non-YouTube URLs', () => {
      expect(extractYouTubeVideoId('https://vimeo.com/123456789')).toBe('');
      expect(extractYouTubeVideoId('')).toBe('');
    });
  });

  describe('getYouTubeThumbnailUrl', () => {
    const videoId = 'dQw4w9WgXcQ';

    it('should generate default thumbnail URL', () => {
      expect(getYouTubeThumbnailUrl(videoId, 'default')).toBe(
        'https://img.youtube.com/vi/dQw4w9WgXcQ/default.jpg'
      );
    });

    it('should generate medium thumbnail URL', () => {
      expect(getYouTubeThumbnailUrl(videoId, 'medium')).toBe(
        'https://img.youtube.com/vi/dQw4w9WgXcQ/mqdefault.jpg'
      );
    });

    it('should generate high quality thumbnail URL', () => {
      expect(getYouTubeThumbnailUrl(videoId, 'high')).toBe(
        'https://img.youtube.com/vi/dQw4w9WgXcQ/hqdefault.jpg'
      );
    });

    it('should generate maximum resolution thumbnail URL', () => {
      expect(getYouTubeThumbnailUrl(videoId, 'maxres')).toBe(
        'https://img.youtube.com/vi/dQw4w9WgXcQ/maxresdefault.jpg'
      );
    });

    it('should default to high quality', () => {
      expect(getYouTubeThumbnailUrl(videoId)).toBe(
        'https://img.youtube.com/vi/dQw4w9WgXcQ/hqdefault.jpg'
      );
    });
  });

  describe('getYouTubeEmbedUrl', () => {
    it('should convert watch URL to embed URL', () => {
      expect(getYouTubeEmbedUrl('https://www.youtube.com/watch?v=dQw4w9WgXcQ')).toBe(
        'https://www.youtube.com/embed/dQw4w9WgXcQ'
      );
    });

    it('should convert short URL to embed URL', () => {
      expect(getYouTubeEmbedUrl('https://youtu.be/dQw4w9WgXcQ')).toBe(
        'https://www.youtube.com/embed/dQw4w9WgXcQ'
      );
    });

    it('should return original URL for non-YouTube URLs', () => {
      const nonYouTubeUrl = 'https://vimeo.com/123456789';
      expect(getYouTubeEmbedUrl(nonYouTubeUrl)).toBe(nonYouTubeUrl);
    });

    it('should return empty string for empty input', () => {
      expect(getYouTubeEmbedUrl('')).toBe('');
    });
  });
});
