<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhArrowLineUp } from '@phosphor-icons/vue';
import {
  buildTocItems,
  calcTocProgress,
  shouldShowTocText,
  type TocItem,
  type HeadingSnapshot,
} from '@/composables/article/floatingToc';

interface Props {
  articleId: number;
  enabled: boolean;
  scrollContainer: HTMLElement | null;
}

const props = defineProps<Props>();
const { t } = useI18n();

const tocItems = ref<TocItem[]>([]);
const tocListEl = ref<HTMLElement | null>(null);
const activeIndex = ref(-1);
const sectionProgress = ref(0);
const articleProgress = ref(0);
const isDesktop = ref(false);

let mediaQuery: ReturnType<typeof window.matchMedia> | null = null;
let containerObserver: InstanceType<typeof window.MutationObserver> | null = null;
let pendingProseObserver: InstanceType<typeof window.MutationObserver> | null = null;
let scrollContainerEl: HTMLElement | null = null;
let rebuildRaf: number | null = null;
let scrollRaf: number | null = null;
let lastAutoScrolledIndex = -1;

function shouldShowText(itemIndex: number): boolean {
  return shouldShowTocText(itemIndex, activeIndex.value, tocItems.value);
}

function queueRebuild(): void {
  if (rebuildRaf !== null) {
    window.cancelAnimationFrame(rebuildRaf);
  }
  rebuildRaf = window.requestAnimationFrame(() => {
    rebuildRaf = null;
    buildToc();
  });
}

function queueScrollSync(): void {
  if (scrollRaf !== null) return;
  scrollRaf = window.requestAnimationFrame(() => {
    scrollRaf = null;
    updateActiveSection();
  });
}

function getMarkerFillPercent(index: number): number {
  if (index !== activeIndex.value) return 0;
  return sectionProgress.value;
}

function autoScrollTocToActive(index: number): void {
  const list = tocListEl.value;
  if (!list || index < 0) return;
  if (list.scrollHeight <= list.clientHeight + 1) return;

  const item = list.querySelector<HTMLElement>(`[data-toc-index="${index}"]`);
  if (!item) return;

  const itemTop = item.offsetTop;
  const targetTop = itemTop - (list.clientHeight - item.offsetHeight) / 2;
  const maxTop = Math.max(0, list.scrollHeight - list.clientHeight);
  const nextTop = Math.max(0, Math.min(targetTop, maxTop));
  const delta = Math.abs(nextTop - list.scrollTop);
  if (delta < 6) return;

  list.scrollTo({
    top: nextTop,
    behavior: 'smooth',
  });
}

function buildToc(): void {
  const container = props.scrollContainer;
  if (!props.enabled || !isDesktop.value || !container) {
    tocItems.value = [];
    activeIndex.value = -1;
    sectionProgress.value = 0;
    articleProgress.value = 0;
    return;
  }

  const proseContainer = container.querySelector('.prose-content');
  if (!proseContainer) {
    tocItems.value = [];
    activeIndex.value = -1;
    sectionProgress.value = 0;
    articleProgress.value = calcTocProgress(
      container.scrollTop,
      container.scrollHeight,
      container.clientHeight,
      []
    ).articleProgress;
    return;
  }

  const headings = Array.from(proseContainer.querySelectorAll<HTMLElement>('h1, h2, h3'));
  const containerRect = container.getBoundingClientRect();
  const articleId = props.articleId || 0;

  const snapshots: HeadingSnapshot[] = headings
    .map((heading, index) => {
      const level = Number(heading.tagName.slice(1));
      if (level < 1 || level > 3) return null;

      const translationEl = heading.nextElementSibling as HTMLElement | null;
      const hasHeadingTranslation =
        translationEl &&
        translationEl.classList.contains('translation-text') &&
        !translationEl.classList.contains('translation-inline') &&
        !translationEl.classList.contains('translation-blockquote');

      return {
        level: level as 1 | 2 | 3,
        offsetTop: heading.getBoundingClientRect().top - containerRect.top + container.scrollTop,
        rawText: heading.textContent || '',
        translatedText: hasHeadingTranslation ? translationEl.textContent || '' : '',
        existingId: heading.id || undefined,
        domIndex: index,
      };
    })
    .filter((item): item is HeadingSnapshot => item !== null);

  const { items, generatedIds } = buildTocItems(snapshots, articleId);
  generatedIds.forEach(({ domIndex, id }) => {
    if (headings[domIndex]) {
      headings[domIndex].id = id;
    }
  });

  tocItems.value = items;
  updateActiveSection();
}

function updateActiveSection(): void {
  const container = props.scrollContainer;
  const items = tocItems.value;

  if (!container || items.length === 0) {
    activeIndex.value = -1;
    sectionProgress.value = 0;
    articleProgress.value = container
      ? calcTocProgress(container.scrollTop, container.scrollHeight, container.clientHeight, [])
          .articleProgress
      : 0;
    lastAutoScrolledIndex = -1;
    return;
  }

  const progress = calcTocProgress(
    container.scrollTop,
    container.scrollHeight,
    container.clientHeight,
    items
  );
  articleProgress.value = progress.articleProgress;
  activeIndex.value = progress.activeIndex;
  sectionProgress.value = progress.sectionProgress;

  if (progress.activeIndex !== lastAutoScrolledIndex) {
    autoScrollTocToActive(progress.activeIndex);
    lastAutoScrolledIndex = progress.activeIndex;
  }
}

function scrollToHeading(item: TocItem): void {
  const container = props.scrollContainer;
  if (!container) return;

  const targetTop = Math.max(0, item.offsetTop - 12);
  container.scrollTo({
    top: targetTop,
    behavior: 'smooth',
  });
}

function scrollToTop(): void {
  const container = props.scrollContainer;
  if (!container) return;

  container.scrollTo({
    top: 0,
    behavior: 'smooth',
  });
}

function handleMediaChange(event: Event): void {
  const mediaEvent = event as Event & { matches?: boolean };
  isDesktop.value = Boolean(mediaEvent.matches);
  queueRebuild();
}

function bindScrollContainer(container: HTMLElement | null): void {
  if (scrollContainerEl) {
    scrollContainerEl.removeEventListener('scroll', queueScrollSync);
  }

  scrollContainerEl = container;

  if (scrollContainerEl) {
    scrollContainerEl.addEventListener('scroll', queueScrollSync, { passive: true });
  }
}

function connectContainerObserver(): void {
  containerObserver?.disconnect();
  pendingProseObserver?.disconnect();
  containerObserver = null;
  pendingProseObserver = null;

  const container = props.scrollContainer;
  if (!container) return;

  const proseContainer = container.querySelector('.prose-content');
  if (!proseContainer) {
    // The article body may render asynchronously. Watch container until prose appears.
    pendingProseObserver = new window.MutationObserver(() => {
      const readyProse = container.querySelector('.prose-content');
      if (!readyProse) return;

      pendingProseObserver?.disconnect();
      pendingProseObserver = null;
      connectContainerObserver();
      queueRebuild();
    });

    pendingProseObserver.observe(container, {
      childList: true,
      subtree: true,
    });
    return;
  }

  containerObserver = new window.MutationObserver(() => queueRebuild());
  containerObserver.observe(proseContainer, {
    childList: true,
    subtree: true,
    characterData: true,
  });
}

onMounted(async () => {
  mediaQuery = window.matchMedia('(min-width: 768px)');
  isDesktop.value = mediaQuery.matches;

  mediaQuery.addEventListener('change', handleMediaChange);
  window.addEventListener('resize', queueRebuild);

  await nextTick();
  bindScrollContainer(props.scrollContainer);
  connectContainerObserver();
  queueRebuild();
});

watch(
  () => [props.articleId, props.enabled] as const,
  async () => {
    await nextTick();
    bindScrollContainer(props.scrollContainer);
    connectContainerObserver();
    queueRebuild();
  }
);

watch(
  () => props.scrollContainer,
  async (container) => {
    await nextTick();
    bindScrollContainer(container);
    connectContainerObserver();
    queueRebuild();
  }
);

onBeforeUnmount(() => {
  if (mediaQuery) {
    mediaQuery.removeEventListener('change', handleMediaChange);
  }

  window.removeEventListener('resize', queueRebuild);
  scrollContainerEl?.removeEventListener('scroll', queueScrollSync);
  containerObserver?.disconnect();
  pendingProseObserver?.disconnect();

  if (rebuildRaf !== null) {
    window.cancelAnimationFrame(rebuildRaf);
  }
  if (scrollRaf !== null) {
    window.cancelAnimationFrame(scrollRaf);
  }
});
</script>

<template>
  <div
    v-if="enabled && isDesktop && tocItems.length > 0"
    class="pointer-events-none absolute right-[8px] top-[76px] bottom-6 z-40 flex w-[max(15%,125px)] flex-col items-end justify-center [container-type:inline-size]"
  >
    <div class="mb-2 w-full text-right text-[10px] font-medium text-text-secondary opacity-75">
      {{ articleProgress }}%
    </div>

    <div class="group/toclist pointer-events-auto relative w-full max-h-[80%]">
      <div
        class="pointer-events-none absolute -inset-y-1.5 -left-2 -right-1 rounded-lg border border-border bg-bg-secondary shadow-lg shadow-black/15 opacity-0 scale-[0.98] transition-all duration-200 group-hover/toclist:opacity-100 group-hover/toclist:scale-100 dark:shadow-black/40"
      ></div>

      <ul
        ref="tocListEl"
        class="toc-list-scroll relative z-[1] flex w-full max-h-full flex-col items-start gap-1 overflow-y-scroll [scrollbar-gutter:stable_both-edges]"
      >
        <li
          v-for="(item, index) in tocItems"
          :key="item.id"
          class="w-full"
          :data-level="item.level"
          :data-toc-index="index"
        >
          <button
            class="group/item flex w-full cursor-pointer items-center justify-start gap-1 rounded py-0.5 transition-colors"
            :style="{ '--toc-level': String(item.level) }"
            @click="scrollToHeading(item)"
          >
            <span
              :class="[
                'toc-text flex-1 min-w-0 truncate text-left text-xs opacity-0 transition-all duration-200 [margin-left:calc((var(--toc-level,1)-1)*12px)]',
                index === activeIndex ? 'text-text-primary' : 'text-text-secondary',
                shouldShowText(index) ? 'toc-text-visible opacity-[0.85] max-w-full' : 'max-w-0',
                'group-hover/toclist:opacity-[0.85] group-hover/toclist:max-w-full group-hover/item:text-text-primary group-hover/item:opacity-100',
              ]"
              :data-level="item.level"
              :title="item.text"
            >
              {{ item.text }}
            </span>
            <span class="ml-auto flex w-[34px] shrink-0 justify-end">
              <span
                :class="[
                  'relative overflow-hidden bg-text-secondary transition-colors duration-150 group-hover/item:bg-text-primary',
                  index === activeIndex ? 'h-[3px] opacity-100' : 'h-[2px] opacity-70',
                ]"
                :style="{ width: `${item.markerWidth}px` }"
              >
                <span
                  class="absolute left-0 top-0 h-full bg-accent transition-all duration-150 group-hover/item:bg-text-primary"
                  :style="{ width: `${getMarkerFillPercent(index)}%` }"
                ></span>
              </span>
            </span>
          </button>
        </li>
      </ul>
    </div>

    <button
      class="pointer-events-auto mt-3 flex h-7 w-7 items-center justify-center self-end rounded bg-transparent text-text-secondary transition-colors hover:bg-[color-mix(in_srgb,var(--bg-tertiary)_70%,transparent)] hover:text-text-primary"
      :title="t('common.back')"
      @click="scrollToTop"
    >
      <PhArrowLineUp :size="14" />
    </button>
  </div>

  <template v-else-if="enabled && !isDesktop">
    <!-- TODO: Add floating TOC UI for mobile devices. -->
  </template>
</template>

<style scoped>
.toc-list-scroll {
  scrollbar-width: thin;
  scrollbar-color: transparent transparent;
}

.group\/toclist:hover .toc-list-scroll {
  scrollbar-color: var(--border-color) transparent;
}

.toc-list-scroll::-webkit-scrollbar {
  width: 6px;
}

.toc-list-scroll::-webkit-scrollbar-track {
  background: transparent;
}

.toc-list-scroll::-webkit-scrollbar-thumb {
  background: transparent;
  border-radius: 3px;
}

.group\/toclist:hover .toc-list-scroll::-webkit-scrollbar-thumb {
  background: var(--border-color);
}

.toc-list-scroll::-webkit-scrollbar-thumb:hover {
  background: transparent;
}

.group\/toclist:hover .toc-list-scroll::-webkit-scrollbar-thumb:hover {
  background: var(--text-secondary);
}

@container (max-width: 150px) {
  .toc-text-visible {
    opacity: 0;
    max-width: 0;
  }
}
</style>
