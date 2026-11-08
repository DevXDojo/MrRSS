import { nextTick, onMounted, onUnmounted, ref, watch, type Ref } from 'vue';
import { tickerNumber } from './tickerPreferences';

interface Options {
  speed: () => number;
  paused: () => boolean;
  content: () => unknown;
}

export function useTickerMarquee(
  viewport: Ref<HTMLElement | null>,
  group: Ref<HTMLElement | null>,
  track: Ref<HTMLElement | null>,
  options: Options
) {
  const copies = ref(2);
  let distance = 0;
  let width = 0;
  let lastFrame: number | null = null;
  let frame: number | undefined;
  let observer: ResizeObserver | undefined;
  let disposed = false;
  let updating = false;

  function paint() {
    if (width > 0) distance = ((distance % width) + width) % width;
    if (track.value) track.value.style.transform = `translate3d(${-distance}px,0,0)`;
  }
  function tick(time: number) {
    frame = undefined;
    if (disposed || options.paused() || updating || !width) {
      lastFrame = null;
      return;
    }
    if (lastFrame !== null)
      distance +=
        (Math.min(64, time - lastFrame) * tickerNumber(options.speed(), 60, 10, 200)) / 1000;
    lastFrame = time;
    paint();
    frame = requestAnimationFrame(tick);
  }
  function synchronize() {
    if (frame !== undefined) cancelAnimationFrame(frame);
    frame = undefined;
    lastFrame = null;
    if (!disposed && !options.paused() && !updating && width) frame = requestAnimationFrame(tick);
  }
  function measure() {
    width = group.value?.scrollWidth ?? 0;
    copies.value = width
      ? Math.min(50, Math.max(2, Math.ceil((viewport.value?.clientWidth ?? 0) / width) + 1))
      : 2;
    paint();
    synchronize();
  }
  const elements = () =>
    Array.from(group.value?.querySelectorAll<HTMLElement>('[data-ticker-article]') ?? []);
  function step(direction: number) {
    const starts = elements().map((element) => element.offsetLeft);
    if (!starts.length) return;
    distance =
      direction > 0
        ? (starts.find((start) => start > distance + 2) ?? starts[0])
        : (starts.filter((start) => start < distance - 2).at(-1) ?? starts.at(-1)!);
    paint();
  }
  // Preserve the visible title and its progress when new headlines arrive or
  // the typography changes, rather than restarting the entire strip.
  watch(options.content, async () => {
    const anchor = elements()
      .filter((element) => element.offsetLeft <= distance)
      .at(-1);
    const id = anchor?.dataset.tickerArticle;
    const progress = anchor ? (distance - anchor.offsetLeft) / Math.max(1, anchor.offsetWidth) : 0;
    updating = true;
    synchronize();
    await nextTick();
    if (disposed) return;
    measure();
    const updated = elements().find((element) => element.dataset.tickerArticle === id);
    distance = updated ? updated.offsetLeft + progress * updated.offsetWidth : 0;
    updating = false;
    paint();
    synchronize();
  });
  watch(options.paused, synchronize);
  watch(
    [group, viewport],
    () => {
      observer?.disconnect();
      if (group.value) observer?.observe(group.value);
      if (viewport.value) observer?.observe(viewport.value);
      measure();
    },
    { flush: 'post' }
  );
  onMounted(() => {
    observer = new ResizeObserver(measure);
    if (group.value) observer.observe(group.value);
    if (viewport.value) observer.observe(viewport.value);
    measure();
  });
  onUnmounted(() => {
    disposed = true;
    observer?.disconnect();
    if (frame !== undefined) cancelAnimationFrame(frame);
  });
  return { copies, step };
}
