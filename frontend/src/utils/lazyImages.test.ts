import { describe, expect, it } from 'vitest';
import { withLazyImages } from './lazyImages';

describe('withLazyImages', () => {
  it('adds lazy loading hints to images', () => {
    const html = '<p>text</p><img src="https://example.com/a.png" alt="a">';
    const result = withLazyImages(html);

    expect(result).toContain('loading="lazy"');
    expect(result).toContain('decoding="async"');
    expect(result).toContain('src="https://example.com/a.png"');
    expect(result).toContain('alt="a"');
  });

  it('adds hints to every image in the document', () => {
    const html = '<div><img src="a.png"><img src="b.png"></div>';
    const result = withLazyImages(html);

    expect(result.match(/loading="lazy"/g)?.length).toBe(2);
    expect(result.match(/decoding="async"/g)?.length).toBe(2);
  });

  it('preserves an existing loading attribute', () => {
    const html = '<img src="a.png" loading="eager">';
    const result = withLazyImages(html);

    expect(result).toContain('loading="eager"');
    expect(result).not.toContain('loading="lazy"');
  });

  it('preserves an existing decoding attribute', () => {
    const html = '<img src="a.png" decoding="sync">';
    const result = withLazyImages(html);

    expect(result).toContain('decoding="sync"');
    expect(result).not.toContain('decoding="async"');
  });

  it('returns HTML without images unchanged', () => {
    const html = '<p>no images here</p>';

    expect(withLazyImages(html)).toBe(html);
  });

  it('returns empty input unchanged', () => {
    expect(withLazyImages('')).toBe('');
  });

  it('keeps surrounding structure intact', () => {
    const html = '<ul><li><a href="x"><img src="i.png"></a></li></ul>';
    const result = withLazyImages(html);

    expect(result).toContain('<ul>');
    expect(result).toContain('<a href="x">');
    expect(result).toContain('loading="lazy"');
  });
});
