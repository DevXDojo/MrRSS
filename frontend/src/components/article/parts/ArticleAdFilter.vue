<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhShieldCheck, PhSparkle, PhSpinnerGap } from '@phosphor-icons/vue';
import type { AdAnalysis } from '@/types/adFilter';
const { t, te } = useI18n();
defineProps<{
  removed: number;
  aiEnabled: boolean;
  analysis: AdAnalysis | null;
  busy: boolean;
  error: string;
  original: boolean;
  applied: boolean;
}>();
const emit = defineEmits<{ analyze: []; apply: []; toggleOriginal: []; cancel: [] }>();
function errorText(code: string) {
  const key = `setting.adFilter.errors.${code}`;
  return te(key) ? t(key) : t('setting.adFilter.errors.ai_unavailable');
}
</script>

<template>
  <section
    class="mb-4 rounded-xl border border-border bg-bg-secondary p-3 text-xs space-y-2"
    :aria-label="t('setting.adFilter.title')"
  >
    <div class="flex flex-wrap items-center gap-2">
      <PhShieldCheck :size="16" class="text-accent" />
      <span class="font-medium">{{ t('setting.adFilter.title') }}</span>
      <span v-if="removed" class="text-text-secondary">{{
        t('setting.adFilter.removed', { count: removed })
      }}</span>
      <span v-if="original" class="text-amber-600">{{ t('setting.adFilter.originalShown') }}</span>
      <button
        v-if="removed || analysis?.decisions.length"
        type="button"
        class="text-accent ml-auto"
        @click="emit('toggleOriginal')"
      >
        {{ t(original ? 'setting.adFilter.showFiltered' : 'setting.adFilter.showOriginal') }}
      </button>
    </div>
    <div v-if="aiEnabled" class="flex flex-wrap items-center gap-3">
      <button
        v-if="!busy"
        type="button"
        class="inline-flex items-center gap-1 text-accent"
        @click="emit('analyze')"
      >
        <PhSparkle :size="14" />{{
          t(analysis ? 'setting.adFilter.analyzeAgain' : 'setting.adFilter.analyze')
        }}
      </button>
      <span v-else role="status" class="inline-flex items-center gap-2 text-text-secondary"
        ><PhSpinnerGap :size="14" class="animate-spin" />{{ t('setting.adFilter.analyzing')
        }}<button type="button" class="text-accent" @click="emit('cancel')">
          {{ t('common.cancel') }}
        </button></span
      >
      <button
        v-if="analysis?.decisions.length && !applied && !busy"
        type="button"
        class="px-2 py-1 bg-accent text-white rounded-md"
        @click="emit('apply')"
      >
        {{ t('setting.adFilter.applySuggestions', { count: analysis.decisions.length }) }}
      </button>
      <span v-if="applied && analysis?.decisions.length && !original" class="text-accent">{{
        t('setting.adFilter.aiApplied', { count: analysis.decisions.length })
      }}</span>
    </div>
    <p v-if="error" role="alert" class="text-red-500">
      {{ errorText(error) }} {{ t('setting.adFilter.contentPreserved') }}
    </p>
    <template v-if="analysis">
      <p v-if="!analysis.decisions.length" class="text-text-secondary">
        {{ t(analysis.guarded ? 'setting.adFilter.guarded' : 'setting.adFilter.noSuggestions') }}
      </p>
      <p v-if="analysis.truncated" class="text-amber-600">
        {{ t('setting.adFilter.partial', { count: analysis.scanned, total: analysis.total }) }}
      </p>
      <details v-if="analysis.decisions.length" class="text-text-secondary">
        <summary class="cursor-pointer">{{ t('setting.adFilter.reviewEvidence') }}</summary>
        <ul class="mt-2 space-y-2">
          <li
            v-for="decision in analysis.decisions"
            :key="decision.id"
            class="border-l-2 border-accent pl-2"
          >
            <span class="font-medium">{{
              t(`setting.adFilter.categories.${decision.category}`)
            }}</span>
            <p class="break-words mt-1">{{ decision.evidence }}</p>
          </li>
        </ul>
        <p class="mt-2">{{ t('setting.adFilter.aiCaution') }}</p>
      </details>
    </template>
  </section>
</template>
