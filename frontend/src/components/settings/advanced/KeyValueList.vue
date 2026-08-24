<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { PhPlus, PhTrash } from '@phosphor-icons/vue';
import KeyValueInput from './KeyValueInput.vue';

export interface KeyValuePair {
  id: string;
  key: string;
  value: string;
}

interface Props {
  modelValue: string; // JSON string of key-value pairs
  keyPlaceholder?: string;
  valuePlaceholder?: string;
  addButtonText?: string;
  removeButtonTitle?: string;
  emptyKeyWarning?: string;
  emptyValueWarning?: string;
  debounceMs?: number;
  /** Only allow ASCII characters (for HTTP headers) */
  asciiOnly?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  keyPlaceholder: 'Key',
  valuePlaceholder: 'Value',
  addButtonText: 'Add',
  removeButtonTitle: 'Remove',
  emptyKeyWarning: '',
  emptyValueWarning: '',
  debounceMs: 500,
  asciiOnly: false,
});

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

// Parse JSON string to array
function parseJsonString(jsonString: string): KeyValuePair[] {
  if (!jsonString || jsonString.trim() === '') return [];
  try {
    const parsed = JSON.parse(jsonString) as Record<string, string>;
    return Object.entries(parsed).map(([key, value], index) => ({
      id: `${Date.now()}-${index}`,
      key,
      value,
    }));
  } catch {
    return [];
  }
}

// Convert array to JSON string
function jsonStringify(pairs: KeyValuePair[]): string {
  const validPairs = pairs.filter((p) => p.key.trim() !== '');
  if (validPairs.length === 0) return '';
  const obj: Record<string, string> = {};
  validPairs.forEach((p) => {
    obj[p.key] = p.value;
  });
  return JSON.stringify(obj);
}

const pairs = ref<KeyValuePair[]>([]);
let saveTimeout: ReturnType<typeof setTimeout> | null = null;

function loadPairs() {
  pairs.value = parseJsonString(props.modelValue);
}

function updateKey(pairId: string, value: string) {
  const pair = pairs.value.find((p) => p.id === pairId);
  if (pair) {
    pair.key = value;
    debouncedSave();
  }
}

function updateValue(pairId: string, value: string) {
  const pair = pairs.value.find((p) => p.id === pairId);
  if (pair) {
    pair.value = value;
    debouncedSave();
  }
}

function savePairs() {
  const jsonString = jsonStringify(pairs.value);
  emit('update:modelValue', jsonString);
}

function addPair() {
  pairs.value.push({
    id: `${Date.now()}-${Math.random()}`,
    key: '',
    value: '',
  });
  debouncedSave();
}

function removePair(id: string) {
  const index = pairs.value.findIndex((p) => p.id === id);
  if (index !== -1) {
    pairs.value.splice(index, 1);
    debouncedSave();
  }
}

function debouncedSave() {
  if (saveTimeout) clearTimeout(saveTimeout);
  saveTimeout = setTimeout(() => {
    savePairs();
    saveTimeout = null;
  }, props.debounceMs);
}

// Watch for external changes
watch(
  () => props.modelValue,
  (newValue, oldValue) => {
    if (newValue !== oldValue) {
      const parsed = parseJsonString(newValue || '');
      const currentIds = new Set(pairs.value.map((p) => p.id));
      const hasNewEntries = parsed.some((p) => !currentIds.has(p.id));
      if (hasNewEntries || parsed.length !== pairs.value.length) {
        pairs.value = parsed;
      }
    }
  }
);

onMounted(() => {
  loadPairs();
});

// Export for component usage
defineOptions({
  name: 'KeyValueList',
});
</script>

<template>
  <div class="key-value-list w-full">
    <!-- Pairs List -->
    <div class="mt-2 sm:mt-3 space-y-1.5 sm:space-y-2">
      <div v-for="pair in pairs" :key="pair.id" class="flex items-center gap-1.5 sm:gap-2">
        <KeyValueInput
          :model-value="pair.key"
          :placeholder="keyPlaceholder"
          :ascii-only="asciiOnly"
          @update:model-value="updateKey(pair.id, $event)"
        />
        <KeyValueInput
          :model-value="pair.value"
          :placeholder="valuePlaceholder"
          :ascii-only="asciiOnly"
          @update:model-value="updateValue(pair.id, $event)"
        />
        <button
          type="button"
          class="p-1.5 sm:p-2 rounded hover:bg-red-50 dark:hover:bg-red-900/20 text-text-secondary hover:text-red-500 transition-all shrink-0"
          :title="removeButtonTitle"
          @click="removePair(pair.id)"
        >
          <PhTrash :size="14" class="sm:w-4 sm:h-4" />
        </button>
      </div>

      <!-- Add Pair Button -->
      <button
        type="button"
        class="!w-full p-1.5 sm:p-2 rounded border border-dashed border-border text-text-secondary hover:border-accent hover:text-accent hover:bg-accent/5 transition-all text-xs font-medium flex items-center justify-center gap-1.5 sm:gap-2"
        @click="addPair"
      >
        <PhPlus :size="14" class="sm:w-4 sm:h-4" />
        <span>{{ addButtonText }}</span>
      </button>
    </div>
  </div>
</template>
