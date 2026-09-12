<script setup lang="ts">
/* eslint-disable vue/no-v-html */
import { PhSpinnerGap, PhArrowSquareOut } from '@phosphor-icons/vue';
import { useI18n } from 'vue-i18n';
import type { FeedPreview } from '@/composables/feed/useFeedPreview';
import { useArticleDateFormat } from '@/composables/article/useArticleDateFormat';
import { openInBrowser } from '@/utils/browser';

defineProps<{ preview: FeedPreview | null; loading: boolean; failed: boolean }>();
defineEmits<{ retry: [] }>();
const { t } = useI18n();
const { formatArticleDateTime } = useArticleDateFormat();

function openLink(raw: string) {
  try {
    const url = new URL(raw);
    if (url.protocol === 'https:' || url.protocol === 'http:') void openInBrowser(url.href);
  } catch {
    /* Invalid source links are not navigable. */
  }
}
function handleContentClick(event: MouseEvent) {
  const link = event.target instanceof Element ? event.target.closest('a') : null;
  if (link) {
    event.preventDefault();
    openLink(link.getAttribute('href') || '');
  }
}
</script>

<template>
  <div class="p-4 sm:p-6 space-y-4" :aria-busy="loading">
    <p class="text-sm text-text-secondary">{{ t('modal.feed.previewHint') }}</p>
    <div v-if="loading" class="flex justify-center items-center gap-2 p-8 text-text-secondary">
      <PhSpinnerGap :size="22" class="animate-spin" />{{ t('modal.feed.previewLoading') }}
    </div>
    <div v-else-if="failed" role="alert" class="text-sm text-text-secondary space-y-2">
      <p>{{ t('modal.feed.previewFailed') }}</p>
      <button type="button" class="text-accent hover:underline" @click="$emit('retry')">
        {{ t('modal.feed.previewRetry') }}
      </button>
    </div>
    <template v-else-if="preview">
      <h2 class="font-semibold text-lg text-text-primary">{{ preview.title }}</h2>
      <p
        v-if="preview.description"
        class="text-sm text-text-secondary whitespace-pre-wrap break-words"
      >
        {{ preview.description }}
      </p>
      <p class="text-xs text-text-secondary">
        {{ t('modal.feed.previewCount', { count: preview.articles.length, total: preview.total }) }}
      </p>
      <p v-if="!preview.articles.length" class="text-sm text-text-secondary">
        {{ t('modal.feed.previewEmpty') }}
      </p>
      <details
        v-for="(article, index) in preview.articles"
        :key="index"
        :open="index === 0"
        class="border border-border rounded-lg bg-bg-secondary"
      >
        <summary class="p-3 cursor-pointer text-text-primary font-medium break-words">
          {{ article.title || t('modal.feed.previewUntitled') }}
        </summary>
        <div class="px-3 pb-3 space-y-3">
          <div class="flex flex-wrap items-center gap-3 text-xs text-text-secondary">
            <time v-if="article.published_at">{{
              formatArticleDateTime(article.published_at)
            }}</time>
            <button
              v-if="article.url"
              type="button"
              class="inline-flex items-center gap-1 text-accent"
              @click="openLink(article.url)"
            >
              <PhArrowSquareOut :size="14" />{{ t('article.action.openInBrowser') }}
            </button>
          </div>
          <!-- content_html has passed PrepareArticleContent on the backend. -->
          <div
            v-if="article.content_html"
            class="preview-content text-sm text-text-primary break-words"
            @click="handleContentClick"
            v-html="article.content_html"
          />
          <p v-else class="text-sm text-text-secondary">{{ t('modal.feed.previewNoContent') }}</p>
          <p v-if="article.truncated" class="text-xs text-text-secondary">
            {{ t('modal.feed.previewTruncated') }}
          </p>
        </div>
      </details>
    </template>
  </div>
</template>

<style scoped>
@reference "../../../style.css";
.preview-content :deep(img),
.preview-content :deep(video),
.preview-content :deep(iframe) {
  max-width: 100%;
  height: auto;
}
.preview-content :deep(pre) {
  overflow-x: auto;
  white-space: pre;
}
.preview-content :deep(p),
.preview-content :deep(ul),
.preview-content :deep(ol),
.preview-content :deep(pre) {
  margin-block: 0.75em;
}
.preview-content :deep(a) {
  @apply text-accent underline;
}
.preview-content :deep(h1),
.preview-content :deep(h2),
.preview-content :deep(h3) {
  font-weight: 600;
  margin-block: 1em 0.5em;
}
.preview-content :deep(ul) {
  list-style: disc;
  padding-left: 1.5em;
}
.preview-content :deep(ol) {
  list-style: decimal;
  padding-left: 1.5em;
}
.preview-content :deep(table) {
  display: block;
  max-width: 100%;
  overflow-x: auto;
}
</style>
