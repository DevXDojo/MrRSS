import { afterEach, describe, expect, it } from 'vitest';
import { mount } from '@vue/test-utils';
import Sidebar from './Sidebar.vue';

afterEach(() => localStorage.removeItem('FeedListPinned'));

describe('floating feed list', () => {
  it.each(['pointerdown', 'focusin', 'blur'])(
    'closes on outside %s while preserving pinned lists',
    async (event) => {
      for (const pinned of [true, false]) {
        localStorage.setItem('FeedListPinned', String(pinned));
        const wrapper = mount(Sidebar, {
          props: { isOpen: true },
          attachTo: document.body,
          global: { stubs: { ActivityBar: true, FeedList: true } },
        });
        wrapper.element.dispatchEvent(new Event('pointerdown', { bubbles: true }));
        expect(wrapper.emitted('toggle')).toBeUndefined();
        (event === 'blur' ? window : document.body).dispatchEvent(
          new Event(event, { bubbles: true })
        );
        expect(wrapper.emitted('toggle')?.length ?? 0).toBe(pinned ? 0 : 1);
        wrapper.unmount();
      }
    }
  );
  it('removes outside listeners when unmounted', () => {
    localStorage.setItem('FeedListPinned', 'false');
    const wrapper = mount(Sidebar, {
      props: { isOpen: true },
      global: { stubs: { ActivityBar: true, FeedList: true } },
    });
    wrapper.unmount();
    document.body.dispatchEvent(new Event('pointerdown', { bubbles: true }));
    expect(wrapper.emitted('toggle')).toBeUndefined();
  });
});
