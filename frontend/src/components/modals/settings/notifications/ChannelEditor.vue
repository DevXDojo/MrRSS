<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhPaperPlaneTilt, PhCheckCircle, PhEye, PhEyeSlash } from '@phosphor-icons/vue';
import { ToggleControl } from '@/components/settings';
import type { NotificationChannel } from '@/types/notification';

const props = defineProps<{
  channel: NotificationChannel;
  saving: boolean;
  test: (channel: NotificationChannel) => Promise<void>;
  errorText: (code: string) => string;
}>();
const emit = defineEmits<{ save: [channel: NotificationChannel]; cancel: [] }>();
const { t } = useI18n();
const draft = ref<NotificationChannel>({ ...props.channel });
const reveal = ref(false);
const testing = ref(false);
const testStatus = ref('');
const testError = ref('');
watch(
  draft,
  () => {
    testStatus.value = '';
    testError.value = '';
  },
  { deep: true }
);
async function sendTest() {
  testing.value = true;
  testStatus.value = '';
  testError.value = '';
  try {
    await props.test(draft.value);
    testStatus.value = t('setting.notifications.testSuccess');
  } catch (e) {
    testError.value = props.errorText(e instanceof Error ? e.message : 'internal');
  } finally {
    testing.value = false;
  }
}
</script>

<template>
  <form class="push-editor space-y-4" @submit.prevent="emit('save', draft)">
    <fieldset :disabled="saving || testing" class="space-y-4 min-w-0">
      <div class="push-grid">
        <label class="push-field"
          ><span>{{ t('setting.notifications.channelName') }}</span
          ><input v-model.trim="draft.name" required maxlength="120" class="push-input"
        /></label>
        <label class="flex items-center gap-3 self-end min-h-10"
          ><ToggleControl
            v-model="draft.enabled"
            :aria-label="t('setting.notifications.channelEnabled')"
          /><span class="text-sm">{{ t('setting.notifications.channelEnabled') }}</span></label
        >
      </div>
      <p class="push-help">{{ t(`setting.notifications.setup.${draft.provider}`) }}</p>
      <div class="flex items-center justify-between gap-3">
        <span class="text-xs text-text-secondary">{{
          t('setting.notifications.credentialsHint')
        }}</span>
        <button
          type="button"
          class="push-icon-button"
          :aria-label="
            t(reveal ? 'setting.notifications.hideSecrets' : 'setting.notifications.showSecrets')
          "
          @click="reveal = !reveal"
        >
          <component :is="reveal ? PhEyeSlash : PhEye" :size="18" />
        </button>
      </div>
      <template v-if="draft.provider === 'telegram'">
        <label class="push-field"
          ><span>{{ t('setting.notifications.token') }}</span
          ><input
            v-model.trim="draft.token"
            :type="reveal ? 'text' : 'password'"
            required
            autocomplete="new-password"
            spellcheck="false"
            placeholder="123456:ABC…"
            class="push-input"
        /></label>
        <div class="push-grid">
          <label class="push-field"
            ><span>{{ t('setting.notifications.chatId') }}</span
            ><input
              v-model.trim="draft.chat_id"
              required
              spellcheck="false"
              placeholder="-1001234567890"
              class="push-input"
          /></label>
          <label class="push-field"
            ><span>{{ t('setting.notifications.threadId') }}</span
            ><input
              v-model.number="draft.thread_id"
              type="number"
              min="0"
              step="1"
              class="push-input"
          /></label>
        </div>
      </template>
      <template v-else>
        <label class="push-field"
          ><span>{{ t('setting.notifications.webhook') }}</span
          ><input
            v-model.trim="draft.webhook"
            :type="reveal ? 'text' : 'password'"
            required
            autocomplete="new-password"
            spellcheck="false"
            :placeholder="
              draft.provider === 'discord'
                ? 'https://discord.com/api/webhooks/…'
                : 'https://open.feishu.cn/open-apis/bot/v2/hook/…'
            "
            class="push-input"
        /></label>
        <label v-if="draft.provider === 'feishu'" class="push-field"
          ><span>{{ t('setting.notifications.signingSecret') }}</span
          ><input
            v-model.trim="draft.secret"
            :type="reveal ? 'text' : 'password'"
            autocomplete="new-password"
            class="push-input"
            :placeholder="t('setting.notifications.optional')"
        /></label>
      </template>
      <p class="push-help">{{ t('setting.notifications.testHint') }}</p>
    </fieldset>
    <p v-if="testStatus" role="status" class="text-sm text-accent flex items-center gap-2">
      <PhCheckCircle :size="18" />{{ testStatus }}
    </p>
    <p v-if="testError" role="alert" class="push-error">{{ testError }}</p>
    <div class="flex flex-wrap items-center gap-2">
      <button type="button" class="push-button" :disabled="saving || testing" @click="sendTest">
        <PhPaperPlaneTilt :size="16" />{{
          t(testing ? 'setting.notifications.testing' : 'setting.notifications.test')
        }}
      </button>
      <div class="flex-1" />
      <button
        type="button"
        class="push-button"
        :disabled="saving || testing"
        @click="emit('cancel')"
      >
        {{ t('common.action.cancel') }}
      </button>
      <button type="submit" class="push-button push-primary" :disabled="saving || testing">
        {{ t(saving ? 'setting.notifications.saving' : 'setting.notifications.saveChannel') }}
      </button>
    </div>
  </form>
</template>
