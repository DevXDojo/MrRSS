<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted, type PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhCaretDown, PhX, PhPlus } from '@phosphor-icons/vue';
import {
  type SelectOption,
  type SelectOptionGroup,
  type SelectOptions,
  isOptionGroup,
  flattenOptions,
  findOption,
} from '@/types/select';
import { useSelect } from '@/composables/ui/useSelect';
import './select.css';

const { t } = useI18n();

// Generate unique ID for this dropdown instance
const dropdownId = `select-${Math.random().toString(36).substring(2, 9)}`;

const props = defineProps({
  modelValue: {
    type: [String, Number] as PropType<string | number>,
    required: true,
  },
  options: {
    type: Array as PropType<SelectOptions>,
    required: true,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  placeholder: {
    type: String,
    default: '',
  },
  width: {
    type: String,
    default: '',
  },
  maxWidth: {
    type: String,
    default: '',
  },
  maxHeight: {
    type: String,
    default: '',
  },
  searchable: {
    type: Boolean,
    default: false, // Reserved for future implementation
  },
  clearable: {
    type: Boolean,
    default: false,
  },
  position: {
    type: String as PropType<'bottom' | 'top' | 'auto'>,
    default: 'bottom',
  },
  size: {
    type: String as PropType<'xs' | 'sm' | 'md'>,
    default: 'sm',
  },
  allowCustomInput: {
    type: Boolean,
    default: false,
  },
  customInputPlaceholder: {
    type: String,
    default: '',
  },
  bgMode: {
    type: String as PropType<'primary' | 'secondary'>,
    default: 'primary',
  },
  allowAdd: {
    type: Boolean,
    default: false,
  },
  addPlaceholder: {
    type: String,
    default: '',
  },
});

const emit = defineEmits<{
  'update:modelValue': [value: string | number];
  'custom-input': [value: string];
  add: [value: string];
}>();

const isOpen = ref(false);
const customInputValue = ref('');
const customInputRef = ref<HTMLInputElement>();
const searchQuery = ref('');
const searchInputRef = ref<HTMLInputElement>();
const addInputValue = ref('');
const addInputRef = ref<HTMLInputElement>();

// Reactive refs for computed values passed to useSelect
const positionRef = computed(() => props.position);
const widthRef = computed(() => props.width);
const maxWidthRef = computed(() => props.maxWidth);
const desiredMaxHeightRef = computed(() => props.maxHeight);

// Filter options based on search query
const filteredOptions = computed(() => {
  if (!props.searchable || !searchQuery.value.trim()) {
    return props.options;
  }

  const query = searchQuery.value.toLowerCase().trim();
  const result: SelectOptions = [];

  for (const item of props.options) {
    if (isOptionGroup(item)) {
      // Filter options within group
      const filteredGroupOptions = item.options.filter((option) =>
        option.label.toLowerCase().includes(query)
      );
      // Only include group if it has matching options
      if (filteredGroupOptions.length > 0) {
        (result as SelectOptionGroup[]).push({
          label: item.label,
          options: filteredGroupOptions,
        });
      }
    } else {
      // Filter single option
      if (item.label.toLowerCase().includes(query)) {
        (result as SelectOption[]).push(item);
      }
    }
  }

  return result;
});

// Flatten filtered options for navigation
const filteredFlatOptions = computed(() => flattenOptions(filteredOptions.value));

// Find current selected option
const selectedOption = computed(() => findOption(props.options, props.modelValue));

// Display text for trigger button
const displayText = computed(() => {
  if (selectedOption.value) {
    return selectedOption.value.label;
  }
  return props.placeholder || t('common.select.placeholder');
});

// Use select composable for shared functionality
const {
  selectedIndex,
  triggerRef,
  dropdownRef,
  dropdownPositionStyle,
  widthClass,
  maxWidthStyle,
  shouldTeleport,
  resetIndex,
  registerAsOpen,
  unregisterAsOpen,
  handleMouseLeave,
  updateDropdownPosition,
  setupScrollListener,
  cleanupScrollListener,
} = useSelect({
  options: [],
  isOpen,
  dropdownId,
  position: positionRef,
  width: widthRef,
  maxWidth: maxWidthRef,
  desiredMaxHeight: desiredMaxHeightRef,
});

// Use dropdown position style directly (maxHeight is calculated dynamically)
const dropdownStyle = computed(() => dropdownPositionStyle.value);

// Reset selected index when search query changes
watch(searchQuery, () => {
  resetIndex();
});

// Toggle dropdown
function toggleDropdown() {
  if (props.disabled) return;
  isOpen.value = !isOpen.value;
  if (isOpen.value) {
    registerAsOpen();
    resetIndex();
    // Set up scroll listener for the container
    setupScrollListener();
    // Calculate position after DOM update
    nextTick(updateDropdownPosition);
  } else {
    unregisterAsOpen();
    cleanupScrollListener();
  }
}

// Select an option
function selectOption(option: SelectOption) {
  if (option.disabled) return;
  emit('update:modelValue', option.value);
  isOpen.value = false;
  unregisterAsOpen();
  cleanupScrollListener();
}

// Clear selection
function clearSelection(event: MouseEvent) {
  event.stopPropagation();
  emit('update:modelValue', '');
}

// Handle custom input
function handleCustomInput() {
  if (customInputValue.value.trim()) {
    emit('custom-input', customInputValue.value.trim());
    customInputValue.value = '';
    isOpen.value = false;
    unregisterAsOpen();
    cleanupScrollListener();
  }
}

// Handle add new option
function handleAddOption() {
  if (addInputValue.value.trim()) {
    emit('add', addInputValue.value.trim());
    addInputValue.value = '';
    // Keep dropdown open for adding more options
    nextTick(() => {
      addInputRef.value?.focus();
    });
  }
}

// Handle click on trigger
function handleTriggerClick(event: MouseEvent) {
  event.stopPropagation();
  toggleDropdown();
}

// Watch for model value changes from parent
watch(
  () => props.modelValue,
  () => {
    // Reset when value changes externally
  }
);

// Focus custom input when dropdown opens with allowCustomInput
watch(isOpen, (open) => {
  if (open && props.allowCustomInput) {
    nextTick(() => {
      customInputRef.value?.focus();
    });
  } else if (open && props.allowAdd) {
    nextTick(() => {
      addInputRef.value?.focus();
    });
  } else if (open && props.searchable) {
    nextTick(() => {
      searchInputRef.value?.focus();
    });
  } else if (!open) {
    // Clear search and add input when dropdown closes
    searchQuery.value = '';
    addInputValue.value = '';
  }
});

// Get option class
function getOptionClass(option: SelectOption, index: number): string {
  const classes = ['select-option'];
  const isSelected = option.value === props.modelValue;
  if (isSelected) {
    classes.push('selected');
  }
  // Don't add focused class to selected items
  if (index === selectedIndex.value && isOpen.value && !isSelected) {
    classes.push('focused');
  }
  if (option.disabled) {
    classes.push('disabled');
  }
  return classes.join(' ');
}

// Get trigger class
function getTriggerClass(): string {
  const classes = ['select-trigger'];
  if (props.disabled) {
    classes.push('disabled');
  }
  if (props.bgMode === 'secondary') {
    classes.push('select-mode-secondary');
  }
  if (!selectedOption.value && props.placeholder) {
    classes.push('select-placeholder');
  }
  return classes.join(' ');
}

// Clean up scroll listener on unmount
onUnmounted(() => {
  cleanupScrollListener();
});
</script>

<template>
  <div :class="['select-container', 'relative', widthClass]" :style="maxWidthStyle">
    <!-- Trigger button -->
    <button
      ref="triggerRef"
      type="button"
      :class="getTriggerClass()"
      :disabled="disabled"
      :tabindex="disabled ? -1 : 0"
      @click="handleTriggerClick"
    >
      <span class="select-text truncate flex-1 text-left">
        {{ displayText }}
      </span>
      <div class="flex items-center gap-1 flex-shrink-0">
        <PhX
          v-if="clearable && selectedOption && !disabled"
          :size="14"
          class="select-clear"
          @click.stop="clearSelection"
        />
        <PhCaretDown :size="14" :class="['select-arrow', { open: isOpen }]" />
      </div>
    </button>

    <!-- Dropdown menu -->
    <Teleport to="body" :disabled="!shouldTeleport">
      <div
        v-if="isOpen"
        ref="dropdownRef"
        class="select-dropdown"
        :style="dropdownStyle"
        @mouseleave="handleMouseLeave"
      >
        <!-- Search input -->
        <div v-if="searchable" class="select-search">
          <input
            ref="searchInputRef"
            v-model="searchQuery"
            type="text"
            :placeholder="t('common.select.searchPlaceholder')"
          />
        </div>

        <!-- Render grouped options -->
        <template v-for="(item, groupIndex) in filteredOptions" :key="groupIndex">
          <!-- Option group -->
          <template v-if="isOptionGroup(item)">
            <div class="select-group-label">
              {{ item.label }}
            </div>
            <div
              v-for="option in item.options"
              :key="option.value"
              :class="getOptionClass(option, filteredFlatOptions.indexOf(option))"
              :style="option.style"
              @click.stop="selectOption(option)"
              @mouseenter="selectedIndex = filteredFlatOptions.indexOf(option)"
            >
              <slot name="option" :option="option">
                {{ option.label }}
              </slot>
            </div>
          </template>

          <!-- Single option -->
          <div
            v-else
            :class="getOptionClass(item, filteredFlatOptions.indexOf(item))"
            :style="item.style"
            @click.stop="selectOption(item)"
            @mouseenter="selectedIndex = filteredFlatOptions.indexOf(item)"
          >
            <slot name="option" :option="item">
              {{ item.label }}
            </slot>
          </div>
        </template>

        <!-- Custom input section -->
        <div v-if="allowCustomInput" class="select-custom-input">
          <input
            ref="customInputRef"
            v-model="customInputValue"
            type="text"
            :placeholder="customInputPlaceholder || t('common.input.customValue')"
            class="input-field w-full text-xs sm:text-sm"
            @keydown.enter="handleCustomInput"
          />
        </div>

        <!-- Add new option section -->
        <div v-if="allowAdd" class="select-add-section">
          <div class="select-add-input-wrapper">
            <input
              ref="addInputRef"
              v-model="addInputValue"
              type="text"
              :placeholder="addPlaceholder || t('common.select.addNewPlaceholder')"
              class="input-field w-full text-xs sm:text-sm"
              @keydown.enter="handleAddOption"
            />
            <button
              type="button"
              class="select-add-button"
              :disabled="!addInputValue.trim()"
              :title="t('common.select.addNew')"
              @click.stop="handleAddOption"
            >
              <PhPlus :size="16" />
            </button>
          </div>
        </div>

        <!-- Empty state -->
        <div v-if="filteredFlatOptions.length === 0 && !allowCustomInput" class="select-empty">
          {{
            searchable && searchQuery ? t('common.select.noResults') : t('common.select.noOptions')
          }}
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
@reference "../../style.css";
.select-container {
  position: relative;
}

.select-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.input-field {
  @apply p-1.5 sm:p-2 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors text-xs sm:text-sm;
}
</style>
