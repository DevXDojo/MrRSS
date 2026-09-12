<script setup lang="ts">
import ModalFooter from '@/components/common/ModalFooter.vue';
import { onBeforeUnmount, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhArrowUp, PhArrowDown } from '@phosphor-icons/vue';
import BaseModal from '@/components/common/BaseModal.vue';
import { setSettingsFromRawData, useSettings } from '@/composables/core/useSettings';
import { parseToolbarLayout, toolbarActions, type ToolbarActionID } from '@/utils/articleToolbar';

const emit = defineEmits<{ close: [] }>();
const { t } = useI18n();
const { settings } = useSettings();
const items = ref(parseToolbarLayout(settings.value.article_toolbar_layout));
const saving = ref(false);
let controller: AbortController | null = null;
onBeforeUnmount(() => controller?.abort());

function label(id: ToolbarActionID) {
  return t(toolbarActions.find((action) => action.id === id)!.label);
}

function move(index: number, direction: number) {
  const target = index + direction;
  if (target < 0 || target >= items.value.length) return;
  [items.value[index], items.value[target]] = [items.value[target], items.value[index]];
}

async function save() {
  if (saving.value) return;
  saving.value = true;
  controller = new AbortController();
  try {
    const response = await fetch('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ article_toolbar_layout: JSON.stringify(items.value) }),
      signal: controller.signal,
    });
    if (!response.ok) throw new Error('Unable to save toolbar settings');
    const data: Record<string, string> = await response.json();
    if (controller.signal.aborted) return;
    setSettingsFromRawData(data);
    window.dispatchEvent(new CustomEvent('settings-updated', { detail: { autoSave: true } }));
    emit('close');
  } catch {
    if (!controller.signal.aborted) window.showToast(t('article.toolbar.saveFailed'), 'error');
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <BaseModal
    :title="t('article.toolbar.customize')"
    size="md"
    :z-index="80"
    :loading="saving"
    @close="emit('close')"
  >
    <div class="p-4 space-y-3">
      <p class="text-sm text-text-secondary">{{ t('article.toolbar.customizeHint') }}</p>
      <div
        v-for="(item, index) in items"
        :key="item.id"
        class="flex items-center gap-2 rounded-lg border border-border p-2"
      >
        <label class="flex flex-1 items-center gap-2 text-sm text-text-primary cursor-pointer">
          <input v-model="item.visible" type="checkbox" :disabled="saving" class="accent-accent" />
          {{ label(item.id) }}
        </label>
        <button
          class="p-1 rounded hover:bg-bg-tertiary disabled:opacity-30"
          :disabled="saving || index === 0"
          :aria-label="t('article.toolbar.moveUp', { name: label(item.id) })"
          @click="move(index, -1)"
        >
          <PhArrowUp :size="18" />
        </button>
        <button
          class="p-1 rounded hover:bg-bg-tertiary disabled:opacity-30"
          :disabled="saving || index === items.length - 1"
          :aria-label="t('article.toolbar.moveDown', { name: label(item.id) })"
          @click="move(index, 1)"
        >
          <PhArrowDown :size="18" />
        </button>
      </div>
    </div>
    <template #footer>
      <ModalFooter
        :primary-button="{ label: t('common.save'), disabled: saving, loading: saving }"
        :secondary-button="{ label: t('common.cancel'), disabled: saving }"
        @primary-click="save"
        @secondary-click="emit('close')"
      >
        <template #left>
          <ModalFooter
            class="sm:mr-auto"
            :secondary-button="{ label: t('article.toolbar.reset'), disabled: saving }"
            @secondary-click="items = parseToolbarLayout('[]')"
          />
        </template>
      </ModalFooter>
    </template>
  </BaseModal>
</template>
