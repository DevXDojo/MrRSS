<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhMagnifyingGlass, PhCaretUp, PhCaretDown, PhX } from '@phosphor-icons/vue';

const { t } = useI18n();

interface Props {
  containerSelector: string; // CSS selector for the content container to search in
  articleId?: number | null; // Article ID to trigger re-search when switching
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
}>();

// Persistent search query that survives article switches
const searchQuery = ref('');
const currentIndex = ref(0);
const totalMatches = ref(0);
const searchInput = ref<HTMLInputElement | null>(null);

// Track if search is focused to handle ESC key properly
const isSearchFocused = ref(false);

// Track current highlights
let currentMarkElements: HTMLElement[] = [];

// Focus input when mounted
onMounted(() => {
  nextTick(() => {
    searchInput.value?.focus();
  });

  // Add focus/blur listeners to track search focus state
  if (searchInput.value) {
    searchInput.value.addEventListener('focus', () => {
      isSearchFocused.value = true;
    });
    searchInput.value.addEventListener('blur', () => {
      isSearchFocused.value = false;
    });
  }
});

// Watch for content changes (article switches) and re-execute search if query exists
watch(
  () => props.articleId,
  () => {
    // Small delay to ensure new content is rendered
    nextTick(() => {
      setTimeout(() => {
        if (searchQuery.value.trim()) {
          performSearch(searchQuery.value);
        }
      }, 100);
    });
  }
);

// Clean up highlights when closing or changing query
function clearHighlights() {
  currentMarkElements.forEach((mark) => {
    const parent = mark.parentNode;
    if (parent) {
      // Replace the mark element with its text content
      parent.replaceChild(document.createTextNode(mark.textContent || ''), mark);
      // Normalize the parent to merge adjacent text nodes
      parent.normalize();
    }
  });
  currentMarkElements = [];
  currentIndex.value = 0;
  totalMatches.value = 0;
}

// Watch for search query changes
watch(searchQuery, (newQuery) => {
  if (!newQuery.trim()) {
    clearHighlights();
    return;
  }
  performSearch(newQuery);
});

function performSearch(query: string) {
  // Clear previous highlights
  clearHighlights();

  if (!query.trim()) {
    return;
  }

  // Find the content container
  const container = document.querySelector(props.containerSelector);
  if (!container) {
    console.warn('Search container not found:', props.containerSelector);
    return;
  }

  // Use TreeWalker to find all text nodes
  const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT, {
    acceptNode: (node) => {
      // Skip if parent is script, style, or our own mark elements
      const parent = node.parentElement;
      if (!parent) return NodeFilter.FILTER_REJECT;
      if (['SCRIPT', 'STYLE', 'MARK', 'CODE', 'PRE'].includes(parent.tagName)) {
        return NodeFilter.FILTER_REJECT;
      }
      // Only accept if contains search text
      if (node.textContent?.toLowerCase().includes(query.toLowerCase())) {
        return NodeFilter.FILTER_ACCEPT;
      }
      return NodeFilter.FILTER_REJECT;
    },
  });

  const textNodes: Text[] = [];
  let node: Node | null;
  while ((node = walker.nextNode())) {
    textNodes.push(node as Text);
  }

  // Highlight matches
  const regex = new RegExp(`(${escapeRegex(query)})`, 'gi');
  const allHighlights: HTMLElement[] = [];

  textNodes.forEach((textNode) => {
    const text = textNode.textContent || '';
    if (!regex.test(text)) return;

    // Split text and wrap matches
    const fragments = text.split(regex);
    const parent = textNode.parentNode;
    if (!parent) return;

    const newNodes: Node[] = [];
    fragments.forEach((fragment) => {
      if (regex.test(fragment)) {
        const mark = document.createElement('mark');
        mark.className = 'search-highlight';
        mark.textContent = fragment;
        newNodes.push(mark);
        allHighlights.push(mark);
      } else if (fragment) {
        newNodes.push(document.createTextNode(fragment));
      }
    });

    // Replace text node with new nodes
    parent.insertBefore(document.createDocumentFragment(), textNode);
    newNodes.forEach((newNode) => {
      parent.insertBefore(newNode, textNode);
    });
    parent.removeChild(textNode);
  });

  currentMarkElements = allHighlights;
  totalMatches.value = allHighlights.length;
  currentIndex.value = allHighlights.length > 0 ? 1 : 0;

  // Scroll to first match
  if (allHighlights.length > 0) {
    scrollToHighlight(0);
  }
}

function escapeRegex(string: string): string {
  return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function scrollToHighlight(index: number) {
  if (index < 0 || index >= currentMarkElements.length) return;

  const element = currentMarkElements[index];
  element.scrollIntoView({
    behavior: 'smooth',
    block: 'center',
  });

  // Update active state
  currentMarkElements.forEach((mark, i) => {
    if (i === index) {
      mark.classList.add('search-highlight-active');
    } else {
      mark.classList.remove('search-highlight-active');
    }
  });
}

function goToNext() {
  if (currentMarkElements.length === 0) return;

  const nextIndex = currentIndex.value % currentMarkElements.length;
  currentIndex.value = nextIndex + 1;
  scrollToHighlight(nextIndex);
}

function goToPrevious() {
  if (currentMarkElements.length === 0) return;

  const prevIndex =
    (currentIndex.value - 2 + currentMarkElements.length) % currentMarkElements.length;
  currentIndex.value = prevIndex + 1;
  scrollToHighlight(prevIndex);
}

function close() {
  clearHighlights();
  emit('close');
}

// Handle ESC key on input - only close if there's search content
function handleEscOnInput() {
  if (searchQuery.value.trim()) {
    close();
  }
}

// Keyboard shortcuts
function handleKeydown(e: KeyboardEvent) {
  // Track focus state
  if (document.activeElement === searchInput.value) {
    isSearchFocused.value = true;
  } else {
    isSearchFocused.value = false;
  }

  if (e.key === 'Enter') {
    if (e.shiftKey) {
      goToPrevious();
    } else {
      goToNext();
    }
    e.preventDefault();
  } else if (e.key === 'Escape') {
    // Only close search if the input is focused or if there's search content
    // This prevents conflict with the article close ESC key
    if (isSearchFocused.value || searchQuery.value.trim()) {
      close();
      e.preventDefault();
      e.stopPropagation();
    }
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown);
});

onBeforeUnmount(() => {
  clearHighlights();
  window.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
  <div class="find-in-page-bar">
    <div class="find-input-wrapper">
      <PhMagnifyingGlass :size="16" class="search-icon" />
      <input
        ref="searchInput"
        v-model="searchQuery"
        type="text"
        class="find-input"
        :placeholder="t('common.findInPage.findInPagePlaceholder')"
        @keydown.enter.exact.prevent="goToNext"
        @keydown.enter.shift.prevent="goToPrevious"
        @keydown.esc.prevent="handleEscOnInput"
      />
      <PhX
        v-if="searchQuery"
        :size="16"
        class="clear-icon"
        :title="t('common.clear')"
        @click="searchQuery = ''"
      />
    </div>

    <div v-if="searchQuery" class="find-navigation">
      <span class="find-count"> {{ currentIndex }} / {{ totalMatches }} </span>
      <button
        class="nav-button"
        :disabled="totalMatches === 0"
        :title="t('common.findInPage.previousMatch')"
        @click="goToPrevious"
      >
        <PhCaretUp :size="16" />
      </button>
      <button
        class="nav-button"
        :disabled="totalMatches === 0"
        :title="t('common.findInPage.nextMatch')"
        @click="goToNext"
      >
        <PhCaretDown :size="16" />
      </button>
      <button class="nav-button" :title="t('common.close')" @click="close">
        <PhX :size="16" />
      </button>
    </div>

    <!-- Close button when no search content -->
    <button v-if="!searchQuery" class="nav-button" :title="t('common.close')" @click="close">
      <PhX :size="16" />
    </button>
  </div>
</template>

<style scoped>
@reference "../../style.css";
.find-in-page-bar {
  @apply fixed top-[4.5rem] right-4 z-50 flex items-center gap-2 bg-bg-secondary border border-border rounded-lg shadow-lg p-2;
  width: auto;
  max-width: min(90vw, 500px);
  animation: slideDown 0.2s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.find-input-wrapper {
  @apply relative flex items-center;
}

.search-icon {
  @apply absolute left-2 text-text-secondary pointer-events-none;
}

.find-input {
  @apply bg-bg-tertiary text-text-primary rounded-md pl-8 pr-6 py-1.5 text-sm min-w-[200px] max-w-[300px] outline-none border border-transparent focus:border-border transition-colors;
}

.clear-icon {
  @apply absolute right-2 text-text-secondary cursor-pointer hover:text-text-primary;
  background: var(--bg-tertiary);
  padding-left: 4px;
}

.find-navigation {
  @apply flex items-center gap-1;
}

.find-count {
  @apply text-xs text-text-secondary px-2 min-w-[3.5rem] text-center;
}

.nav-button {
  @apply flex items-center justify-center w-7 h-7 rounded hover:bg-bg-tertiary text-text-primary disabled:opacity-30 disabled:cursor-not-allowed transition-colors;
}

@media (max-width: 640px) {
  .find-in-page-bar {
    @apply right-2 left-2;
    max-width: none;
  }

  .find-input {
    @apply min-w-[150px] max-w-[200px];
  }
}
</style>
