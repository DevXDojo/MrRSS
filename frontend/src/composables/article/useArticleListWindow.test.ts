import { beforeEach, describe, expect, it } from 'vitest';
import { mount } from '@vue/test-utils';
import { computed, nextTick, ref } from 'vue';
import { useArticleListWindow } from './useArticleListWindow';
import type { Article } from '@/types/models';

function makeArticles(count: number): Article[] {
  return Array.from(
    { length: count },
    (_, index) => ({ id: index + 1, title: `a${index}` }) as Article
  );
}

function mountWindow(count: number, scrollTop = 0, clientHeight = 800, enabled = ref(true)) {
  const items = ref<Article[]>(makeArticles(count));
  const container = document.createElement('div');
  Object.defineProperty(container, 'scrollTop', { value: scrollTop, writable: true });
  Object.defineProperty(container, 'clientHeight', { value: clientHeight, writable: true });

  let api!: ReturnType<typeof useArticleListWindow>;
  const wrapper = mount({
    setup() {
      api = useArticleListWindow(
        computed(() => items.value),
        ref(container as unknown as HTMLElement),
        { enabled }
      );
      return () => null;
    },
  });
  return { items, container, api, wrapper };
}

describe('article list windowing', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
  });

  it('renders short lists in full without spacers', () => {
    const { api } = mountWindow(120);
    expect(api.isVirtualized.value).toBe(false);
    expect(api.windowItems.value).toHaveLength(120);
    expect(api.topSpacerHeight.value).toBe(0);
    expect(api.bottomSpacerHeight.value).toBe(0);
  });

  it('preserves complete rendering for layouts that cannot be windowed', async () => {
    const enabled = ref(true);
    const { api } = mountWindow(3000, 9600, 800, enabled);
    api.updateFromScroll();
    enabled.value = false;
    await nextTick();
    expect(api.windowItems.value).toHaveLength(3000);
    expect(api.topSpacerHeight.value).toBe(0);
    expect(api.bottomSpacerHeight.value).toBe(0);
  });

  it('caps rendered rows for long lists and keeps the window near the viewport', () => {
    const { api } = mountWindow(3000);
    api.updateFromScroll();
    expect(api.isVirtualized.value).toBe(true);
    expect(api.windowItems.value.length).toBeLessThanOrEqual(160);
    expect(api.windowItems.value.length).toBeGreaterThanOrEqual(60);
    expect(api.windowItems.value[0].id).toBe(1);
  });

  it('moves the window when scrolled and reports spacer heights', () => {
    const { api, container } = mountWindow(3000);
    api.updateFromScroll();

    container.scrollTop = 9600; // ≈ row 100 at the 96px fallback height
    api.updateFromScroll();

    const firstId = api.windowItems.value[0].id;
    expect(firstId).toBeGreaterThan(80);
    expect(firstId).toBeLessThan(100);
    expect(api.topSpacerHeight.value).toBeGreaterThan(7000);
    expect(api.bottomSpacerHeight.value).toBeGreaterThan(0);
  });

  it('brings a distant article into the window on request', async () => {
    const { api } = mountWindow(3000);
    api.updateFromScroll();
    expect(api.windowItems.value.some((article) => article.id === 2500)).toBe(false);

    await api.ensureArticleVisible(2500);
    expect(api.windowItems.value.some((article) => article.id === 2500)).toBe(true);
  });

  it('uses measured row heights to place the window', () => {
    const { api, container } = mountWindow(3000);
    container.scrollTop = 0;
    api.updateFromScroll();

    // Two mounted rows measuring 200px each; the scroller is the real container
    // so heights can be fed to the composable for later scroll maths.
    const rows = [2000, 2001].map((id) => {
      const el = document.createElement('div');
      el.dataset.articleId = String(id);
      return el;
    });
    rows[0].getBoundingClientRect = () => ({ top: 0, height: 200 }) as DOMRect;
    rows[1].getBoundingClientRect = () => ({ top: 200, height: 200 }) as DOMRect;
    rows.forEach((row) => container.appendChild(row));
    Object.defineProperty(rows[1], 'offsetHeight', { value: 200 });
    api.measureMountedRows();

    container.scrollTop = 20000; // 100 measured rows at 200px
    api.updateFromScroll();
    const firstId = api.windowItems.value[0].id;
    expect(firstId).toBeGreaterThan(80);
    expect(firstId).toBeLessThan(100);
  });

  it('keeps the window valid when the list shrinks past the current range', async () => {
    const { api, container, items } = mountWindow(3000);
    api.updateFromScroll();
    container.scrollTop = 20000;
    api.updateFromScroll();
    expect(api.rangeStart.value).toBeGreaterThan(100);

    items.value = makeArticles(300);
    await nextTick();

    expect(api.windowItems.value.length).toBeGreaterThan(0);
    expect(Number.isFinite(api.topSpacerHeight.value)).toBe(true);
    expect(Number.isFinite(api.bottomSpacerHeight.value)).toBe(true);
    expect(api.bottomSpacerHeight.value).toBeGreaterThanOrEqual(0);
  });

  it('splits a multi-column row height between the cards in that row', () => {
    const { api, container } = mountWindow(3000);
    api.updateFromScroll();

    // Card mode mounts two cards side by side (same top) per grid row.
    const rows = [1, 2, 3, 4].map((id) => {
      const el = document.createElement('div');
      el.dataset.articleId = String(id);
      return el;
    });
    rows[0].getBoundingClientRect = () => ({ top: 0, height: 120 }) as DOMRect;
    rows[1].getBoundingClientRect = () => ({ top: 0, height: 120 }) as DOMRect;
    rows[2].getBoundingClientRect = () => ({ top: 120, height: 120 }) as DOMRect;
    rows[3].getBoundingClientRect = () => ({ top: 120, height: 120 }) as DOMRect;
    rows.forEach((row) => container.appendChild(row));
    Object.defineProperty(rows[3], 'offsetHeight', { value: 120 });
    api.measureMountedRows();

    container.scrollTop = 6000; // 100 items at the measured 60px per card
    api.updateFromScroll();
    const firstId = api.windowItems.value[0].id;
    expect(firstId).toBeGreaterThan(80);
    expect(firstId).toBeLessThan(100);
    expect(api.topSpacerHeight.value).toBeLessThan(6000);
  });

  it('leaves the scroll offset alone when a window slide keeps the anchor in place', async () => {
    const { api, container } = mountWindow(3000);
    api.updateFromScroll();

    const anchor = document.createElement('div');
    anchor.dataset.articleId = '1';
    anchor.getBoundingClientRect = () => ({ top: 100, height: 96 }) as DOMRect;
    container.appendChild(anchor);

    container.scrollTop = 9600;
    api.updateFromScroll();
    await nextTick();
    await nextTick();

    expect(container.scrollTop).toBe(9600);
  });

  it('corrects the scroll offset by the drift of the anchor row', async () => {
    const { api, container } = mountWindow(3000);
    api.updateFromScroll();

    let anchorTop = 100;
    const anchor = document.createElement('div');
    anchor.dataset.articleId = '1';
    anchor.getBoundingClientRect = () => ({ top: anchorTop, height: 96 }) as DOMRect;
    container.appendChild(anchor);

    container.scrollTop = 9600;
    api.updateFromScroll();
    anchorTop = 150; // the height model drifted by 50px before the DOM update ran
    await nextTick();
    await nextTick();

    expect(container.scrollTop).toBe(9650);
  });
});
