<script setup lang="ts">
import { computed, ref } from 'vue';
import { PhFolder, PhFolderDashed, PhCaretDown } from '@phosphor-icons/vue';
import { useI18n } from 'vue-i18n';
import type { Feed } from '@/types/models';
import type { DropPreview } from '@/composables/ui/useDragDrop';
import SidebarFeed from './SidebarFeed.vue';

const { t } = useI18n();

// Track click timeout to distinguish single click from double click
const clickTimeout = ref<ReturnType<typeof setTimeout> | null>(null);

interface TreeNode {
  _feeds: Feed[];
  _children: Record<string, TreeNode>;
  isOpen: boolean;
}

interface Props {
  name: string;
  feeds: Feed[];
  isOpen: boolean;
  isActive: boolean;
  isUncategorized?: boolean;
  unreadCount: number;
  currentFeedId: number | null;
  feedUnreadCounts: Record<number, number>;
  isDragOver?: boolean;
  isEditMode?: boolean;
  dropPreview?: DropPreview;
  draggingFeedId?: number | null;
  // Multi-level support
  children?: Record<string, TreeNode>;
  level?: number;
  categoryPath?: string;
  isCategoryOpen?: (path: string) => boolean;
  compactMode?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  children: undefined,
  level: 0,
  categoryPath: '',
  isCategoryOpen: undefined,
  dropPreview: undefined,
  draggingFeedId: null,
  compactMode: false,
});

const emit = defineEmits<{
  toggle: [];
  selectCategory: [path: string];
  selectFeed: [feedId: number];
  categoryContextMenu: [event: MouseEvent, path: string];
  feedContextMenu: [event: MouseEvent, feed: Feed];
  feedDragOver: [feedId: number | null, event: Event];
  drop: [];
  dragstart: [feedId: number, event: Event];
  dragend: [];
  dragleave: [categoryName: string, event: Event];
  categoryDragOver: [categoryName: string, event: Event];
  // Multi-level events
  childToggle: [path: string];
  childSelectCategory: [path: string];
  childContextMenu: [event: MouseEvent, path: string];
}>();

// Handle dragover on the feeds-list container using event delegation
function handleFeedsListDragOver(event: DragEvent) {
  // Prevent default to allow drop
  event.preventDefault();
  // Stop propagation to prevent triggering parent handlers
  event.stopPropagation();

  // Find which feed item we're hovering over
  const target = event.target as HTMLElement;
  const feedItem = target.closest('.feed-item');

  if (feedItem) {
    // Get the feed ID from the data attribute
    const feedIdStr = feedItem.getAttribute('data-feed-id');
    const feedId = feedIdStr ? parseInt(feedIdStr, 10) : null;
    console.log('[SidebarCategory] Emitting feedDragOver with feedId:', feedId, 'event:', event);
    emit('feedDragOver', feedId, event);
  } else {
    // Not hovering over any specific feed (in gaps between feeds or empty space)
    // Calculate which feed we're closest to based on Y position
    const feedsList = event.currentTarget as HTMLElement;
    if (feedsList) {
      const feedItems = Array.from(feedsList.querySelectorAll('.feed-item'));

      // If there are no feed items (empty category), emit null immediately
      if (feedItems.length === 0) {
        console.log('[SidebarCategory] Empty category, emitting feedDragOver with null feedId');
        emit('feedDragOver', null, event);
        return;
      }

      const listRect = feedsList.getBoundingClientRect();
      const mouseY = event.clientY - listRect.top;

      // Find the feed item closest to the mouse Y position
      let closestFeedId: number | null = null;
      let minDistance = Infinity;

      for (const item of feedItems) {
        const rect = item.getBoundingClientRect();
        const itemCenterY = (rect.top + rect.bottom) / 2 - listRect.top;
        const distance = Math.abs(mouseY - itemCenterY);

        if (distance < minDistance) {
          minDistance = distance;
          const feedIdStr = item.getAttribute('data-feed-id');
          closestFeedId = feedIdStr ? parseInt(feedIdStr, 10) : null;
        }
      }

      console.log('[SidebarCategory] In gap, closest feedId:', closestFeedId, 'event:', event);
      emit('feedDragOver', closestFeedId, event);
    } else {
      console.log('[SidebarCategory] Emitting feedDragOver with null feedId, event:', event);
      emit('feedDragOver', null, event);
    }
  }
}

function handleDrop(event: DragEvent) {
  event.preventDefault();
  emit('drop');
}

// Handle dragover on category container (for dropping at category level)
function handleCategoryDragOver(event: DragEvent) {
  event.preventDefault();
  event.stopPropagation();
  // Notify parent that we're dragging over this category
  emit('categoryDragOver', props.name, event);
  // Also emit feedDragOver for drop preview
  emit('feedDragOver', null, event);
}

// Computed properties for child categories
const hasChildren = computed(() => {
  return props.children && Object.keys(props.children).length > 0;
});

// Get the full category path for this node
const fullPath = computed(() => {
  // For uncategorized category, use empty string to match database (feeds with no category)
  if (props.isUncategorized) {
    return '';
  }
  return props.categoryPath ? `${props.categoryPath}/${props.name}` : props.name;
});

// Check if a category should be open
const checkIsOpen = (path: string) => {
  if (props.isCategoryOpen) {
    return props.isCategoryOpen(path);
  }
  return false;
};

// Check if this category is exclusively for FreshRSS feeds
// Only show the icon if ALL feeds in this category are from FreshRSS
const isFreshRSSCategory = computed(() => {
  if (!props.feeds || props.feeds.length === 0) {
    return false;
  }
  // Only show FreshRSS icon if ALL feeds in this category are FreshRSS sources
  return props.feeds.every((feed) => feed.is_freshrss_source);
});

// Handle category header click - delays to distinguish from double click
function handleCategoryClick() {
  // Clear any existing timeout
  if (clickTimeout.value) {
    clearTimeout(clickTimeout.value);
  }

  // Set timeout to execute single-click action if no double-click follows
  clickTimeout.value = setTimeout(() => {
    emit('selectCategory', fullPath.value);
    clickTimeout.value = null;
  }, 250); // 250ms delay to wait for potential double-click
}

// Handle category header double-click - toggles expand/collapse
function handleCategoryDoubleClick() {
  // Clear the timeout to prevent single-click action from executing
  if (clickTimeout.value) {
    clearTimeout(clickTimeout.value);
    clickTimeout.value = null;
  }
  // Toggle expand/collapse
  emit('toggle');
}

// Handle caret click - toggles expand/collapse and ensures context menu closes
function handleCaretClick() {
  emit('toggle');
  // Manually trigger a click event to ensure context menu closes
  // The click.stop modifier prevents event bubbling, so we need to manually trigger it
  document.dispatchEvent(new MouseEvent('click', { bubbles: true }));
}
</script>

<template>
  <div
    :class="[
      'category-container',
      isDragOver ? 'drag-over' : '',
      props.compactMode ? 'mb-0.5' : 'mb-1',
    ]"
    :data-level="level"
    @dragover.self="handleCategoryDragOver"
    @dragleave.self="(e) => $emit('dragleave', props.name, e)"
    @drop.self.prevent="handleDrop"
  >
    <div
      :class="['category-header', isActive ? 'active' : '', props.compactMode ? 'compact' : '']"
      @click="handleCategoryClick"
      @dblclick="handleCategoryDoubleClick"
      @contextmenu="(e) => emit('categoryContextMenu', e, fullPath)"
      @dragover="handleCategoryDragOver"
    >
      <span class="flex-1 flex items-center gap-2">
        <PhFolderDashed v-if="isUncategorized" :size="20" />
        <PhFolder v-else :size="20" :weight="'fill'" />
        {{ name }}
        <!-- FreshRSS indicator on category -->
        <!-- Only show if ALL feeds in this category are from FreshRSS -->
        <img
          v-if="isFreshRSSCategory"
          src="/assets/plugin_icons/freshrss.svg"
          class="w-3.5 h-3.5 shrink-0"
          :title="t('setting.freshrss.syncedFeed')"
          alt="FreshRSS"
        />
      </span>
      <span v-if="unreadCount > 0" class="unread-badge mr-1">{{ unreadCount }}</span>
      <PhCaretDown
        :size="20"
        class="p-1 cursor-pointer transition-transform text-text-secondary"
        :class="{ 'rotate-180': isOpen }"
        @click.stop="handleCaretClick"
      />
    </div>
    <div
      v-show="isOpen"
      class="feeds-list"
      :class="props.compactMode ? 'pl-1' : 'pl-2'"
      @dragover="handleFeedsListDragOver"
      @drop.prevent="handleDrop"
    >
      <template v-for="feed in feeds" :key="feed.id">
        <div class="feed-wrapper">
          <!-- Drop indicator above this feed -->
          <div
            v-if="
              isDragOver &&
              dropPreview &&
              dropPreview.targetFeedId === feed.id &&
              dropPreview.beforeTarget
            "
            class="drop-indicator"
            style="top: -1.5px"
          ></div>
          <SidebarFeed
            :feed="feed"
            :is-active="currentFeedId === feed.id"
            :unread-count="feedUnreadCounts[feed.id] || 0"
            :is-edit-mode="isEditMode"
            :level="level"
            :compact-mode="props.compactMode"
            @click="emit('selectFeed', feed.id)"
            @contextmenu="(e) => emit('feedContextMenu', e, feed)"
            @dragstart="(e) => emit('dragstart', feed.id, e)"
            @dragend="emit('dragend')"
          />
          <!-- Drop indicator below this feed -->
          <div
            v-if="
              isDragOver &&
              dropPreview &&
              dropPreview.targetFeedId === feed.id &&
              !dropPreview.beforeTarget
            "
            class="drop-indicator"
            style="bottom: -1.5px"
          ></div>
        </div>
      </template>

      <!-- Drop indicator for empty category or at the end when dragging over -->
      <div
        v-if="isDragOver && dropPreview && dropPreview.targetFeedId === null"
        class="drop-indicator end-indicator"
        :class="{ 'empty-category-indicator': feeds.length === 0 }"
      ></div>

      <!-- Child categories (multi-level support) -->
      <template v-if="hasChildren">
        <SidebarCategory
          v-for="(childData, childName) in children"
          :key="childName"
          :name="childName"
          :feeds="childData._feeds"
          :children="childData._children"
          :level="level + 1"
          :category-path="fullPath"
          :is-open="checkIsOpen(fullPath + '/' + childName)"
          :is-active="false"
          :unread-count="0"
          :current-feed-id="currentFeedId"
          :feed-unread-counts="feedUnreadCounts"
          :is-drag-over="false"
          :is-edit-mode="isEditMode"
          :dragging-feed-id="draggingFeedId"
          :is-category-open="props.isCategoryOpen"
          :compact-mode="props.compactMode"
          @toggle="emit('childToggle', fullPath + '/' + childName)"
          @select-category="(path) => emit('childSelectCategory', path)"
          @category-context-menu="(e, path) => emit('childContextMenu', e, path)"
          @child-toggle="(path) => emit('childToggle', path)"
          @child-select-category="(path) => emit('childSelectCategory', path)"
          @child-context-menu="(e, path) => emit('childContextMenu', e, path)"
          @select-feed="(feedId) => emit('selectFeed', feedId)"
          @feed-context-menu="(e, feed) => emit('feedContextMenu', e, feed)"
        />
      </template>
    </div>
  </div>
</template>

<style scoped>
@reference "../../style.css";
.category-header {
  @apply px-2 sm:px-3 py-1.5 sm:py-2 cursor-pointer font-semibold text-xs sm:text-sm text-text-secondary flex items-center justify-between hover:bg-bg-tertiary hover:text-text-primary transition-colors;
  @apply sticky z-10 bg-bg-secondary;
  top: -0.375rem; /* matches container's p-1.5 */
  margin-left: -0.375rem;
  margin-right: -0.375rem;
  padding-left: calc(0.5rem + 0.375rem);
  padding-right: calc(0.75rem + 0.375rem);
}

/* Compact mode: reduce padding for category headers */
.category-header.compact {
  @apply px-1.5 sm:px-2 py-1 sm:py-1.5;
}

/* Indentation for nested categories */
.category-container[data-level='1'] .category-header {
  padding-left: calc(0.5rem + 0.375rem + 1rem);
}

.category-container[data-level='2'] .category-header {
  padding-left: calc(0.5rem + 0.375rem + 2rem);
}

.category-container[data-level='3'] .category-header {
  padding-left: calc(0.5rem + 0.375rem + 3rem);
}

.category-container[data-level='4'] .category-header {
  padding-left: calc(0.5rem + 0.375rem + 4rem);
}

/* Compact mode indentation for nested categories */
.category-container[data-level='1'] .category-header.compact {
  padding-left: calc(0.375rem + 0.375rem + 1rem);
}

.category-container[data-level='2'] .category-header.compact {
  padding-left: calc(0.375rem + 0.375rem + 2rem);
}

.category-container[data-level='3'] .category-header.compact {
  padding-left: calc(0.375rem + 0.375rem + 3rem);
}

.category-container[data-level='4'] .category-header.compact {
  padding-left: calc(0.375rem + 0.375rem + 4rem);
}

/* Special styling for category header when its container is a drag target */
.category-container.drag-over .category-header {
  @apply text-accent font-bold;
  background-color: transparent;
}
@media (min-width: 640px) {
  .category-header {
    top: -0.5rem; /* matches container's sm:p-2 */
    margin-left: -0.5rem;
    margin-right: -0.5rem;
    padding-left: calc(0.75rem + 0.5rem);
    padding-right: calc(0.75rem + 0.5rem);
  }

  /* Compact mode at sm breakpoint */
  .category-header.compact {
    top: -0.25rem; /* matches container's compact p-1 */
    margin-left: -0.25rem;
    margin-right: -0.25rem;
    padding-left: calc(0.5rem + 0.25rem);
    padding-right: calc(0.5rem + 0.25rem);
  }

  /* Compact mode indentation for nested categories at sm breakpoint */
  .category-container[data-level='1'] .category-header.compact {
    padding-left: calc(0.5rem + 0.25rem + 1rem);
  }

  .category-container[data-level='2'] .category-header.compact {
    padding-left: calc(0.5rem + 0.25rem + 2rem);
  }

  .category-container[data-level='3'] .category-header.compact {
    padding-left: calc(0.5rem + 0.25rem + 3rem);
  }

  .category-container[data-level='4'] .category-header.compact {
    padding-left: calc(0.5rem + 0.25rem + 4rem);
  }
}
.category-header.active {
  @apply bg-bg-tertiary text-accent;
}

/* Container drag-over styling */
.category-container.drag-over {
  @apply rounded-lg;
  outline: 2px solid var(--accent-color, #007bff);
  outline-offset: -2px;
  background-color: var(--bg-tertiary, rgba(0, 123, 255, 0.05));
}

.feeds-list {
  position: relative;
  min-height: 40px; /* Ensure empty categories have a drop zone */
}

/* Wrapper to position drop indicators relative to each feed */
.feed-wrapper {
  position: relative;
}

/* Drop indicator positioned absolutely to avoid layout shift */
.drop-indicator {
  position: absolute;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent, var(--accent-color, #007bff), transparent);
  border-radius: 1.5px;
  animation: pulse-indicator 1.5s ease-in-out infinite;
  pointer-events: none;
  z-index: 10;
}

/* End indicator positioned relative to feeds list */
.drop-indicator.end-indicator {
  position: relative;
  margin-top: 2px;
  margin-bottom: 2px;
}

/* Empty category indicator - more prominent */
.drop-indicator.empty-category-indicator {
  height: 4px;
  margin-top: 8px;
  margin-bottom: 8px;
}

@keyframes pulse-indicator {
  0%,
  100% {
    opacity: 0.6;
  }
  50% {
    opacity: 1;
  }
}

.unread-badge {
  @apply text-[9px] sm:text-[10px] font-medium rounded-full min-w-[14px] sm:min-w-[16px] h-[14px] sm:h-[16px] px-0.5 sm:px-1 flex items-center justify-center;
  background-color: rgba(120, 120, 120, 0.15);
  color: #666666;
}
</style>

<style>
@reference "../../style.css";
.dark-mode .unread-badge {
  /* This style will be applied to child components, so it can not use scoped */
  background-color: rgba(100, 100, 100, 0.4) !important;
  color: #d0d0d0 !important;
}
</style>
