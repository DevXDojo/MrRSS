<script setup lang="ts">
import { computed } from 'vue';

interface ButtonAction {
  label: string;
  type?: 'primary' | 'secondary' | 'danger' | 'ghost';
  disabled?: boolean;
  loading?: boolean;
  onClick?: () => void;
}

interface Props {
  // Button alignment
  align?: 'left' | 'center' | 'right' | 'space-between';
  // Buttons configuration
  primaryButton?: ButtonAction;
  secondaryButton?: ButtonAction;
  dangerButton?: ButtonAction;
  // Additional classes
  class?: string;
  // Show as text (no background)
  textOnly?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  align: 'right',
  primaryButton: undefined,
  secondaryButton: undefined,
  dangerButton: undefined,
  class: '',
  textOnly: false,
});

const emit = defineEmits<{
  primaryClick: [];
  secondaryClick: [];
  dangerClick: [];
}>();

// Computed alignment classes
const alignClasses = computed(() => {
  const alignMap = {
    left: 'justify-start',
    center: 'justify-center',
    right: 'justify-end',
    'space-between': 'justify-between',
  };
  return alignMap[props.align];
});

// Button type classes
function getButtonClasses(type: 'primary' | 'secondary' | 'danger' | 'ghost' = 'primary') {
  const baseClasses = [
    'text-sm',
    'sm:text-base',
    'px-4',
    'sm:px-5',
    'py-2',
    'sm:py-2.5',
    'rounded-lg',
    'cursor-pointer',
    'font-semibold',
    'transition-colors',
    'disabled:opacity-70',
    'disabled:cursor-not-allowed',
  ];

  const typeClasses = {
    primary: 'bg-accent text-white border-none hover:bg-accent-hover',
    secondary: 'bg-transparent border border-border text-text-primary hover:bg-bg-tertiary',
    danger:
      'bg-transparent border border-red-300 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 dark:border-red-400 dark:text-red-400',
    ghost:
      'bg-transparent border-none text-text-primary hover:text-text-secondary hover:bg-bg-tertiary',
  };

  return [...baseClasses, ...typeClasses[type].split(' ')];
}

// Handle button clicks
function handlePrimaryClick() {
  if (!props.primaryButton?.disabled && !props.primaryButton?.loading) {
    if (props.primaryButton?.onClick) {
      props.primaryButton.onClick();
    }
    emit('primaryClick');
  }
}

function handleSecondaryClick() {
  if (!props.secondaryButton?.disabled && !props.secondaryButton?.loading) {
    if (props.secondaryButton?.onClick) {
      props.secondaryButton.onClick();
    }
    emit('secondaryClick');
  }
}

function handleDangerClick() {
  if (!props.dangerButton?.disabled && !props.dangerButton?.loading) {
    if (props.dangerButton?.onClick) {
      props.dangerButton.onClick();
    }
    emit('dangerClick');
  }
}
</script>

<template>
  <div :class="['flex flex-col-reverse sm:flex-row gap-2 sm:gap-3', alignClasses, props.class]">
    <!-- Left side buttons (for space-between alignment) -->
    <slot name="left"></slot>

    <!-- Secondary button (shown first on mobile, left on desktop) -->
    <button
      v-if="secondaryButton"
      :class="getButtonClasses('secondary')"
      :disabled="secondaryButton.disabled || secondaryButton.loading"
      @click="handleSecondaryClick"
    >
      {{ secondaryButton.loading ? '...' : secondaryButton.label }}
    </button>

    <!-- Danger button -->
    <button
      v-if="dangerButton"
      :class="getButtonClasses('danger')"
      :disabled="dangerButton.disabled || dangerButton.loading"
      @click="handleDangerClick"
    >
      {{ dangerButton.loading ? '...' : dangerButton.label }}
    </button>

    <!-- Primary button (shown last on mobile, right on desktop) -->
    <button
      v-if="primaryButton"
      :class="getButtonClasses(primaryButton.type || 'primary')"
      :disabled="primaryButton.disabled || primaryButton.loading"
      @click="handlePrimaryClick"
    >
      {{ primaryButton.loading ? '...' : primaryButton.label }}
    </button>

    <!-- Right side buttons (custom content) -->
    <slot name="right"></slot>
  </div>
</template>

<style scoped>
@reference "../../style.css";

button {
  box-sizing: border-box;
}
</style>
