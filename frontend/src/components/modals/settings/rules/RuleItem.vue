<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import {
  PhDotsSixVertical,
  PhFunnel,
  PhListChecks,
  PhPlay,
  PhPencil,
  PhTrash,
} from '@phosphor-icons/vue';
import type { Condition } from '@/composables/rules/useRuleOptions';

const { t } = useI18n();

interface Rule {
  id: number;
  name: string;
  enabled: boolean;
  conditions: Condition[];
  actions: string[];
  position?: number;
}

interface Props {
  rule: Rule;
  isApplying: boolean;
  isDragging?: boolean;
  isDragOver?: boolean;
  isDropBefore?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  'toggle-enabled': [];
  apply: [];
  edit: [];
  delete: [];
  'drag-start': [event: DragEvent];
  'drag-end': [];
  'drag-over': [event: DragEvent];
  'drag-leave': [event: DragEvent];
  drop: [event: DragEvent];
}>();

// Format condition for display
function formatCondition(rule: Rule): string {
  if (!rule.conditions || rule.conditions.length === 0) {
    return t('modal.filter.conditionAlways');
  }

  // Simplified display - show first condition
  const first = rule.conditions[0];
  let text = formatSingleCondition(first);

  if (rule.conditions.length > 1) {
    text += ` ${t('common.text.andNMore', { count: rule.conditions.length - 1 })}`;
  }

  return text;
}

function formatSingleCondition(condition: Condition): string {
  const fieldLabels: Record<string, string> = {
    feed_name: t('modal.feed.feedName'),
    feed_category: t('modal.feed.feedCategory'),
    article_title: t('article.parts.articleTitle'),
    published_after: t('modal.filter.publishedAfter'),
    published_before: t('modal.filter.publishedBefore'),
    is_read: t('modal.filter.readStatus'),
    is_favorite: t('modal.filter.favoriteStatus'),
    is_hidden: t('modal.filter.hiddenStatus'),
  };

  const field = fieldLabels[condition.field] || condition.field;
  const value =
    condition.value || (condition.values && condition.values.length > 0 ? condition.values[0] : '');

  if (condition.negate) {
    return `${t('modal.filter.not')} ${field}: ${value}`;
  }

  return `${field}: ${value}`;
}

// Format actions for display
function formatActions(rule: Rule): string {
  if (!rule.actions || rule.actions.length === 0) {
    return '-';
  }

  const actionLabels: Record<string, string> = {
    favorite: t('setting.rule.actionFavorite'),
    unfavorite: t('setting.rule.actionUnfavorite'),
    hide: t('setting.rule.actionHide'),
    unhide: t('setting.rule.actionUnhide'),
    mark_read: t('setting.rule.actionMarkRead'),
    mark_unread: t('setting.rule.actionMarkUnread'),
    read_later: t('setting.rule.actionReadLater'),
    remove_read_later: t('setting.rule.actionRemoveReadLater'),
  };

  return rule.actions.map((a: string) => actionLabels[a] || a).join(', ');
}
</script>

<template>
  <div
    class="rule-item-container"
    :class="{
      'drop-target': isDragOver && !isDragging,
    }"
  >
    <!-- Drop Indicator Line -->
    <div
      v-if="isDragOver && !isDragging"
      class="drop-indicator"
      :class="{
        'drop-top': isDropBefore,
        'drop-bottom': !isDropBefore,
      }"
    ></div>

    <div
      class="rule-item"
      :class="{
        dragging: isDragging,
      }"
      draggable="true"
      @dragstart="emit('drag-start', $event)"
      @dragend="emit('drag-end')"
      @dragover="emit('drag-over', $event)"
      @dragleave="emit('drag-leave', $event)"
      @drop="emit('drop', $event)"
    >
      <div class="flex items-start gap-2 sm:gap-4">
        <!-- Drag Handle -->
        <div class="drag-handle" @mousedown.stop>
          <PhDotsSixVertical :size="20" class="sm:w-6 sm:h-6" />
        </div>

        <!-- Toggle and Info -->
        <div class="flex-1 flex items-start gap-2 sm:gap-3 min-w-0">
          <input
            type="checkbox"
            :checked="rule.enabled"
            class="toggle mt-1"
            @change="emit('toggle-enabled')"
          />
          <div class="flex-1 min-w-0">
            <div
              class="font-medium mb-1 text-sm sm:text-base truncate"
              :class="{ 'text-text-secondary': !rule.enabled }"
            >
              {{ rule.name || t('modal.rule.rules') + ' #' + rule.id }}
            </div>
            <div class="text-xs text-text-secondary flex flex-wrap items-center gap-1 sm:gap-2">
              <span class="condition-badge">
                <PhFunnel :size="12" />
                {{ formatCondition(rule) }}
              </span>
              <span class="text-text-tertiary">→</span>
              <span class="action-badge">
                <PhListChecks :size="12" />
                {{ formatActions(rule) }}
              </span>
            </div>
          </div>
        </div>

        <!-- Action buttons -->
        <div class="flex items-center gap-1 sm:gap-2 shrink-0">
          <button
            class="action-btn"
            :disabled="isApplying"
            :title="t('setting.rule.applyRuleNow')"
            @click="emit('apply')"
          >
            <PhPlay v-if="!isApplying" :size="18" class="sm:w-5 sm:h-5" />
            <span v-else class="animate-spin text-sm">⟳</span>
          </button>
          <button class="action-btn" :title="t('modal.rule.editRule')" @click="emit('edit')">
            <PhPencil :size="18" class="sm:w-5 sm:h-5" />
          </button>
          <button
            class="action-btn danger"
            :title="t('modal.rule.deleteRule')"
            @click="emit('delete')"
          >
            <PhTrash :size="18" class="sm:w-5 sm:h-5" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";
/* Container for drop zone */
.rule-item-container {
  @apply relative;
  transition: all 0.2s ease;
}

/* The actual rule item */
.rule-item {
  @apply p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border transition-all;
  position: relative;
  cursor: default;
  user-select: none;
}

/* Drag handle styling */
.drag-handle {
  @apply text-text-secondary hover:text-text-primary cursor-grab active:cursor-grabbing shrink-0 mt-1 p-1 rounded -ml-1;
  transition: all 0.2s;
}

.drag-handle:hover {
  @apply bg-bg-tertiary;
}

/* Dragging state - lifted effect */
.rule-item.dragging {
  @apply opacity-40 scale-[0.98];
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
  cursor: grabbing;
}

/* Drop target zone gets a highlight */
.rule-item-container.drop-target .rule-item {
  @apply border-accent;
  opacity: 1;
  background-color: rgb(var(--accent-color) / 0.05);
  border-color: rgb(var(--accent-color) / 0.5);
}

/* Drop indicator line - thick and obvious */
.drop-indicator {
  @apply absolute left-0 right-0 h-0.5 bg-accent rounded-full;
  z-index: 10;
  transition: all 0.15s ease;
}

.drop-indicator.drop-top {
  @apply top-0;
  transform: translateY(-1px);
}

.drop-indicator.drop-bottom {
  @apply bottom-0;
  transform: translateY(1px);
}

/* Make the line thicker and add shadow */
.drop-indicator::after {
  content: '';
  @apply absolute left-0 right-0 h-2 rounded-full -top-[3px];
  background-color: rgb(var(--accent-color) / 0.2);
}

.drop-indicator.drop-bottom::after {
  @apply -bottom-[3px];
}

/* Toggle switch */
.toggle {
  @apply w-10 h-5 appearance-none bg-bg-tertiary rounded-full relative cursor-pointer border border-border transition-colors checked:bg-accent checked:border-accent shrink-0;
}
.toggle::after {
  content: '';
  @apply absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow-sm transition-transform;
}
.toggle:checked::after {
  transform: translateX(20px);
}

/* Badges */
.condition-badge,
.action-badge {
  @apply inline-flex items-center gap-1 px-1.5 sm:px-2 py-0.5 sm:py-1 rounded text-[10px] sm:text-xs bg-bg-tertiary;
}

/* Action buttons */
.action-btn {
  @apply p-1.5 sm:p-2 rounded-lg bg-transparent border-none cursor-pointer text-text-secondary hover:bg-bg-tertiary hover:text-text-primary transition-all;
}

.action-btn.danger:hover {
  @apply text-red-500 bg-red-500/10;
}

.action-btn:disabled {
  @apply opacity-50 cursor-not-allowed;
}

/* Loading spinner */
.animate-spin {
  animation: spin 1s linear infinite;
  display: inline-block;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
