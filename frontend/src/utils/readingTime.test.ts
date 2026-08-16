import { describe, expect, it } from 'vitest';
import { estimateReadingTime } from './readingTime';

describe('estimateReadingTime', () => {
  it('counts English words without HTML tags', () => {
    const result = estimateReadingTime(`<nav>Menu</nav><p>${'word '.repeat(440)}</p>`);
    expect(result.minutes).toBe(2);
    expect(result.latinWords).toBe(440);
  });

  it('counts Chinese characters', () => {
    expect(estimateReadingTime(`<p>${'中文'.repeat(500)}</p>`).minutes).toBe(2);
  });

  it('combines CJK and Latin reading effort', () => {
    const result = estimateReadingTime(`${'内容'.repeat(250)} ${'word '.repeat(220)}`);
    expect(result.minutes).toBe(2);
  });

  it('returns a one-minute minimum for empty content', () => {
    expect(estimateReadingTime('')).toEqual({ minutes: 1, latinWords: 0, cjkCharacters: 0 });
  });

  it('returns one minute for very short content', () => {
    expect(estimateReadingTime('<p>Hello 世界</p>').minutes).toBe(1);
  });
});
