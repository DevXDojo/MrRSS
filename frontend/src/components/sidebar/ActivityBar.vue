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
  PhTextOutdent,
  PhSidebar,
} from '@phosphor-icons/vue';
import { ref, onMounted } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { useArticleFilter } from '@/composables/article/useArticleFilter';
import LogoSvg from '../../../public/assets/logo.svg';

const store = useAppStore();
const { t } = useI18n();
const { clearAllFilters } = useArticleFilter();

interface Props {
  isCollapsed?: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'select-filter': [filterType: string];
  'add-feed': [];
  settings: [];
  'toggle-feed-drawer': [];
  ready: [{ expanded: boolean; pinned: boolean }];
  'toggle-activity-bar': [];
}>();

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
  // Clear any active saved filters when clicking main filter buttons
  clearAllFilters();
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
  <Transition name="activity-bar-slide">
    <div
      v-if="!props.isCollapsed"
      class="smart-activity-bar flex flex-col items-center py-3 bg-bg-tertiary border-r border-border h-full select-none shrink-0 relative z-30"
    >
      <!-- Logo -->
      <div class="mb-6">
        <img :src="LogoSvg" alt="MrRSS" class="w-6 h-6" />
      </div>

      <!-- Divider -->
      <div class="w-8 h-px bg-border mb-3"></div>

      <!-- Navigation Items -->
      <div
        class="flex-1 flex flex-col items-center gap-1 w-full overflow-y-auto overflow-x-hidden nav-items-container"
      >
        <TransitionGroup name="nav-item">
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
              :is="
                store.currentFilter === item.filterType ? item.activeIcon || item.icon : item.icon
              "
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
        </TransitionGroup>
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
          <PhSidebar :size="24" :weight="isFeedListExpanded ? 'fill' : 'regular'" />
        </button>

        <button
          class="relative flex items-center justify-center text-text-secondary flex-shrink-0 transition-all hover:text-accent"
          style="width: 44px; height: 44px"
          :title="t('setting.tab.settings')"
          @click="emit('settings')"
        >
          <PhGear :size="24" weight="regular" class="transition-all" />
        </button>

        <!-- Divider -->
        <div class="w-8 h-px bg-border my-2"></div>

        <!-- Collapse Button (at the bottom) -->
        <button
          class="relative flex items-center justify-center text-text-secondary flex-shrink-0 transition-all hover:text-accent"
          style="width: 44px; height: 44px"
          :title="t('sidebar.activity.collapseActivityBar')"
          @click="emit('toggle-activity-bar')"
        >
          <PhTextOutdent :size="24" weight="regular" class="transition-all" />
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
/* Activity bar slide transition */
.activity-bar-slide-enter-active {
  transition:
    transform 0.25s cubic-bezier(0.4, 0, 0.2, 1),
    opacity 0.2s ease;
  will-change: transform, opacity;
}

.activity-bar-slide-leave-active {
  transition:
    transform 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    opacity 0.18s ease;
  will-change: transform, opacity;
}

.activity-bar-slide-enter-from {
  opacity: 0;
  transform: translateX(-12px);
}

.activity-bar-slide-leave-to {
  opacity: 0;
  transform: translateX(-12px);
}

.activity-bar-slide-enter-to,
.activity-bar-slide-leave-from {
  opacity: 1;
  transform: translateX(0);
}

.smart-activity-bar {
  width: 56px;
  min-width: 56px;
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 15;
  /* Prevent layout shift during animations */
  backface-visibility: hidden;
  -webkit-font-smoothing: antialiased;
}

/* Navigation items smooth transitions */
.nav-items-container {
  /* Smooth height transition when items are added/removed */
  transition: height 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  /* Reserve the scrollbar space on both sides so centered icons stay centered. */
  scrollbar-gutter: stable both-edges;
  scrollbar-width: thin;
  scrollbar-color: transparent transparent;
}

.nav-items-container:hover,
.nav-items-container:focus-within {
  scrollbar-color: var(--border-color) transparent;
}

.nav-items-container::-webkit-scrollbar {
  width: 4px;
}

.nav-items-container::-webkit-scrollbar-track {
  background: transparent;
}

.nav-items-container::-webkit-scrollbar-thumb {
  background: transparent;
  border-radius: 2px;
}

.nav-items-container:hover::-webkit-scrollbar-thumb,
.nav-items-container:focus-within::-webkit-scrollbar-thumb {
  background: var(--border-color);
}

.nav-items-container::-webkit-scrollbar-thumb:hover {
  background: var(--text-secondary);
}

/* Nav item enter/leave transitions */
.nav-item-enter-active,
.nav-item-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  will-change: opacity, transform;
}

.nav-item-enter-from {
  opacity: 0;
  transform: scale(0.9) translateY(-10px);
}

.nav-item-leave-to {
  opacity: 0;
  transform: scale(0.9) translateY(10px);
}

.nav-item-move {
  transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  will-change: transform;
}

/* Ensure smooth transitions for icon scale changes */
.smart-activity-bar button .ph,
.smart-activity-bar button svg {
  transition:
    transform 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    color 0.2s ease;
  will-change: transform;
}

/* Improve button hover transition */
.smart-activity-bar button {
  transition:
    color 0.2s ease,
    background-color 0.2s ease;
  will-change: color, background-color;
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
