<script setup lang="ts">
import { PhNewspaper, PhCaretLeft, PhCaretRight } from '@phosphor-icons/vue';
import { useArticleDetail } from '@/composables/article/useArticleDetail';
import ArticleToolbar from './ArticleToolbar.vue';
import ArticleContent from './ArticleContent.vue';
import ImageViewer from '../common/ImageViewer.vue';
import FindInPage from '../common/FindInPage.vue';

import { ref, onMounted, onBeforeUnmount } from 'vue';

const {
  article,
  showContent,
  articleContent,
  isLoadingContent,
  imageViewerSrc,
  imageViewerAlt,
  imageViewerImages,
  imageViewerInitialIndex,
  hasPreviousArticle,
  hasNextArticle,
  close,
  toggleRead,
  toggleFavorite,
  toggleReadLater,
  openOriginal,
  toggleContentView,
  reloadArticleContent,
  closeImageViewer,
  attachImageEventListeners,
  exportToObsidian,
  exportToNotion,
  exportToZotero,
  handleRetryLoadContent,
  goToPreviousArticle,
  goToNextArticle,
  t,
} = useArticleDetail();

const showTranslations = ref(true);
const showFindInPage = ref(false);

function toggleTranslations() {
  showTranslations.value = !showTranslations.value;
}

function openFindInPage() {
  showFindInPage.value = true;
}

function closeFindInPage() {
  showFindInPage.value = false;
}

function handleKeydown(e: KeyboardEvent) {
  // Open find in page with Ctrl+F or Cmd+F
  if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
    // Only if we're showing an article in content mode (not webpage view)
    if (article.value && showContent.value) {
      e.preventDefault();
      openFindInPage();
    }
  }

  // Note: FindInPage component handles its own ESC key to close
  // We don't handle ESC here to avoid conflicts - FindInPage will stopPropagation
  // when it needs to handle the key (when search is focused or has content)

  // Note: Arrow key navigation is now handled by the global keyboard shortcuts system
  // See useKeyboardShortcuts.ts which properly checks for editable elements
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown);
});

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
  <main
    :class="[
      'flex-1 bg-bg-primary flex flex-col h-full absolute w-full md:static md:w-auto z-30 transition-transform duration-300',
      article ? 'translate-x-0' : 'translate-x-full md:translate-x-0',
    ]"
  >
    <div
      v-if="!article"
      class="hidden md:flex flex-col items-center justify-center h-full text-text-secondary text-center px-4"
    >
      <PhNewspaper :size="48" class="mb-4 sm:mb-5 opacity-50 sm:w-16 sm:h-16" />
      <p class="text-sm sm:text-base">{{ t('article.content.selectArticle') }}</p>
    </div>

    <div v-else class="flex flex-col h-full bg-bg-primary">
      <ArticleToolbar
        :article="article"
        :show-content="showContent"
        :show-translations="showTranslations"
        @close="close"
        @toggle-content-view="toggleContentView"
        @toggle-read="toggleRead"
        @toggle-favorite="toggleFavorite"
        @toggle-read-later="toggleReadLater"
        @open-original="openOriginal"
        @toggle-translations="toggleTranslations"
        @reload-content="reloadArticleContent"
        @export-to-obsidian="exportToObsidian"
        @export-to-notion="exportToNotion"
        @export-to-zotero="exportToZotero"
      />

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
        @retry-load-content="handleRetryLoadContent"
      />

      <!-- Navigation buttons -->
      <div
        v-if="hasPreviousArticle || hasNextArticle"
        class="flex items-center justify-between bg-bg-primary px-3 py-1.5"
      >
        <button
          v-if="hasPreviousArticle"
          :title="t('article.navigation.previousArticle') || 'Previous article'"
          class="flex items-center gap-1.5 px-2 py-1 rounded text-text-secondary/70 hover:text-text-primary hover:bg-bg-secondary/50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          @click="goToPreviousArticle"
        >
          <PhCaretLeft :size="16" />
          <span class="text-xs">{{ t('article.navigation.previousArticle') || 'Previous' }}</span>
        </button>

        <div v-else class="w-16"></div>

        <button
          v-if="hasNextArticle"
          :title="t('article.navigation.nextArticle') || 'Next article'"
          class="flex items-center gap-1.5 px-2 py-1 rounded text-text-secondary/70 hover:text-text-primary hover:bg-bg-secondary/50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          @click="goToNextArticle"
        >
          <span class="text-xs">{{ t('article.navigation.nextArticle') || 'Next' }}</span>
          <PhCaretRight :size="16" />
        </button>

        <div v-else class="w-16"></div>
      </div>
    </div>

    <!-- Find in Page (only shown in content mode) -->
    <FindInPage
      v-if="showFindInPage && showContent"
      container-selector=".prose-content"
      :article-id="article?.id"
      @close="closeFindInPage"
    />

    <!-- Image Viewer Modal -->
    <ImageViewer
      v-if="imageViewerSrc"
      :src="imageViewerSrc"
      :alt="imageViewerAlt"
      :images="imageViewerImages"
      :initial-index="imageViewerInitialIndex"
      @close="closeImageViewer"
    />
  </main>
</template>
