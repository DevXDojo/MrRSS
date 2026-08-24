export interface TocItem {
  id: string;
  text: string;
  level: 1 | 2 | 3;
  offsetTop: number;
  markerWidth: number;
  parentIndex: number | null;
  isFallback: boolean;
}

export interface HeadingSnapshot {
  level: 1 | 2 | 3;
  offsetTop: number;
  rawText: string;
  translatedText?: string;
  existingId?: string;
  domIndex: number;
}

export interface TocProgressResult {
  activeIndex: number;
  sectionProgress: number;
  articleProgress: number;
}

export function getMarkerWidth(level: 1 | 2 | 3): number {
  if (level === 1) return 34;
  if (level === 2) return 24;
  return 14;
}

export function sanitizeHeadingText(text: string): string {
  // Keep numbered prefixes like "1.2.", but remove markdown heading markers like "###".
  return text
    .replace(/^\s*#+\s*/, '')
    .replace(/\s+/g, ' ')
    .trim();
}

export function resolveHeadingDisplayText(rawText: string, translatedText?: string): string {
  const normalizedTranslated = sanitizeHeadingText(translatedText || '');
  if (normalizedTranslated) return normalizedTranslated;
  return sanitizeHeadingText(rawText);
}

export function buildTocItems(
  headings: HeadingSnapshot[],
  articleId: number
): { items: TocItem[]; generatedIds: Array<{ domIndex: number; id: string }> } {
  const items: TocItem[] = [];
  const generatedIds: Array<{ domIndex: number; id: string }> = [];
  const minLevel = headings.reduce<1 | 2 | 3>((min, heading) => {
    return heading.level < min ? heading.level : min;
  }, 3);

  let lastH1Index: number | null = null;
  let lastH2Index: number | null = null;

  for (const heading of headings) {
    const text = resolveHeadingDisplayText(heading.rawText, heading.translatedText);
    if (!text) continue;

    const id = heading.existingId || `toc-heading-${articleId}-${heading.domIndex}`;
    if (!heading.existingId) {
      generatedIds.push({ domIndex: heading.domIndex, id });
    }

    const normalizedLevel = Math.max(1, heading.level - (minLevel - 1)) as 1 | 2 | 3;

    let parentIndex: number | null = null;
    if (normalizedLevel === 2) {
      parentIndex = lastH1Index;
    } else if (normalizedLevel === 3) {
      parentIndex = lastH2Index ?? lastH1Index;
    }

    items.push({
      id,
      text,
      level: normalizedLevel,
      offsetTop: Math.max(0, Math.round(heading.offsetTop)),
      markerWidth: getMarkerWidth(normalizedLevel),
      parentIndex,
      isFallback: false,
    });

    const itemIndex = items.length - 1;
    if (normalizedLevel === 1) {
      lastH1Index = itemIndex;
      lastH2Index = null;
    } else if (normalizedLevel === 2) {
      lastH2Index = itemIndex;
    }
  }

  if (items.length === 0) {
    items.push({
      id: `toc-fallback-${articleId}`,
      text: '',
      level: 1,
      offsetTop: 0,
      markerWidth: getMarkerWidth(1),
      parentIndex: null,
      isFallback: true,
    });
  }

  return { items, generatedIds };
}

export function shouldShowTocText(
  itemIndex: number,
  activeIndex: number,
  items: TocItem[]
): boolean {
  const item = items[itemIndex];
  if (!item || item.isFallback) return false;
  if (activeIndex < 0 || activeIndex >= items.length) return false;
  if (itemIndex === activeIndex) return true;

  let parent = items[activeIndex].parentIndex;
  while (parent !== null) {
    if (parent === itemIndex) return true;
    parent = items[parent].parentIndex;
  }
  return false;
}

export function calcArticleProgress(
  scrollTop: number,
  scrollHeight: number,
  clientHeight: number
): number {
  const maxScrollTop = Math.max(0, scrollHeight - clientHeight);
  if (maxScrollTop <= 0) return 100;
  const progress = Math.max(0, Math.min(1, scrollTop / maxScrollTop));
  return Math.round(progress * 100);
}

export function calcTocProgress(
  scrollTop: number,
  scrollHeight: number,
  clientHeight: number,
  items: TocItem[],
  pointerOffset: number = 12
): TocProgressResult {
  const articleProgress = calcArticleProgress(scrollTop, scrollHeight, clientHeight);
  if (items.length === 0) {
    return { activeIndex: -1, sectionProgress: 0, articleProgress };
  }

  const pointer = scrollTop + pointerOffset;
  let activeIndex = 0;
  for (let i = 0; i < items.length; i += 1) {
    if (items[i].offsetTop <= pointer) {
      activeIndex = i;
    } else {
      break;
    }
  }

  const start = items[activeIndex].offsetTop;
  const maxScrollTop = Math.max(0, scrollHeight - clientHeight);
  const end =
    activeIndex < items.length - 1
      ? items[activeIndex + 1].offsetTop
      : Math.max(start + 1, maxScrollTop);
  const segmentSize = Math.max(1, end - start);
  const sectionProgress = Math.round(
    Math.max(0, Math.min(1, (pointer - start) / segmentSize)) * 100
  );

  return { activeIndex, sectionProgress, articleProgress };
}
