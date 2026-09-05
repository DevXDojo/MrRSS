<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
const props = defineProps<{ feedId: number }>();
const { t } = useI18n();
const contentSelector = ref('');
const removeSelector = ref('');
const cookie = ref('');
const cookieOrigin = ref('');
const hasCookie = ref(false);
const clearCookie = ref(false);
const loading = ref(true);
const saving = ref(false);
watch(
  () => props.feedId,
  async (id, _, onCleanup) => {
    const controller = new AbortController();
    onCleanup(() => controller.abort());
    loading.value = true;
    try {
      const res = await fetch(`/api/feeds/content-options?id=${id}`, { signal: controller.signal });
      if (!res.ok) throw new Error('Content options');
      const options = await res.json();
      if (controller.signal.aborted) return;
      contentSelector.value = options.content_selector || '';
      removeSelector.value = options.remove_selector || '';
      cookieOrigin.value = options.cookie_origin || '';
      hasCookie.value = options.has_cookie === true;
      cookie.value = '';
      clearCookie.value = false;
      loading.value = false;
    } catch {
      if (!controller.signal.aborted)
        window.showToast(t('modal.feed.contentOptionsError'), 'error');
    }
  },
  { immediate: true }
);
async function save() {
  if (loading.value || saving.value) return;
  saving.value = true;
  try {
    const res = await fetch(`/api/feeds/content-options?id=${props.feedId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        content_selector: contentSelector.value,
        remove_selector: removeSelector.value,
        cookie_origin: cookieOrigin.value,
        ...(clearCookie.value ? { cookie: '' } : cookie.value ? { cookie: cookie.value } : {}),
      }),
    });
    if (!res.ok) throw new Error('Content options');
    const saved = await res.json();
    hasCookie.value = saved.has_cookie === true;
    cookie.value = '';
    clearCookie.value = false;
    window.showToast(t('modal.feed.contentOptionsSaved'), 'success');
  } catch {
    window.showToast(t('modal.feed.contentOptionsError'), 'error');
  } finally {
    saving.value = false;
  }
}
</script>
<template>
  <fieldset class="space-y-3 border-t border-border pt-4" :disabled="loading || saving">
    <legend class="text-sm font-medium text-text-primary">
      {{ t('modal.feed.contentOptions') }}
    </legend>
    <p class="text-xs text-text-secondary">{{ t('modal.feed.contentOptionsHelp') }}</p>
    <label class="block text-sm text-text-primary">
      {{ t('modal.feed.contentSelector') }}
      <input
        v-model="contentSelector"
        data-testid="content-selector"
        maxlength="4096"
        class="w-full mt-1 px-3 py-2 bg-bg-secondary border border-border rounded-lg"
        placeholder="article .content"
      />
    </label>
    <label class="block text-sm text-text-primary">
      {{ t('modal.feed.removeSelector') }}
      <input
        v-model="removeSelector"
        data-testid="remove-selector"
        maxlength="4096"
        class="w-full mt-1 px-3 py-2 bg-bg-secondary border border-border rounded-lg"
        placeholder=".advertisement, .related"
      />
    </label>
    <label class="block text-sm text-text-primary">
      {{ t('modal.feed.cookieOrigin') }}
      <input
        v-model="cookieOrigin"
        data-testid="cookie-origin"
        class="w-full mt-1 px-3 py-2 bg-bg-secondary border border-border rounded-lg"
        placeholder="https://example.com"
      />
    </label>
    <label class="block text-sm text-text-primary">
      {{ t('modal.feed.cookie') }}
      <input
        v-model="cookie"
        data-testid="feed-cookie"
        type="password"
        autocomplete="new-password"
        maxlength="8192"
        class="w-full mt-1 px-3 py-2 bg-bg-secondary border border-border rounded-lg"
        :placeholder="hasCookie ? t('modal.feed.cookieSaved') : 'name=value; name2=value2'"
      />
    </label>
    <p class="text-xs text-text-secondary">{{ t('modal.feed.cookieHelp') }}</p>
    <label v-if="hasCookie" class="flex items-center gap-2 text-sm"
      ><input v-model="clearCookie" type="checkbox" />{{ t('modal.feed.clearCookie') }}</label
    >
    <button
      type="button"
      class="px-3 py-2 bg-accent text-white rounded-lg disabled:opacity-50"
      @click="save"
    >
      {{ t('modal.feed.saveContentOptions') }}
    </button>
  </fieldset>
</template>
