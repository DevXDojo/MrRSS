<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhRobot,
  PhPlus,
  PhPencil,
  PhTrash,
  PhArrowClockwise,
  PhCheck,
  PhX,
} from '@phosphor-icons/vue';
import { SettingGroup } from '@/components/settings';
import type { AIProfile, AIProfileTestResult, AIProfileFormData } from '@/types/aiProfile';
import { useAIProfiles } from '@/composables/ai/useAIProfiles';
import { getProviderIconUrl } from '@/composables/ai/useAIProvider';
import AIProfileModal from './AIProfileModal.vue';

const { t } = useI18n();
const {
  profiles,
  isLoading,
  fetchProfiles,
  getProfile,
  deleteProfile,
  testProfile,
  testAllProfiles,
  profileToFormData,
} = useAIProfiles();

// Modal state
const isModalOpen = ref(false);
const editProfileId = ref<number | null>(null);
const editProfileData = ref<AIProfileFormData | undefined>(undefined);

// Test state
const testResults = ref<Map<number, AIProfileTestResult>>(new Map());
const testingProfiles = ref<Set<number>>(new Set());
const isTestingAll = ref(false);

// Load profiles on mount
onMounted(() => {
  fetchProfiles();
});

// Open modal for adding new profile
function openAddModal() {
  editProfileId.value = null;
  editProfileData.value = undefined;
  isModalOpen.value = true;
}

// Open modal for editing existing profile
async function openEditModal(profile: AIProfile) {
  editProfileId.value = profile.id;
  // Fetch full profile data (with masked API key)
  const fullProfile = await getProfile(profile.id);
  if (fullProfile) {
    editProfileData.value = profileToFormData(fullProfile);
  } else {
    editProfileData.value = profileToFormData(profile);
  }
  isModalOpen.value = true;
}

// Close modal
function closeModal() {
  isModalOpen.value = false;
  editProfileId.value = null;
  editProfileData.value = undefined;
}

// Handle modal save
function handleModalSaved() {
  fetchProfiles();
}

// Delete profile with confirmation
async function handleDelete(profile: AIProfile) {
  const confirmed = await window.showConfirm({
    title: t('setting.ai.deleteProfileTitle'),
    message: t('setting.ai.deleteProfileConfirm', { name: profile.name }),
    isDanger: true,
  });

  if (confirmed) {
    const success = await deleteProfile(profile.id);
    if (success) {
      window.showToast(t('setting.ai.profileDeleted'), 'success');
      // Clear test result for this profile
      testResults.value.delete(profile.id);
    } else {
      window.showToast(t('setting.ai.deleteProfileFailed'), 'error');
    }
  }
}

// Test single profile
async function handleTestProfile(profile: AIProfile) {
  testingProfiles.value.add(profile.id);

  try {
    const result = await testProfile(profile.id);
    if (result) {
      testResults.value.set(profile.id, result);
    }
  } finally {
    testingProfiles.value.delete(profile.id);
  }
}

// Test all profiles
async function handleTestAllProfiles() {
  isTestingAll.value = true;
  testResults.value.clear();

  try {
    const results = await testAllProfiles();
    for (const result of results) {
      testResults.value.set(result.profile_id, result);
    }
  } finally {
    isTestingAll.value = false;
  }
}

// Get test result status for display
function getTestStatus(profileId: number): 'success' | 'error' | 'unknown' {
  const result = testResults.value.get(profileId);
  if (!result) return 'unknown';
  return result.config_valid && result.connection_success ? 'success' : 'error';
}
</script>

<template>
  <SettingGroup :icon="PhRobot" :title="t('setting.ai.aiProfiles')">
    <!-- Action Buttons -->
    <div class="flex flex-wrap items-center gap-2">
      <button type="button" class="btn-secondary" @click="openAddModal">
        <PhPlus :size="16" />
        {{ t('setting.ai.addProfile') }}
      </button>

      <button
        type="button"
        class="btn-secondary"
        :disabled="isTestingAll || profiles.length === 0"
        @click="handleTestAllProfiles"
      >
        <PhArrowClockwise :size="16" :class="{ 'animate-spin': isTestingAll }" />
        {{ isTestingAll ? t('setting.ai.testingAll') : t('setting.ai.testAllProfiles') }}
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading && profiles.length === 0" class="py-8 text-center text-text-secondary">
      <PhArrowClockwise :size="24" class="animate-spin mx-auto mb-2" />
      {{ t('common.state.loading') }}
    </div>

    <!-- Empty State -->
    <div v-else-if="profiles.length === 0" class="py-8 text-center">
      <PhRobot :size="48" class="mx-auto mb-3 text-text-tertiary" />
      <div class="text-text-secondary mb-2">{{ t('setting.ai.noProfiles') }}</div>
      <div class="text-xs text-text-tertiary">{{ t('setting.ai.noProfilesHint') }}</div>
    </div>

    <!-- Profiles List -->
    <div v-else class="space-y-2">
      <div v-for="profile in profiles" :key="profile.id" class="profile-item">
        <div class="flex items-center gap-2 sm:gap-4">
          <!-- AI Provider Logo -->
          <div class="shrink-0 w-8 h-8 flex items-center justify-center">
            <img
              v-if="getProviderIconUrl(profile.model) ?? undefined"
              :src="getProviderIconUrl(profile.model) ?? undefined"
              :alt="profile.name"
              class="w-7 h-7 object-contain"
            />
            <PhRobot v-else :size="28" class="text-text-tertiary" />
          </div>

          <!-- Profile Info -->
          <div class="flex-1 min-w-0">
            <div class="font-medium text-sm sm:text-base truncate">
              {{ profile.name }}
            </div>
            <div class="mt-1.5 flex flex-wrap gap-1.5">
              <div
                class="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-bg-tertiary text-xs"
              >
                <span class="text-text-tertiary">{{ t('setting.ai.endpoint') }}</span>
                <span class="text-text-secondary font-mono">{{ profile.endpoint }}</span>
              </div>
              <div
                class="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-bg-tertiary text-xs"
              >
                <span class="text-text-tertiary">{{ t('setting.ai.model') }}</span>
                <span class="text-text-secondary font-medium">{{ profile.model }}</span>
              </div>
            </div>
          </div>

          <!-- Test Result Indicator (only show when tested or testing) -->
          <div class="shrink-0">
            <div v-if="testingProfiles.has(profile.id)" class="status-indicator">
              <PhArrowClockwise :size="16" class="animate-spin text-text-secondary" />
            </div>
            <div
              v-else-if="getTestStatus(profile.id) === 'success'"
              class="status-indicator status-success"
              :title="t('setting.ai.connectionSuccess')"
            >
              <PhCheck :size="14" class="text-green-500" />
            </div>
            <div
              v-else-if="getTestStatus(profile.id) === 'error'"
              class="status-indicator status-error"
              :title="testResults.get(profile.id)?.error_message"
            >
              <PhX :size="14" class="text-red-500" />
            </div>
          </div>

          <!-- Action buttons -->
          <div class="flex items-center gap-1 sm:gap-2 shrink-0">
            <button
              class="action-btn"
              :disabled="testingProfiles.has(profile.id)"
              :title="t('setting.ai.testProfile')"
              @click="handleTestProfile(profile)"
            >
              <PhTestTube :size="18" class="sm:w-5 sm:h-5" />
            </button>
            <button
              class="action-btn"
              :title="t('setting.ai.editProfile')"
              @click="openEditModal(profile)"
            >
              <PhPencil :size="18" class="sm:w-5 sm:h-5" />
            </button>
            <button
              class="action-btn danger"
              :title="t('setting.ai.deleteProfile')"
              @click="handleDelete(profile)"
            >
              <PhTrash :size="18" class="sm:w-5 sm:h-5" />
            </button>
          </div>
        </div>

        <!-- Test Result Details (shown when there's an error) -->
        <div
          v-if="testResults.get(profile.id)?.error_message && getTestStatus(profile.id) === 'error'"
          class="mt-2 text-xs text-red-500 bg-red-500/5 rounded p-2 break-words"
        >
          {{ testResults.get(profile.id)?.error_message }}
        </div>
      </div>
    </div>
  </SettingGroup>

  <!-- Profile Modal -->
  <AIProfileModal
    :is-open="isModalOpen"
    :edit-profile-id="editProfileId"
    :initial-data="editProfileData"
    @close="closeModal"
    @saved="handleModalSaved"
  />
</template>

<style scoped>
@reference "../../../../style.css";
/* Profile item styling - similar to RuleItem */
.profile-item {
  @apply p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border transition-all;
}

.profile-item:hover {
  @apply bg-bg-tertiary;
}

/* Info badges */
.info-badge {
  @apply inline-flex items-center px-1.5 sm:px-2 py-0.5 rounded text-[10px] sm:text-xs bg-bg-tertiary;
}

/* Status indicator */
.status-indicator {
  @apply w-6 h-6 flex items-center justify-center rounded-full bg-bg-tertiary;
}

.status-indicator.status-success {
  @apply bg-green-500/10;
}

.status-indicator.status-error {
  @apply bg-red-500/10;
}

/* Action buttons */
.action-btn {
  @apply p-1.5 sm:p-2 rounded-lg bg-transparent border-none cursor-pointer text-text-secondary hover:bg-bg-tertiary hover:text-text-primary transition-all;
}

.action-btn.danger:hover {
  @apply text-red-500 bg-red-500/10;
}

.action-btn:disabled {
  @apply opacity-50 cursor-not-allowed;
}

/* Loading spinner */
.animate-spin {
  animation: spin 1s linear infinite;
  display: inline-block;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
