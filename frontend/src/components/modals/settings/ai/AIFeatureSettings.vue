<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhRobot,
  PhChatCircleText,
  PhTrash,
  PhBroom,
  PhMagnifyingGlass,
} from '@phosphor-icons/vue';
import {
  TipBox,
  SettingGroup,
  SettingWithToggle,
  NestedSettingsContainer,
  SubSettingItem,
} from '@/components/settings';
import AIProfileSelector from './AIProfileSelector.vue';
import '@/components/settings/styles.css';
import type { SettingsData } from '@/types/settings';

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

const isDeleting = ref(false);

async function clearAllChatSessions() {
  const confirmed = await window.showConfirm({
    title: t('setting.ai.clearAllChats'),
    message: t('setting.ai.clearAllChatsConfirm'),
    isDanger: true,
  });
  if (!confirmed) return;

  isDeleting.value = true;
  try {
    const response = await fetch('/api/ai/chat/sessions/delete-all', {
      method: 'DELETE',
    });

    if (response.ok) {
      const data = await response.json();
      window.showToast(t('setting.ai.clearAllChatsSuccess', { count: data.count || 0 }), 'success');
    } else {
      const errorText = await response.text();
      console.error('Server error:', response.status, errorText);
      window.showToast(t('setting.ai.clearAllChatsFailed'), 'error');
    }
  } catch (error) {
    console.error('Failed to clear chat sessions:', error);
    window.showToast(t('setting.ai.clearAllChatsFailed'), 'error');
  } finally {
    isDeleting.value = false;
  }
}
</script>

<template>
  <SettingGroup :icon="PhRobot" :title="t('setting.ai.aiFeatures')">
    <!-- AI Search -->
    <TipBox type="info" :title="t('setting.ai.isBeta')" />
    <SettingWithToggle
      :icon="PhMagnifyingGlass"
      :title="t('setting.ai.aiSearchEnabled')"
      :description="t('setting.ai.aiSearchEnabledDesc')"
      :model-value="props.settings.ai_search_enabled"
      @update:model-value="updateSetting('ai_search_enabled', $event)"
    />

    <NestedSettingsContainer v-if="props.settings.ai_search_enabled">
      <SubSettingItem
        :icon="PhRobot"
        :title="t('setting.ai.selectProfile')"
        :description="t('setting.ai.selectProfileForSearch')"
      >
        <AIProfileSelector
          :model-value="props.settings.ai_search_profile_id"
          @update:model-value="updateSetting('ai_search_profile_id', $event)"
        />
      </SubSettingItem>
    </NestedSettingsContainer>

    <!-- AI Chat -->
    <SettingWithToggle
      :icon="PhChatCircleText"
      :title="t('setting.ai.aiChatEnabled')"
      :description="t('setting.ai.aiChatEnabledDesc')"
      :model-value="props.settings.ai_chat_enabled"
      @update:model-value="updateSetting('ai_chat_enabled', $event)"
    />

    <NestedSettingsContainer v-if="props.settings.ai_chat_enabled">
      <SubSettingItem
        :icon="PhRobot"
        :title="t('setting.ai.selectProfile')"
        :description="t('setting.ai.selectProfileForChat')"
      >
        <AIProfileSelector
          :model-value="props.settings.ai_chat_profile_id"
          @update:model-value="updateSetting('ai_chat_profile_id', $event)"
        />
      </SubSettingItem>

      <SubSettingItem
        :icon="PhTrash"
        :title="t('setting.ai.clearAllChats')"
        :description="t('setting.ai.clearAllChatsDesc')"
      >
        <button
          type="button"
          :disabled="isDeleting"
          class="btn-secondary"
          @click="clearAllChatSessions"
        >
          <PhBroom :size="16" class="sm:w-5 sm:h-5" />
          {{ isDeleting ? t('setting.database.cleaning') : t('setting.ai.clearAllChatsButton') }}
        </button>
      </SubSettingItem>
    </NestedSettingsContainer>
  </SettingGroup>
</template>

<style scoped></style>
