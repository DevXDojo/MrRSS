<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhFunnel, PhPencil, PhTrash } from '@phosphor-icons/vue';
import type { SavedFilter } from '@/types/filter';

const { t } = useI18n();

interface Props {
  filter: SavedFilter;
  isActive: boolean;
  isDragging?: boolean;
  isEditMode?: boolean;
  compactMode?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  isDragging: false,
  isEditMode: false,
  compactMode: false,
});

const emit = defineEmits<{
  click: [];
  contextmenu: [event: MouseEvent];
  dragstart: [event: Event];
  dragend: [];
  edit: [filter: SavedFilter];
  delete: [filter: SavedFilter];
}>();

// Handle click
function handleClick() {
  emit('click');
}

// Handle context menu
function handleContextMenu(event: MouseEvent) {
  emit('contextmenu', event);
}

// Handle drag start
function handleDragStart(event: Event) {
  emit('dragstart', event);
}

// Handle drag end
function handleDragEnd() {
  emit('dragend');
}

// Handle edit
function handleEdit(event: Event) {
  event.stopPropagation();
  emit('edit', props.filter);
}

// Handle delete
function handleDelete(event: Event) {
  event.stopPropagation();
  emit('delete', props.filter);
}
</script>

<template>
  <div
    :class="[
      'cursor-pointer select-none rounded-md transition-all duration-200 border border-transparent flex items-center justify-between gap-1.5 sm:gap-2.5',
      isActive
        ? 'bg-bg-tertiary hover:bg-bg-tertiary text-accent font-medium'
        : 'bg-transparent hover:bg-bg-tertiary',
      props.compactMode ? 'px-1 sm:px-1.5 py-0.5 sm:py-1' : 'px-2 sm:px-3 py-1.5 sm:py-2 mx-1',
      isDragging ? 'opacity-50' : '',
    ]"
    draggable="true"
    @click="handleClick"
    @contextmenu="handleContextMenu"
    @dragstart="handleDragStart"
    @dragend="handleDragEnd"
  >
    <!-- Icon and filter name -->
    <div class="flex items-center gap-0.5 sm:gap-1 flex-1 min-w-0">
      <PhFunnel :size="18" class="flex-shrink-0 text-accent" />
      <span
        :class="[
          'whitespace-nowrap overflow-hidden text-ellipsis text-xs sm:text-sm font-medium',
          isActive ? 'text-accent' : 'text-text-primary',
        ]"
        >{{ filter.name }}</span
      >
    </div>

    <!-- Edit mode actions -->
    <div v-if="isEditMode" class="flex gap-1">
      <button
        class="bg-transparent border-0 p-1 cursor-pointer text-text-secondary rounded transition-all duration-200 hover:bg-bg-secondary hover:text-text-primary"
        :title="t('common.edit')"
        @click="handleEdit"
      >
        <PhPencil :size="14" />
      </button>
      <button
        class="bg-transparent border-0 p-1 cursor-pointer text-text-secondary rounded transition-all duration-200 hover:text-red-500 hover:bg-[rgba(239,68,68,0.1)]"
        :title="t('common.delete')"
        @click="handleDelete"
      >
        <PhTrash :size="14" />
      </button>
    </div>
  </div>
</template>
