<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import BaseModal from '@/components/common/BaseModal.vue';
import XPathPreviewNodeView from './XPathPreviewNode.vue';
import { flattenPreview, previewField, relativePickerXPath, type XPathPreviewNode, type XPathPickerField, type XPathSelection } from '@/utils/xpathPicker';

const props = defineProps<{ url: string; proxyEnabled?: boolean; proxyUrl?: string }>();
const emit = defineEmits<{ close: []; apply: [selection: XPathSelection] }>();
const { t } = useI18n();
const tree = ref<XPathPreviewNode | null>(null);
const loading = ref(false);
const failed = ref(false);
const candidate = ref<XPathPreviewNode | null>(null);
const item = ref<XPathPreviewNode | null>(null);
const active = ref<XPathPickerField>('item');
const emptySelection = (): XPathSelection => ({ item: '', title: '', uri: '', timestamp: '', content: '', thumbnail: '' });
const selection = ref(emptySelection());
const fields: XPathPickerField[] = ['item', 'title', 'uri', 'timestamp', 'content', 'thumbnail'];
const fieldLabel = (field: XPathPickerField) => t(`modal.feed.picker.${field === 'title' ? 'titleField' : field}`);
const nodes = computed(() => tree.value ? flattenPreview(tree.value).filter((node) => node.path && node.tag) : []);
const byPath = computed(() => new Map(nodes.value.map((node) => [node.path!, node])));
const matches = computed(() => nodes.value.filter((node) => node.group === selection.value.item));
const parent = computed(() => candidate.value?.path ? byPath.value.get(candidate.value.path.slice(0, candidate.value.path.lastIndexOf('/'))) : undefined);
const generated = computed(() => {
  if (!candidate.value) return null;
  if (active.value === 'item') return candidate.value.group ?? null;
  return item.value ? relativePickerXPath(item.value, candidate.value, active.value) : null;
});
let controller: AbortController | null = null;
let generation = 0;

async function load() {
  const request = ++generation;
  controller?.abort();
  controller = new AbortController();
  loading.value = true; failed.value = false; tree.value = null;
  item.value = null; candidate.value = null; selection.value = emptySelection(); active.value = 'item';
  try {
    const response = await fetch('/api/feeds/xpath-preview', { method: 'POST', headers: { 'Content-Type': 'application/json' }, signal: controller.signal, body: JSON.stringify({ url: props.url, proxy_enabled: props.proxyEnabled, proxy_url: props.proxyUrl }) });
    if (!response.ok) throw new Error('Preview failed');
    const data: XPathPreviewNode = await response.json();
    if (request === generation) tree.value = data;
  } catch {
    if (request === generation) { failed.value = true; window.showToast(t('modal.feed.picker.loadFailed'), 'error'); }
  } finally { if (request === generation) loading.value = false; }
}

function choose() {
  if (!candidate.value || !generated.value) return;
  if (active.value === 'item') {
    item.value = candidate.value;
    selection.value = { ...emptySelection(), item: generated.value };
    active.value = 'title';
  } else selection.value[active.value] = generated.value;
}

onMounted(load);
onBeforeUnmount(() => { generation++; controller?.abort(); });
</script>

<template>
  <BaseModal :title="t('modal.feed.picker.title')" size="full" height="full" :z-index="80" show-footer @close="emit('close')">
    <p class="text-sm text-text-secondary mb-3">{{ t('modal.feed.picker.hint') }}</p>
    <p class="text-xs text-text-secondary mb-3">{{ t('modal.feed.picker.staticHint') }}</p>
    <div class="flex flex-wrap gap-2 mb-3">
      <button v-for="field in fields" :key="field" type="button" class="px-3 py-2 rounded border border-border text-sm" :class="active === field ? 'bg-accent text-white' : 'bg-bg-secondary text-text-primary'" :disabled="field !== 'item' && !item" @click="active = field">{{ fieldLabel(field) }}</button>
    </div>
    <p v-if="loading" class="text-text-secondary">{{ t('common.loading') }}</p>
    <button v-else-if="failed" type="button" class="text-accent" @click="load">{{ t('modal.feed.picker.retry') }}</button>
    <div v-else-if="tree" class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <div class="lg:col-span-2 max-h-[55vh] overflow-auto p-4 border border-border rounded bg-bg-primary text-text-primary text-sm">
        <XPathPreviewNodeView :node="tree" :selected="candidate?.path" @pick="candidate = $event" />
      </div>
      <div class="space-y-3 text-sm min-w-0">
        <p class="break-all text-text-secondary">{{ candidate?.path || t('modal.feed.picker.selectElement') }}</p>
        <div class="flex gap-2">
          <button type="button" class="px-3 py-2 rounded border border-border disabled:opacity-50" :disabled="!parent" @click="candidate = parent ?? null">{{ t('modal.feed.picker.parent') }}</button>
          <button type="button" class="px-3 py-2 rounded bg-accent text-white disabled:opacity-50" :disabled="!generated" @click="choose">{{ t('modal.feed.picker.choose') }}</button>
        </div>
        <p v-if="item" class="text-text-secondary">{{ t('modal.feed.picker.withinItem') }}</p>
        <dl class="space-y-2">
          <template v-for="field in fields" :key="field"><dt class="font-medium">{{ fieldLabel(field) }}</dt><dd class="break-all text-text-secondary">{{ selection[field] || '—' }}</dd></template>
        </dl>
        <p>{{ t('modal.feed.picker.matches', { count: matches.length }) }}</p>
        <ul class="space-y-2 max-h-40 overflow-auto">
          <li v-for="match in matches.slice(0, 5)" :key="match.path" class="p-2 border border-border rounded break-all">
            <p>{{ previewField(match, selection.title || '.', byPath) }}</p>
            <p class="text-xs text-text-secondary">{{ selection.uri ? previewField(match, selection.uri, byPath) : '' }}</p>
          </li>
        </ul>
      </div>
    </div>
    <template #footer>
      <button type="button" class="px-4 py-2 border border-border rounded" @click="emit('close')">{{ t('common.cancel') }}</button>
      <button type="button" class="px-4 py-2 bg-accent text-white rounded disabled:opacity-50" :disabled="!selection.item || !selection.title || !selection.uri || matches.length === 0" @click="emit('apply', selection)">{{ t('modal.feed.picker.apply') }}</button>
    </template>
  </BaseModal>
</template>
