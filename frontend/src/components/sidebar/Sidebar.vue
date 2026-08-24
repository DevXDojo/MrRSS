<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue';
import { PhCaretRight } from '@phosphor-icons/vue';
import { useI18n } from 'vue-i18n';
import ActivityBar from './ActivityBar.vue';
import FeedList from './FeedList.vue';

interface Props {
  isOpen?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  toggle: [];
}>();

const { t } = useI18n();

// Feed drawer state
const isFeedListExpanded = ref(false);
const isFeedListPinned = ref(false);
const activityBarRef = ref<InstanceType<typeof ActivityBar> | null>(null);

// Activity bar collapse state - use localStorage for persistence
const savedActivityBarCollapsed = localStorage.getItem('ActivityBarCollapsed');
const isActivityBarCollapsed = ref(savedActivityBarCollapsed === 'true');

// Save activity bar state to localStorage
function saveActivityBarState() {
  localStorage.setItem('ActivityBarCollapsed', String(isActivityBarCollapsed.value));
}

// Handle ready event from ActivityBar
function handleActivityBarReady(state: { expanded: boolean; pinned: boolean }) {
  isFeedListExpanded.value = state.expanded;
  isFeedListPinned.value = state.pinned;
}

// Initialize state from ActivityBar after mount (fallback)
onMounted(async () => {
  await nextTick();

  // Fallback: if ready event doesn't fire, try reading state after delay
  setTimeout(() => {
    if (activityBarRef.value) {
      const expanded = activityBarRef.value.isFeedListExpanded;
      const pinned = activityBarRef.value.isFeedListPinned;

      // Only update if not already set by ready event
      if (isFeedListExpanded.value === false && expanded === true) {
        isFeedListExpanded.value = expanded;
        isFeedListPinned.value = pinned;
      }
    }
  }, 300);
});

function handleFeedListExpand() {
  isFeedListExpanded.value = true;
  updateActivityBarState();
}

function handleFeedListCollapse() {
  isFeedListExpanded.value = false;
  updateActivityBarState();
}

function handlePinFeedList() {
  isFeedListPinned.value = true;
  isFeedListExpanded.value = true;
  updateActivityBarState();
}

function handleUnpinFeedList() {
  isFeedListPinned.value = false;
  // Keep expanded when unpinning - don't collapse
  updateActivityBarState();
}

function handleToggleFeedList() {
  // Only toggle expand/collapse state
  // Pinned state should remain unchanged and only be controlled via the pin button in FeedList
  isFeedListExpanded.value = !isFeedListExpanded.value;
  updateActivityBarState();
}

// Update activity bar state when drawer state changes
function updateActivityBarState() {
  if (activityBarRef.value) {
    activityBarRef.value.handleFeedListStateChange(
      isFeedListExpanded.value,
      isFeedListPinned.value
    );
  }
}

const emitShowAddFeed = () => window.dispatchEvent(new CustomEvent('show-add-feed'));
const emitShowSettings = () => window.dispatchEvent(new CustomEvent('show-settings'));

function toggleActivityBar() {
  isActivityBarCollapsed.value = !isActivityBarCollapsed.value;
  saveActivityBarState();
}
</script>

<template>
  <div
    class="compact-sidebar-wrapper flex h-full relative"
    :class="{ 'width-collapsed': isActivityBarCollapsed }"
  >
    <!-- Shared container for ActivityBar and Edge Toggle -->
    <div class="sidebar-toggle-container">
      <!-- Edge Toggle Button (visible when ActivityBar is collapsed) -->
      <Transition name="edge-toggle-fade">
        <button
          v-if="isActivityBarCollapsed"
          class="edge-toggle-button flex items-center justify-center text-text-secondary hover:text-accent hover:bg-bg-secondary transition-all"
          :title="t('sidebar.activity.expandActivityBar')"
          @click="toggleActivityBar"
        >
          <PhCaretRight :size="20" weight="regular" />
        </button>
      </Transition>

      <!-- Smart Activity Bar (Left) -->
      <ActivityBar
        ref="activityBarRef"
        :is-collapsed="isActivityBarCollapsed"
        @add-feed="emitShowAddFeed"
        @settings="emitShowSettings"
        @toggle-feed-drawer="handleToggleFeedList"
        @toggle-activity-bar="toggleActivityBar"
        @ready="handleActivityBarReady"
      />
    </div>

    <!-- Feed Drawer -->
    <Transition name="drawer-position">
      <div
        v-if="isFeedListExpanded"
        class="feed-drawer-wrapper"
        :class="[
          { pinned: isFeedListPinned },
          { 'activity-bar-collapsed': isActivityBarCollapsed },
        ]"
      >
        <FeedList
          :is-expanded="isFeedListExpanded"
          :is-pinned="isFeedListPinned"
          @expand="handleFeedListExpand"
          @collapse="handleFeedListCollapse"
          @pin="handlePinFeedList"
          @unpin="handleUnpinFeedList"
        />
      </div>
    </Transition>

    <!-- Overlay for mobile -->
    <Transition name="overlay-fade">
      <div
        v-if="isOpen && isFeedListExpanded"
        class="fixed inset-0 bg-black/50 z-20 md:hidden"
        @click="emit('toggle')"
      ></div>
    </Transition>
  </div>
</template>

<style scoped>
.compact-sidebar-wrapper {
  position: relative;
  z-index: 20;
  display: flex;
  align-items: stretch;
  /* Smooth width transition between collapsed/expanded states */
  transition: width 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  will-change: width;
}

/* Container for both ActivityBar and Edge Toggle - uses absolute positioning */
.sidebar-toggle-container {
  position: relative;
  width: 56px;
  min-width: 56px;
  height: 100%;
  flex-shrink: 0;
  /* Width transition happens after button animations */
  transition:
    width 0.25s cubic-bezier(0.4, 0, 0.2, 1) 0.15s,
    min-width 0.25s cubic-bezier(0.4, 0, 0.2, 1) 0.15s;
  will-change: width, min-width;
}

/* When collapsed, container shrinks to edge toggle button width */
.compact-sidebar-wrapper.width-collapsed .sidebar-toggle-container {
  width: 16px;
  min-width: 16px;
}

/* Edge toggle button - absolutely positioned in shared space */
.edge-toggle-button {
  position: absolute;
  left: 0;
  top: 0;
  width: 16px;
  height: 100%;
  border-right: 1px solid var(--color-border);
  background-color: var(--color-bg-secondary);
  cursor: pointer;
  z-index: 16;
  transition: background-color 0.2s;
}

.edge-toggle-button:hover {
  background-color: var(--color-bg-tertiary);
}

/* Edge toggle fade transition - faster than container width change */
.edge-toggle-fade-enter-active,
.edge-toggle-fade-leave-active {
  transition: opacity 0.15s cubic-bezier(0.4, 0, 0.2, 1);
  will-change: opacity;
}

.edge-toggle-fade-enter-from,
.edge-toggle-fade-leave-to {
  opacity: 0;
}

.edge-toggle-fade-enter-to,
.edge-toggle-fade-leave-from {
  opacity: 1;
}

/* Smaller screens (laptops, tablets) */
@media (max-width: 1400px) {
  .sidebar-toggle-container {
    width: 48px;
    min-width: 48px;
  }

  .compact-sidebar-wrapper.width-collapsed .sidebar-toggle-container {
    width: 16px;
    min-width: 16px;
  }
}

/* Mobile devices */
@media (max-width: 767px) {
  .sidebar-toggle-container {
    width: 44px;
    min-width: 44px;
  }

  .compact-sidebar-wrapper.width-collapsed .sidebar-toggle-container {
    width: 16px;
    min-width: 16px;
  }
}

.feed-drawer-wrapper {
  position: relative;
  height: 100%;
  flex-shrink: 0;
}

.feed-drawer-wrapper:not(.pinned) {
  position: absolute;
  left: 56px;
  top: 0;
  bottom: 0;
  z-index: 20;
}

/* When activity bar is collapsed, feed drawer should start from edge toggle button */
.feed-drawer-wrapper:not(.pinned).activity-bar-collapsed {
  left: 16px;
}

/* Smaller screens (laptops, tablets) */
@media (max-width: 1400px) {
  .feed-drawer-wrapper:not(.pinned) {
    left: 48px;
  }

  .feed-drawer-wrapper:not(.pinned).activity-bar-collapsed {
    left: 16px;
  }
}

/* Mobile devices */
@media (max-width: 767px) {
  .feed-drawer-wrapper:not(.pinned) {
    left: 44px;
  }

  .feed-drawer-wrapper:not(.pinned).activity-bar-collapsed {
    left: 16px;
  }
}

/* Drawer position transition */
.drawer-position-enter-active {
  transition:
    transform 0.3s cubic-bezier(0.4, 0, 0.2, 1),
    opacity 0.2s ease;
  will-change: transform, opacity;
}

.drawer-position-leave-active {
  transition:
    transform 0.25s cubic-bezier(0.4, 0, 0.2, 1),
    opacity 0.2s ease;
  will-change: transform, opacity;
}

.drawer-position-enter-from {
  opacity: 0;
  transform: translateX(-16px);
}

.drawer-position-leave-to {
  opacity: 0;
  transform: translateX(-16px);
}

.drawer-position-enter-to,
.drawer-position-leave-from {
  opacity: 1;
  transform: translateX(0);
}

/* Optimize feed drawer rendering */
.feed-drawer-wrapper {
  backface-visibility: hidden;
  -webkit-font-smoothing: antialiased;
  transform: translateZ(0);
}

/* Overlay transition */
.overlay-fade-enter-active,
.overlay-fade-leave-active {
  transition: opacity 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  will-change: opacity;
}

.overlay-fade-enter-from,
.overlay-fade-leave-to {
  opacity: 0;
}

.overlay-fade-enter-to,
.overlay-fade-leave-from {
  opacity: 1;
}
</style>
