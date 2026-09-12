<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhCircle, PhStar, PhClock } from '@phosphor-icons/vue';
import { useArticleDateFormat } from '@/composables/article/useArticleDateFormat';
import { useArticleHoverRead } from '@/composables/article/useArticleHoverRead';
import { useSettings } from '@/composables/core/useSettings';
import type { Article } from '@/types/models';
import type { ArticleTableColumn } from '@/utils/articleTable';

const props = defineProps<{
  article: Article;
  columns: ArticleTableColumn[];
  isActive: boolean;
  disabled?: boolean;
}>();
const emit = defineEmits<{
  click: [];
  contextmenu: [event: MouseEvent];
  observeElement: [element: Element | null];
  hoverMarkAsRead: [id: number];
}>();
const { t } = useI18n();
const { settings } = useSettings();
const { formatArticleDate, formatArticleDateTime } = useArticleDateFormat();
const hover = useArticleHoverRead(
  () => props.article,
  (id) => emit('hoverMarkAsRead', id),
  () => props.disabled === true
);
const hasTranslation = computed(
  () => props.article.translated_title && props.article.translated_title !== props.article.title
);
</script>

<template>
  <tr
    :ref="(element) => emit('observeElement', element as Element | null)"
    :data-article-id="article.id"
    :class="[
      'cursor-pointer border-b border-border hover:bg-bg-secondary',
      isActive ? 'bg-accent/10' : '',
      article.is_read ? 'text-text-secondary' : 'text-text-primary font-medium',
    ]"
    @click="emit('click')"
    @contextmenu="emit('contextmenu', $event)"
    @mouseenter="hover.enter"
    @mouseleave="hover.leave"
  >
    <td v-for="column in columns" :key="column" class="px-3 py-2 text-sm align-middle">
      <button
        v-if="column === 'title'"
        class="block w-full truncate text-left focus-visible:outline-2 focus-visible:outline-accent"
        :title="article.title"
        :aria-current="isActive ? 'true' : undefined"
        @click.stop="emit('click')"
      >
        <span v-if="hasTranslation">{{ article.translated_title }}</span>
        <span
          v-if="!hasTranslation || !settings.translation_only_mode"
          :class="hasTranslation ? 'ml-2 text-text-secondary font-normal' : ''"
          >{{ article.title }}</span
        >
      </button>
      <span v-else-if="column === 'feed'" class="block truncate" :title="article.feed_title">{{
        article.feed_title
      }}</span>
      <span v-else-if="column === 'author'" class="block truncate" :title="article.author">{{
        article.author || '—'
      }}</span>
      <time
        v-else-if="column === 'date'"
        class="block truncate"
        :datetime="article.published_at"
        :title="formatArticleDateTime(article.published_at)"
        >{{ formatArticleDate(article.published_at) }}</time
      >
      <span v-else-if="column === 'status'" class="flex items-center gap-2">
        <PhCircle
          :size="14"
          :weight="article.is_read ? 'regular' : 'fill'"
          :aria-label="t(article.is_read ? 'article.table.read' : 'article.table.unread')"
          role="img"
        />
        <PhStar
          v-if="article.is_favorite"
          :size="14"
          weight="fill"
          class="text-yellow-500"
          :aria-label="t('sidebar.activity.favorites')"
          role="img"
        />
        <PhClock
          v-if="article.is_read_later"
          :size="14"
          class="text-accent"
          :aria-label="t('sidebar.activity.readLater')"
          role="img"
        />
      </span>
    </td>
  </tr>
</template>
