<script setup lang="ts">
import { computed } from 'vue';
import { useAppStore } from '@/stores/app';
const props = defineProps<{ provider?: string }>();
const store = useAppStore();
const source = computed(() =>
  props.provider === 'miniflux'
    ? `/assets/plugin_icons/miniflux${store.theme === 'dark' ? '-dark' : ''}.svg`
    : '/assets/plugin_icons/freshrss.svg'
);
</script>
<template>
  <img :src="source" :alt="provider === 'miniflux' ? 'Miniflux' : 'FreshRSS'" />
</template>
