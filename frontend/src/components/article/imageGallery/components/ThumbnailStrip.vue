<script setup lang="ts">
import { ref, watch, nextTick, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { getProxiedMediaUrl, isMediaCacheEnabled } from '@/utils/mediaProxy';

interface Props {
  images: string[];
  currentIndex: number;
  show: boolean;
  coverImageURL?: string; // Optional: URL of the cover image (always cached)
}

const props = defineProps<Props>();

const emit = defineEmits<{
  select: [index: number];
  toggle: [];
}>();

const { t } = useI18n();

// Track media cache setting
const mediaCacheEnabled = ref(false);

// Check media cache setting on mount
onMounted(async () => {
  mediaCacheEnabled.value = await isMediaCacheEnabled();
});

const thumbnailStripRef = ref<HTMLElement | null>(null);
const thumbnailStripWidth = ref(0);

// Update thumbnail strip width when the ref is available
watch(
  () => thumbnailStripRef.value,
  () => {
    if (thumbnailStripRef.value) {
      thumbnailStripWidth.value = thumbnailStripRef.value.offsetWidth;
    }
  }
);

// Watch for images changes and update thumbnail strip width
watch(
  () => props.images,
  async () => {
    await nextTick();
    if (thumbnailStripRef.value) {
      thumbnailStripWidth.value = thumbnailStripRef.value.offsetWidth;
    }
  }
);

// Watch for currentIndex changes and auto-scroll to keep current thumbnail visible
watch(
  () => props.currentIndex,
  async (newIndex) => {
    await nextTick();
    if (!thumbnailStripRef.value || newIndex < 0) return;

    // Each thumbnail is 64px (w-16) + 8px (gap-2) = 72px
    const thumbnailWidth = 72;
    const containerWidth = thumbnailStripRef.value.offsetWidth;
    const scrollLeft = thumbnailStripRef.value.scrollLeft;

    // Calculate thumbnail position
    const thumbnailLeft = newIndex * thumbnailWidth;
    const thumbnailRight = thumbnailLeft + thumbnailWidth;

    // Scroll into view if needed
    if (thumbnailLeft < scrollLeft) {
      // Thumbnail is to the left of viewport, scroll to show it
      thumbnailStripRef.value.scrollTo({
        left: thumbnailLeft,
        behavior: 'smooth',
      });
    } else if (thumbnailRight > scrollLeft + containerWidth) {
      // Thumbnail is to the right of viewport, scroll to show it
      thumbnailStripRef.value.scrollTo({
        left: thumbnailRight - containerWidth,
        behavior: 'smooth',
      });
    }
  }
);

/**
 * Check if thumbnails should be centered (when they don't fill the container)
 */
const shouldCenterThumbnails = () => {
  if (props.images.length === 0) return false;
  // Each thumbnail is 64px (w-16) + 8px (gap-2) = 72px
  const thumbnailWidth = 72;
  const totalThumbnailsWidth = props.images.length * thumbnailWidth;
  return totalThumbnailsWidth < thumbnailStripWidth.value;
};

/**
 * Handle mouse wheel on thumbnail strip for horizontal scrolling
 */
function handleThumbnailWheel(e: WheelEvent): void {
  if (!thumbnailStripRef.value) return;

  // Prevent vertical scrolling
  e.preventDefault();

  // Scroll horizontally with smooth behavior
  thumbnailStripRef.value.scrollBy({
    left: e.deltaY,
    behavior: 'smooth',
  });
}

/**
 * Handle thumbnail selection
 */
function selectThumbnail(index: number): void {
  emit('select', index);
}

/**
 * Get proxied URL for a thumbnail
 * Cover images are always cached
 * Other images are cached only if global media cache is enabled
 */
function getProxiedUrl(url: string): string {
  // Check if this is the cover image
  const isCoverImage = props.coverImageURL && url === props.coverImageURL;

  // Cover images are always cached to ensure they display correctly
  if (isCoverImage) {
    return getProxiedMediaUrl(url, undefined, true);
  }

  // Non-cover images: only cache if global media cache is enabled
  if (mediaCacheEnabled.value) {
    return getProxiedMediaUrl(url, undefined, true);
  }

  // If cache is disabled for non-cover images, use original URL directly
  return url;
}
</script>

<template>
  <!-- Thumbnail strip (only shown when there are multiple images) -->
  <div v-if="images.length > 1" class="w-full shrink-0" @click.stop>
    <!-- Collapsed state: show expand hint -->
    <div
      v-if="!show"
      class="relative w-full py-3 flex items-center justify-center"
      @click="emit('toggle')"
    >
      <div
        class="h-1 w-12 bg-white/30 rounded-full cursor-pointer hover:bg-white/50 hover:w-16 transition-all duration-300"
      ></div>
    </div>

    <!-- Expanded state: show thumbnails with collapse handle -->
    <template v-else>
      <!-- Collapse handle above thumbnails -->
      <div class="relative w-full py-2 flex items-center justify-center" @click="emit('toggle')">
        <div
          class="h-1 w-16 bg-white/20 rounded-full cursor-pointer hover:bg-white/40 hover:w-20 transition-all duration-300"
        ></div>
      </div>

      <!-- Thumbnail strip -->
      <div class="w-full px-2" @click.stop>
        <div
          ref="thumbnailStripRef"
          class="flex gap-2 overflow-x-auto pb-2 scrollbar-hide scroll-smooth"
          :class="shouldCenterThumbnails() ? 'justify-center' : 'justify-start'"
          @wheel="handleThumbnailWheel"
        >
          <button
            v-for="(image, index) in images"
            :key="index"
            class="relative shrink-0 w-16 h-16 rounded overflow-hidden border-2 transition-all duration-200"
            :class="
              index === currentIndex
                ? 'border-accent shadow-lg shadow-accent/30'
                : 'border-white/30 hover:border-white/60'
            "
            @click="selectThumbnail(index)"
          >
            <img
              :src="getProxiedUrl(image)"
              :alt="`${t('common.text.image')} ${index + 1}`"
              class="w-full h-full object-cover"
              loading="lazy"
            />
            <!-- Active indicator -->
            <div
              v-if="index === currentIndex"
              class="absolute inset-0 bg-accent/20 pointer-events-none"
            ></div>
          </button>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
/* Hide scrollbar but keep functionality */
.scrollbar-hide {
  -ms-overflow-style: none; /* IE and Edge */
  scrollbar-width: none; /* Firefox */
}

.scrollbar-hide::-webkit-scrollbar {
  display: none; /* Chrome, Safari and Opera */
}
</style>
