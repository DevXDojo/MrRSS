<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAppStore } from '@/stores/app';
import { PhCaretLeft, PhCaretRight } from '@phosphor-icons/vue';
import ArticleToolbar from './ArticleToolbar.vue';
import ArticleContent from './ArticleContent.vue';
import ImageViewer from '../common/ImageViewer.vue';
import FindInPage from '../common/FindInPage.vue';
import type { Article } from '@/types/models';
import { openInBrowser } from '@/utils/browser';
import { useSettings } from '@/composables/core/useSettings';

interface Props {
  article: Article;
  articleContent: string;
  isLoadingContent: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  previous: [];
  next: [];
  toggleRead: [];
  toggleFavorite: [];
  toggleReadLater: [];
  retryLoadContent: [];
}>();

const { t } = useI18n();
const store = useAppStore();
const { settings, fetchSettings } = useSettings();

// View state
const showContent = ref(true);
const showTranslations = ref(true);
const showFindInPage = ref(false);

// Image viewer state
const imageViewerSrc = ref<string | null>(null);
const imageViewerAlt = ref('');
const imageViewerImages = ref<string[]>([]);
const imageViewerInitialIndex = ref(0);

// Export to Obsidian
async function exportToObsidian() {
  if (!props.article) return;

  try {
    window.showToast(t('setting.plugins.obsidian.exporting'), 'info');

    const response = await fetch('/api/articles/export/obsidian', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        article_id: props.article.id,
      }),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(error);
    }

    const data = await response.json();

    // Show success message with file path
    const message = data.message || t('setting.plugins.obsidian.exported');
    const filePath = data.file_path ? ` (${data.file_path})` : '';
    window.showToast(message + filePath, 'success');
  } catch (error) {
    console.error('Failed to export to Obsidian:', error);
    const message =
      error instanceof Error ? error.message : t('setting.plugins.obsidian.exportFailed');
    window.showToast(message, 'error');
  }
}

// Navigation
const currentArticleIndex = computed(() => {
  if (!props.article) return -1;
  return store.articles.findIndex((a) => a.id === props.article.id);
});

const hasPreviousArticle = computed(() => currentArticleIndex.value > 0);
const hasNextArticle = computed(
  () => currentArticleIndex.value >= 0 && currentArticleIndex.value < store.articles.length - 1
);

// Load default view mode on mount
onMounted(async () => {
  try {
    await fetchSettings();
    // Apply default view mode
    showContent.value = settings.value.default_view_mode === 'rendered';
  } catch (e) {
    console.error('Error loading settings:', e);
  }

  // Add keyboard listener
  window.addEventListener('keydown', handleKeydown);
});

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown);
});

// Watch for article changes
watch(
  () => props.article?.id,
  () => {
    // Reset image viewer when article changes
    imageViewerSrc.value = null;
    imageViewerAlt.value = '';
    imageViewerImages.value = [];
    imageViewerInitialIndex.value = 0;

    // Apply default view mode for new article
    const feed = store.feeds.find((f) => f.id === props.article?.feed_id);
    if (feed?.article_view_mode === 'webpage') {
      showContent.value = false;
    } else if (feed?.article_view_mode === 'rendered') {
      showContent.value = true;
    } else {
      showContent.value = settings.value.default_view_mode === 'rendered';
    }
  }
);

function handleKeydown(e: KeyboardEvent) {
  // ESC to close
  if (e.key === 'Escape') {
    if (showFindInPage.value) {
      showFindInPage.value = false;
    } else if (imageViewerSrc.value) {
      closeImageViewer();
    } else {
      emit('close');
    }
    return;
  }

  // Ctrl+F to find
  if ((e.ctrlKey || e.metaKey) && e.key === 'f' && showContent.value) {
    e.preventDefault();
    showFindInPage.value = true;
    return;
  }

  // Arrow navigation
  if (e.key === 'ArrowLeft' && hasPreviousArticle.value) {
    emit('previous');
  } else if (e.key === 'ArrowRight' && hasNextArticle.value) {
    emit('next');
  }
}

function toggleContentView() {
  showContent.value = !showContent.value;
}

function toggleTranslations() {
  showTranslations.value = !showTranslations.value;
}

function openOriginal() {
  if (props.article?.url) {
    openInBrowser(props.article.url);
  }
}

function closeImageViewer() {
  imageViewerSrc.value = null;
  imageViewerAlt.value = '';
  imageViewerImages.value = [];
  imageViewerInitialIndex.value = 0;
}

// Attach image event listeners for the image viewer
function attachImageEventListeners() {
  setTimeout(() => {
    const contentEl = document.querySelector('.modal-prose-content');
    if (!contentEl) return;

    const images = contentEl.querySelectorAll('img');
    const allImages: string[] = [];

    images.forEach((img) => {
      const src = img.getAttribute('src');
      if (src) {
        allImages.push(src);
      }
    });

    images.forEach((img, index) => {
      img.style.cursor = 'pointer';
      img.addEventListener('click', () => {
        const src = img.getAttribute('src');
        if (src) {
          imageViewerSrc.value = src;
          imageViewerAlt.value = img.getAttribute('alt') || '';
          imageViewerImages.value = allImages;
          imageViewerInitialIndex.value = index;
        }
      });
    });
  }, 100);
}

function handleRetryLoadContent() {
  emit('retryLoadContent');
}

function handleOverlayClick(e: MouseEvent) {
  // Only close if clicking directly on the overlay, not its children
  if (e.target === e.currentTarget) {
    emit('close');
  }
}
</script>

<template>
  <Teleport to="body">
    <div class="article-modal-overlay" @click="handleOverlayClick">
      <div class="article-modal" @click.stop>
        <!-- Reuse ArticleToolbar with modal mode -->
        <ArticleToolbar
          :article="article"
          :show-content="showContent"
          :show-translations="showTranslations"
          :is-modal="true"
          @close="emit('close')"
          @toggle-content-view="toggleContentView"
          @toggle-read="emit('toggleRead')"
          @toggle-favorite="emit('toggleFavorite')"
          @toggle-read-later="emit('toggleReadLater')"
          @open-original="openOriginal"
          @toggle-translations="toggleTranslations"
          @export-to-obsidian="exportToObsidian"
        />

        <!-- Modal content -->
        <div class="modal-content">
          <!-- Original webpage view -->
          <div v-if="!showContent" class="flex-1 bg-bg-primary w-full">
            <iframe
              :key="article.id"
              :src="`/api/webpage/proxy?url=${encodeURIComponent(article.url)}`"
              class="w-full h-full border-none"
              sandbox="allow-scripts allow-same-origin allow-popups"
            ></iframe>
          </div>

          <!-- RSS content view -->
          <ArticleContent
            v-else
            :article="article"
            :article-content="articleContent"
            :is-loading-content="isLoadingContent"
            :attach-image-event-listeners="attachImageEventListeners"
            :show-translations="showTranslations"
            :show-content="showContent"
            class="modal-prose-content"
            @retry-load-content="handleRetryLoadContent"
          />
        </div>

        <!-- Navigation buttons -->
        <div v-if="hasPreviousArticle || hasNextArticle" class="modal-navigation">
          <button
            v-if="hasPreviousArticle"
            class="nav-btn"
            :title="t('article.navigation.previousArticle')"
            @click="emit('previous')"
          >
            <PhCaretLeft :size="16" />
            <span>{{ t('article.navigation.previousArticle') }}</span>
          </button>
          <div v-else class="w-24"></div>

          <button
            v-if="hasNextArticle"
            class="nav-btn"
            :title="t('article.navigation.nextArticle')"
            @click="emit('next')"
          >
            <span>{{ t('article.navigation.nextArticle') }}</span>
            <PhCaretRight :size="16" />
          </button>
          <div v-else class="w-24"></div>
        </div>
      </div>

      <!-- Find in Page -->
      <FindInPage
        v-if="showFindInPage && showContent"
        container-selector=".modal-prose-content"
        :article-id="article?.id"
        @close="showFindInPage = false"
      />

      <!-- Image Viewer -->
      <ImageViewer
        v-if="imageViewerSrc"
        :src="imageViewerSrc"
        :alt="imageViewerAlt"
        :images="imageViewerImages"
        :initial-index="imageViewerInitialIndex"
        @close="closeImageViewer"
      />
    </div>
  </Teleport>
</template>

<style scoped>
@reference "../../style.css";

.article-modal-overlay {
  @apply fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4;
}

.article-modal {
  @apply bg-bg-primary w-full max-w-4xl h-[90vh] rounded-2xl shadow-2xl border border-border flex flex-col overflow-hidden;
}

.modal-content {
  @apply flex-1 overflow-hidden flex flex-col;
}

.modal-navigation {
  @apply flex items-center justify-between px-3 py-2 border-t border-border bg-bg-primary;
}

.nav-btn {
  @apply flex items-center gap-1.5 px-3 py-1.5 rounded text-sm text-text-secondary hover:text-text-primary hover:bg-bg-secondary transition-colors;
}

/* Override ArticleContent styling inside modal */
:deep(.modal-prose-content) {
  @apply flex-1 overflow-auto;
}
</style>
