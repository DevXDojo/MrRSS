<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhFolderOpen } from '@phosphor-icons/vue';
import { SettingItem } from '@/components/settings';
const { t } = useI18n();
interface StorageStatus {
  data_directory: string;
  pending_directory: string;
  last_error: string;
  managed: boolean;
}
const status = ref<StorageStatus | null>(null);
const path = ref('');
const busy = ref(false);
async function request(url: string, options?: RequestInit) {
  const response = await fetch(url, options);
  if (!response.ok) {
    let message = t('setting.database.directoryFailed');
    try {
      const data = await response.json();
      message = data.error?.message || message;
    } catch {
      /* no JSON */
    }
    throw new Error(message);
  }
  return response.json();
}
onMounted(async () => {
  try {
    const response = await fetch('/api/settings/data-directory');
    if (response.status === 501) return;
    if (!response.ok) throw new Error(t('setting.database.directoryFailed'));
    status.value = await response.json();
  } catch {
    window.showToast(t('setting.database.directoryFailed'), 'error');
  }
});

async function browse() {
  busy.value = true;
  try {
    const data = await request('/api/settings/data-directory/select', { method: 'POST' });
    if (data.path) path.value = data.path;
  } catch (error) {
    window.showToast(
      error instanceof Error ? error.message : t('setting.database.directoryFailed'),
      'error'
    );
  } finally {
    busy.value = false;
  }
}
async function save() {
  const confirmed = await window.showConfirm({
    title: t('setting.database.dataDirectory'),
    message: t('setting.database.directoryConfirm', { path: path.value }),
  });
  if (!confirmed) return;
  busy.value = true;
  try {
    status.value = await request('/api/settings/data-directory', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path: path.value }),
    });
    path.value = '';
    window.showToast(t('setting.database.directoryScheduled'), 'success');
  } catch (error) {
    window.showToast(
      error instanceof Error ? error.message : t('setting.database.directoryFailed'),
      'error'
    );
  } finally {
    busy.value = false;
  }
}
async function cancel() {
  busy.value = true;
  try {
    status.value = await request('/api/settings/data-directory', { method: 'DELETE' });
  } catch (error) {
    window.showToast(
      error instanceof Error ? error.message : t('setting.database.directoryFailed'),
      'error'
    );
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <SettingItem
    v-if="status"
    :icon="PhFolderOpen"
    :title="t('setting.database.dataDirectory')"
    :description="t('setting.database.directoryDescription')"
    layout="column"
  >
    <div class="space-y-3">
      <p class="text-xs break-all">
        <span class="text-text-secondary">{{ t('setting.database.directoryCurrent') }} </span
        >{{ status.data_directory }}
      </p>
      <p v-if="!status.managed" class="text-xs text-text-secondary">
        {{ t('setting.database.directoryOverridden') }}
      </p>
      <template v-else>
        <p v-if="status.last_error" role="alert" class="text-xs text-red-500">
          {{ t('setting.database.directoryMigrationFailed') }} {{ status.last_error }}
        </p>
        <div v-if="status.pending_directory" class="space-y-2">
          <p class="text-xs break-all">
            {{ t('setting.database.directoryPending') }} {{ status.pending_directory }}
          </p>
          <p class="text-xs text-text-secondary">
            {{ t('setting.database.directoryScheduled') }}
          </p>
          <button type="button" class="btn-secondary" :disabled="busy" @click="cancel">
            {{ t('setting.database.directoryCancel') }}
          </button>
        </div>
        <div v-else class="flex flex-wrap gap-2">
          <input
            v-model="path"
            class="flex-1 min-w-0 rounded border border-border bg-bg-tertiary p-2 text-sm"
            :disabled="busy"
            :aria-label="t('setting.database.directoryDestination')"
            :placeholder="t('setting.database.directoryDestination')"
          />
          <button type="button" class="btn-secondary" :disabled="busy" @click="browse">
            {{ t('setting.database.directoryBrowse') }}
          </button>
          <button
            type="button"
            class="rounded bg-accent px-3 py-2 text-white disabled:opacity-40"
            :disabled="busy || !path.trim()"
            @click="save"
          >
            {{ t('setting.database.directoryChange') }}
          </button>
        </div>
      </template>
    </div>
  </SettingItem>
</template>
