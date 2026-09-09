<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhArrowClockwise, PhCircle, PhList, PhTextT, PhTextTSlash } from '@phosphor-icons/vue';

interface Props {
  title: string;
  isRefreshing: boolean;
  showTextOverlay: boolean;
  showOnlyUnread: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  toggleSidebar: [];
  refresh: [];
  toggleTextOverlay: [];
  toggleShowOnlyUnread: [];
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
