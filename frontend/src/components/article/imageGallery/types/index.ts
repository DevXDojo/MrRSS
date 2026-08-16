import type { Article } from '@/types/models';

/**
 * Image gallery configuration constants
 */
interface ImageGalleryConfig {
  /** Number of items to fetch per page */
  itemsPerPage: number;
  /** Distance from bottom (in pixels) to trigger infinite scroll */
  scrollThreshold: number;
  /** Minimum zoom scale */
  minScale: number;
  /** Maximum zoom scale */
  maxScale: number;
  /** Zoom step per increment */
  scaleStep: number;
  /** Target width for each column in pixels */
  targetColumnWidth: number;
  /** Minimum number of columns */
  minColumns: number;
}

/**
 * Image viewer state
 */
interface ImageViewerState {
  /** Currently selected article */
  article: Article | null;
  /** All images from the current article */
  images: string[];
  /** Index of currently displayed image */
  currentIndex: number;
  /** Current zoom scale */
  scale: number;
  /** Current pan position */
  position: { x: number; y: number };
  /** Whether user is currently dragging */
  isDragging: boolean;
  /** Whether current image is loading */
  isLoading: boolean;
}

/**
 * Context menu action types for images
 */
type ImageAction =
  'toggleRead' | 'toggleFavorite' | 'copyTitle' | 'copyLink' | 'downloadImage' | 'openBrowser';

/**
 * Navigation direction for image viewer
 */
type NavigationDirection = 'prev' | 'next';

/**
 * Thumbnail strip configuration
 */
interface ThumbnailStripConfig {
  /** Width of each thumbnail in pixels (64px w-16 + 8px gap-2 = 72px) */
  thumbnailWidth: number;
  /** Threshold for collapsing thumbnail strip */
  collapseThreshold: number;
}

/**
 * Image gallery data return type from useImageGalleryData composable
 */
export interface ImageGalleryDataReturn {
  // State
  articles: import('vue').Ref<Article[]>;
  isLoading: import('vue').Ref<boolean>;
  page: import('vue').Ref<number>;
  hasMore: import('vue').Ref<boolean>;
  imageCountCache: import('vue').Ref<Map<number, number>>;
  showOnlyUnread: import('vue').Ref<boolean>;

  // Methods
  fetchImages: (loadMore?: boolean) => Promise<void>;
  fetchImageCount: (articleId: number) => Promise<void>;
  getImageCount: (article: Article) => number;
  refresh: () => Promise<void>;
  toggleShowOnlyUnread: () => void;
}

/**
 * Masonry layout return type from useMasonryLayout composable
 */
export interface MasonryLayoutReturn {
  columns: import('vue').Ref<Article[][]>;
  columnCount: import('vue').Ref<number>;
  containerRef: import('vue').Ref<HTMLElement | null>;
  calculateColumns: () => void;
  arrangeColumns: () => void;
  setupResizeObserver: () => void;
  cleanupResizeObserver: () => void;
}

/**
 * Image viewer return type from useImageViewer composable
 */
export interface ImageViewerReturn {
  // State
  scale: import('vue').Ref<number>;
  position: import('vue').Ref<{ x: number; y: number }>;
  isDragging: import('vue').Ref<boolean>;
  currentImageLoading: import('vue').Ref<boolean>;
  dragStart: import('vue').Ref<{ x: number; y: number }>;

  // Computed
  imageStyle: import('vue').ComputedRef<{ transform: string }>;
  canNavigatePrevious: import('vue').ComputedRef<boolean>;
  canNavigateNext: import('vue').ComputedRef<boolean>;

  // Methods
  zoomIn: () => void;
  zoomOut: () => void;
  resetView: () => void;
  startDrag: (e: MouseEvent) => void;
  onDrag: (e: MouseEvent) => void;
  stopDrag: () => void;
  handleImageWheel: (e: globalThis.WheelEvent) => void;
  handleImageLoad: () => void;
  handleImageError: () => void;
}

/**
 * Image actions return type from useImageActions composable
 */
export interface ImageActionsReturn {
  toggleFavorite: (article: Article, event?: Event) => Promise<void>;
  markAsRead: (article: Article) => Promise<void>;
  toggleReadStatus: (article: Article) => Promise<void>;
  downloadImage: (src: string) => Promise<void>;
  copyImage: (src: string) => Promise<void>;
  openOriginal: (article: Article) => void;
  copyArticleTitle: (article: Article) => Promise<void>;
  copyArticleLink: (article: Article) => Promise<void>;
}

/**
 * Gallery keyboard return type from useGalleryKeyboard composable
 */
export interface GalleryKeyboardReturn {
  enable: () => void;
  disable: () => void;
}

// Global type declarations
declare global {
  interface Window {
    store: any;
  }
}
