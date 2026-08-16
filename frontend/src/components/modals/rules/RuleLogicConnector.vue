<script setup lang="ts">
import { useI18n } from 'vue-i18n';

interface Props {
  logic: 'and' | 'or' | null;
}

defineProps<Props>();

const emit = defineEmits<{
  update: [logic: 'and' | 'or'];
}>();

const { t } = useI18n();

const logicOptions: Array<{ value: 'and' | 'or'; labelKey: string }> = [
  { value: 'and', labelKey: 'and' },
  { value: 'or', labelKey: 'or' },
];
</script>

<template>
  <div class="flex items-center justify-center my-3">
    <div class="flex-1 h-px bg-border"></div>
    <div class="logic-connector mx-3">
      <button
        v-for="opt in logicOptions"
        :key="opt.value"
        :class="['logic-btn', logic === opt.value ? 'active' : '']"
        @click="emit('update', opt.value)"
      >
        {{ t(opt.labelKey) }}
      </button>
    </div>
    <div class="flex-1 h-px bg-border"></div>
  </div>
</template>

<style scoped>
@reference "../../../style.css";
.logic-connector {
  @apply flex items-center gap-1 bg-bg-tertiary rounded-full p-1;
}

.logic-btn {
  @apply px-3 py-1 text-xs font-bold rounded-full transition-all cursor-pointer;
  @apply text-text-secondary bg-transparent;
}

.logic-btn:hover {
  @apply text-text-primary bg-bg-secondary;
}

.logic-btn.active {
  @apply text-white bg-accent;
}
</style>
