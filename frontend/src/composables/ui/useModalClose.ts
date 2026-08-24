import { onMounted, onUnmounted } from 'vue';

// Global modal stack to track nested modals
const modalStack: Array<{ zIndex: number; close: () => void }> = [];

// Base z-index for modals
// Note: SettingsModal should use a lower z-index (40) so its nested modals (default 50) appear on top
const BASE_Z_INDEX = 50;
// Z-index increment for nested modals
const Z_INDEX_INCREMENT = 10;

// Special z-index for large modals like SettingsModal
export const LARGE_MODAL_Z_INDEX = 40;

// Get the next available z-index
function getNextZIndex(baseZIndex?: number): number {
  if (modalStack.length === 0) {
    return baseZIndex || BASE_Z_INDEX;
  }

  const highestZIndex = Math.max(...modalStack.map((m) => m.zIndex));
  return Math.max(highestZIndex + Z_INDEX_INCREMENT, baseZIndex || BASE_Z_INDEX);
}

export function useModalClose(onClose: () => void, modalZIndex?: number) {
  const zIndex = modalZIndex || getNextZIndex(); // Auto-assign z-index if not provided

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      // Find the modal with the highest z-index
      const highestModal = modalStack.reduce(
        (highest, modal) => {
          return modal.zIndex > (highest?.zIndex || 0) ? modal : highest;
        },
        null as { zIndex: number; close: () => void } | null
      );

      // Only close if this modal is the highest one
      if (highestModal && zIndex === highestModal.zIndex) {
        event.preventDefault();
        event.stopImmediatePropagation(); // Stop other listeners from executing
        onClose();
      }
    }
  }

  onMounted(() => {
    modalStack.push({ zIndex, close: onClose });
    // Use capture phase to ensure this runs before other listeners
    document.addEventListener('keydown', handleKeyDown, { capture: true });
  });

  onUnmounted(() => {
    const index = modalStack.findIndex((m) => m.zIndex === zIndex && m.close === onClose);
    if (index !== -1) {
      modalStack.splice(index, 1);
    }
    // Remove with capture option to match the addEventListener call
    document.removeEventListener('keydown', handleKeyDown, { capture: true } as any);
  });

  return {
    handleKeyDown,
    zIndex, // Return the actual z-index being used
  };
}
