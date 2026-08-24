<script setup lang="ts">
/* eslint-disable vue/no-v-html */
import { ref, computed, watch, onUnmounted } from 'vue';
import {
  PhTextAlignLeft,
  PhSpinnerGap,
  PhPlay,
  PhWarning,
  PhBrain,
  PhCaretDown,
  PhArrowsClockwise,
  PhCopy,
} from '@phosphor-icons/vue';
import { useI18n } from 'vue-i18n';

interface Props {
  summaryResult: {
    summary: string;
    html?: string;
    sentence_count: number;
    is_too_short: boolean;
    limit_reached?: boolean;
    used_fallback?: boolean;
    thinking?: string;
    error?: string;
  } | null;
  isLoadingSummary: boolean;
  translatedSummary?: {
    text: string;
    html?: string;
  } | null;
  isTranslatingSummary?: boolean;
  translationEnabled: boolean;
  summaryProvider?: string;
  summaryTriggerMode?: string;
  isLoadingContent?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  summaryProvider: 'local',
  summaryTriggerMode: 'auto',
  isLoadingContent: false,
  translatedSummary: null,
  isTranslatingSummary: false,
});

const emit = defineEmits<{
  'generate-summary': [];
}>();

const { t } = useI18n();

const showSummary = ref(true);
const showThinking = ref(false);
const isAnimating = ref(false);
const isCopying = ref(false);

// Enhanced loading states
const loadingTime = ref(0);
const loadingStartTime = ref<number | null>(null);

// Track loading time for better UX
watch(
  () => props.isLoadingSummary,
  (isLoading: boolean) => {
    if (isLoading && !loadingStartTime.value) {
      loadingStartTime.value = Date.now();
      loadingTime.value = 0;
    } else if (!isLoading && loadingStartTime.value) {
      loadingStartTime.value = null;
      loadingTime.value = 0;
    }
  }
);

let intervalId: number | null = null;

// Update loading time display
intervalId = window.setInterval(() => {
  if (props.isLoadingSummary && loadingStartTime.value) {
    loadingTime.value = Math.floor((Date.now() - loadingStartTime.value) / 1000);
  }
}, 100);

// Cleanup interval on unmount
onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId);
  }
});

// Check if should show manual trigger button
const shouldShowManualTrigger = computed(() => {
  return (
    props.summaryProvider === 'ai' &&
    props.summaryTriggerMode === 'manual' &&
    !props.summaryResult &&
    !props.isLoadingSummary &&
    !props.isLoadingContent
  );
});

const shouldReserveAutoSummary = computed(() => {
  if (props.summaryResult || props.isLoadingSummary || props.isLoadingContent) return false;
  if (props.summaryProvider === 'ai') return props.summaryTriggerMode === 'auto';
  return props.summaryProvider === 'local' || props.summaryProvider === 'rss';
});

const shouldShowSummaryContainer = computed(() => {
  return (
    !!props.summaryResult ||
    props.isLoadingSummary ||
    shouldShowManualTrigger.value ||
    shouldReserveAutoSummary.value
  );
});

const shouldShowTranslatedSummary = computed(() => {
  return Boolean(
    props.translationEnabled &&
    props.summaryProvider === 'rss' &&
    props.translatedSummary?.text &&
    props.translatedSummary.text !== props.summaryResult?.summary
  );
});

// Generate summary on button click
function handleGenerateSummary() {
  emit('generate-summary');
}

// Toggle summary with animation
function toggleSummary() {
  isAnimating.value = true;
  showSummary.value = !showSummary.value;
  setTimeout(() => {
    isAnimating.value = false;
  }, 300);
}

// Copy summary to clipboard
async function copySummary() {
  if (!props.summaryResult?.summary) return;

  isCopying.value = true;
  try {
    await navigator.clipboard.writeText(props.summaryResult.summary);
    window.showToast(t('common.toast.copiedToClipboard'), 'success');
  } catch (error) {
    console.error('Failed to copy summary:', error);
    window.showToast(t('common.errors.failedToCopy'), 'error');
  } finally {
    setTimeout(() => {
      isCopying.value = false;
    }, 500);
  }
}

// Handle link clicks in summary to open in default browser
async function handleSummaryLinkClick(event: MouseEvent) {
  const target = event.target as HTMLElement;
  const anchor = target.closest('a');

  if (anchor && anchor.href) {
    event.preventDefault();
    event.stopPropagation();

    try {
      await fetch('/api/browser/open', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: anchor.href }),
      });
    } catch (error) {
      console.error('Failed to open link:', error);
      window.showToast(t('common.errors.failedToOpenLink'), 'error');
    }
  }
}
</script>

<template>
  <!-- Summary Section -->
  <div
    v-if="shouldShowSummaryContainer"
    class="summary-container mb-4 p-3 rounded-lg border border-border bg-bg-secondary"
  >
    <!-- Summary Header -->
    <div
      class="flex items-center justify-between gap-2 cursor-pointer select-none"
      @click="toggleSummary"
    >
      <div class="flex items-center gap-2">
        <PhTextAlignLeft :size="20" class="text-accent" />
        <span class="text-base font-medium text-text-primary">{{
          t('article.summary.articleSummary')
        }}</span>
      </div>
      <div class="flex items-center gap-1">
        <!-- Copy Button (show when summary exists and expanded) -->
        <button
          v-if="summaryResult?.summary && showSummary"
          class="p-1.5 rounded hover:bg-bg-tertiary text-text-secondary hover:text-text-primary transition-colors"
          :title="t('common.copy')"
          @click.stop="copySummary"
        >
          <PhCopy :size="18" />
        </button>
        <!-- Regenerate Button -->
        <button
          v-if="summaryResult || shouldShowManualTrigger"
          class="p-1.5 rounded hover:bg-bg-tertiary text-text-secondary hover:text-text-primary transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          :title="t('setting.content.regenerateSummary')"
          :disabled="isLoadingSummary"
          @click.stop="handleGenerateSummary"
        >
          <PhSpinnerGap v-if="isLoadingSummary" :size="18" class="animate-spin" />
          <PhArrowsClockwise v-else :size="18" />
        </button>
        <!-- Toggle Icon -->
        <PhCaretDown
          :size="20"
          class="text-text-secondary transition-transform duration-200"
          :class="{ 'rotate-180': showSummary }"
        />
      </div>
    </div>

    <!-- Summary Content -->
    <Transition name="summary-content">
      <div v-if="showSummary" class="summary-content mt-3">
        <!-- Loading State -->
        <div
          v-if="isLoadingSummary || shouldReserveAutoSummary"
          class="summary-loading-state flex flex-col items-center justify-center gap-3 py-4"
        >
          <PhSpinnerGap :size="24" class="animate-spin text-accent" />
          <div class="text-sm text-text-primary">
            {{
              props.summaryProvider === 'ai'
                ? t('setting.content.generatingAISummary')
                : t('setting.content.generatingSummary')
            }}
          </div>
          <div class="text-xs text-text-secondary">
            {{ t('article.summary.generatingSummaryTime', { seconds: loadingTime }) }}
          </div>
        </div>

        <!-- Manual Trigger Button -->
        <div v-else-if="shouldShowManualTrigger" class="flex flex-col items-center gap-2 py-4">
          <div class="text-sm text-text-secondary text-center">
            {{ t('setting.content.summaryManualTriggerDesc') }}
          </div>
          <button
            class="flex items-center gap-2 px-4 py-2 bg-accent text-white rounded-lg hover:bg-accent/90 transition-colors"
            @click.stop="handleGenerateSummary"
          >
            <PhPlay :size="16" />
            <span class="text-sm">{{ t('setting.content.generateSummary') }}</span>
          </button>
        </div>

        <!-- Too Short Warning -->
        <div v-else-if="summaryResult?.is_too_short" class="flex flex-col items-center gap-2 py-4">
          <div class="flex items-center gap-2 text-amber-600 dark:text-amber-400">
            <PhWarning :size="18" />
            <span class="text-sm">{{ t('setting.content.summaryTooShort') }}</span>
          </div>
          <div class="text-xs text-text-secondary">{{ t('article.summary.articleTooShort') }}</div>
        </div>

        <!-- Summary Display -->
        <div v-else-if="summaryResult?.summary" class="summary-display">
          <!-- Thinking Toggle -->
          <button
            v-if="summaryResult.thinking"
            class="flex items-center gap-1.5 px-2 py-1 mb-2 text-xs bg-bg-tertiary/50 text-text-secondary rounded hover:bg-bg-tertiary transition-colors"
            @click.stop="showThinking = !showThinking"
          >
            <PhBrain :size="12" />
            <span>{{
              showThinking ? t('article.chat.hideThinking') : t('article.chat.showThinking')
            }}</span>
          </button>

          <!-- Thinking Section -->
          <Transition name="thinking-section">
            <div
              v-if="summaryResult.thinking && showThinking"
              class="mb-2 p-2 text-xs bg-bg-tertiary/50 rounded"
            >
              {{ summaryResult.thinking }}
            </div>
          </Transition>

          <!-- Summary Content -->
          <div v-if="shouldShowTranslatedSummary || isTranslatingSummary" class="mb-3">
            <div class="mb-1 text-[11px] text-text-secondary">
              {{ t('article.summary.translatedSummary') }}
            </div>
            <div
              v-if="shouldShowTranslatedSummary"
              class="text-xs text-text-primary leading-snug select-text prose prose-xs max-w-none"
              @click="handleSummaryLinkClick"
              v-html="translatedSummary?.html || translatedSummary?.text"
            ></div>
            <div v-else class="flex items-center gap-1 text-xs text-text-secondary">
              <PhSpinnerGap :size="12" class="animate-spin" />
              <span>{{ t('article.summary.translatingSummary') }}</span>
            </div>
          </div>

          <div v-if="shouldShowTranslatedSummary" class="mt-1 text-[11px] text-text-secondary">
            {{ t('article.summary.originalSummary') }}
          </div>
          <div
            class="text-xs text-text-primary leading-snug select-text prose prose-xs max-w-none"
            @click="handleSummaryLinkClick"
            v-html="summaryResult.html || summaryResult.summary"
          ></div>
        </div>

        <!-- Error State -->
        <div v-else-if="summaryResult?.error" class="flex flex-col items-center gap-2 py-4">
          <div class="flex items-center gap-2 text-red-500">
            <PhWarning :size="18" />
            <span class="text-sm">{{ t('setting.content.summaryGenerationFailed') }}</span>
          </div>
          <div class="text-xs text-text-secondary text-center max-w-xs break-words">
            {{ summaryResult.error }}
          </div>
        </div>

        <!-- No Summary Available -->
        <div v-else class="py-4 text-sm text-text-secondary text-center">
          {{ t('setting.content.noSummaryAvailable') }}
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.summary-container {
  min-height: 8rem;
  transition:
    min-height 0.2s ease,
    background-color 0.2s ease,
    border-color 0.2s ease;
}

.summary-loading-state {
  min-height: 4.75rem;
}

/* Content Transitions */
.summary-content-enter-active,
.summary-content-leave-active {
  transition: all 0.2s ease;
  max-height: 1000px;
  opacity: 1;
}

.summary-content-enter-from,
.summary-content-leave-to {
  max-height: 0;
  opacity: 0;
}

/* Thinking section transition */
.thinking-section-enter-active,
.thinking-section-leave-active {
  transition: all 0.2s ease;
  max-height: 500px;
  opacity: 1;
}

.thinking-section-enter-from,
.thinking-section-leave-to {
  max-height: 0;
  opacity: 0;
}

/* Prose Styles */
.prose {
  color: inherit;
}

/* Fix for dark mode: ensure prose inherits text color correctly */
.dark-mode .summary-display .prose {
  color: var(--text-primary);
}

.dark-mode .summary-display .prose p,
.dark-mode .summary-display .prose li,
.dark-mode .summary-display .prose span {
  color: var(--text-primary);
}

.prose.select-text,
.prose.select-text * {
  user-select: text !important;
  -webkit-user-select: text !important;
}

.prose p {
  margin: 0.25rem 0;
  line-height: 1.5;
}

.prose ul {
  list-style-type: disc;
  padding-left: 1.25rem;
  margin: 0.25rem 0;
}

.prose ol {
  list-style-type: decimal;
  padding-left: 1.25rem;
  margin: 0.25rem 0;
}

.prose li {
  margin: 0.125rem 0;
}

.prose code {
  font-family: 'Courier New', Courier, monospace;
  font-size: 0.875em;
  background-color: var(--bg-tertiary);
  padding: 0.125rem 0.25rem;
  border-radius: 0.25rem;
}

.prose a {
  color: var(--accent-color);
  text-decoration: none;
  cursor: pointer;
}

.prose a:hover {
  text-decoration: underline;
}
</style>
