<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhShieldCheck, PhSparkle } from '@phosphor-icons/vue';
import { SettingGroup, ToggleControl } from '@/components/settings';
import AIProfileSelector from '../ai/AIProfileSelector.vue';
import type { SettingsData } from '@/types/settings';
import type { AdFilterConfig, AdAnalysis } from '@/types/adFilter';
import { defaultAdFilterConfig } from '@/types/adFilter';
import { clearArticleContentCache } from '@/utils/articleContentCache';

const props = defineProps<{ settings: SettingsData }>();
const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
  editing: [value: boolean];
  busy: [value: boolean];
}>();
const { t, te } = useI18n();
const config = ref(defaultAdFilterConfig());
const baseline = ref('');
const loading = ref(true);
const saving = ref(false);
const testing = ref(false);
const error = ref('');
const sampleURL = ref('https://news.example.org/story');
const sampleHTML = ref(
  '<article><h1>Understanding database indexes</h1><p>Indexes help databases find the rows needed for a query. Measure real workloads and inspect execution plans before changing an index. A useful plan avoids reading unnecessary rows while preserving the ordering and filtering required by the application. Keep a baseline and repeat measurements with representative data.</p><p>Sponsored message: Buy Acme Cloud now and save 40% with code NEWS40.</p><ins class="adsbygoogle" data-ad-client="ca-pub-123" data-ad-slot="456">Advertising slot</ins></article>'
);
const result = ref('');
const analysis = ref<AdAnalysis | null>(null);
const controller = new AbortController();
let previewController: AbortController | null = null;
let previewGeneration = 0;
const dirty = computed(() => !!baseline.value && JSON.stringify(config.value) !== baseline.value);
watch(dirty, (value) => emit('editing', value));
watch(saving, (value) => emit('busy', value));
function errorText(code: string) {
  const [key, line] = code.split(':');
  return te(`setting.adFilter.errors.${key}`)
    ? t(`setting.adFilter.errors.${key}`, { line })
    : t('setting.adFilter.errors.save_failed');
}
async function request<T>(
  path: string,
  method = 'GET',
  body?: unknown,
  signal = controller.signal
): Promise<T> {
  const response = await fetch(`/api/ad-filter/${path}`, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
    signal,
  });
  const data = await response.json();
  if (!response.ok) throw new Error(data.error?.message || 'save_failed');
  return data as T;
}
async function load() {
  loading.value = true;
  error.value = '';
  try {
    config.value = await request<AdFilterConfig>('config');
    baseline.value = JSON.stringify(config.value);
  } catch (e) {
    if (!controller.signal.aborted)
      error.value = errorText(e instanceof Error ? e.message : 'config_unavailable');
  } finally {
    loading.value = false;
  }
}
async function save() {
  saving.value = true;
  error.value = '';
  try {
    config.value = await request<AdFilterConfig>('config', 'PUT', config.value);
    baseline.value = JSON.stringify(config.value);
    clearArticleContentCache();
    window.dispatchEvent(new CustomEvent('ad-filter-changed', { detail: config.value }));
    window.showToast(t('setting.adFilter.saved'), 'success');
  } catch (e) {
    if (!controller.signal.aborted)
      error.value = errorText(e instanceof Error ? e.message : 'save_failed');
  } finally {
    saving.value = false;
  }
}
function cancelDraft() {
  config.value = JSON.parse(baseline.value) as AdFilterConfig;
  error.value = '';
}
function invalidatePreview() {
  previewGeneration++;
  previewController?.abort();
  testing.value = false;
  result.value = '';
  analysis.value = null;
}
watch([config, sampleHTML, sampleURL], invalidatePreview, { deep: true, flush: 'sync' });
async function preview(ai: boolean) {
  invalidatePreview();
  const generation = previewGeneration;
  previewController = new AbortController();
  testing.value = true;
  error.value = '';
  const { revision: _revision, ...options } = config.value;
  try {
    const output = await request<AdAnalysis | { content: string }>(
      ai ? 'analyze' : 'preview',
      'POST',
      { options, url: sampleURL.value, html: sampleHTML.value },
      previewController.signal
    );
    if (generation !== previewGeneration) return;
    result.value = output.content;
    if ('decisions' in output) analysis.value = output;
  } catch (e) {
    if (generation === previewGeneration && !previewController.signal.aborted)
      error.value = errorText(e instanceof Error ? e.message : 'ai_unavailable');
  } finally {
    if (generation === previewGeneration) testing.value = false;
  }
}
onMounted(load);
onBeforeUnmount(() => {
  controller.abort();
  previewController?.abort();
  emit('editing', false);
  emit('busy', false);
});
</script>

<template>
  <SettingGroup :icon="PhShieldCheck" :title="t('setting.adFilter.title')">
    <div class="space-y-4 p-1 text-sm">
      <p class="text-text-secondary">{{ t('setting.adFilter.description') }}</p>
      <p v-if="loading" role="status">{{ t('setting.adFilter.loading') }}</p>
      <div
        v-if="error"
        role="alert"
        class="rounded-lg border border-red-500/30 bg-red-500/5 p-3 text-red-500 break-words"
      >
        {{ error
        }}<button v-if="!dirty" type="button" class="block mt-2 text-accent" @click="load">
          {{ t('setting.adFilter.reload') }}
        </button>
      </div>
      <fieldset
        v-if="!loading && baseline"
        :disabled="saving"
        class="space-y-4 disabled:opacity-70"
      >
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="font-medium">{{ t('setting.adFilter.basic') }}</p>
            <p class="text-xs text-text-secondary mt-1">{{ t('setting.adFilter.basicHint') }}</p>
          </div>
          <ToggleControl v-model="config.enabled" :aria-label="t('setting.adFilter.basic')" />
        </div>
        <div class="rounded-xl border border-border bg-bg-secondary p-3 space-y-3">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="font-medium flex items-center gap-2">
                <PhSparkle :size="17" />{{ t('setting.adFilter.ai') }}
              </p>
              <p class="text-xs text-text-secondary mt-1">{{ t('setting.adFilter.aiHint') }}</p>
            </div>
            <ToggleControl
              v-model="config.ai_enabled"
              :disabled="!config.enabled"
              :aria-label="t('setting.adFilter.ai')"
            />
          </div>
          <template v-if="config.ai_enabled">
            <div class="space-y-2">
              <p class="font-medium text-xs">{{ t('setting.adFilter.profile') }}</p>
              <AIProfileSelector
                :model-value="settings.ai_ad_filter_profile_id"
                allow-default
                @update:model-value="
                  emit('update:settings', {
                    ...props.settings,
                    ai_ad_filter_profile_id: $event || '0',
                  })
                "
              />
              <p class="text-xs text-text-secondary">{{ t('setting.adFilter.profileHint') }}</p>
            </div>
            <label class="block space-y-1"
              ><span class="text-xs font-medium">{{ t('setting.adFilter.mode') }}</span
              ><select v-model="config.ai_mode" class="filter-input">
                <option value="review">{{ t('setting.adFilter.review') }}</option>
                <option value="automatic">{{ t('setting.adFilter.automatic') }}</option>
              </select></label
            >
            <p v-if="config.ai_mode === 'automatic'" class="text-xs text-text-secondary">
              {{ t('setting.adFilter.automaticHint') }}
            </p>
            <label class="flex items-center gap-2 text-xs"
              ><input v-model="config.include_self_promotion" type="checkbox" />{{
                t('setting.adFilter.selfPromotion')
              }}</label
            >
            <p class="text-xs text-text-secondary">{{ t('setting.adFilter.selfPromotionHint') }}</p>
          </template>
        </div>
        <label class="block space-y-1"
          ><span class="font-medium">{{ t('setting.adFilter.exceptions') }}</span
          ><textarea
            v-model="config.allowlist"
            rows="2"
            class="filter-input font-mono"
            placeholder="example.org"
            spellcheck="false"
          /><span class="block text-xs text-text-secondary">{{
            t('setting.adFilter.exceptionsHint')
          }}</span></label
        >
        <details class="rounded-lg border border-border p-3">
          <summary class="cursor-pointer text-xs font-medium">
            {{ t('setting.adFilter.rules') }}
          </summary>
          <p class="text-xs text-text-secondary mt-2 mb-2">{{ t('setting.adFilter.rulesHint') }}</p>
          <textarea
            v-model="config.rules"
            :aria-label="t('setting.adFilter.rules')"
            rows="4"
            class="filter-input font-mono"
            placeholder="example.org##.promotion&#10;example.org#@#.keep"
            spellcheck="false"
          />
        </details>
        <p class="text-xs text-text-secondary">{{ t('setting.adFilter.scope') }}</p>
        <div class="flex flex-wrap gap-2">
          <button
            type="button"
            class="px-3 py-2 rounded-lg bg-accent text-white disabled:opacity-40"
            :disabled="!dirty"
            @click="save"
          >
            {{ t('setting.adFilter.save') }}</button
          ><button
            v-if="dirty"
            type="button"
            class="px-3 py-2 rounded-lg border border-border"
            @click="cancelDraft"
          >
            {{ t('setting.adFilter.reset') }}
          </button>
        </div>
      </fieldset>
      <details v-if="baseline" class="rounded-xl border border-border p-3">
        <summary class="font-medium cursor-pointer">{{ t('setting.adFilter.preview') }}</summary>
        <div class="mt-3 space-y-3">
          <p class="text-xs text-text-secondary">{{ t('setting.adFilter.previewHint') }}</p>
          <label class="block text-xs space-y-1"
            ><span>{{ t('setting.adFilter.sampleURL') }}</span
            ><input v-model="sampleURL" class="filter-input" type="url" /></label
          ><label class="block text-xs space-y-1"
            ><span>{{ t('setting.adFilter.sampleHTML') }}</span
            ><textarea
              v-model="sampleHTML"
              rows="5"
              class="filter-input font-mono"
              spellcheck="false"
            />
          </label>
          <div class="flex flex-wrap gap-2">
            <button
              type="button"
              class="px-3 py-2 rounded-lg border border-border disabled:opacity-40"
              :disabled="testing || saving || !config.enabled"
              @click="preview(false)"
            >
              {{ t('setting.adFilter.previewBasic') }}</button
            ><button
              type="button"
              class="px-3 py-2 rounded-lg border border-border text-accent disabled:opacity-40"
              :disabled="testing || saving || !config.enabled || !config.ai_enabled"
              @click="preview(true)"
            >
              {{ t('setting.adFilter.previewAI') }}
            </button>
          </div>
          <p v-if="testing" role="status" class="text-xs text-text-secondary">
            {{ t('setting.adFilter.analyzing') }}
          </p>
          <div v-if="result" class="space-y-2">
            <p class="text-xs font-medium">{{ t('setting.adFilter.previewResult') }}</p>
            <pre
              class="p-3 bg-bg-tertiary rounded-lg whitespace-pre-wrap break-all text-xs max-h-64 overflow-auto"
              >{{ result }}</pre>
            <p v-if="analysis" class="text-xs text-text-secondary">
              {{
                t(
                  analysis.decisions.length
                    ? 'setting.adFilter.aiApplied'
                    : analysis.guarded
                      ? 'setting.adFilter.guarded'
                      : 'setting.adFilter.noSuggestions',
                  { count: analysis.decisions.length }
                )
              }}
            </p>
          </div>
        </div>
      </details>
    </div>
  </SettingGroup>
</template>

<style scoped>
@reference "../../../../style.css";
.filter-input {
  @apply block w-full min-w-0 rounded-lg border border-border bg-bg-primary p-2 text-sm text-text-primary focus:border-accent focus:outline-none;
}
</style>
