<script setup lang="ts">
import { computed, ref, watch, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhClock, PhTranslate } from '@phosphor-icons/vue';
import { translationQueue } from '@/utils/translationRequest';

const { t } = useI18n();
const seconds = ref(0);
const visible = computed(
  () => translationQueue.value.pending > 0 && translationQueue.value.recovering
);
let timer: ReturnType<typeof setInterval> | undefined;

watch(
  () => [translationQueue.value.pending, translationQueue.value.retryAt],
  () => {
    if (timer) clearInterval(timer);
    timer = undefined;
    const update = () => {
      seconds.value = Math.max(0, Math.ceil((translationQueue.value.retryAt - Date.now()) / 1000));
      if (!seconds.value && timer) {
        clearInterval(timer);
        timer = undefined;
      }
    };
    update();
    if (translationQueue.value.pending && seconds.value) timer = setInterval(update, 1000);
  },
  { immediate: true }
);

onBeforeUnmount(() => {
  if (timer) clearInterval(timer);
});
</script>

<template>
  <div v-if="visible" class="flex items-start gap-2 text-xs text-text-secondary" role="status">
    <PhClock v-if="seconds" :size="14" class="mt-0.5 shrink-0" />
    <PhTranslate v-else :size="14" class="mt-0.5 shrink-0" />
    <div>
      <span>{{ t(seconds ? 'article.translation.waiting' : 'article.translation.resuming') }}</span>
      <span v-if="seconds" aria-hidden="true" aria-live="off" class="ml-1">
        {{ t('article.translation.retryCountdown', { seconds }) }}
      </span>
    </div>
  </div>
</template>
