<script setup lang="ts">
import { useI18n } from 'vue-i18n';

interface Props {
  onClick: () => void;
}

defineProps<Props>();

const { t } = useI18n();
</script>

<template>
  <button
    class="js-article-chat-button fixed bottom-12 right-6 w-14 h-14 rounded-full shadow-lg transition-all hover:scale-110 z-40 overflow-hidden"
    :title="t('article.chat.aiChat')"
    @click="onClick"
  >
    <img src="/assets/logo-circle.png" alt="" class="w-full h-full object-cover" />
  </button>
</template>
