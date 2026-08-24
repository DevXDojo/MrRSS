<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhHardDrives, PhUpload, PhDownload, PhBroom } from '@phosphor-icons/vue';
import { ButtonControl } from '@/components/settings';
import { SettingGroup } from '@/components/settings';

const { t } = useI18n();

const emit = defineEmits<{
  'import-opml': [];
  'export-opml': [];
  'cleanup-database': [];
}>();

function handleImportOPML() {
  emit('import-opml');
}

function handleExportOPML() {
  emit('export-opml');
}

function handleCleanupDatabase() {
  emit('cleanup-database');
}
</script>

<template>
  <SettingGroup :icon="PhHardDrives" :title="t('setting.database.dataManagement')">
    <div class="flex flex-col sm:flex-row gap-2 sm:gap-3">
      <ButtonControl
        :label="t('modal.opml.import')"
        :icon="PhDownload"
        type="secondary"
        class="flex-1 justify-center text-sm sm:text-base"
        @click="handleImportOPML"
      />
      <ButtonControl
        :label="t('modal.opml.export')"
        :icon="PhUpload"
        type="secondary"
        class="flex-1 justify-center text-sm sm:text-base"
        @click="handleExportOPML"
      />
    </div>
    <ButtonControl
      :label="t('setting.database.cleanDatabase')"
      :icon="PhBroom"
      type="danger"
      class="w-full justify-center text-sm sm:text-base"
      @click="handleCleanupDatabase"
    />
  </SettingGroup>
</template>

<style scoped></style>
