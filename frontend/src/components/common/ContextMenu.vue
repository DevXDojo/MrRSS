<script setup lang="ts">
import {
  ref,
  onMounted,
  onUnmounted,
  computed,
  nextTick,
  watch,
  type Ref,
  type Component,
} from 'vue';
import * as PhosphorIcons from '@phosphor-icons/vue';
import type { ContextMenuItem } from '@/types/context-menu';

interface Props {
  items: ContextMenuItem[];
  x: number;
  y: number;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  action: [action: string];
}>();

const menuRef: Ref<HTMLDivElement | null> = ref(null);
const adjustedPosition = ref({ top: 0, left: 0 });

// Map old icon names to new component names
const iconMap: Record<string, string> = {
  'ph-link': 'PhLink',
  'ph-text-t': 'PhTextT',
  'ph-check-circle': 'PhCheckCircle',
  'ph-globe': 'PhGlobe',
  'ph-pencil': 'PhPencil',
  'ph-trash': 'PhTrash',
  'ph-envelope': 'PhEnvelope',
  'ph-envelope-open': 'PhEnvelopeOpen',
  'ph-star': 'PhStar',
  'ph-article': 'PhArticle',
  'ph-eye': 'PhEye',
  'ph-eye-slash': 'PhEyeSlash',
  'ph-arrow-square-out': 'PhArrowSquareOut',
  'ph-clock-countdown': 'PhClockCountdown',
  'ph-arrow-bend-right-up': 'PhArrowBendRightUp',
  'ph-arrow-bend-left-down': 'PhArrowBendLeftDown',
  PhMagnifyingGlass: 'PhMagnifyingGlass',
  PhArrowsClockwise: 'PhArrowsClockwise',
  PhMagnifyingGlassPlus: 'PhMagnifyingGlassPlus',
  PhDownloadSimple: 'PhDownloadSimple',
};

// Get icon component from icon string
function getIconComponent(iconName?: string): Component | null {
  if (!iconName) return null;
  const componentName = iconMap[iconName] || iconName;
  return (PhosphorIcons as Record<string, Component>)[componentName] || null;
}

function handleClickOutside(event: MouseEvent) {
  if (menuRef.value && event.target instanceof Node && !menuRef.value.contains(event.target)) {
    emit('close');
  }
}

// Adjust position to keep menu within viewport
function adjustMenuPosition() {
  nextTick(() => {
    if (!menuRef.value) {
      adjustedPosition.value = { top: props.y, left: props.x };
      return;
    }

    const menuRect = menuRef.value.getBoundingClientRect();
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;

    let newTop = props.y;
    let newLeft = props.x;

    // Check if menu goes beyond right edge
    if (props.x + menuRect.width > viewportWidth - 10) {
      newLeft = props.x - menuRect.width;
      if (newLeft < 10) newLeft = 10;
    }

    // Check if menu goes beyond bottom edge
    if (props.y + menuRect.height > viewportHeight - 10) {
      newTop = props.y - menuRect.height;
      if (newTop < 10) newTop = 10;
    }

    adjustedPosition.value = { top: newTop, left: newLeft };
  });
}

onMounted(() => {
  // Adjust position immediately
  adjustMenuPosition();

  // Use setTimeout to avoid catching the event that opened the menu
  setTimeout(() => {
    document.addEventListener('click', handleClickOutside);
    document.addEventListener('contextmenu', handleClickOutside);
  }, 0);
});

// Watch for position changes and re-adjust
watch(
  () => [props.x, props.y],
  () => {
    adjustMenuPosition();
  }
);

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
  document.removeEventListener('contextmenu', handleClickOutside);
});

function handleAction(item: ContextMenuItem) {
  if (item.disabled) return;
  if (item.action) {
    emit('action', item.action);
  }
  emit('close');
}

const menuStyle = computed(() => ({
  top: `${adjustedPosition.value.top}px`,
  left: `${adjustedPosition.value.left}px`,
}));
</script>

<template>
  <div
    ref="menuRef"
    class="fixed z-50 bg-bg-primary border border-border rounded-lg shadow-xl py-1 min-w-[180px] animate-fade-in"
    :style="menuStyle"
  >
    <template v-for="(item, index) in items" :key="index">
      <div v-if="item.separator" class="h-px bg-border my-1"></div>
      <div
        v-else
        class="px-4 py-2 flex items-center gap-3 cursor-pointer hover:bg-bg-tertiary text-sm transition-colors"
        :class="[
          item.disabled ? 'opacity-50 cursor-not-allowed' : '',
          item.danger
            ? 'text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20'
            : 'text-text-primary',
        ]"
        @click="handleAction(item)"
      >
        <component
          :is="getIconComponent(item.icon)"
          v-if="item.icon && getIconComponent(item.icon)"
          :size="20"
          :weight="item.iconWeight || 'regular'"
          :class="
            item.iconColor ||
            (item.danger ? 'text-red-600 dark:text-red-400' : 'text-text-secondary')
          "
        />
        <span>{{ item.label }}</span>
      </div>
    </template>
  </div>
</template>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.1s ease-out;
}
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
