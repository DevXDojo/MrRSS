import { ref, onUnmounted } from 'vue';
import type { Feed } from '@/types/models';

export interface DropPreview {
  targetFeedId: number | null;
  beforeTarget: boolean;
}

export function useDragDrop() {
  const draggingFeedId = ref<number | null>(null);
  const dragOverCategory = ref<string | null>(null);
  const dropPreview = ref<DropPreview>({ targetFeedId: null, beforeTarget: true });
  let scrollTimer: ReturnType<typeof setInterval> | null = null;
  let scrollContainer: HTMLElement | null = null;
  let pointerX = 0;
  let pointerY = 0;

  function resetPreview() {
    dragOverCategory.value = null;
    dropPreview.value = { targetFeedId: null, beforeTarget: true };
  }

  function trackDrag(event: DragEvent) {
    pointerX = event.clientX;
    pointerY = event.clientY;
    if (event.target instanceof Element && !event.target.closest('.categories-list')) resetPreview();
  }

  function onDragStart(feedId: number, event: Event) {
    onDragEnd();
    const drag = event as DragEvent;
    draggingFeedId.value = feedId;
    drag.dataTransfer?.setData('text/plain', String(feedId));
    if (drag.dataTransfer) drag.dataTransfer.effectAllowed = 'move';
    const source = event.target instanceof Element ? event.target.closest('.feed-item') : null;
    source?.classList.add('dragging');
    scrollContainer = source?.closest<HTMLElement>('.categories-list') || null;
    pointerX = drag.clientX;
    pointerY = drag.clientY;
    document.addEventListener('dragover', trackDrag);
    scrollTimer = setInterval(() => {
      if (!scrollContainer) return;
      const rect = scrollContainer.getBoundingClientRect();
      if (pointerX < rect.left || pointerX > rect.right) return;
      if (pointerY < rect.top + 50) scrollContainer.scrollTop -= 10;
      else if (pointerY > rect.bottom - 50) scrollContainer.scrollTop += 10;
    }, 16);
  }

  function onDragEnd() {
    document.querySelectorAll('.feed-item.dragging').forEach((el) => el.classList.remove('dragging'));
    draggingFeedId.value = null;
    resetPreview();
    if (scrollTimer) clearInterval(scrollTimer);
    scrollTimer = null;
    scrollContainer = null;
    document.removeEventListener('dragover', trackDrag);
  }

  function onDragOver(category: string, targetFeedId: number | null, event: Event) {
    if (draggingFeedId.value === null) return;
    const drag = event as DragEvent;
    drag.preventDefault();
    drag.stopPropagation();
    if (drag.dataTransfer) drag.dataTransfer.dropEffect = 'move';
    pointerX = drag.clientX;
    pointerY = drag.clientY;
    dragOverCategory.value = category;
    if (targetFeedId === draggingFeedId.value) {
      dropPreview.value = { targetFeedId, beforeTarget: true };
      return;
    }
    // Resolve from the stable row, even when the event originated on an SVG or in a gap.
    const target = targetFeedId === null ? null : scrollContainer?.querySelector<HTMLElement>(`[data-feed-id="${targetFeedId}"]`);
    let beforeTarget = true;
    if (target) {
      const rect = target.getBoundingClientRect();
      const middle = rect.top + rect.height / 2;
      // Keep the current side in a small dead zone to prevent jitter at the midpoint.
      beforeTarget = dropPreview.value.targetFeedId === targetFeedId && Math.abs(drag.clientY - middle) < 3
        ? dropPreview.value.beforeTarget : drag.clientY < middle;
    }
    if (dropPreview.value.targetFeedId !== targetFeedId || dropPreview.value.beforeTarget !== beforeTarget) {
      dropPreview.value = { targetFeedId, beforeTarget };
    }
  }

  function onDragLeave(_category: string, event: Event) {
    const drag = event as DragEvent;
    const container = drag.currentTarget;
    if (container instanceof Element && drag.relatedTarget instanceof Node && container.contains(drag.relatedTarget)) return;
    // A following dragover chooses the new row. Clear only when leaving the list entirely.
    if (drag.relatedTarget instanceof Element && drag.relatedTarget.closest('.categories-list')) return;
    resetPreview();
  }

  async function onDrop(category: string, feeds: Feed[]): Promise<{ success: boolean; error?: string }> {
    const feedId = draggingFeedId.value;
    if (feedId === null) return { success: false, error: 'No feed being dragged' };
    const targetCategory = (dragOverCategory.value ?? category) === 'uncategorized' ? '' : (dragOverCategory.value ?? category);
    const { targetFeedId, beforeTarget } = dropPreview.value;
    const sorted = [...feeds].sort((a, b) => (a.position || 0) - (b.position || 0) || a.id - b.id);
    // Dropping on the source row is a no-op, not a move to the end.
    if (targetFeedId === feedId) return { success: true };
    const others = sorted.filter((feed) => feed.id !== feedId);
    const targetIndex = others.findIndex((feed) => feed.id === targetFeedId);
    const position = targetIndex < 0 ? others.length : targetIndex + (beforeTarget ? 0 : 1);
    try {
      const response = await fetch('/api/feeds/reorder', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ feed_id: feedId, category: targetCategory, position }),
      });
      if (!response.ok) throw new Error(await response.text());
      return { success: true };
    } catch (error) {
      return { success: false, error: error instanceof Error ? error.message : 'Reorder failed' };
    } finally {
      onDragEnd();
    }
  }

  onUnmounted(onDragEnd);
  return { draggingFeedId, dragOverCategory, dropPreview, onDragStart, onDragEnd, onDragOver, onDragLeave, onDrop };
}
