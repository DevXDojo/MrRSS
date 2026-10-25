import { onMounted, onBeforeUnmount, ref, type Ref } from 'vue';
import { Events, System, Window } from '@wailsio/runtime';

/** Own only fullscreen entered by this viewer; restore it when the viewer closes. */
export function useViewerFullscreen(element: Ref<HTMLElement | null>, onError: () => void) {
  const isFullscreen = ref(false);
  const supported = ref(false);
  const pending = ref(false);
  let desktop = false;
  let ownsNativeFullscreen = false;
  let disposed = false;
  let removeNativeListener: (() => void) | undefined;

  function syncBrowserState() {
    isFullscreen.value = !!element.value && document.fullscreenElement === element.value;
  }

  async function leave() {
    if (desktop) {
      if (ownsNativeFullscreen) {
        await Window.UnFullscreen();
        ownsNativeFullscreen = false;
      }
    } else if (element.value && document.fullscreenElement === element.value) {
      await document.exitFullscreen();
    }
    isFullscreen.value = false;
  }

  async function toggle() {
    if (pending.value || disposed || !element.value) return;
    pending.value = true;
    const target = element.value;
    try {
      if (isFullscreen.value) {
        await leave();
      } else if (desktop) {
        const alreadyFullscreen = await Window.IsFullscreen();
        if (disposed) return;
        if (!alreadyFullscreen) {
          ownsNativeFullscreen = true;
          await Window.Fullscreen();
        }
        if (disposed) await leave();
        else isFullscreen.value = true;
      } else {
        await target.requestFullscreen();
        if (disposed && document.fullscreenElement === target) await document.exitFullscreen();
        else syncBrowserState();
      }
    } catch {
      if (!disposed) onError();
    } finally {
      pending.value = false;
    }
  }

  // The gallery's capture-phase Escape handler calls this before closing.
  function exitFullscreen(): boolean {
    if (!isFullscreen.value) return false;
    if (!pending.value) void toggle();
    return true;
  }

  onMounted(() => {
    desktop = System.IsDesktop();
    supported.value = desktop || typeof element.value?.requestFullscreen === 'function';
    document.addEventListener('fullscreenchange', syncBrowserState);
    if (desktop) {
      removeNativeListener = Events.On('common:WindowUnFullscreen', () => {
        isFullscreen.value = false;
        ownsNativeFullscreen = false;
      });
    }
  });

  onBeforeUnmount(() => {
    disposed = true;
    document.removeEventListener('fullscreenchange', syncBrowserState);
    removeNativeListener?.();
    void leave().catch(onError);
  });

  return { isFullscreen, supported, pending, toggle, exitFullscreen };
}
