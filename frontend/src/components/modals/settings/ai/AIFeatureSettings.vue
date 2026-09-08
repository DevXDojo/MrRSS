<script setup lang="ts">
import { nextTick, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhRobot,
  PhChatCircleText,
  PhTrash,
  PhBroom,
  PhMagnifyingGlass,
  PhLightning,
} from '@phosphor-icons/vue';
import {
  TipBox,
  SettingGroup,
  SettingWithToggle,
  NestedSettingsContainer,
  SubSettingItem,
  TextAreaControl,
} from '@/components/settings';
import AIProfileSelector from './AIProfileSelector.vue';
import AIChatQuickPromptsSettings from './AIChatQuickPromptsSettings.vue';
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

async function updateSetting(key: keyof SettingsData, value: any) {
  const anchor =
    document.activeElement instanceof HTMLElement
      ? document.activeElement.closest<HTMLElement>('.setting-item')
      : null;
  const scrollContainer = anchor?.closest<HTMLElement>('[data-settings-content]');
  const scrollTop = scrollContainer?.scrollTop;

  emit('update:settings', {
    ...props.settings,
    [key]: value,
  });

  await nextTick();
  if (scrollContainer && scrollTop !== undefined) {
    scrollContainer.scrollTop = scrollTop;
  }
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

      <div class="sub-setting-item-col">
        <label for="ai-chat-response-preferences" class="font-medium text-xs sm:text-sm">
          {{ t('setting.ai.responsePreferences') }}
        </label>
        <p class="text-text-secondary text-xs">
          {{ t('setting.ai.responsePreferencesDesc') }}
        </p>
        <TextAreaControl
          id="ai-chat-response-preferences"
          :model-value="props.settings.ai_chat_response_preferences"
          :placeholder="t('setting.ai.responsePreferencesPlaceholder')"
          :rows="4"
          @update:model-value="updateSetting('ai_chat_response_preferences', $event)"
        />
      </div>

      <SubSettingItem
        :icon="PhLightning"
        :title="t('setting.ai.quickPrompts')"
        :description="t('setting.ai.quickPromptsDesc')"
      >
        <AIChatQuickPromptsSettings
          :model-value="props.settings.ai_chat_quick_prompts"
          @update:model-value="updateSetting('ai_chat_quick_prompts', $event)"
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
