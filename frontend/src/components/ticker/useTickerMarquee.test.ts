import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, nextTick, onUpdated, ref } from 'vue';
import { mount, type VueWrapper } from '@vue/test-utils';
import { useTickerMarquee } from './useTickerMarquee';

describe('continuous ticker motion', () => {
  let wrapper: VueWrapper;
  let track: HTMLElement;
  let group: HTMLElement;
  const paused = ref(false),
    speed = ref(100),
    content = ref(0);
  let frameId = 0;
  const frames = new Map<number, FrameRequestCallback>();
  function advance(time: number) {
    const callbacks = [...frames.values()];
    frames.clear();
    callbacks.forEach((callback) => callback(time));
  }
  function expectOffset(x: number) {
    expect(track.style.transform).toMatch(
      new RegExp(`^translate3d\\(${x}px,\\s*0(?:px)?,\\s*0(?:px)?\\)$`)
    );
  }
  beforeEach(() => {
    paused.value = false;
    speed.value = 100;
    content.value = 0;
    frames.clear();
    vi.stubGlobal('requestAnimationFrame', (callback: FrameRequestCallback) => {
      frames.set(++frameId, callback);
      return frameId;
    });
    vi.stubGlobal('cancelAnimationFrame', (id: number) => frames.delete(id));
    vi.stubGlobal(
      'ResizeObserver',
      class {
        observe() {}
        disconnect() {}
      }
    );
    track = document.createElement('div');
    group = document.createElement('div');
    const viewport = document.createElement('div');
    Object.defineProperty(viewport, 'clientWidth', { value: 400 });
    Object.defineProperty(group, 'scrollWidth', { value: 800, configurable: true });
    for (let index = 0; index < 4; index++) {
      const item = document.createElement('button');
      item.dataset.tickerArticle = String(index + 1);
      Object.defineProperty(item, 'offsetLeft', { value: index * 200, configurable: true });
      Object.defineProperty(item, 'offsetWidth', { value: 200, configurable: true });
      group.appendChild(item);
    }
    wrapper = mount(
      defineComponent({
        setup() {
          useTickerMarquee(ref(viewport), ref(group), ref(track), {
            speed: () => speed.value,
            paused: () => paused.value,
            content: () => content.value,
          });
          onUpdated(() => {
            if (content.value) {
              Object.defineProperty(group.children[0], 'offsetLeft', {
                value: 100,
                configurable: true,
              });
              Object.defineProperty(group, 'scrollWidth', { value: 900, configurable: true });
            }
          });
          return { content };
        },
        template: '<div>{{ content }}</div>',
      })
    );
  });
  afterEach(() => {
    wrapper.unmount();
    vi.unstubAllGlobals();
  });
  it('resumes without a time jump and changes speed without restarting', async () => {
    advance(0);
    advance(50);
    expectOffset(-5);
    paused.value = true;
    await nextTick();
    expect(frames.size).toBe(0);
    paused.value = false;
    await nextTick();
    advance(10000);
    advance(10050);
    expectOffset(-10);
    speed.value = 200;
    advance(10075);
    expectOffset(-15);
  });
  it('keeps the visible title in place when new titles are inserted before it', async () => {
    advance(0);
    advance(50);
    paused.value = true;
    await nextTick();
    content.value++;
    await nextTick();
    await nextTick();
    expectOffset(-105);
    expect(frames.size).toBe(0);
  });
});
