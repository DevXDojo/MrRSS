import { describe, expect, it } from 'vitest';
import { isVideoArticle, safeVideoUrl } from './video';
import { isYouTubeUrl } from './youtube';
import { isBilibiliUrl } from './bilibili';

describe('generic videos', () => {
  it('accepts signed extensionless HTTP media and rejects active or local URLs', () => {
    expect(isVideoArticle({ video_url: 'https://example.org/media?token=abc' })).toBe(true);
    for (const url of ['javascript:alert(1)', 'file:///secret', 'data:text/html,x', 'https://user:pass@example.org/v']) {
      expect(safeVideoUrl(url)).toBe('');
    }
    expect(isVideoArticle({})).toBe(false);
  });
  it('does not classify untrusted origins as embedded platforms', () => {
    expect(isYouTubeUrl('https://evil.org/youtube.com/embed/abc')).toBe(false);
    expect(isBilibiliUrl('https://evilbilibili.com/blackboard/html5mobileplayer.html')).toBe(false);
  });
});
