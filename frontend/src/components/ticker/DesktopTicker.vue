<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { Call } from '@wailsio/runtime';
import { useI18n } from 'vue-i18n';
import { PhPause, PhPlay, PhGear, PhX, PhArrowUp, PhArrowDown } from '@phosphor-icons/vue';
import { useSettings } from '@/composables/core/useSettings';
import { openInBrowser } from '@/utils/browser';
import type { Article, Feed } from '@/types/models';

const { t } = useI18n();
const { settings } = useSettings();
const feeds = ref<Feed[]>([]);
const articles = ref<Article[]>([]);
const paused = ref(window.matchMedia('(prefers-reduced-motion: reduce)').matches);
const configuring = ref(false);
const error = ref('');
const notice = ref('');
const previousToast = window.showToast;
const tickerToast: NonNullable<Window['showToast']> = (message, type) => {
  if (type === 'error') error.value = message;
  else notice.value = message;
};
window.showToast = tickerToast;
const group = ref<HTMLElement | null>(null);
const width = ref(1000);
const duration = computed(() => `${width.value / Math.min(200, Math.max(10, settings.value.ticker_speed))}s`);
const color = (value: string, fallback: string) => /^#[0-9a-f]{6}$/i.test(value) ? value : fallback;
const style = computed(() => ({
  color: color(settings.value.ticker_text_color, '#f8fafc'),
  background: `linear-gradient(90deg, ${color(settings.value.ticker_background_color, '#1e293b')}, ${color(settings.value.ticker_background_color_end, '#0f172a')})`,
  fontSize: `${Math.min(40, Math.max(12, settings.value.ticker_font_size))}px`,
}));
let poll: ReturnType<typeof setInterval> | undefined;
let saveTimer: ReturnType<typeof setTimeout> | undefined;
let controller: AbortController | undefined;
const lifetime = new AbortController();
let observer: ResizeObserver | undefined;
let disposed = false;

async function load() {
  controller?.abort();
  const current = new AbortController();
  controller = current;
  try {
    const params = new URLSearchParams({ filter: 'all', limit: '50' });
    if (settings.value.ticker_feed_id > 0) params.set('feed_id', String(settings.value.ticker_feed_id));
    const res = await fetch(`/api/articles?${params}`, { signal: current.signal });
    if (!res.ok) throw new Error();
    const data: Article[] | null = await res.json();
    if (current.signal.aborted || disposed) return;
    articles.value = data ?? [];
    error.value = '';
    await nextTick();
    if (group.value) width.value = group.value.scrollWidth;
  } catch {
    if (!current.signal.aborted && !disposed) error.value = t('setting.ticker.loadFailed');
  }
}

const preferencePayload = () => ({
  ticker_feed_id: String(settings.value.ticker_feed_id),
  ticker_speed: String(settings.value.ticker_speed),
  ticker_font_size: String(settings.value.ticker_font_size),
  ticker_text_color: settings.value.ticker_text_color,
  ticker_background_color: settings.value.ticker_background_color,
  ticker_background_color_end: settings.value.ticker_background_color_end,
});
async function savePreferences() {
  try {
    const res = await fetch('/api/settings', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(preferencePayload()), signal: lifetime.signal });
    if (!res.ok) throw new Error();
    error.value = '';
  } catch { if (!disposed) error.value = t('setting.ticker.saveFailed'); }
}
watch(preferencePayload, () => {
  clearTimeout(saveTimer);
  saveTimer = setTimeout(savePreferences, 500);
});
watch(() => settings.value.ticker_feed_id, load);
watch(configuring, (expanded) => nativeAction('SetTickerExpanded', expanded));
watch(group, (element) => { observer?.disconnect(); if (element) observer?.observe(element); });

async function nativeAction(action: 'HideTicker' | 'PositionTicker' | 'SetTickerExpanded', top?: boolean) {
  try {
    await Call.ByName(`main.DesktopCapabilities.${action}`, ...(top === undefined ? [] : [top]));
  } catch { error.value = t('setting.ticker.openFailed'); }
}

onMounted(async () => {
  observer = new ResizeObserver(() => { if (group.value) width.value = group.value.scrollWidth; });
  if (group.value) observer.observe(group.value);
  try {
    const res = await fetch('/api/feeds', { signal: lifetime.signal });
    if (!res.ok) throw new Error();
    feeds.value = (await res.json()) ?? [];
  } catch { if (!disposed) error.value = t('setting.ticker.loadFailed'); }
  if (disposed) return;
  await load();
  poll = setInterval(load, 60000);
});
onUnmounted(() => {
  disposed = true;
  clearInterval(poll);
  clearTimeout(saveTimer);
  controller?.abort();
  lifetime.abort();
  observer?.disconnect();
  if (window.showToast === tickerToast) window.showToast = previousToast;
});
</script>

<template>
  <main class="ticker flex h-screen min-w-0 flex-col overflow-hidden" :style="style">
    <div class="flex min-h-8 shrink-0 items-center gap-2 px-2 text-xs">
      <span class="ticker-grip cursor-move font-semibold" :title="t('setting.ticker.drag')">MrRSS</span>
      <select v-model.number="settings.ticker_feed_id" class="min-w-0 max-w-60 rounded bg-bg-secondary px-1 text-text-primary" :aria-label="t('setting.ticker.feed')">
        <option :value="0">{{ t('sidebar.activity.allArticles') }}</option>
        <option v-for="feed in feeds" :key="feed.id" :value="feed.id">{{ feed.title }}</option>
      </select>
      <button :title="t(paused ? 'setting.ticker.play' : 'setting.ticker.pause')" :aria-label="t(paused ? 'setting.ticker.play' : 'setting.ticker.pause')" @click="paused = !paused"><PhPlay v-if="paused" :size="16" /><PhPause v-else :size="16" /></button>
      <button :title="t('setting.ticker.top')" :aria-label="t('setting.ticker.top')" @click="nativeAction('PositionTicker', true)"><PhArrowUp :size="16" /></button>
      <button :title="t('setting.ticker.bottom')" :aria-label="t('setting.ticker.bottom')" @click="nativeAction('PositionTicker', false)"><PhArrowDown :size="16" /></button>
      <button :title="t('setting.ticker.preferences')" :aria-label="t('setting.ticker.preferences')" :aria-expanded="configuring" @click="configuring = !configuring"><PhGear :size="16" /></button>
      <span class="ticker-grip min-w-2 flex-1 self-stretch"></span>
      <button :title="t('common.close')" :aria-label="t('common.close')" @click="nativeAction('HideTicker')"><PhX :size="16" /></button>
    </div>
    <div v-if="configuring" class="flex shrink-0 flex-wrap items-center gap-3 px-2 text-xs">
      <label>{{ t('setting.ticker.speed') }} <input v-model.number="settings.ticker_speed" type="range" min="10" max="200" /></label>
      <label>{{ t('setting.ticker.fontSize') }} <input v-model.number="settings.ticker_font_size" type="number" min="12" max="40" class="w-12 bg-bg-secondary text-text-primary" /></label>
      <label>{{ t('setting.ticker.textColor') }} <input v-model="settings.ticker_text_color" type="color" /></label>
      <label>{{ t('setting.ticker.background') }} <input v-model="settings.ticker_background_color" type="color" /><input v-model="settings.ticker_background_color_end" type="color" :aria-label="t('setting.ticker.backgroundEnd')" /></label>
    </div>
    <p v-if="error" role="alert" class="m-0 px-2 text-xs">{{ error }} <button class="underline" @click="() => { savePreferences(); load(); }">{{ t('setting.ticker.retry') }}</button></p>
    <p v-else-if="notice" role="status" class="m-0 px-2 text-xs">{{ notice }}</p>
    <div v-if="articles.length" class="min-h-0 flex-1 overflow-hidden" :aria-label="t('setting.ticker.title')">
      <div class="ticker-track flex w-max items-center h-full" :class="{ paused }" :style="{ animationDuration: duration }">
        <div v-for="copy in 2" :key="copy" :ref="(el) => { if (copy === 1) group = el as HTMLElement; }" class="ticker-group flex shrink-0 items-center gap-8 pr-8" :aria-hidden="copy === 2 ? 'true' : undefined">
          <button v-for="article in articles" :key="article.id" :tabindex="copy === 2 ? -1 : 0" class="whitespace-nowrap hover:underline" @click="openInBrowser(article.url)">
            <span class="opacity-65">{{ article.feed_title }} · </span>{{ article.title }}
          </button>
        </div>
      </div>
    </div>
    <p v-else class="m-0 px-2 text-sm">{{ t('setting.ticker.empty') }}</p>
  </main>
</template>

<style scoped>
.ticker-grip { --wails-draggable: drag; }
.ticker-track { animation: ticker-scroll linear infinite; }
.ticker-track.paused, .ticker-track:hover, .ticker-track:focus-within { animation-play-state: paused; }
.ticker input[type='color'] { width: 1.5rem; height: 1.25rem; vertical-align: middle; }
.ticker input[type='range'] { width: 5rem; vertical-align: middle; }
@keyframes ticker-scroll { to { transform: translateX(-50%); } }
</style>
