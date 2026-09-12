import { DOMWrapper, mount } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';
import ArticleListMoreMenu from './ArticleListMoreMenu.vue';
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));

function createMenu() {
  return mount(ArticleListMoreMenu, {
    props: { sortOrder: 'newest', groupBy: 'none', filterCount: 2 },
    attachTo: document.body,
  });
}
function menuPanel() {
  return new DOMWrapper(document.body.querySelector<HTMLElement>('[role="dialog"]')!);
}

afterEach(() => {
  vi.restoreAllMocks();
  document.body.innerHTML = '';
});

describe('article list more menu', () => {
  it('applies sorting and grouping without closing, and opens filters as a separate action', async () => {
    const wrapper = createMenu();
    try {
      const trigger = wrapper.get('button');
      expect(document.querySelector('[role="dialog"]')).toBeNull();
      await trigger.trigger('click');
      const panel = menuPanel();
      expect(panel.findAll('input[type="radio"]')).toHaveLength(5);
      expect(panel.get('input[value="newest"]').element).toBe(document.activeElement);
      await panel.get('input[value="oldest"]').setValue(true);
      expect(wrapper.emitted('sort')).toEqual([['oldest']]);
      await wrapper.setProps({ sortOrder: 'oldest' });
      await panel.get('input[value="feed"]').setValue(true);
      expect(wrapper.emitted('group')).toEqual([['feed']]);
      await wrapper.setProps({ groupBy: 'feed' });
      expect(panel.findAll('input:checked').map((input) => input.attributes('value'))).toEqual([
        'oldest',
        'feed',
      ]);
      expect(trigger.attributes('aria-expanded')).toBe('true');
      expect(panel.get('button').text()).toContain('2');
      await panel.get('button').trigger('click');
      expect(wrapper.emitted('filter')).toHaveLength(1);
      expect(document.querySelector('[role="dialog"]')).toBeNull();
    } finally {
      wrapper.unmount();
    }
  });

  it('supports keyboard opening and Escape, and dismisses when focus or pointer moves outside', async () => {
    const wrapper = createMenu();
    const outside = document.createElement('button');
    document.body.append(outside);
    try {
      const trigger = wrapper.get('button');
      await trigger.trigger('keydown', { key: 'ArrowDown' });
      await menuPanel().get('input[value="newest"]').trigger('keydown', { key: 'Escape' });
      expect(document.querySelector('[role="dialog"]')).toBeNull();
      expect(document.activeElement).toBe(trigger.element);
      await trigger.trigger('click');
      outside.focus();
      await wrapper.vm.$nextTick();
      expect(trigger.attributes('aria-expanded')).toBe('false');
      await trigger.trigger('click');
      outside.dispatchEvent(new Event('pointerdown', { bubbles: true }));
      await wrapper.vm.$nextTick();
      expect(trigger.attributes('aria-expanded')).toBe('false');
    } finally {
      wrapper.unmount();
    }
  });

  it('fits near viewport edges, allows panel scrolling, and closes when its anchor scrolls', async () => {
    const wrapper = createMenu();
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (
      this: HTMLElement
    ) {
      return this.tagName === 'BUTTON'
        ? ({
            left: 10,
            right: 40,
            top: window.innerHeight - 40,
            bottom: window.innerHeight - 10,
          } as DOMRect)
        : ({ height: 260 } as DOMRect);
    });
    vi.spyOn(HTMLElement.prototype, 'offsetWidth', 'get').mockReturnValue(288);
    try {
      await wrapper.get('button').trigger('click');
      const panel = menuPanel();
      expect(panel.element.style.left).toBe('8px');
      expect(panel.element.style.top).toBe(`${window.innerHeight - 40 - 6 - 260}px`);
      await panel.trigger('scroll');
      expect(wrapper.get('button').attributes('aria-expanded')).toBe('true');
      document.dispatchEvent(new Event('scroll'));
      await wrapper.vm.$nextTick();
      expect(wrapper.get('button').attributes('aria-expanded')).toBe('false');
    } finally {
      wrapper.unmount();
    }
  });
});
