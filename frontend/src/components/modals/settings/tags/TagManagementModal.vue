<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAppStore } from '@/stores/app';
import type { Tag } from '@/types/models';
import { PhPlus, PhPencil, PhTrash } from '@phosphor-icons/vue';
import BaseModal from '@/components/common/BaseModal.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';
import TagFormModal from './TagFormModal.vue';

const { t } = useI18n();
const store = useAppStore();

const emit = defineEmits<{
  close: [];
}>();

// State
const tags = ref<Tag[]>([]);
const editingTag = ref<Tag | null>(null);
const showAddForm = ref(false);

// Fetch tags on mount
onMounted(async () => {
  await fetchTags();
});

async function fetchTags() {
  try {
    const res = await fetch('/api/tags');
    tags.value = await res.json();
  } catch (e) {
    console.error('Failed to fetch tags:', e);
  }
}

function openAddForm() {
  editingTag.value = null;
  showAddForm.value = true;
}

function openEditForm(tag: Tag) {
  editingTag.value = tag;
  showAddForm.value = true;
}

async function handleSaveTag(name: string, color: string) {
  try {
    if (editingTag.value) {
      // Update existing tag
      await fetch('/api/tags/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: editingTag.value.id,
          name,
          color,
          position: editingTag.value.position,
        }),
      });
    } else {
      // Create new tag
      await fetch('/api/tags', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, color }),
      });
    }

    await fetchTags();
    showAddForm.value = false;
    editingTag.value = null;

    // Refresh store tags
    await store.fetchTags();
  } catch (e) {
    console.error('Failed to save tag:', e);
  }
}

async function deleteTag(tag: Tag) {
  if (!window.confirm(t('modal.tag.confirmDelete'))) {
    return;
  }

  try {
    await fetch('/api/tags/delete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: tag.id }),
    });

    await fetchTags();

    // Refresh store tags
    await store.fetchTags();
  } catch (e) {
    console.error('Failed to delete tag:', e);
  }
}

function getFeedCount(tag: Tag): number {
  return store.feeds.filter((f) => f.tags?.some((t) => t.id === tag.id)).length;
}

function handleEditTag(tag: Tag) {
  openEditForm(tag);
}

function handleDeleteTag(tag: Tag) {
  deleteTag(tag);
}

function closeForm() {
  showAddForm.value = false;
  editingTag.value = null;
}

function close() {
  emit('close');
}
</script>

<template>
  <!-- Tag Management Modal -->
  <BaseModal :title="t('modal.tag.manageTags')" size="2xl" :z-index="100" @close="close">
    <!-- Content -->
    <div class="p-4 sm:p-6">
      <!-- Tag List -->
      <div v-if="tags.length > 0">
        <!-- Tags displayed in a flex wrap layout -->
        <div class="flex flex-wrap gap-2">
          <div
            v-for="tag in tags"
            :key="tag.id"
            class="inline-flex items-center rounded-l rounded-r border border-border hover:border-accent/50 transition-all overflow-hidden"
          >
            <!-- Left part: Tag name and count with colored background -->
            <div
              class="flex items-center gap-1.5 px-2.5 py-1.5"
              :style="{ backgroundColor: tag.color }"
            >
              <!-- Tag name -->
              <span class="text-sm font-medium text-white">{{ tag.name }}</span>

              <!-- Feed count badge -->
              <span class="feed-count-badge">{{ getFeedCount(tag) }}</span>
            </div>

            <!-- Right part: Action buttons -->
            <div
              class="flex items-center justify-center px-2 h-full bg-bg-secondary gap-2 border-l border-border self-stretch"
            >
              <!-- Edit button -->
              <button
                class="text-text-secondary hover:text-accent transition-colors"
                @click="handleEditTag(tag)"
              >
                <PhPencil :size="18" />
              </button>

              <!-- Delete button -->
              <button
                class="text-text-secondary hover:text-red-500 transition-colors"
                @click="handleDeleteTag(tag)"
              >
                <PhTrash :size="18" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty state -->
      <div v-else class="text-center py-12 text-text-secondary">
        <div class="text-4xl mb-4">🏷️</div>
        <p>{{ t('modal.tag.noTags') }}</p>
        <button
          class="mt-4 px-4 py-2 text-sm font-medium text-white bg-accent rounded-md hover:bg-accent/90 transition-colors"
          @click="openAddForm"
        >
          <PhPlus :size="20" class="inline mr-1" />
          {{ t('modal.tag.createTag') }}
        </button>
      </div>
    </div>

    <!-- Footer -->
    <template #footer>
      <ModalFooter
        align="right"
        :primary-button="{
          label: t('modal.tag.addNew'),
          onClick: openAddForm,
        }"
      />
    </template>

    <!-- Tag Form Modal (Nested) -->
    <TagFormModal
      v-if="showAddForm"
      :editing-tag="editingTag"
      @close="closeForm"
      @save="handleSaveTag"
    />
  </BaseModal>
</template>

<style scoped>
@reference "../../../../style.css";

.btn-primary {
  @apply bg-accent text-white border-none px-4 sm:px-5 py-2 sm:py-2.5 rounded-lg cursor-pointer font-semibold hover:bg-accent-hover transition-colors disabled:opacity-70;
}

.feed-count-badge {
  @apply text-[10px] font-medium rounded-full min-w-[16px] h-[16px] px-1 flex items-center justify-center;
  background-color: rgba(255, 255, 255, 0.25);
  color: #ffffff;
}

.dark-mode .feed-count-badge {
  background-color: rgba(0, 0, 0, 0.25) !important;
  color: #ffffff !important;
}
</style>
