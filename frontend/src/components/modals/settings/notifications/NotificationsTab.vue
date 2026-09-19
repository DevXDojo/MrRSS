<script setup lang="ts">
import { computed, ref, watch, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhBellRinging,
  PhPlus,
  PhPencilSimple,
  PhTrash,
  PhClock,
  PhFunnel,
  PhArrowClockwise,
  PhPaperPlaneTilt,
  PhCaretDown,
  PhCaretRight,
} from '@phosphor-icons/vue';
import { SettingGroup, ToggleControl } from '@/components/settings';
import { useNotifications } from '@/composables/notification/useNotifications';
import { newChannel, newRule } from '@/types/notification';
import type {
  NotificationChannel,
  NotificationRule,
  NotificationProvider,
  NotificationConfig,
} from '@/types/notification';
import ChannelEditor from './ChannelEditor.vue';
import RuleEditor from './RuleEditor.vue';
import telegramLogo from '@/assets/brands/telegram.svg';
import discordLogo from '@/assets/brands/discord.svg';
import feishuLogo from '@/assets/brands/feishu.svg';
import './notifications.css';

const { t, locale } = useI18n();
const emit = defineEmits<{ editing: [value: boolean]; busy: [value: boolean] }>();
const {
  config,
  history,
  loading,
  saving,
  error,
  historyError,
  load,
  save,
  test,
  preview,
  refreshHistory,
  errorText,
} = useNotifications();
const channelDraft = ref<NotificationChannel | null>(null);
const ruleDraft = ref<NotificationRule | null>(null);
const expanded = ref<NotificationProvider[]>([]);
const historyOpen = ref(false);
const providers = [
  { id: 'telegram' as const, name: 'Telegram', logo: telegramLogo },
  { id: 'discord' as const, name: 'Discord', logo: discordLogo },
  { id: 'feishu' as const, name: 'Feishu / Lark', logo: feishuLogo },
];
const browserZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
const zones = computed(() => [
  ...new Set([
    'Local',
    browserZone,
    config.value?.timezone || 'Local',
    'Asia/Shanghai',
    'Asia/Tokyo',
    'Asia/Singapore',
    'UTC',
    'Europe/London',
    'Europe/Berlin',
    'America/New_York',
    'America/Los_Angeles',
    'Australia/Sydney',
  ]),
]);
const activeChannels = computed(() => config.value?.channels.filter((c) => c.enabled).length || 0);
const activeRules = computed(() => config.value?.rules.filter((r) => r.enabled).length || 0);
const editing = computed(() => !!channelDraft.value || !!ruleDraft.value);
watch(editing, (value) => emit('editing', value));
watch(saving, (value) => emit('busy', value));
onBeforeUnmount(() => {
  emit('editing', false);
  emit('busy', false);
});

function channelsFor(provider: NotificationProvider) {
  return config.value?.channels.filter((c) => c.provider === provider) || [];
}
function toggleProvider(provider: NotificationProvider) {
  expanded.value = expanded.value.includes(provider)
    ? expanded.value.filter((p) => p !== provider)
    : [...expanded.value, provider];
}
function addChannel(provider: NotificationProvider) {
  if (editing.value) return;
  const name =
    provider === 'feishu'
      ? t('setting.notifications.feishu')
      : providers.find((p) => p.id === provider)!.name;
  channelDraft.value = newChannel(provider, name);
  if (!expanded.value.includes(provider)) expanded.value.push(provider);
}
async function saveChannel(channel: NotificationChannel) {
  if (!config.value) return;
  const channels = config.value.channels.filter((c) => c.id !== channel.id);
  channels.push(channel);
  if (await save({ ...config.value, channels })) channelDraft.value = null;
}
async function removeChannel(channel: NotificationChannel) {
  if (!config.value) return;
  if (config.value.rules.some((r) => r.channel_ids.includes(channel.id))) {
    window.showToast(t('setting.notifications.channelInUse'), 'error');
    return;
  }
  if (
    !(await window.showConfirm({
      title: t('setting.notifications.removeChannel'),
      message: t('setting.notifications.removeChannelConfirm', { name: channel.name }),
      isDanger: true,
    }))
  )
    return;
  await save({
    ...config.value,
    channels: config.value.channels.filter((c) => c.id !== channel.id),
  });
}
function addRule(mode: 'alert' | 'digest') {
  if (!config.value || editing.value) return;
  ruleDraft.value = newRule(
    mode,
    t(`setting.notifications.${mode === 'alert' ? 'newAlert' : 'newDigest'}`),
    config.value.channels.filter((c) => c.enabled).map((c) => c.id)
  );
}
async function saveRule(rule: NotificationRule) {
  if (!config.value) return;
  const rules = config.value.rules.filter((r) => r.id !== rule.id);
  rules.push(rule);
  if (await save({ ...config.value, rules })) ruleDraft.value = null;
}
async function removeRule(rule: NotificationRule) {
  if (!config.value) return;
  if (
    !(await window.showConfirm({
      title: t('setting.notifications.removeRule'),
      message: t('setting.notifications.removeRuleConfirm', { name: rule.name }),
      isDanger: true,
    }))
  )
    return;
  await save({ ...config.value, rules: config.value.rules.filter((r) => r.id !== rule.id) });
}
async function common(patch: Partial<NotificationConfig>) {
  if (config.value) await save({ ...config.value, ...patch });
}
function value(event: Event) {
  return (event.target as HTMLInputElement).value;
}
function channelNames(ids: string[]) {
  return ids
    .map((id) => config.value?.channels.find((c) => c.id === id)?.name)
    .filter(Boolean)
    .join(' · ');
}
function date(seconds: number) {
  return new Date(seconds * 1000).toLocaleString(locale.value);
}
</script>

<template>
  <div class="push-tab space-y-6">
    <p v-if="loading" role="status" class="text-sm text-text-secondary py-8 text-center">
      {{ t('setting.notifications.loading') }}
    </p>
    <div
      v-if="error"
      role="alert"
      class="push-error flex flex-wrap items-center justify-between gap-2"
    >
      <span>{{ error }}</span
      ><button v-if="!editing" class="push-button" :disabled="loading || saving" @click="load">
        {{ t('setting.notifications.reload') }}
      </button>
    </div>
    <template v-if="config && !loading">
      <section class="push-hero">
        <div class="flex items-start gap-3">
          <div class="push-hero-icon"><PhBellRinging :size="25" weight="duotone" /></div>
          <div class="flex-1 min-w-0">
            <h2 class="text-lg font-semibold">{{ t('setting.notifications.title') }}</h2>
            <p class="text-sm text-text-secondary mt-1 leading-relaxed">
              {{ t('setting.notifications.description') }}
            </p>
          </div>
          <ToggleControl
            :model-value="config.enabled"
            :disabled="saving || editing"
            :aria-label="t('setting.notifications.enable')"
            @update:model-value="common({ enabled: $event })"
          />
        </div>
        <div class="flex flex-wrap items-center gap-2 mt-4">
          <span class="push-badge" :class="{ 'push-selected': config.enabled }">{{
            t(config.enabled ? 'setting.notifications.running' : 'setting.notifications.off')
          }}</span
          ><span class="text-xs text-text-secondary">{{
            t('setting.notifications.overview', { channels: activeChannels, rules: activeRules })
          }}</span>
        </div>
        <p class="push-help mt-3">{{ t('setting.notifications.runtimeHint') }}</p>
      </section>

      <SettingGroup
        :icon="PhClock"
        :title="t('setting.notifications.commonSettings')"
        :description="t('setting.notifications.commonHint')"
      >
        <fieldset :disabled="saving || editing" class="push-panel space-y-4 min-w-0">
          <label class="push-field"
            ><span>{{ t('setting.notifications.timezone') }}</span
            ><select
              :value="config.timezone"
              class="push-input"
              @change="common({ timezone: value($event) })"
            >
              <option v-for="zone in zones" :key="zone" :value="zone">
                {{ zone === 'Local' ? t('setting.notifications.systemTimezone') : zone }}
              </option>
            </select></label
          >
          <label class="flex items-center justify-between gap-3"
            ><span class="text-sm font-medium">{{ t('setting.notifications.quietHours') }}</span
            ><ToggleControl
              :model-value="config.quiet_enabled"
              :aria-label="t('setting.notifications.quietHours')"
              @update:model-value="common({ quiet_enabled: $event })"
          /></label>
          <div v-if="config.quiet_enabled" class="push-grid">
            <label class="push-field"
              ><span>{{ t('setting.notifications.quietStart') }}</span
              ><input
                :value="config.quiet_start"
                type="time"
                class="push-input"
                @change="common({ quiet_start: value($event) })" /></label
            ><label class="push-field"
              ><span>{{ t('setting.notifications.quietEnd') }}</span
              ><input
                :value="config.quiet_end"
                type="time"
                class="push-input"
                @change="common({ quiet_end: value($event) })"
            /></label>
          </div>
          <p v-if="config.quiet_enabled" class="push-help">
            {{ t('setting.notifications.quietHint') }}
          </p>
        </fieldset>
      </SettingGroup>

      <SettingGroup
        :icon="PhPaperPlaneTilt"
        :title="t('setting.notifications.channels')"
        :description="t('setting.notifications.channelsHint')"
      >
        <section v-for="provider in providers" :key="provider.id" class="push-provider">
          <button
            type="button"
            class="push-provider-heading"
            :aria-expanded="expanded.includes(provider.id)"
            @click="toggleProvider(provider.id)"
          >
            <img :src="provider.logo" :alt="provider.name" class="push-logo" />
            <span class="flex-1 text-left min-w-0"
              ><span class="block text-sm font-semibold">{{
                provider.id === 'feishu' ? t('setting.notifications.feishu') : provider.name
              }}</span
              ><span class="block text-xs text-text-secondary mt-1">{{
                t(`setting.notifications.providerHint.${provider.id}`)
              }}</span></span
            >
            <span class="push-badge">{{
              channelsFor(provider.id).length
                ? t('setting.notifications.configured', { count: channelsFor(provider.id).length })
                : t('setting.notifications.notConfigured')
            }}</span>
            <component
              :is="expanded.includes(provider.id) ? PhCaretDown : PhCaretRight"
              :size="17"
              class="shrink-0"
            />
          </button>
          <div v-if="expanded.includes(provider.id)" class="px-4 pb-4 space-y-3">
            <div v-for="channel in channelsFor(provider.id)" :key="channel.id" class="push-row">
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium break-words">{{ channel.name }}</p>
                <span class="text-xs text-text-secondary">{{
                  t(
                    channel.enabled ? 'setting.notifications.ready' : 'setting.notifications.paused'
                  )
                }}</span>
              </div>
              <button
                class="push-icon-button"
                :disabled="saving || editing"
                :aria-label="t('setting.notifications.editChannel', { name: channel.name })"
                @click="channelDraft = { ...channel }"
              >
                <PhPencilSimple :size="18" /></button
              ><button
                class="push-icon-button"
                :disabled="saving || editing"
                :aria-label="t('setting.notifications.removeChannel')"
                @click="removeChannel(channel)"
              >
                <PhTrash :size="18" />
              </button>
            </div>
            <ChannelEditor
              v-if="channelDraft?.provider === provider.id"
              :key="channelDraft.id"
              :channel="channelDraft"
              :saving="saving"
              :test="test"
              :error-text="errorText"
              @save="saveChannel"
              @cancel="channelDraft = null"
            />
            <button
              v-else
              class="push-button"
              :disabled="saving || editing || config.channels.length >= 20"
              @click="addChannel(provider.id)"
            >
              <PhPlus :size="16" />{{ t('setting.notifications.addChannel') }}
            </button>
          </div>
        </section>
      </SettingGroup>

      <SettingGroup
        :icon="PhFunnel"
        :title="t('setting.notifications.rules')"
        :description="t('setting.notifications.rulesHint')"
      >
        <div v-if="!config.rules.length && !ruleDraft" class="push-empty">
          <PhBellRinging :size="30" class="text-accent mx-auto mb-3" weight="duotone" />
          <p class="text-sm font-medium">{{ t('setting.notifications.emptyRules') }}</p>
          <p class="push-help mt-2">
            {{
              t(
                config.channels.length
                  ? 'setting.notifications.emptyRulesHint'
                  : 'setting.notifications.connectFirst'
              )
            }}
          </p>
        </div>
        <div v-for="rule in config.rules" :key="rule.id" class="push-row push-panel">
          <component
            :is="rule.mode === 'digest' ? PhClock : PhBellRinging"
            :size="21"
            class="text-accent shrink-0"
          />
          <div class="flex-1 min-w-0">
            <p class="text-sm font-semibold break-words">{{ rule.name }}</p>
            <p class="push-help mt-1">
              {{
                rule.mode === 'digest'
                  ? t('setting.notifications.dailyAt', { time: rule.time })
                  : t('setting.notifications.everyMinutes', { minutes: rule.interval_minutes })
              }}
              ·
              {{ t(rule.enabled ? 'setting.notifications.ready' : 'setting.notifications.paused') }}
            </p>
            <p class="push-help break-words">{{ channelNames(rule.channel_ids) }}</p>
          </div>
          <button
            class="push-icon-button"
            :disabled="saving || editing"
            :aria-label="t('setting.notifications.editRule', { name: rule.name })"
            @click="ruleDraft = rule"
          >
            <PhPencilSimple :size="18" /></button
          ><button
            class="push-icon-button"
            :disabled="saving || editing"
            :aria-label="t('setting.notifications.removeRule')"
            @click="removeRule(rule)"
          >
            <PhTrash :size="18" />
          </button>
        </div>
        <RuleEditor
          v-if="ruleDraft"
          :key="ruleDraft.id"
          :rule="ruleDraft"
          :channels="config.channels"
          :saving="saving"
          :preview="preview"
          :error-text="errorText"
          @save="saveRule"
          @cancel="ruleDraft = null"
        />
        <div v-else class="flex flex-wrap gap-2">
          <button
            class="push-button"
            :disabled="saving || editing || !config.channels.length || config.rules.length >= 50"
            @click="addRule('alert')"
          >
            <PhPlus :size="16" />{{ t('setting.notifications.addAlert') }}</button
          ><button
            class="push-button"
            :disabled="saving || editing || !config.channels.length || config.rules.length >= 50"
            @click="addRule('digest')"
          >
            <PhClock :size="16" />{{ t('setting.notifications.addDigest') }}
          </button>
        </div>
      </SettingGroup>

      <section class="push-panel">
        <div class="flex items-center justify-between gap-2">
          <button
            class="flex items-center gap-2 text-sm font-semibold"
            :aria-expanded="historyOpen"
            @click="historyOpen = !historyOpen"
          >
            <component :is="historyOpen ? PhCaretDown : PhCaretRight" :size="17" />{{
              t('setting.notifications.history')
            }}<span class="push-badge">{{ history.length }}</span></button
          ><button
            class="push-icon-button"
            :aria-label="t('setting.notifications.refreshHistory')"
            @click="refreshHistory"
          >
            <PhArrowClockwise :size="18" />
          </button>
        </div>
        <div v-if="historyOpen" class="space-y-3 mt-4">
          <p class="push-help">{{ t('setting.notifications.historyHint') }}</p>
          <p v-if="historyError" role="alert" class="push-error">
            {{ t('setting.notifications.historyError') }}
          </p>
          <p v-else-if="!history.length" class="push-help py-3">
            {{ t('setting.notifications.historyEmpty') }}
          </p>
          <div
            v-for="item in history"
            :key="item.id"
            class="border-t border-border pt-3 text-xs space-y-1"
          >
            <div class="flex flex-wrap justify-between gap-2">
              <span class="font-medium"
                >{{ item.label }} ·
                {{
                  channelNames([item.channel_id]) || t('setting.notifications.deletedChannel')
                }}</span
              ><span class="push-badge" :class="{ 'push-selected': item.status === 'sent' }">{{
                t(`setting.notifications.status.${item.status}`)
              }}</span>
            </div>
            <p class="text-text-secondary">
              {{ date(item.created_at) }} ·
              {{ t('setting.notifications.attempts', { count: item.attempts }) }}
            </p>
            <p v-if="item.error_code" class="text-text-secondary">
              {{ errorText(item.error_code)
              }}<template v-if="item.status === 'pending'">
                ·
                {{
                  t('setting.notifications.retryAt', { time: date(item.next_attempt) })
                }}</template
              >
            </p>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>
