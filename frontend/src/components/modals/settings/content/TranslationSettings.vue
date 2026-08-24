<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhGlobe,
  PhTranslate,
  PhList,
  PhLink,
  PhPackage,
  PhSliders,
  PhCode,
  PhTrash,
  PhBroom,
  PhTimer,
  PhRobot,
  PhKey,
} from '@phosphor-icons/vue';
import {
  SettingGroup,
  SettingWithToggle,
  NestedSettingsContainer,
  SubSettingItem,
  TextAreaControl,
  InputControl,
  NumberControl,
  ToggleControl,
  KeyValueList,
} from '@/components/settings';
import BaseSelect from '@/components/common/BaseSelect.vue';
import AIProfileSelector from '@/components/modals/settings/ai/AIProfileSelector.vue';
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

const isClearingCache = ref(false);
const showCustomTemplates = ref(false);

// Preset templates for common translation services
const customTemplates = [
  {
    name: 'DeepLX',
    endpoint: 'http://localhost:8080/translate',
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    bodyTemplate: '{"text": %text%, "source_lang": "auto", "target_lang": "%target_lang%"}',
    responsePath: 'data',
  },
  {
    name: 'LibreTranslate',
    endpoint: 'https://libretranslate.com/translate',
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    bodyTemplate: '{"q": %text%, "source": "auto", "target": "%target_lang%", "format": "text"}',
    responsePath: 'translatedText',
  },
  {
    name: 'Argos Translate',
    endpoint: 'https://translate.argosopentech.com/translate',
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    bodyTemplate: '{"q": %text%, "source": "auto", "target": "%target_lang%"}',
    responsePath: 'translatedText',
  },
];

function applyTemplate(template: (typeof customTemplates)[0]) {
  // Convert placeholders back to {{}} format
  const bodyTemplate = template.bodyTemplate
    .replace(/%text%/g, '{{text}}')
    .replace(/%target_lang%/g, '{{target_lang}}')
    .replace(/%source_lang%/g, '{{source_lang}}');

  emit('update:settings', {
    ...props.settings,
    custom_translation_endpoint: template.endpoint,
    custom_translation_method: template.method,
    custom_translation_headers: JSON.stringify(template.headers),
    custom_translation_body_template: bodyTemplate,
    custom_translation_response_path: template.responsePath,
  });
  showCustomTemplates.value = false;
}

async function clearTranslationCache() {
  const confirmed = await window.showConfirm({
    title: t('setting.content.clearTranslationCache'),
    message: t('setting.content.clearTranslationCacheConfirm'),
    isDanger: true,
  });
  if (!confirmed) return;

  isClearingCache.value = true;
  try {
    const response = await fetch('/api/articles/clear-translations', {
      method: 'POST',
    });

    if (response.ok) {
      window.showToast(t('setting.content.clearTranslationCacheSuccess'), 'success');
      // Refresh article list to show updated translations
      window.dispatchEvent(new CustomEvent('refresh-articles'));
    } else {
      console.error('Server error:', response.status);
      window.showToast(t('setting.content.clearTranslationCacheFailed'), 'error');
    }
  } catch (error) {
    console.error('Failed to clear translation cache:', error);
    window.showToast(t('setting.content.clearTranslationCacheFailed'), 'error');
  } finally {
    isClearingCache.value = false;
  }
}
</script>

<template>
  <SettingGroup :icon="PhGlobe" :title="t('setting.content.translation')">
    <SettingWithToggle
      :icon="PhTranslate"
      :title="t('setting.content.enableTranslation')"
      :description="t('setting.content.enableTranslationDesc')"
      :model-value="settings.translation_enabled"
      @update:model-value="updateSetting('translation_enabled', $event)"
    />

    <NestedSettingsContainer v-if="settings.translation_enabled">
      <SubSettingItem
        :icon="PhTranslate"
        :title="t('setting.content.translationOnlyMode')"
        :description="t('setting.content.translationOnlyModeDesc')"
      >
        <ToggleControl
          :model-value="settings.translation_only_mode"
          @update:model-value="updateSetting('translation_only_mode', $event)"
        />
      </SubSettingItem>

      <SubSettingItem
        :icon="PhPackage"
        :title="t('setting.content.translationProvider')"
        :description="t('setting.content.translationProviderDesc')"
      >
        <BaseSelect
          :model-value="settings.translation_provider"
          :options="[
            { value: 'google', label: t('setting.content.googleTranslate') },
            { value: 'deepl', label: t('setting.content.deeplApi') },
            { value: 'baidu', label: t('setting.content.baiduTranslate') },
            { value: 'microsoft', label: t('setting.content.microsoftTranslate') },
            { value: 'tencent', label: t('setting.content.tencentTranslate') },
            { value: 'ai', label: t('setting.content.aiTranslation') },
            { value: 'custom', label: t('setting.translation.custom.title') },
          ]"
          :searchable="true"
          width="w-32 sm:w-48"
          @update:model-value="updateSetting('translation_provider', $event)"
        />
      </SubSettingItem>

      <!-- Google Translate Endpoint -->
      <SubSettingItem
        v-if="settings.translation_provider === 'google'"
        :icon="PhLink"
        :title="t('setting.content.googleTranslateEndpoint')"
        :description="t('setting.content.googleTranslateEndpointDesc')"
      >
        <BaseSelect
          :model-value="settings.google_translate_endpoint"
          :options="[
            {
              value: 'translate.googleapis.com',
              label: t('setting.content.googleTranslateEndpointDefault'),
            },
            {
              value: 'clients5.google.com',
              label: t('setting.content.googleTranslateEndpointAlternate'),
            },
          ]"
          width="w-32 sm:w-48"
          @update:model-value="updateSetting('google_translate_endpoint', $event)"
        />
      </SubSettingItem>

      <!-- DeepL API Key -->
      <SubSettingItem
        v-if="settings.translation_provider === 'deepl'"
        :icon="PhKey"
        :title="t('setting.content.deeplApiKey')"
        :description="t('setting.content.deeplApiKeyDesc')"
        :required="!settings.deepl_endpoint?.trim()"
      >
        <InputControl
          :model-value="settings.deepl_api_key"
          type="password"
          :placeholder="t('setting.content.deeplApiKeyPlaceholder')"
          :error="
            settings.translation_provider === 'deepl' &&
            !settings.deepl_api_key?.trim() &&
            !settings.deepl_endpoint?.trim()
          "
          width="md"
          @update:model-value="updateSetting('deepl_api_key', $event)"
        />
      </SubSettingItem>

      <!-- DeepL Custom Endpoint (deeplx) -->
      <SubSettingItem
        v-if="settings.translation_provider === 'deepl'"
        :icon="PhLink"
        :title="t('setting.content.deeplEndpoint')"
        :description="t('setting.content.deeplEndpointDesc')"
      >
        <InputControl
          :model-value="settings.deepl_endpoint"
          type="text"
          :placeholder="t('setting.content.deeplEndpointPlaceholder')"
          width="md"
          @update:model-value="updateSetting('deepl_endpoint', $event)"
        />
      </SubSettingItem>

      <!-- Baidu Translate Settings -->
      <template v-if="settings.translation_provider === 'baidu'">
        <SubSettingItem
          :icon="PhKey"
          :title="t('setting.content.baiduAppId')"
          :description="t('setting.content.baiduAppIdDesc')"
          required
        >
          <InputControl
            :model-value="settings.baidu_app_id"
            type="text"
            :placeholder="t('setting.content.baiduAppIdPlaceholder')"
            :error="settings.translation_provider === 'baidu' && !settings.baidu_app_id?.trim()"
            width="md"
            @update:model-value="updateSetting('baidu_app_id', $event)"
          />
        </SubSettingItem>

        <SubSettingItem
          :icon="PhKey"
          :title="t('setting.content.baiduSecretKey')"
          :description="t('setting.content.baiduSecretKeyDesc')"
          required
        >
          <InputControl
            :model-value="settings.baidu_secret_key"
            type="password"
            :placeholder="t('setting.content.baiduSecretKeyPlaceholder')"
            :error="settings.translation_provider === 'baidu' && !settings.baidu_secret_key?.trim()"
            width="md"
            @update:model-value="updateSetting('baidu_secret_key', $event)"
          />
        </SubSettingItem>
      </template>

      <!-- Microsoft Translate Settings -->
      <template v-if="settings.translation_provider === 'microsoft'">
        <SubSettingItem
          :icon="PhKey"
          :title="t('setting.content.microsoftApiKey')"
          :description="t('setting.content.microsoftApiKeyDesc')"
          required
        >
          <InputControl
            :model-value="settings.microsoft_api_key"
            type="password"
            :placeholder="t('setting.content.microsoftApiKeyPlaceholder')"
            :error="
              settings.translation_provider === 'microsoft' && !settings.microsoft_api_key?.trim()
            "
            width="md"
            @update:model-value="updateSetting('microsoft_api_key', $event)"
          />
        </SubSettingItem>

        <SubSettingItem
          :icon="PhLink"
          :title="t('setting.content.microsoftRegion')"
          :description="t('setting.content.microsoftRegionDesc')"
        >
          <InputControl
            :model-value="settings.microsoft_region"
            type="text"
            :placeholder="t('setting.content.microsoftRegionPlaceholder')"
            width="md"
            @update:model-value="updateSetting('microsoft_region', $event)"
          />
        </SubSettingItem>

        <SubSettingItem
          :icon="PhLink"
          :title="t('setting.content.microsoftEndpoint')"
          :description="t('setting.content.microsoftEndpointDesc')"
        >
          <InputControl
            :model-value="settings.microsoft_endpoint"
            type="text"
            :placeholder="t('setting.content.microsoftEndpointPlaceholder')"
            width="lg"
            @update:model-value="updateSetting('microsoft_endpoint', $event)"
          />
        </SubSettingItem>
      </template>

      <!-- Tencent Translate Settings -->
      <template v-if="settings.translation_provider === 'tencent'">
        <SubSettingItem
          :icon="PhKey"
          :title="t('setting.content.tencentSecretId')"
          :description="t('setting.content.tencentSecretIdDesc')"
          required
        >
          <InputControl
            :model-value="settings.tencent_secret_id"
            type="text"
            :placeholder="t('setting.content.tencentSecretIdPlaceholder')"
            :error="
              settings.translation_provider === 'tencent' && !settings.tencent_secret_id?.trim()
            "
            width="md"
            @update:model-value="updateSetting('tencent_secret_id', $event)"
          />
        </SubSettingItem>

        <SubSettingItem
          :icon="PhKey"
          :title="t('setting.content.tencentSecretKey')"
          :description="t('setting.content.tencentSecretKeyDesc')"
          required
        >
          <InputControl
            :model-value="settings.tencent_secret_key"
            type="password"
            :placeholder="t('setting.content.tencentSecretKeyPlaceholder')"
            :error="
              settings.translation_provider === 'tencent' && !settings.tencent_secret_key?.trim()
            "
            width="md"
            @update:model-value="updateSetting('tencent_secret_key', $event)"
          />
        </SubSettingItem>

        <SubSettingItem
          :icon="PhGlobe"
          :title="t('setting.content.tencentRegion')"
          :description="t('setting.content.tencentRegionDesc')"
        >
          <BaseSelect
            :model-value="settings.tencent_region || 'ap-guangzhou'"
            :options="[
              { value: 'ap-guangzhou', label: ' Guangzhou (ap-guangzhou)' },
              { value: 'ap-shanghai', label: 'Shanghai (ap-shanghai)' },
              { value: 'ap-beijing', label: 'Beijing (ap-beijing)' },
              { value: 'ap-chengdu', label: 'Chengdu (ap-chengdu)' },
              { value: 'ap-hongkong', label: 'Hong Kong (ap-hongkong)' },
              { value: 'ap-singapore', label: 'Singapore (ap-singapore)' },
              { value: 'ap-tokyo', label: 'Tokyo (ap-tokyo)' },
              { value: 'na-toronto', label: 'Toronto (na-toronto)' },
              { value: 'na-ashburn', label: 'Ashburn (na-ashburn)' },
              { value: 'na-siliconvalley', label: 'Silicon Valley (na-siliconvalley)' },
              { value: 'eu-frankfurt', label: 'Frankfurt (eu-frankfurt)' },
            ]"
            width="w-40 sm:w-64"
            @update:model-value="updateSetting('tencent_region', $event)"
          />
        </SubSettingItem>
      </template>

      <!-- AI Translation Prompt -->
      <template v-if="settings.translation_provider === 'ai'">
        <!-- AI Profile Selection -->
        <SubSettingItem
          :icon="PhRobot"
          :title="t('setting.ai.selectProfile')"
          :description="t('setting.ai.selectProfileForTranslation')"
        >
          <AIProfileSelector
            :model-value="settings.ai_translation_profile_id"
            @update:model-value="updateSetting('ai_translation_profile_id', $event)"
          />
        </SubSettingItem>

        <div class="sub-setting-item-col">
          <div class="flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
            <PhRobot :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
            <div class="flex-1 min-w-0">
              <div class="font-medium mb-0 sm:mb-1 text-xs sm:text-sm">
                {{ t('setting.content.aiTranslationPrompt') }}
              </div>
              <div class="text-[10px] sm:text-xs text-text-secondary hidden sm:block">
                {{ t('setting.content.aiTranslationPromptDesc') }}
              </div>
            </div>
          </div>
          <TextAreaControl
            :model-value="settings.ai_translation_prompt"
            :placeholder="t('setting.content.aiTranslationPromptPlaceholder')"
            :rows="3"
            @update:model-value="updateSetting('ai_translation_prompt', $event)"
          />
        </div>
      </template>

      <!-- Custom Translation Provider -->
      <template v-if="settings.translation_provider === 'custom'">
        <!-- Template Selection -->
        <div class="sub-setting-item">
          <div class="flex items-center sm:items-start justify-between gap-2 sm:gap-4 w-full">
            <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
              <PhList :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
              <div class="flex-1 min-w-0">
                <div class="font-medium mb-0 sm:mb-1 text-xs sm:text-sm">
                  {{ t('setting.translation.custom.template') }}
                </div>
                <div class="text-[10px] sm:text-xs text-text-secondary hidden sm:block">
                  {{ t('setting.translation.custom.templateDesc') }}
                </div>
              </div>
            </div>
            <div class="relative shrink-0">
              <button
                type="button"
                class="btn-secondary"
                @click="showCustomTemplates = !showCustomTemplates"
              >
                {{ t('setting.content.custom.selectTemplate') || 'Select Template' }}
              </button>
              <div
                v-if="showCustomTemplates"
                class="absolute top-full right-0 mt-1 z-50 bg-bg-secondary border border-border rounded-lg shadow-lg overflow-hidden"
              >
                <button
                  v-for="tmpl in customTemplates"
                  :key="tmpl.name"
                  type="button"
                  class="w-full px-4 py-2 text-left hover:bg-bg-tertiary text-sm"
                  @click="applyTemplate(tmpl)"
                >
                  {{ tmpl.name }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Custom Translation Endpoint -->
        <SubSettingItem
          :icon="PhLink"
          :title="t('setting.translation.custom.endpoint')"
          :description="t('setting.translation.custom.endpointDesc')"
          required
        >
          <InputControl
            :model-value="settings.custom_translation_endpoint"
            type="text"
            :placeholder="t('setting.translation.custom.endpointPlaceholder')"
            :error="
              settings.translation_provider === 'custom' &&
              !settings.custom_translation_endpoint?.trim()
            "
            width="lg"
            @update:model-value="updateSetting('custom_translation_endpoint', $event)"
          />
        </SubSettingItem>

        <!-- Custom Translation Method -->
        <SubSettingItem
          :icon="PhCode"
          :title="t('setting.translation.custom.method')"
          :description="t('setting.translation.custom.methodDesc')"
        >
          <BaseSelect
            :model-value="settings.custom_translation_method || 'POST'"
            :options="[
              { value: 'GET', label: 'GET' },
              { value: 'POST', label: 'POST' },
            ]"
            width="w-24 sm:w-32"
            @update:model-value="updateSetting('custom_translation_method', $event)"
          />
        </SubSettingItem>

        <!-- Custom Translation Headers -->
        <div class="sub-setting-item-col">
          <div class="flex items-center gap-2 sm:gap-3">
            <PhSliders :size="20" class="text-text-secondary shrink-0 sm:w-6 sm:h-6" />
            <div class="flex-1 min-w-0">
              <div class="font-medium text-sm">
                {{ t('setting.translation.custom.headers') }}
              </div>
              <div class="text-xs text-text-secondary">
                {{ t('setting.translation.custom.headersDesc') }}
              </div>
            </div>
          </div>

          <KeyValueList
            :model-value="settings.custom_translation_headers"
            :key-placeholder="t('setting.content.custom.headerName') || 'Header name'"
            :value-placeholder="t('setting.content.custom.headerValue') || 'Value'"
            :add-button-text="t('setting.content.addHeader')"
            :remove-button-title="t('common.action.remove') || 'Remove'"
            ascii-only
            @update:model-value="updateSetting('custom_translation_headers', $event)"
          />
        </div>

        <!-- Custom Translation Body Template -->
        <div
          v-if="(settings.custom_translation_method || 'POST') === 'POST'"
          class="sub-setting-item-col"
        >
          <div class="flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
            <PhCode :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
            <div class="flex-1 min-w-0">
              <div class="font-medium mb-0 sm:mb-1 text-sm">
                {{ t('setting.translation.custom.bodyTemplate') }}
                <span class="text-red-500">*</span>
              </div>
              <div class="text-xs text-text-secondary hidden sm:block">
                {{ t('setting.translation.custom.bodyTemplateDesc') }}
              </div>
            </div>
          </div>
          <TextAreaControl
            :model-value="settings.custom_translation_body_template"
            :placeholder="
              t('setting.translation.custom.bodyTemplatePlaceholder') ||
              'Enter request body template'
            "
            :rows="4"
            :resize="false"
            :font-mono="true"
            @update:model-value="updateSetting('custom_translation_body_template', $event)"
          />
        </div>

        <!-- Custom Translation Response Path -->
        <SubSettingItem
          :icon="PhCode"
          :title="t('setting.translation.custom.responsePath')"
          :description="t('setting.translation.custom.responsePathDesc')"
          required
        >
          <InputControl
            :model-value="settings.custom_translation_response_path"
            type="text"
            :placeholder="t('setting.translation.custom.responsePathPlaceholder')"
            :error="
              settings.translation_provider === 'custom' &&
              !settings.custom_translation_response_path?.trim()
            "
            width="lg"
            @update:model-value="updateSetting('custom_translation_response_path', $event)"
          />
        </SubSettingItem>

        <!-- Custom Translation Language Mapping -->
        <div class="sub-setting-item-col">
          <div class="flex items-center gap-2 sm:gap-3">
            <PhGlobe :size="20" class="text-text-secondary shrink-0 sm:w-6 sm:h-6" />
            <div class="flex-1 min-w-0">
              <div class="font-medium text-sm">
                {{ t('setting.translation.custom.langMapping') }}
              </div>
              <div class="text-xs text-text-secondary">
                {{ t('setting.translation.custom.langMappingDesc') }}
              </div>
            </div>
          </div>

          <KeyValueList
            :model-value="settings.custom_translation_lang_mapping"
            :key-placeholder="
              t('setting.content.custom.mrssLangCode') || 'MrRSS code (en, zh, ...)'
            "
            :value-placeholder="t('setting.content.apiLangCode') || 'API code'"
            :add-button-text="t('setting.content.addLangMapping')"
            :remove-button-title="t('common.action.remove') || 'Remove'"
            @update:model-value="updateSetting('custom_translation_lang_mapping', $event)"
          />
        </div>

        <!-- Custom Translation Timeout -->
        <SubSettingItem
          :icon="PhTimer"
          :title="t('setting.translation.custom.timeout')"
          :description="t('setting.translation.custom.timeoutDesc')"
        >
          <NumberControl
            :model-value="settings.custom_translation_timeout || 10"
            :min="1"
            :max="60"
            :suffix="t('common.time.seconds')"
            width="sm"
            @update:model-value="updateSetting('custom_translation_timeout', $event)"
          />
        </SubSettingItem>
      </template>

      <SubSettingItem
        :icon="PhGlobe"
        :title="t('setting.content.targetLanguage')"
        :description="t('setting.content.targetLanguageDesc')"
      >
        <BaseSelect
          :model-value="settings.target_language"
          :options="[
            { value: 'en', label: t('common.language.english') },
            { value: 'es', label: t('common.language.spanish') },
            { value: 'fr', label: t('common.language.french') },
            { value: 'de', label: t('common.language.german') },
            { value: 'zh', label: t('common.language.simplifiedChinese') },
            { value: 'zh-TW', label: t('common.language.traditionalChinese') },
            { value: 'ja', label: t('common.language.japanese') },
            { value: 'tr', label: t('common.language.turkish') },
          ]"
          width="w-24 sm:w-48"
          @update:model-value="updateSetting('target_language', $event)"
        />
      </SubSettingItem>

      <!-- Cache Management -->
      <SubSettingItem
        :icon="PhTrash"
        :title="t('setting.content.clearTranslationCache')"
        :description="t('setting.content.clearTranslationCacheDesc')"
      >
        <button
          type="button"
          :disabled="isClearingCache"
          class="btn-secondary"
          @click="clearTranslationCache"
        >
          <PhBroom :size="16" class="sm:w-5 sm:h-5" />
          {{
            isClearingCache
              ? t('setting.database.cleaning')
              : t('setting.content.clearTranslationCacheButton')
          }}
        </button>
      </SubSettingItem>
    </NestedSettingsContainer>
  </SettingGroup>
</template>

<style scoped>
/* Styles are now handled by BaseSelect and select.css */
</style>
