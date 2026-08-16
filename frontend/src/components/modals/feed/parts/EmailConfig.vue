<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhGlobe, PhCheckCircle, PhXCircle, PhSpinner } from '@phosphor-icons/vue';

interface Props {
  mode: 'add' | 'edit';
  emailAddress?: string;
  imapServer?: string;
  imapPort?: number;
  username?: string;
  password?: string;
  folder?: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:emailAddress': [value: string];
  'update:imapServer': [value: string];
  'update:imapPort': [value: number];
  'update:username': [value: string];
  'update:password': [value: string];
  'update:folder': [value: string];
}>();

const { t } = useI18n();

// Test connection state
const isTesting = ref(false);
const testResult = ref<'success' | 'error' | null>(null);
const testMessage = ref('');

// Common IMAP providers
const providers = [
  { name: 'Gmail', server: 'imap.gmail.com', port: 993 },
  { name: 'Outlook', server: 'outlook.office365.com', port: 993 },
  { name: 'QQ', server: 'imap.qq.com', port: 993 },
  { name: '163', server: 'imap.163.com', port: 993 },
  { name: 'iCloud', server: 'imap.mail.me.com', port: 993 },
  { name: 'Yahoo', server: 'imap.mail.yahoo.com', port: 993 },
];

// Use computed with v-model pattern for two-way binding
const emailAddress = computed({
  get: () => props.emailAddress || '',
  set: (val) => emit('update:emailAddress', val),
});

const imapServer = computed({
  get: () => props.imapServer || '',
  set: (val) => emit('update:imapServer', val),
});

const imapPort = computed({
  get: () => props.imapPort ?? 993,
  set: (val: string | number | null) => {
    const numVal = val === '' || val === null ? 993 : Number(val);
    emit('update:imapPort', numVal);
  },
});

const username = computed({
  get: () => props.username || '',
  set: (val) => emit('update:username', val),
});

const password = computed({
  get: () => props.password || '',
  set: (val) => emit('update:password', val),
});

const folder = computed({
  get: () => props.folder || 'INBOX',
  set: (val) => emit('update:folder', val),
});

// Select provider
function selectProvider(provider: (typeof providers)[0]) {
  emit('update:imapServer', provider.server);
  emit('update:imapPort', provider.port);
}

// Test IMAP connection
async function testConnection() {
  if (!imapServer.value || !username.value || !password.value) {
    testMessage.value = t('modal.feed.emailFillRequired');
    testResult.value = 'error';
    return;
  }

  isTesting.value = true;
  testResult.value = null;
  testMessage.value = '';

  const requestBody = {
    email_imap_server: imapServer.value,
    email_imap_port: imapPort.value,
    email_username: username.value,
    email_password: password.value,
    email_folder: folder.value,
  };

  try {
    const response = await fetch('/api/feeds/test-imap', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(requestBody),
    });

    const data = await response.json();

    if (response.ok) {
      testResult.value = 'success';
      testMessage.value = data.message || t('common.connectionSuccessful');
    } else {
      testResult.value = 'error';
      testMessage.value = data.error || t('common.connectionFailed');
    }
  } catch {
    testResult.value = 'error';
    testMessage.value = t('modal.feed.emailConnectionError');
  } finally {
    isTesting.value = false;
  }
}

// Form validation
const isValid = computed(() => {
  return emailAddress.value && imapServer.value && username.value && password.value;
});

defineExpose({
  isValid,
});
</script>

<template>
  <div class="email-config space-y-4">
    <!-- Email Address -->
    <div class="mb-3">
      <label class="block mb-1 sm:mb-1.5 font-semibold text-xs sm:text-sm text-text-secondary">
        {{ t('modal.feed.emailAddress') }}
      </label>
      <input
        v-model="emailAddress"
        type="email"
        placeholder="newsletter@example.com"
        class="input-field w-full"
      />
      <div class="text-xs text-text-secondary mt-1">{{ t('modal.feed.emailAddressHint') }}</div>
    </div>

    <!-- IMAP Server -->
    <div class="mb-3">
      <label class="block mb-1 sm:mb-1.5 font-semibold text-xs sm:text-sm text-text-secondary">
        {{ t('modal.feed.emailServer') }} <span class="text-red-500">*</span>
      </label>

      <!-- Quick select providers -->
      <div class="mb-2 flex flex-wrap gap-1.5">
        <button
          v-for="provider in providers"
          :key="provider.name"
          type="button"
          class="text-xs px-2 py-1 rounded bg-bg-tertiary text-text-secondary hover:bg-accent hover:text-white transition-colors"
          @click="selectProvider(provider)"
        >
          {{ provider.name }}
        </button>
      </div>

      <div class="flex flex-wrap gap-2">
        <input
          v-model="imapServer"
          type="text"
          placeholder="imap.gmail.com"
          class="input-field flex-1 min-w-[120px]"
        />
        <input
          v-model="imapPort"
          type="number"
          placeholder="993"
          class="input-field w-20 sm:w-24"
        />
      </div>
    </div>

    <!-- Username -->
    <div class="mb-3">
      <label class="block mb-1 sm:mb-1.5 font-semibold text-xs sm:text-sm text-text-secondary">
        {{ t('modal.feed.emailUsername') }} <span class="text-red-500">*</span>
      </label>
      <input
        v-model="username"
        type="text"
        placeholder="your-email@example.com"
        class="input-field w-full"
      />
    </div>

    <!-- Password -->
    <div class="mb-3">
      <label class="block mb-1 sm:mb-1.5 font-semibold text-xs sm:text-sm text-text-secondary">
        {{ t('modal.feed.emailPassword') }} <span class="text-red-500">*</span>
      </label>
      <input
        v-model="password"
        type="password"
        :placeholder="t('modal.feed.emailPasswordPlaceholder')"
        class="input-field w-full"
      />
    </div>

    <!-- Folder (Optional) -->
    <div class="mb-3">
      <label class="block mb-1 sm:mb-1.5 font-semibold text-xs sm:text-sm text-text-secondary">
        {{ t('modal.feed.emailFolder') }}
      </label>
      <input v-model="folder" type="text" placeholder="INBOX" class="input-field w-full" />
    </div>

    <!-- Test Connection Button -->
    <div class="text-center">
      <button
        type="button"
        :disabled="isTesting || !imapServer || !username || !password"
        class="inline-flex items-center gap-2 text-sm px-4 py-2 rounded-lg border border-border bg-bg-tertiary hover:bg-bg-secondary disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        @click="testConnection"
      >
        <PhSpinner v-if="isTesting" :size="16" class="animate-spin" />
        <PhCheckCircle v-else-if="testResult === 'success'" :size="16" class="text-green-500" />
        <PhXCircle v-else-if="testResult === 'error'" :size="16" class="text-red-500" />
        <PhGlobe v-else :size="16" />
        <span>{{ t('modal.feed.emailTestConnection') }}</span>
      </button>

      <!-- Test result message -->
      <div
        v-if="testMessage"
        :class="[
          'mt-2 text-xs p-2 rounded',
          testResult === 'success'
            ? 'bg-green-500/10 text-green-500'
            : 'bg-red-500/10 text-red-500',
        ]"
      >
        {{ testMessage }}
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";
.input-field {
  @apply p-2.5 border border-border rounded-md bg-bg-tertiary text-text-primary text-sm focus:border-accent focus:outline-none transition-colors;
}
</style>
