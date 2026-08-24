<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhArrowClockwise, PhArrowsClockwise, PhBell, PhClock } from '@phosphor-icons/vue';
import {
  SettingGroup,
  SettingWithToggle,
  SettingWithSelect,
  SubSettingItem,
  NumberControl,
  NestedSettingsContainer,
} from '@/components/settings';
import '@/components/settings/styles.css';
import type { SettingsData } from '@/types/settings';
import { formatDate } from '@/utils/date';

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
  <SettingGroup :icon="PhArrowClockwise" :title="t('setting.update.updates')">
    <SettingWithToggle
      :icon="PhBell"
      :title="t('setting.update.updateCheckEnabled')"
      :description="t('setting.update.updateCheckEnabledDesc')"
      :model-value="settings.update_check_enabled"
      @update:model-value="updateSetting('update_check_enabled', $event)"
    />

    <!-- Refresh Mode -->
    <SettingWithSelect
      :icon="PhArrowsClockwise"
      :title="t('setting.feed.refreshMode')"
      :description="t('setting.feed.refreshModeDesc')"
      :model-value="settings.refresh_mode"
      :options="[
        { value: 'fixed', label: t('setting.feed.fixedInterval') },
        { value: 'intelligent', label: t('setting.feed.intelligentInterval') },
        { value: 'never', label: t('setting.feed.neverRefresh') },
      ]"
      width="md"
      @update:model-value="updateSetting('refresh_mode', $event)"
    />

    <!-- Auto Update Interval (shown when fixed mode is selected) -->
    <NestedSettingsContainer v-if="settings.refresh_mode === 'fixed'">
      <SubSettingItem
        :icon="PhClock"
        :title="t('setting.update.autoUpdateInterval')"
        :description="t('setting.update.autoUpdateIntervalDesc')"
      >
        <template #extraInfo>
          <div class="text-xs text-text-secondary mt-1">
            {{ t('sidebar.activity.lastGlobalRefresh') }}:
            <span class="theme-number">{{
              formatDate(props.settings.last_global_refresh, props.settings.language, t)
            }}</span>
          </div>
        </template>
        <NumberControl
          :model-value="settings.update_interval"
          :min="1"
          :suffix="t('common.time.minutes')"
          @update:model-value="updateSetting('update_interval', $event)"
        />
      </SubSettingItem>
    </NestedSettingsContainer>
  </SettingGroup>
</template>

<style scoped></style>
