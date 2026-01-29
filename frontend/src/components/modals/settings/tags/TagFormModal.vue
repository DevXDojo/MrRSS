<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAppStore } from '@/stores/app';
import type { Tag } from '@/types/models';
import { PhCheck } from '@phosphor-icons/vue';
import BaseModal from '@/components/common/BaseModal.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';

interface Props {
  editingTag: Tag | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  save: [name: string, color: string];
}>();

const { t } = useI18n();
const store = useAppStore();

const newTagName = ref('');
const newTagColor = ref('#3B82F6');

// Predefined colors - sorted by color wheel
const predefinedColors = [
  '#EF4444', // red
  '#F97316', // orange
  '#EAB308', // yellow
  '#84CC16', // lime
  '#10B981', // green
  '#06B6D4', // cyan
  '#3B82F6', // blue
  '#6366F1', // indigo
  '#8B5CF6', // violet
  '#EC4899', // pink
];

// Initialize form when editingTag changes
watch(
  () => props.editingTag,
  (tag) => {
    if (tag) {
      newTagName.value = tag.name;
      newTagColor.value = tag.color;
    } else {
      newTagName.value = '';
      newTagColor.value = '#3B82F6';
    }
  },
  { immediate: true }
);

// Computed feeds for the tag being edited
const feedsForTag = computed(() => {
  if (!props.editingTag) return [];
  return store.feeds.filter((f) => f.tags?.some((t) => t.id === props.editingTag!.id));
});

function getFavicon(url: string): string {
  try {
    return `https://www.google.com/s2/favicons?domain=${new URL(url).hostname}`;
  } catch {
    return '';
  }
}

function saveTag() {
  const name = newTagName.value.trim();
  if (!name) {
    return;
  }
  emit('save', name, newTagColor.value);
}

function closeForm() {
  emit('close');
}

// Computed title
const modalTitle = computed(() => {
  return props.editingTag ? t('modal.tag.editTag') : t('modal.tag.createTag');
});

// Computed button text
const saveButtonText = computed(() => {
  return props.editingTag ? t('common.action.save') : t('common.action.add');
});
</script>

<template>
  <!-- Tag Form Modal -->
  <BaseModal :title="modalTitle" size="md" :z-index="110" @close="closeForm">
    <!-- Form -->
    <div class="p-4 sm:p-6 space-y-4">
      <!-- Name -->
      <div>
        <label class="block mb-1.5 text-sm font-medium text-text-secondary">
          {{ t('modal.tag.name') }}
        </label>
        <input
          v-model="newTagName"
          type="text"
          class="input-field w-full"
          :placeholder="t('modal.tag.name')"
          @keyup.enter="saveTag"
        />
      </div>

      <!-- Color -->
      <div>
        <label class="block mb-1.5 text-sm font-medium text-text-secondary">
          {{ t('modal.tag.color') }}
        </label>
        <div class="flex gap-2 flex-wrap">
          <button
            v-for="color in predefinedColors"
            :key="color"
            class="w-8 h-8 rounded-md border-2 transition-all hover:scale-110 flex items-center justify-center"
            :class="newTagColor === color ? 'border-accent scale-110' : 'border-transparent'"
            :style="{ backgroundColor: color }"
            @click="newTagColor = color"
          >
            <PhCheck v-if="newTagColor === color" :size="20" weight="bold" class="text-white" />
          </button>
        </div>
      </div>

      <!-- Feeds using this tag (only for editing) -->
      <div v-if="editingTag && feedsForTag.length > 0" class="pt-4 border-t border-border">
        <label class="block mb-2 text-sm font-medium text-text-primary">
          {{ t('modal.tag.assignedFeeds', { count: feedsForTag.length }) }}
        </label>
        <div class="max-h-48 overflow-y-auto space-y-1 pr-1">
          <div
            v-for="feed in feedsForTag"
            :key="feed.id"
            class="flex items-center gap-2 px-2 py-1.5 bg-bg-tertiary rounded text-sm"
          >
            <!-- Favicon -->
            <img
              :src="getFavicon(feed.url)"
              class="w-4 h-4 flex-shrink-0 object-contain"
              @error="
                (e: Event) => {
                  (e.target as HTMLImageElement).style.display = 'none';
                }
              "
            />
            <!-- Feed title -->
            <span class="flex-1 min-w-0 truncate text-text-primary">{{ feed.title }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <template #footer>
      <ModalFooter
        align="right"
        :secondary-button="{
          label: t('common.action.cancel'),
          onClick: closeForm,
        }"
        :primary-button="{
          label: saveButtonText,
          onClick: saveTag,
        }"
      />
    </template>
  </BaseModal>
</template>

<style scoped>
@reference "../../../../style.css";

.input-field {
  @apply w-full p-2 sm:p-2.5 border border-border rounded-md bg-bg-tertiary text-text-primary text-xs sm:text-sm focus:border-accent focus:outline-none transition-colors;
  box-sizing: border-box;
}
</style>
