import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref } from 'vue';
import { clampArticleTableSplit, useArticleTableSplit } from './useArticleTableSplit';

afterEach(() => {
  localStorage.removeItem('articleTableSplit');
  vi.restoreAllMocks();
});

describe('article table split', () => {
  it('keeps both panes available even with an invalid saved ratio', () => {
    expect(clampArticleTableSplit(NaN)).toBe(40);
    expect(clampArticleTableSplit(-20)).toBe(25);
    expect(clampArticleTableSplit(120)).toBe(75);
  });

  it('persists keyboard resizing and releases pointer listeners on unmount', async () => {
    localStorage.setItem('articleTableSplit', '40');
    let split!: ReturnType<typeof useArticleTableSplit>;
    const wrapper = mount({
      setup() {
        const container = ref<HTMLElement | null>(null);
        split = useArticleTableSplit(container);
        return { container };
      },
      template: '<div ref="container" />',
    });
    const previousCursor = document.body.style.cursor;
    const previousSelect = document.body.style.userSelect;
    split.keydown(new KeyboardEvent('keydown', { key: 'ArrowDown' }));
    expect(split.split.value).toBe(45);
    expect(localStorage.getItem('articleTableSplit')).toBe('45');
    const removeListener = vi.spyOn(window, 'removeEventListener');
    split.start(new MouseEvent('pointerdown', { button: 0 }) as PointerEvent);
    expect(split.resizing.value).toBe(true);
    wrapper.unmount();
    expect(split.resizing.value).toBe(false);
    expect(document.body.style.cursor).toBe(previousCursor);
    expect(document.body.style.userSelect).toBe(previousSelect);
    expect(removeListener).toHaveBeenCalledWith('pointermove', expect.any(Function));
    expect(removeListener).toHaveBeenCalledWith('pointerup', expect.any(Function));
  });
});
