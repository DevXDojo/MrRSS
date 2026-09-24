import { ref, onBeforeUnmount } from 'vue';

export function useResizablePanels() {
  const sidebarWidth = ref<number>(256);
  const articleListWidth = ref<number>(350);
  const isResizingSidebar = ref<boolean>(false);
  const isResizingArticleList = ref<boolean>(false);
  const compactMode = ref<boolean>(false);

  let previousCursor = '';
  let previousSelect = '';

  const storageKey = () => `articleListWidth:${compactMode.value ? 'compact' : 'normal'}`;

  function restoreArticleListWidth(): void {
    const fallback = compactMode.value ? 500 : 350;
    try {
      const saved = localStorage.getItem(storageKey());
      const width = saved === null ? fallback : Number(saved);
      articleListWidth.value = Number.isFinite(width) && width > 0
        ? Math.min(compactMode.value ? 800 : 600, Math.max(compactMode.value ? 300 : 280, width))
        : fallback;
    } catch {
      articleListWidth.value = fallback;
    }
  }

  // Track initial mouse position when starting resize
  const initialMouseX = ref<number>(0);
  const initialArticleListWidth = ref<number>(400);

  function setCompactMode(enabled: boolean): void {
    stopResizeArticleList();
    compactMode.value = enabled;
    restoreArticleListWidth();
  }

  restoreArticleListWidth();

  // Sidebar resize handlers
  function startResizeSidebar(): void {
    isResizingSidebar.value = true;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
    window.addEventListener('mousemove', handleResizeSidebar);
    window.addEventListener('mouseup', stopResizeSidebar);
  }

  function handleResizeSidebar(event: MouseEvent): void {
    if (!isResizingSidebar.value) return;
    const newWidth = event.clientX;
    if (newWidth >= 180 && newWidth <= 450) {
      sidebarWidth.value = newWidth;
    }
  }

  function stopResizeSidebar(): void {
    isResizingSidebar.value = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    window.removeEventListener('mousemove', handleResizeSidebar);
    window.removeEventListener('mouseup', stopResizeSidebar);
  }

  // Article list resize handlers
  function startResizeArticleList(event: MouseEvent): void {
    if (event.button !== 0 || isResizingArticleList.value) return;
    event.preventDefault();
    isResizingArticleList.value = true;
    // Store initial mouse position and article list width
    initialMouseX.value = event.clientX;
    initialArticleListWidth.value = articleListWidth.value;
    previousCursor = document.body.style.cursor;
    previousSelect = document.body.style.userSelect;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
    window.addEventListener('mousemove', handleResizeArticleList);
    window.addEventListener('mouseup', stopResizeArticleList);
    window.addEventListener('blur', stopResizeArticleList);
  }

  function handleResizeArticleList(event: MouseEvent): void {
    if (!isResizingArticleList.value) return;
    const currentMouseX = event.clientX;
    // Calculate the delta from the initial position and apply to initial width
    const deltaX = currentMouseX - initialMouseX.value;
    const newWidth = initialArticleListWidth.value + deltaX;
    // In compact mode, allow wider range (300-800), in normal mode (250-600)
    const minWidth = compactMode.value ? 300 : 280;
    const maxWidth = compactMode.value ? 800 : 600;
    if (newWidth >= minWidth && newWidth <= maxWidth) {
      articleListWidth.value = newWidth;
    }
  }

  function stopResizeArticleList(): void {
    if (!isResizingArticleList.value) return;
    isResizingArticleList.value = false;
    document.body.style.cursor = previousCursor;
    document.body.style.userSelect = previousSelect;
    window.removeEventListener('mousemove', handleResizeArticleList);
    window.removeEventListener('mouseup', stopResizeArticleList);
    window.removeEventListener('blur', stopResizeArticleList);
    try {
      localStorage.setItem(storageKey(), String(articleListWidth.value));
    } catch {
      // Keep resizing usable when browser storage is unavailable.
    }
  }

  // Cleanup
  onBeforeUnmount(() => {
    stopResizeArticleList();
    window.removeEventListener('mousemove', handleResizeSidebar);
    window.removeEventListener('mouseup', stopResizeSidebar);
    window.removeEventListener('mousemove', handleResizeArticleList);
    window.removeEventListener('mouseup', stopResizeArticleList);
  });

  return {
    sidebarWidth,
    articleListWidth,
    startResizeSidebar,
    startResizeArticleList,
    setCompactMode,
  };
}
