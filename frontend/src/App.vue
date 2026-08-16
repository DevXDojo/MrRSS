<script setup lang="ts">
import { useAppStore } from './stores/app';
import { useI18n } from 'vue-i18n';
import Sidebar from './components/sidebar/Sidebar.vue';
import ArticleList from './components/article/ArticleList.vue';
import ArticleDetail from './components/article/ArticleDetail.vue';
import ImageGalleryView from './components/article/imageGallery/index.vue';
import AddFeedModal from './components/modals/feed/AddFeedModal.vue';
import EditFeedModal from './components/modals/feed/EditFeedModal.vue';
import SettingsModal from './components/modals/SettingsModal.vue';
import DiscoverFeedsModal from './components/modals/discovery/DiscoverFeedsModal.vue';
import UpdateAvailableDialog from './components/modals/update/UpdateAvailableDialog.vue';
import ContextMenu from './components/common/ContextMenu.vue';
import ConfirmDialog from './components/modals/common/ConfirmDialog.vue';
import InputDialog from './components/modals/common/InputDialog.vue';
import MultiSelectDialog from './components/modals/common/MultiSelectDialog.vue';
import Toast from './components/common/Toast.vue';
import { onMounted, onUnmounted, ref, computed, watchEffect } from 'vue';
import { useNotifications } from './composables/ui/useNotifications';
import { useKeyboardShortcuts } from './composables/ui/useKeyboardShortcuts';
import { useContextMenu } from './composables/ui/useContextMenu';
import { useResizablePanels } from './composables/ui/useResizablePanels';
import { useWindowState } from './composables/core/useWindowState';
import { useAppUpdates } from './composables/core/useAppUpdates';
import { useSettings } from './composables/core/useSettings';
import { resolveFontFamily } from './utils/fontDetector';
import type { Feed } from './types/models';

const store = useAppStore();
const { t } = useI18n();
const { settings } = useSettings();

const uiFontSize = computed(() => {
  const value = Number(settings.value.ui_font_size);
  return Number.isFinite(value) ? Math.min(20, Math.max(12, value)) : 16;
});

watchEffect(() => {
  const rootStyle = document.documentElement.style;
  rootStyle.setProperty('--ui-font-family', resolveFontFamily(settings.value.ui_font_family));
  rootStyle.setProperty('--ui-font-size', `${uiFontSize.value}px`);
  rootStyle.setProperty('--ui-font-scale', String(uiFontSize.value / 16));
});

onUnmounted(() => {
  const rootStyle = document.documentElement.style;
  rootStyle.removeProperty('--ui-font-family');
  rootStyle.removeProperty('--ui-font-size');
  rootStyle.removeProperty('--ui-font-scale');
});

const showAddFeed = ref(false);
const showEditFeed = ref(false);
const feedToEdit = ref<Feed | null>(null);
const showSettings = ref(false);
const showDiscoverBlogs = ref(false);
const feedToDiscover = ref<Feed | null>(null);
const isSidebarOpen = ref(true);

// Check if we're in image gallery mode
const isImageGalleryMode = computed(() => store.currentFilter === 'imageGallery');

// Check if we're in card mode
const isCardMode = ref(false);

// Use composables
const {
  confirmDialog,
  inputDialog,
  multiSelectDialog,
  toasts,
  removeToast,
  installGlobalHandlers,
} = useNotifications();

const { contextMenu, openContextMenu, handleContextMenuAction } = useContextMenu();

const {
  sidebarWidth,
  articleListWidth,
  startResizeArticleList,
  setArticleListWidth,
  setCompactMode,
} = useResizablePanels();

// Use app updates composable
const {
  updateInfo,
  checkForUpdates,
  downloadAndInstallUpdate,
  downloadingUpdate,
  installingUpdate,
  downloadProgress,
} = useAppUpdates();

// Update dialog state
const showUpdateDialog = ref(false);

// Initialize window state management
const windowState = useWindowState();
windowState.init();

// Initialize keyboard shortcuts
const { shortcuts } = useKeyboardShortcuts({
  onOpenSettings: () => {
    showSettings.value = true;
  },
  onAddFeed: () => {
    showAddFeed.value = true;
  },
  onMarkAllRead: async () => {
    // Reuse the article-list action so the shortcut has the same scope,
    // confirmation dialog, and filtered-article handling as the toolbar button.
    window.dispatchEvent(new CustomEvent('mark-all-as-read'));
  },
});

onMounted(async () => {
  // Install global notification handlers
  installGlobalHandlers();

  // Initialize theme system immediately (lightweight)
  store.initTheme();

  // Load remaining settings (theme and other settings are already loaded in main.ts)
  let updateInterval = 10;
  let lastGlobalRefresh = '';
  let updateCheckEnabled = true;

  try {
    const res = await fetch('/api/settings');
    const data = await res.json();

    // Set initial article list width based on layout mode setting
    const layoutMode = data.layout_mode || 'normal';
    const isCompactModeLayout = layoutMode === 'compact';
    isCardMode.value = layoutMode === 'card';
    // First set the compact mode, then set the width (order matters)
    setCompactMode(isCompactModeLayout);
    setArticleListWidth(isCompactModeLayout ? 500 : 350);

    // Notify all components that settings have been loaded
    window.dispatchEvent(new CustomEvent('settings-loaded'));

    // Apply saved theme preference (already applied in main.ts, but ensure it's set)
    if (data.theme) {
      store.setTheme(data.theme);
    }

    // Apply other settings
    if (data.update_interval) {
      updateInterval = parseInt(data.update_interval);
      store.startAutoRefresh(updateInterval);
    }

    if (data.last_global_refresh) {
      lastGlobalRefresh = data.last_global_refresh;
    }

    updateCheckEnabled = data.update_check_enabled !== 'false';

    // Load saved shortcuts
    if (data.shortcuts) {
      try {
        const parsed = JSON.parse(data.shortcuts);
        shortcuts.value = { ...shortcuts.value, ...parsed };
      } catch (e) {
        console.error('Error parsing shortcuts:', e);
      }
    }
  } catch (e) {
    console.error('Error loading initial settings:', e);
  }

  if (updateCheckEnabled) {
    // Check for updates on startup (silent mode - don't show toast if up to date)
    setTimeout(async () => {
      try {
        await checkForUpdates(true);

        // If update is available, show dialog for user to manually confirm
        if (updateInfo.value && updateInfo.value.has_update) {
          showUpdateDialog.value = true;
        }
      } catch (e) {
        console.error('Error checking for updates:', e);
      }
    }, 3000); // Check 3 seconds after startup
  }

  // Defer heavy operations to allow UI to render first
  setTimeout(() => {
    // Load feeds and articles in background
    store.fetchFeeds();
    store.fetchArticles();

    // Check if backend is already refreshing (e.g., from auto-refresh on startup)
    // and start polling progress if so
    setTimeout(async () => {
      try {
        const progressRes = await fetch('/api/progress');
        const progressData = await progressRes.json();

        if (progressData.is_running) {
          // Backend is already refreshing, start polling
          store.refreshProgress = {
            ...store.refreshProgress,
            isRunning: true,
            pool_task_count: progressData.pool_task_count,
            article_click_count: progressData.article_click_count,
            queue_task_count: progressData.queue_task_count,
          };
          store.pollProgress();
          return; // Don't trigger another refresh
        }
      } catch (e) {
        console.error('Error checking initial refresh progress:', e);
      }

      // Only trigger feed refresh if enough time has passed since last update
      // and backend is not already refreshing

      // Re-fetch the latest last_global_refresh from backend to ensure we have
      // the most recent value (in case a previous update just completed)
      let latestLastGlobalRefresh = lastGlobalRefresh;
      try {
        const settingsRes = await fetch('/api/settings');
        const settingsData = await settingsRes.json();
        if (settingsData.last_global_refresh) {
          latestLastGlobalRefresh = settingsData.last_global_refresh;
        }
      } catch (e) {
        console.error('Error fetching latest last_global_refresh:', e);
      }

      const shouldRefresh = shouldTriggerRefresh(latestLastGlobalRefresh, updateInterval);
      if (shouldRefresh) {
        store.refreshFeeds();
      }
    }, 500);
  }, 100);
});

// Listen for events from Sidebar (moved outside onMounted to ensure proper capture)
window.addEventListener('show-add-feed', () => {
  showAddFeed.value = true;
});
window.addEventListener('show-edit-feed', (e) => {
  const customEvent = e as CustomEvent<any>;
  feedToEdit.value = customEvent.detail;
  showEditFeed.value = true;
});
window.addEventListener('show-settings', () => {
  showSettings.value = true;
});
window.addEventListener('show-discover-blogs', (e) => {
  const customEvent = e as CustomEvent<any>;
  feedToDiscover.value = customEvent.detail;
  showDiscoverBlogs.value = true;
});

// Listen for compact mode changes to update article list width
window.addEventListener('layout-mode-changed', (e) => {
  const customEvent = e as CustomEvent<{ mode: string }>;
  const mode = customEvent.detail.mode;
  const isCompactModeLayout = mode === 'compact';
  isCardMode.value = mode === 'card';
  setCompactMode(isCompactModeLayout);
  if (!isCardMode.value) {
    setArticleListWidth(isCompactModeLayout ? 600 : 400);
  }
});

// Global Context Menu Event Listener
window.addEventListener('open-context-menu', (e) => {
  openContextMenu(e as CustomEvent<any>);
});

// Check if we should trigger refresh based on last update time and interval
function shouldTriggerRefresh(lastUpdate: string, intervalMinutes: number): boolean {
  if (!lastUpdate) {
    return true; // Never updated, should refresh
  }

  try {
    const lastUpdateTime = new Date(lastUpdate).getTime();
    const now = Date.now();
    const intervalMs = intervalMinutes * 60 * 1000;

    // Refresh if more than interval time has passed since last update
    return now - lastUpdateTime >= intervalMs;
  } catch {
    return true; // Invalid date, should refresh
  }
}

function toggleSidebar(): void {
  isSidebarOpen.value = !isSidebarOpen.value;
}

async function onFeedAdded(feedId?: number): Promise<void> {
  await store.fetchFeeds();
  if (feedId) {
    await store.setFilter('all');
    store.setFeed(feedId);
  }
  // Start polling for progress as the backend is now fetching articles for the new feed
  if (!feedId) {
    store.pollProgress();
  }
}

function onFeedUpdated(): void {
  store.fetchFeeds();
  // Refresh articles to immediately apply hide_from_timeline changes
  store.fetchArticles();
}
</script>

<template>
  <div
    class="app-container flex h-screen w-full bg-bg-primary text-text-primary overflow-hidden"
    :style="{
      '--sidebar-width': sidebarWidth + 'px',
      '--article-list-width': articleListWidth + 'px',
    }"
  >
    <Sidebar :is-open="isSidebarOpen" @toggle="toggleSidebar" />

    <!-- Show ImageGalleryView when in image gallery mode -->
    <template v-if="isImageGalleryMode">
      <ImageGalleryView :is-sidebar-open="isSidebarOpen" @toggle-sidebar="toggleSidebar" />
    </template>

    <!-- Show ArticleList and ArticleDetail when not in image gallery mode -->
    <template v-else>
      <ArticleList :is-sidebar-open="isSidebarOpen" @toggle-sidebar="toggleSidebar" />

      <!-- Hide resizer and ArticleDetail when in card mode -->
      <template v-if="!isCardMode">
        <div class="resizer hidden md:block" @mousedown="startResizeArticleList"></div>

        <ArticleDetail />
      </template>
    </template>

    <AddFeedModal v-if="showAddFeed" @close="showAddFeed = false" @added="onFeedAdded" />
    <EditFeedModal
      v-if="showEditFeed && feedToEdit"
      :feed="feedToEdit"
      @close="showEditFeed = false"
      @updated="onFeedUpdated"
    />
    <SettingsModal v-if="showSettings" @close="showSettings = false" />
    <DiscoverFeedsModal
      v-if="showDiscoverBlogs && feedToDiscover"
      :feed="feedToDiscover"
      :show="showDiscoverBlogs"
      @close="showDiscoverBlogs = false"
    />

    <UpdateAvailableDialog
      v-if="showUpdateDialog && updateInfo"
      :update-info="updateInfo"
      :downloading-update="downloadingUpdate"
      :installing-update="installingUpdate"
      :download-progress="downloadProgress"
      @close="showUpdateDialog = false"
      @update="downloadAndInstallUpdate"
    />

    <ContextMenu
      v-if="contextMenu.show"
      :x="contextMenu.x"
      :y="contextMenu.y"
      :items="contextMenu.items"
      @close="contextMenu.show = false"
      @action="handleContextMenuAction"
    />

    <!-- Global Notification System -->
    <ConfirmDialog
      v-if="confirmDialog"
      :title="confirmDialog.title"
      :message="confirmDialog.message"
      :confirm-text="confirmDialog.confirmText"
      :cancel-text="confirmDialog.cancelText"
      :is-danger="confirmDialog.isDanger"
      @confirm="confirmDialog.onConfirm"
      @cancel="confirmDialog.onCancel"
      @close="confirmDialog = null"
    />

    <InputDialog
      v-if="inputDialog"
      :title="inputDialog.title"
      :message="inputDialog.message"
      :placeholder="inputDialog.placeholder"
      :default-value="inputDialog.defaultValue"
      :confirm-text="inputDialog.confirmText"
      :cancel-text="inputDialog.cancelText"
      :suggestions="inputDialog.suggestions"
      @confirm="inputDialog.onConfirm"
      @cancel="inputDialog.onCancel"
      @close="inputDialog = null"
    />

    <MultiSelectDialog
      v-if="multiSelectDialog"
      :title="multiSelectDialog.title"
      :message="multiSelectDialog.message"
      :options="multiSelectDialog.options"
      :confirm-text="multiSelectDialog.confirmText"
      :cancel-text="multiSelectDialog.cancelText"
      @confirm="multiSelectDialog.onConfirm"
      @cancel="multiSelectDialog.onCancel"
      @close="multiSelectDialog = null"
    />

    <div class="toast-container">
      <Toast
        v-for="toast in toasts"
        :key="toast.id"
        :message="toast.message"
        :type="toast.type"
        :duration="toast.duration"
        @close="removeToast(toast.id)"
      />
    </div>
  </div>
</template>

<style>
.toast-container {
  position: fixed;
  top: 10px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 60;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  pointer-events: none;
}

.toast-container > * {
  top: 42px; /* Account for MacOS top padding */
}

.toast-container > * {
  pointer-events: auto;
}
@media (min-width: 640px) {
  .toast-container {
    top: 20px;
    gap: 10px;
  }
  .app-container.macos-padding .toast-container {
    top: 52px; /* Account for MacOS top padding on larger screens */
  }
}
.resizer {
  width: 4px;
  cursor: col-resize;
  background-color: transparent;
  flex-shrink: 0;
  transition: background-color 0.2s;
  z-index: 10;
  margin-left: -2px;
  margin-right: -2px;
}
.resizer:hover,
.resizer:active {
  background-color: var(--color-accent, #3b82f6);
}
/* Global styles if needed */
</style>
