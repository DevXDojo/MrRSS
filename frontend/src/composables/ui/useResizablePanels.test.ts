import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { useResizablePanels } from './useResizablePanels';

function setup() {
  let panels!: ReturnType<typeof useResizablePanels>;
  const wrapper = mount({
    setup() {
      panels = useResizablePanels();
    },
    template: '<div />',
  });
  return { panels, wrapper };
}

afterEach(() => {
  localStorage.removeItem('articleListWidth:normal');
  localStorage.removeItem('articleListWidth:compact');
  vi.restoreAllMocks();
});

describe('article list width', () => {
  it('remembers independent normal and compact widths across remounts', () => {
    const { panels, wrapper } = setup();
    panels.startResizeArticleList(new MouseEvent('mousedown', { clientX: 500 }));
    window.dispatchEvent(new MouseEvent('mousemove', { clientX: 610 }));
    window.dispatchEvent(new MouseEvent('mouseup'));
    expect(localStorage.getItem('articleListWidth:normal')).toBe('460');
    panels.setCompactMode(true);
    panels.startResizeArticleList(new MouseEvent('mousedown', { clientX: 500 }));
    window.dispatchEvent(new MouseEvent('mousemove', { clientX: 700 }));
    window.dispatchEvent(new Event('blur'));
    expect(localStorage.getItem('articleListWidth:compact')).toBe('700');
    wrapper.unmount();
    const remounted = setup();
    expect(remounted.panels.articleListWidth.value).toBe(460);
    remounted.panels.setCompactMode(true);
    expect(remounted.panels.articleListWidth.value).toBe(700);
    remounted.panels.setCompactMode(false);
    expect(remounted.panels.articleListWidth.value).toBe(460);
    remounted.wrapper.unmount();
  });

  it('bounds stored widths and restores document styles on unmount', () => {
    localStorage.setItem('articleListWidth:normal', '9000');
    localStorage.setItem('articleListWidth:compact', 'NaN');
    const { panels, wrapper } = setup();
    expect(panels.articleListWidth.value).toBe(600);
    panels.setCompactMode(true);
    expect(panels.articleListWidth.value).toBe(500);
    const cursor = document.body.style.cursor;
    panels.startResizeArticleList(new MouseEvent('mousedown'));
    wrapper.unmount();
    expect(document.body.style.cursor).toBe(cursor);
    window.dispatchEvent(new MouseEvent('mousemove', { clientX: 200 }));
    expect(panels.articleListWidth.value).toBe(500);
  });
});
