<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick, watch } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import type { Article } from '@/types/models';
import {
  PhImage,
  PhHeart,
  PhList,
  PhCopy,
  PhDownloadSimple,
  PhX,
  PhTextT,
  PhTextTSlash,
  PhMagnifyingGlassPlus,
  PhMagnifyingGlassMinus,
} from '@phosphor-icons/vue';
import { openInBrowser } from '@/utils/browser';

const store = useAppStore();
const { t } = useI18n();

interface Props {
  isSidebarOpen?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  toggleSidebar: [];
}>();

// Constants
const ITEMS_PER_PAGE = 30;
const SCROLL_THRESHOLD_PX = 500; // Start loading more items when user is 500px from bottom
const MIN_SCALE = 0.5;
const MAX_SCALE = 5;
const SCALE_STEP = 0.25;

const articles = ref<Article[]>([]);
const isLoading = ref(false);
const page = ref(1);
const hasMore = ref(true);
const selectedArticle = ref<Article | null>(null);
const showImageViewer = ref(false);
const allImages = ref<string[]>([]);
const currentImageIndex = ref(0);
const currentImageLoading = ref(false);
const columns = ref<Article[][]>([]);
const columnCount = ref(4);
const containerRef = ref<HTMLElement | null>(null);
// eslint-disable-next-line no-undef
let resizeObserver: ResizeObserver | null = null;
const imageCountCache = ref<Map<number, number>>(new Map());
const showTextOverlay = ref(true);
const showThumbnailStrip = ref(true);
const thumbnailStripRef = ref<HTMLElement | null>(null);
const thumbnailStripWidth = ref(0);

// Image viewer zoom and pan
const scale = ref(1);
const position = ref<{ x: number; y: number }>({ x: 0, y: 0 });
const isDragging = ref(false);
const dragStart = ref<{ x: number; y: number }>({ x: 0, y: 0 });
const imageContainerRef = ref<HTMLElement | null>(null);

// Load showTextOverlay preference from localStorage
const savedShowTextOverlay = localStorage.getItem('imageGalleryShowTextOverlay');
if (savedShowTextOverlay !== null) {
  showTextOverlay.value = savedShowTextOverlay === 'true';
}

// Load showThumbnailStrip preference from localStorage
const savedShowThumbnailStrip = localStorage.getItem('imageGalleryShowThumbnailStrip');
if (savedShowThumbnailStrip !== null) {
  showThumbnailStrip.value = savedShowThumbnailStrip === 'true';
}

// Watch for changes and save to localStorage
watch(showTextOverlay, (newValue) => {
  localStorage.setItem('imageGalleryShowTextOverlay', String(newValue));
});

watch(showThumbnailStrip, (newValue) => {
  localStorage.setItem('imageGalleryShowThumbnailStrip', String(newValue));
});

// Compute which feed ID to fetch (if viewing a specific feed)
const feedId = computed(() => store.currentFeedId);

// Compute which category to fetch (if viewing a specific category)
const category = computed(() => store.currentCategory);

// Get current image URL
const currentImageUrl = computed(() => {
  if (allImages.value.length > 0 && currentImageIndex.value < allImages.value.length) {
    return allImages.value[currentImageIndex.value];
  }
  return selectedArticle.value?.image_url || '';
});

// Image style for zoom and pan
const imageStyle = computed(() => ({
  transform: `translate(${position.value.x}px, ${position.value.y}px) scale(${scale.value})`,
}));

// Find current article index in articles array
const currentArticleIndex = computed(() => {
  if (!selectedArticle.value) return -1;
  return articles.value.findIndex((a) => a.id === selectedArticle.value!.id);
});

// Check if we can navigate to previous image/article
const canNavigatePrevious = computed(() => {
  // Can navigate if not at first image of first article
  if (currentImageIndex.value > 0) return true;
  if (currentArticleIndex.value > 0) return true;
  return false;
});

// Check if we can navigate to next image/article
const canNavigateNext = computed(() => {
  // Can navigate if not at last image of last article
  if (currentImageIndex.value < allImages.value.length - 1) return true;
  if (currentArticleIndex.value >= 0 && currentArticleIndex.value < articles.value.length - 1)
    return true;
  return false;
});

// Check if thumbnails should be centered (when they don't fill the container)
const shouldCenterThumbnails = computed(() => {
  if (allImages.value.length === 0) return false;
  // Each thumbnail is 64px (w-16) + 8px (gap-2) = 72px
  const thumbnailWidth = 72;
  const totalThumbnailsWidth = allImages.value.length * thumbnailWidth;
  return totalThumbnailsWidth < thumbnailStripWidth.value;
});

// Fetch image gallery articles
async function fetchImages(loadMore = false) {
  if (isLoading.value) return;

  isLoading.value = true;
  try {
    let url = `/api/articles/images?page=${page.value}&limit=${ITEMS_PER_PAGE}`;
    if (feedId.value) {
      url += `&feed_id=${feedId.value}`;
    } else if (category.value !== null) {
      url += `&category=${encodeURIComponent(category.value)}`;
    }

    const res = await fetch(url);
    if (res.ok) {
      const data = await res.json();

      // Validate that data is an array
      if (!Array.isArray(data)) {
        console.error('API response is not an array:', data);
        return;
      }

      const newArticles = data;

      if (loadMore) {
        articles.value = [...articles.value, ...newArticles];
      } else {
        articles.value = newArticles;
      }

      hasMore.value = newArticles.length >= ITEMS_PER_PAGE;

      // Preload image counts for new articles
      newArticles.forEach((article: Article) => {
        if (!imageCountCache.value.has(article.id)) {
          fetchImageCount(article.id);
        }
      });
    }
  } catch (e) {
    console.error('Failed to load images:', e);
  } finally {
    isLoading.value = false;
  }
}

// Fetch image count for an article
async function fetchImageCount(articleId: number) {
  try {
    const res = await fetch(`/api/articles/extract-images?id=${articleId}`);
    if (res.ok) {
      const data = await res.json();
      if (data.images && Array.isArray(data.images)) {
        imageCountCache.value.set(articleId, data.images.length);
      }
    }
  } catch (e) {
    console.error('Failed to fetch image count:', e);
  }
}

// Get image count for an article
function getImageCount(article: Article): number {
  return imageCountCache.value.get(article.id) || 1;
}

// Calculate number of columns based on container width dynamically
function calculateColumns() {
  if (!containerRef.value) return;
  const width = containerRef.value.offsetWidth;

  // Target column width: 250px for optimal image viewing
  // Minimum 2 columns, no maximum
  const targetColumnWidth = 250;
  const calculatedColumns = Math.floor(width / targetColumnWidth);

  // Ensure at least 2 columns
  columnCount.value = Math.max(2, calculatedColumns);

  // Rearrange columns after calculating new count
  arrangeColumns();
}

// Arrange articles into columns by time, balancing heights
function arrangeColumns() {
  if (articles.value.length === 0) {
    columns.value = [];
    return;
  }

  // Initialize columns
  const cols: Article[][] = Array.from({ length: columnCount.value }, () => []);
  const colHeights: number[] = Array(columnCount.value).fill(0);

  // Sort articles by published date (newest first)
  const sortedArticles = [...articles.value].sort((a, b) => {
    return new Date(b.published_at).getTime() - new Date(a.published_at).getTime();
  });

  // Place each article in the shortest column
  sortedArticles.forEach((article) => {
    const shortestColIndex = colHeights.indexOf(Math.min(...colHeights));
    cols[shortestColIndex].push(article);
    // Estimate height: 200px for image + 80px for info
    colHeights[shortestColIndex] += 280;
  });

  columns.value = cols;
}

// Handle scroll for infinite loading
function handleScroll() {
  if (!containerRef.value) return;

  const scrollTop = containerRef.value.scrollTop;
  const containerHeight = containerRef.value.clientHeight;
  const scrollHeight = containerRef.value.scrollHeight;

  if (
    scrollTop + containerHeight >= scrollHeight - SCROLL_THRESHOLD_PX &&
    !isLoading.value &&
    hasMore.value
  ) {
    // Increment page before fetching
    const nextPage = page.value + 1;
    page.value = nextPage;
    fetchImages(true);
  }
}

// Toggle favorite
async function toggleFavorite(article: Article, event?: Event) {
  if (event) {
    event.stopPropagation();
  }
  try {
    const res = await fetch(`/api/articles/favorite?id=${article.id}`, {
      method: 'POST',
    });
    if (res.ok) {
      article.is_favorite = !article.is_favorite;
      if (selectedArticle.value && selectedArticle.value.id === article.id) {
        selectedArticle.value.is_favorite = article.is_favorite;
      }
      // Update filter counts after toggling favorite status
      await store.fetchFilterCounts();
    }
  } catch (e) {
    console.error('Failed to toggle favorite:', e);
  }
}

// Open image viewer
async function openImage(article: Article) {
  selectedArticle.value = article;
  showImageViewer.value = true;
  currentImageLoading.value = true;
  // Reset zoom and position
  scale.value = 1;
  position.value = { x: 0, y: 0 };

  // Fetch all images from the article
  await fetchArticleImages(article);

  // Mark as read
  if (!article.is_read) {
    markAsRead(article);
  }
}

// Fetch all images from article content
async function fetchArticleImages(article: Article) {
  try {
    const res = await fetch(`/api/articles/extract-images?id=${article.id}`);
    if (res.ok) {
      const data = await res.json();
      if (data.images && Array.isArray(data.images) && data.images.length > 0) {
        allImages.value = data.images;
        // Find the index of the article's main image
        currentImageIndex.value = data.images.findIndex((img: string) => img === article.image_url);
        if (currentImageIndex.value < 0) {
          currentImageIndex.value = 0;
        }
      } else {
        // Fallback to just the article's main image
        allImages.value = [article.image_url || ''];
        currentImageIndex.value = 0;
      }
    } else {
      // Fallback on error
      allImages.value = [article.image_url || ''];
      currentImageIndex.value = 0;
    }
  } catch (e) {
    console.error('Failed to fetch article images:', e);
    // Fallback on error
    allImages.value = [article.image_url || ''];
    currentImageIndex.value = 0;
  }
}

// Navigate to previous image (with cross-article support)
async function previousImage() {
  // Check if we can navigate backward
  if (!canNavigatePrevious.value) return;

  // Reset zoom and position when navigating
  resetView();

  if (currentImageIndex.value > 0) {
    currentImageIndex.value--;
    // Reset loading state
    currentImageLoading.value = true;
  } else {
    // At first image of current article, go to previous article
    const prevArticle = articles.value[currentArticleIndex.value - 1];
    // Update selected article without closing viewer
    const wasFavorite = selectedArticle.value?.is_favorite;
    selectedArticle.value = prevArticle;
    if (wasFavorite !== undefined) {
      selectedArticle.value.is_favorite = wasFavorite;
    }
    currentImageLoading.value = true;

    await fetchArticleImages(prevArticle);
    // Move to last image
    currentImageIndex.value = allImages.value.length - 1;

    if (!prevArticle.is_read) {
      markAsRead(prevArticle);
    }
  }
}

// Navigate to next image (with cross-article support)
async function nextImage() {
  // Check if we can navigate forward
  if (!canNavigateNext.value) {
    // Try to load more articles if available
    if (hasMore.value) {
      await fetchImages(true);
      // Check again after loading
      if (!canNavigateNext.value) return;
    } else {
      return;
    }
  }

  // Reset zoom and position when navigating
  resetView();

  if (currentImageIndex.value < allImages.value.length - 1) {
    currentImageIndex.value++;
    // Reset loading state
    currentImageLoading.value = true;
  } else {
    // At last image of current article, go to next article
    const nextArticle = articles.value[currentArticleIndex.value + 1];
    // Update selected article without closing viewer
    const wasFavorite = selectedArticle.value?.is_favorite;
    selectedArticle.value = nextArticle;
    if (wasFavorite !== undefined) {
      selectedArticle.value.is_favorite = wasFavorite;
    }
    currentImageLoading.value = true;

    await fetchArticleImages(nextArticle);
    // Start at first image
    currentImageIndex.value = 0;

    if (!nextArticle.is_read) {
      markAsRead(nextArticle);
    }
  }
}

// Handle image load
function handleImageLoad() {
  currentImageLoading.value = false;
}

// Handle image error
function handleImageError() {
  currentImageLoading.value = false;
}

// Mark article as read
async function markAsRead(article: Article) {
  try {
    const res = await fetch(`/api/articles/read?id=${article.id}&read=true`, {
      method: 'POST',
    });
    if (res.ok) {
      article.is_read = true;
      // Update unread counts after marking as read
      await store.fetchUnreadCounts();
      await store.fetchFilterCounts();
    }
  } catch (e) {
    console.error('Failed to mark as read:', e);
  }
}

// Close image viewer
function closeImageViewer() {
  showImageViewer.value = false;
  selectedArticle.value = null;
  allImages.value = [];
  currentImageIndex.value = 0;
  scale.value = 1;
  position.value = { x: 0, y: 0 };
}

// Format date
function formatDate(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));

  if (days === 0) {
    const hours = Math.floor(diff / (1000 * 60 * 60));
    if (hours === 0) {
      const minutes = Math.floor(diff / (1000 * 60));
      return minutes <= 0
        ? t('common.time.justNow')
        : t('common.time.minutesAgo', { count: minutes });
    }
    return t('common.time.hoursAgo', { count: hours });
  } else if (days < 7) {
    return t('common.time.daysAgo', { count: days });
  }
  return date.toLocaleDateString();
}

// Handle right-click context menu
function handleContextMenu(event: MouseEvent, article: Article) {
  event.preventDefault();
  event.stopPropagation();

  // Use global context menu system to avoid conflicts with sidebar/context menu
  const menuItems = [
    {
      label: article.is_read ? t('article.action.markAsUnread') : t('article.action.markAsRead'),
      action: 'toggleRead',
      icon: article.is_read ? 'ph-envelope' : 'ph-envelope-open',
    },
    {
      label: article.is_favorite
        ? t('article.action.removeFromFavorites')
        : t('article.imageGallery.addToFavorite'),
      action: 'toggleFavorite',
      icon: 'ph-star',
      iconWeight: article.is_favorite ? 'fill' : 'regular',
      iconColor: article.is_favorite ? 'text-yellow-500' : '',
    },
    { separator: true },
    {
      label: t('common.contextMenu.copyTitle'),
      action: 'copyTitle',
      icon: 'ph-text-t',
    },
    {
      label: t('common.contextMenu.copyLink'),
      action: 'copyLink',
      icon: 'ph-link',
    },
    { separator: true },
    {
      label: t('common.contextMenu.downloadImage'),
      action: 'downloadImage',
      icon: 'PhDownloadSimple',
    },
    {
      label: t('article.action.openInBrowser'),
      action: 'openBrowser',
      icon: 'ph-globe',
    },
  ];

  window.dispatchEvent(
    new CustomEvent('open-context-menu', {
      detail: {
        x: event.clientX,
        y: event.clientY,
        items: menuItems,
        data: article,
        callback: handleImageAction,
      },
    })
  );
}

// Handle context menu actions
async function handleImageAction(action: string, article: Article): Promise<void> {
  if (action === 'toggleRead') {
    await toggleReadStatus(article);
  } else if (action === 'toggleFavorite') {
    await toggleFavorite(article);
  } else if (action === 'copyTitle') {
    await copyArticleTitle(article);
  } else if (action === 'copyLink') {
    await copyArticleLink(article);
  } else if (action === 'downloadImage') {
    await downloadImage(article.image_url || '');
  } else if (action === 'openBrowser') {
    openOriginal(article);
  }
}

// Download image
async function downloadImage(src: string) {
  try {
    const response = await fetch(src);
    const blob = await response.blob();

    // Extract and sanitize filename from URL
    let filename = 'image';
    try {
      const url = new URL(src);
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

    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  } catch (e) {
    console.error('Failed to download image:', e);
    window.showToast(t('common.toast.downloadFailed'), 'error');
  }
}

// Copy image (convert to PNG)
async function copyImage(src: string) {
  try {
    const response = await fetch(src);
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

// Open original article
function openOriginal(article: Article) {
  openInBrowser(article.url);
}

// Toggle article read status
async function toggleReadStatus(article: Article) {
  const newState = !article.is_read;
  article.is_read = newState;
  try {
    await fetch(`/api/articles/read?id=${article.id}&read=${newState}`, {
      method: 'POST',
    });
    // Update unread counts after toggling read status
    await store.fetchUnreadCounts();
    await store.fetchFilterCounts();
  } catch (e) {
    console.error('Error toggling read status:', e);
    // Revert the state change on error
    article.is_read = !newState;
    window.showToast(t('common.errors.savingSettings'), 'error');
  }
}

// Copy article title
async function copyArticleTitle(article: Article) {
  try {
    await navigator.clipboard.writeText(article.title);
    window.showToast(t('common.toast.copiedToClipboard'), 'success');
  } catch (error) {
    console.error('Failed to copy title:', error);
    window.showToast(t('common.errors.failedToCopy'), 'error');
  }
}

// Copy article link
async function copyArticleLink(article: Article) {
  try {
    await navigator.clipboard.writeText(article.url);
    window.showToast(t('common.toast.copiedToClipboard'), 'success');
  } catch (error) {
    console.error('Failed to copy link:', error);
    window.showToast(t('common.errors.failedToCopy'), 'error');
  }
}

// Open article in detail view
function openArticleDetail() {
  if (!selectedArticle.value) return;

  // Set the current article ID in the store
  store.currentArticleId = selectedArticle.value.id;

  // Switch to 'all' filter to exit image gallery mode and show article detail
  store.setFilter('all');

  // Close the image viewer
  closeImageViewer();
}

// Zoom functions
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

function resetView() {
  scale.value = 1;
  position.value = { x: 0, y: 0 };
}

// Drag functions
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

// Handle keyboard shortcuts
function handleKeyDown(e: KeyboardEvent) {
  // Handle Escape key - stop it immediately when image viewer is open
  if (e.key === 'Escape' && showImageViewer.value) {
    e.stopImmediatePropagation();
    closeImageViewer();
    return;
  }

  // Only handle other keyboard shortcuts when image viewer is open
  if (!showImageViewer.value) return;

  if (e.key === 'ArrowLeft') {
    e.preventDefault();
    previousImage();
  } else if (e.key === 'ArrowRight') {
    e.preventDefault();
    nextImage();
  } else if (e.key === '+' || e.key === '=') {
    e.preventDefault();
    zoomIn();
  } else if (e.key === '-' || e.key === '_') {
    e.preventDefault();
    zoomOut();
  }
}

// Handle mouse wheel on thumbnail strip for horizontal scrolling
function handleThumbnailWheel(e: WheelEvent) {
  if (!thumbnailStripRef.value) return;

  // Prevent vertical scrolling
  e.preventDefault();

  // Scroll horizontally with smooth behavior
  thumbnailStripRef.value.scrollBy({
    left: e.deltaY,
    behavior: 'smooth',
  });
}

// Handle mouse wheel on main image area for navigation
function handleImageWheel(e: WheelEvent) {
  // For single image articles, allow navigation across articles
  // For multiple image articles, navigate within the article

  // Determine direction
  const isNavigatingForward = e.deltaY > 0 || e.deltaX > 0;
  const isNavigatingBackward = e.deltaY < 0 || e.deltaX < 0;

  // Check if navigation is possible
  if (isNavigatingForward && !canNavigateNext.value) return;
  if (isNavigatingBackward && !canNavigatePrevious.value) return;

  // Prevent default scrolling only if we can navigate
  e.preventDefault();

  // Navigate
  if (isNavigatingForward) {
    nextImage();
  } else if (isNavigatingBackward) {
    previousImage();
  }
}

// Watch for articles changes and rearrange
watch(articles, () => {
  nextTick(() => {
    arrangeColumns();
  });
});

// Watch for feed ID changes and refetch
watch(feedId, async () => {
  // Close image viewer when switching feeds
  closeImageViewer();

  page.value = 1;
  articles.value = [];
  hasMore.value = true;
  await fetchImages();
  // Recalculate columns after fetching new articles
  await nextTick();
  calculateColumns();
});

// Watch for category changes and refetch
watch(category, async () => {
  // Close image viewer when switching categories
  closeImageViewer();

  page.value = 1;
  articles.value = [];
  hasMore.value = true;
  await fetchImages();
  // Recalculate columns after fetching new articles
  await nextTick();
  calculateColumns();
});

// Update thumbnail strip width when the ref is available
watch(thumbnailStripRef, () => {
  if (thumbnailStripRef.value) {
    thumbnailStripWidth.value = thumbnailStripRef.value.offsetWidth;
  }
});

// Watch for allImages changes and update thumbnail strip width
watch(allImages, async () => {
  await nextTick();
  if (thumbnailStripRef.value) {
    thumbnailStripWidth.value = thumbnailStripRef.value.offsetWidth;
  }
});

onMounted(() => {
  fetchImages();
  if (containerRef.value) {
    containerRef.value.addEventListener('scroll', handleScroll);
  }
  // Use capture phase to handle Escape before other listeners
  window.addEventListener('keydown', handleKeyDown, { capture: true } as any);

  // Set up ResizeObserver to watch for container size changes
  if (containerRef.value) {
    // eslint-disable-next-line no-undef
    resizeObserver = new ResizeObserver(() => {
      calculateColumns();
    });
    resizeObserver.observe(containerRef.value);
  }
});

onUnmounted(() => {
  if (containerRef.value) {
    containerRef.value.removeEventListener('scroll', handleScroll);
  }
  if (resizeObserver && containerRef.value) {
    resizeObserver.unobserve(containerRef.value);
    resizeObserver.disconnect();
    resizeObserver = null;
  }
  // Remove with capture option to match the addEventListener call
  window.removeEventListener('keydown', handleKeyDown, { capture: true } as any);
});
</script>

<template>
  <div class="flex flex-col flex-1 h-full bg-bg-primary">
    <!-- Header -->
    <div
      class="flex-shrink-0 bg-bg-primary border-b border-border p-2 sm:p-4 flex items-center gap-3"
    >
      <button
        class="p-2 rounded-lg hover:bg-bg-tertiary text-text-primary transition-colors md:hidden"
        :title="t('shortcut.toggle.sidebar')"
        @click="emit('toggleSidebar')"
      >
        <PhList :size="24" />
      </button>
      <div class="flex items-center gap-2 sm:gap-2 flex-1">
        <h1 class="text-base sm:text-lg font-bold text-text-primary line-height-fixed-32">
          {{ t('sidebar.activity.imageGallery') }}
        </h1>
      </div>
      <button
        class="p-1 sm:p-1.5 rounded hover:bg-bg-tertiary text-text-primary transition-colors"
        :title="showTextOverlay ? t('setting.reading.hideText') : t('setting.reading.showText')"
        @click="showTextOverlay = !showTextOverlay"
      >
        <PhTextTSlash v-if="showTextOverlay" :size="20" />
        <PhTextT v-else :size="20" />
      </button>
    </div>

    <!-- Scrollable content area -->
    <div ref="containerRef" class="flex-1 overflow-y-scroll scroll-smooth">
      <!-- Masonry Grid -->
      <div v-if="articles.length > 0" class="p-4 flex gap-4">
        <div
          v-for="(column, colIndex) in columns"
          :key="colIndex"
          class="flex-1 flex flex-col gap-4"
        >
          <div
            v-for="article in column"
            :key="article.id"
            class="cursor-pointer group"
            @click="openImage(article)"
            @contextmenu="handleContextMenu($event, article)"
          >
            <div
              class="relative overflow-hidden rounded-lg bg-bg-secondary transition-transform duration-200 hover:scale-[1.02]"
            >
              <img
                :src="article.image_url"
                :alt="article.title"
                class="w-full h-auto block"
                loading="lazy"
              />
              <!-- Image count indicator -->
              <div
                v-if="getImageCount(article) > 1"
                class="absolute bottom-2 left-2 px-2 py-1 rounded-full bg-black/60 text-white text-xs font-semibold backdrop-blur-sm z-10 flex items-center gap-1"
              >
                <PhImage :size="14" />
                <span class="ml-1">{{ getImageCount(article) }}</span>
              </div>
              <div
                class="absolute inset-0 bg-black/0 hover:bg-black/30 transition-all duration-200 flex items-start justify-end p-2"
              >
                <button
                  class="opacity-0 group-hover:opacity-100 transition-opacity duration-200 bg-black/50 rounded-full p-1.5 hover:bg-black/70"
                  @click="toggleFavorite(article, $event)"
                >
                  <PhHeart
                    :size="20"
                    :weight="article.is_favorite ? 'fill' : 'regular'"
                    :class="article.is_favorite ? 'text-red-500' : 'text-white'"
                  />
                </button>
              </div>
              <!-- Hover overlay when text is hidden -->
              <div
                v-if="!showTextOverlay"
                class="absolute inset-x-0 bottom-0 p-3 bg-gradient-to-t from-black/80 via-black/50 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-200"
              >
                <p class="text-sm font-medium text-white line-clamp-2 mb-1">
                  {{ article.title }}
                </p>
                <div class="flex items-center justify-between text-xs text-white/80">
                  <span class="truncate flex-1">{{ article.feed_title }}</span>
                  <span class="ml-2 shrink-0">{{ formatDate(article.published_at) }}</span>
                </div>
              </div>
            </div>
            <div v-if="showTextOverlay" class="p-2">
              <p class="text-sm font-medium text-text-primary line-clamp-2 mb-1">
                {{ article.title }}
              </p>
              <div class="flex items-center justify-between text-xs text-text-secondary">
                <span class="truncate flex-1">{{ article.feed_title }}</span>
                <span class="ml-2 shrink-0">{{ formatDate(article.published_at) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div
        v-else-if="!isLoading"
        class="flex flex-col items-center justify-center h-full w-full gap-4"
      >
        <PhImage :size="64" class="text-text-secondary opacity-50" />
        <p class="text-text-secondary">{{ t('article.content.noArticles') }}</p>
      </div>

      <!-- Loading Indicator -->
      <div v-if="isLoading" class="flex justify-center py-8">
        <div
          class="w-8 h-8 border-4 border-accent border-t-transparent rounded-full animate-spin"
        ></div>
      </div>
    </div>

    <!-- Image Viewer Modal -->
    <div
      v-if="showImageViewer && selectedArticle"
      class="fixed inset-0 z-50 bg-black/90 flex flex-col p-4"
      data-image-viewer="true"
      @click="closeImageViewer"
    >
      <!-- Top bar: Close button, Image counter, Zoom controls, Action buttons -->
      <div class="relative shrink-0 mb-2" @click.stop>
        <!-- Left: Image counter -->
        <div class="absolute left-0 top-0 flex items-center gap-2">
          <div
            v-if="allImages.length > 1"
            class="px-2 py-1 rounded bg-black/50 text-white text-sm font-medium min-w-[60px] text-center backdrop-blur-sm"
          >
            {{ currentImageIndex + 1 }} / {{ allImages.length }}
          </div>
        </div>

        <!-- Center: Zoom controls and Action buttons -->
        <div class="flex items-center justify-center gap-2">
          <!-- Zoom controls -->
          <button
            class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
            :disabled="scale <= MIN_SCALE"
            :title="t('common.imageViewer.zoomOut')"
            @click="zoomOut"
          >
            <PhMagnifyingGlassMinus :size="20" />
          </button>
          <span
            class="px-2 py-1.5 rounded bg-black/50 text-white text-sm font-medium min-w-[60px] text-center"
          >
            {{ Math.round(scale * 100) }}%
          </span>
          <button
            class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
            :disabled="scale >= MAX_SCALE"
            :title="t('common.imageViewer.zoomIn')"
            @click="zoomIn"
          >
            <PhMagnifyingGlassPlus :size="20" />
          </button>

          <!-- Action buttons -->
          <button
            class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
            :title="t('common.contextMenu.copyImage')"
            @click="copyImage(currentImageUrl)"
          >
            <PhCopy :size="20" />
          </button>
          <button
            class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
            :title="t('common.contextMenu.downloadImage')"
            @click="downloadImage(currentImageUrl)"
          >
            <PhDownloadSimple :size="20" />
          </button>
          <button
            class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
            :title="
              selectedArticle.is_favorite
                ? t('article.imageGallery.actionUnfavorite')
                : t('article.imageGallery.actionFavorite')
            "
            @click="toggleFavorite(selectedArticle)"
          >
            <PhHeart
              :size="20"
              :weight="selectedArticle.is_favorite ? 'fill' : 'regular'"
              :class="selectedArticle.is_favorite ? 'text-red-500' : 'text-white'"
            />
          </button>
        </div>

        <!-- Right: Close button -->
        <div class="absolute right-0 top-0">
          <button
            class="w-8 h-8 bg-black/50 hover:bg-black/70 rounded-full text-white flex items-center justify-center transition-colors"
            @click="closeImageViewer"
          >
            <PhX :size="20" />
          </button>
        </div>
      </div>

      <!-- Navigation buttons -->
      <template v-if="canNavigatePrevious">
        <button
          class="absolute top-[calc(50%-64px-8px)] left-4 -translate-y-1/2 w-12 h-12 rounded text-white text-4xl flex items-center justify-center transition-all duration-200 hover:scale-110 active:scale-95 z-10"
          style="
            text-shadow:
              0 1px 3px rgba(0, 0, 0, 0.8),
              0 1px 2px rgba(0, 0, 0, 0.6);
          "
          @click.stop="previousImage"
        >
          ‹
        </button>
      </template>
      <template v-if="canNavigateNext">
        <button
          class="absolute top-[calc(50%-64px-8px)] right-4 -translate-y-1/2 w-12 h-12 rounded text-white text-4xl flex items-center justify-center transition-all duration-200 hover:scale-110 active:scale-95 z-10"
          style="
            text-shadow:
              0 1px 3px rgba(0, 0, 0, 0.8),
              0 1px 2px rgba(0, 0, 0, 0.6);
          "
          @click.stop="nextImage"
        >
          ›
        </button>
      </template>

      <div class="flex-1 flex flex-col items-center justify-center min-h-0 relative" @click.stop>
        <div
          ref="imageContainerRef"
          class="flex-1 flex items-center justify-center w-full min-h-0 overflow-hidden"
          :class="{
            'cursor-grab': !isDragging,
            'cursor-grabbing': isDragging,
          }"
          @wheel="handleImageWheel"
          @mousedown="startDrag"
          @mousemove="onDrag"
          @mouseup="stopDrag"
          @mouseleave="stopDrag"
        >
          <!-- Loading placeholder -->
          <div
            v-if="currentImageLoading"
            class="absolute inset-0 flex items-center justify-center z-10"
          >
            <div
              class="w-12 h-12 border-4 border-white/20 border-t-white rounded-full animate-spin"
            ></div>
          </div>

          <img
            :src="currentImageUrl"
            :alt="selectedArticle.title"
            class="max-w-full max-h-full object-contain select-none"
            :class="[
              isDragging ? '' : 'transition-transform duration-150',
              { 'opacity-0': currentImageLoading },
            ]"
            :style="imageStyle"
            @load="handleImageLoad"
            @error="handleImageError"
            @dragstart.prevent
          />
        </div>

        <!-- Thumbnail strip toggle handle (always visible when there are multiple images) -->
        <div v-if="allImages.length > 1" class="w-full shrink-0" @click.stop>
          <!-- Collapsed state: show expand hint -->
          <div
            v-if="!showThumbnailStrip"
            class="relative w-full py-3 flex items-center justify-center"
          >
            <div
              class="h-1 w-12 bg-white/30 rounded-full cursor-pointer hover:bg-white/50 hover:w-16 transition-all duration-300"
              @click="showThumbnailStrip = true"
            ></div>
          </div>

          <!-- Expanded state: show thumbnails with collapse handle -->
          <template v-else>
            <!-- Collapse handle above thumbnails -->
            <div class="relative w-full py-2 flex items-center justify-center">
              <div
                class="h-1 w-16 bg-white/20 rounded-full cursor-pointer hover:bg-white/40 hover:w-20 transition-all duration-300"
                @click="showThumbnailStrip = false"
              ></div>
            </div>

            <!-- Thumbnail strip -->
            <div class="w-full px-2" @click.stop>
              <div
                ref="thumbnailStripRef"
                class="flex gap-2 overflow-x-auto pb-2 scrollbar-hide scroll-smooth"
                :class="shouldCenterThumbnails ? 'justify-center' : 'justify-start'"
                @wheel="handleThumbnailWheel"
              >
                <button
                  v-for="(image, index) in allImages"
                  :key="index"
                  class="relative shrink-0 w-16 h-16 rounded overflow-hidden border-2 transition-all duration-200"
                  :class="
                    index === currentImageIndex
                      ? 'border-accent shadow-lg shadow-accent/30'
                      : 'border-white/30 hover:border-white/60'
                  "
                  @click="
                    currentImageIndex = index;
                    currentImageLoading = true;
                    resetView();
                  "
                >
                  <img
                    :src="image"
                    :alt="`${t('common.text.image')} ${index + 1}`"
                    class="w-full h-full object-cover"
                    loading="lazy"
                  />
                  <!-- Active indicator -->
                  <div
                    v-if="index === currentImageIndex"
                    class="absolute inset-0 bg-accent/20 pointer-events-none"
                  ></div>
                </button>
              </div>
            </div>
          </template>
        </div>
      </div>

      <!-- Info bar with expandable content -->
      <div class="mt-2 px-3 py-3 rounded-lg bg-black/60 backdrop-blur-sm shrink-0" @click.stop>
        <!-- Basic info -->
        <div class="flex items-center justify-between gap-4 mb-2">
          <h2 class="text-base font-bold text-white flex-1 line-clamp-2">
            {{ selectedArticle.title }}
          </h2>
          <div class="flex items-center gap-2 shrink-0">
            <a
              :href="selectedArticle.url"
              target="_blank"
              rel="noopener noreferrer"
              class="px-3 py-1.5 bg-accent hover:bg-accent-hover text-white rounded-md text-sm whitespace-nowrap transition-colors duration-200"
            >
              {{ t('article.action.viewOriginal') }}
            </a>
            <button
              class="px-3 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded-md text-sm whitespace-nowrap transition-all duration-200"
              :title="t('article.action.viewArticle')"
              @click="openArticleDetail"
            >
              {{ t('article.action.viewArticle') }}
            </button>
          </div>
        </div>
        <div class="flex items-center gap-4 text-sm text-white/80">
          <span class="truncate flex-1">{{ selectedArticle.feed_title }}</span>
          <span class="shrink-0">{{ formatDate(selectedArticle.published_at) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Define keyframes for spinner animation */
@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* Hide scrollbar but keep functionality */
.scrollbar-hide {
  -ms-overflow-style: none; /* IE and Edge */
  scrollbar-width: none; /* Firefox */
}

.scrollbar-hide::-webkit-scrollbar {
  display: none; /* Chrome, Safari and Opera */
}

/* Prose content styling */
.prose-content {
  line-height: 1.6;
}

.prose-content :deep(img) {
  max-width: 100%;
  height: auto;
}

.prose-content :deep(p) {
  margin-bottom: 0.75rem;
}

.prose-content :deep(a) {
  color: #4daafc;
  text-decoration: underline;
}

.dark-mode .prose-content :deep(a) {
  color: #4daafc;
}

.flex-col.flex-1.h-full > div:first-child {
  --wails-draggable: drag;
}

.flex-col.flex-1.h-full > div:first-child > * {
  --wails-draggable: no-drag;
}
</style>
