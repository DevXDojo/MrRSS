<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhChartBar,
  PhArrowCounterClockwise,
  PhCalendar,
  PhEye,
  PhStar,
  PhReadCvLogo,
  PhCaretRight,
  PhCaretLeft,
  PhFlagPennant,
  PhChats,
  PhCalendarDots,
  PhCalendarPlus,
  PhCalendarX,
  PhCalendarStar,
} from '@phosphor-icons/vue';
import { ButtonControl } from '@/components/settings';
import BaseSelect from '@/components/common/BaseSelect.vue';
import BaseDatePicker from '@/components/common/BaseDatePicker.vue';
import { formatCalendarDate } from '@/utils/date';

const { t } = useI18n();

type Period = 'week' | 'month' | 'year' | 'all' | 'custom';

interface StatSummary {
  period: Period;
  start_date: string;
  end_date: string;
  totals: Record<string, number>;
  daily_data?: Record<string, Record<string, number>>;
  can_navigate: boolean;
  has_previous: boolean;
  has_next: boolean;
  display_label: string;
}

const selectedPeriod = ref<Period>('week');
const currentOffset = ref(0);
const loading = ref(false);
const error = ref<string>();
const stats = ref<StatSummary>();
const isResetting = ref(false);

// Separate state for the select dropdown to preserve value when switching away
const selectedInterval = ref<Period>('week');

// Watch for period changes and sync the interval select
watch(
  () => selectedPeriod.value,
  (newPeriod) => {
    if (['week', 'month', 'year'].includes(newPeriod)) {
      selectedInterval.value = newPeriod;
    }
  }
);

// Custom date range
const today = new Date();
const customEndDate = ref(formatCalendarDate(today, '', 'yyyy-mm-dd'));
const customStartDate = ref(`${customEndDate.value.slice(0, 7)}-01`);
const hasCustomRange = computed(() => Boolean(customStartDate.value && customEndDate.value));
let statisticsRequest: AbortController | undefined;

// All stat types to display (even if zero)
const allStatTypes = [
  'feed_refresh',
  'article_read',
  'article_view',
  'ai_chat',
  'ai_summary',
  'article_favorite',
] as const;

// Use computed for statLabels to avoid top-level t() calls
const statLabels = computed<Record<string, string>>(() => ({
  feed_refresh: t('modal.feed.refreshes'),
  article_read: t('setting.statistic.articlesRead'),
  article_view: t('setting.statistic.articlesViewed'),
  ai_chat: t('setting.statistic.aiChats'),
  ai_summary: t('setting.statistic.aiSummaries'),
  article_favorite: t('setting.statistic.articlesFavorited'),
}));

const statIcons: Record<string, any> = {
  feed_refresh: PhArrowCounterClockwise,
  article_read: PhReadCvLogo,
  article_view: PhEye,
  ai_chat: PhChats,
  ai_summary: PhFlagPennant,
  article_favorite: PhStar,
};

const statColors: Record<string, string> = {
  feed_refresh: 'var(--accent-color)',
  article_read: 'var(--accent-color)',
  article_view: 'var(--accent-color)',
  ai_chat: 'var(--accent-color)',
  ai_summary: 'var(--accent-color)',
  article_favorite: 'var(--accent-color)',
};

// Use computed for intervalOptions to avoid top-level t() calls
const intervalOptions = computed(() => [
  { value: 'week' as Period, label: t('setting.statistic.byWeek'), icon: PhCalendar },
  { value: 'month' as Period, label: t('setting.statistic.byMonth'), icon: PhCalendar },
  { value: 'year' as Period, label: t('setting.statistic.byYear'), icon: PhCalendar },
]);

// Whether the fixed-interval group is the active period
const isIntervalPeriod = computed(() => ['week', 'month', 'year'].includes(selectedPeriod.value));

const totalStats = computed(() => {
  const currentStats = stats.value;
  if (!currentStats?.totals) return [];

  return allStatTypes.map((key) => ({
    key,
    label: statLabels.value[key] || key,
    value: currentStats.totals[key] || 0,
    icon: statIcons[key] || PhChartBar,
    color: statColors[key] || '#6b7280',
  }));
});

const showCustomDatePickers = computed(() => selectedPeriod.value === 'custom');
const showNavigation = computed(
  () =>
    stats.value?.can_navigate &&
    (selectedPeriod.value === 'week' ||
      selectedPeriod.value === 'month' ||
      selectedPeriod.value === 'year')
);
const displayLabel = computed(() => stats.value?.display_label || '');

async function fetchStatistics() {
  statisticsRequest?.abort();
  if (selectedPeriod.value === 'custom' && !hasCustomRange.value) {
    loading.value = false;
    return;
  }
  const request = new AbortController();
  statisticsRequest = request;
  loading.value = true;
  error.value = undefined;

  try {
    let url = `/api/statistics?period=${selectedPeriod.value}`;

    if (selectedPeriod.value !== 'custom' && selectedPeriod.value !== 'all') {
      url += `&offset=${currentOffset.value}`;
    } else if (selectedPeriod.value === 'custom' && customStartDate.value && customEndDate.value) {
      url += `&start_date=${customStartDate.value}&end_date=${customEndDate.value}`;
    }

    const response = await fetch(url, { signal: request.signal });
    if (!response.ok) throw new Error('Failed to fetch statistics');
    const result: StatSummary = await response.json();
    if (!request.signal.aborted) stats.value = result;
  } catch (e) {
    if (request.signal.aborted) return;
    error.value = e instanceof Error ? e.message : 'Unknown error';
    console.error('Error fetching statistics:', e);
  } finally {
    if (!request.signal.aborted) loading.value = false;
  }
}

function navigatePeriod(direction: number) {
  currentOffset.value += direction;
  fetchStatistics();
}

function setPeriod(period: Period) {
  selectedPeriod.value = period;
  currentOffset.value = 0;
  fetchStatistics();
}

function handleIntervalClick() {
  // If clicking the interval button but not on one of the interval modes,
  // switch to the interval that's currently selected in the dropdown
  if (!['week', 'month', 'year'].includes(selectedPeriod.value)) {
    setPeriod(selectedInterval.value);
  }
}

function onIntervalChange(value: string | number) {
  selectedInterval.value = value as Period;
  setPeriod(selectedInterval.value);
}

async function resetStatistics() {
  const confirmed = await window.showConfirm({
    title: t('setting.statistic.resetToDefault'),
    message: t('setting.statistic.resetConfirm'),
    isDanger: true,
  });
  if (!confirmed) return;

  isResetting.value = true;
  try {
    const response = await fetch('/api/statistics', {
      method: 'DELETE',
    });

    if (response.ok) {
      window.showToast(t('setting.statistic.resetSuccess'), 'success');
      // Refresh statistics after reset
      await fetchStatistics();
    } else {
      console.error('Server error:', response.status);
      window.showToast(t('setting.statistic.resetFailed'), 'error');
    }
  } catch (error) {
    console.error('Failed to reset statistics:', error);
    window.showToast(t('setting.statistic.resetFailed'), 'error');
  } finally {
    isResetting.value = false;
  }
}

onMounted(fetchStatistics);
onBeforeUnmount(() => statisticsRequest?.abort());
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between mb-3">
      <div class="flex items-center gap-2 sm:gap-3">
        <PhChartBar :size="20" class="text-text-secondary sm:w-6 sm:h-6" />
        <div>
          <h3 class="font-semibold text-sm sm:text-base">
            {{ t('setting.statistic.statistics') }}
          </h3>
          <p class="text-xs text-text-secondary hidden sm:block">
            {{ t('setting.statistic.description') }}
          </p>
        </div>
      </div>
      <ButtonControl
        :label="t('setting.statistic.resetToDefault')"
        :icon="PhArrowCounterClockwise"
        :disabled="isResetting"
        type="secondary"
        @click="resetStatistics"
      />
    </div>

    <!-- Period Selector -->
    <div class="flex flex-col gap-1.5">
      <div class="flex flex-wrap gap-1.5">
        <!-- All Time -->
        <button
          :class="[
            'period-btn',
            selectedPeriod === 'all' ? 'period-btn-active' : 'period-btn-inactive',
          ]"
          @click="setPeriod('all')"
        >
          <PhCalendarStar :size="16" />
          {{ t('setting.statistic.allTime') }}
        </button>

        <!-- Fixed Interval with Dropdown -->
        <div
          :class="[
            'interval-group',
            isIntervalPeriod ? 'period-btn-active' : 'period-btn-inactive',
          ]"
        >
          <button
            type="button"
            class="interval-group-label"
            :aria-pressed="isIntervalPeriod"
            @click="handleIntervalClick"
          >
            <PhCalendar :size="16" />
            {{ t('setting.feed.fixedInterval') }}
          </button>
          <span class="interval-group-divider" aria-hidden="true"></span>
          <BaseSelect
            class="interval-group-select"
            dropdown-class="select-dropdown-neutral"
            :model-value="selectedInterval"
            :options="intervalOptions"
            width="w-24"
            size="xs"
            @update:model-value="onIntervalChange"
          />
        </div>

        <!-- Custom Range -->
        <button
          :class="[
            'period-btn',
            selectedPeriod === 'custom' ? 'period-btn-active' : 'period-btn-inactive',
          ]"
          @click="setPeriod('custom')"
        >
          <PhCalendarDots :size="16" />
          {{ t('setting.statistic.customRange') }}
        </button>
      </div>
    </div>

    <!-- Custom Date Range Pickers -->
    <div
      v-if="showCustomDatePickers"
      class="flex flex-wrap gap-3 p-3 bg-bg-secondary rounded-lg items-end"
    >
      <div class="flex flex-col gap-1.5 flex-1 min-w-[150px]">
        <label class="flex items-center gap-1.5 text-xs text-text-primary font-medium">
          <PhCalendarPlus :size="16" />
          {{ t('setting.statistic.startDate') }}:
        </label>
        <BaseDatePicker
          v-model="customStartDate"
          :max="customEndDate"
          :placeholder="t('common.datePicker.placeholder')"
          :aria-label="t('setting.statistic.startDate')"
          @change="fetchStatistics"
        />
      </div>
      <div class="flex flex-col gap-1.5 flex-1 min-w-[150px]">
        <label class="flex items-center gap-1.5 text-xs text-text-primary font-medium">
          <PhCalendarX :size="16" />
          {{ t('setting.statistic.endDate') }}:
        </label>
        <BaseDatePicker
          v-model="customEndDate"
          :min="customStartDate"
          :placeholder="t('common.datePicker.placeholder')"
          :aria-label="t('setting.statistic.endDate')"
          @change="fetchStatistics"
        />
      </div>
      <ButtonControl
        type="primary"
        :label="t('common.form.apply')"
        :disabled="!hasCustomRange"
        class="h-[var(--ui-control-height)]"
        @click="fetchStatistics"
      />
    </div>

    <!-- Navigation -->
    <div
      v-if="showNavigation"
      class="flex items-center justify-center gap-1.5 p-2 bg-bg-secondary rounded-lg"
    >
      <button
        class="nav-btn"
        :disabled="!stats?.has_previous || loading"
        @click="navigatePeriod(-1)"
      >
        <PhCaretLeft :size="20" />
      </button>
      <div class="text-xs font-semibold text-text-primary min-w-[180px] text-center">
        {{ displayLabel }}
      </div>
      <button class="nav-btn" :disabled="!stats?.has_next || loading" @click="navigatePeriod(1)">
        <PhCaretRight :size="20" />
      </button>
    </div>

    <!-- Error State -->
    <div v-if="error" class="flex items-center justify-center p-8 gap-4">
      <p class="text-red-500">{{ t('common.error') }}: {{ error }}</p>
    </div>

    <!-- Statistics Display -->
    <div v-else class="flex flex-col gap-6">
      <div class="grid grid-cols-3 gap-3">
        <div v-for="stat in totalStats" :key="stat.key" class="stat-card">
          <div class="stat-icon-wrapper">
            <component :is="stat.icon" :size="28" class="text-text-tertiary" />
          </div>
          <div class="flex flex-col gap-1 flex-1">
            <p class="text-xs font-semibold uppercase tracking-wider m-0 text-text-secondary">
              {{ stat.label }}
            </p>
            <p class="text-[1.75rem] font-bold text-accent m-0 leading-none">{{ stat.value }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";
.period-btn {
  @apply flex items-center gap-1.5 px-3 py-2 border rounded-lg cursor-pointer text-xs font-medium transition-all;
}

/* Fixed-interval segmented control: one pill made of a label + inline select.
   It is filled with the accent color while one of the fixed intervals is active. */
.interval-group {
  @apply inline-flex items-stretch border rounded-lg cursor-pointer text-xs font-medium transition-all;
}

.interval-group-label {
  @apply flex items-center gap-1.5 px-3 py-2 bg-transparent border-0 text-xs font-medium cursor-pointer;
  /* Inherit the pill's text color */
  color: inherit;
  white-space: nowrap;
  border-start-start-radius: calc(var(--ui-radius-control) - 1px);
  border-end-start-radius: calc(var(--ui-radius-control) - 1px);
}

.interval-group-divider {
  @apply w-px flex-shrink-0 self-stretch;
  background-color: var(--border-color);
}

/* Let the select blend into the pill instead of showing its own boxed trigger */
.interval-group-select :deep(.select-trigger) {
  @apply flex items-center justify-between w-full cursor-pointer;
  height: auto;
  min-height: 0;
  padding: 0.5rem;
  border: 0;
  border-radius: 0 calc(var(--ui-radius-control) - 1px) calc(var(--ui-radius-control) - 1px) 0;
  background-color: transparent;
  color: inherit;
  font-size: 0.75rem;
  line-height: 1rem;
  font-weight: 500;
  transition: background-color var(--ui-transition-fast);
}

.interval-group-select :deep(.select-trigger:hover) {
  background-color: color-mix(in srgb, currentColor 10%, transparent);
}

.interval-group-select :deep(.select-arrow) {
  @apply ml-1 flex-shrink-0;
  color: inherit;
  opacity: 0.85;
}

/* Divider and hover tint need more contrast on the filled (active) pill */
.interval-group.period-btn-active .interval-group-divider {
  background-color: color-mix(in srgb, #ffffff 45%, transparent);
}

.interval-group.period-btn-active .interval-group-select :deep(.select-trigger:hover) {
  background-color: color-mix(in srgb, #ffffff 18%, transparent);
}

.period-btn-active {
  @apply bg-accent text-white border-accent hover:border-accent;
}

.period-btn-inactive {
  @apply border-border bg-bg-secondary text-text-primary hover:border-accent;
}

.nav-btn {
  @apply flex items-center justify-center w-7 h-7 border border-border bg-bg-primary text-text-secondary rounded-md cursor-pointer transition-all hover:bg-accent hover:text-white hover:border-accent disabled:opacity-40 disabled:cursor-not-allowed;
}

.stat-card {
  @apply relative flex items-center gap-3 px-4 py-3 bg-bg-secondary border-2 border-border rounded-lg transition-all hover:shadow-sm;
}

.stat-card:hover {
  border-color: color-mix(in srgb, var(--accent-color) 50%, transparent);
}

.stat-icon-wrapper {
  @apply flex items-center justify-center w-12 h-12 rounded-lg flex-shrink-0;
  background-color: color-mix(in srgb, var(--bg-tertiary) 50%, transparent);
}
</style>
