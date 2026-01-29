<script setup lang="ts">
import { computed, ref } from 'vue';
import DataManagementSettings from './DataManagementSettings.vue';
import FeedManagementSettings from './FeedManagementSettings.vue';
import DiscoverySettings from './DiscoverySettings.vue';
import TagManagementModal from '../tags/TagManagementModal.vue';
import type { Feed } from '@/types/models';
import type { SettingsData } from '@/types/settings';
import { useSettingsAutoSave } from '@/composables/core/useSettingsAutoSave';

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'import-opml': [];
  'export-opml': [];
  'cleanup-database': [];
  'add-feed': [];
  'edit-feed': [feed: Feed];
  'delete-feed': [id: number];
  'batch-delete': [ids: number[]];
  'batch-move': [ids: number[]];
  'batch-add-tags': [ids: number[]];
  'batch-set-image-mode': [ids: number[]];
  'batch-unset-image-mode': [ids: number[]];
  'discover-all': [];
  'update:settings': [settings: SettingsData];
  'select-feed': [feedId: number];
}>();

// Create a computed ref that returns the settings object
// This ensures reactivity while allowing modifications
const settingsRef = computed(() => props.settings);

// Use composable for auto-save with reactivity
useSettingsAutoSave(settingsRef);

// Tag management modal state
const showTagManagement = ref(false);

// Event handlers that pass through to parent
function handleImportOPML() {
  emit('import-opml');
}

function handleExportOPML() {
  emit('export-opml');
}

function handleCleanupDatabase() {
  emit('cleanup-database');
}

function handleDiscoverAll() {
  emit('discover-all');
}

function handleAddFeed() {
  emit('add-feed');
}

function handleEditFeed(feed: Feed) {
  emit('edit-feed', feed);
}

function handleDeleteFeed(id: number) {
  emit('delete-feed', id);
}

function handleBatchDelete(ids: number[]) {
  emit('batch-delete', ids);
}

function handleBatchMove(ids: number[]) {
  emit('batch-move', ids);
}

function handleBatchAddTags(ids: number[]) {
  emit('batch-add-tags', ids);
}

function handleBatchSetImageMode(ids: number[]) {
  emit('batch-set-image-mode', ids);
}

function handleBatchUnsetImageMode(ids: number[]) {
  emit('batch-unset-image-mode', ids);
}

function handleSelectFeed(feedId: number) {
  emit('select-feed', feedId);
}

function handleManageTags() {
  showTagManagement.value = true;
}
</script>

<template>
  <div class="space-y-4 sm:space-y-6">
    <DataManagementSettings
      @import-opml="handleImportOPML"
      @export-opml="handleExportOPML"
      @cleanup-database="handleCleanupDatabase"
    />

    <FeedManagementSettings
      @add-feed="handleAddFeed"
      @edit-feed="handleEditFeed"
      @delete-feed="handleDeleteFeed"
      @batch-delete="handleBatchDelete"
      @batch-move="handleBatchMove"
      @batch-add-tags="handleBatchAddTags"
      @batch-set-image-mode="handleBatchSetImageMode"
      @batch-unset-image-mode="handleBatchUnsetImageMode"
      @select-feed="handleSelectFeed"
      @manage-tags="handleManageTags"
    />

    <DiscoverySettings @discover-all="handleDiscoverAll" />
  </div>

  <!-- Tag Management Modal (Teleported to body) -->
  <Teleport to="body">
    <TagManagementModal v-if="showTagManagement" @close="showTagManagement = false" />
  </Teleport>
</template>
