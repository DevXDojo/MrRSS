<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, computed, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhSpinnerGap, PhArticleNyTimes } from '@phosphor-icons/vue';
import type { Article } from '@/types/models';
import ArticleTitle from './parts/ArticleTitle.vue';
import ArticleSummary from './parts/ArticleSummary.vue';
import ArticleLoading from './parts/ArticleLoading.vue';
import ArticleBody from './parts/ArticleBody.vue';
import AudioPlayer from './parts/AudioPlayer.vue';
import VideoPlayer from './parts/VideoPlayer.vue';
import ArticleChatButton from './ArticleChatButton.vue';
import ArticleChatPanel from './ArticleChatPanel.vue';
import { useArticleSummary } from '@/composables/article/useArticleSummary';
import { useArticleTranslation } from '@/composables/article/useArticleTranslation';
import { useArticleRendering } from '@/composables/article/useArticleRendering';
import {
  extractTextWithPlaceholders,
  restorePreservedElements,
  hasOnlyPreservedContent,
} from '@/composables/article/useContentTranslation';
import { useSettings } from '@/composables/core/useSettings';
import { useAppStore } from '@/stores/app';
import { proxyImagesInHtml, isMediaCacheEnabled } from '@/utils/mediaProxy';
import './ArticleContent.css';

interface SummaryResult {
  summary: string;
  html?: string;
  sentence_count: number;
  is_too_short: boolean;
  limit_reached?: boolean;
  used_fallback?: boolean;
  thinking?: string;
  error?: string;
}

interface Props {
  article: Article;
  articleContent: string;
  isLoadingContent: boolean;
  attachImageEventListeners?: () => void;
  showTranslations?: boolean;
  showContent?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  showTranslations: true,
  attachImageEventListeners: undefined,
  showContent: true,
});

const emit = defineEmits<{
  retryLoadContent: [];
}>();

const { t } = useI18n();

// Handle retry load content
function handleRetryLoad() {
  emit('retryLoadContent');
}

// Chat state
const { settings: appSettings, fetchSettings } = useSettings();
const store = useAppStore();
const isChatPanelOpen = ref(false);

// Full-text fetching state
const isFetchingFullArticle = ref(false);
const fullArticleContent = ref('');
const autoShowAllContent = ref(false);

// Computed property to determine if auto-expand should be enabled for this feed
const shouldAutoExpandContent = computed(() => {
  // First check if feed has auto_expand_content setting
  const feed = store.feeds.find((f) => f.id === props.article.feed_id);

  // Special case: For XPath feeds without content xpath, always auto-expand regardless of settings
  const isXPathFeedWithoutContent =
    feed &&
    (feed.type === 'HTML+XPath' || feed.type === 'XML+XPath') &&
    !feed.xpath_item_content &&
    feed.xpath_item_uri;

  // For XPath feeds without content xpath, always return true
  if (isXPathFeedWithoutContent) {
    return true;
  }

  if (feed?.auto_expand_content) {
    if (feed.auto_expand_content === 'enabled') return true;
    if (feed.auto_expand_content === 'disabled') return false;
    // If 'global', fall through to global setting
  }

  // Fall back to global setting
  return autoShowAllContent.value;
});

// Fetch settings on mount to get actual values
onMounted(async () => {
  try {
    const data = await fetchSettings();
    autoShowAllContent.value = data.auto_show_all_content === true;
  } catch (e) {
    console.error('Error fetching settings for chat:', e);
  }

  // Listen for auto show all content setting changes
  window.addEventListener(
    'auto-show-all-content-changed',
    onAutoShowAllContentChanged as EventListener
  );

  // Listen for summary settings changes
  window.addEventListener('summary-settings-changed', onSummarySettingsChanged as EventListener);
});

// Computed to check if chat should be shown
const showChatButton = computed(() => {
  return (
    appSettings.value.ai_chat_enabled && !props.isLoadingContent && props.articleContent
    // Removed: props.showContent requirement - chat should work in both modes
  );
});

// Computed to check if full-text fetching should be shown
const showFullTextButton = computed(() => {
  // For XPath feeds without content, show button even if articleContent is empty
  const feed = store.feeds.find((f) => f.id === props.article.feed_id);
  const isXPathFeedWithoutContent =
    feed && (feed.type === 'HTML+XPath' || feed.type === 'XML+XPath') && !props.articleContent;

  return (
    appSettings.value.full_text_fetch_enabled &&
    !props.isLoadingContent &&
    (props.articleContent || isXPathFeedWithoutContent) && // Allow empty content for XPath feeds
    props.article?.url &&
    props.showContent &&
    !fullArticleContent.value // Don't show if we already have full content
  );
});

// Computed for the content to display (full article if available, otherwise RSS content)
const displayContent = computed(() => {
  return fullArticleContent.value || props.articleContent;
});

// Use composables for summary and translation
const {
  summarySettings,
  loadSummarySettings,
  generateSummary: generateSummaryComposable,
  isSummaryLoading,
  cancelSummaryGeneration,
} = useArticleSummary();

const { translationSettings, loadTranslationSettings } = useArticleTranslation();

// Use composable for enhanced rendering (math formulas, etc.)
const { enhanceRendering, renderMathFormulas, highlightCodeBlocks } = useArticleRendering();

// Computed properties for easier access
const summaryEnabled = computed(() => summarySettings.value.enabled);
const summaryProvider = computed(() => summarySettings.value.provider);
const summaryTriggerMode = computed(() => summarySettings.value.triggerMode);
const translationEnabled = computed(() => translationSettings.value.enabled);
const targetLanguage = computed(() => translationSettings.value.targetLang);

// Current article summary
const summaryResult = ref<SummaryResult | null>(null);
const isLoadingSummary = computed(() =>
  props.article ? isSummaryLoading(props.article.id) : false
);

// Additional state for translation
const translatedTitle = ref('');
const isTranslatingTitle = ref(false);
const isTranslatingContent = ref(false);
const lastTranslatedArticleId = ref<number | null>(null);
const lastTranslatedContentHash = ref<string>(''); // Track translated content by hash
const translationSkipped = ref(false);

// Load settings using composables
async function loadSettings() {
  await loadSummarySettings();
  await loadTranslationSettings();
}

// Translate text using the API
async function translateText(
  text: string,
  force: boolean = false
): Promise<{ text: string; html: string }> {
  if (!text || !translationEnabled.value) {
    return { text: '', html: '' };
  }

  const requestBody = {
    text: text,
    target_language: targetLanguage.value,
    force: force,
  };

  try {
    const res = await fetch('/api/articles/translate-text', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(requestBody),
    });

    if (res.ok) {
      const data = await res.json();

      // Check if translation was skipped
      if (data.skipped === 'true' || data.skipped === true) {
        if (data.reason === 'already_target_language') {
          translationSkipped.value = true;
        }
      } else {
        // Reset skip flags on successful translation
        translationSkipped.value = false;
      }

      return {
        text: data.translated_text || '',
        html: data.html || '',
      };
    } else {
      window.showToast(t('common.errors.translatingContent'), 'error');
    }
  } catch {
    window.showToast(t('common.errors.translating'), 'error');
  }
  return { text: '', html: '' };
}

// Force translate content
async function forceTranslateContent() {
  if (!props.articleContent) return;

  await translateContentParagraphs(props.articleContent);
}

// Fetch full article content from the original URL
// @param showErrors - whether to show error toasts (default: true for manual clicks, false for auto-fetch)
async function fetchFullArticle(showErrors: boolean = true) {
  if (!props.article?.id) return;

  isFetchingFullArticle.value = true;
  try {
    const res = await fetch(`/api/articles/fetch-full?id=${props.article.id}`, {
      method: 'POST',
    });

    if (res.ok) {
      const data = await res.json();
      let content = data.content || '';

      // Proxy images if media cache is enabled
      const cacheEnabled = await isMediaCacheEnabled();
      if (cacheEnabled && content) {
        // Use feed URL as referer for anti-hotlinking (more reliable than article URL)
        const feedUrl = data.feed_url || props.article.url;
        content = proxyImagesInHtml(content, feedUrl);
      }

      fullArticleContent.value = content;
      if (showErrors) {
        window.showToast(t('article.action.fullArticleFetched'), 'success');
      }

      // After fetching full content, regenerate summary and trigger translation
      if (props.article) {
        if (shouldAutoGenerateSummary()) {
          setTimeout(() => generateSummary(props.article), 100);
        }
        if (translationEnabled.value) {
          // Only translate content, not title (title translation is cached in DB)
          // Content hash will automatically detect new content and trigger translation
          // Wait for DOM to update with new content before translating
          await nextTick();
          await translateContentParagraphs(fullArticleContent.value);
        }
      }
    } else {
      console.error('Error fetching full article:', res.status);
      if (showErrors) {
        window.showToast(t('common.errors.fetchingFullArticle'), 'error');
      }
    }
  } catch (e) {
    console.error('Error fetching full article:', e);
    if (showErrors) {
      window.showToast(t('common.errors.fetchingFullArticle'), 'error');
    }
  } finally {
    isFetchingFullArticle.value = false;
  }
}

// Generate summary for the current article
async function generateSummary(article: Article, force: boolean = false) {
  if (!summaryEnabled.value || !article) {
    return;
  }

  // Only clear state if forcing regeneration
  if (force) {
    summaryResult.value = null;
  }

  const result = await generateSummaryComposable(article, displayContent.value, force);

  // Update the article summary in store for caching
  if (result?.summary) {
    store.updateArticleSummary(article.id, result.summary);
  }

  // Set summary result
  summaryResult.value = result;
}

// Check if should auto-generate summary
function shouldAutoGenerateSummary(): boolean {
  if (!summaryEnabled.value) return false;

  // For local provider, always auto-generate
  if (summaryProvider.value === 'local') return true;

  // For AI provider, check trigger mode
  if (summaryProvider.value === 'ai') {
    return summaryTriggerMode.value === 'auto';
  }

  return false;
}

// Translate title
async function translateTitle(article: Article) {
  if (!translationEnabled.value || !article?.title) return;

  isTranslatingTitle.value = true;
  const translation = await translateText(article.title);
  translatedTitle.value = translation.text;
  isTranslatingTitle.value = false;
}

// Simple hash function for content (for detecting content changes)
function simpleHash(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash = hash & hash; // Convert to 32bit integer
  }
  return hash.toString(36);
}

// Translate content paragraphs while preserving inline elements (formulas, code, images)
async function translateContentParagraphs(content: string) {
  if (!translationEnabled.value || !content) {
    return;
  }

  // Calculate content hash to detect if content has changed
  const contentHash = simpleHash(content);

  // Prevent duplicate translations for the same content
  // Check both article ID and content hash to handle RSS content vs full content
  if (
    lastTranslatedArticleId.value === props.article?.id &&
    lastTranslatedContentHash.value === contentHash
  ) {
    return;
  }

  isTranslatingContent.value = true;
  lastTranslatedArticleId.value = props.article?.id || null;
  lastTranslatedContentHash.value = contentHash;

  // Wait for content to render
  await nextTick();

  // Find all text elements in the prose content
  const proseContainer = document.querySelector('.prose-content');
  if (!proseContainer) {
    isTranslatingContent.value = false;
    return;
  }

  // Remove any existing translations first
  const existingTranslations = proseContainer.querySelectorAll('.translation-text');
  existingTranslations.forEach((el) => el.remove());

  // Check if content is plain text (no HTML tags) and wrap it in <p> tags
  // This handles cases where article content is stored as plain text without HTML structure
  const hasHTMLTags = /<[^>]+>/.test(proseContainer.innerHTML);
  if (!hasHTMLTags && proseContainer.textContent && proseContainer.textContent.trim().length > 0) {
    const textContent = proseContainer.innerHTML;
    proseContainer.innerHTML = `<p>${textContent}</p>`;
  }

  // Find all translatable elements
  // For lists: translate individual li items, translation stays inside the same li
  // For tables: translate td/th cells, translation stays inside the same cell
  // For blockquotes: translate inner paragraphs, not the blockquote itself
  const textTags = [
    'P',
    'H1',
    'H2',
    'H3',
    'H4',
    'H5',
    'H6',
    'LI',
    'TD',
    'TH',
    'FIGCAPTION',
    'DT',
    'DD',
  ];

  // Track which elements we've already translated to avoid duplicates
  const translatedElements = new Set<HTMLElement>();

  // Process elements level by level to handle nested structures correctly
  // First, get all elements and sort them by depth (shallowest first)
  const allElements = Array.from(proseContainer.querySelectorAll(textTags.join(',')));

  // Sort by depth (number of ancestors) to process outermost elements first
  allElements.sort((a, b) => {
    const getDepth = (el: Element): number => {
      let depth = 0;
      let parent = el.parentElement;
      while (parent && parent !== proseContainer) {
        depth++;
        parent = parent.parentElement;
      }
      return depth;
    };
    return getDepth(a) - getDepth(b);
  });

  // Helper function to check if an element can contain nested translatable content
  const canContainNestedTranslatableElements = (el: HTMLElement): boolean => {
    // These elements can contain other translatable elements
    const nestableTags = ['LI', 'BLOCKQUOTE', 'DD', 'DT', 'TD', 'TH'];
    return nestableTags.includes(el.tagName);
  };

  // Helper function to get nested translatable children (direct children only)
  const getNestedTranslatableChildren = (el: HTMLElement): Element[] => {
    return Array.from(el.children).filter((child) => textTags.includes(child.tagName));
  };

  for (const el of allElements) {
    const htmlEl = el as HTMLElement;

    // Skip if inside a translation element
    if (htmlEl.closest('.translation-text')) continue;

    // Skip if already has translation inside
    if (htmlEl.querySelector('.translation-text')) continue;

    // Skip if we've already translated this element
    if (translatedElements.has(htmlEl)) continue;

    // Skip if this element's parent or any ancestor was already translated
    // EXCEPTION: For LI/BLOCKQUOTE/DD/DT/TD/TH elements, if the parent element
    // also contains nested elements, allow translation
    let hasTranslatedAncestor = false;
    let ancestor = htmlEl.parentElement;
    while (ancestor && ancestor !== proseContainer) {
      if (translatedElements.has(ancestor)) {
        // If current element is also nestable (like LI), check if ancestor has nested children
        if (canContainNestedTranslatableElements(htmlEl)) {
          const ancestorNested = getNestedTranslatableChildren(ancestor as HTMLElement);
          if (ancestorNested.length > 0) {
            // Both are nestable and ancestor has nested children, allow this one
            ancestor = ancestor.parentElement;
            continue;
          }
        }
        hasTranslatedAncestor = true;
        break;
      }
      ancestor = ancestor.parentElement;
    }
    if (hasTranslatedAncestor) continue;

    // Skip elements that are entirely technical content (no translatable text)
    if (
      htmlEl.closest('pre') ||
      htmlEl.tagName === 'CODE' ||
      htmlEl.closest('kbd') ||
      htmlEl.classList.contains('katex') ||
      htmlEl.classList.contains('katex-display') ||
      htmlEl.classList.contains('katex-inline')
    ) {
      continue;
    }

    // Skip elements that only contain preserved content (no translatable text)
    if (hasOnlyPreservedContent(htmlEl)) {
      continue;
    }

    // Extract text with placeholders for inline elements (formulas, code, images) and hyperlinks
    const {
      text: textWithPlaceholders,
      preservedElements,
      hyperlinks,
    } = extractTextWithPlaceholders(htmlEl);

    if (!textWithPlaceholders || textWithPlaceholders.length < 2) continue;

    // Translate the text (with placeholders and link markers)
    const translation = await translateText(textWithPlaceholders);
    const translatedText = translation.text;

    // Skip if translation is same as original or empty
    if (!translatedText || translatedText === textWithPlaceholders) {
      continue;
    }

    // Restore preserved elements and hyperlinks in the translated text
    const translatedHTML = restorePreservedElements(translatedText, preservedElements, hyperlinks);

    // Determine how to insert translation based on element type
    const tagName = htmlEl.tagName;

    if (
      tagName === 'LI' ||
      tagName === 'TD' ||
      tagName === 'TH' ||
      tagName === 'DD' ||
      tagName === 'DT'
    ) {
      // For list items, table cells, definition list items: append translation inside the same element
      const translationEl = document.createElement('div');
      translationEl.className = 'translation-text translation-inline';
      translationEl.innerHTML = translatedHTML;
      htmlEl.appendChild(translationEl);
    } else if (htmlEl.closest('blockquote')) {
      // For elements inside blockquote: append translation inside, styled differently
      const translationEl = document.createElement('div');
      translationEl.className = 'translation-text translation-blockquote';
      translationEl.innerHTML = translatedHTML;
      htmlEl.appendChild(translationEl);
    } else {
      // For standalone paragraphs, headings, figcaption: insert after as sibling
      const translationEl = document.createElement('div');
      translationEl.className = 'translation-text';
      translationEl.innerHTML = translatedHTML;
      htmlEl.parentNode?.insertBefore(translationEl, htmlEl.nextSibling);
    }

    // Mark this element as translated
    translatedElements.add(htmlEl);
  }

  // Re-apply rendering enhancements to translation elements (for math formulas)
  await nextTick();
  proseContainer.querySelectorAll('.translation-text').forEach((el) => {
    renderMathFormulas(el as HTMLElement);
    highlightCodeBlocks(el as HTMLElement);
  });

  // Re-attach ALL event listeners after translation modifies the DOM
  // This includes unwrapping images from links, attaching image handlers, and link handlers
  await reattachImageInteractions();

  isTranslatingContent.value = false;
}

async function reattachImageInteractions() {
  if (!props.attachImageEventListeners || !props.articleContent) return;
  await nextTick();
  props.attachImageEventListeners();
}

// Clear text selection when clicking outside the selected content
function handleContainerClick(event: MouseEvent) {
  const selection = window.getSelection();
  if (!selection || selection.toString().length === 0) return;

  const target = event.target as HTMLElement;

  // Don't clear if clicking on:
  // - Links, buttons, or interactive elements
  // - Inputs, textareas
  // - Elements within the selection
  const isInteractive =
    target.tagName === 'A' ||
    target.tagName === 'BUTTON' ||
    target.tagName === 'INPUT' ||
    target.tagName === 'TEXTAREA' ||
    target.closest('a') !== null ||
    target.closest('button') !== null;

  if (isInteractive) return;

  // Check if target is within the current selection
  try {
    if (selection.containsNode(target, true)) {
      return;
    }
  } catch {
    // containsNode can throw in some cases, ignore and proceed
  }

  // Clear the selection
  selection.removeAllRanges();
}

// Handle auto show all content setting change
function onAutoShowAllContentChanged(e: Event): void {
  const customEvent = e as CustomEvent<{ value: boolean }>;
  autoShowAllContent.value = customEvent.detail.value;
}

// Handle summary settings change
async function onSummarySettingsChanged(): Promise<void> {
  // Reload summary settings to get the latest enabled state
  await loadSummarySettings();

  // Clear cached summary when settings change
  if (props.article) {
    summaryResult.value = null;
    // Auto-generate summary if newly enabled
    if (shouldAutoGenerateSummary() && props.articleContent) {
      setTimeout(() => generateSummary(props.article), 100);
    }
  }
}

// Watch for article changes and regenerate summary + translations
watch(
  () => props.article?.id,
  async (newId, oldId) => {
    if (newId !== oldId) {
      // Cancel any ongoing summary generation for the previous article
      if (oldId !== undefined) {
        cancelSummaryGeneration(oldId);
      }

      summaryResult.value = null;
      translatedTitle.value = '';
      lastTranslatedArticleId.value = null; // Reset translation tracking
      fullArticleContent.value = ''; // Reset full article content when switching articles

      if (props.article) {
        // Check if article has a cached summary first
        if (props.article.summary && props.article.summary.trim() !== '') {
          // Load the cached summary by calling API to get HTML
          // Don't use on-the-fly summarization, let backend convert cached markdown to HTML
          const result = await generateSummaryComposable(props.article, '', false);

          // Set summary result
          if (result) {
            summaryResult.value = result;
          }
        } else if (shouldAutoGenerateSummary()) {
          // Only auto-generate if no cached summary exists
          setTimeout(() => generateSummary(props.article), 100);
        }

        // Translate title
        if (translationEnabled.value) {
          translateTitle(props.article);
        }
      }
    }
  }
);

// Watch for article content changes to trigger translation
// This handles both cases:
// 1. Content is loaded from cache (isLoadingContent never changes)
// 2. Content is fetched and becomes available
watch(
  () => [props.article?.id, props.articleContent, translationEnabled.value] as const,
  async (newValue, oldValue) => {
    const [newArticleId, newContent, newTranslationEnabled] = newValue || [
      undefined,
      undefined,
      false,
    ];
    const [oldArticleId, oldContent, oldTranslationEnabled] = oldValue || [
      undefined,
      undefined,
      false,
    ];

    // Trigger when:
    // 1. Article changes AND content is present
    // 2. Same article but content changes (from empty to loaded) AND translation is enabled
    // 3. Translation setting changes from false to true AND content is present
    const articleChanged = newArticleId !== oldArticleId;
    const contentJustLoaded =
      newArticleId && oldContent === '' && newContent && newContent !== oldContent;
    const translationJustEnabled =
      oldTranslationEnabled === false && newTranslationEnabled === true;

    const shouldTrigger =
      newContent && newArticleId && (articleChanged || contentJustLoaded || translationJustEnabled);

    if (shouldTrigger) {
      // Wait for DOM to update with the new content
      await nextTick();

      // Enhance rendering first (math formulas, etc.)
      enhanceRendering('.prose-content');

      // Re-attach image event listeners after rendering enhancements
      await reattachImageInteractions();

      // Auto-fetch full article if setting is enabled
      if (shouldAutoExpandContent.value && !fullArticleContent.value) {
        setTimeout(() => fetchFullArticle(false), 200);
      }

      // Generate summary if needed
      if (shouldAutoGenerateSummary()) {
        setTimeout(() => generateSummary(props.article), 100);
      }

      // Translate content if enabled
      if (newTranslationEnabled && lastTranslatedArticleId.value !== newArticleId) {
        await nextTick();
        translateContentParagraphs(newContent);
      }
    }
  },
  { immediate: true } // Run immediately on component mount
);

onMounted(async () => {
  await loadSettings();
  if (props.article) {
    // Check for cached summary first
    if (props.article.summary && props.article.summary.trim() !== '') {
      // Load the cached summary by calling API to get HTML
      const result = await generateSummaryComposable(props.article, '', false);

      // Set summary result
      if (result) {
        summaryResult.value = result;
      }
    } else if (shouldAutoGenerateSummary() && props.articleContent) {
      // Only auto-generate if no cached summary exists
      setTimeout(() => generateSummary(props.article), 100);
    }

    // Translate title
    if (translationEnabled.value) {
      translateTitle(props.article);
    }
    // Content translation is handled by the watch on [article.id, articleContent]

    // Enhance rendering if content is already loaded
    if (props.articleContent && !props.isLoadingContent) {
      await nextTick();
      enhanceRendering('.prose-content');
      // Re-attach image event listeners after rendering
      await reattachImageInteractions();

      // Auto-fetch full article if setting is enabled and content is already loaded
      if (
        shouldAutoExpandContent.value &&
        !fullArticleContent.value &&
        !isFetchingFullArticle.value
      ) {
        setTimeout(() => fetchFullArticle(false), 200);
      }
    }
  }
});

// Ensure image interactions stay attached when content is (re)rendered
watch(
  () => props.articleContent,
  async (content) => {
    if (content) {
      // Wait for v-html to update the DOM before attaching event listeners
      await nextTick();
      await reattachImageInteractions();
    }
  },
  { immediate: true }
);

// Watch for full article content changes and reattach event listeners
// This is necessary because displayContent uses fullArticleContent when available,
// but the watch above only monitors props.articleContent
watch(fullArticleContent, async (content) => {
  if (content) {
    // Wait for v-html to update the DOM before attaching event listeners
    await nextTick();
    await reattachImageInteractions();
  }
});

// Clean up event listeners
onBeforeUnmount(() => {
  // Cancel any ongoing summary generation
  if (props.article?.id) {
    cancelSummaryGeneration(props.article.id);
  }

  window.removeEventListener(
    'auto-show-all-content-changed',
    onAutoShowAllContentChanged as EventListener
  );

  window.removeEventListener('summary-settings-changed', onSummarySettingsChanged as EventListener);
});
</script>

<template>
  <div
    class="flex-1 overflow-y-auto bg-bg-primary p-3 sm:p-6 scroll-smooth"
    @click="handleContainerClick"
  >
    <div
      class="max-w-3xl mx-auto bg-bg-primary"
      :class="{
        'hide-translations': !showTranslations,
        'translation-only-mode': translationSettings.translationOnlyMode,
      }"
    >
      <ArticleTitle
        :article="article"
        :translated-title="translatedTitle"
        :is-translating-title="isTranslatingTitle"
        :translation-enabled="translationEnabled"
        :translation-skipped="translationSkipped"
        :is-translating-content="isTranslatingContent"
        @force-translate="forceTranslateContent"
      />

      <!-- Audio Player (if article has audio) -->
      <AudioPlayer
        v-if="article.audio_url"
        :audio-url="article.audio_url"
        :article-title="article.title"
      />

      <!-- Video Player (if article has video) -->
      <VideoPlayer
        v-if="article.video_url"
        :video-url="article.video_url"
        :article-title="article.title"
      />

      <ArticleSummary
        v-if="summaryEnabled"
        :summary-result="summaryResult"
        :is-loading-summary="isLoadingSummary"
        :translation-enabled="translationEnabled"
        :summary-provider="summaryProvider"
        :summary-trigger-mode="summaryTriggerMode"
        :is-loading-content="props.isLoadingContent"
        @generate-summary="generateSummary(props.article, true)"
      />

      <ArticleLoading v-if="isLoadingContent" />

      <ArticleBody
        v-else
        :article-content="displayContent"
        :is-translating-content="isTranslatingContent"
        :has-media-content="!!(article.audio_url || article.video_url)"
        :is-loading-content="isLoadingContent"
        @retry-load="handleRetryLoad"
      />

      <!-- Full-text fetch button -->
      <div v-if="showFullTextButton" class="flex justify-center mt-4 mb-4">
        <button
          :disabled="isFetchingFullArticle"
          class="btn-secondary-compact flex items-center gap-2"
          @click="() => fetchFullArticle()"
        >
          <PhSpinnerGap v-if="isFetchingFullArticle" :size="14" class="animate-spin" />
          <PhArticleNyTimes v-else :size="14" />
          <span>{{
            isFetchingFullArticle
              ? t('article.action.fetchingFullArticle')
              : t('article.action.fetchFullArticle')
          }}</span>
        </button>
      </div>
    </div>

    <!-- Chat Button (shown when content is loaded and chat is enabled) -->
    <ArticleChatButton v-if="showChatButton && !isChatPanelOpen" @click="isChatPanelOpen = true" />

    <!-- Chat Panel -->
    <ArticleChatPanel
      v-if="isChatPanelOpen"
      :article="article"
      :article-content="articleContent"
      :settings="{ ai_chat_enabled: appSettings.ai_chat_enabled }"
      @close="isChatPanelOpen = false"
    />
  </div>
</template>

<style scoped>
@reference "../../../../style.css";

.btn-secondary {
  @apply bg-bg-tertiary border border-border text-text-primary px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium hover:bg-bg-secondary transition-colors;
}

.btn-secondary-compact {
  @apply border border-border px-3 py-1.5 rounded-md cursor-pointer flex items-center gap-2 text-sm font-normal transition-all duration-200;
  background-color: var(--bg-tertiary);
  color: var(--text-secondary);
  opacity: 0.6;
}

.btn-secondary-compact:hover {
  opacity: 1;
  color: var(--text-primary);
}
</style>
