import { ref } from 'vue';
import type {
  ConfirmDialogOptions,
  InputDialogOptions,
  MultiSelectDialogOptions,
  ToastType,
} from '@/types/global';

interface Toast {
  id: number;
  message: string;
  type: ToastType;
  duration: number;
}

interface ConfirmDialogState extends ConfirmDialogOptions {
  onConfirm: () => void;
  onCancel: () => void;
}

interface InputDialogState extends InputDialogOptions {
  onConfirm: (string) => void;
  onCancel: () => void;
}

interface MultiSelectDialogState extends MultiSelectDialogOptions {
  onConfirm: (values: string[]) => void;
  onCancel: () => void;
}

export function useNotifications() {
  const confirmDialog = ref<ConfirmDialogState | null>(null);
  const inputDialog = ref<InputDialogState | null>(null);
  const multiSelectDialog = ref<MultiSelectDialogState | null>(null);
  const toasts = ref<Toast[]>([]);

  // Track recent toast messages to prevent duplicates
  // Map of message -> last shown timestamp
  const recentToasts = ref<Map<string, number>>(new Map());
  const TOAST_DEBOUNCE_MS = 2000; // Wait 2 seconds before showing same toast again

  function showConfirm(options: ConfirmDialogOptions): Promise<boolean> {
    return new Promise((resolve) => {
      confirmDialog.value = {
        ...options,
        onConfirm: () => {
          confirmDialog.value = null;
          resolve(true);
        },
        onCancel: () => {
          confirmDialog.value = null;
          resolve(false);
        },
      };
    });
  }

  function showInput(options: InputDialogOptions): Promise<string | null> {
    return new Promise((resolve) => {
      inputDialog.value = {
        ...options,
        onConfirm: (value: string) => {
          inputDialog.value = null;
          resolve(value);
        },
        onCancel: () => {
          inputDialog.value = null;
          resolve(null);
        },
      };
    });
  }

  function showMultiSelect(options: MultiSelectDialogOptions): Promise<string[] | null> {
    return new Promise((resolve) => {
      multiSelectDialog.value = {
        ...options,
        onConfirm: (values: string[]) => {
          multiSelectDialog.value = null;
          resolve(values);
        },
        onCancel: () => {
          multiSelectDialog.value = null;
          resolve(null);
        },
      };
    });
  }

  function showToast(message: string, type: ToastType = 'info', duration: number = 3000): void {
    const now = Date.now();
    const toastKey = `${type}:${message}`;

    // Check if we've shown this exact toast recently
    const lastShown = recentToasts.value.get(toastKey);
    if (lastShown && now - lastShown < TOAST_DEBOUNCE_MS) {
      // Skip showing this toast - it was shown recently
      console.debug(`[Toast] Debounced duplicate toast: "${message}"`);
      return;
    }

    // Mark this toast as shown
    recentToasts.value.set(toastKey, now);

    // Clean up old entries to prevent memory leak
    setTimeout(() => {
      const entryTime = recentToasts.value.get(toastKey);
      if (entryTime && now - entryTime >= TOAST_DEBOUNCE_MS) {
        recentToasts.value.delete(toastKey);
      }
    }, TOAST_DEBOUNCE_MS + 100);

    const id = Date.now();
    toasts.value.push({ id, message, type, duration });
  }

  function removeToast(id: number): void {
    toasts.value = toasts.value.filter((t) => t.id !== id);
  }

  // Make these available globally
  function installGlobalHandlers(): void {
    window.showConfirm = showConfirm;
    window.showInput = showInput;
    window.showMultiSelect = showMultiSelect;
    window.showToast = showToast;
  }

  return {
    confirmDialog,
    inputDialog,
    multiSelectDialog,
    toasts,
    showConfirm,
    showInput,
    showMultiSelect,
    showToast,
    removeToast,
    installGlobalHandlers,
  };
}
