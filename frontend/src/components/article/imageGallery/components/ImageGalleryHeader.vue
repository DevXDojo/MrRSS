<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhArrowClockwise, PhCheckCircle, PhCircle, PhList, PhTextT, PhTextTSlash } from '@phosphor-icons/vue';
import type { MediaTypeFilter } from '../types';

interface Props {
  title: string;
  isRefreshing: boolean;
  showTextOverlay: boolean;
  showOnlyUnread: boolean;
  mediaType: MediaTypeFilter;
}

defineProps<Props>();

const emit = defineEmits<{
  toggleSidebar: [];
  refresh: [];
  toggleTextOverlay: [];
  toggleShowOnlyUnread: [];
  updateMediaType: [mediaType: MediaTypeFilter];
  markAllRead: [];
}>();

const { t } = useI18n();
</script>

<template>
  <div
    class="flex-shrink-0 bg-bg-primary border-b border-border p-2 sm:p-4 flex items-center gap-3"
  >
    <!-- Sidebar toggle button (mobile only) -->
    <button
      class="p-2 rounded-lg hover:bg-bg-tertiary text-text-primary transition-colors md:hidden"
      :title="t('shortcut.toggle.sidebar')"
      @click="emit('toggleSidebar')"
    >
      <PhList :size="24" />
    </button>

    <!-- Title -->
    <div class="flex items-center gap-2 sm:gap-2 flex-1">
      <h1 class="text-base sm:text-lg font-bold text-text-primary line-height-fixed-32">
        {{ title }}
      </h1>
    </div>

    <div class="flex items-center gap-2">
      <div class="flex items-center rounded-md bg-bg-secondary p-0.5" role="group" :aria-label="t('article.imageGallery.mediaFilter')">
        <button
          v-for="filter in (['all', 'images', 'videos'] as MediaTypeFilter[])"
          :key="filter"
          type="button"
          class="cursor-pointer rounded px-2 py-1 text-xs text-text-secondary transition-colors hover:bg-bg-tertiary hover:text-text-primary"
          :class="mediaType === filter ? 'bg-bg-tertiary text-text-primary' : ''"
          @click="emit('updateMediaType', filter)"
        >
          {{ t(`article.imageGallery.filter.${filter}`) }}
        </button>
      </div>

      <!-- Show only unread toggle button -->
      <button
        class="p-1 sm:p-1.5 rounded hover:bg-bg-tertiary text-text-secondary transition-colors cursor-pointer"
        :class="showOnlyUnread ? 'text-accent' : ''"
        :title="
          showOnlyUnread
            ? t('setting.reading.showAllArticles')
            : t('setting.reading.showOnlyUnread')
        "
        @click="emit('toggleShowOnlyUnread')"
      >
        <PhCircle :size="20" :weight="showOnlyUnread ? 'fill' : 'regular'" />
      </button>

      <!-- Toggle text overlay button -->
      <button
        class="p-1 sm:p-1.5 rounded hover:bg-bg-tertiary text-text-secondary transition-colors cursor-pointer"
        :title="showTextOverlay ? t('setting.reading.hideText') : t('setting.reading.showText')"
        @click="emit('toggleTextOverlay')"
      >
        <PhTextTSlash v-if="showTextOverlay" :size="20" />
        <PhTextT v-else :size="20" />
      </button>

      <button
        class="p-1 sm:p-1.5 rounded hover:bg-bg-tertiary text-text-secondary transition-colors cursor-pointer"
        :title="t('article.imageGallery.markAllRead')"
        @click="emit('markAllRead')"
      >
        <PhCheckCircle :size="20" />
      </button>

      <button
        class="p-1 sm:p-1.5 rounded hover:bg-bg-tertiary text-text-secondary transition-colors cursor-pointer disabled:cursor-not-allowed disabled:opacity-60"
        :title="t('article.action.refresh')"
        :disabled="isRefreshing"
        @click="emit('refresh')"
      >
        <PhArrowClockwise :size="20" :class="isRefreshing ? 'animate-spin' : ''" />
      </button>
    </div>
  </div>
</template>
