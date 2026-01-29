<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useSettings } from '@/composables/core/useSettings';
import { onMounted } from 'vue';
import {
  PhArrowLeft,
  PhX,
  PhGlobe,
  PhArticle,
  PhEnvelopeOpen,
  PhEnvelope,
  PhStar,
  PhClockCountdown,
  PhArrowSquareOut,
  PhTranslate,
} from '@phosphor-icons/vue';
import type { Article } from '@/types/models';

const { t } = useI18n();
const { settings, fetchSettings } = useSettings();

onMounted(async () => {
  try {
    await fetchSettings();
  } catch (e) {
    console.error('Error loading settings:', e);
  }
});

interface Props {
  article: Article;
  showContent: boolean;
  showTranslations?: boolean;
  isModal?: boolean;
}

withDefaults(defineProps<Props>(), {
  showTranslations: true,
  isModal: false,
});

defineEmits<{
  close: [];
  toggleContentView: [];
  toggleRead: [];
  toggleFavorite: [];
  toggleReadLater: [];
  openOriginal: [];
  toggleTranslations: [];
  exportToObsidian: [];
}>();
</script>

<template>
  <div
    class="p-2 sm:p-4 border-b border-border flex justify-between items-center bg-bg-primary shrink-0"
  >
    <!-- Modal mode: X button always visible -->
    <button
      v-if="isModal"
      class="flex items-center gap-1.5 sm:gap-2 text-text-secondary hover:text-text-primary text-sm sm:text-base"
      :title="t('common.close')"
      @click="$emit('close')"
    >
      <PhX :size="20" class="sm:w-5 sm:h-5" />
    </button>
    <!-- Normal mode: Back button on mobile -->
    <button
      v-else
      class="md:hidden flex items-center gap-1.5 sm:gap-2 text-text-secondary hover:text-text-primary text-sm sm:text-base"
      @click="$emit('close')"
    >
      <PhArrowLeft :size="18" class="sm:w-5 sm:h-5" />
      <span class="hidden xs:inline">{{ t('common.back') }}</span>
    </button>
    <div class="flex gap-1 sm:gap-2 ml-auto">
      <button
        class="action-btn"
        :title="showContent ? t('article.action.viewOriginal') : t('article.action.viewContent')"
        @click="$emit('toggleContentView')"
      >
        <PhGlobe v-if="showContent" :size="18" class="sm:w-5 sm:h-5" />
        <PhArticle v-else :size="18" class="sm:w-5 sm:h-5" />
      </button>
      <button
        v-if="showContent && settings.translation_enabled && !settings.translation_only_mode"
        class="action-btn"
        :title="
          showTranslations
            ? t('setting.reading.hideTranslations')
            : t('setting.reading.showTranslations')
        "
        @click="$emit('toggleTranslations')"
      >
        <PhTranslate
          :size="18"
          class="sm:w-5 sm:h-5"
          :weight="showTranslations ? 'fill' : 'regular'"
        />
      </button>
      <button
        class="action-btn"
        :title="article.is_read ? t('article.action.markAsUnread') : t('article.action.markAsRead')"
        @click="$emit('toggleRead')"
      >
        <PhEnvelopeOpen v-if="article.is_read" :size="18" class="sm:w-5 sm:h-5" />
        <PhEnvelope v-else :size="18" class="sm:w-5 sm:h-5" />
      </button>
      <button
        :class="[
          'action-btn',
          article.is_favorite ? 'text-yellow-500 hover:text-yellow-600' : 'hover:text-yellow-500',
        ]"
        :title="
          article.is_favorite
            ? t('article.action.removeFromFavorite')
            : t('article.toolbar.addToFavorite')
        "
        @click="$emit('toggleFavorite')"
      >
        <PhStar
          :size="18"
          class="sm:w-5 sm:h-5"
          :weight="article.is_favorite ? 'fill' : 'regular'"
        />
      </button>
      <button
        :class="[
          'action-btn',
          article.is_read_later ? 'text-blue-500 hover:text-blue-600' : 'hover:text-blue-500',
        ]"
        :title="
          article.is_read_later
            ? t('article.action.removeFromReadLater')
            : t('article.toolbar.addToReadLater')
        "
        @click="$emit('toggleReadLater')"
      >
        <PhClockCountdown
          :size="18"
          class="sm:w-5 sm:h-5"
          :weight="article.is_read_later ? 'fill' : 'regular'"
        />
      </button>
      <button
        class="action-btn"
        :title="t('article.action.openInBrowser')"
        @click="$emit('openOriginal')"
      >
        <PhArrowSquareOut :size="18" class="sm:w-5 sm:h-5" />
      </button>
      <button
        v-if="settings.obsidian_enabled"
        class="action-btn"
        :title="t('setting.plugins.obsidian.exportTo')"
        @click="$emit('exportToObsidian')"
      >
        <img
          src="/assets/plugin_icons/obsidian.svg"
          class="w-[18px] h-[18px] sm:w-5 sm:h-5"
          alt="Obsidian"
        />
      </button>
    </div>
  </div>
</template>

<style scoped>
@reference "../../style.css";

.action-btn {
  @apply text-lg sm:text-xl cursor-pointer text-text-secondary p-1 sm:p-1.5 rounded-md transition-colors hover:bg-bg-tertiary hover:text-text-primary;
}
</style>
