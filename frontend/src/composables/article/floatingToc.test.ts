import { describe, expect, it } from 'vitest';
import {
  buildTocItems,
  calcTocProgress,
  resolveHeadingDisplayText,
  sanitizeHeadingText,
  shouldShowTocText,
  type TocItem,
} from './floatingToc';

describe('floatingToc helpers', () => {
  it('sanitizes heading text by removing leading # and keeping numbering', () => {
    expect(sanitizeHeadingText('### Hello World')).toBe('Hello World');
    expect(sanitizeHeadingText('  ## 1.2. Heading')).toBe('1.2. Heading');
  });

  it('prefers translated heading text when available', () => {
    expect(resolveHeadingDisplayText('Original title', ' 翻译标题 ')).toBe('翻译标题');
    expect(resolveHeadingDisplayText('Original title', '')).toBe('Original title');
  });

  it('builds hierarchical toc items with fallback support', () => {
    const result = buildTocItems(
      [
        { level: 1, offsetTop: 10, rawText: 'H1', domIndex: 0 },
        { level: 2, offsetTop: 20, rawText: 'H2', domIndex: 1, existingId: 'given-id' },
        { level: 3, offsetTop: 30, rawText: 'H3', domIndex: 2 },
      ],
      42
    );

    expect(result.items).toHaveLength(3);
    expect(result.items[0].parentIndex).toBeNull();
    expect(result.items[1].parentIndex).toBe(0);
    expect(result.items[2].parentIndex).toBe(1);
    expect(result.items[1].id).toBe('given-id');
    expect(result.generatedIds).toEqual([
      { domIndex: 0, id: 'toc-heading-42-0' },
      { domIndex: 2, id: 'toc-heading-42-2' },
    ]);

    const fallback = buildTocItems([], 7);
    expect(fallback.items).toHaveLength(1);
    expect(fallback.items[0].isFallback).toBe(true);
    expect(fallback.items[0].id).toBe('toc-fallback-7');
  });

  it('promotes heading levels when h1 or h1/h2 are missing', () => {
    const noH1 = buildTocItems(
      [
        { level: 2, offsetTop: 10, rawText: 'H2 as top', domIndex: 0 },
        { level: 3, offsetTop: 20, rawText: 'H3 child', domIndex: 1 },
      ],
      99
    );
    expect(noH1.items[0].level).toBe(1);
    expect(noH1.items[1].level).toBe(2);
    expect(noH1.items[1].parentIndex).toBe(0);

    const onlyH3 = buildTocItems(
      [
        { level: 3, offsetTop: 10, rawText: 'H3 promoted', domIndex: 0 },
        { level: 3, offsetTop: 20, rawText: 'H3 sibling', domIndex: 1 },
      ],
      100
    );
    expect(onlyH3.items[0].level).toBe(1);
    expect(onlyH3.items[1].level).toBe(1);
    expect(onlyH3.items[0].parentIndex).toBeNull();
    expect(onlyH3.items[1].parentIndex).toBeNull();
  });

  it('shows only current + ancestor path when not hovered', () => {
    const items: TocItem[] = [
      {
        id: 'a',
        text: 'A',
        level: 1,
        offsetTop: 0,
        markerWidth: 34,
        parentIndex: null,
        isFallback: false,
      },
      {
        id: 'b',
        text: 'B',
        level: 2,
        offsetTop: 10,
        markerWidth: 24,
        parentIndex: 0,
        isFallback: false,
      },
      {
        id: 'c',
        text: 'C',
        level: 3,
        offsetTop: 20,
        markerWidth: 14,
        parentIndex: 1,
        isFallback: false,
      },
      {
        id: 'd',
        text: 'D',
        level: 2,
        offsetTop: 30,
        markerWidth: 24,
        parentIndex: 0,
        isFallback: false,
      },
    ];

    // active is index 2 (C), path should be A -> B -> C
    expect(shouldShowTocText(0, 2, items)).toBe(true);
    expect(shouldShowTocText(1, 2, items)).toBe(true);
    expect(shouldShowTocText(2, 2, items)).toBe(true);
    expect(shouldShowTocText(3, 2, items)).toBe(false);
  });

  it('calculates active item, section progress and article progress', () => {
    const items: TocItem[] = [
      {
        id: 'h1',
        text: 'H1',
        level: 1,
        offsetTop: 0,
        markerWidth: 34,
        parentIndex: null,
        isFallback: false,
      },
      {
        id: 'h2',
        text: 'H2',
        level: 2,
        offsetTop: 100,
        markerWidth: 24,
        parentIndex: 0,
        isFallback: false,
      },
      {
        id: 'h3',
        text: 'H3',
        level: 3,
        offsetTop: 200,
        markerWidth: 14,
        parentIndex: 1,
        isFallback: false,
      },
    ];

    const result = calcTocProgress(150, 1000, 200, items);
    expect(result.activeIndex).toBe(1);
    expect(result.sectionProgress).toBeGreaterThanOrEqual(60);
    expect(result.sectionProgress).toBeLessThanOrEqual(65);
    expect(result.articleProgress).toBe(19);
  });
});
