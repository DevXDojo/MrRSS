<script setup lang="ts">
/* eslint-disable vue/no-v-html */
import { computed } from 'vue';
import { PhSpinnerGap, PhArticle, PhArrowClockwise } from '@phosphor-icons/vue';
import { useI18n } from 'vue-i18n';
import { useSettings } from '@/composables/core/useSettings';
import { resolveFontFamily } from '@/utils/fontDetector';

const { t } = useI18n();
const { settings } = useSettings();

interface Props {
  articleContent: string;
  isTranslatingContent: boolean;
  hasMediaContent?: boolean; // Whether article has audio/video content
  isLoadingContent?: boolean; // Whether content is currently loading
}

const props = withDefaults(defineProps<Props>(), {
  hasMediaContent: false,
  isLoadingContent: false,
});

// Emits
const emit = defineEmits<{
  retryLoad: [];
}>();

const hasCustomCSS = computed(() => !!settings.value.custom_css_file);

// Content styling based on settings
const contentStyle = computed(() => {
  const fontFamily = settings.value.content_font_family;
  const fontSize = parseInt(settings.value.content_font_size as any) || 16;
  const lineHeight = settings.value.content_line_height || '1.6';

  const style = {
    fontFamily: resolveFontFamily(fontFamily),
    fontSize: `${fontSize}px`, // Always apply font size
    lineHeight: lineHeight,
    '--content-line-height': lineHeight,
  } as Record<string, string>;

  return style;
});

</script>

<template>
  <!-- Content display with inline translations -->
  <div v-if="articleContent">
    <div
      class="prose prose-sm sm:prose-lg max-w-none text-text-primary prose-content"
      :class="{ 'custom-css-active': hasCustomCSS }"
      :style="contentStyle"
      v-html="articleContent"
    ></div>
    <!-- Translation loading indicator -->
    <div v-if="isTranslatingContent" class="flex items-center gap-2 mt-4 text-text-secondary">
      <PhSpinnerGap :size="16" class="animate-spin" />
      <span class="text-sm">{{ t('setting.content.translatingContent') }}</span>
    </div>
  </div>

  <!-- Content loading state -->
  <div
    v-else-if="props.isLoadingContent && !hasMediaContent"
    class="flex flex-col items-center justify-center text-center text-text-secondary py-6 sm:py-8"
  >
    <PhSpinnerGap :size="36" class="mb-2 sm:mb-3 animate-spin opacity-70" />
    <p class="text-sm sm:text-base">{{ t('article.content.fetchingArticleContent') }}</p>
  </div>

  <!-- No content available with retry option -->
  <div v-else-if="!hasMediaContent" class="text-center text-text-secondary py-6 sm:py-8">
    <PhArticle :size="48" class="mb-2 sm:mb-3 opacity-50 mx-auto sm:w-16 sm:h-16" />
    <p class="text-sm sm:text-base mb-4">{{ t('article.content.noContentAvailable') }}</p>
    <button
      v-if="!props.isLoadingContent"
      class="btn-secondary-compact flex items-center gap-1.5 mx-auto"
      @click="emit('retryLoad')"
    >
      <PhArrowClockwise :size="12" />
      <span class="text-xs">{{ t('setting.content.retrySummary') }}</span>
    </button>
  </div>
</template>
