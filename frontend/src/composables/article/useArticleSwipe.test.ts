import { describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref } from 'vue';
import { useArticleSwipe } from './useArticleSwipe';

function fixture() {
  const id = ref(1);
  const previous = vi.fn();
  const next = vi.fn();
  const wrapper = mount({
    setup() {
      return {
        swipe: useArticleSwipe({ articleId: () => id.value, enabled: () => true, previous, next }),
      };
    },
    template:
      '<div @touchstart="swipe.touchstart" @touchmove="swipe.touchmove" @touchend="swipe.touchend" @touchcancel="swipe.touchcancel" @wheel="swipe.wheel"><p>Text</p><button>Control</button><pre>Code</pre></div>',
  });
  function touch(type: string, x: number, y = 0, target = 'p', count = 1) {
    const point = { identifier: 1, clientX: x, clientY: y };
    const event = new Event(type, { bubbles: true, cancelable: true });
    Object.defineProperties(event, {
      touches: { value: type === 'touchend' ? [] : Array(count).fill(point) },
      changedTouches: { value: [point] },
    });
    wrapper.get(target).element.dispatchEvent(event);
  }
  return { id, wrapper, previous, next, touch };
}

describe('article swipe', () => {
  it('navigates in both directions only after a deliberate horizontal swipe', () => {
    const f = fixture();
    f.touch('touchstart', 200);
    f.touch('touchend', 100);
    f.touch('touchstart', 100);
    f.touch('touchend', 200);
    f.touch('touchstart', 100);
    f.touch('touchend', 125);
    expect(f.next).toHaveBeenCalledTimes(1);
    expect(f.previous).toHaveBeenCalledTimes(1);
    f.wrapper.unmount();
  });

  it('ignores vertical scrolling, controls, code, multitouch and cancelled/stale gestures', () => {
    const f = fixture();
    for (const target of ['button', 'pre']) {
      f.touch('touchstart', 200, 0, target);
      f.touch('touchend', 0, 0, target);
    }
    f.touch('touchstart', 200);
    f.touch('touchmove', 195, 50);
    f.touch('touchend', 0);
    f.touch('touchstart', 200, 0, 'p', 2);
    f.touch('touchend', 0);
    f.touch('touchstart', 200);
    f.touch('touchcancel', 100);
    f.touch('touchend', 0);
    f.touch('touchstart', 200);
    f.id.value = 2;
    f.touch('touchend', 0);
    expect(f.next).not.toHaveBeenCalled();
    expect(f.previous).not.toHaveBeenCalled();
    f.wrapper.unmount();
  });

  it('ignores selection and horizontal scrollers, and consumes trackpad inertia', () => {
    const f = fixture();
    const selection = vi
      .spyOn(window, 'getSelection')
      .mockReturnValue({ toString: () => 'Selected' } as Selection);
    f.touch('touchstart', 200);
    f.touch('touchend', 0);
    expect(f.next).not.toHaveBeenCalled();
    selection.mockRestore();
    const paragraph = f.wrapper.get('p').element;
    Object.defineProperties(paragraph, {
      scrollWidth: { value: 800, configurable: true },
      clientWidth: { value: 200 },
    });
    paragraph.style.overflowX = 'auto';
    f.touch('touchstart', 200);
    f.touch('touchend', 0);
    expect(f.next).not.toHaveBeenCalled();
    Object.defineProperty(paragraph, 'scrollWidth', { value: 200 });
    for (let i = 0; i < 8; i++)
      paragraph.dispatchEvent(new WheelEvent('wheel', { bubbles: true, deltaX: 40 }));
    expect(f.next).toHaveBeenCalledTimes(1);
    f.wrapper.unmount();
  });
});
