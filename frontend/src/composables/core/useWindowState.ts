import { onMounted, onUnmounted } from 'vue';

interface WindowState {
  x: number;
  y: number;
  width: number;
  height: number;
  maximized: boolean;
}

export function useWindowState() {
  let saveTimeout: NodeJS.Timeout | null = null;
  let isRestoringState = false;

  /**
   * Load and restore window state from database
   *
   * NOTE: Actual window restoration happens in main.go during OnStartup.
   * This frontend function is kept for API compatibility and logging.
   * The window state is restored by the Go backend which has access to
   * Wails runtime context.
   */
  async function restoreWindowState() {
    try {
      isRestoringState = true;

      const response = await fetch('/api/window/state');
      if (!response.ok) {
        console.warn('Failed to load window state');
        return;
      }

      const data = await response.json();
    } catch (error) {
      // Log fetch errors during state restoration for debugging
      console.debug('Failed to restore window state:', error);
    } finally {
      // Wait a bit before allowing saves
      setTimeout(() => {
        isRestoringState = false;
      }, 1000);
    }
  }

  /**
   * Save current window state to database
   */
  async function saveWindowState() {
    // Don't save while we're restoring state
    if (isRestoringState) {
      return;
    }

    try {
      // Use browser window properties to get approximate state
      // Note: These values may not be 100% accurate due to browser limitations,
      // but they provide a reasonable approximation for window state persistence
      const currentStateResponse = await fetch('/api/window/state');
      const currentState = currentStateResponse.ok ? await currentStateResponse.json() : {};

      const state: WindowState = {
        x: window.screenX || 0,
        y: window.screenY || 0,
        width: window.innerWidth || 1024,
        height: window.innerHeight || 768,
        maximized: currentState.maximized === true || currentState.maximized === 'true',
      };

      // Only save if values are reasonable
      if (
        state.width >= 400 &&
        state.height >= 300 &&
        state.width <= 4000 &&
        state.height <= 3000
      ) {
        await fetch('/api/window/save', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(state),
        });
      }
    } catch (error) {
      // Log save errors for debugging (non-critical functionality)
      console.debug('Failed to save window state:', error);
    }
  }

  /**
   * Debounced save to avoid excessive writes
   */
  function debouncedSave() {
    if (saveTimeout) {
      clearTimeout(saveTimeout);
    }
    saveTimeout = setTimeout(saveWindowState, 2000); // Increased from 500ms to 2000ms
  }

  /**
   * Setup window event listeners
   */
  function setupListeners() {
    // Listen to window resize and move events
    // We use multiple approaches to catch window state changes:

    // 1. Browser resize event (fires when window size changes)
    const handleResize = () => {
      debouncedSave();
    };
    window.addEventListener('resize', handleResize);

    // 2. Visibility change (fires when window is minimized/maximized)
    const handleVisibilityChange = () => {
      if (!document.hidden) {
        debouncedSave();
      }
    };
    document.addEventListener('visibilitychange', handleVisibilityChange);

    // 3. Window blur event (fires when window loses focus)
    const handleBlur = () => {
      debouncedSave();
    };
    window.addEventListener('blur', handleBlur);

    return () => {
      window.removeEventListener('resize', handleResize);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      window.removeEventListener('blur', handleBlur);
      if (saveTimeout) {
        clearTimeout(saveTimeout);
      }
    };
  }

  /**
   * Initialize window state management
   */
  function init() {
    let cleanup: (() => void) | null = null;

    onMounted(async () => {
      // Load state for logging
      await restoreWindowState();

      // Setup event listeners to save window state changes
      cleanup = setupListeners();
    });

    onUnmounted(() => {
      // Cleanup event listeners
      if (cleanup) {
        cleanup();
      }
      if (saveTimeout) {
        clearTimeout(saveTimeout);
      }
    });
  }

  return {
    init,
    restoreWindowState,
    saveWindowState,
  };
}
