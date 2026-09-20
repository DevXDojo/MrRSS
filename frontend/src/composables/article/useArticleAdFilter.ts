import { computed, ref, watch, onMounted, onBeforeUnmount } from 'vue';
import type { AdAnalysis, AdFilterConfig, ArticleFilterInfo } from '@/types/adFilter';
import { defaultAdFilterConfig, isAdFilterApplicable } from '@/types/adFilter';
import { proxyImagesInHtml } from '@/utils/mediaProxy';

interface Inputs {
  id: () => number;
  url: () => string;
  content: () => string;
  info: () => ArticleFilterInfo | undefined;
  mediaCache: () => boolean;
}

export function useArticleAdFilter(inputs: Inputs) {
  const options = ref<AdFilterConfig>(defaultAdFilterConfig());
  const analysis = ref<AdAnalysis | null>(null);
  const busy = ref(false);
  const error = ref('');
  const original = ref(false);
  const applied = ref(false);
  const details = ref(false);
  let generation = 0;
  let controller: AbortController | null = null;
  const configController = new AbortController();
  let automaticAttempt = -1;
  const applicable = computed(() => isAdFilterApplicable(options.value, inputs.url()));
  const source = computed(() => inputs.info()?.sourceContent ?? inputs.content());
  const removed = computed(() => inputs.info()?.report.removed || 0);
  const content = computed(() => {
    let output: string | undefined;
    if (original.value) output = inputs.info()?.originalContent || source.value;
    else if (applied.value && analysis.value) output = analysis.value.content;
    if (output === undefined) return inputs.content();
    return inputs.mediaCache() ? proxyImagesInHtml(output, inputs.url()) : output;
  });
  function reset() {
    generation++;
    controller?.abort();
    busy.value = false;
    error.value = '';
    analysis.value = null;
    original.value = false;
    applied.value = false;
  }
  function cancel() {
    generation++;
    controller?.abort();
    busy.value = false;
    automaticAttempt = generation;
  }
  async function analyze() {
    if (busy.value || !source.value || !options.value.ai_enabled || !applicable.value) return;
    const request = generation;
    automaticAttempt = generation;
    controller?.abort();
    controller = new AbortController();
    busy.value = true;
    error.value = '';
    try {
      const { revision: _revision, ...filterOptions } = options.value;
      const response = await fetch('/api/ad-filter/analyze', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        signal: controller.signal,
        body: JSON.stringify({ url: inputs.url(), html: source.value, options: filterOptions }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error?.message || 'ai_unavailable');
      if (request !== generation) return;
      analysis.value = data as AdAnalysis;
      if (options.value.ai_mode === 'automatic' && !original.value) applied.value = true;
    } catch (err) {
      if (request === generation && !controller?.signal.aborted)
        error.value = err instanceof Error ? err.message : 'ai_unavailable';
    } finally {
      if (request === generation) busy.value = false;
    }
  }
  function apply() {
    applied.value = true;
    original.value = false;
  }
  function toggleOriginal() {
    original.value = !original.value;
  }
  watch([inputs.id, inputs.content, inputs.info], reset, { flush: 'sync' });
  watch(
    [source, applicable, options],
    () => {
      if (options.value.ai_mode === 'automatic' && automaticAttempt !== generation) void analyze();
    },
    { flush: 'post', deep: true }
  );
  async function loadOptions() {
    try {
      const response = await fetch('/api/ad-filter/config', { signal: configController.signal });
      if (!response.ok) return;
      options.value = { ...defaultAdFilterConfig(), ...(await response.json()) };
    } catch {
      /* Article reading remains available if settings cannot be read. */
    }
  }
  function settingsChanged(event: Event) {
    configController.abort();
    reset();
    options.value = (event as CustomEvent<AdFilterConfig>).detail || defaultAdFilterConfig();
  }
  onMounted(() => {
    void loadOptions();
    window.addEventListener('ad-filter-changed', settingsChanged);
  });
  onBeforeUnmount(() => {
    reset();
    configController.abort();
    window.removeEventListener('ad-filter-changed', settingsChanged);
  });
  return {
    content,
    options,
    analysis,
    busy,
    error,
    original,
    applied,
    details,
    applicable,
    removed,
    analyze,
    apply,
    toggleOriginal,
    cancel,
  };
}
