<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhSparkle, PhCopy, PhArrowSquareOut, PhSpinner } from '@phosphor-icons/vue';
import BaseModal from '@/components/common/BaseModal.vue';
import BaseSelect from '@/components/common/BaseSelect.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';
import { useAIProfiles } from '@/composables/ai/useAIProfiles';
import { useReadingReport } from '@/composables/ai/useReadingReport';
import { copyToClipboard } from '@/utils/clipboard';
import { openInBrowser } from '@/utils/browser';
import type { Article } from '@/types/models';

const props = defineProps<{ articles: Article[] }>();
const emit = defineEmits<{ close: [] }>();
const { t } = useI18n();
// Freeze the selection when opened, so background refresh cannot change a report.
const candidates = props.articles.slice(0, 20);
const count = ref(Math.min(10, candidates.length));
const focus = ref('');
const profile = ref('0');
const { profiles, fetchProfiles } = useAIProfiles();
const { sources, result, previewing, generating, error, preview, generate, cancel } =
  useReadingReport();
const ids = computed(() => candidates.slice(0, Number(count.value)).map((article) => article.id));
const validCount = computed(
  () =>
    Number.isInteger(Number(count.value)) &&
    Number(count.value) >= 1 &&
    Number(count.value) <= candidates.length
);
const available = computed(
  () => sources.value.filter((source) => source.kind !== 'missing').length
);
const profileOptions = computed(() => [
  { value: '0', label: t('article.report.summaryProfile') },
  ...profiles.value.map((item) => ({ value: String(item.id), label: item.name })),
]);
watch(
  ids,
  () => {
    if (validCount.value) void preview(ids.value);
  },
  { immediate: true }
);
onMounted(() => {
  void fetchProfiles();
});

function openSettings() {
  emit('close');
  window.dispatchEvent(new CustomEvent('show-settings', { detail: { tab: 'ai' } }));
}
function sourceByID(id: number) {
  return result.value?.sources.find((source) => source.id === id);
}
function showSource(id: number) {
  const source = sourceByID(id);
  if (source?.url) void openInBrowser(source.url);
}
async function copyReport() {
  if (!result.value) return;
  const { report, sources: used } = result.value;
  const escape = (text: string) => text.replace(/[\\`*_[\]<>#]/g, '\\$&');
  const lines = [`# ${t('article.report.title')}`, '', escape(report.overview), ''];
  for (const topic of report.topics)
    lines.push(
      `## ${escape(topic.title)}`,
      '',
      escape(topic.summary),
      '',
      topic.source_ids.map((id) => `[${id}]`).join(' '),
      ''
    );
  if (report.reading_order.length)
    lines.push(
      `## ${t('article.report.readingOrder')}`,
      '',
      ...report.reading_order.map((item) => `- [${item.source_id}] ${escape(item.reason)}`),
      ''
    );
  lines.push(
    `## ${t('article.report.sources')}`,
    '',
    ...used.map(
      (source) =>
        `[${source.id}] ${escape(source.title)} — ${source.url} (${t(`article.report.${source.kind}`)}${source.truncated ? `, ${t('article.report.truncated')}` : ''})`
    )
  );
  if (report.caveats.length)
    lines.push(
      '',
      `## ${t('article.report.caveats')}`,
      '',
      ...report.caveats.map((item) => `- ${escape(item)}`)
    );
  const copied = await copyToClipboard(lines.join('\n'));
  window.showToast(
    t(copied ? 'common.toast.copiedToClipboard' : 'common.errors.failedToCopy'),
    copied ? 'success' : 'error'
  );
}
</script>

<template>
  <BaseModal :title="t('article.report.title')" size="4xl" @close="emit('close')">
    <div class="space-y-5 p-4 sm:p-6 text-text-primary" data-testid="reading-report">
      <p class="text-sm text-text-secondary">{{ t('article.report.description') }}</p>
      <div v-if="candidates.length" class="grid gap-4 sm:grid-cols-2">
        <label class="space-y-2 text-sm">
          <span>{{ t('article.report.count', { max: candidates.length }) }}</span>
          <input
            v-model.number="count"
            type="number"
            min="1"
            :max="candidates.length"
            :disabled="generating"
            class="block w-full rounded-lg border border-border bg-bg-secondary px-3 py-2"
            data-testid="report-count"
          />
        </label>
        <div class="space-y-2 text-sm">
          <span>{{ t('article.report.profile') }}</span>
          <BaseSelect
            v-model="profile"
            :options="profileOptions"
            :disabled="generating"
            width="w-full"
          />
        </div>
      </div>
      <label class="block space-y-2 text-sm">
        <span>{{ t('article.report.focus') }}</span>
        <textarea
          v-model="focus"
          :disabled="generating"
          maxlength="500"
          rows="2"
          :placeholder="t('article.report.focusPlaceholder')"
          class="block w-full rounded-lg border border-border bg-bg-secondary px-3 py-2"
          data-testid="report-focus"
        />
      </label>
      <div class="rounded-xl border border-border bg-bg-secondary p-3">
        <div class="flex items-center justify-between gap-2 text-sm font-medium">
          <span>{{ t('article.report.coverage', { available, total: sources.length }) }}</span>
          <PhSpinner v-if="previewing" :size="18" class="animate-spin" />
        </div>
        <p class="mt-1 text-xs text-text-secondary">{{ t('article.report.localOnly') }}</p>
        <ol class="mt-3 max-h-52 space-y-2 overflow-y-auto">
          <li v-for="source in sources" :key="source.id" class="text-sm">
            <details>
              <summary class="cursor-pointer">
                <span class="font-medium">[{{ source.id }}] {{ source.title }}</span>
                <span class="text-xs text-text-secondary"
                  >· {{ t(`article.report.${source.kind}`)
                  }}{{ source.truncated ? ` · ${t('article.report.truncated')}` : '' }}</span
                >
              </summary>
              <p class="mt-1 pl-4 text-xs text-text-secondary">
                {{ source.excerpt || t('article.report.noExcerpt') }}
              </p>
            </details>
          </li>
        </ol>
        <p v-if="!previewing && !available" class="mt-2 text-sm text-text-secondary">
          {{ t('article.report.noContent') }}
        </p>
      </div>
      <div v-if="error" role="alert" class="rounded-lg border border-border p-3 text-sm">
        {{ error }}
        <button class="ml-2 text-accent hover:underline" @click="openSettings">
          {{ t('article.report.settings') }}
        </button>
        <button
          v-if="!generating && !previewing"
          class="ml-2 text-accent hover:underline"
          @click="preview(ids)"
        >
          {{ t('article.report.retryPreview') }}
        </button>
      </div>
      <p
        v-if="generating"
        role="status"
        class="flex items-center gap-2 text-sm text-text-secondary"
      >
        <PhSparkle :size="18" class="text-accent" />{{ t('article.report.generating') }}
      </p>
      <section
        v-if="result"
        class="space-y-5 border-t border-border pt-5"
        data-testid="report-result"
      >
        <div class="flex items-center justify-between gap-3">
          <div>
            <h4 class="font-semibold">{{ t('article.report.result') }}</h4>
            <p class="text-xs text-text-secondary">
              {{
                t('article.report.provenance', {
                  count: result.sources.length,
                  model: result.model,
                })
              }}
            </p>
          </div>
          <button
            class="flex items-center gap-1 rounded-lg border border-border px-3 py-2 text-sm hover:bg-bg-tertiary"
            @click="copyReport"
          >
            <PhCopy :size="16" />{{ t('article.report.copy') }}
          </button>
        </div>
        <p class="whitespace-pre-line leading-relaxed">{{ result.report.overview }}</p>
        <div
          v-for="(topic, index) in result.report.topics"
          :key="index"
          class="rounded-xl border border-border p-4"
        >
          <h4 class="mb-2 font-semibold">{{ topic.title }}</h4>
          <p class="whitespace-pre-line text-sm leading-relaxed">{{ topic.summary }}</p>
          <div class="mt-3 flex flex-wrap gap-2">
            <button
              v-for="id in topic.source_ids"
              :key="id"
              :disabled="!sourceByID(id)?.url"
              :title="sourceByID(id)?.title"
              class="flex items-center gap-1 rounded-full bg-accent/10 px-2 py-1 text-xs text-accent disabled:opacity-50"
              @click="showSource(id)"
            >
              [{{ id }}] {{ sourceByID(id)?.title }}<PhArrowSquareOut :size="12" />
            </button>
          </div>
        </div>
        <div v-if="result.report.reading_order.length">
          <h4 class="mb-2 font-semibold">{{ t('article.report.readingOrder') }}</h4>
          <ol class="space-y-2 text-sm">
            <li v-for="item in result.report.reading_order" :key="item.source_id">
              <button
                :disabled="!sourceByID(item.source_id)?.url"
                class="text-accent hover:underline"
                @click="showSource(item.source_id)"
              >
                [{{ item.source_id }}] {{ sourceByID(item.source_id)?.title }}
              </button>
              <p class="mt-1 text-text-secondary">{{ item.reason }}</p>
            </li>
          </ol>
        </div>
        <div v-if="result.report.caveats.length">
          <h4 class="mb-2 font-semibold">{{ t('article.report.caveats') }}</h4>
          <ul class="list-disc space-y-1 pl-5 text-sm text-text-secondary">
            <li v-for="(item, index) in result.report.caveats" :key="index">{{ item }}</li>
          </ul>
        </div>
        <p class="text-xs text-text-secondary">{{ t('article.report.transient') }}</p>
        <details class="text-sm">
          <summary class="cursor-pointer font-medium">{{ t('article.report.sources') }}</summary>
          <ul class="mt-2 space-y-1 text-text-secondary">
            <li v-for="source in result.sources" :key="source.id">
              [{{ source.id }}] {{ source.title }} · {{ t(`article.report.${source.kind}`)
              }}{{ source.truncated ? ` · ${t('article.report.truncated')}` : '' }}
            </li>
          </ul>
        </details>
      </section>
    </div>
    <template #footer>
      <ModalFooter
        :secondary-button="{ label: generating ? t('article.report.stop') : t('common.close') }"
        :primary-button="{
          label: result ? t('article.report.regenerate') : t('article.report.generate'),
          loading: generating,
          disabled: !validCount || previewing || !available,
        }"
        @secondary-click="generating ? cancel() : emit('close')"
        @primary-click="generate(ids, Number(profile), focus)"
      />
    </template>
  </BaseModal>
</template>
