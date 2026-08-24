<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted, type PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhCaretDown, PhCheck, PhChecks, PhPlus } from '@phosphor-icons/vue';
import { type SelectOption, flattenOptions } from '@/types/select';
import { useSelect } from '@/composables/ui/useSelect';
import './select.css';

const { t } = useI18n();

// Generate unique ID for this dropdown instance
const dropdownId = `multiselect-${Math.random().toString(36).substring(2, 9)}`;

const props = defineProps({
  modelValue: {
    type: Array as PropType<(string | number)[]>,
    default: () => [],
  },
  options: {
    type: Array as PropType<SelectOption[]>,
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
  position: {
    type: String as PropType<'bottom' | 'top' | 'auto'>,
    default: 'bottom',
  },
  displayMode: {
    type: String as PropType<'chips' | 'counter'>,
    default: 'counter',
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
  'update:modelValue': [value: (string | number)[]];
  add: [value: string];
}>();

const isOpen = ref(false);
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
  return props.options.filter((option) => option.label.toLowerCase().includes(query));
});

// Flatten options for navigation
const flatOptions = computed(() => flattenOptions(filteredOptions.value));

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

// Check if an option is selected
function isSelected(value: string | number): boolean {
  return props.modelValue.includes(value);
}

// Toggle an option
function toggleOption(option: SelectOption) {
  if (option.disabled) return;

  const newValues = isSelected(option.value)
    ? props.modelValue.filter((v) => v !== option.value)
    : [...props.modelValue, option.value];

  emit('update:modelValue', newValues);
}

// Display text for trigger button
const displayText = computed(() => {
  const count = props.modelValue.length;

  if (count === 0) {
    return props.placeholder || t('common.search.selectItems');
  }

  if (count === 1) {
    const option = flatOptions.value.find((opt) => opt.value === props.modelValue[0]);
    return option?.label || `${count} ${t('common.search.item')}`;
  }

  return t('common.search.itemsSelected', { count });
});

// Toggle dropdown
function toggleDropdown() {
  if (props.disabled) return;
  isOpen.value = !isOpen.value;
  if (isOpen.value) {
    registerAsOpen();
    resetIndex();
    setupScrollListener();
    nextTick(updateDropdownPosition);
    if (props.allowAdd) {
      nextTick(() => {
        addInputRef.value?.focus();
      });
    } else if (props.searchable) {
      nextTick(() => {
        searchInputRef.value?.focus();
      });
    }
  } else {
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

// Watch for dropdown state changes to clear search
watch(isOpen, (open) => {
  if (!open) {
    searchQuery.value = '';
    addInputValue.value = '';
  }
});

// Handle click on trigger
function handleTriggerClick(event: MouseEvent) {
  event.stopPropagation();
  toggleDropdown();
}

// Get option class
function getOptionClass(option: SelectOption, index: number): string {
  const classes = ['select-option'];
  const isOptionSelected = isSelected(option.value);
  if (isOptionSelected) {
    classes.push('selected');
  }
  // Don't add focused class to selected items
  if (index === selectedIndex.value && isOpen.value && !isOptionSelected) {
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
  return classes.join(' ');
}

// Render selected chips
const selectedOptions = computed(() => {
  return props.modelValue
    .map((value) => flatOptions.value.find((opt) => opt.value === value))
    .filter(Boolean) as SelectOption[];
});

// Clean up scroll listener on unmount
onUnmounted(() => {
  cleanupScrollListener();
});
</script>

<template>
  <div :class="['select-container', 'relative', widthClass]" :style="maxWidthStyle">
    <!-- Display mode: chips -->
    <div v-if="displayMode === 'chips' && selectedOptions.length > 0" class="select-chips">
      <span
        v-for="option in selectedOptions"
        :key="option.value"
        class="select-chip"
        :style="{
          backgroundColor: option.color || 'var(--accent-color)',
          color: 'white',
        }"
      >
        {{ option.label }}
        <button type="button" class="hover:text-gray-200 ml-1" @click.stop="toggleOption(option)">
          ×
        </button>
      </span>
    </div>

    <!-- Trigger button -->
    <button
      ref="triggerRef"
      type="button"
      :class="getTriggerClass()"
      :disabled="disabled"
      :tabindex="disabled ? -1 : 0"
      @click="handleTriggerClick"
    >
      <PhChecks :size="16" class="text-accent flex-shrink-0 mr-1.5" />
      <span class="select-text truncate flex-1 text-left">
        {{ displayText }}
      </span>
      <PhCaretDown :size="14" :class="['select-arrow', { open: isOpen }]" />
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

        <!-- Options -->
        <div
          v-for="(option, index) in filteredOptions"
          :key="option.value"
          :class="getOptionClass(option, index)"
          @click.stop="toggleOption(option)"
          @mouseenter="selectedIndex = index"
        >
          <span
            v-if="option.color"
            class="w-3 h-3 rounded-full flex-shrink-0"
            :style="{ backgroundColor: option.color }"
          />
          <span class="truncate flex-1">{{ option.label }}</span>
          <PhCheck
            v-if="isSelected(option.value)"
            :size="16"
            :class="['flex-shrink-0', isSelected(option.value) ? 'text-white' : 'text-accent']"
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
        <div v-if="flatOptions.length === 0" class="select-empty">
          {{
            searchable && searchQuery ? t('common.select.noResults') : t('common.select.noOptions')
          }}
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.select-container {
  position: relative;
}

.select-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
