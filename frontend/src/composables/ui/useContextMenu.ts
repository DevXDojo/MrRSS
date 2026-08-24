import { ref } from 'vue';
import type { ContextMenuItem, ContextMenuState } from '@/types/context-menu';

export function useContextMenu() {
  const contextMenu = ref<ContextMenuState>({
    show: false,
    x: 0,
    y: 0,
    items: [],
    data: null,
  });

  function openContextMenu(event: CustomEvent): void {
    contextMenu.value = {
      show: true,
      x: event.detail.x,
      y: event.detail.y,
      items: event.detail.items,
      data: event.detail.data,
      callback: event.detail.callback,
    };
  }

  function closeContextMenu(): void {
    contextMenu.value.show = false;
  }

  async function handleContextMenuAction(action: string): Promise<void> {
    if (contextMenu.value.callback) {
      await contextMenu.value.callback(action, contextMenu.value.data);
    }
  }

  return {
    contextMenu,
    openContextMenu,
    closeContextMenu,
    handleContextMenuAction,
  };
}
