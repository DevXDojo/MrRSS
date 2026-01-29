<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAppStore } from '@/stores/app';
import { PhX, PhPlus } from '@phosphor-icons/vue';
import BaseModal from '@/components/common/BaseModal.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';
import TagFormModal from '../tags/TagFormModal.vue';

interface Props {
  show: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  confirm: [tagIds: number[]];
}>();

const { t } = useI18n();
const store = useAppStore();

const selectedTags = ref<number[]>([]);

const availableTags = computed(() => store.tags || []);

// New tag creation state
const showNewTagModal = ref(false);

// Reset selection when modal opens/closes
watch(
  () => props.show,
  (newValue) => {
    if (!newValue) {
      selectedTags.value = [];
    }
  }
);

function toggleTag(tagId: number) {
  if (selectedTags.value.includes(tagId)) {
    selectedTags.value = selectedTags.value.filter((id) => id !== tagId);
  } else {
    selectedTags.value = [...selectedTags.value, tagId];
  }
}

function removeTag(tagId: number) {
  selectedTags.value = selectedTags.value.filter((id) => id !== tagId);
}

function openNewTagModal() {
  showNewTagModal.value = true;
}

function closeNewTagModal() {
  showNewTagModal.value = false;
}

async function handleSaveTag(name: string, color: string) {
  try {
    const res = await fetch('/api/tags', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, color }),
    });

    if (res.ok) {
      const newTag = await res.json();
      // Refresh tags in store
      await store.fetchTags();
      // Auto-select the newly created tag
      selectedTags.value = [...selectedTags.value, newTag.id];
      closeNewTagModal();
      window.showToast(t('modal.tag.tagCreated'), 'success');
    } else {
      window.showToast(t('common.errors.createFailed'), 'error');
    }
  } catch (e) {
    console.error('Failed to create tag:', e);
    window.showToast(t('common.errors.createFailed'), 'error');
  }
}

function handleConfirm() {
  emit('confirm', selectedTags.value);
  emit('close');
}

function handleCancel() {
  emit('close');
}

function handleClose() {
  emit('close');
}
</script>

<template>
  <BaseModal
    v-if="show"
    :title="t('common.action.addTags')"
    size="md"
    :closable="false"
    @close="handleClose"
  >
    <!-- Body -->
    <div class="p-3 sm:p-5">
      <p class="m-0 mb-3 text-text-primary text-sm sm:text-base">
        {{ t('modal.feed.selectTagsToAdd') }}
      </p>

      <!-- Selected tags as chips -->
      <div v-if="selectedTags.length > 0" class="flex flex-wrap gap-2 mb-3">
        <span
          v-for="tagId in selectedTags"
          :key="tagId"
          class="inline-flex items-center gap-1 px-2 py-1 text-xs rounded text-white"
          :style="{ backgroundColor: store.tagMap.get(tagId)?.color || '#3B82F6' }"
        >
          {{ store.tagMap.get(tagId)?.name }}
          <button class="hover:text-gray-200" @click="removeTag(tagId)">
            <PhX :size="14" />
          </button>
        </span>
      </div>

      <!-- Available tags grid -->
      <div class="border border-border rounded-lg bg-bg-tertiary p-3 mb-3 max-h-64 overflow-y-auto">
        <div v-if="availableTags.length === 0" class="text-center text-text-secondary text-sm py-4">
          {{ t('modal.tag.noTags') }}
        </div>
        <div v-else class="grid grid-cols-2 sm:grid-cols-3 gap-2">
          <div
            v-for="tag in availableTags"
            :key="tag.id"
            :class="[
              'px-3 py-2 rounded-md cursor-pointer transition-all text-sm font-medium border',
              selectedTags.includes(tag.id)
                ? 'border-accent bg-accent/10 text-accent'
                : 'border-border bg-bg-secondary text-text-primary hover:border-accent/50',
            ]"
            @click="toggleTag(tag.id)"
          >
            <div class="flex items-center gap-2">
              <span
                class="w-3 h-3 rounded-full shrink-0 border border-white/20"
                :style="{ backgroundColor: tag.color }"
              />
              <span class="truncate">{{ tag.name }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Selection count -->
      <div v-if="selectedTags.length > 0" class="text-xs text-text-secondary mb-3">
        {{
          t('common.search.totalAndSelected', {
            total: availableTags.length,
            selected: selectedTags.length,
          })
        }}
      </div>

      <!-- Create new tag button -->
      <button
        type="button"
        class="text-xs text-accent hover:text-accent-hover transition-colors flex items-center gap-1"
        @click="openNewTagModal"
      >
        <PhPlus :size="14" />
        {{ t('modal.tag.createNew') }}
      </button>
    </div>

    <!-- Footer -->
    <template #footer>
      <ModalFooter
        :secondary-button="{
          label: t('common.cancel'),
          onClick: handleCancel,
        }"
        :primary-button="{
          label: t('common.action.add'),
          onClick: handleConfirm,
        }"
      />
    </template>
  </BaseModal>

  <!-- New tag modal -->
  <Teleport to="body">
    <TagFormModal
      v-if="showNewTagModal"
      :editing-tag="null"
      @close="closeNewTagModal"
      @save="handleSaveTag"
    />
  </Teleport>
</template>

<style scoped>
@reference "../../../style.css";
</style>
