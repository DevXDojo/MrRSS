<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { Call } from '@wailsio/runtime';
import { useI18n } from 'vue-i18n';
import {
  PhRss,
  PhEye,
  PhFunnel,
  PhPalette,
  PhTextT,
  PhSlidersHorizontal,
  PhMonitor,
  PhArrowSquareOut,
  PhPause,
  PhClock,
  PhPushPin,
  PhWarningCircle,
} from '@phosphor-icons/vue';
import BaseMultiSelect from '@/components/common/BaseMultiSelect.vue';
import { SubSettingItem, SelectControl, ToggleControl, NumberControl } from '@/components/settings';
import type { SettingsData } from '@/types/settings';
import type { Feed } from '@/types/models';
import {
  tickerNumber,
  tickerFeedIDs,
  tickerPalette,
  tickerPalettes,
  tickerStyle,
} from '@/components/ticker/tickerPreferences';

const props = defineProps<{ settings: SettingsData }>();
const emit = defineEmits<{ 'update:settings': [settings: SettingsData] }>();
const { t } = useI18n();
const feeds = ref<Feed[]>([]);
const feedError = ref(false);
const loadingFeeds = ref(false);
const displays = ref<Array<{ id: string; name: string; primary: boolean }>>([]);
const sourceOptions = computed(() =>
  feeds.value.map((feed) => ({ value: feed.id, label: feed.title }))
);
const selectedFeedIDs = computed(() => tickerFeedIDs(props.settings));
const displayOptions = computed(() => [
  { value: '', label: t('setting.ticker.currentDisplay') },
  ...displays.value.map((display) => ({
    value: display.id,
    label: display.name + (display.primary ? ` · ${t('setting.ticker.primaryDisplay')}` : ''),
  })),
]);
const behavior = [
  {
    icon: PhPause,
    key: 'ticker_pause_on_hover',
    label: 'pauseOnHover',
    description: 'pauseOnHoverHint',
  },
  { icon: PhRss, key: 'ticker_show_source', label: 'showSource', description: 'showSourceHint' },
  { icon: PhClock, key: 'ticker_show_time', label: 'showTime', description: 'showTimeHint' },
  {
    icon: PhPushPin,
    key: 'ticker_always_on_top',
    label: 'alwaysOnTop',
    description: 'alwaysOnTopHint',
  },
] as const;
let controller: AbortController | undefined;
let disposed = false;

function update<K extends keyof SettingsData>(key: K, value: SettingsData[K]) {
  emit('update:settings', { ...props.settings, [key]: value });
}
function updateSources(values: (string | number)[]) {
  const ids = values.map(Number);
  emit('update:settings', {
    ...props.settings,
    ticker_feed_ids: ids.join(',') || '0',
    ticker_feed_id: ids.length === 1 ? ids[0] : 0,
  });
}
const activePalette = computed(() => tickerPalette(props.settings).id);
const paletteOptions = computed(() =>
  tickerPalettes.map((palette) => ({
    value: palette.id,
    label: t(`setting.ticker.palettes.${palette.id}`),
  }))
);
function choosePalette(value: string | number) {
  const palette = tickerPalettes.find((item) => item.id === value);
  if (palette)
    emit('update:settings', {
      ...props.settings,
      ticker_text_color: palette.text,
      ticker_background_color: palette.start,
      ticker_background_color_end: palette.end,
    });
}
async function loadFeeds() {
  controller?.abort();
  const current = new AbortController();
  controller = current;
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
    if (!disposed && !current.signal.aborted) feedError.value = true;
  } finally {
    if (!disposed && !current.signal.aborted) loadingFeeds.value = false;
  }
}
onMounted(async () => {
  void loadFeeds();
  try {
    const result = await Call.ByName('main.DesktopCapabilities.GetTickerDisplays');
    if (!disposed) displays.value = (result as typeof displays.value) ?? [];
  } catch {
    /* Placement defaults to the current display. */
  }
});
onUnmounted(() => {
  disposed = true;
  controller?.abort();
});
</script>

<template>
  <SubSettingItem
    :icon="PhEye"
    :title="t('setting.ticker.preview')"
    :description="t('setting.ticker.previewHint')"
  >
    <div class="ticker-preview w-48 sm:w-64" :style="tickerStyle(settings)">
      <PhRss :size="16" class="shrink-0 opacity-60" />
      <span v-if="settings.ticker_show_source" class="preview-source">{{
        feeds.find((feed) => selectedFeedIDs.includes(feed.id))?.title ||
        t('setting.ticker.allFeeds')
      }}</span>
      <span class="truncate">{{ t('setting.ticker.previewTitle') }}</span>
    </div>
  </SubSettingItem>
  <SubSettingItem
    :icon="PhRss"
    :title="t('setting.ticker.feed')"
    :description="t('setting.ticker.localRefreshHint')"
  >
    <BaseMultiSelect
      :model-value="selectedFeedIDs"
      :options="sourceOptions"
      searchable
      :disabled="loadingFeeds || feedError"
      width="w-48 sm:w-64"
      :placeholder="t('setting.ticker.allFeeds')"
      @update:model-value="updateSources"
    />
  </SubSettingItem>
  <SubSettingItem
    v-if="feedError"
    :icon="PhWarningCircle"
    :title="t('setting.ticker.loadFailed')"
    :description="t('setting.ticker.sourcesRetryHint')"
  >
    <button class="text-xs text-accent hover:underline" @click="loadFeeds">
      {{ t('setting.ticker.retry') }}
    </button>
  </SubSettingItem>
  <SubSettingItem
    :icon="PhFunnel"
    :title="t('setting.ticker.onlyUnread')"
    :description="t('setting.ticker.onlyUnreadHint')"
  >
    <ToggleControl
      :model-value="settings.ticker_only_unread"
      @update:model-value="update('ticker_only_unread', $event)"
    />
  </SubSettingItem>
  <SubSettingItem
    :icon="PhPalette"
    :title="t('setting.ticker.appearance')"
    :description="t('setting.ticker.appearanceHint')"
  >
    <SelectControl
      :model-value="activePalette"
      :options="paletteOptions"
      @update:model-value="choosePalette"
    />
  </SubSettingItem>
  <SubSettingItem
    :icon="PhSlidersHorizontal"
    :title="t('setting.ticker.speed')"
    :description="t('setting.ticker.speedHint')"
  >
    <NumberControl
      :model-value="settings.ticker_speed"
      :min="10"
      :max="200"
      :suffix="t('setting.ticker.pixelsPerSecond')"
      @update:model-value="update('ticker_speed', tickerNumber($event, 60, 10, 200))"
    />
  </SubSettingItem>
  <SubSettingItem
    :icon="PhTextT"
    :title="t('setting.ticker.fontSize')"
    :description="t('setting.ticker.fontSizeHint')"
  >
    <NumberControl
      :model-value="settings.ticker_font_size"
      :min="12"
      :max="40"
      suffix="px"
      @update:model-value="update('ticker_font_size', tickerNumber($event, 18, 12, 40))"
    />
  </SubSettingItem>
  <SubSettingItem
    :icon="PhMonitor"
    :title="t('setting.ticker.position')"
    :description="t('setting.ticker.positionHint')"
  >
    <SelectControl
      :model-value="settings.ticker_position"
      :options="
        ['top', 'bottom', 'floating'].map((value) => ({
          value,
          label: t(`setting.ticker.positions.${value}`),
        }))
      "
      @update:model-value="update('ticker_position', String($event))"
    />
  </SubSettingItem>
  <SubSettingItem
    v-if="displays.length > 1"
    :icon="PhMonitor"
    :title="t('setting.ticker.display')"
    :description="t('setting.ticker.displayHint')"
  >
    <SelectControl
      :model-value="settings.ticker_screen_id"
      :options="displayOptions"
      width="lg"
      @update:model-value="update('ticker_screen_id', String($event))"
    />
  </SubSettingItem>
  <SubSettingItem
    v-for="item in behavior"
    :key="item.key"
    :icon="item.icon"
    :title="t(`setting.ticker.${item.label}`)"
    :description="t(`setting.ticker.${item.description}`)"
  >
    <ToggleControl
      :model-value="settings[item.key]"
      @update:model-value="update(item.key, $event)"
    />
  </SubSettingItem>
  <SubSettingItem
    :icon="PhArrowSquareOut"
    :title="t('setting.ticker.openArticle')"
    :description="t('setting.ticker.openArticleHint')"
  >
    <SelectControl
      :model-value="settings.ticker_open_in_browser ? 1 : 0"
      :options="[
        { value: 0, label: t('setting.ticker.inReader') },
        { value: 1, label: t('setting.ticker.inBrowser') },
      ]"
      @update:model-value="update('ticker_open_in_browser', Number($event) === 1)"
    />
  </SubSettingItem>
</template>

<style scoped>
.ticker-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 44px;
  padding: 10px 16px;
  border-radius: var(--ui-radius-surface);
  background: linear-gradient(110deg, var(--ticker-start), var(--ticker-end));
  color: var(--ticker-ink);
  font-size: var(--ticker-size);
  overflow: hidden;
}
.preview-source {
  flex-shrink: 0;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 10px;
  opacity: 0.65;
}
</style>
