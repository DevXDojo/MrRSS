<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watchEffect } from 'vue';
import { Call, Events } from '@wailsio/runtime';
import { useI18n } from 'vue-i18n';
import {
  PhRss,
  PhMagnifyingGlass,
  PhCheck,
  PhPalette,
  PhSlidersHorizontal,
  PhArrowUp,
  PhArrowDown,
  PhCornersOut,
  PhKeyboard,
  PhSpinnerGap,
  PhWarningCircle,
} from '@phosphor-icons/vue';
import { ToggleControl } from '@/components/settings';
import { useSettings } from '@/composables/core/useSettings';
import { resolveFontFamily } from '@/utils/fontDetector';
import type { Feed } from '@/types/models';
import { tickerPalettes, tickerStyle } from './tickerPreferences';
import { useTickerPreferences } from './useTickerPreferences';

const { t } = useI18n();
const { settings } = useSettings();
const { status, saveNow } = useTickerPreferences(settings);
const feeds = ref<Feed[]>([]);
const search = ref('');
const feedError = ref(false);
const loadingFeeds = ref(false);
const displays = ref<Array<{ id: string; name: string; primary: boolean }>>([]);
const themeQuery = window.matchMedia('(prefers-color-scheme: dark)');
const systemDark = ref(themeQuery.matches);
const filteredFeeds = computed(() =>
  feeds.value.filter((feed) =>
    `${feed.title} ${feed.category || ''}`
      .toLocaleLowerCase()
      .includes(search.value.trim().toLocaleLowerCase())
  )
);
const sourceName = computed(
  () =>
    feeds.value.find((feed) => feed.id === settings.value.ticker_feed_id)?.title ||
    t('setting.ticker.allFeeds')
);
const previewStyle = computed(() => tickerStyle(settings.value));
const activePalette = computed(
  () =>
    tickerPalettes.find(
      (palette) =>
        palette.start === settings.value.ticker_background_color &&
        palette.end === settings.value.ticker_background_color_end &&
        palette.text === settings.value.ticker_text_color
    )?.id
);
const positions = [
  { id: 'top', icon: PhArrowUp },
  { id: 'bottom', icon: PhArrowDown },
  { id: 'floating', icon: PhCornersOut },
] as const;
const behavior = [
  { key: 'ticker_pause_on_hover', label: 'pauseOnHover', description: 'pauseOnHoverHint' },
  { key: 'ticker_show_source', label: 'showSource', description: 'showSourceHint' },
  { key: 'ticker_show_time', label: 'showTime', description: 'showTimeHint' },
  { key: 'ticker_always_on_top', label: 'alwaysOnTop', description: 'alwaysOnTopHint' },
] as const;
let feedController: AbortController | undefined;
let disposed = false;
let disposeShown: (() => void) | undefined;

watchEffect(() => {
  document.documentElement.classList.toggle(
    'dark-mode',
    settings.value.theme === 'dark' || (settings.value.theme === 'auto' && systemDark.value)
  );
  document.documentElement.style.setProperty(
    '--ui-font-family',
    resolveFontFamily(settings.value.ui_font_family)
  );
  document.title = t('setting.ticker.preferences');
});
function themeChanged(event: MediaQueryListEvent) {
  systemDark.value = event.matches;
}
function choosePalette(palette: (typeof tickerPalettes)[number]) {
  settings.value.ticker_background_color = palette.start;
  settings.value.ticker_background_color_end = palette.end;
  settings.value.ticker_text_color = palette.text;
}
async function loadFeeds() {
  feedController?.abort();
  const current = new AbortController();
  feedController = current;
  loadingFeeds.value = true;
  try {
    const response = await fetch('/api/feeds', { signal: current.signal });
    if (!response.ok) throw new Error('Could not load subscriptions');
    const result: Feed[] | null = await response.json();
    if (result !== null && !Array.isArray(result)) throw new Error('Invalid subscriptions');
    if (disposed || current.signal.aborted) return;
    feeds.value = (result ?? []).sort((a, b) => a.title.localeCompare(b.title));
    feedError.value = false;
  } catch {
    if (!current.signal.aborted && !disposed) feedError.value = true;
  } finally {
    if (!current.signal.aborted && !disposed) loadingFeeds.value = false;
  }
}
async function finish() {
  if (!(await saveNow())) return;
  try {
    await Call.ByName('main.DesktopCapabilities.HideTickerSettings');
  } catch {
    status.value = 'error';
  }
}
onMounted(async () => {
  themeQuery.addEventListener?.('change', themeChanged);
  disposeShown = Events.On('ticker:preferences-shown', () => {
    void loadFeeds();
  });
  void loadFeeds();
  try {
    displays.value =
      ((await Call.ByName(
        'main.DesktopCapabilities.GetTickerDisplays'
      )) as typeof displays.value) ?? [];
  } catch {
    /* Display placement still works on the current screen. */
  }
});
onUnmounted(() => {
  disposed = true;
  feedController?.abort();
  disposeShown?.();
  themeQuery.removeEventListener?.('change', themeChanged);
});
</script>

<template>
  <main class="ticker-preferences">
    <header class="preferences-header">
      <div class="preferences-icon"><PhRss :size="20" weight="bold" /></div>
      <div class="min-w-0 flex-1">
        <h1>{{ t('setting.ticker.title') }}</h1>
        <p>{{ t('setting.ticker.settingsHint') }}</p>
      </div>
      <ToggleControl
        id="ticker-enabled"
        v-model="settings.ticker_enabled"
        :aria-label="t('setting.ticker.enabled')"
      />
    </header>
    <div class="preferences-body">
      <section
        class="ticker-preview"
        :style="previewStyle"
        :aria-label="t('setting.ticker.preview')"
      >
        <div class="preview-caption">
          {{ t('setting.ticker.preview')
          }}<span>{{
            t(settings.ticker_enabled ? 'setting.ticker.enabled' : 'setting.ticker.disabled')
          }}</span>
        </div>
        <div class="preview-line">
          <PhRss :size="15" /><span v-if="settings.ticker_show_source" class="preview-source">{{
            sourceName
          }}</span
          ><span class="preview-title">{{ t('setting.ticker.previewTitle') }}</span>
        </div>
      </section>
      <p class="preferences-note">{{ t('setting.ticker.enabledHint') }}</p>

      <section class="preferences-card">
        <div class="card-heading">
          <PhRss :size="16" />
          <h2>{{ t('setting.ticker.content') }}</h2>
          <span class="heading-value">{{
            settings.ticker_feed_id ? sourceName : t('setting.ticker.allFeeds')
          }}</span>
        </div>
        <div class="feed-search">
          <PhMagnifyingGlass :size="15" /><input
            v-model="search"
            type="search"
            :placeholder="t('setting.ticker.searchFeeds')"
            :aria-label="t('setting.ticker.searchFeeds')"
          />
        </div>
        <div
          class="feed-options"
          role="group"
          :aria-label="t('setting.ticker.feed')"
          :aria-busy="loadingFeeds"
        >
          <button
            class="feed-option"
            :aria-pressed="settings.ticker_feed_id === 0"
            @click="settings.ticker_feed_id = 0"
          >
            <span>{{ t('setting.ticker.allFeeds') }}</span
            ><PhCheck v-if="settings.ticker_feed_id === 0" :size="15" weight="bold" />
          </button>
          <button
            v-for="feed in filteredFeeds"
            :key="feed.id"
            class="feed-option"
            :aria-pressed="settings.ticker_feed_id === feed.id"
            @click="settings.ticker_feed_id = feed.id"
          >
            <span class="min-w-0"
              ><span class="feed-name">{{ feed.title }}</span
              ><small v-if="feed.category">{{ feed.category }}</small></span
            ><PhCheck v-if="settings.ticker_feed_id === feed.id" :size="15" weight="bold" />
          </button>
          <p v-if="search && !filteredFeeds.length" class="feed-empty">
            {{ t('setting.ticker.noMatchingFeeds') }}
          </p>
          <button v-if="feedError" class="feed-empty text-accent" @click="loadFeeds">
            {{ t('setting.ticker.loadFailed') }} · {{ t('setting.ticker.retry') }}
          </button>
        </div>
        <label class="preference-row"
          ><span>{{ t('setting.ticker.onlyUnread') }}</span
          ><ToggleControl
            v-model="settings.ticker_only_unread"
            :aria-label="t('setting.ticker.onlyUnread')"
        /></label>
        <p class="card-note">{{ t('setting.ticker.localRefreshHint') }}</p>
      </section>

      <section class="preferences-card">
        <div class="card-heading">
          <PhPalette :size="16" />
          <h2>{{ t('setting.ticker.appearance') }}</h2>
        </div>
        <div class="palette-options">
          <button
            v-for="palette in tickerPalettes"
            :key="palette.id"
            :aria-pressed="activePalette === palette.id"
            @click="choosePalette(palette)"
          >
            <span
              class="palette-swatch"
              :style="{
                background: `linear-gradient(110deg,${palette.start},${palette.end})`,
                color: palette.text,
              }"
              >Aa<PhCheck v-if="activePalette === palette.id" :size="12" weight="bold"
            /></span>
            <span>{{ t(`setting.ticker.palettes.${palette.id}`) }}</span>
          </button>
        </div>
        <div class="color-options">
          <label
            ><input v-model="settings.ticker_text_color" type="color" /><span>{{
              t('setting.ticker.textColor')
            }}</span></label
          >
          <label
            ><input v-model="settings.ticker_background_color" type="color" /><span>{{
              t('setting.ticker.backgroundStart')
            }}</span></label
          >
          <label
            ><input v-model="settings.ticker_background_color_end" type="color" /><span>{{
              t('setting.ticker.backgroundEnd')
            }}</span></label
          >
        </div>
        <label class="range-setting"
          ><span
            >{{ t('setting.ticker.speed')
            }}<output
              >{{ settings.ticker_speed }} {{ t('setting.ticker.pixelsPerSecond') }}</output
            ></span
          ><input v-model.number="settings.ticker_speed" type="range" min="10" max="200" /><span
            class="range-hints"
            ><small>{{ t('setting.ticker.slow') }}</small
            ><small>{{ t('setting.ticker.fast') }}</small></span
          ></label
        >
        <label class="range-setting"
          ><span
            >{{ t('setting.ticker.fontSize')
            }}<output>{{ settings.ticker_font_size }} px</output></span
          ><input v-model.number="settings.ticker_font_size" type="range" min="12" max="40"
        /></label>
      </section>

      <section class="preferences-card">
        <div class="card-heading">
          <PhSlidersHorizontal :size="16" />
          <h2>{{ t('setting.ticker.windowAndBehavior') }}</h2>
        </div>
        <div class="position-options" :aria-label="t('setting.ticker.position')">
          <button
            v-for="position in positions"
            :key="position.id"
            :aria-pressed="settings.ticker_position === position.id"
            @click="settings.ticker_position = position.id"
          >
            <component :is="position.icon" :size="16" />{{
              t(`setting.ticker.positions.${position.id}`)
            }}
          </button>
        </div>
        <label v-if="settings.ticker_position === 'floating'" class="range-setting"
          ><span
            >{{ t('setting.ticker.width') }}<output>{{ settings.ticker_width }} px</output></span
          ><input
            v-model.number="settings.ticker_width"
            type="range"
            min="360"
            max="1800"
            step="20"
        /></label>
        <label v-if="displays.length > 1" class="preference-row"
          ><span>{{ t('setting.ticker.display') }}</span
          ><select v-model="settings.ticker_screen_id">
            <option value="">{{ t('setting.ticker.currentDisplay') }}</option>
            <option v-for="display in displays" :key="display.id" :value="display.id">
              {{ display.name
              }}{{ display.primary ? ` · ${t('setting.ticker.primaryDisplay')}` : '' }}
            </option>
          </select></label
        >
        <label v-for="item in behavior" :key="item.key" class="preference-row"
          ><span
            >{{ t(`setting.ticker.${item.label}`)
            }}<small>{{ t(`setting.ticker.${item.description}`) }}</small></span
          ><ToggleControl
            v-model="settings[item.key]"
            :aria-label="t(`setting.ticker.${item.label}`)"
        /></label>
        <label class="preference-row"
          ><span>{{ t('setting.ticker.openArticle') }}</span
          ><select v-model="settings.ticker_open_in_browser">
            <option :value="false">{{ t('setting.ticker.inReader') }}</option>
            <option :value="true">{{ t('setting.ticker.inBrowser') }}</option>
          </select></label
        >
      </section>
      <div class="keyboard-note">
        <PhKeyboard :size="16" />
        <p>{{ t('setting.ticker.shortcuts') }}</p>
      </div>
    </div>
    <footer class="preferences-footer">
      <span class="save-state" :class="{ failed: status === 'error' }" role="status"
        ><PhSpinnerGap v-if="status === 'saving'" :size="14" class="animate-spin" /><PhWarningCircle
          v-else-if="status === 'error'"
          :size="14"
        /><PhCheck v-else :size="14" />{{ t(`setting.ticker.${status}`)
        }}<button v-if="status === 'error'" @click="saveNow">
          {{ t('setting.ticker.retry') }}
        </button></span
      >
      <button class="preferences-done" @click="finish">{{ t('setting.ticker.done') }}</button>
    </footer>
  </main>
</template>

<style scoped>
.ticker-preferences {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--bg-primary);
  color: var(--text-primary);
  font-family: var(--ui-font-family);
  font-size: 13px;
}
.preferences-header {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  gap: 12px;
  padding: 22px 22px 17px;
  border-bottom: 1px solid var(--border-color);
}
.preferences-icon {
  display: grid;
  place-items: center;
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  border-radius: 12px;
  background: var(--bg-tertiary);
  color: var(--accent-color);
}
.preferences-header h1 {
  margin: 0;
  font-size: 17px;
  font-weight: 650;
  letter-spacing: -0.025em;
}
.preferences-header p {
  margin: 4px 0 0;
  color: var(--text-secondary);
  font-size: 11px;
  line-height: 1.6;
}
.preferences-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 20px 22px 14px;
  scrollbar-width: thin;
}
.ticker-preview {
  overflow: hidden;
  padding: 13px 15px;
  border-radius: 11px;
  background: linear-gradient(110deg, var(--ticker-start), var(--ticker-end));
  color: var(--ticker-ink);
  box-shadow: 0 4px 12px #0000000b;
}
.preview-caption {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--ticker-chrome);
  opacity: 0.55;
}
.preview-line {
  display: flex;
  align-items: center;
  gap: 9px;
  height: 40px;
  margin-top: 6px;
  white-space: nowrap;
  font-size: var(--ticker-size);
}
.preview-line > svg {
  flex-shrink: 0;
  opacity: 0.6;
}
.preview-source {
  max-width: 90px;
  overflow: hidden;
  text-overflow: ellipsis;
  padding: 3px 6px;
  border-radius: 4px;
  font-size: 10px;
  color: var(--ticker-chrome);
  background: color-mix(in srgb, var(--ticker-chrome) 10%, transparent);
}
.preview-title {
  overflow: hidden;
  text-overflow: ellipsis;
}
.preferences-note,
.card-note {
  color: var(--text-tertiary);
  font-size: 11px;
  line-height: 1.7;
}
.preferences-note {
  margin: 9px 2px 18px;
}
.preferences-card {
  padding: 16px;
  margin-block: 14px;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  background: var(--bg-secondary);
}
.card-heading {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
  color: var(--text-secondary);
}
.card-heading h2 {
  color: var(--text-primary);
  font-size: 12px;
  font-weight: 650;
  margin: 0;
  white-space: nowrap;
}
.heading-value {
  margin-left: auto;
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 11px;
}
.feed-search {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border: 1px solid var(--border-color);
  border-radius: 7px;
  background: var(--bg-primary);
  color: var(--text-tertiary);
}
.feed-search input {
  min-width: 0;
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-primary);
  font-size: 11px;
}
.feed-search:focus-within {
  border-color: var(--accent-color);
}
.feed-options {
  max-height: 150px;
  overflow: auto;
  margin: 8px -4px 4px;
  scrollbar-width: thin;
}
.feed-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  border-radius: 6px;
  padding: 8px;
  text-align: left;
  cursor: pointer;
  font-size: 12px;
}
.feed-option[aria-pressed='true'] {
  color: var(--accent-color);
  background: var(--bg-tertiary);
  font-weight: 550;
}
.feed-option:hover {
  background: var(--bg-tertiary);
}
.feed-option > svg {
  flex-shrink: 0;
}
.feed-name {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.feed-option small {
  display: block;
  font-weight: 400;
  color: var(--text-tertiary);
  font-size: 11px;
  margin-top: 3px;
}
.feed-empty {
  padding: 9px;
  font-size: 11px;
  color: var(--text-secondary);
}
.preference-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-block: 12px;
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  font-size: 12px;
}
.preference-row:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}
.preference-row small {
  display: block;
  font-size: 11px;
  line-height: 1.6;
  color: var(--text-tertiary);
  margin-top: 3px;
}
.preference-row select {
  max-width: 54%;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 6px;
  font-size: 11px;
}
.card-note {
  margin: 8px 0 0;
}
.palette-options {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 7px;
}
.palette-options button {
  min-width: 0;
  text-align: center;
  font-size: 11px;
  color: var(--text-secondary);
  cursor: pointer;
}
.palette-swatch {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 34px;
  padding: 0 10px;
  border-radius: 6px;
  margin-bottom: 6px;
  font-size: 15px;
  border: 1px solid #0000000c;
}
.palette-options [aria-pressed='true'] .palette-swatch {
  outline: 2px solid var(--accent-color);
  outline-offset: 2px;
}
.color-options {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin: 18px 0 14px;
}
.color-options label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-secondary);
  cursor: pointer;
}
.color-options input {
  height: 22px;
  width: 24px;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: pointer;
}
.range-setting {
  display: block;
  margin-top: 16px;
  font-size: 12px;
}
.range-setting > span {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.range-setting output {
  color: var(--text-secondary);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}
.range-setting input {
  width: 100%;
  accent-color: var(--accent-color);
  margin-top: 11px;
  cursor: pointer;
}
.range-hints {
  color: var(--text-tertiary);
}
.range-hints small {
  font-size: 10px;
}
.position-options {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 5px;
  padding: 4px;
  background: var(--bg-tertiary);
  border-radius: 8px;
  margin-bottom: 4px;
}
.position-options button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 3px;
  border-radius: 5px;
  font-size: 11px;
  color: var(--text-secondary);
  cursor: pointer;
}
.position-options [aria-pressed='true'] {
  background: var(--bg-primary);
  color: var(--text-primary);
  box-shadow: 0 1px 4px #0000000b;
}
.keyboard-note {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin: 18px 2px 10px;
  color: var(--text-tertiary);
}
.keyboard-note svg {
  flex-shrink: 0;
  margin-top: 2px;
}
.keyboard-note p {
  margin: 0;
  font-size: 11px;
  line-height: 1.8;
}
.preferences-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  gap: 12px;
  padding: 13px 22px;
  background: var(--bg-primary);
  border-top: 1px solid var(--border-color);
}
.save-state {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  color: var(--text-tertiary);
}
.save-state.failed {
  color: var(--accent-color);
}
.save-state button {
  text-decoration: underline;
}
.preferences-done {
  background: var(--accent-color);
  color: white;
  padding: 7px 21px;
  border-radius: 7px;
  font-size: 12px;
  font-weight: 550;
  cursor: pointer;
}
.ticker-preferences button:focus-visible,
.ticker-preferences select:focus-visible,
.ticker-preferences input:focus-visible {
  outline: 2px solid var(--accent-color);
  outline-offset: 2px;
}
</style>
