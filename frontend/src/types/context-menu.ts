/**
 * Shared type definitions for context menu system
 */

export interface ContextMenuItem {
  label?: string;
  action?: string;
  icon?: string;
  iconWeight?: 'regular' | 'bold' | 'light' | 'fill' | 'duotone' | 'thin';
  iconColor?: string;
  disabled?: boolean;
  danger?: boolean;
  separator?: boolean;
}

export interface ContextMenuState {
  show: boolean;
  x: number;
  y: number;
  items: ContextMenuItem[];
  data: unknown;
  callback?: (action: string, data: unknown) => void | Promise<void>;
}
