<script setup lang="ts">
import type { Component } from 'vue';
import SettingItem from '../base/SettingItem.vue';
import SelectControl from '../base/SettingControl/SelectControl.vue';

interface Option {
  value: string | number;
  label: string;
  disabled?: boolean;
}

interface Props {
  icon?: Component;
  title: string;
  description?: string;
  modelValue: string | number;
  options: Array<Option>;
  required?: boolean;
  disabled?: boolean;
  width?: string;
}

defineProps<Props>();

const emit = defineEmits<{
  'update:modelValue': [value: string | number];
}>();
</script>

<template>
  <SettingItem :icon="icon" :title="title" :description="description" :required="required">
    <SelectControl
      :model-value="modelValue"
      :options="options"
      :disabled="disabled"
      :width="width"
      @update:model-value="emit('update:modelValue', $event)"
    />
  </SettingItem>
</template>
