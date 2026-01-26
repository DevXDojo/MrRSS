<script setup lang="ts">
import {
  PhListDashes,
  PhSquaresFour,
  PhTray,
  PhStar,
  PhClockCountdown,
  PhImages,
  PhPlus,
  PhGear,
  PhTextIndent,
  PhTextOutdent,
} from '@phosphor-icons/vue';
import { ref, onMounted } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import LogoSvg from '../../../assets/logo.svg';

const store = useAppStore();
const { t } = useI18n();

interface NavItem {
  id: string;
  icon: any;
  label: string;
  activeIcon?: any;
  filterType: 'all' | 'unread' | 'favorites' | 'readLater' | 'imageGallery';
}

const navItems: NavItem[] = [
  {
    id: 'all',
    icon: PhListDashes,
    activeIcon: PhSquaresFour,
    label: t('sidebar.activity.allArticles'),
    filterType: 'all',
  },
  {
    id: 'unread',
    icon: PhTray,
    label: t('sidebar.feedList.unread'),
    filterType: 'unread',
  },
  {
    id: 'favorites',
    icon: PhStar,
    label: t('sidebar.activity.favorites'),
    filterType: 'favorites',
  },
  {
    id: 'readLater',
    icon: PhClockCountdown,
    label: t('sidebar.activity.readLater'),
    filterType: 'readLater',
  },
  {
    id: 'imageGallery',
    icon: PhImages,
    label: t('sidebar.activity.imageGallery'),
    filterType: 'imageGallery',
  },
];

// Check if image gallery feature is enabled
const imageGalleryEnabled = ref(false);

async function loadImageGallerySetting() {
  try {
    const res = await fetch('/api/settings');
    if (res.ok) {
      const data = await res.json();
      imageGalleryEnabled.value = data.image_gallery_enabled === 'true';
    }
  } catch (e) {
    console.error('Failed to load settings:', e);
  }
}

const emit = defineEmits<{
  'select-filter': [filterType: string];
  'add-feed': [];
  settings: [];
  'toggle-feed-drawer': [];
  ready: [{ expanded: boolean; pinned: boolean }];
}>();

// Feed drawer state - use localStorage just like category open/pinned state
const savedPinnedState = localStorage.getItem('FeedListPinned');
const savedExpandedState = localStorage.getItem('FeedListExpanded');

const isFeedListPinned = ref(savedPinnedState === 'true' || savedPinnedState === null); // Default: pinned
const isFeedListExpanded = ref(savedExpandedState === 'true' || savedExpandedState === null); // Default: expanded

// Save state to localStorage
function saveDrawerState() {
  localStorage.setItem('FeedListPinned', String(isFeedListPinned.value));
  localStorage.setItem('FeedListExpanded', String(isFeedListExpanded.value));
}

// Load state from localStorage (called on mount)
function loadDrawerState() {
  const pinned = localStorage.getItem('FeedListPinned');
  const expanded = localStorage.getItem('FeedListExpanded');
  isFeedListPinned.value = pinned === 'true' || pinned === null;
  isFeedListExpanded.value = expanded === 'true' || expanded === null;
}

onMounted(async () => {
  await loadImageGallerySetting();
  loadDrawerState();

  // Notify parent that initialization is complete
  emit('ready', {
    expanded: isFeedListExpanded.value,
    pinned: isFeedListPinned.value,
  });

  // Listen for settings changes
  window.addEventListener('image-gallery-setting-changed', (e: Event) => {
    const customEvent = e as CustomEvent;
    imageGalleryEnabled.value = customEvent.detail.enabled;
  });
});

function handleNavClick(item: NavItem) {
  store.setFilter(item.filterType);
  emit('select-filter', item.filterType);

  // Don't auto-expand feed panel when clicking nav items
  // Only expand when clicking the Feed button
}

function toggleFeedList() {
  // Only toggle expand/collapse state
  // Pinned state should remain unchanged and only be controlled via the pin button in FeedList
  isFeedListExpanded.value = !isFeedListExpanded.value;
  saveDrawerState();
  emit('toggle-feed-drawer');
}

function pinFeedList() {
  isFeedListPinned.value = true;
  isFeedListExpanded.value = true;
  saveDrawerState();
  emit('toggle-feed-drawer');
}

function unpinFeedList() {
  isFeedListPinned.value = false;
  // Keep expanded when unpinning - don't collapse
  saveDrawerState();
  emit('toggle-feed-drawer');
}

// Listen for drawer state changes from parent
function handleFeedListStateChange(expanded: boolean, pinned?: boolean) {
  isFeedListExpanded.value = expanded;
  // Only update pinned if it's provided (not undefined)
  if (pinned !== undefined) {
    isFeedListPinned.value = pinned;
  }
  saveDrawerState();
}

// Expose functions and state to parent
defineExpose({
  toggleFeedList,
  pinFeedList,
  unpinFeedList,
  handleFeedListStateChange,
  loadDrawerState,
  // Expose refs as computed getters
  get isFeedListExpanded() {
    return isFeedListExpanded.value;
  },
  get isFeedListPinned() {
    return isFeedListPinned.value;
  },
});
</script>

<template>
  <div
    class="smart-activity-bar flex flex-col items-center py-3 bg-bg-tertiary border-r border-border h-full select-none shrink-0 relative"
  >
    <!-- Logo -->
    <div class="mt-3 mb-6">
      <img :src="LogoSvg" alt="MrRSS" class="w-6 h-6" />
    </div>

    <!-- Divider -->
    <div class="w-8 h-px bg-border mb-3"></div>

    <!-- Navigation Items -->
    <div class="flex-1 flex flex-col items-center gap-1 w-full overflow-y-auto overflow-x-hidden">
      <button
        v-for="item in navItems"
        v-show="item.id !== 'imageGallery' || imageGalleryEnabled"
        :key="item.id"
        :class="[
          'relative flex items-center justify-center text-text-secondary flex-shrink-0 transition-all hover:text-accent',
          store.currentFilter === item.filterType ? 'text-accent' : '',
        ]"
        style="width: 44px; height: 44px"
        :title="item.label"
        @click="handleNavClick(item)"
      >
        <!-- Icon -->
        <component
          :is="store.currentFilter === item.filterType ? item.activeIcon || item.icon : item.icon"
          :size="24"
          :weight="store.currentFilter === item.filterType ? 'fill' : 'regular'"
          :class="[
            store.currentFilter === item.filterType ? 'text-accent scale-105' : '',
            'transition-all',
          ]"
        />

        <!-- Unread Badge (only for 'all' button) -->
        <span
          v-if="item.id === 'all' && store.unreadCounts?.total > 0"
          class="absolute bottom-0.5 right-0.5 min-w-[14px] h-[14px] px-0.5 text-[9px] font-medium flex items-center justify-center rounded-full text-white"
          style="background-color: #999999"
        >
          {{ store.unreadCounts?.total > 99 ? '99+' : store.unreadCounts?.total }}
        </span>
      </button>
    </div>

    <!-- Bottom Actions -->
    <div class="flex flex-col items-center gap-1 mt-auto w-full">
      <button
        class="relative flex items-center justify-center text-text-secondary flex-shrink-0 transition-all hover:text-accent"
        style="width: 44px; height: 44px"
        :title="t('sidebar.activity.addFeed')"
        @click="emit('add-feed')"
      >
        <PhPlus :size="24" weight="regular" class="transition-all" />
      </button>

      <!-- Feed List Button -->
      <button
        class="relative flex items-center justify-center text-text-secondary flex-shrink-0 transition-all hover:text-accent"
        style="width: 44px; height: 44px"
        :title="
          isFeedListExpanded
            ? t('sidebar.activity.collapseFeedList')
            : t('sidebar.activity.expandFeedList')
        "
        @click="toggleFeedList"
      >
        <PhTextOutdent v-if="isFeedListExpanded" :size="24" />
        <PhTextIndent v-else :size="24" />
      </button>

      <button
        class="relative flex items-center justify-center text-text-secondary flex-shrink-0 transition-all hover:text-accent"
        style="width: 44px; height: 44px"
        :title="t('setting.tab.settings')"
        @click="emit('settings')"
      >
        <PhGear :size="24" weight="regular" class="transition-all" />
      </button>
    </div>
  </div>
</template>

<style scoped>
@reference "../../style.css";

.smart-activity-bar {
  width: 56px;
  min-width: 56px;
  position: relative;
  z-index: 20;
}

.smart-activity-bar {
  --wails-draggable: drag;
}

.smart-activity-bar button,
.smart-activity-bar .w-8.h-px,
.smart-activity-bar img {
  --wails-draggable: no-drag;
}

/* Smaller screens (laptops, tablets) */
@media (max-width: 1400px) {
  .smart-activity-bar {
    width: 48px;
    min-width: 48px;
  }

  button[style*='width: 44px'] {
    width: 40px !important;
    height: 40px !important;
  }
}

/* Mobile devices */
@media (max-width: 767px) {
  .smart-activity-bar {
    width: 44px;
    min-width: 44px;
  }

  button[style*='width: 44px'] {
    width: 36px !important;
    height: 36px !important;
  }
}
</style>

<style>
/* Dark mode for unread badge - keep accent color */
</style>
