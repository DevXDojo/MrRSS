<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { useDragDrop } from '@/composables/ui/useDragDrop';
import { useSidebar } from '@/composables/core/useSidebar';
import { useSettings } from '@/composables/core/useSettings';
import { useArticleFilter } from '@/composables/article/useArticleFilter';
import { useSavedFilters } from '@/composables/article/useSavedFilters';
import SidebarCategory from './SidebarCategory.vue';
import SavedFilterItem from './SavedFilterItem.vue';
import SavedFilterModal from '@/components/modals/filter/SavedFilterModal.vue';
import {
  PhMagnifyingGlass,
  PhX,
  PhPencil,
  PhCheck,
  PhPushPin,
  PhFloppyDisk,
} from '@phosphor-icons/vue';
import type { Feed } from '@/types/models';
import type { FilterCondition, SavedFilter } from '@/types/filter';

const props = defineProps<{
  isExpanded?: boolean;
  isPinned?: boolean;
}>();

const emit = defineEmits<{
  expand: [];
  collapse: [];
  pin: [];
  unpin: [];
}>();

const store = useAppStore();
const { t } = useI18n();
const { settings, fetchSettings } = useSettings();

// Saved filters
const {
  savedFilters,
  fetchSavedFilters,
  createSavedFilter,
  updateSavedFilter,
  deleteSavedFilter,
  parseConditions,
} = useSavedFilters();

const { activeFilters, fetchFilteredArticles } = useArticleFilter();

// Safe computed for active filters check
const hasActiveFilters = computed(() => {
  return activeFilters.value && activeFilters.value.length > 0;
});

// Deep compare two arrays of filter conditions
function conditionsEqual(a: FilterCondition[], b: FilterCondition[]): boolean {
  if (a.length !== b.length) return false;

  for (let i = 0; i < a.length; i++) {
    const condA = a[i];
    const condB = b[i];

    // Compare each field
    if (
      condA.field !== condB.field ||
      condA.operator !== condB.operator ||
      condA.value !== condB.value ||
      condA.negate !== condB.negate
    ) {
      return false;
    }

    // Compare values array if exists
    const valuesA = condA.values || [];
    const valuesB = condB.values || [];
    if (valuesA.length !== valuesB.length) return false;

    for (let j = 0; j < valuesA.length; j++) {
      if (valuesA[j] !== valuesB[j]) {
        return false;
      }
    }
  }
  return true;
}

// Safe current filters value for modal
const currentFiltersValue = computed(() => {
  return activeFilters.value || [];
});

// Create a computed Set of active filter IDs for efficient lookup
const activeFilterIds = computed(() => {
  if (!activeFilters.value || activeFilters.value.length === 0) {
    return new Set<number>();
  }
  const ids = new Set<number>();
  for (const filter of safeSavedFilters.value) {
    const savedConditions = parseConditions(filter.conditions);
    if (conditionsEqual(savedConditions, activeFilters.value)) {
      ids.add(filter.id);
    }
  }
  return ids;
});

const isFilterActive = (filter: SavedFilter) => {
  return activeFilterIds.value.has(filter.id);
};

// Safe computed for saved filters
const safeSavedFilters = computed(() => {
  return Array.isArray(savedFilters.value) ? savedFilters.value : [];
});

// Saved filters UI state
const showSaveFilterModal = ref(false);
const showEditFilterModal = ref(false);
const editingFilter = ref<SavedFilter | null>(null);
const draggingFilterId = ref<number | null>(null);

// Compact mode setting (layout_mode === 'compact')
const compactMode = computed(() => {
  return settings.value.layout_mode === 'compact';
});

// Initialize settings on mount
onMounted(async () => {
  try {
    await fetchSettings();
    await fetchSavedFilters();
  } catch (e) {
    console.error('Error loading settings in FeedList:', e);
  }

  // Listen for layout mode changes
  window.addEventListener('layout-mode-changed', handleLayoutModeChange);
  // Listen for category expansion events
  window.addEventListener('categories-expanded', handleCategoriesExpanded);
});

// Handle layout mode changes
function handleLayoutModeChange() {
  fetchSettings().catch((e) => {
    console.error('Error re-fetching settings after layout mode change:', e);
  });
}

// Handle categories expanded event
function handleCategoriesExpanded() {
  // Reload openCategories from localStorage
  const savedCategories = localStorage.getItem('openCategories');
  if (savedCategories) {
    const categories = JSON.parse(savedCategories);
    openCategories.value = new Set(categories);
  }
}

onUnmounted(() => {
  window.removeEventListener('layout-mode-changed', handleLayoutModeChange);
  window.removeEventListener('categories-expanded', handleCategoriesExpanded);
});

// Edit mode for drag reordering
const isEditMode = ref(false);

// Local state to track if user is actively dragging
const isDragging = ref(false);
let dropHandled = false;

function toggleEditMode() {
  isEditMode.value = !isEditMode.value;
}

const {
  tree,
  categoryUnreadCounts,
  feedUnreadCounts,
  toggleCategory,
  isCategoryOpen: checkIsCategoryOpen,
  searchQuery,
  onFeedContextMenu,
  onCategoryContextMenu,
  openCategories,
} = useSidebar();

// Track if we should collapse after selection
let shouldCollapseAfterSelection = false;

// Watch for feed/category selection to auto-collapse
watch(
  () => store.currentFeedId,
  (newVal, oldVal) => {
    // Only collapse if:
    // 1. Not pinned
    // 2. Is currently expanded
    // 3. The change was triggered by user action (not initial load or programmatic change)
    // 4. We haven't just collapsed (prevent double-collapse)
    if (!props.isPinned && props.isExpanded && shouldCollapseAfterSelection && newVal !== oldVal) {
      shouldCollapseAfterSelection = false;
      setTimeout(() => {
        emit('collapse');
      }, 200);
    }
  }
);

watch(
  () => store.currentCategory,
  (newVal, oldVal) => {
    // Only collapse if:
    // 1. Not pinned
    // 2. Is currently expanded
    // 3. The change was triggered by user action
    // 4. We haven't just collapsed
    if (!props.isPinned && props.isExpanded && shouldCollapseAfterSelection && newVal !== oldVal) {
      shouldCollapseAfterSelection = false;
      setTimeout(() => {
        emit('collapse');
      }, 200);
    }
  }
);

// Mark that we should collapse after a feed/category is selected
function handleFeedOrCategorySelect() {
  if (!props.isPinned) {
    shouldCollapseAfterSelection = true;
  }
}

// Wrapper functions for feed/category selection
function handleSelectFeed(feedId: number) {
  handleFeedOrCategorySelect();
  store.setFeed(feedId);
}

function handleSelectCategory(category: string) {
  handleFeedOrCategorySelect();
  store.setCategory(category);
}

// Drag and drop functionality
const {
  draggingFeedId,
  dragOverCategory,
  dropPreview,
  onDragStart,
  onDragEnd,
  onDragOver: onDragOverComposable,
  onDragLeave: onDragLeaveComposable,
  onDrop,
} = useDragDrop();

// Handle drag events
function handleDragStart(feedId: number, event: Event) {
  const feed = store.feeds?.find((f) => f.id === feedId);
  if (feed?.is_freshrss_source) {
    event.preventDefault();
    window.showToast(t('setting.freshrss.feedLocked'), 'info');
    return;
  }

  isDragging.value = true;
  dropHandled = false;
  onDragStart(feedId, event);
}

function handleDragEnd() {
  onDragEnd();
}

function handleDragOver(categoryName: string, feedId: number | null, event: Event) {
  onDragOverComposable(categoryName, feedId, event);
}

function handleCategoryDragOver(categoryName: string, event: Event) {
  onDragOverComposable(categoryName, null, event);
}

function handleDragLeave(categoryName: string, event: Event) {
  onDragLeaveComposable(categoryName, event);
}

async function handleDrop(categoryName: string, feeds: any[]) {
  if (dropHandled) {
    return;
  }

  dropHandled = true;

  const draggedFeed = store.feeds?.find((f) => f.id === draggingFeedId.value);

  // Prevent dropping into FreshRSS categories
  const targetCategoryFeeds = feeds.filter(
    (f) => f.category === categoryName || (categoryName === 'uncategorized' && !f.category)
  );
  const hasFreshRSSFeedInTarget = targetCategoryFeeds.some((f) => f.is_freshrss_source);
  const isFreshRSSCategoryByName =
    categoryName.endsWith(' (FreshRSS)') || categoryName.match(/ \(FreshRSS \d+\)$/);

  if (draggedFeed && !draggedFeed.is_freshrss_source) {
    if (hasFreshRSSFeedInTarget || isFreshRSSCategoryByName) {
      window.showToast(t('setting.freshrss.feedLocked'), 'info');
      isDragging.value = false;
      return;
    }
  }

  if (draggedFeed && draggedFeed.is_freshrss_source) {
    if (!hasFreshRSSFeedInTarget && !isFreshRSSCategoryByName && targetCategoryFeeds.length > 0) {
      window.showToast(t('setting.freshrss.feedLocked'), 'info');
      isDragging.value = false;
      return;
    }
  }

  try {
    const result = await onDrop(categoryName, feeds);

    if (result.success) {
      await store.fetchFeeds();
      window.showToast(t('modal.feed.feedReordered'), 'success');
    } else {
      window.showToast(t('common.errors.reorderingFeed') + ': ' + result.error, 'error');
    }
  } finally {
    isDragging.value = false;
  }
}

// Auto-expand collapsed categories when dragging over them
let autoExpandTimeout: ReturnType<typeof setTimeout> | null = null;
watch(dragOverCategory, (newCategory) => {
  if (autoExpandTimeout) {
    clearTimeout(autoExpandTimeout);
    autoExpandTimeout = null;
  }

  if (newCategory && isDragging.value) {
    const isClosed = !checkIsCategoryOpen(newCategory);

    if (isClosed) {
      autoExpandTimeout = setTimeout(() => {
        if (dragOverCategory.value === newCategory && isDragging.value) {
          toggleCategory(newCategory);
        }
      }, 300);
    }
  }
});

watch(draggingFeedId, (newValue, oldValue) => {
  if (oldValue !== null && newValue === null && !dropHandled) {
    setTimeout(() => {
      isDragging.value = false;
    }, 100);
  }
});

// Drawer type based on current filter
const drawerType = computed(() => {
  switch (store.currentFilter) {
    case 'all':
    case 'unread':
    case 'favorites':
    case 'readLater':
    case 'imageGallery':
      return 'feeds';
    default:
      return 'feeds';
  }
});

// Filter tree based on current filter
const filteredTree = computed(() => {
  if (drawerType.value !== 'feeds' || !tree.value) return { tree: {}, uncategorized: [] };

  const imageModeOnly = store.currentFilter === 'imageGallery';

  // Filter feeds in categories
  const filteredTree: Record<string, any> = {};

  const treeData = tree.value.tree || {};
  for (const [name, data] of Object.entries(treeData)) {
    const filteredFeeds = data._feeds.filter((f: Feed) => !imageModeOnly || f.is_image_mode);

    // Filter children recursively
    const filterChildren = (children: Record<string, any>): Record<string, any> => {
      const result: Record<string, any> = {};
      for (const [childName, childData] of Object.entries(children)) {
        const childFeeds = childData._feeds.filter((f: Feed) => !imageModeOnly || f.is_image_mode);
        const childChildren = filterChildren(childData._children);

        if (childFeeds.length > 0 || Object.keys(childChildren).length > 0) {
          result[childName] = {
            ...childData,
            _feeds: childFeeds,
            _children: childChildren,
          };
        }
      }
      return result;
    };

    const filteredChildren = filterChildren(data._children);

    if (filteredFeeds.length > 0 || Object.keys(filteredChildren).length > 0) {
      filteredTree[name] = {
        ...data,
        _feeds: filteredFeeds,
        _children: filteredChildren,
      };
    }
  }

  // Filter uncategorized feeds
  const uncategorizedFeeds = tree.value?.uncategorized || [];
  const filteredUncategorized = uncategorizedFeeds.filter(
    (f: Feed) => !imageModeOnly || f.is_image_mode
  );

  return {
    tree: filteredTree,
    uncategorized: filteredUncategorized,
  };
});

// Get drawer title
const drawerTitle = computed(() => {
  // Always show "Feeds" regardless of filter
  return t('sidebar.feedList.feeds');
});

function handleClose() {
  // Always allow closing, regardless of pinned state
  // Pinned state only affects positioning, not ability to close
  emit('collapse');
}

function handleTogglePin() {
  if (props.isPinned) {
    emit('unpin');
  } else {
    emit('pin');
  }
}

// Saved filters functions
async function applySavedFilter(filter: SavedFilter) {
  // Check if this filter is currently applied
  const isCurrentlyActive = isFilterActive(filter);

  if (isCurrentlyActive) {
    // Cancel the filter if clicking the currently applied one
    activeFilters.value = [];
    await fetchFilteredArticles([]);
  } else {
    // Apply the filter
    const conditions = parseConditions(filter.conditions);
    activeFilters.value = conditions;
    await fetchFilteredArticles(conditions);
  }
}

async function handleSaveFilter(name: string, conditions: FilterCondition[]) {
  try {
    const result = await createSavedFilter(name, conditions);
    if (result) {
      window.showToast(t('sidebar.savedFilters.filterSaved'), 'success');
      await fetchSavedFilters();
    }
  } catch (e) {
    // Show the error message from server
    const errorMessage = e instanceof Error ? e.message : t('sidebar.savedFilters.saveFailed');
    window.showToast(errorMessage, 'error');
  }
}

async function handleEditFilter(name: string, conditions: FilterCondition[]) {
  if (!editingFilter.value) return;

  const success = await updateSavedFilter(editingFilter.value.id, name, conditions);
  if (success) {
    window.showToast(t('sidebar.savedFilters.filterUpdated'), 'success');
    await fetchSavedFilters();
    closeEditModal();
  } else {
    window.showToast(t('sidebar.savedFilters.updateFailed'), 'error');
  }
}

async function handleDeleteFilter(filter: SavedFilter) {
  const confirmed = await window.showConfirm({
    title: t('sidebar.savedFilters.deleteConfirmTitle'),
    message: t('sidebar.savedFilters.deleteConfirmMessage', { name: filter.name }),
    isDanger: true,
  });

  if (confirmed) {
    const success = await deleteSavedFilter(filter.id);
    if (success) {
      window.showToast(t('sidebar.savedFilters.filterDeleted'), 'success');
      await fetchSavedFilters();
    } else {
      window.showToast(t('sidebar.savedFilters.deleteFailed'), 'error');
    }
  }
}

// Handle saved filter context menu using global context menu system
function onFilterContextMenu(event: MouseEvent, filter: SavedFilter) {
  event.preventDefault();
  event.stopPropagation();

  const menuItems = [
    {
      label: t('common.edit'),
      action: 'edit',
      icon: 'PhPencil',
    },
    {
      label: t('common.delete'),
      action: 'delete',
      icon: 'PhTrash',
      danger: true,
    },
  ];

  window.dispatchEvent(
    new CustomEvent('open-context-menu', {
      detail: {
        x: event.clientX,
        y: event.clientY,
        items: menuItems,
        data: filter,
        callback: handleFilterAction,
      },
    })
  );
}

async function handleFilterAction(action: string, filter: SavedFilter) {
  switch (action) {
    case 'edit':
      editingFilter.value = filter;
      showEditFilterModal.value = true;
      break;
    case 'delete':
      await handleDeleteFilter(filter);
      break;
  }
}

function openSaveModal() {
  showSaveFilterModal.value = true;
}

function closeSaveModal() {
  showSaveFilterModal.value = false;
}

function openEditModal(filter: SavedFilter) {
  editingFilter.value = filter;
  showEditFilterModal.value = true;
}

function closeEditModal() {
  showEditFilterModal.value = false;
  editingFilter.value = null;
}

// Drag and drop for saved filters
function handleFilterDragStart(filterId: number) {
  draggingFilterId.value = filterId;
}

function handleFilterDragEnd() {
  draggingFilterId.value = null;
}
</script>

<template>
  <Transition
    enter-active-class="transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
    enter-from-class="opacity-0 -translate-x-5"
    enter-to-class="opacity-100 translate-x-0"
    leave-active-class="transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]"
    leave-from-class="opacity-100 translate-x-0"
    leave-to-class="opacity-0 -translate-x-5"
  >
    <div
      v-if="isExpanded || isPinned"
      class="w-[280px] min-w-[280px] max-w-[80vw] md:w-[280px] md:min-w-[280px] flex flex-col h-full flex-shrink-0 relative border-r border-border feed-drawer-width z-20"
      :class="[isPinned ? 'bg-bg-primary' : 'bg-bg-secondary shadow-2xl']"
    >
      <!-- Drawer Header -->
      <div
        class="p-2 sm:p-4 border-b border-border flex items-center justify-between flex-shrink-0 bg-bg-primary"
      >
        <h3 class="m-0 text-base sm:text-lg font-semibold">{{ drawerTitle }}</h3>
        <div class="flex items-center gap-1 sm:gap-2">
          <!-- Pin/Unpin Button -->
          <button
            class="text-text-secondary hover:text-text-primary hover:bg-bg-tertiary p-1 sm:p-1.5 rounded transition-colors"
            :class="isPinned ? 'text-accent' : ''"
            :title="isPinned ? t('sidebar.feedList.unpin') : t('sidebar.feedList.pin')"
            @click="handleTogglePin"
          >
            <PhPushPinSlash v-if="isPinned" :size="18" class="sm:w-5 sm:h-5" />
            <PhPushPin v-else :size="18" class="sm:w-5 sm:h-5" />
          </button>
          <!-- Close Button -->
          <button
            class="text-text-secondary hover:text-text-primary hover:bg-bg-tertiary p-1 sm:p-1.5 rounded transition-colors"
            :title="t('common.close')"
            @click="handleClose"
          >
            <PhX :size="18" class="sm:w-5 sm:h-5" />
          </button>
        </div>
      </div>

      <!-- Drawer Content -->
      <div class="flex-1 overflow-hidden flex flex-col">
        <!-- Feeds Drawer (for all filters including imageGallery) -->
        <template v-if="drawerType === 'feeds'">
          <!-- Search Box -->
          <div class="border-b border-border">
            <div class="flex items-center">
              <div class="relative flex-1">
                <input
                  v-model="searchQuery"
                  type="text"
                  :placeholder="t('common.search.searchFeeds')"
                  class="w-full bg-bg-tertiary px-3 py-2 pl-8 text-sm focus:outline-none transition-colors"
                />
                <PhMagnifyingGlass
                  :size="14"
                  class="absolute left-2.5 top-1/2 -translate-y-1/2 text-text-secondary"
                />
                <button
                  v-if="searchQuery"
                  class="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-text-secondary hover:text-text-primary"
                  @click="searchQuery = ''"
                >
                  <PhX :size="12" />
                </button>
              </div>
              <!-- Edit Toggle Button -->
              <button
                class="text-text-secondary hover:text-accent p-1 sm:p-1.5 transition-colors flex-shrink-0"
                :class="isEditMode ? 'text-accent' : ''"
                :title="isEditMode ? t('common.done') : t('common.edit')"
                @click="toggleEditMode"
              >
                <PhPencil v-if="!isEditMode" :size="16" class="sm:w-5 sm:h-5" />
                <PhCheck v-else :size="16" class="sm:w-5 sm:h-5" />
              </button>
            </div>
          </div>

          <!-- Categories List -->
          <div
            class="categories-list sidebar-hover-scrollbar flex-1 overflow-y-auto overflow-x-hidden"
          >
            <SidebarCategory
              v-for="(data, name) in filteredTree.tree"
              :key="name"
              :name="name"
              :feeds="data._feeds"
              :children="data._children"
              :level="0"
              :is-open="checkIsCategoryOpen(name)"
              :is-active="store.currentCategory === name"
              :unread-count="categoryUnreadCounts[name] || 0"
              :current-feed-id="store.currentFeedId"
              :feed-unread-counts="feedUnreadCounts"
              :is-drag-over="dragOverCategory === name"
              :is-edit-mode="isEditMode"
              :drop-preview="dropPreview"
              :dragging-feed-id="draggingFeedId"
              :is-category-open="checkIsCategoryOpen"
              :compact-mode="compactMode"
              @toggle="() => toggleCategory(name)"
              @select-category="() => handleSelectCategory(name)"
              @select-feed="(feedId: number) => handleSelectFeed(feedId)"
              @category-context-menu="(e: MouseEvent) => onCategoryContextMenu(e, name)"
              @child-toggle="toggleCategory"
              @child-select-category="(category: string) => handleSelectCategory(category)"
              @child-context-menu="(e: MouseEvent, path: string) => onCategoryContextMenu(e, path)"
              @feed-context-menu="onFeedContextMenu"
              @dragstart="(feedId: number, e: Event) => handleDragStart(feedId, e)"
              @dragend="handleDragEnd"
              @feed-drag-over="(feedId: number | null, e: Event) => handleDragOver(name, feedId, e)"
              @category-drag-over="
                (categoryName: string, e: Event) => handleCategoryDragOver(categoryName, e)
              "
              @dragleave="(categoryName: string, e: Event) => handleDragLeave(categoryName, e)"
              @drop="() => handleDrop(name, data._feeds)"
            />

            <!-- Uncategorized -->
            <SidebarCategory
              v-if="filteredTree.uncategorized.length > 0 || isDragging"
              :name="t('sidebar.feedList.uncategorized')"
              :feeds="filteredTree.uncategorized"
              :is-open="
                checkIsCategoryOpen('uncategorized') ||
                (filteredTree.uncategorized.length === 0 && isDragging)
              "
              :is-active="store.currentCategory === ''"
              :is-uncategorized="true"
              :unread-count="categoryUnreadCounts['uncategorized'] || 0"
              :current-feed-id="store.currentFeedId"
              :feed-unread-counts="feedUnreadCounts"
              :is-drag-over="dragOverCategory === 'uncategorized'"
              :is-edit-mode="isEditMode"
              :drop-preview="dropPreview"
              :dragging-feed-id="draggingFeedId"
              :is-category-open="checkIsCategoryOpen"
              :compact-mode="compactMode"
              @toggle="toggleCategory('uncategorized')"
              @select-category="(path: string) => handleSelectCategory(path)"
              @select-feed="(feedId: number) => handleSelectFeed(feedId)"
              @category-context-menu="(e: MouseEvent) => onCategoryContextMenu(e, 'uncategorized')"
              @feed-context-menu="onFeedContextMenu"
              @dragstart="(feedId: number, e: Event) => handleDragStart(feedId, e)"
              @dragend="handleDragEnd"
              @feed-drag-over="
                (feedId: number | null, e: Event) => handleDragOver('uncategorized', feedId, e)
              "
              @category-drag-over="
                (categoryName: string, e: Event) => handleCategoryDragOver(categoryName, e)
              "
              @dragleave="(categoryName: string, e: Event) => handleDragLeave(categoryName, e)"
              @drop="() => handleDrop('uncategorized', filteredTree.uncategorized)"
            />
          </div>

          <!-- Saved Filters Section - positioned at bottom, only show when viewing All Articles -->
          <div
            v-if="
              store.currentFilter === 'all' && (hasActiveFilters || safeSavedFilters.length > 0)
            "
            class="flex-shrink-0 max-h-[50%] flex flex-col border-t border-border"
          >
            <!-- Saved Filters Header -->
            <div
              :class="[
                'flex-shrink-0 transition-colors duration-200 bg-bg-secondary cursor-default flex items-center justify-between',
                compactMode ? 'px-1.5 sm:px-2 py-1 sm:py-1.5' : 'px-3 py-1.5 sm:px-3 sm:py-2',
              ]"
            >
              <div class="flex items-center gap-1.5 sm:gap-2">
                <span class="font-semibold text-xs sm:text-sm text-text-secondary">
                  {{ t('sidebar.savedFilters.title') }}
                </span>
              </div>

              <!-- Save Current Filter Button -->
              <button
                :class="[
                  'bg-transparent border-0 cursor-pointer text-text-secondary rounded transition-all duration-200 flex items-center justify-center hover:not(:disabled):bg-bg-tertiary hover:not(:disabled):text-accent disabled:opacity-40 disabled:cursor-not-allowed',
                  'w-8 h-8',
                ]"
                :disabled="!hasActiveFilters"
                :title="
                  !hasActiveFilters
                    ? t('sidebar.savedFilters.conditionsRequired')
                    : t('sidebar.savedFilters.saveCurrentFilter')
                "
                @click="openSaveModal"
              >
                <PhFloppyDisk :size="18" />
              </button>
            </div>

            <!-- Saved Filters List -->
            <div
              :class="[
                'sidebar-hover-scrollbar flex-1 overflow-y-auto min-h-0',
                compactMode ? 'py-0.5 sm:py-1' : 'pt-1 pb-1 sm:pt-1.5 sm:pb-1.5',
              ]"
            >
              <SavedFilterItem
                v-for="filter in safeSavedFilters"
                :key="filter.id"
                :filter="filter"
                :is-active="isFilterActive(filter)"
                :is-dragging="draggingFilterId === filter.id"
                :is-edit-mode="isEditMode"
                :compact-mode="compactMode"
                @click="applySavedFilter(filter)"
                @contextmenu="onFilterContextMenu($event, filter)"
                @dragstart="handleFilterDragStart(filter.id)"
                @dragend="handleFilterDragEnd"
                @edit="openEditModal"
                @delete="handleDeleteFilter"
              />
            </div>
          </div>
        </template>
      </div>
    </div>
  </Transition>

  <!-- Save Filter Modal (Teleported to body) -->
  <Teleport to="body">
    <SavedFilterModal
      :show="showSaveFilterModal"
      :current-filters="currentFiltersValue"
      @close="closeSaveModal"
      @save="handleSaveFilter"
    />
  </Teleport>

  <!-- Edit Filter Modal (Teleported to body) -->
  <Teleport to="body">
    <SavedFilterModal
      :show="showEditFilterModal"
      :edit-filter="editingFilter"
      @close="closeEditModal"
      @save="handleEditFilter"
    />
  </Teleport>
</template>

<style scoped>
.sidebar-hover-scrollbar {
  scrollbar-gutter: stable;
  scrollbar-width: thin;
  scrollbar-color: transparent transparent;
}

.sidebar-hover-scrollbar:hover,
.sidebar-hover-scrollbar:focus-within {
  scrollbar-color: var(--border-color) transparent;
}

.sidebar-hover-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.sidebar-hover-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar-hover-scrollbar::-webkit-scrollbar-thumb {
  background: transparent;
  border-radius: 3px;
}

.sidebar-hover-scrollbar:hover::-webkit-scrollbar-thumb,
.sidebar-hover-scrollbar:focus-within::-webkit-scrollbar-thumb {
  background: var(--border-color);
}

.sidebar-hover-scrollbar::-webkit-scrollbar-thumb:hover {
  background: var(--text-secondary);
}

/* Responsive width for feed drawer on medium screens */
@media (max-width: 1400px) {
  .feed-drawer-width {
    width: 240px !important;
    min-width: 240px !important;
  }
}
</style>

<style>
.dark-mode .sidebar-hover-scrollbar:hover::-webkit-scrollbar-thumb,
.dark-mode .sidebar-hover-scrollbar:focus-within::-webkit-scrollbar-thumb {
  background: #444;
}

.dark-mode .sidebar-hover-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #666;
}
</style>
