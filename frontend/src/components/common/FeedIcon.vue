<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { PhRss } from '@phosphor-icons/vue';
import type { Feed } from '@/types/models';

const props = defineProps<{ feed: Feed; lazy?: boolean }>();
const attempt = ref(0);
const generation = ref(0);
const image = ref<HTMLImageElement | null>(null);

const sources = computed(() => {
  let favicon = '';
  // Use the parsed website link for hosted feeds such as RSSHub.
  for (const candidate of [props.feed.link, props.feed.website_url, props.feed.url]) {
    if (!candidate) continue;
    try {
      const url = new URL(candidate);
      if (!['http:', 'https:'].includes(url.protocol)) continue;
      favicon = `https://www.google.com/s2/favicons?domain=${encodeURIComponent(url.hostname)}`;
      break;
    } catch {
      // Try the next website/feed URL.
    }
  }
  return [...new Set([props.feed.image_url, favicon].filter((url): url is string => !!url))];
});
const source = computed(() => sources.value[attempt.value] || '');
const imageKey = computed(() => `${generation.value}:${attempt.value}`);

function reset() {
  attempt.value = 0;
  generation.value++;
}

function retryFailed() {
  if (attempt.value > 0) reset();
}

function handleError(event: Event) {
  // An error from a replaced image must not discard its replacement.
  if (event.target !== image.value || image.value?.dataset.attempt !== imageKey.value) return;
  attempt.value++;
}

watch([() => props.feed.id, () => JSON.stringify(sources.value)], reset);
watch(() => props.feed.last_updated || props.feed.last_fetched_at, retryFailed);
onMounted(() => window.addEventListener('online', retryFailed));
onUnmounted(() => window.removeEventListener('online', retryFailed));
</script>

<template>
  <span class="inline-flex items-center justify-center shrink-0" aria-hidden="true">
    <img
      v-if="source"
      :key="imageKey"
      ref="image"
      :data-attempt="imageKey"
      :src="source"
      alt=""
      class="w-full h-full object-contain"
      :loading="lazy ? 'lazy' : undefined"
      @error="handleError"
    />
    <PhRss v-else class="w-4/5 h-4/5 text-text-tertiary" data-testid="feed-icon-fallback" />
  </span>
</template>
