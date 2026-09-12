<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhDotsThree, PhFunnel, PhCheck } from '@phosphor-icons/vue';
import type { ArticleGroupBy } from '@/utils/articleGrouping';

defineProps<{ groupBy: ArticleGroupBy; filterCount: number }>();
const emit = defineEmits<{ group: [value: ArticleGroupBy]; filter: [] }>();
const { t } = useI18n();
const open = ref(false);
const root = ref<HTMLElement | null>(null);
const trigger = ref<HTMLButtonElement | null>(null);
const modes: ArticleGroupBy[] = ['none', 'date', 'feed'];
function close(restoreFocus = false) {
  open.value = false;
  if (restoreFocus) trigger.value?.focus();
}
function outside(event: Event) {
  if (event.target instanceof Node && !root.value?.contains(event.target)) close();
}
onMounted(() => document.addEventListener('pointerdown', outside));
onBeforeUnmount(() => document.removeEventListener('pointerdown', outside));
</script>

<template>
  <div
    ref="root"
    class="relative"
    @keydown.esc.stop.prevent="close(true)"
    @focusout="
      (event) => {
        if (!root?.contains(event.relatedTarget as Node)) close();
      }
    "
  >
    <button
      ref="trigger"
      type="button"
      class="relative rounded p-1 text-text-secondary transition-colors hover:bg-bg-tertiary hover:text-text-primary sm:p-1.5"
      :class="{ 'text-accent': filterCount > 0 || groupBy !== 'none' }"
      :title="t('article.list.more')"
      :aria-label="t('article.list.more')"
      :aria-expanded="open"
      @click="open = !open"
    >
      <PhDotsThree :size="18" class="sm:h-5 sm:w-5" weight="bold" />
      <span
        v-if="filterCount > 0"
        class="absolute right-0 top-0 h-1.5 w-1.5 rounded-full bg-accent"
      />
    </button>
    <div
      v-if="open"
      class="absolute right-0 top-full z-40 mt-1 w-52 overflow-hidden rounded-lg border border-border bg-bg-primary p-1 shadow-xl"
    >
      <button
        type="button"
        class="flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm text-text-primary hover:bg-bg-tertiary"
        @click="
          close();
          emit('filter');
        "
      >
        <PhFunnel :size="16" />
        <span class="flex-1">{{ t('modal.filter.filter') }}</span>
        <span v-if="filterCount" class="text-xs text-accent">{{ filterCount }}</span>
      </button>
      <div class="my-1 border-t border-border" />
      <div class="px-3 py-1.5 text-xs text-text-secondary">
        {{ t('article.list.grouping.label') }}
      </div>
      <div role="group" :aria-label="t('article.list.grouping.label')">
        <button
          v-for="mode in modes"
          :key="mode"
          type="button"
          class="flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm hover:bg-bg-tertiary"
          :class="groupBy === mode ? 'text-accent' : 'text-text-primary'"
          :aria-pressed="groupBy === mode"
          @click="
            emit('group', mode);
            close(true);
          "
        >
          <PhCheck :size="16" :class="groupBy === mode ? 'opacity-100' : 'opacity-0'" />
          {{ t(`article.list.grouping.${mode}`) }}
        </button>
      </div>
    </div>
  </div>
</template>
