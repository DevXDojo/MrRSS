import { watch } from 'vue';

interface Options {
  articleId: () => number;
  enabled: () => boolean;
  previous: () => void;
  next: () => void;
}

// Only the reader surface participates. Leave controls, selection and nested
// horizontal scrollers (including code blocks/tables) to their own interactions.
function canSwipe(event: Event): boolean {
  if (window.getSelection()?.toString()) return false;
  let element = event.target instanceof Element ? event.target : null;
  if (!element || element.closest('a,button,input,textarea,select,[contenteditable]:not([contenteditable="false"]),audio,video,iframe,[role="slider"],pre,table')) return false;
  while (element && element !== event.currentTarget) {
    if (element.scrollWidth > element.clientWidth + 1 && /auto|scroll/.test(getComputedStyle(element).overflowX)) return false;
    element = element.parentElement;
  }
  return true;
}

export function useArticleSwipe(options: Options) {
  let start: { x: number; y: number; id: number; time: number } | null = null;
  let wheelTime = -Infinity;
  let wheelDistance = 0;
  let wheelHandled = false;

  function cancel() { start = null; }
  watch(options.articleId, cancel, { flush: 'sync' });

  function touchstart(event: TouchEvent) {
    cancel();
    if (!options.enabled() || event.touches.length !== 1 || !canSwipe(event)) return;
    const touch = event.touches[0];
    start = { x: touch.clientX, y: touch.clientY, id: touch.identifier, time: event.timeStamp };
  }

  function touchmove(event: TouchEvent) {
    if (!start) return;
    if (event.touches.length !== 1 || !canSwipe(event)) return cancel();
    const touch = event.touches[0];
    const dx = Math.abs(touch.clientX - start.x);
    const dy = Math.abs(touch.clientY - start.y);
    if (touch.identifier !== start.id || (dy > 12 && dy >= dx)) return cancel();
    if (dx > 12 && dx > dy * 2 && event.cancelable) event.preventDefault();
  }

  function touchend(event: TouchEvent) {
    const origin = start;
    cancel();
    if (!origin || !options.enabled() || event.touches.length || !canSwipe(event)) return;
    const touch = Array.from(event.changedTouches).find((item) => item.identifier === origin.id);
    if (!touch || event.timeStamp - origin.time > 1000) return;
    const dx = touch.clientX - origin.x;
    const dy = touch.clientY - origin.y;
    if (Math.abs(dx) < 70 || Math.abs(dx) <= Math.abs(dy) * 2) return;
    if (event.cancelable) event.preventDefault();
    if (dx < 0) options.next();
    else options.previous();
  }

  function wheel(event: WheelEvent) {
    if (!options.enabled() || event.ctrlKey || event.shiftKey || !canSwipe(event)) return;
    if (Math.abs(event.deltaX) <= Math.abs(event.deltaY) * 2) return;
    if (event.timeStamp - wheelTime > 250) {
      wheelDistance = 0;
      wheelHandled = false;
    }
    wheelTime = event.timeStamp;
    if (event.cancelable) event.preventDefault();
    if (wheelHandled) return;
    // A trackpad emits many events, including inertia after navigation. Consume
    // the whole burst so one gesture can never skip several articles.
    const delta = event.deltaX * (event.deltaMode === 1 ? 16 : event.deltaMode === 2 ? 100 : 1);
    if (Math.sign(delta) !== Math.sign(wheelDistance)) wheelDistance = 0;
    wheelDistance += delta;
    if (Math.abs(wheelDistance) < 100) return;
    wheelHandled = true;
    if (wheelDistance > 0) options.next();
    else options.previous();
  }

  return { touchstart, touchmove, touchend, touchcancel: cancel, wheel };
}
