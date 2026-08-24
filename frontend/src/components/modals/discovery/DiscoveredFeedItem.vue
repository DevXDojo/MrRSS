<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhCheck, PhGlobe, PhRss } from '@phosphor-icons/vue';

const { t } = useI18n();

interface RecentArticle {
  title: string;
  date?: string;
}

interface DiscoveredFeed {
  name: string;
  homepage: string;
  rss_feed: string;
  icon_url?: string;
  recent_articles?: Array<RecentArticle | string>;
}

interface Props {
  feed: DiscoveredFeed;
  isSelected: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  toggle: [];
}>();

function handleImageError(event: Event): void {
  const target = event.target as HTMLImageElement;
  if (target) {
    target.style.display = 'none';
  }
}
</script>

<template>
  <div
    :class="[
      'border rounded-xl p-4 cursor-pointer transition-all duration-200',
      isSelected
        ? 'bg-accent/10 border-accent ring-2 ring-accent/20 shadow-md'
        : 'bg-bg-secondary hover:bg-bg-tertiary border-border hover:shadow-sm',
    ]"
    @click="emit('toggle')"
  >
    <div class="flex items-start gap-4">
      <!-- Checkbox -->
      <div class="mt-1 shrink-0">
        <div
          :class="[
            'w-5 h-5 rounded border-2 flex items-center justify-center transition-all',
            isSelected ? 'bg-accent border-accent scale-110' : 'border-border bg-bg-primary',
          ]"
        >
          <PhCheck v-if="isSelected" :size="14" weight="bold" class="text-white" />
        </div>
      </div>

      <!-- Feed Info -->
      <div class="flex-1 min-w-0">
        <div class="flex items-start gap-3 mb-3">
          <div
            class="shrink-0 w-10 h-10 rounded-lg overflow-hidden bg-bg-primary border border-border flex items-center justify-center"
          >
            <img
              v-if="feed.icon_url"
              :src="feed.icon_url"
              class="w-full h-full object-cover"
              :alt="feed.name"
              @error="handleImageError"
            />
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="font-semibold text-text-primary truncate text-base">{{ feed.name }}</h3>
            <a
              :href="feed.homepage"
              target="_blank"
              class="text-xs text-accent hover:text-accent-hover flex items-center gap-1 mt-1 hover:underline"
              @click.stop
            >
              <PhGlobe :size="14" />
              <span class="truncate">{{ feed.homepage }}</span>
            </a>
          </div>
        </div>

        <!-- Recent Articles -->
        <div v-if="feed.recent_articles && feed.recent_articles.length > 0" class="mt-3">
          <p class="text-xs font-semibold text-text-secondary mb-2 flex items-center gap-1">
            <PhRss :size="12" />
            {{ t('sidebar.feedList.recentArticles') }}
          </p>
          <div class="space-y-1.5">
            <div
              v-for="(article, aIndex) in feed.recent_articles"
              :key="aIndex"
              class="flex flex-col gap-0.5 py-1.5 border-l-2 border-border pl-2"
            >
              <span class="text-sm text-text-primary line-clamp-2 leading-snug">
                {{ typeof article === 'string' ? article : article.title }}
              </span>
              <span
                v-if="typeof article !== 'string' && article.date"
                class="text-xs text-text-tertiary"
              >
                {{ article.date }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
