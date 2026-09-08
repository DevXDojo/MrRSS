<script setup lang="ts">
import { ref, watch } from 'vue';
import ActivityBar from './ActivityBar.vue';
import FeedList from './FeedList.vue';

interface Props {
  isOpen: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  toggle: [];
}>();

const isFeedListPinned = ref(localStorage.getItem('FeedListPinned') !== 'false');
watch(isFeedListPinned, (pinned) => {
  localStorage.setItem('FeedListPinned', String(pinned));
});

function handleFeedListExpand() {
  if (!props.isOpen) emit('toggle');
}

function handleFeedListCollapse() {
  if (props.isOpen) emit('toggle');
}

function handlePinFeedList() {
  isFeedListPinned.value = true;
  handleFeedListExpand();
}

function handleUnpinFeedList() {
  isFeedListPinned.value = false;
}

const emitShowAddFeed = () => window.dispatchEvent(new CustomEvent('show-add-feed'));
const emitShowSettings = () => window.dispatchEvent(new CustomEvent('show-settings'));
</script>

<template>
  <div class="compact-sidebar-wrapper flex h-full relative">
    <div class="sidebar-toggle-container">
      <ActivityBar
        :is-feed-list-expanded="isOpen"
        @add-feed="emitShowAddFeed"
        @settings="emitShowSettings"
        @toggle-feed-drawer="emit('toggle')"
      />
    </div>

    <!-- Feed Drawer -->
    <Transition name="drawer-position">
      <div v-if="isOpen" class="feed-drawer-wrapper" :class="{ pinned: isFeedListPinned }">
        <FeedList
          :is-expanded="isOpen"
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
        v-if="isOpen"
        class="fixed inset-0 bg-black/50 z-10 md:hidden"
        @click="handleFeedListCollapse"
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
}

/* Fixed-width navigation stays available while the feed drawer is closed. */
.sidebar-toggle-container {
  position: relative;
  width: 56px;
  min-width: 56px;
  height: 100%;
  flex-shrink: 0;
}

/* Smaller screens (laptops, tablets) */
@media (max-width: 1400px) {
  .sidebar-toggle-container {
    width: 48px;
    min-width: 48px;
  }
}

/* Mobile devices */
@media (max-width: 767px) {
  .sidebar-toggle-container {
    width: 44px;
    min-width: 44px;
  }
}

.feed-drawer-wrapper {
  position: relative;
  z-index: 20;
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
