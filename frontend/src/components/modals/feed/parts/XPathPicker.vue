<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import BaseModal from '@/components/common/BaseModal.vue';
import { createXPathSnapshot } from '@/utils/xpathSnapshot';
import {
  flattenPreview,
  previewField,
  relativePickerXPath,
  matchesPickerGroup,
  containingPickerItem,
  pickerLink,
  suggestPickerItems,
  inferPickerField,
  suggestOptionalPickerFields,
  type XPathPreviewNode,
  type XPathPickerField,
  type XPathSelection,
} from '@/utils/xpathPicker';

const props = defineProps<{ url: string; proxyEnabled?: boolean; proxyUrl?: string }>();
const emit = defineEmits<{ close: []; apply: [selection: XPathSelection] }>();
const { t } = useI18n();
const tree = ref<XPathPreviewNode | null>(null);
const frame = ref<HTMLIFrameElement | null>(null);
const token = ref('');
const loading = ref(false);
const failed = ref(false);
const candidate = ref<XPathPreviewNode | null>(null);
const item = ref<XPathPreviewNode | null>(null);
const active = ref<XPathPickerField>('item');
const manual = ref(false);
const seed = ref<XPathPreviewNode | null>(null);
const secondSeed = ref<XPathPreviewNode | null>(null);
const calibrating = ref(false);
const fieldCalibrating = ref(false);
const fieldSeeds = ref<Partial<Record<XPathPickerField, XPathPreviewNode>>>({});
const simplified = ref(false);
const resultPage = ref(0);
const missingOnly = ref(false);
const pageSize = 5;
function savedState() {
  return {
    selection: { ...selection.value },
    item: item.value,
    candidate: candidate.value,
    manual: manual.value,
    seed: seed.value,
    secondSeed: secondSeed.value,
    active: active.value,
    fieldSeeds: { ...fieldSeeds.value },
  };
}
type PickerState = ReturnType<typeof savedState>;
const undoStack = ref<PickerState[]>([]);
const redoStack = ref<PickerState[]>([]);
function checkpoint() {
  undoStack.value = [...undoStack.value.slice(-29), savedState()];
  redoStack.value = [];
}
function restore(state: PickerState) {
  selection.value = { ...state.selection };
  item.value = state.item;
  manual.value = state.manual;
  seed.value = state.seed;
  secondSeed.value = state.secondSeed;
  active.value = state.active;
  fieldSeeds.value = { ...state.fieldSeeds };
  candidate.value = state.candidate;
  calibrating.value = false;
  fieldCalibrating.value = false;
}
function undo() {
  const state = undoStack.value.pop();
  if (!state) return;
  redoStack.value.push(savedState());
  restore(state);
}
function redo() {
  const state = redoStack.value.pop();
  if (!state) return;
  undoStack.value.push(savedState());
  restore(state);
}
const suggestions = computed(() =>
  seed.value ? suggestPickerItems(seed.value, byPath.value, secondSeed.value) : []
);
const guidedStart = computed(() => !item.value && !manual.value);
const emptySelection = (): XPathSelection => ({
  item: '',
  title: '',
  uri: '',
  timestamp: '',
  content: '',
  thumbnail: '',
});
const selection = ref(emptySelection());
const fields: XPathPickerField[] = ['item', 'title', 'uri', 'timestamp', 'content', 'thumbnail'];
const fieldLabel = (field: XPathPickerField) =>
  t(`modal.feed.picker.${field === 'title' ? 'titleField' : field}`);
const nodes = computed(() =>
  tree.value ? flattenPreview(tree.value).filter((node) => node.path && node.tag) : []
);
const byPath = computed(() => new Map(nodes.value.map((node) => [node.path!, node])));
const matches = computed(() =>
  item.value ? nodes.value.filter((node) => matchesPickerGroup(node, item.value!)) : []
);
const highlightedItems = computed(() =>
  active.value === 'item' && candidate.value
    ? nodes.value.filter((node) => matchesPickerGroup(node, candidate.value!))
    : matches.value
);
const ancestors = computed(() => {
  const result: XPathPreviewNode[] = [];
  let node = candidate.value;
  while (node?.path) {
    if (!['html', 'body'].includes(node.tag ?? '')) result.unshift(node);
    node = byPath.value.get(node.path.slice(0, node.path.lastIndexOf('/'))) ?? null;
  }
  const root =
    candidate.value &&
    active.value !== 'item' &&
    containingPickerItem(candidate.value, matches.value);
  return root
    ? result.filter((node) => node.path === root.path || node.path?.startsWith(root.path + '/'))
    : result;
});
const candidateItem = computed(() =>
  candidate.value ? containingPickerItem(candidate.value, matches.value) : undefined
);
const fieldNode = computed(() => {
  if (!candidate.value || !candidateItem.value) return null;
  return active.value === 'uri'
    ? pickerLink(candidate.value, candidateItem.value, byPath.value)
    : candidate.value;
});
const generated = computed(() => {
  if (calibrating.value) return null;
  if (!candidate.value) return null;
  if (active.value === 'item') return candidate.value.group ?? null;
  if (fieldCalibrating.value) {
    const first = fieldSeeds.value[active.value];
    const firstItem = first && containingPickerItem(first, matches.value);
    return first && firstItem && fieldNode.value && candidateItem.value
      ? inferPickerField(
          first,
          firstItem,
          fieldNode.value,
          candidateItem.value,
          active.value,
          byPath.value
        )
      : null;
  }
  return candidateItem.value && fieldNode.value
    ? relativePickerXPath(candidateItem.value, fieldNode.value, active.value)
    : null;
});
const displaySelection = computed(() =>
  item.value && active.value !== 'item' && generated.value
    ? { ...selection.value, [active.value]: generated.value }
    : selection.value
);
const validSamples = computed(() =>
  !displaySelection.value.title || !displaySelection.value.uri
    ? []
    : matches.value.filter(
        (node) =>
          previewField(node, displaySelection.value.title, byPath.value).trim() &&
          previewField(node, displaySelection.value.uri, byPath.value).trim()
      )
);
const fieldCoverage = computed(() =>
  Object.fromEntries(
    fields.map((field) => [
      field,
      displaySelection.value[field]
        ? matches.value.filter((node) =>
            previewField(node, displaySelection.value[field], byPath.value).trim()
          ).length
        : 0,
    ])
  )
);
const resultItems = computed(() =>
  highlightedItems.value.filter(
    (node) =>
      !missingOnly.value ||
      !previewField(node, displaySelection.value.title, byPath.value).trim() ||
      !previewField(node, displaySelection.value.uri, byPath.value).trim()
  )
);
const pageCount = computed(() => Math.max(1, Math.ceil(resultItems.value.length / pageSize)));
const visibleResults = computed(() =>
  resultItems.value.slice(resultPage.value * pageSize, (resultPage.value + 1) * pageSize)
);
watch([resultItems, missingOnly], () => {
  resultPage.value = 0;
});
const snapshot = computed(() =>
  tree.value?.html
    ? createXPathSnapshot(tree.value.html, tree.value.base_url || props.url, token.value)
    : ''
);
function highlight() {
  frame.value?.contentWindow?.postMessage(
    {
      type: 'mrrss-xpath-highlight',
      token: token.value,
      path: candidate.value?.path,
      matches: highlightedItems.value.map((node) => node.path),
      simplified: simplified.value,
    },
    '*'
  );
}
function receive(event: MessageEvent) {
  if (event.source !== frame.value?.contentWindow || event.data?.token !== token.value) return;
  if (event.data.type === 'mrrss-xpath-pick' && typeof event.data.path === 'string') {
    const picked = byPath.value.get(event.data.path) ?? null;
    if (guidedStart.value) {
      if (calibrating.value) {
        secondSeed.value = picked;
        calibrating.value = false;
      } else {
        seed.value = picked;
        secondSeed.value = null;
      }
      candidate.value = suggestions.value[0] ?? null;
    } else candidate.value = picked;
  }
  if (event.data.type === 'mrrss-xpath-confirm' && generated.value) choose();
  if (event.data.type === 'mrrss-xpath-undo') undo();
  if (event.data.type === 'mrrss-xpath-redo') redo();
  if (event.data.type === 'mrrss-xpath-navigate') navigateCandidate(event.data.direction);
  if (event.data.type === 'mrrss-xpath-ready') highlight();
}
watch([candidate, highlightedItems, simplified], highlight);
function locate(node: XPathPreviewNode) {
  frame.value?.contentWindow?.postMessage(
    { type: 'mrrss-xpath-locate', token: token.value, path: node.path },
    '*'
  );
}
function navigateCandidate(direction: string) {
  if (!candidate.value?.path) return;
  const parent = byPath.value.get(
    candidate.value.path.slice(0, candidate.value.path.lastIndexOf('/'))
  );
  const next =
    direction === 'parent'
      ? parent
      : candidate.value.children?.find((node) => node.path && node.tag);
  if (!next || ['html', 'body'].includes(next.tag ?? '')) return;
  if (active.value !== 'item' && !containingPickerItem(next, matches.value)) return;
  candidate.value = next;
  locate(next);
}
function clearField() {
  checkpoint();
  selection.value[active.value] = '';
  candidate.value = null;
  fieldCalibrating.value = false;
}
let controller: AbortController | null = null;
let generation = 0;
async function load() {
  const request = ++generation;
  controller?.abort();
  controller = new AbortController();
  loading.value = true;
  failed.value = false;
  tree.value = null;
  item.value = null;
  candidate.value = null;
  selection.value = emptySelection();
  active.value = 'item';
  manual.value = false;
  seed.value = null;
  secondSeed.value = null;
  calibrating.value = false;
  fieldCalibrating.value = false;
  fieldSeeds.value = {};
  undoStack.value = [];
  redoStack.value = [];
  token.value = Array.from(crypto.getRandomValues(new Uint8Array(24)), (byte) =>
    byte.toString(16).padStart(2, '0')
  ).join('');
  try {
    const response = await fetch('/api/feeds/xpath-preview', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      signal: controller.signal,
      body: JSON.stringify({
        url: props.url,
        proxy_enabled: props.proxyEnabled,
        proxy_url: props.proxyUrl,
      }),
    });
    if (!response.ok) throw new Error('Preview failed');
    const data: XPathPreviewNode = await response.json();
    if (request === generation) tree.value = data;
  } catch {
    if (request === generation) {
      failed.value = true;
      window.showToast(t('modal.feed.picker.loadFailed'), 'error');
    }
  } finally {
    if (request === generation) loading.value = false;
  }
}
function choose() {
  if (!candidate.value || !generated.value) return;
  checkpoint();
  if (active.value === 'item') {
    const inferFields = guidedStart.value;
    item.value = candidate.value;
    selection.value = { ...emptySelection(), item: generated.value };
    if (inferFields && seed.value) {
      selection.value.title = relativePickerXPath(candidate.value, seed.value, 'title') ?? '';
      const link = pickerLink(seed.value, candidate.value, byPath.value);
      if (link) selection.value.uri = relativePickerXPath(candidate.value, link, 'uri') ?? '';
      fieldSeeds.value = { title: seed.value, ...(link ? { uri: link } : {}) };
      if (secondSeed.value) {
        const secondItem = containingPickerItem(secondSeed.value, matches.value);
        const secondLink = secondItem && pickerLink(secondSeed.value, secondItem, byPath.value);
        if (secondItem)
          selection.value.title =
            inferPickerField(
              seed.value,
              candidate.value,
              secondSeed.value,
              secondItem,
              'title',
              byPath.value
            ) ?? '';
        if (secondItem && link && secondLink)
          selection.value.uri =
            inferPickerField(link, candidate.value, secondLink, secondItem, 'uri', byPath.value) ??
            '';
      }
    }
    if (inferFields) {
      Object.assign(
        selection.value,
        suggestOptionalPickerFields(candidate.value, matches.value, byPath.value)
      );
    }
    active.value = selection.value.uri ? 'content' : 'title';
  } else {
    selection.value[active.value] = generated.value;
    if (!fieldCalibrating.value && fieldNode.value)
      fieldSeeds.value[active.value] = fieldNode.value;
    if (active.value === 'title' && candidateItem.value) {
      const link = pickerLink(candidate.value, candidateItem.value, byPath.value);
      if (link && !selection.value.uri) {
        selection.value.uri = relativePickerXPath(candidateItem.value, link, 'uri') ?? '';
        fieldSeeds.value.uri = link;
      }
      active.value = selection.value.uri ? 'timestamp' : 'uri';
    }
  }
  candidate.value = null;
  fieldCalibrating.value = false;
}
function startManual() {
  checkpoint();
  manual.value = true;
  active.value = 'item';
  item.value = null;
  selection.value = emptySelection();
  candidate.value = null;
  seed.value = null;
  secondSeed.value = null;
  calibrating.value = false;
  fieldCalibrating.value = false;
  fieldSeeds.value = {};
}
function clearCalibration() {
  secondSeed.value = null;
  calibrating.value = false;
  candidate.value = suggestions.value[0] ?? null;
}
function apply() {
  if (candidate.value && generated.value) choose();
  emit('apply', { ...selection.value });
}
onMounted(() => {
  window.addEventListener('message', receive);
  load();
});
onBeforeUnmount(() => {
  generation++;
  controller?.abort();
  window.removeEventListener('message', receive);
});
</script>

<template>
  <Teleport to="body">
    <BaseModal
      :title="t('modal.feed.picker.title')"
      size="full"
      height="full"
      :z-index="80"
      show-footer
      body-class="p-3 sm:p-4 flex flex-col min-h-0"
      footer-class="flex items-center justify-between gap-3"
      @close="emit('close')"
    >
      <div class="flex items-center justify-between gap-3 mb-3 text-sm">
        <p class="font-medium" role="status">
          {{ t(item ? 'modal.feed.picker.stepFields' : 'modal.feed.picker.stepTitle') }}
        </p>
        <details class="text-xs text-text-secondary max-w-lg">
          <summary class="cursor-pointer">{{ t('modal.feed.picker.previewAbout') }}</summary>
          <p class="mt-1">{{ t('modal.feed.picker.staticHint') }}</p>
        </details>
      </div>
      <div v-if="tree" class="flex flex-wrap items-center gap-2 mb-3 text-xs">
        <button
          type="button"
          class="px-3 py-1.5 border border-border rounded-lg disabled:opacity-40"
          :disabled="!undoStack.length"
          @click="undo"
        >
          {{ t('modal.feed.picker.undo') }}
        </button>
        <button
          type="button"
          class="px-3 py-1.5 border border-border rounded-lg disabled:opacity-40"
          :disabled="!redoStack.length"
          @click="redo"
        >
          {{ t('modal.feed.picker.redo') }}
        </button>
        <label class="flex items-center gap-2 ml-auto cursor-pointer"
          ><input v-model="simplified" type="checkbox" />{{
            t('modal.feed.picker.simplified')
          }}</label
        >
      </div>
      <p v-if="loading" class="text-text-secondary" role="status">
        {{ t('common.state.loading') }}
      </p>
      <button v-else-if="failed" type="button" class="text-accent" @click="load">
        {{ t('modal.feed.picker.retry') }}
      </button>
      <div
        v-else-if="tree"
        class="flex-1 min-h-0 grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_22rem] gap-4 overflow-auto lg:overflow-hidden"
      >
        <div class="min-h-0 flex flex-col">
          <div
            class="flex flex-wrap items-center justify-between gap-2 text-xs text-text-secondary pb-2"
          >
            <span>{{ t('modal.feed.picker.pageInstruction') }}</span>
            <span>{{ t('modal.feed.picker.highlightLegend') }}</span>
          </div>
          <iframe
            ref="frame"
            :srcdoc="snapshot"
            sandbox="allow-scripts"
            credentialless
            referrerpolicy="no-referrer"
            :title="t('modal.feed.picker.pagePreview')"
            class="w-full h-[55vh] lg:flex-1 lg:h-auto min-h-64 border border-border rounded-lg bg-white"
          />
        </div>
        <aside class="space-y-4 text-sm min-w-0 lg:overflow-auto pr-1">
          <section
            v-if="guidedStart"
            class="p-3 bg-bg-secondary rounded-lg border border-border space-y-3"
          >
            <h3 class="font-semibold">{{ t('modal.feed.picker.startTitle') }}</h3>
            <p class="text-text-secondary">{{ t('modal.feed.picker.startHint') }}</p>
            <template v-if="seed && suggestions.length">
              <p class="text-accent font-medium">
                {{ t('modal.feed.picker.suggested', { count: highlightedItems.length }) }}
              </p>
              <p class="text-xs text-text-secondary">
                {{ t('modal.feed.picker.confirmGroupHint') }}
              </p>
              <label v-if="suggestions.length > 1" class="block text-xs">
                {{ t('modal.feed.picker.otherGroup') }}
                <select
                  class="w-full mt-1 p-2 rounded border border-border bg-bg-primary"
                  :value="candidate?.path"
                  @change="
                    candidate =
                      suggestions.find(
                        (node) => node.path === ($event.target as HTMLSelectElement).value
                      ) ?? null
                  "
                >
                  <option v-for="node in suggestions" :key="node.path" :value="node.path">
                    {{ node.tag }}{{ node.classes?.length ? '.' + node.classes.join('.') : '' }}
                  </option>
                </select>
              </label>
            </template>
            <p v-else-if="seed" class="text-amber-600">{{ t('modal.feed.picker.noSuggestion') }}</p>
            <div v-if="seed" class="space-y-2 border-t border-border pt-2 text-xs">
              <p class="line-clamp-2">
                {{ t('modal.feed.picker.firstExample') }} {{ previewField(seed, '.', byPath) }}
              </p>
              <p v-if="secondSeed" class="line-clamp-2">
                {{ t('modal.feed.picker.secondExample') }}
                {{ previewField(secondSeed, '.', byPath) }}
              </p>
              <p v-if="calibrating" class="text-accent font-medium" role="status">
                {{ t('modal.feed.picker.calibrationHint') }}
              </p>
              <button v-else type="button" class="text-accent" @click="calibrating = true">
                {{ t('modal.feed.picker.calibrate') }}
              </button>
              <button
                v-if="secondSeed || calibrating"
                type="button"
                class="block text-text-secondary underline"
                @click="clearCalibration"
              >
                {{ t('modal.feed.picker.clearCalibration') }}
              </button>
            </div>
            <button
              type="button"
              class="text-xs text-accent underline underline-offset-2"
              @click="startManual"
            >
              {{ t('modal.feed.picker.manual') }}
            </button>
          </section>
          <section v-else class="space-y-3">
            <div class="flex items-center justify-between gap-2">
              <h3 class="font-semibold">{{ t('modal.feed.picker.fields') }}</h3>
              <button v-if="item" type="button" class="text-xs text-accent" @click="startManual">
                {{ t('modal.feed.picker.changeGroup') }}
              </button>
            </div>
            <div class="grid grid-cols-2 gap-2">
              <button
                v-for="field in fields.filter((field) => field !== 'item')"
                :key="field"
                type="button"
                class="text-left px-3 py-2 rounded-lg border disabled:opacity-40"
                :class="
                  active === field
                    ? 'border-accent bg-bg-tertiary text-accent'
                    : 'border-border bg-bg-secondary text-text-primary'
                "
                :aria-pressed="active === field"
                :disabled="!item"
                @click="
                  active = field;
                  candidate = null;
                  fieldCalibrating = false;
                "
              >
                <span class="block font-medium"
                  >{{ fieldLabel(field) }} {{ selection[field] ? '✓' : '' }}</span
                >
                <span class="block text-xs text-text-secondary mt-1">{{
                  displaySelection[field]
                    ? t('modal.feed.picker.coverage', {
                        count: fieldCoverage[field],
                        total: matches.length,
                      })
                    : t(
                        ['title', 'uri'].includes(field)
                          ? 'modal.feed.picker.required'
                          : 'modal.feed.picker.optional'
                      )
                }}</span>
              </button>
            </div>
            <p class="font-medium">
              {{ t('modal.feed.picker.selecting', { field: fieldLabel(active) }) }}
            </p>
            <p class="text-xs text-text-secondary">
              {{
                t(
                  active === 'item'
                    ? 'modal.feed.picker.containerHint'
                    : 'modal.feed.picker.withinItem'
                )
              }}
            </p>
            <p v-if="fieldCalibrating" class="text-accent text-xs" role="status">
              {{ t('modal.feed.picker.fieldCalibrationHint', { field: fieldLabel(active) }) }}
            </p>
            <button
              v-if="
                active !== 'item' && fieldSeeds[active] && selection[active] && !fieldCalibrating
              "
              type="button"
              class="block text-xs text-accent"
              @click="
                fieldCalibrating = true;
                candidate = null;
              "
            >
              {{ t('modal.feed.picker.calibrateField') }}
            </button>
            <button
              v-if="fieldCalibrating"
              type="button"
              class="block text-xs text-text-secondary underline"
              @click="
                fieldCalibrating = false;
                candidate = null;
              "
            >
              {{ t('modal.feed.picker.cancelCalibration') }}
            </button>
            <button
              v-if="active !== 'item' && selection[active]"
              type="button"
              class="block text-xs text-accent"
              @click="clearField"
            >
              {{ t('modal.feed.picker.clearField', { field: fieldLabel(active) }) }}
            </button>
            <div
              v-if="candidate"
              class="flex flex-wrap gap-1"
              :aria-label="t('modal.feed.picker.ancestors')"
            >
              <button
                v-for="node in ancestors"
                :key="node.path"
                type="button"
                class="px-2 py-1 rounded border border-border hover:text-accent break-all"
                :class="node.path === candidate?.path ? 'bg-bg-tertiary text-accent' : ''"
                @click="candidate = node"
              >
                {{ node.tag }}{{ node.classes?.length ? '.' + node.classes.join('.') : '' }}
              </button>
            </div>
            <div v-if="candidate" class="flex flex-wrap gap-3 text-xs">
              <button type="button" class="text-accent" @click="navigateCandidate('parent')">
                {{ t('modal.feed.picker.parent') }}
              </button>
              <button type="button" class="text-accent" @click="navigateCandidate('child')">
                {{ t('modal.feed.picker.child') }}
              </button>
              <span class="text-text-secondary">{{ t('modal.feed.picker.keyboardHint') }}</span>
            </div>
            <p v-if="candidate && !generated" class="text-amber-600 text-xs">
              {{
                t(
                  fieldCalibrating
                    ? 'modal.feed.picker.incompatibleExamples'
                    : 'modal.feed.picker.invalidElement'
                )
              }}
            </p>
          </section>
          <button
            v-if="generated"
            type="button"
            class="w-full px-3 py-2.5 rounded-lg bg-accent text-white font-medium"
            @click="choose"
          >
            {{ t(guidedStart ? 'modal.feed.picker.confirmGroup' : 'modal.feed.picker.choose') }}
          </button>
          <section
            v-if="item || (guidedStart && candidate)"
            class="border-t border-border pt-3 space-y-3"
          >
            <div class="flex justify-between items-center gap-2">
              <h3 class="font-semibold">{{ t('modal.feed.picker.results') }}</h3>
              <span class="text-xs text-text-secondary">{{
                t('modal.feed.picker.matches', { count: highlightedItems.length })
              }}</span>
            </div>
            <p
              v-if="displaySelection.title && displaySelection.uri"
              class="text-xs"
              :class="validSamples.length === matches.length ? 'text-accent' : 'text-amber-600'"
            >
              {{
                t('modal.feed.picker.validSamples', {
                  count: validSamples.length,
                  total: matches.length,
                })
              }}
            </p>
            <p v-if="validSamples.length" class="text-xs text-text-secondary">
              {{ t('modal.feed.picker.ready') }}
            </p>
            <ul class="space-y-2">
              <li
                v-for="match in visibleResults"
                :key="match.path"
                class="p-3 bg-bg-secondary border border-border rounded-lg space-y-1"
              >
                <button
                  type="button"
                  class="text-xs text-accent float-right ml-2"
                  @click="locate(match)"
                >
                  {{ t('modal.feed.picker.locate') }}
                </button>
                <p class="font-medium line-clamp-2">
                  {{ previewField(match, displaySelection.title || '.', byPath).slice(0, 200) }}
                </p>
                <p
                  v-if="displaySelection.uri"
                  class="text-xs text-text-secondary break-all line-clamp-2"
                >
                  {{ previewField(match, displaySelection.uri, byPath) }}
                </p>
                <p v-if="displaySelection.timestamp" class="text-xs text-text-secondary">
                  {{ previewField(match, displaySelection.timestamp, byPath) }}
                </p>
                <p v-if="displaySelection.content" class="text-xs text-text-secondary line-clamp-3">
                  {{ previewField(match, displaySelection.content, byPath) }}
                </p>
                <p
                  v-if="
                    item &&
                    (!previewField(match, displaySelection.title, byPath).trim() ||
                      !previewField(match, displaySelection.uri, byPath).trim())
                  "
                  class="text-xs text-amber-600"
                >
                  {{ t('modal.feed.picker.missingFields') }}
                </p>
              </li>
            </ul>
            <label v-if="item" class="flex items-center gap-2 text-xs"
              ><input v-model="missingOnly" type="checkbox" />{{
                t('modal.feed.picker.missingOnly')
              }}</label
            >
            <div v-if="pageCount > 1" class="flex items-center justify-between text-xs">
              <button
                type="button"
                class="text-accent disabled:opacity-40"
                :disabled="resultPage === 0"
                @click="resultPage--"
              >
                {{ t('modal.feed.picker.previous') }}
              </button>
              <span>{{ resultPage + 1 }} / {{ pageCount }}</span>
              <button
                type="button"
                class="text-accent disabled:opacity-40"
                :disabled="resultPage + 1 >= pageCount"
                @click="resultPage++"
              >
                {{ t('modal.feed.picker.next') }}
              </button>
            </div>
          </section>
          <details
            v-if="item || candidate"
            class="border-t border-border pt-3 text-xs text-text-secondary"
          >
            <summary class="cursor-pointer">{{ t('modal.feed.picker.advanced') }}</summary>
            <p v-if="generated" class="break-all mt-2">{{ generated }}</p>
            <dl class="mt-2 space-y-2">
              <template v-for="field in fields" :key="field">
                <dt>{{ fieldLabel(field) }}</dt>
                <dd class="break-all">{{ selection[field] || '—' }}</dd>
              </template>
            </dl>
          </details>
        </aside>
      </div>
      <template #footer>
        <p class="text-xs text-text-secondary">
          {{
            t(validSamples.length ? 'modal.feed.picker.applyHint' : 'modal.feed.picker.needFields')
          }}
        </p>
        <div class="flex gap-3 shrink-0">
          <button
            type="button"
            class="px-3 py-2 rounded-lg border border-border"
            @click="emit('close')"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            type="button"
            class="px-4 py-2 rounded-lg bg-accent text-white disabled:opacity-40"
            :disabled="
              !validSamples.length || active === 'item' || (fieldCalibrating && !generated)
            "
            @click="apply"
          >
            {{ t('modal.feed.picker.apply') }}
          </button>
        </div>
      </template>
    </BaseModal>
  </Teleport>
</template>
