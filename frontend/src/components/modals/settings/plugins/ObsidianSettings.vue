<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhArchive, PhFolders } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';
import { NestedSettingsContainer, SubSettingItem, InputControl } from '@/components/settings';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

function updateSetting(key: keyof SettingsData, value: any) {
  emit('update:settings', {
    ...props.settings,
    [key]: value,
  });
}
</script>

<template>
  <!-- Enable Obsidian Integration -->
  <div class="setting-item">
    <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
      <img
        src="/assets/plugin_icons/obsidian.svg"
        alt="Obsidian"
        class="w-5 h-5 sm:w-6 sm:h-6 mt-0.5 shrink-0"
      />
      <div class="flex-1 min-w-0">
        <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
          {{ t('setting.plugins.obsidian.integration') }}
        </div>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.plugins.obsidian.integrationDescription') }}
        </div>
      </div>
    </div>
    <input
      type="checkbox"
      :checked="props.settings.obsidian_enabled"
      class="toggle"
      @change="updateSetting('obsidian_enabled', ($event.target as HTMLInputElement).checked)"
    />
  </div>

  <NestedSettingsContainer v-if="props.settings.obsidian_enabled">
    <!-- Vault Name -->
    <SubSettingItem
      :icon="PhArchive"
      :title="t('setting.plugins.obsidian.vaultName')"
      :description="t('setting.plugins.obsidian.vaultNameDesc')"
    >
      <InputControl
        :model-value="props.settings.obsidian_vault"
        :placeholder="t('setting.plugins.obsidian.vaultNamePlaceholder')"
        width="md"
        @update:model-value="updateSetting('obsidian_vault', $event)"
      />
    </SubSettingItem>

    <!-- Vault Path -->
    <SubSettingItem
      :icon="PhFolders"
      :title="t('setting.plugins.obsidian.vaultPath')"
      :description="t('setting.plugins.obsidian.vaultPathDesc')"
      required
    >
      <InputControl
        :model-value="props.settings.obsidian_vault_path"
        placeholder="C:\Users\username\Documents\Obsidian Vault"
        width="lg"
        @update:model-value="updateSetting('obsidian_vault_path', $event)"
      />
    </SubSettingItem>
  </NestedSettingsContainer>
</template>

<style scoped>
@reference "../../../../style.css";
.toggle {
  @apply w-10 h-5 appearance-none bg-bg-tertiary rounded-full relative cursor-pointer border border-border transition-colors checked:bg-accent checked:border-accent shrink-0;
}
.toggle::after {
  content: '';
  @apply absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow-sm transition-transform;
}
.toggle:checked::after {
  transform: translateX(20px);
}

.setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border;
}
</style>
