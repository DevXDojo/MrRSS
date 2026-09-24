<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import BaseSelect from '@/components/common/BaseSelect.vue';
import type { SelectOption, SelectOptionGroup } from '@/types/select';
import { getRecommendedFonts, resolveFontFamily } from '@/utils/fontDetector';

const props = defineProps<{
  modelValue: string;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: string | number];
}>();

const { t } = useI18n();

const availableFonts = ref({
  serif: [] as string[],
  sansSerif: [] as string[],
  monospace: [] as string[],
});

const fontOptions = computed<SelectOptionGroup[]>(() => {
  const groups: SelectOptionGroup[] = [
    {
      label: t('setting.typography.fontSystem'),
      options: [
        {
          value: 'system',
          label: t('setting.typography.fontSystemDefault'),
        },
      ],
    },
  ];

  if (availableFonts.value.serif.length > 0) {
    const options: SelectOption[] = [
      {
        value: 'serif',
        label: t('setting.typography.fontSerifDefault'),
      },
      ...availableFonts.value.serif.map((font) => ({
        value: font,
        label: font,
        style: { fontFamily: resolveFontFamily(font) },
      })),
    ];
    groups.push({ label: t('setting.typography.fontSerif'), options });
  }

  if (availableFonts.value.sansSerif.length > 0) {
    const options: SelectOption[] = [
      {
        value: 'sans-serif',
        label: t('setting.typography.fontSansSerifDefault'),
      },
      ...availableFonts.value.sansSerif.map((font) => ({
        value: font,
        label: font,
        style: { fontFamily: resolveFontFamily(font) },
      })),
    ];
    groups.push({ label: t('setting.typography.fontSansSerif'), options });
  }

  if (availableFonts.value.monospace.length > 0) {
    const options: SelectOption[] = [
      {
        value: 'monospace',
        label: t('setting.typography.fontMonospaceDefault'),
      },
      ...availableFonts.value.monospace.map((font) => ({
        value: font,
        label: font,
        style: { fontFamily: resolveFontFamily(font) },
      })),
    ];
    groups.push({ label: t('setting.typography.fontMonospace'), options });
  }

  if (props.modelValue && !groups.some((group) => group.options.some((option) => option.value === props.modelValue))) {
    groups.push({
      label: t('setting.typography.fontCustom'),
      options: [{ value: props.modelValue, label: props.modelValue, style: { fontFamily: resolveFontFamily(props.modelValue) } }],
    });
  }
  return groups;
});

onMounted(() => {
  try {
    availableFonts.value = getRecommendedFonts();
  } catch (error) {
    console.error('Failed to detect system fonts:', error);
  }
});
</script>

<template>
  <BaseSelect
    :model-value="modelValue"
    :options="fontOptions"
    :searchable="true"
    :allow-custom-input="true"
    :custom-input-placeholder="t('setting.typography.fontCustomPlaceholder')"
    width="w-36 sm:w-48"
    max-height="max-h-60"
    @update:model-value="emit('update:modelValue', $event)"
    @custom-input="emit('update:modelValue', $event)"
  >
    <template #option="{ option }">
      <span :style="option.style">{{ option.label }}</span>
    </template>
  </BaseSelect>
</template>
