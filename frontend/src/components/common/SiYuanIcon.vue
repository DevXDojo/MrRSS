<script setup lang="ts">
withDefaults(defineProps<{ size?: number }>(), { size: 20 });
</script>
<template>
  <img
    src="/assets/plugin_icons/siyuan.svg"
    :width="size"
    :height="size"
    alt=""
    aria-hidden="true"
  />
</template>
