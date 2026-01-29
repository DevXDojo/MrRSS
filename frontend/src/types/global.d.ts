// Global type declarations

export interface ConfirmDialogOptions {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  isDanger?: boolean;
}

export interface InputDialogOptions {
  title: string;
  message: string;
  placeholder?: string;
  defaultValue?: string;
  confirmText?: string;
  cancelText?: string;
  suggestions?: string[];
}

export interface MultiSelectOption {
  value: string;
  label: string;
  color?: string;
}

export interface MultiSelectDialogOptions {
  title: string;
  message: string;
  options: MultiSelectOption[];
  confirmText?: string;
  cancelText?: string;
}

export type ToastType = 'success' | 'error' | 'info' | 'warning';

declare global {
  interface Window {
    showConfirm: (ConfirmDialogOptions) => Promise<boolean>;
    showInput: (InputDialogOptions) => Promise<string | null>;
    showMultiSelect: (MultiSelectDialogOptions) => Promise<string[] | null>;
    showToast: (string, ToastType?, number?) => void;
  }

  // Browser APIs
  const confirm: (message: string) => boolean;
  const alert: (message: string) => void;
  const prompt: (message: string, defaultValue?: string) => string | null;
}

export {};
