import { ref, onUnmounted, type Ref } from 'vue';
import type { Feed } from '@/types/models';

export interface DropPreview {
  targetFeedId: number | null;
  beforeTarget: boolean; // true = insert before target, false = insert after target
}

// Auto-scroll configuration
const SCROLL_THRESHOLD = 50; // Distance from edge to trigger scrolling (pixels)
const SCROLL_SPEED = 10; // Pixels per scroll step
const SCROLL_INTERVAL = 16; // milliseconds between scroll steps (~60fps)

export function useDragDrop() {
  const draggingFeedId: Ref<number | null> = ref(null);
  const dragOverCategory: Ref<string | null> = ref(null);
  const dropPreview: Ref<DropPreview> = ref({ targetFeedId: null, beforeTarget: true });

  // Auto-scroll state
  let scrollInterval: ReturnType<typeof setInterval> | null = null;
  let scrollableContainer: HTMLElement | null = null;
  let sidebarContainer: HTMLElement | null = null;

  // Store the last dragged feed ID for use in drop after dragend clears it
  let lastDraggedFeedId: number | null = null;

  function onDragStart(feedId: number, event: Event) {
    const dragEvent = event as DragEvent;
    draggingFeedId.value = feedId;
    lastDraggedFeedId = feedId; // Store for later use in drop
    if (dragEvent.dataTransfer) {
      dragEvent.dataTransfer.effectAllowed = 'move';
      dragEvent.dataTransfer.setData('text/plain', String(feedId));
    }
    // Add dragging class to the source element for visual feedback
    if (dragEvent.target instanceof HTMLElement) {
      const feedItem = dragEvent.target.closest('.feed-item');
      if (feedItem) {
        feedItem.classList.add('dragging');
      }
    }

    // Find the sidebar and scrollable containers
    sidebarContainer = document.querySelector('.sidebar');
    scrollableContainer = document.querySelector('.sidebar .flex-1.overflow-y-auto');
    if (!sidebarContainer) {
      console.warn('[onDragStart] Could not find sidebar container');
    }
    if (!scrollableContainer) {
      console.warn('[onDragStart] Could not find scrollable container');
    }

    // Start monitoring mouse position for auto-scroll
    startAutoScrollMonitor();

    console.log('[onDragStart] Started dragging feed:', feedId);
  }

  function onDragEnd() {
    console.log('[onDragEnd] Ended dragging feed:', draggingFeedId.value);
    // Remove dragging class from all feed items
    document.querySelectorAll('.feed-item.dragging').forEach((el) => {
      el.classList.remove('dragging');
    });
    draggingFeedId.value = null;
    // Don't clear lastDraggedFeedId yet - onDrop may still need it
    // It will be cleared after a short delay

    dragOverCategory.value = null;
    dropPreview.value = { targetFeedId: null, beforeTarget: true };

    // Stop auto-scroll
    stopAutoScrollMonitor();
    scrollableContainer = null;
    sidebarContainer = null;

    // Clear lastDraggedFeedId after a short delay to allow onDrop to complete
    setTimeout(() => {
      lastDraggedFeedId = null;
    }, 100);
  }

  function onDragOver(category: string, targetFeedId: number | null, event: Event) {
    if (!event || !(event instanceof DragEvent)) {
      console.log('[onDragOver] Invalid event:', event);
      return;
    }

    event.preventDefault();

    // Check if we're in the middle of a drag operation (dragend may have cleared draggingFeedId)
    const currentDraggingId = draggingFeedId.value || lastDraggedFeedId;
    if (!currentDraggingId) {
      console.log('[onDragOver] No dragging feed');
      return;
    }

    // Always update the dragOverCategory when over a valid category
    dragOverCategory.value = category;

    // Don't allow dropping on itself
    if (targetFeedId === currentDraggingId) {
      dropPreview.value = { targetFeedId: null, beforeTarget: true };
      console.log('[onDragOver] Dropping on itself, clearing preview');
      return;
    }

    // Calculate drop position based on mouse Y position relative to element
    let beforeTarget: boolean;
    if (targetFeedId !== null && event.target instanceof HTMLElement) {
      // Use target instead of currentTarget to get the actual element being hovered
      const target = event.target;
      // Get the feed-item element (might need to traverse up)
      const feedItem = target.closest('.feed-item');
      if (feedItem) {
        const rect = feedItem.getBoundingClientRect();
        const relativeY = event.clientY - rect.top;
        const threshold = rect.height / 2;
        beforeTarget = relativeY < threshold;

        // Debounce: only update if target or position changed significantly
        const newPreview = { targetFeedId, beforeTarget };
        if (
          dropPreview.value.targetFeedId !== newPreview.targetFeedId ||
          dropPreview.value.beforeTarget !== newPreview.beforeTarget
        ) {
          dropPreview.value = newPreview;
        }

        console.log(
          '[onDragOver] category:',
          category,
          'targetFeedId:',
          targetFeedId,
          'relativeY:',
          relativeY.toFixed(1),
          'threshold:',
          threshold.toFixed(1),
          'beforeTarget:',
          beforeTarget
        );
      } else {
        console.log('[onDragOver] Could not find .feed-item element');
      }
    } else {
      console.log('[onDragOver] No specific target, dropping at end. targetFeedId:', targetFeedId);
      // Only update if different
      if (dropPreview.value.targetFeedId !== null) {
        dropPreview.value = { targetFeedId: null, beforeTarget: true };
      }
    }

    console.log('[onDragOver] Updated dropPreview:', dropPreview.value);
  }

  function onDragLeave(category: string, event: Event) {
    if (!event || !(event instanceof DragEvent)) {
      return;
    }

    // Only clear the preview if we're actually leaving the category container
    // Check if the relatedTarget (where we're going) is outside the category
    const target = event.target as HTMLElement;
    const relatedTarget = event.relatedTarget as HTMLElement;

    // If moving to a child element, don't clear the preview
    if (relatedTarget && target.contains(relatedTarget)) {
      return;
    }

    // If moving from one category to another, the new category will handle it
    // Don't clear dragOverCategory here - let handleGlobalDragOver handle it
    // Only clear the drop preview
    dropPreview.value = { targetFeedId: null, beforeTarget: true };
    console.log('[onDragLeave] Cleared preview for category:', category);
  }

  async function onDrop(
    currentCategory: string,
    feeds: Feed[]
  ): Promise<{ success: boolean; error?: string }> {
    // Use lastDraggedFeedId as fallback since dragend may have cleared draggingFeedId
    const feedId = draggingFeedId.value || lastDraggedFeedId;

    if (!feedId) {
      return { success: false, error: 'No feed being dragged' };
    }
    // Use dragOverCategory if set, otherwise fall back to currentCategory
    // This handles cases where the drop happens but dragOverCategory was cleared
    let targetCategory = dragOverCategory.value || currentCategory;

    // Convert 'uncategorized' to empty string for the API
    if (targetCategory === 'uncategorized') {
      targetCategory = '';
    }

    const { targetFeedId, beforeTarget } = dropPreview.value;

    console.log('[onDrop] Starting drop operation:', {
      feedId,
      currentCategory,
      targetCategory,
      targetFeedId,
      beforeTarget,
      feedsCount: feeds.length,
    });

    // Sort feeds by position to get correct order
    const sortedFeeds = [...feeds].sort((a, b) => (a.position || 0) - (b.position || 0));

    // Find the dragging feed's current index (0-based)
    const draggingIndex = sortedFeeds.findIndex((f) => f.id === feedId);

    // Calculate the visual target index (0-based)
    // This is where the feed should appear in the sorted list
    let targetIndex: number;

    if (targetFeedId !== null) {
      const targetIdx = sortedFeeds.findIndex((f) => f.id === targetFeedId);
      if (targetIdx !== -1) {
        if (beforeTarget) {
          // Insert before the target feed
          targetIndex = targetIdx;
        } else {
          // Insert after the target feed
          targetIndex = targetIdx + 1;
        }
      } else {
        // Target feed not found, append to end
        targetIndex = sortedFeeds.length;
      }
    } else {
      // No specific target, append to end
      targetIndex = sortedFeeds.length;
    }

    // Calculate the final position index considering the dragging feed will be removed
    let newPosition = targetIndex;
    if (draggingIndex !== -1 && targetIndex > draggingIndex) {
      // Moving forward: after removing the dragging feed, indices shift down by 1
      newPosition = targetIndex - 1;
    }

    console.log('[onDrop] Calculated position:', {
      feedId,
      targetCategory,
      newPosition,
      draggingIndex,
      targetIndex,
      targetFeedId,
      beforeTarget,
      feedsInCategory: sortedFeeds.length,
    });

    try {
      const response = await fetch('/api/feeds/reorder', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          feed_id: feedId,
          category: targetCategory,
          position: newPosition,
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to reorder feed');
      }

      console.log('[onDrop] Successfully reordered feed');
      return { success: true };
    } catch (error) {
      console.error('[onDrop] Error:', error);
      return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
    }
  }

  // Auto-scroll functions
  function startAutoScrollMonitor() {
    if (scrollInterval) {
      clearInterval(scrollInterval);
    }

    // Add mousemove listener to track mouse position globally
    document.addEventListener('mousemove', trackMousePosition);

    // Add global dragover listener to detect when mouse leaves sidebar
    document.addEventListener('dragover', handleGlobalDragOver);

    // Start the scroll interval
    scrollInterval = setInterval(() => {
      if (!scrollableContainer) {
        return;
      }

      // Check if we're in the middle of a drag operation (dragend may have cleared draggingFeedId)
      const currentDraggingId = draggingFeedId.value || lastDraggedFeedId;
      if (!currentDraggingId) {
        return;
      }

      performAutoScroll();
    }, SCROLL_INTERVAL);
  }

  function stopAutoScrollMonitor() {
    if (scrollInterval) {
      clearInterval(scrollInterval);
      scrollInterval = null;
    }
    document.removeEventListener('mousemove', trackMousePosition);
    document.removeEventListener('dragover', handleGlobalDragOver);
  }

  // Handle global dragover to detect when mouse leaves the sidebar
  function handleGlobalDragOver(e: Event) {
    if (!e || !(e instanceof DragEvent) || !sidebarContainer) {
      return;
    }

    // Check if we're in the middle of a drag operation (dragend may have cleared draggingFeedId)
    const currentDraggingId = draggingFeedId.value || lastDraggedFeedId;
    if (!currentDraggingId) {
      return;
    }

    const dragEvent = e as DragEvent;
    const rect = sidebarContainer.getBoundingClientRect();

    // Check if mouse is outside the sidebar bounds (including scroll zone)
    // Expand the bounds by SCROLL_THRESHOLD to allow scrolling near edges
    // Use a more generous threshold to avoid clearing state during drop
    const DROP_THRESHOLD = SCROLL_THRESHOLD * 2; // More generous threshold
    const isOutside =
      dragEvent.clientX < rect.left - DROP_THRESHOLD ||
      dragEvent.clientX > rect.right + DROP_THRESHOLD ||
      dragEvent.clientY < rect.top - DROP_THRESHOLD ||
      dragEvent.clientY > rect.bottom + DROP_THRESHOLD;

    if (isOutside) {
      // Clear the drag-over state and drop preview when outside sidebar
      dragOverCategory.value = null;
      dropPreview.value = { targetFeedId: null, beforeTarget: true };
      console.log('[handleGlobalDragOver] Cleared state - mouse is outside sidebar');
    }
  }

  // Store latest mouse position
  let mouseX = 0;
  let mouseY = 0;

  function trackMousePosition(e: MouseEvent) {
    mouseX = e.clientX;
    mouseY = e.clientY;
  }

  function performAutoScroll() {
    if (!scrollableContainer) {
      return;
    }

    const rect = scrollableContainer.getBoundingClientRect();

    // Calculate distance from edges (can be negative when outside)
    const distanceFromTop = mouseY - rect.top;
    const distanceFromBottom = rect.bottom - mouseY;
    const distanceFromLeft = mouseX - rect.left;
    const distanceFromRight = rect.right - mouseX;

    // Check if mouse is horizontally within the container bounds
    const isWithinHorizontalBounds = distanceFromLeft >= 0 && distanceFromRight >= 0;

    // Only scroll if mouse is horizontally aligned with the container
    if (!isWithinHorizontalBounds) {
      return;
    }

    // Check if mouse is near the top edge (inside or above)
    // Allow scrolling even when mouse is slightly above (up to SCROLL_THRESHOLD)
    if (distanceFromTop > -SCROLL_THRESHOLD && distanceFromTop < SCROLL_THRESHOLD) {
      if (scrollableContainer.scrollTop > 0) {
        // Calculate scroll speed based on proximity to edge
        // When above the top edge (negative), scroll faster
        const speedMultiplier =
          distanceFromTop >= 0
            ? // Inside: closer = faster (1x to 2x)
              1 + (SCROLL_THRESHOLD - distanceFromTop) / SCROLL_THRESHOLD
            : // Above: farther = faster (1x to 2x)
              1 + (SCROLL_THRESHOLD + distanceFromTop) / SCROLL_THRESHOLD;
        scrollableContainer.scrollTop -= SCROLL_SPEED * speedMultiplier;
      }
    }
    // Check if mouse is near the bottom edge (inside or below)
    // Allow scrolling even when mouse is slightly below (up to SCROLL_THRESHOLD)
    else if (distanceFromBottom > -SCROLL_THRESHOLD && distanceFromBottom < SCROLL_THRESHOLD) {
      const maxScroll = scrollableContainer.scrollHeight - scrollableContainer.clientHeight;
      if (scrollableContainer.scrollTop < maxScroll) {
        // Calculate scroll speed based on proximity to edge
        // When below the bottom edge (negative), scroll faster
        const speedMultiplier =
          distanceFromBottom >= 0
            ? // Inside: closer = faster (1x to 2x)
              1 + (SCROLL_THRESHOLD - distanceFromBottom) / SCROLL_THRESHOLD
            : // Below: farther = faster (1x to 2x)
              1 + (SCROLL_THRESHOLD + distanceFromBottom) / SCROLL_THRESHOLD;
        scrollableContainer.scrollTop += SCROLL_SPEED * speedMultiplier;
      }
    }
  }

  // Clean up on component unmount
  onUnmounted(() => {
    stopAutoScrollMonitor();
  });

  return {
    draggingFeedId,
    dragOverCategory,
    dropPreview,
    onDragStart,
    onDragEnd,
    onDragOver,
    onDragLeave,
    onDrop,
  };
}
