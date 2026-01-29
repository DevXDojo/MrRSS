<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, type Ref, type CSSProperties } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhX,
  PhMagnifyingGlassMinus,
  PhMagnifyingGlassPlus,
  PhDownloadSimple,
  PhCopy,
} from '@phosphor-icons/vue';

const { t } = useI18n();

interface Props {
  src: string;
  alt?: string;
  images?: string[];
  initialIndex?: number;
}

const props = withDefaults(defineProps<Props>(), {
  alt: '',
  images: () => [],
  initialIndex: 0,
});

const emit = defineEmits<{
  close: [];
}>();

interface Position {
  x: number;
  y: number;
}

const scale = ref(1);
const position = ref<Position>({ x: 0, y: 0 });
const isDragging = ref(false);
const dragStart = ref<Position>({ x: 0, y: 0 });
const imageRef: Ref<HTMLImageElement | null> = ref(null);
const currentImageIndex = ref(0);

const MIN_SCALE = 0.5;
const MAX_SCALE = 5;
const SCALE_STEP = 0.25;

// Computed properties for multiple image support
const hasMultipleImages = computed(() => props.images && props.images.length > 1);
const currentSrc = computed(() => {
  if (hasMultipleImages.value && props.images[currentImageIndex.value]) {
    return props.images[currentImageIndex.value];
  }
  return props.src;
});
const currentImageNumber = computed(() => currentImageIndex.value + 1);
const totalImages = computed(() => (hasMultipleImages.value ? props.images.length : 1));

// Initialize index from props or find src in images
onMounted(() => {
  if (hasMultipleImages.value) {
    if (props.initialIndex >= 0 && props.initialIndex < props.images.length) {
      currentImageIndex.value = props.initialIndex;
    } else {
      // Try to find the current src in the images array
      const foundIndex = props.images.findIndex((img) => img === props.src);
      if (foundIndex >= 0) {
        currentImageIndex.value = foundIndex;
      }
    }
  }
  // Use capture phase to handle Escape before other listeners
  document.addEventListener('keydown', handleKeyDown, { capture: true } as any);
});

onUnmounted(() => {
  // Remove with capture option to match the addEventListener call
  document.removeEventListener('keydown', handleKeyDown, { capture: true } as any);
});

function close() {
  emit('close');
}

function zoomIn() {
  if (scale.value < MAX_SCALE) {
    scale.value = Math.min(scale.value + SCALE_STEP, MAX_SCALE);
  }
}

function zoomOut() {
  if (scale.value > MIN_SCALE) {
    scale.value = Math.max(scale.value - SCALE_STEP, MIN_SCALE);
    // Reset position if zooming out to 1 or less
    if (scale.value <= 1) {
      position.value = { x: 0, y: 0 };
    }
  }
}

function previousImage() {
  if (!hasMultipleImages.value) return;
  if (currentImageIndex.value > 0) {
    currentImageIndex.value--;
  } else {
    // Wrap to last image
    currentImageIndex.value = props.images.length - 1;
  }
  // Reset zoom and position when changing images
  resetView();
}

function nextImage() {
  if (!hasMultipleImages.value) return;
  if (currentImageIndex.value < props.images.length - 1) {
    currentImageIndex.value++;
  } else {
    // Wrap to first image
    currentImageIndex.value = 0;
  }
  // Reset zoom and position when changing images
  resetView();
}

function resetView() {
  scale.value = 1;
  position.value = { x: 0, y: 0 };
}

function handleWheel(e: WheelEvent) {
  e.preventDefault();
  if (e.deltaY < 0) {
    zoomIn();
  } else {
    zoomOut();
  }
}

function startDrag(e: MouseEvent) {
  isDragging.value = true;
  dragStart.value = {
    x: e.clientX - position.value.x,
    y: e.clientY - position.value.y,
  };
}

function onDrag(e: MouseEvent) {
  if (isDragging.value) {
    position.value = {
      x: e.clientX - dragStart.value.x,
      y: e.clientY - dragStart.value.y,
    };
  }
}

function stopDrag() {
  isDragging.value = false;
}

function handleKeyDown(e: KeyboardEvent) {
  // Only handle Escape key - stop it immediately to prevent other modals from closing
  if (e.key === 'Escape') {
    e.stopImmediatePropagation();
    close();
    return;
  }

  // Check if image viewer is open for other keys
  const imageViewer = document.querySelector('[data-image-viewer="true"]');
  if (!imageViewer) return;

  if (e.key === '+' || e.key === '=') {
    zoomIn();
  } else if (e.key === '-' || e.key === '_') {
    zoomOut();
  } else if (e.key === 'ArrowLeft') {
    e.preventDefault();
    previousImage();
  } else if (e.key === 'ArrowRight') {
    e.preventDefault();
    nextImage();
  } else if (e.key === 's' && (e.ctrlKey || e.metaKey)) {
    e.preventDefault();
    downloadImage();
  }
}

// Download image to local storage
async function downloadImage() {
  try {
    const response = await fetch(currentSrc.value);

    // Check if the response is successful
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const blob = await response.blob();

    // Extract and sanitize filename from URL
    let filename = 'image';
    try {
      const url = new URL(currentSrc.value);
      const pathname = url.pathname;
      const pathSegments = pathname.split('/').filter((segment) => segment.length > 0);
      if (pathSegments.length > 0) {
        const lastSegment = pathSegments[pathSegments.length - 1];
        // Remove query params and sanitize filename
        filename = lastSegment.split('?')[0].replace(/[^a-zA-Z0-9._-]/g, '_') || 'image';
      }
    } catch {
      // If URL parsing fails, use default filename
      filename = 'image';
    }

    // Ensure it has a valid extension based on MIME type
    if (!filename.match(/\.(jpg|jpeg|png|gif|webp|svg|bmp)$/i)) {
      const mimeType = blob.type;
      const ext = mimeType.split('/')[1]?.replace('jpeg', 'jpg') || 'png';
      filename = `${filename}.${ext}`;
    }

    // Create download link
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  } catch (error) {
    console.error('Failed to download image:', error);
    // Fallback: open image in new tab for manual saving
    window.open(currentSrc.value, '_blank');
  }
}

// Copy image (convert to PNG)
async function copyImage() {
  try {
    const response = await fetch(currentSrc.value);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const blob = await response.blob();

    // Convert to PNG for maximum clipboard compatibility
    const pngBlob = await new Promise<Blob>((resolve, reject) => {
      const img = new Image();
      img.crossOrigin = 'anonymous';

      img.onload = () => {
        const canvas = document.createElement('canvas');
        canvas.width = img.width;
        canvas.height = img.height;
        const ctx = canvas.getContext('2d');
        if (ctx) {
          ctx.drawImage(img, 0, 0);
          canvas.toBlob((convertedBlob) => {
            if (convertedBlob) {
              resolve(convertedBlob);
            } else {
              reject(new Error('Failed to convert image to PNG'));
            }
          }, 'image/png');
        } else {
          reject(new Error('Failed to get canvas context'));
        }
      };

      img.onerror = () => {
        reject(new Error('Failed to load image for conversion'));
      };

      img.src = URL.createObjectURL(blob);
    });

    // Copy to clipboard using only PNG format (widely supported)
    await navigator.clipboard.write([
      new ClipboardItem({
        'image/png': pngBlob,
      }),
    ]);

    window.showToast(t('common.toast.copiedToClipboard'), 'success');
  } catch (error) {
    console.error('Failed to copy image:', error);
    window.showToast(t('common.errors.failedToCopy'), 'error');
  }
}

const imageStyle = computed<CSSProperties>(() => ({
  transform: `translate(${position.value.x}px, ${position.value.y}px) scale(${scale.value})`,
}));
</script>

<template>
  <div class="fixed inset-0 z-50 flex flex-col bg-black/90" data-image-viewer="true">
    <!-- Close button -->
    <button class="close-btn" @click="close">
      <PhX :size="20" />
    </button>

    <!-- Image counter (when multiple images) -->
    <div v-if="hasMultipleImages" class="image-counter">
      {{ currentImageNumber }} / {{ totalImages }}
    </div>

    <!-- Navigation buttons (when multiple images) -->
    <template v-if="hasMultipleImages">
      <button class="nav-btn nav-btn-prev" @click.stop="previousImage">‹</button>
      <button class="nav-btn nav-btn-next" @click.stop="nextImage">›</button>
    </template>

    <!-- Controls -->
    <div class="controls-container" @click.stop>
      <button
        class="control-btn"
        :disabled="scale <= MIN_SCALE"
        :title="t('common.imageViewer.zoomOut')"
        @click="zoomOut"
      >
        <PhMagnifyingGlassMinus :size="20" />
      </button>
      <span class="control-btn scale-display">{{ Math.round(scale * 100) }}%</span>
      <button
        class="control-btn"
        :disabled="scale >= MAX_SCALE"
        :title="t('common.imageViewer.zoomIn')"
        @click="zoomIn"
      >
        <PhMagnifyingGlassPlus :size="20" />
      </button>
      <button class="control-btn" :title="t('common.contextMenu.copyImage')" @click="copyImage">
        <PhCopy :size="20" />
      </button>
      <button
        class="control-btn"
        :title="t('common.contextMenu.downloadImage')"
        @click="downloadImage"
      >
        <PhDownloadSimple :size="20" />
      </button>
    </div>

    <!-- Image Container -->
    <div class="flex-1 flex items-center justify-center min-h-0 relative" @click="close">
      <div
        class="relative w-full h-full flex items-center justify-center overflow-hidden image-container"
        :class="{ 'cursor-grab': !isDragging, 'cursor-grabbing': isDragging }"
        @click.stop
        @wheel="handleWheel"
        @mousedown="startDrag"
        @mousemove="onDrag"
        @mouseup="stopDrag"
        @mouseleave="stopDrag"
      >
        <img
          ref="imageRef"
          :src="currentSrc"
          :alt="alt"
          :style="imageStyle"
          class="max-w-full max-h-full object-contain select-none"
          :class="[isDragging ? '' : 'transition-transform duration-150']"
          @dragstart.prevent
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../style.css";

/* Close button */
.close-btn {
  @apply absolute top-4 right-4 w-8 h-8 bg-black/50 hover:bg-black/70;
  @apply rounded-full text-white flex items-center justify-center;
  @apply transition-colors duration-200 z-10 shrink-0;
}

/* Image counter */
.image-counter {
  @apply absolute top-4 left-4 px-2 py-1 rounded;
  @apply text-white text-sm font-medium min-w-[60px] text-center;
  @apply z-10 text-shadow;
}

/* Navigation buttons */
.nav-btn {
  @apply absolute top-1/2 -translate-y-1/2 w-12 h-12 rounded;
  @apply text-white text-4xl;
  @apply flex items-center justify-center transition-all duration-200;
  @apply z-10 text-shadow;
}

.nav-btn-prev {
  @apply left-4;
}

.nav-btn-next {
  @apply right-4;
}

.nav-btn:hover {
  @apply scale-110;
}

.nav-btn:active {
  @apply scale-95;
}

/* Text shadow for better visibility on images */
.text-shadow {
  text-shadow:
    0 1px 3px rgba(0, 0, 0, 0.8),
    0 1px 2px rgba(0, 0, 0, 0.6);
}

/* Controls container */
.controls-container {
  @apply absolute bottom-4 left-1/2 -translate-x-1/2 flex gap-2 px-3 py-2 rounded-lg;
  @apply bg-black/50 backdrop-blur-sm z-10;
}

.control-btn {
  @apply px-2 py-1.5 rounded transition-colors flex items-center justify-center min-w-[40px];
  @apply text-white hover:bg-white/10 active:bg-white/20;
  background-color: transparent;
}

.control-btn:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.scale-display {
  @apply text-sm font-medium w-[60px] text-center pointer-events-none;
}

/* Image container cursor */
.image-container {
  cursor: default;
}

.image-container.cursor-grab {
  cursor: grab;
}

.image-container.cursor-grabbing {
  cursor: grabbing;
}
</style>
