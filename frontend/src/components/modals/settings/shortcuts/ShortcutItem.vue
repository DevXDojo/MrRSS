<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { Component } from 'vue';

const { t } = useI18n();

const isMacOS =
  typeof navigator !== 'undefined' &&
  /Mac|iPhone|iPad|iPod/.test(navigator.platform || navigator.userAgent);

interface ShortcutItem {
  key: string;
  label: string;
  icon: Component;
}

interface Props {
  item: ShortcutItem;
  shortcutValue: string;
  isEditing: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  edit: [];
}>();

// Format key for display
function formatKey(key: string): string {
  if (!key) return t('setting.shortcut.notSet');

  // Convert key combinations to display format
  const parts = key.split('+');
  return parts
    .map((part: string) => {
      // Special key symbols
      if (part === 'Shift') return isMacOS ? '⇧' : 'Shift';
      if (part === 'Control' || part === 'Ctrl') return isMacOS ? '⌃' : 'Ctrl';
      if (part === 'Alt') return isMacOS ? '⌥' : 'Alt';
      if (part === 'Meta' || part === 'Cmd') return isMacOS ? '⌘' : 'Win';
      if (part === 'Enter') return '↵';
      if (part === 'Escape') return 'Esc';
      if (part === 'ArrowUp') return '↑';
      if (part === 'ArrowDown') return '↓';
      if (part === 'ArrowLeft') return '←';
      if (part === 'ArrowRight') return '→';
      if (part === 'Space') return '␣';
      return part.charAt(0).toUpperCase() + part.slice(1);
    })
    .join(' ');
}
</script>

<template>
  <div class="shortcut-row">
    <div class="flex items-center gap-2 flex-1 min-w-0">
      <component :is="item.icon" :size="20" class="text-text-secondary shrink-0" />
      <span class="text-sm font-medium text-text-primary truncate">{{ item.label }}</span>
    </div>
    <button :class="['shortcut-key', isEditing ? 'recording' : '']" @click="emit('edit')">
      <span v-if="isEditing" class="text-accent animate-pulse text-xs sm:text-sm">
        {{ t('shortcut.pressKey') }}
      </span>
      <span v-else>{{ formatKey(shortcutValue) }}</span>
    </button>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";
.shortcut-row {
  @apply flex items-center justify-between gap-3 p-2 rounded-lg bg-bg-secondary border border-border;
}

.shortcut-key {
  @apply px-3 py-1 rounded-md bg-bg-primary border border-border text-sm font-mono text-text-primary min-w-[80px] sm:min-w-[100px] text-center cursor-pointer transition-all;
}

.shortcut-key:hover {
  @apply border-accent bg-bg-tertiary;
}

.shortcut-key.recording {
  @apply border-accent;
  background-color: rgba(59, 130, 246, 0.1);
}

.animate-pulse {
  animation: pulse 1.5s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
