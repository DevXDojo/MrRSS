<script setup lang="ts">
import { ref, computed, nextTick, useId, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhDotsThree,
  PhFunnel,
  PhCaretRight,
  PhSortAscending,
  PhSortDescending,
} from '@phosphor-icons/vue';
import type { ArticleSortOrder } from '@/stores/app';
import type { ArticleGroupBy } from '@/utils/articleGrouping';

const props = defineProps<{
  sortOrder: ArticleSortOrder;
  groupBy: ArticleGroupBy;
  filterCount: number;
}>();
const emit = defineEmits<{
  sort: [value: ArticleSortOrder];
  group: [value: ArticleGroupBy];
  filter: [];
}>();
const { t } = useI18n();
const id = useId();
const open = ref(false);
const trigger = ref<HTMLButtonElement | null>(null);
const panel = ref<HTMLElement | null>(null);
const position = ref({ left: '0px', top: '0px', maxHeight: 'none' });
const orders: ArticleSortOrder[] = ['newest', 'oldest'];
const modes: ArticleGroupBy[] = ['none', 'date', 'feed'];
const customized = computed(
  () => props.filterCount > 0 || props.groupBy !== 'none' || props.sortOrder !== 'newest'
);

function close(restoreFocus = false) {
  open.value = false;
  if (restoreFocus) trigger.value?.focus();
}

function contains(target: EventTarget | null) {
  return (
    target instanceof Node && (trigger.value?.contains(target) || panel.value?.contains(target))
  );
}

function outside(event: Event) {
  if (open.value && !contains(event.target)) close();
}

function placePanel() {
  if (!trigger.value || !panel.value) return;
  const anchor = trigger.value.getBoundingClientRect();
  const height = panel.value.getBoundingClientRect().height;
  const margin = 8;
  const gap = 6;
  const below = window.innerHeight - anchor.bottom - gap - margin;
  const above = anchor.top - gap - margin;
  const useAbove = height > below && above > below;
  position.value = {
    left: `${Math.max(margin, Math.min(anchor.right - panel.value.offsetWidth, window.innerWidth - panel.value.offsetWidth - margin))}px`,
    top: `${Math.max(margin, useAbove ? anchor.top - gap - Math.min(height, above) : anchor.bottom + gap)}px`,
    maxHeight: `${Math.max(0, useAbove ? above : below)}px`,
  };
}

async function show() {
  open.value = true;
  await nextTick();
  if (!open.value) return;
  placePanel();
  panel.value?.querySelector<HTMLInputElement>('input:checked')?.focus({ preventScroll: true });
}

function viewportChanged(event: Event) {
  // Article reloads may scroll the list; only dismiss if the anchor itself moves.
  const target = event.target;
  if (
    open.value &&
    (target === window ||
      target === document ||
      (target instanceof Element && trigger.value && target.contains(trigger.value)))
  )
    close();
}

onMounted(() => {
  document.addEventListener('pointerdown', outside);
  document.addEventListener('focusin', outside);
  document.addEventListener('scroll', viewportChanged, true);
  window.addEventListener('resize', viewportChanged);
});
onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', outside);
  document.removeEventListener('focusin', outside);
  document.removeEventListener('scroll', viewportChanged, true);
  window.removeEventListener('resize', viewportChanged);
});
</script>

<template>
  <div @keydown.esc.stop.prevent="close(true)">
    <button
      ref="trigger"
      type="button"
      class="relative rounded p-1 transition-colors hover:bg-bg-tertiary focus-visible:outline focus-visible:outline-2 focus-visible:outline-accent sm:p-1.5"
      :class="open || customized ? 'text-accent' : 'text-text-secondary hover:text-text-primary'"
      :title="t('article.list.more')"
      :aria-label="t('article.list.more')"
      :aria-expanded="open"
      :aria-controls="open ? id : undefined"
      aria-haspopup="dialog"
      @click="open ? close() : show()"
      @keydown.down.prevent="show()"
    >
      <PhDotsThree :size="18" class="sm:h-5 sm:w-5" weight="bold" />
      <span v-if="customized" class="absolute right-0 top-0 h-1.5 w-1.5 rounded-full bg-accent" />
    </button>
    <Teleport to="body">
      <div
        v-if="open"
        :id="id"
        ref="panel"
        role="dialog"
        :aria-labelledby="`${id}-title`"
        class="fixed z-50 w-72 max-w-[calc(100vw-1rem)] overflow-y-auto rounded-xl border border-border bg-bg-primary text-text-primary shadow-xl"
        :style="position"
        @keydown.esc.stop.prevent="close(true)"
      >
        <div class="flex items-center justify-between gap-2 px-3 pb-1 pt-3">
          <span :id="`${id}-title`" class="text-sm font-semibold">{{
            t('article.list.options')
          }}</span>
          <span class="text-xs text-text-secondary">{{
            t('article.list.appliesImmediately')
          }}</span>
        </div>
        <div class="space-y-3 p-3">
          <fieldset class="min-w-0">
            <legend class="mb-2 text-xs font-medium text-text-secondary">
              {{ t('article.list.sorting.label') }}
            </legend>
            <div class="grid grid-cols-2 gap-1 rounded-lg bg-bg-secondary p-1">
              <label v-for="order in orders" :key="order" class="relative min-w-0 cursor-pointer">
                <input
                  type="radio"
                  :name="`${id}-sort`"
                  :value="order"
                  :checked="sortOrder === order"
                  class="peer sr-only"
                  @change="emit('sort', order)"
                />
                <span
                  class="flex items-center justify-center gap-1.5 rounded-md px-2 py-2 text-sm text-text-secondary transition-colors hover:text-text-primary peer-checked:bg-bg-primary peer-checked:font-medium peer-checked:text-accent peer-checked:shadow-sm peer-focus-visible:outline peer-focus-visible:outline-2 peer-focus-visible:outline-accent"
                >
                  <component
                    :is="order === 'newest' ? PhSortDescending : PhSortAscending"
                    :size="16"
                    aria-hidden="true"
                  />
                  {{ t(`article.list.sorting.${order}`) }}
                </span>
              </label>
            </div>
          </fieldset>
          <fieldset class="min-w-0">
            <legend class="mb-2 text-xs font-medium text-text-secondary">
              {{ t('article.list.grouping.label') }}
            </legend>
            <div class="grid grid-cols-3 gap-1 rounded-lg bg-bg-secondary p-1">
              <label v-for="mode in modes" :key="mode" class="relative min-w-0 cursor-pointer">
                <input
                  type="radio"
                  :name="`${id}-group`"
                  :value="mode"
                  :checked="groupBy === mode"
                  class="peer sr-only"
                  @change="emit('group', mode)"
                />
                <span
                  class="flex h-full items-center justify-center rounded-md px-1 py-2 text-center text-xs text-text-secondary transition-colors hover:text-text-primary peer-checked:bg-bg-primary peer-checked:font-medium peer-checked:text-accent peer-checked:shadow-sm peer-focus-visible:outline peer-focus-visible:outline-2 peer-focus-visible:outline-accent"
                >
                  {{ t(`article.list.grouping.${mode}`) }}
                </span>
              </label>
            </div>
          </fieldset>
        </div>
        <div class="border-t border-border p-1.5">
          <button
            type="button"
            class="flex w-full items-center gap-2 rounded-lg px-2 py-2 text-left text-sm transition-colors hover:bg-bg-tertiary focus-visible:outline focus-visible:outline-2 focus-visible:outline-accent"
            @click="
              close();
              emit('filter');
            "
          >
            <PhFunnel :size="16" :class="filterCount ? 'text-accent' : 'text-text-secondary'" />
            <span class="flex-1">{{ t('modal.filter.filter') }}</span>
            <span
              v-if="filterCount"
              class="rounded-full bg-accent/10 px-2 py-0.5 text-xs font-medium text-accent"
              >{{ filterCount }}</span
            >
            <PhCaretRight :size="14" class="text-text-secondary" />
          </button>
        </div>
      </div>
    </Teleport>
  </div>
</template>
