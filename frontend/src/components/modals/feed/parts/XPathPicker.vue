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
const ancestors = computed(() => {
  const result: XPathPreviewNode[] = [];
  let node = candidate.value;
  while (node?.path) {
    if (!['html', 'body'].includes(node.tag ?? '')) result.unshift(node);
    node = byPath.value.get(node.path.slice(0, node.path.lastIndexOf('/'))) ?? null;
  }
  return result;
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
  if (!candidate.value) return null;
  if (active.value === 'item') return candidate.value.group ?? null;
  return candidateItem.value && fieldNode.value
    ? relativePickerXPath(candidateItem.value, fieldNode.value, active.value)
    : null;
});
const validSamples = computed(() =>
  !selection.value.title || !selection.value.uri
    ? []
    : matches.value.filter(
        (node) =>
          previewField(node, selection.value.title, byPath.value).trim() &&
          previewField(node, selection.value.uri, byPath.value).trim()
      )
);
const snapshot = computed(() =>
  tree.value?.html
    ? createXPathSnapshot(tree.value.html, tree.value.base_url || props.url, token.value)
    : ''
);
function highlight() {
  frame.value?.contentWindow?.postMessage(
    { type: 'mrrss-xpath-highlight', token: token.value, path: candidate.value?.path },
    '*'
  );
}
function receive(event: MessageEvent) {
  if (event.source !== frame.value?.contentWindow || event.data?.token !== token.value) return;
  if (event.data.type === 'mrrss-xpath-pick' && typeof event.data.path === 'string')
    candidate.value = byPath.value.get(event.data.path) ?? null;
  if (event.data.type === 'mrrss-xpath-ready') highlight();
}
watch(candidate, highlight);
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
  if (active.value === 'item') {
    item.value = candidate.value;
    selection.value = { ...emptySelection(), item: generated.value };
    active.value = 'title';
  } else {
    selection.value[active.value] = generated.value;
    if (active.value === 'title' && candidateItem.value) {
      const link = pickerLink(candidate.value, candidateItem.value, byPath.value);
      if (link && !selection.value.uri)
        selection.value.uri = relativePickerXPath(candidateItem.value, link, 'uri') ?? '';
      active.value = selection.value.uri ? 'timestamp' : 'uri';
    }
  }
  candidate.value = null;
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
      body-class="p-4 flex flex-col min-h-0"
      footer-class="flex justify-end gap-3"
      @close="emit('close')"
    >
      <p class="text-sm text-text-secondary mb-2">{{ t('modal.feed.picker.hint') }}</p>
      <p class="text-xs text-text-secondary mb-3">{{ t('modal.feed.picker.staticHint') }}</p>
      <p v-if="loading" class="text-text-secondary" role="status">
        {{ t('common.state.loading') }}
      </p>
      <button v-else-if="failed" type="button" class="text-accent" @click="load">
        {{ t('modal.feed.picker.retry') }}
      </button>
      <div
        v-else-if="tree"
        class="flex-1 min-h-0 grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_22rem] gap-4"
      >
        <iframe
          ref="frame"
          :srcdoc="snapshot"
          sandbox="allow-scripts"
          credentialless
          referrerpolicy="no-referrer"
          :title="t('modal.feed.picker.pagePreview')"
          class="w-full h-[55vh] lg:h-full min-h-64 border border-border rounded bg-white"
        />
        <aside class="space-y-3 text-sm min-w-0 overflow-auto pr-1">
          <div class="flex flex-wrap gap-2">
            <button
              v-for="field in fields"
              :key="field"
              type="button"
              class="px-3 py-2 rounded border border-border disabled:opacity-40"
              :class="
                active === field ? 'bg-accent text-white' : 'bg-bg-secondary text-text-primary'
              "
              :disabled="field !== 'item' && !item"
              @click="active = field"
            >
              {{ fieldLabel(field) }} {{ selection[field] ? '✓' : '' }}
            </button>
          </div>
          <p class="font-semibold">
            {{ t('modal.feed.picker.selecting', { field: fieldLabel(active) }) }}
          </p>
          <p class="text-text-secondary">
            {{
              t(
                active === 'item'
                  ? 'modal.feed.picker.containerHint'
                  : 'modal.feed.picker.withinItem'
              )
            }}
          </p>
          <div class="flex flex-wrap gap-1" :aria-label="t('modal.feed.picker.ancestors')">
            <button
              v-for="node in ancestors"
              :key="node.path"
              type="button"
              class="px-2 py-1 rounded border border-border hover:text-accent break-all"
              :class="node.path === candidate?.path ? 'bg-bg-tertiary' : ''"
              @click="candidate = node"
            >
              {{ node.tag }}{{ node.classes?.length ? '.' + node.classes.join('.') : '' }}
            </button>
          </div>
          <p class="break-all text-xs text-text-secondary">
            {{ candidate?.path || t('modal.feed.picker.selectElement') }}
          </p>
          <p v-if="candidate && !generated" class="text-amber-600">
            {{ t('modal.feed.picker.invalidElement') }}
          </p>
          <p v-if="generated" class="break-all text-xs text-accent">{{ generated }}</p>
          <button
            type="button"
            class="px-3 py-2 rounded bg-accent text-white disabled:opacity-40"
            :disabled="!generated"
            @click="choose"
          >
            {{ t('modal.feed.picker.choose') }}
          </button>
          <dl class="space-y-2 border-t border-border pt-3">
            <template v-for="field in fields" :key="field">
              <dt class="font-medium flex justify-between">
                {{ fieldLabel(field) }}
                <button
                  v-if="field !== 'item' && selection[field]"
                  type="button"
                  class="text-xs text-accent"
                  @click="
                    selection[field] = '';
                    active = field;
                  "
                >
                  {{ t('modal.feed.picker.clear') }}
                </button>
              </dt>
              <dd class="break-all text-xs text-text-secondary">{{ selection[field] || '—' }}</dd>
            </template>
          </dl>
          <p>{{ t('modal.feed.picker.matches', { count: matches.length }) }}</p>
          <p v-if="selection.title && selection.uri" class="text-xs text-text-secondary">
            {{
              t('modal.feed.picker.validSamples', {
                count: validSamples.length,
                total: matches.length,
              })
            }}
          </p>
          <ul class="space-y-2">
            <li
              v-for="match in matches.slice(0, 5)"
              :key="match.path"
              class="p-2 border border-border rounded break-all"
            >
              <p>{{ previewField(match, selection.title || '.', byPath).slice(0, 200) }}</p>
              <p v-if="selection.uri" class="text-xs text-text-secondary">
                {{ previewField(match, selection.uri, byPath) }}
              </p>
              <p v-if="selection.timestamp" class="text-xs text-text-secondary">
                {{ previewField(match, selection.timestamp, byPath) }}
              </p>
            </li>
          </ul>
        </aside>
      </div>
      <template #footer>
        <button type="button" class="btn-secondary" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          class="px-3 py-2 rounded bg-accent text-white disabled:opacity-40"
          :disabled="!validSamples.length"
          @click="emit('apply', { ...selection })"
        >
          {{ t('modal.feed.picker.apply') }}
        </button>
      </template>
    </BaseModal>
  </Teleport>
</template>
