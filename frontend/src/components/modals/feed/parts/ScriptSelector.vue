<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhCode, PhBookOpen } from '@phosphor-icons/vue';
import BaseSelect from '@/components/common/BaseSelect.vue';
import { openInBrowser } from '@/utils/browser';
import type { SelectOption } from '@/types/select';

interface Props {
  modelValue: string;
  mode: 'add' | 'edit';
  isInvalid?: boolean;
  availableScripts: Array<{ path: string; name: string; type: string }>;
  scriptsDir: string;
}

const props = withDefaults(defineProps<Props>(), {
  isInvalid: false,
});

const emit = defineEmits<{
  'update:modelValue': [value: string];
  'open-scripts-folder': [];
}>();

const { t, locale } = useI18n();

// Build options for BaseSelect
const scriptOptions = computed<SelectOption[]>(() => {
  return [
    { value: '', label: t('setting.customization.selectScriptPlaceholder') },
    ...props.availableScripts.map((script) => ({
      value: script.path,
      label: `${script.name} (${script.type})`,
    })),
  ];
});

function openScriptsFolder() {
  emit('open-scripts-folder');
}

function openDocumentation() {
  const docUrl = locale.value.startsWith('zh')
    ? 'https://github.com/DevXDojo/MrRSS/blob/main/docs/CUSTOM_SCRIPT_MODE.zh.md'
    : 'https://github.com/DevXDojo/MrRSS/blob/main/docs/CUSTOM_SCRIPT_MODE.md';
  openInBrowser(docUrl);
}
</script>

<template>
  <div class="mb-3 sm:mb-4">
    <label class="block mb-1 sm:mb-1.5 font-semibold text-xs sm:text-sm text-text-secondary"
      >{{ t('setting.customization.selectScript') }}
      <span v-if="props.mode === 'add'" class="text-red-500">*</span></label
    >
    <div v-if="props.availableScripts.length > 0" class="mb-2">
      <BaseSelect
        :model-value="props.modelValue"
        :options="scriptOptions"
        :class="{ 'border-red-500': props.mode === 'add' && props.isInvalid }"
        :searchable="true"
        @update:model-value="emit('update:modelValue', String($event))"
      />
    </div>
    <div
      v-else
      class="text-xs sm:text-sm text-text-secondary bg-bg-secondary rounded-md p-2 sm:p-3 border border-border"
    >
      <p class="mb-2">{{ t('setting.customization.scriptsNotFound') }}</p>
    </div>
    <div class="flex flex-col sm:flex-row gap-2 sm:gap-3 mt-3">
      <button
        type="button"
        class="text-xs sm:text-sm text-accent hover:underline flex items-center gap-1"
        @click="openDocumentation"
      >
        <PhBookOpen :size="14" />
        {{ t('setting.customization.scriptDoc') }}
      </button>
      <button
        type="button"
        class="text-xs sm:text-sm text-accent hover:underline flex items-center gap-1"
        @click="openScriptsFolder"
      >
        <PhCode :size="14" />
        {{ t('setting.customization.scriptsFolder') }}
      </button>
    </div>
  </div>
</template>

<style scoped>
/* Styles are now handled by BaseSelect and select.css */
</style>
