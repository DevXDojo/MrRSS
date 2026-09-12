<script setup lang="ts">
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhLink, PhUser, PhKey, PhArrowClockwise, PhCloudCheck } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';
import { useAppStore } from '@/stores/app';
import { NestedSettingsContainer, SubSettingItem, InputControl } from '@/components/settings';
import ReaderProviderIcon from '@/components/common/ReaderProviderIcon.vue';

const props = defineProps<{ settings: SettingsData; provider: 'freshrss' | 'miniflux' }>();
const emit = defineEmits<{ 'update:settings': [settings: SettingsData] }>();
const { t } = useI18n();
const store = useAppStore();
const name = computed(() => (props.provider === 'miniflux' ? 'Miniflux' : 'FreshRSS'));
type Field = 'enabled' | 'server_url' | 'username' | 'api_password';
function key(field: Field) {
  return `${props.provider}_${field}` as const;
}
const enabled = computed(() => props.settings[`${props.provider}_enabled`]);
function update(field: Field, value: string | boolean) {
  emit('update:settings', { ...props.settings, [key(field)]: value });
}
async function toggle(event: Event) {
  const target = event.target as HTMLInputElement;
  if (
    !target.checked &&
    !(await window.showConfirm({
      title: name.value,
      message: t('setting.freshrss.disableConfirm'),
      isDanger: true,
    }))
  ) {
    target.checked = true;
    return;
  }
  update('enabled', target.checked);
}
const syncing = ref(false);
const lastSync = ref<string | null>(null);
const controller = new AbortController();
let poll: ReturnType<typeof setInterval> | undefined;
let fetchingStatus = false;
async function status() {
  if (!enabled.value || fetchingStatus) return;
  fetchingStatus = true;
  try {
    const res = await fetch(`/api/${props.provider}/status`, { signal: controller.signal });
    if (res.ok) lastSync.value = (await res.json()).last_sync_time;
  } catch {
    /* Polling resumes on the next interval. */
  } finally {
    fetchingStatus = false;
  }
}
async function syncNow() {
  if (syncing.value) return;
  syncing.value = true;
  try {
    // Persist this integration's current values before starting sync, even if autosave is pending.
    const body = Object.fromEntries(
      (['enabled', 'server_url', 'username', 'api_password'] as Field[]).map((field) => [
        key(field),
        String(props.settings[key(field)]),
      ])
    );
    const saved = await fetch('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: controller.signal,
    });
    if (!saved.ok) throw new Error();
    await store.startFreshRSSStatusPolling();
    const res = await fetch(`/api/${props.provider}/sync`, {
      method: 'POST',
      signal: controller.signal,
    });
    if (!res.ok) throw new Error();
    window.showToast(t('setting.freshrss.syncStarted'), 'success');
  } catch {
    if (!controller.signal.aborted) window.showToast(t('setting.freshrss.syncFailed'), 'error');
  } finally {
    syncing.value = false;
  }
}
onMounted(() => {
  void status();
  poll = setInterval(status, 5000);
});
onBeforeUnmount(() => {
  controller.abort();
  clearInterval(poll);
});
</script>

<template>
  <div>
    <div
      class="flex items-center justify-between gap-3 rounded-lg border border-border bg-bg-secondary p-2 sm:items-start sm:gap-4 sm:p-3"
    >
      <div class="flex min-w-0 flex-1 items-start gap-2 sm:gap-3">
        <ReaderProviderIcon :provider="provider" class="mt-0.5 h-5 w-5 shrink-0 sm:h-6 sm:w-6" />
        <div>
          <div class="text-sm font-medium sm:mb-1 sm:text-base">
            {{ t('setting.freshrss.integration', { name }) }}
          </div>
          <div class="text-xs text-text-secondary">
            {{ t('setting.freshrss.integrationDesc', { name }) }}
          </div>
        </div>
      </div>
      <input
        type="checkbox"
        class="toggle"
        :checked="enabled"
        :aria-label="t('setting.freshrss.integration', { name })"
        @change="toggle"
      />
    </div>
    <NestedSettingsContainer v-if="enabled">
      <SubSettingItem
        :icon="PhLink"
        :title="t('setting.freshrss.serverUrl')"
        :description="t('setting.freshrss.serverUrlDesc')"
        required
      >
        <InputControl
          type="url"
          :model-value="settings[`${provider}_server_url`]"
          :placeholder="`https://${provider}.example.com`"
          width="md"
          @update:model-value="update('server_url', $event)"
        />
      </SubSettingItem>
      <SubSettingItem
        :icon="PhUser"
        :title="t('setting.freshrss.username')"
        :description="t('setting.freshrss.usernameDesc')"
        required
      >
        <InputControl
          :model-value="settings[`${provider}_username`]"
          :placeholder="t('setting.freshrss.usernamePlaceholder')"
          width="md"
          @update:model-value="update('username', $event)"
        />
      </SubSettingItem>
      <SubSettingItem
        :icon="PhKey"
        :title="t('setting.freshrss.apiPassword')"
        :description="t('setting.freshrss.apiPasswordDesc')"
      >
        <InputControl
          type="password"
          :model-value="settings[`${provider}_api_password`]"
          :placeholder="t('setting.freshrss.apiPasswordPlaceholder')"
          width="md"
          @update:model-value="update('api_password', $event)"
        />
      </SubSettingItem>
      <SubSettingItem :icon="PhCloudCheck" :title="t('setting.freshrss.syncNow')">
        <template #extraInfo>
          {{ t('setting.freshrss.syncNowDesc') }}
          <div class="mt-1 text-xs text-text-secondary">
            {{ t('setting.freshrss.lastSync') }}:
            {{ lastSync ? new Date(lastSync).toLocaleString() : t('setting.freshrss.never') }}
          </div>
        </template>
        <button class="btn-secondary" :disabled="syncing" @click="syncNow">
          <PhArrowClockwise :size="16" :class="{ 'animate-spin': syncing }" />
          {{ t(syncing ? 'setting.freshrss.syncing' : 'setting.freshrss.sync') }}
        </button>
      </SubSettingItem>
    </NestedSettingsContainer>
  </div>
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
.btn-secondary {
  @apply bg-bg-tertiary border border-border text-text-primary px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium hover:bg-bg-secondary transition-colors disabled:opacity-50 disabled:cursor-wait;
}
</style>
