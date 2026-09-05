import { onMounted, onUnmounted, watch } from 'vue';

// One application-level style survives article navigation and empty reader views.
export function useCustomCSS(getFile: () => string) {
  let style: HTMLStyleElement | null = null;
  let controller: AbortController | null = null;
  let generation = 0;

  function removeStyle() {
    style?.remove();
    style = null;
  }

  async function reload() {
    const current = ++generation;
    controller?.abort();
    controller = null;
    if (!getFile()) {
      removeStyle();
      return;
    }
    const request = new AbortController();
    controller = request;
    try {
      const response = await fetch('/api/custom-css', {
        signal: request.signal,
        cache: 'no-store',
      });
      if (!response.ok) throw new Error(`Custom CSS: HTTP ${response.status}`);
      const css = await response.text();
      if (current !== generation || request.signal.aborted) return;
      if (!style) {
        style = document.createElement('style');
        style.dataset.customCss = 'application';
        document.head.appendChild(style);
      }
      style.textContent = css;
    } catch (error) {
      if (current !== generation || request.signal.aborted) return;
      removeStyle();
      console.error('Failed to load custom CSS:', error);
    }
  }

  const stop = watch(getFile, reload, { immediate: true });
  onMounted(() => window.addEventListener('custom-css-changed', reload));
  onUnmounted(() => {
    ++generation;
    controller?.abort();
    stop();
    removeStyle();
    window.removeEventListener('custom-css-changed', reload);
  });
}
