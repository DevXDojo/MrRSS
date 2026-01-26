<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue';
import ActivityBar from './ActivityBar.vue';
import FeedList from './FeedList.vue';

interface Props {
  isOpen?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  toggle: [];
}>();

// Feed drawer state
const isFeedListExpanded = ref(false);
const isFeedListPinned = ref(false);
const activityBarRef = ref<InstanceType<typeof ActivityBar> | null>(null);

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
</script>

<template>
  <div class="compact-sidebar-wrapper flex h-full relative">
    <!-- Smart Activity Bar (Left) -->
    <ActivityBar
      ref="activityBarRef"
      @add-feed="emitShowAddFeed"
      @settings="emitShowSettings"
      @toggle-feed-drawer="handleToggleFeedList"
      @ready="handleActivityBarReady"
    />

    <!-- Feed Drawer -->
    <Transition name="drawer-position">
      <div
        v-if="isFeedListExpanded"
        class="feed-drawer-wrapper"
        :class="{ pinned: isFeedListPinned }"
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
@reference "../../style.css";

.compact-sidebar-wrapper {
  position: relative;
  z-index: 10;
  display: flex;
  align-items: stretch;
}

:global(.macos) .compact-sidebar-wrapper {
  padding-top: 0;
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
  z-index: 10;
}

/* Smaller screens (laptops, tablets) */
@media (max-width: 1400px) {
  .feed-drawer-wrapper:not(.pinned) {
    left: 48px;
  }
}

/* Mobile devices */
@media (max-width: 767px) {
  .feed-drawer-wrapper:not(.pinned) {
    left: 44px;
  }
}

/* Drawer position transition */
.drawer-position-enter-active,
.drawer-position-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.drawer-position-enter-from,
.drawer-position-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

.drawer-position-enter-to,
.drawer-position-leave-from {
  opacity: 1;
  transform: translateX(0);
}

/* Overlay transition */
.overlay-fade-enter-active,
.overlay-fade-leave-active {
  transition: opacity 0.3s ease;
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
