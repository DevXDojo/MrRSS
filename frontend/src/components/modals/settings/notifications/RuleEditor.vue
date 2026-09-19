<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhPlus, PhTrash, PhEye } from '@phosphor-icons/vue';
import { ToggleControl } from '@/components/settings';
import type {
  NotificationRule,
  NotificationChannel,
  NotificationCondition,
  NotificationPreview,
} from '@/types/notification';

const props = defineProps<{
  rule: NotificationRule;
  channels: NotificationChannel[];
  saving: boolean;
  preview: (rule: NotificationRule) => Promise<NotificationPreview>;
  errorText: (code: string) => string;
}>();
const emit = defineEmits<{ save: [rule: NotificationRule]; cancel: [] }>();
const { t } = useI18n();
const draft = ref<NotificationRule>(JSON.parse(JSON.stringify(props.rule)) as NotificationRule);
const result = ref<NotificationPreview | null>(null);
const previewing = ref(false);
const previewError = ref('');
let revision = 0;
watch(
  draft,
  () => {
    revision++;
    result.value = null;
    previewError.value = '';
  },
  { deep: true, flush: 'sync' }
);
const fields: NotificationCondition['field'][] = [
  'title',
  'summary',
  'feed',
  'category',
  'author',
  'url',
  'favorite',
  'read_later',
];
const operators: NotificationCondition['operator'][] = [
  'contains',
  'not_contains',
  'equals',
  'not_equals',
];
function isBoolean(c: NotificationCondition) {
  return c.field === 'favorite' || c.field === 'read_later';
}
function changeField(c: NotificationCondition) {
  c.operator = isBoolean(c) ? 'equals' : 'contains';
  c.value = isBoolean(c) ? 'true' : '';
}
function toggleDay(day: number) {
  draft.value.weekdays = draft.value.weekdays.includes(day)
    ? draft.value.weekdays.filter((d) => d !== day)
    : [...draft.value.weekdays, day];
}
async function showPreview() {
  const version = revision;
  previewing.value = true;
  previewError.value = '';
  try {
    const p = await props.preview(draft.value);
    if (version === revision) result.value = p;
  } catch (e) {
    if (version === revision)
      previewError.value = props.errorText(e instanceof Error ? e.message : 'internal');
  } finally {
    previewing.value = false;
  }
}
</script>

<template>
  <form class="push-editor space-y-5" @submit.prevent="emit('save', draft)">
    <fieldset :disabled="saving" class="space-y-5 min-w-0">
      <div class="push-grid">
        <label class="push-field"
          ><span>{{ t('setting.notifications.ruleName') }}</span
          ><input v-model.trim="draft.name" required maxlength="120" class="push-input"
        /></label>
        <label class="flex items-center gap-3 self-end min-h-10"
          ><ToggleControl
            v-model="draft.enabled"
            :aria-label="t('setting.notifications.ruleEnabled')"
          /><span class="text-sm">{{ t('setting.notifications.ruleEnabled') }}</span></label
        >
      </div>
      <div>
        <p class="push-label">{{ t('setting.notifications.sendTo') }}</p>
        <div class="flex flex-wrap gap-2">
          <label
            v-for="channel in channels"
            :key="channel.id"
            class="push-choice"
            :class="{ 'push-selected': draft.channel_ids.includes(channel.id) }"
          >
            <input v-model="draft.channel_ids" type="checkbox" :value="channel.id" />
            {{ channel.name
            }}<span v-if="!channel.enabled" class="text-text-secondary"
              >({{ t('setting.notifications.paused') }})</span
            >
          </label>
        </div>
      </div>
      <div class="push-grid">
        <label class="push-field"
          ><span>{{ t('setting.notifications.deliveryMode') }}</span
          ><select v-model="draft.mode" class="push-input">
            <option value="alert">{{ t('setting.notifications.alert') }}</option>
            <option value="digest">{{ t('setting.notifications.digest') }}</option>
          </select></label
        >
        <label v-if="draft.mode === 'alert'" class="push-field"
          ><span>{{ t('setting.notifications.interval') }}</span
          ><input
            v-model.number="draft.interval_minutes"
            type="number"
            min="1"
            max="1440"
            required
            class="push-input"
        /></label>
        <label v-else class="push-field"
          ><span>{{ t('setting.notifications.deliveryTime') }}</span
          ><input v-model="draft.time" type="time" required class="push-input"
        /></label>
      </div>
      <div v-if="draft.mode === 'digest'">
        <p class="push-label">{{ t('setting.notifications.weekdays') }}</p>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="day in [1, 2, 3, 4, 5, 6, 0]"
            :key="day"
            type="button"
            class="push-choice"
            :class="{ 'push-selected': draft.weekdays.includes(day) }"
            :aria-pressed="draft.weekdays.includes(day)"
            @click="toggleDay(day)"
          >
            {{ t(`setting.notifications.days.${day}`) }}
          </button>
        </div>
        <p class="push-help mt-2">{{ t('setting.notifications.everyDayHint') }}</p>
      </div>
      <div class="space-y-3">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <p class="push-label mb-0">{{ t('setting.notifications.conditions') }}</p>
          <select
            v-model="draft.match"
            class="push-input w-auto!"
            :aria-label="t('setting.notifications.matchLogic')"
          >
            <option value="all">{{ t('setting.notifications.matchAll') }}</option>
            <option value="any">{{ t('setting.notifications.matchAny') }}</option>
          </select>
        </div>
        <p v-if="!draft.conditions.length" class="push-help">
          {{ t('setting.notifications.noConditions') }}
        </p>
        <div v-for="(condition, index) in draft.conditions" :key="index" class="push-condition">
          <select
            v-model="condition.field"
            class="push-input"
            :aria-label="t('setting.notifications.field')"
            @change="changeField(condition)"
          >
            <option v-for="field in fields" :key="field" :value="field">
              {{ t(`setting.notifications.fields.${field}`) }}
            </option>
          </select>
          <select
            v-model="condition.operator"
            class="push-input"
            :aria-label="t('setting.notifications.operator')"
          >
            <option
              v-for="operator in isBoolean(condition) ? ['equals', 'not_equals'] : operators"
              :key="operator"
              :value="operator"
            >
              {{ t(`setting.notifications.operators.${operator}`) }}
            </option>
          </select>
          <select
            v-if="isBoolean(condition)"
            v-model="condition.value"
            class="push-input"
            :aria-label="t('setting.notifications.value')"
          >
            <option value="true">{{ t('setting.notifications.yes') }}</option>
            <option value="false">{{ t('setting.notifications.no') }}</option>
          </select>
          <input
            v-else
            v-model="condition.value"
            required
            maxlength="500"
            class="push-input"
            :aria-label="t('setting.notifications.value')"
            :placeholder="t('setting.notifications.keywordPlaceholder')"
          />
          <button
            type="button"
            class="push-icon-button"
            :aria-label="t('setting.notifications.removeCondition')"
            @click="draft.conditions.splice(index, 1)"
          >
            <PhTrash :size="17" />
          </button>
        </div>
        <button
          type="button"
          class="push-button"
          :disabled="draft.conditions.length >= 10"
          @click="draft.conditions.push({ field: 'title', operator: 'contains', value: '' })"
        >
          <PhPlus :size="16" />{{ t('setting.notifications.addCondition') }}
        </button>
      </div>
      <div class="push-grid">
        <label class="push-field"
          ><span>{{ t('setting.notifications.maxItems') }}</span
          ><input
            v-model.number="draft.max_items"
            type="number"
            min="1"
            max="20"
            required
            class="push-input"
        /></label>
        <label class="flex items-center gap-2 self-end min-h-10"
          ><input v-model="draft.unread_only" type="checkbox" /><span class="text-sm">{{
            t('setting.notifications.unreadOnly')
          }}</span></label
        >
      </div>
      <div class="flex flex-wrap gap-5 text-sm">
        <label class="flex items-center gap-2"
          ><input v-model="draft.include_summary" type="checkbox" />{{
            t('setting.notifications.includeSummary')
          }}</label
        ><label class="flex items-center gap-2"
          ><input v-model="draft.include_link" type="checkbox" />{{
            t('setting.notifications.includeLink')
          }}</label
        >
      </div>
      <p class="push-help">{{ t('setting.notifications.ruleHint') }}</p>
    </fieldset>
    <div class="flex flex-wrap gap-2 items-center">
      <button
        type="button"
        class="push-button"
        :disabled="previewing || saving"
        @click="showPreview"
      >
        <PhEye :size="16" />{{
          t(previewing ? 'setting.notifications.previewing' : 'setting.notifications.preview')
        }}
      </button>
      <div class="flex-1" />
      <button type="button" class="push-button" :disabled="saving" @click="emit('cancel')">
        {{ t('common.action.cancel') }}</button
      ><button
        type="submit"
        class="push-button push-primary"
        :disabled="saving || !draft.channel_ids.length"
      >
        {{ t(saving ? 'setting.notifications.saving' : 'setting.notifications.saveRule') }}
      </button>
    </div>
    <p v-if="previewError" role="alert" class="push-error">{{ previewError }}</p>
    <div v-if="result" role="status" class="push-preview space-y-3">
      <p class="text-sm font-medium">
        {{
          t('setting.notifications.previewResult', {
            matched: result.matched,
            scanned: result.scanned,
          })
        }}
      </p>
      <p class="push-help">{{ t('setting.notifications.previewHint') }}</p>
      <p v-if="!result.messages.length" class="text-sm text-text-secondary">
        {{ t('setting.notifications.noMatches') }}
      </p>
      <pre
        v-for="(message, index) in result.messages"
        :key="index"
        class="whitespace-pre-wrap break-words text-sm font-sans select-text"
        >{{ message }}</pre>
    </div>
  </form>
</template>
