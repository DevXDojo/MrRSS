import { onBeforeUnmount, ref, type Ref } from 'vue';

export function clampArticleTableSplit(value: number): number {
  return Number.isFinite(value) ? Math.min(75, Math.max(25, value)) : 40;
}

export function useArticleTableSplit(container: Ref<HTMLElement | null>) {
  const saved = localStorage.getItem('articleTableSplit');
  const split = ref(clampArticleTableSplit(saved === null ? 40 : Number(saved)));
  const resizing = ref(false);
  let previousCursor = '';
  let previousSelect = '';

  function setSplit(value: number): void {
    split.value = clampArticleTableSplit(value);
  }

  function move(event: PointerEvent): void {
    const bounds = container.value?.getBoundingClientRect();
    if (bounds && bounds.height > 0) setSplit((100 * (event.clientY - bounds.top)) / bounds.height);
  }

  function stop(): void {
    if (!resizing.value) return;
    resizing.value = false;
    localStorage.setItem('articleTableSplit', String(split.value));
    document.body.style.cursor = previousCursor;
    document.body.style.userSelect = previousSelect;
    window.removeEventListener('pointermove', move);
    window.removeEventListener('pointerup', stop);
    window.removeEventListener('pointercancel', stop);
    window.removeEventListener('blur', stop);
  }

  function start(event: PointerEvent): void {
    if (event.button !== 0 || resizing.value) return;
    event.preventDefault();
    resizing.value = true;
    previousCursor = document.body.style.cursor;
    previousSelect = document.body.style.userSelect;
    document.body.style.cursor = 'row-resize';
    document.body.style.userSelect = 'none';
    window.addEventListener('pointermove', move);
    window.addEventListener('pointerup', stop);
    window.addEventListener('pointercancel', stop);
    window.addEventListener('blur', stop);
  }

  function keydown(event: KeyboardEvent): void {
    const values: Record<string, number> = {
      ArrowUp: split.value - 5,
      ArrowDown: split.value + 5,
      Home: 25,
      End: 75,
    };
    if (!(event.key in values)) return;
    event.preventDefault();
    event.stopPropagation();
    setSplit(values[event.key]);
    localStorage.setItem('articleTableSplit', String(split.value));
  }

  onBeforeUnmount(stop);
  return { split, resizing, start, stop, keydown };
}
