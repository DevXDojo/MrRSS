import type { GalleryKeyboardReturn } from '../types';

/**
 * Composable for handling keyboard shortcuts in image gallery
 * @param handlers - Keyboard event handlers
 * @returns Keyboard control methods
 */
export function useGalleryKeyboard(handlers: {
  onClose: () => void;
  onPrevious: () => void;
  onNext: () => void;
  onZoomIn: () => void;
  onZoomOut: () => void;
  isImageViewerOpen: () => boolean;
}): GalleryKeyboardReturn {
  /**
   * Handle keyboard shortcuts
   */
  function handleKeyDown(e: KeyboardEvent): void {
    // Handle Escape key - stop it immediately when image viewer is open
    if (e.key === 'Escape' && handlers.isImageViewerOpen()) {
      e.stopImmediatePropagation();
      handlers.onClose();
      return;
    }

    // Only handle other keyboard shortcuts when image viewer is open
    if (!handlers.isImageViewerOpen()) return;

    if (e.key === 'ArrowLeft') {
      e.preventDefault();
      handlers.onPrevious();
    } else if (e.key === 'ArrowRight') {
      e.preventDefault();
      handlers.onNext();
    } else if (e.key === '+' || e.key === '=') {
      e.preventDefault();
      handlers.onZoomIn();
    } else if (e.key === '-' || e.key === '_') {
      e.preventDefault();
      handlers.onZoomOut();
    }
  }

  /**
   * Enable keyboard event listeners
   */
  function enable(): void {
    // Use capture phase to handle Escape before other listeners
    window.addEventListener('keydown', handleKeyDown, { capture: true } as any);
  }

  /**
   * Disable keyboard event listeners
   */
  function disable(): void {
    // Remove with capture option to match the addEventListener call
    window.removeEventListener('keydown', handleKeyDown, { capture: true } as any);
  }

  return {
    enable,
    disable,
  };
}
