<script setup lang="ts">
// Container component for nested settings with left border and spacing
// This replaces the repetitive div: class="ml-2 sm:ml-4 mt-2 sm:mt-3 space-y-2 sm:space-y-3 border-l-2 border-border pl-2 sm:pl-4"
</script>

<template>
  <div class="nested-settings-container">
    <slot />
  </div>
</template>

<style scoped>
@reference "../../../style.css";
.nested-settings-container {
  @apply ml-2 sm:ml-4 mt-2 sm:mt-3 space-y-2 sm:space-y-3 border-l-2 border-border pl-2 sm:pl-4;
}
</style>
