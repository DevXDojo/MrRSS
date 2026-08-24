<script setup lang="ts">
import type { Component } from 'vue';
import SettingItem from '../base/SettingItem.vue';
import ToggleControl from '../base/SettingControl/ToggleControl.vue';

interface Props {
  icon?: Component;
  title: string;
  description?: string;
  modelValue: boolean;
  required?: boolean;
  disabled?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
}>();
</script>

<template>
  <SettingItem :icon="icon" :title="title" :description="description" :required="required">
    <ToggleControl
      :model-value="modelValue"
      :disabled="disabled"
      @update:model-value="emit('update:modelValue', $event)"
    />
  </SettingItem>
</template>
