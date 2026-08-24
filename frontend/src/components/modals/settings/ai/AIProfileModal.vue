<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhRobot,
  PhKey,
  PhLink,
  PhBrain,
  PhSliders,
  PhArrowClockwise,
  PhBookOpen,
} from '@phosphor-icons/vue';
import { SettingItem, KeyValueList, StatusBoxGroup } from '@/components/settings';
import BaseModal from '@/components/common/BaseModal.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';
import type { AIProfileFormData, AIProfileTestResult } from '@/types/aiProfile';
import type { Status } from '@/components/settings/base/StatusBox.vue';
import { defaultAIProfileFormData } from '@/types/aiProfile';
import { useAIProfiles } from '@/composables/ai/useAIProfiles';
import { openInBrowser } from '@/utils/browser';

const { t, locale } = useI18n();
const { testProfile, testConfig, createProfile, updateProfile } = useAIProfiles();

interface Props {
  isOpen: boolean;
  editProfileId?: number | null;
  initialData?: AIProfileFormData;
}

const props = withDefaults(defineProps<Props>(), {
  editProfileId: null,
  initialData: undefined,
});

const emit = defineEmits<{
  close: [];
  saved: [profileId: number];
}>();

// Form data
const formData = ref<AIProfileFormData>({ ...defaultAIProfileFormData });
const isSaving = ref(false);
const saveError = ref<string | null>(null);

// Test state
const isTesting = ref(false);
const testResult = ref<AIProfileTestResult | null>(null);
const testError = ref<string | null>(null);

// Computed
const isEditMode = computed(
  () => props.editProfileId !== null && props.editProfileId !== undefined
);
const modalTitle = computed(() =>
  isEditMode.value ? t('setting.ai.editProfile') : t('setting.ai.addProfile')
);

// Initialize form data when modal opens or editProfileId changes
watch(
  () => [props.isOpen, props.editProfileId, props.initialData],
  async () => {
    if (props.isOpen) {
      if (props.initialData) {
        formData.value = { ...props.initialData };
      } else {
        formData.value = { ...defaultAIProfileFormData };
      }
      // Reset test state
      testResult.value = null;
      testError.value = null;
      saveError.value = null;
    }
  },
  { immediate: true }
);

// Test the current configuration
async function testConfiguration() {
  if (!formData.value.endpoint || !formData.value.model) {
    testError.value = t('setting.ai.configIncomplete');
    return;
  }

  isTesting.value = true;
  testError.value = null;
  testResult.value = null;

  try {
    let result: AIProfileTestResult | null;

    if (isEditMode.value && props.editProfileId && !formData.value.api_key.startsWith('****')) {
      // For existing profiles with unchanged API key, test via profile ID
      result = await testProfile(props.editProfileId);
    } else {
      // For new profiles or changed API keys, test the configuration directly
      result = await testConfig({
        api_key: formData.value.api_key,
        endpoint: formData.value.endpoint,
        model: formData.value.model,
        custom_headers: formData.value.custom_headers,
      });
    }

    if (result) {
      testResult.value = result;
      if (!result.config_valid || !result.connection_success) {
        testError.value = result.error_message || t('setting.ai.aiTestFailed');
      }
    } else {
      testError.value = t('setting.ai.aiTestFailed');
    }
  } catch (e) {
    console.error('Test failed:', e);
    testError.value = t('setting.ai.aiTestFailed');
  } finally {
    isTesting.value = false;
  }
}

// Save the profile
async function saveProfile() {
  // Validate required fields
  if (!formData.value.name.trim()) {
    saveError.value = t('setting.ai.nameRequired');
    return;
  }
  if (!formData.value.endpoint.trim()) {
    saveError.value = t('setting.ai.endpointRequired');
    return;
  }
  if (!formData.value.model.trim()) {
    saveError.value = t('setting.ai.modelRequired');
    return;
  }

  isSaving.value = true;
  saveError.value = null;

  try {
    let savedProfile;
    if (isEditMode.value && props.editProfileId) {
      savedProfile = await updateProfile(props.editProfileId, formData.value);
    } else {
      savedProfile = await createProfile(formData.value);
    }

    if (savedProfile) {
      window.showToast(
        isEditMode.value ? t('setting.ai.profileUpdated') : t('setting.ai.profileCreated'),
        'success'
      );
      emit('saved', savedProfile.id);
      emit('close');
    }
  } catch (e) {
    console.error('Save failed:', e);
    saveError.value = e instanceof Error ? e.message : t('setting.ai.saveFailed');
  } finally {
    isSaving.value = false;
  }
}

// Open documentation
function openDocumentation() {
  const docUrl = locale.value.startsWith('zh')
    ? 'https://github.com/DevXDojo/MrRSS/blob/main/docs/AI_CONFIGURATION.zh.md'
    : 'https://github.com/DevXDojo/MrRSS/blob/main/docs/AI_CONFIGURATION.md';
  openInBrowser(docUrl);
}

// Status display for test results
const testStatuses = computed((): Status[] => {
  return [
    {
      label: t('setting.ai.configValid'),
      value: testResult.value
        ? testResult.value.config_valid
          ? t('common.action.yes')
          : t('common.action.no')
        : '-',
      type: testResult.value
        ? testResult.value.config_valid
          ? ('success' as const)
          : ('error' as const)
        : ('neutral' as const),
    },
    {
      label: t('setting.ai.connectionSuccess'),
      value: testResult.value
        ? testResult.value.connection_success
          ? t('common.action.yes')
          : t('common.action.no')
        : '-',
      type: testResult.value
        ? testResult.value.connection_success
          ? ('success' as const)
          : ('error' as const)
        : ('neutral' as const),
    },
    {
      label: t('setting.ai.responseTime'),
      value:
        testResult.value && testResult.value.response_time_ms > 0
          ? testResult.value.response_time_ms
          : '-',
      unit: testResult.value && testResult.value.response_time_ms > 0 ? t('common.time.ms') : '',
      type:
        testResult.value && testResult.value.response_time_ms > 0
          ? ('neutral' as const)
          : ('neutral' as const),
    },
  ];
});

// Handle close
function handleClose() {
  emit('close');
}
</script>

<template>
  <Teleport to="body">
    <BaseModal v-if="isOpen" :title="modalTitle" size="2xl" :z-index="60" @close="handleClose">
      <!-- Form Content -->
      <div class="p-4 sm:p-6 space-y-4">
        <!-- Profile Name -->
        <SettingItem
          :icon="PhRobot"
          :title="t('setting.ai.profileName')"
          :description="t('setting.ai.profileNameDesc')"
          required
        >
          <input
            v-model="formData.name"
            type="text"
            :placeholder="t('setting.ai.profileNamePlaceholder')"
            class="input-field w-full sm:w-48 text-xs sm:text-sm"
          />
        </SettingItem>

        <!-- API Key -->
        <SettingItem
          :icon="PhKey"
          :title="t('setting.ai.aiApiKey')"
          :description="t('setting.ai.aiApiKeyDesc')"
        >
          <input
            v-model="formData.api_key"
            type="password"
            :placeholder="t('setting.ai.aiApiKeyPlaceholder')"
            class="input-field w-full sm:w-48 text-xs sm:text-sm"
          />
        </SettingItem>

        <!-- Endpoint -->
        <SettingItem
          :icon="PhLink"
          :title="t('setting.ai.aiEndpoint')"
          :description="t('setting.ai.aiEndpointDesc')"
          required
        >
          <input
            v-model="formData.endpoint"
            type="text"
            :placeholder="t('setting.ai.aiEndpointPlaceholder')"
            class="input-field w-full sm:w-48 text-xs sm:text-sm"
          />
        </SettingItem>

        <!-- Model -->
        <SettingItem
          :icon="PhBrain"
          :title="t('setting.ai.aiModel')"
          :description="t('setting.ai.aiModelDesc')"
          required
        >
          <input
            v-model="formData.model"
            type="text"
            :placeholder="t('setting.ai.aiModelPlaceholder')"
            class="input-field w-full sm:w-48 text-xs sm:text-sm"
          />
        </SettingItem>

        <!-- Custom Headers -->
        <div class="setting-item-col">
          <div class="flex items-center gap-2 sm:gap-3">
            <PhSliders :size="20" class="text-text-secondary shrink-0 sm:w-6 sm:h-6" />
            <div class="flex-1 min-w-0">
              <div class="font-medium text-sm">{{ t('setting.ai.aiCustomHeaders') }}</div>
              <div class="text-xs text-text-secondary hidden sm:block">
                {{ t('setting.ai.aiCustomHeadersDesc') }}
              </div>
            </div>
          </div>

          <KeyValueList
            v-model="formData.custom_headers"
            :key-placeholder="t('setting.ai.aiCustomHeadersName')"
            :value-placeholder="t('setting.ai.aiCustomHeadersValue')"
            :add-button-text="t('setting.ai.aiCustomHeadersAdd')"
            :remove-button-title="t('setting.ai.aiCustomHeadersRemove')"
            ascii-only
          />
        </div>

        <!-- Test Configuration Section -->
        <div class="border-t border-border pt-4 mt-4">
          <!-- Test Status Display -->
          <StatusBoxGroup
            :statuses="testStatuses"
            :action-button="{
              label: isTesting ? t('setting.ai.testing') : t('setting.ai.testAIConfig'),
              icon: PhArrowClockwise,
              loading: isTesting,
              disabled: !formData.endpoint || !formData.model,
              onClick: testConfiguration,
            }"
          />

          <!-- Test Error -->
          <div
            v-if="testError"
            class="bg-red-500/10 border border-red-500/30 rounded-lg p-2 sm:p-3 text-xs sm:text-sm text-red-500 mt-3"
          >
            {{ testError }}
          </div>

          <!-- Test Success (when all checks pass) -->
          <div
            v-if="testResult?.config_valid && testResult?.connection_success && !testError"
            class="bg-green-500/10 border border-green-500/30 rounded-lg p-2 sm:p-3 text-xs sm:text-sm text-green-500 mt-3"
          >
            {{ t('setting.ai.aiConfigAllGood') }}
          </div>
        </div>

        <!-- Documentation Link -->
        <div class="mt-3">
          <button
            type="button"
            class="text-xs sm:text-sm text-accent hover:underline flex items-center gap-1"
            @click="openDocumentation"
          >
            <PhBookOpen :size="14" />
            {{ t('setting.ai.aiConfigurationGuide') }}
          </button>
        </div>
      </div>

      <!-- Footer -->
      <template #footer>
        <div class="flex items-center justify-between">
          <!-- Error Message -->
          <div v-if="saveError" class="text-xs sm:text-sm text-red-500 flex-1 mr-4">
            {{ saveError }}
          </div>
          <div v-else class="flex-1" />

          <!-- Actions -->
          <ModalFooter
            align="right"
            :secondary-button="{
              label: t('common.action.cancel'),
              onClick: handleClose,
            }"
            :primary-button="{
              label: isSaving ? t('common.action.saving') : t('common.action.save'),
              disabled: isSaving,
              loading: isSaving,
              onClick: saveProfile,
            }"
          />
        </div>
      </template>
    </BaseModal>
  </Teleport>
</template>

<style scoped>
@reference "../../../../style.css";
.input-field {
  @apply p-1.5 sm:p-2.5 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors;
}
</style>
