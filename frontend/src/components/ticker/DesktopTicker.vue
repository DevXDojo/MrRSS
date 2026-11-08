<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watchEffect } from 'vue';
import { Call, Events } from '@wailsio/runtime';
import { useI18n } from 'vue-i18n';
import {
  PhRss,
  PhPause,
  PhPlay,
  PhSlidersHorizontal,
  PhX,
  PhCaretLeft,
  PhCaretRight,
  PhArrowClockwise,
  PhPushPin,
  PhWarningCircle,
  PhSpinnerGap,
} from '@phosphor-icons/vue';
import { useSettings } from '@/composables/core/useSettings';
import { openInBrowser } from '@/utils/browser';
import type { Article } from '@/types/models';
import { tickerStyle } from './tickerPreferences';
import { useTickerPreferences } from './useTickerPreferences';
import { useTickerArticles } from './useTickerArticles';
import { useTickerMarquee } from './useTickerMarquee';

const { t, locale } = useI18n();
const { settings } = useSettings();
const preferences = useTickerPreferences(settings);
const nativeVisible = ref(true);
const pageVisible = ref(!document.hidden);
const visible = computed(() => nativeVisible.value && pageVisible.value);
const { articles, loading, failed, refresh } = useTickerArticles(settings, visible);
const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)');
const paused = ref(reducedMotion.matches);
const hovering = ref(false);
const focused = ref(false);
const stopped = computed(
  () =>
    paused.value ||
    !visible.value ||
    focused.value ||
    (hovering.value && settings.value.ticker_pause_on_hover)
);
const viewport = ref<HTMLElement | null>(null);
const group = ref<HTMLElement | null>(null);
const track = ref<HTMLElement | null>(null);
const { copies, step } = useTickerMarquee(viewport, group, track, {
  speed: () => settings.value.ticker_speed,
  paused: () => stopped.value,
  content: () => [
    articles.value,
    settings.value.ticker_font_size,
    settings.value.ticker_show_source,
    settings.value.ticker_show_time,
    settings.value.translation_enabled,
  ],
});
const style = computed(() => tickerStyle(settings.value));
const previousBackground = document.body.style.backgroundColor;
watchEffect(() => {
  document.body.style.backgroundColor = String(style.value['--ticker-end']);
});
const message = ref('');
const messageIsError = ref(false);
let messageTimer: ReturnType<typeof setTimeout> | undefined;
const previousToast = window.showToast;
const tickerToast: Window['showToast'] = (text, type) => {
  message.value = text;
  messageIsError.value = type === 'error';
  clearTimeout(messageTimer);
  messageTimer = setTimeout(
    () => {
      message.value = '';
    },
    type === 'error' ? 6000 : 2600
  );
};
window.showToast = tickerToast;
const disposers: Array<() => void> = [];

async function nativeAction(action: string, ...args: Array<number | boolean>) {
  try {
    await Call.ByName(`main.DesktopCapabilities.${action}`, ...args);
  } catch {
    tickerToast(t('setting.ticker.openFailed'), 'error');
  }
}
function openPreferences() {
  void nativeAction('ShowTickerSettings');
}
async function disableTicker() {
  settings.value.ticker_enabled = false;
  if (!(await preferences.saveNow())) tickerToast(t('setting.ticker.saveFailed'), 'error');
}
function navigate(direction: number) {
  paused.value = true;
  step(direction);
}
function headline(article: Article) {
  return settings.value.translation_enabled && article.translated_title
    ? article.translated_title
    : article.title;
}
function articleTime(article: Article) {
  const date = new Date(article.published_at);
  return Number.isNaN(date.getTime())
    ? ''
    : date.toLocaleTimeString(locale.value, { hour: '2-digit', minute: '2-digit' });
}
function articleTooltip(article: Article) {
  const date = new Date(article.published_at);
  return [
    article.title,
    article.feed_title,
    Number.isNaN(date.getTime()) ? '' : date.toLocaleString(locale.value),
    t(
      settings.value.ticker_open_in_browser
        ? 'setting.ticker.browserHint'
        : 'setting.ticker.readerHint'
    ),
  ]
    .filter(Boolean)
    .join('\n');
}
function openArticle(article: Article, event: MouseEvent) {
  if (settings.value.ticker_open_in_browser || event.ctrlKey || event.metaKey)
    void openInBrowser(article.url);
  else void nativeAction('OpenTickerArticle', article.id, article.feed_id);
}
function onKeydown(event: KeyboardEvent) {
  const target = event.target instanceof Element ? event.target : null;
  if (target?.closest('input,select,textarea,[contenteditable="true"]')) return;
  if (event.key === 'Escape') {
    event.preventDefault();
    void disableTicker();
  } else if ((event.ctrlKey || event.metaKey) && event.key === ',') {
    event.preventDefault();
    openPreferences();
  } else if (event.key === 'ArrowRight' || event.key === 'ArrowLeft') {
    event.preventDefault();
    navigate(event.key === 'ArrowRight' ? 1 : -1);
  } else if (event.code === 'Space' && !target?.closest('button')) {
    event.preventDefault();
    paused.value = !paused.value;
  }
}
function visibilityChanged() {
  pageVisible.value = !document.hidden;
}
function onBlur() {
  focused.value = false;
  hovering.value = false;
}
function motionChanged(event: MediaQueryListEvent) {
  if (event.matches) paused.value = true;
}

onMounted(() => {
  document.title = t('setting.ticker.title');
  disposers.push(
    Events.On('ticker:visibility', (event) => {
      nativeVisible.value = event.data === true;
      if (!nativeVisible.value) onBlur();
      else void preferences.reload();
    })
  );
  disposers.push(
    Events.On('ticker:request-disable', () => {
      void disableTicker();
    })
  );
  document.addEventListener('visibilitychange', visibilityChanged);
  document.addEventListener('keydown', onKeydown);
  window.addEventListener('blur', onBlur);
  reducedMotion.addEventListener?.('change', motionChanged);
});
onUnmounted(() => {
  document.body.style.backgroundColor = previousBackground;
  clearTimeout(messageTimer);
  disposers.forEach((dispose) => dispose());
  document.removeEventListener('visibilitychange', visibilityChanged);
  document.removeEventListener('keydown', onKeydown);
  window.removeEventListener('blur', onBlur);
  reducedMotion.removeEventListener?.('change', motionChanged);
  if (window.showToast === tickerToast) window.showToast = previousToast;
});
</script>

<template>
  <main
    class="ticker-strip"
    :class="{ 'is-floating': settings.ticker_position === 'floating', 'is-paused': paused }"
    :style="style"
    @mouseenter="hovering = true"
    @mouseleave="hovering = false"
    @focusin="focused = true"
    @focusout="focused = false"
  >
    <div
      class="ticker-grip"
      :title="
        t(
          settings.ticker_position === 'floating'
            ? 'setting.ticker.drag'
            : 'setting.ticker.dockedHint'
        )
      "
    >
      <PhRss :size="15" weight="bold" /><span class="ticker-brand">MrRSS</span>
      <span
        class="ticker-status-dot"
        :class="{ paused: paused || (hovering && settings.ticker_pause_on_hover) }"
        :title="t(paused ? 'setting.ticker.paused' : 'setting.ticker.live')"
      ></span>
    </div>
    <div ref="viewport" class="ticker-viewport" :aria-label="t('setting.ticker.title')">
      <div v-if="articles.length" ref="track" class="ticker-track">
        <div
          v-for="copy in copies"
          :key="copy"
          :ref="
            (element) => {
              if (copy === 1) group = element as HTMLElement;
            }
          "
          class="ticker-group"
          :aria-hidden="copy > 1 ? 'true' : undefined"
        >
          <button
            v-for="article in articles"
            :key="article.id"
            class="ticker-headline"
            :class="{ 'is-read': article.is_read }"
            :data-ticker-article="article.id"
            :tabindex="copy > 1 ? -1 : 0"
            :title="articleTooltip(article)"
            @click="openArticle(article, $event)"
            @auxclick.middle.prevent="openInBrowser(article.url)"
          >
            <span v-if="settings.ticker_show_source && article.feed_title" class="ticker-source">{{
              article.feed_title
            }}</span>
            <span
              v-if="!article.is_read"
              class="ticker-unread-dot"
              :aria-label="t('setting.ticker.unread')"
            ></span>
            <span>{{ headline(article) }}</span>
            <time
              v-if="settings.ticker_show_time"
              class="ticker-time"
              :datetime="article.published_at"
              >{{ articleTime(article) }}</time
            >
            <span class="ticker-separator" aria-hidden="true">/</span>
          </button>
        </div>
      </div>
      <div v-else class="ticker-empty">
        <PhSpinnerGap v-if="loading" :size="16" class="animate-spin" />
        <span>{{
          t(
            loading
              ? 'setting.ticker.loading'
              : failed
                ? 'setting.ticker.loadFailed'
                : settings.ticker_only_unread
                  ? 'setting.ticker.noUnread'
                  : 'setting.ticker.empty'
          )
        }}</span>
        <button v-if="!loading" @click="failed ? refresh() : openPreferences()">
          {{ t(failed ? 'setting.ticker.retry' : 'setting.ticker.chooseFeed') }}
        </button>
      </div>
    </div>
    <button
      v-if="failed && articles.length"
      class="ticker-warning"
      :title="t('setting.ticker.loadFailed')"
      @click="refresh"
    >
      <PhWarningCircle :size="15" />{{ t('setting.ticker.retry') }}
    </button>
    <div
      v-if="message"
      class="ticker-message"
      :class="{ error: messageIsError }"
      :role="messageIsError ? 'alert' : 'status'"
      :title="message"
    >
      {{ message }}
    </div>
    <div class="ticker-rest-state" aria-hidden="true">
      <PhPause v-if="paused" :size="13" /><span>{{ articles.length }}</span>
    </div>
    <nav class="ticker-controls" :aria-label="t('setting.ticker.controls')">
      <button
        class="ticker-step"
        :title="t('setting.ticker.previous')"
        :aria-label="t('setting.ticker.previous')"
        :disabled="!articles.length"
        @click="navigate(-1)"
      >
        <PhCaretLeft :size="16" />
      </button>
      <button
        :title="t(paused ? 'setting.ticker.play' : 'setting.ticker.pause')"
        :aria-label="t(paused ? 'setting.ticker.play' : 'setting.ticker.pause')"
        :aria-pressed="paused"
        @click="paused = !paused"
      >
        <PhPlay v-if="paused" :size="15" weight="fill" /><PhPause v-else :size="15" weight="fill" />
      </button>
      <button
        class="ticker-step"
        :title="t('setting.ticker.next')"
        :aria-label="t('setting.ticker.next')"
        :disabled="!articles.length"
        @click="navigate(1)"
      >
        <PhCaretRight :size="16" />
      </button>
      <span class="ticker-control-divider"></span>
      <button
        :title="t('setting.ticker.refresh')"
        :aria-label="t('setting.ticker.refresh')"
        :disabled="loading"
        @click="refresh"
      >
        <PhArrowClockwise :size="15" :class="{ 'animate-spin': loading }" />
      </button>
      <button
        class="ticker-pin"
        :class="{ active: settings.ticker_always_on_top }"
        :title="t('setting.ticker.alwaysOnTop')"
        :aria-label="t('setting.ticker.alwaysOnTop')"
        :aria-pressed="settings.ticker_always_on_top"
        @click="settings.ticker_always_on_top = !settings.ticker_always_on_top"
      >
        <PhPushPin :size="15" :weight="settings.ticker_always_on_top ? 'fill' : 'regular'" />
      </button>
      <button
        :title="t('setting.ticker.preferences')"
        :aria-label="t('setting.ticker.preferences')"
        @click="openPreferences"
      >
        <PhSlidersHorizontal :size="16" />
      </button>
      <button
        :title="t('setting.ticker.disable')"
        :aria-label="t('setting.ticker.disable')"
        @click="disableTicker"
      >
        <PhX :size="15" />
      </button>
    </nav>
    <button
      v-if="preferences.status.value === 'error'"
      class="ticker-save-error"
      role="alert"
      @click="preferences.saveNow"
    >
      {{ t('setting.ticker.saveFailed') }} · {{ t('setting.ticker.retry') }}
    </button>
  </main>
</template>

<style scoped>
.ticker-strip {
  position: relative;
  display: flex;
  align-items: center;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: linear-gradient(110deg, var(--ticker-start), var(--ticker-end));
  color: var(--ticker-ink);
  font-size: var(--ticker-size);
  border-block: 1px solid color-mix(in srgb, var(--ticker-chrome) 12%, transparent);
  isolation: isolate;
}
.ticker-strip.is-floating {
  border: 1px solid color-mix(in srgb, var(--ticker-chrome) 18%, transparent);
  border-radius: 10px;
}
.ticker-grip {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  gap: 7px;
  padding: 0 15px;
  height: 100%;
  color: var(--ticker-chrome);
  cursor: default;
  user-select: none;
}
.is-floating .ticker-grip {
  --wails-draggable: drag;
  cursor: grab;
}
.ticker-brand {
  font-size: 10px;
  font-weight: 750;
  letter-spacing: 0.03em;
  opacity: 0.7;
}
.ticker-status-dot,
.ticker-unread-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  flex-shrink: 0;
  background: currentColor;
}
.ticker-status-dot {
  width: 5px;
  height: 5px;
  background: #4ade80;
  box-shadow: 0 0 7px #4ade8038;
}
.ticker-status-dot.paused {
  background: currentColor;
  box-shadow: none;
  opacity: 0.4;
}
.ticker-viewport {
  height: 100%;
  min-width: 0;
  flex: 1;
  overflow: hidden;
  mask-image: linear-gradient(90deg, transparent, #000 14px, #000 calc(100% - 42px), transparent);
}
.ticker-track {
  display: flex;
  width: max-content;
  align-items: center;
  height: 100%;
  will-change: transform;
}
.ticker-group {
  position: relative;
  display: flex;
  flex-shrink: 0;
  align-items: center;
  height: 100%;
}
.ticker-headline {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  gap: 10px;
  height: 100%;
  padding-left: 16px;
  white-space: nowrap;
  cursor: pointer;
  line-height: 1.2;
  text-align: left;
}
.ticker-headline:hover {
  text-decoration: underline;
  text-underline-offset: 4px;
}
.ticker-headline:focus-visible {
  outline: 1px solid currentColor;
  outline-offset: -6px;
  border-radius: 8px;
}
.ticker-headline.is-read {
  opacity: 0.65;
}
.ticker-source {
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 10px;
  font-weight: 650;
  letter-spacing: 0.02em;
  padding: 3px 6px;
  border-radius: 4px;
  color: var(--ticker-chrome);
  background: color-mix(in srgb, var(--ticker-chrome) 8%, transparent);
  text-decoration: none;
}
.ticker-time {
  color: var(--ticker-chrome);
  opacity: 0.5;
  font-size: 10px;
  font-variant-numeric: tabular-nums;
}
.ticker-separator {
  margin: 0 24px 0 20px;
  opacity: 0.18;
  font-size: 16px;
  font-weight: 300;
}
.ticker-rest-state {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 0 13px 0 4px;
  font-size: 10px;
  font-variant-numeric: tabular-nums;
  color: var(--ticker-chrome);
  opacity: 0.35;
}
.ticker-controls {
  position: absolute;
  z-index: 2;
  right: 0;
  top: 0;
  display: flex;
  align-items: center;
  gap: 2px;
  height: 100%;
  padding: 0 9px 0 24px;
  color: var(--ticker-chrome);
  background: linear-gradient(90deg, transparent, var(--ticker-end) 20px);
  opacity: 0;
  pointer-events: none;
  transition: opacity 150ms ease;
}
.ticker-strip:hover .ticker-controls,
.ticker-strip:focus-within .ticker-controls {
  opacity: 1;
  pointer-events: auto;
}
.ticker-controls button {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  opacity: 0.7;
  cursor: pointer;
}
.ticker-controls button:hover,
.ticker-controls button:focus-visible {
  opacity: 1;
  background: color-mix(in srgb, var(--ticker-chrome) 12%, transparent);
  outline: 1px solid color-mix(in srgb, var(--ticker-chrome) 22%, transparent);
}
.ticker-controls button:disabled {
  opacity: 0.25;
  cursor: default;
}
.ticker-controls .active {
  opacity: 1;
}
.ticker-control-divider {
  height: 13px;
  width: 1px;
  margin-inline: 4px;
  background: currentColor;
  opacity: 0.15;
}
.ticker-empty {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 100%;
  font-size: 12px;
  opacity: 0.7;
  padding-inline: 14px;
}
.ticker-empty button {
  text-decoration: underline;
  text-underline-offset: 3px;
}
.ticker-message,
.ticker-warning,
.ticker-save-error {
  position: absolute;
  z-index: 3;
  right: 48px;
  top: 50%;
  transform: translateY(-50%);
  max-width: 65%;
  border-radius: 5px;
  background: var(--ticker-end);
  color: var(--ticker-chrome);
  padding: 5px 9px;
  font-size: 11px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  box-shadow: 0 0 16px var(--ticker-end);
}
.ticker-warning {
  display: flex;
  align-items: center;
  gap: 5px;
}
.ticker-message.error,
.ticker-warning,
.ticker-save-error {
  color: color-mix(in srgb, #f59e0b 75%, var(--ticker-chrome));
}
.ticker-strip:hover .ticker-warning {
  right: 260px;
}
.ticker-save-error {
  right: 260px;
  cursor: pointer;
}
@media (max-width: 600px) {
  .ticker-brand,
  .ticker-step,
  .ticker-pin {
    display: none !important;
  }
  .ticker-grip {
    padding-inline: 11px;
  }
  .ticker-source {
    max-width: 100px;
  }
  .ticker-strip:hover .ticker-warning,
  .ticker-save-error {
    right: 170px;
  }
}
@media (prefers-reduced-motion: reduce) {
  .ticker-controls {
    transition: none;
  }
}
</style>
