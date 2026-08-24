<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
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
const customStartDate = ref('');
const customEndDate = ref('');

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
  loading.value = true;
  error.value = undefined;

  try {
    let url = `/api/statistics?period=${selectedPeriod.value}`;

    if (selectedPeriod.value !== 'custom' && selectedPeriod.value !== 'all') {
      url += `&offset=${currentOffset.value}`;
    } else if (selectedPeriod.value === 'custom' && customStartDate.value && customEndDate.value) {
      url += `&start_date=${customStartDate.value}&end_date=${customEndDate.value}`;
    }

    const response = await fetch(url);
    if (!response.ok) throw new Error('Failed to fetch statistics');
    stats.value = await response.json();
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Unknown error';
    console.error('Error fetching statistics:', e);
  } finally {
    loading.value = false;
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

function onIntervalChange(event: Event) {
  const target = event.target as HTMLSelectElement;
  setPeriod(target.value as Period);
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

onMounted(async () => {
  await fetchStatistics();

  // Set default custom dates to today
  const today = new Date();
  customEndDate.value = today.toISOString().split('T')[0];
  customStartDate.value = new Date(today.getFullYear(), today.getMonth(), 1)
    .toISOString()
    .split('T')[0];
});
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
            'flex items-center justify-between p-0 overflow-hidden border rounded-lg cursor-pointer text-xs font-medium transition-all',
            ['week', 'month', 'year'].includes(selectedPeriod)
              ? 'period-btn-active'
              : 'period-btn-inactive',
          ]"
          @click="handleIntervalClick"
        >
          <div class="flex items-center gap-1.5 px-3 py-2 flex-1">
            <PhCalendar :size="16" />
            <span class="text-xs font-medium">{{ t('setting.feed.fixedInterval') }}</span>
          </div>
          <div class="relative flex items-center" @click.stop>
            <select
              v-model="selectedInterval"
              class="appearance-none -webkit-appearance-none -moz-appearance-none flex items-center gap-1.5 px-3 pr-6 py-2 bg-bg-primary text-text-primary border-none border-l border-border text-xs font-medium cursor-pointer transition-all hover:bg-bg-primary hover:border-l-accent focus:outline-none focus:bg-bg-primary focus:border-l-accent"
              @change="onIntervalChange"
            >
              <option
                v-for="option in intervalOptions"
                :key="option.value"
                :value="option.value"
                class="bg-bg-primary text-text-primary"
              >
                {{ option.label }}
              </option>
            </select>
            <!-- Custom dropdown arrow -->
            <div class="absolute right-2.5 top-1/2 -translate-y-1/2 pointer-events-none opacity-60">
              <svg
                class="w-0 h-0 border-l-[3px] border-l-text-primary border-t-[2px] border-t-transparent border-b-[2px] border-b-transparent"
                :class="['week', 'month', 'year'].includes(selectedPeriod) ? 'border-l-white' : ''"
              ></svg>
            </div>
          </div>
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
        <input v-model="customStartDate" type="date" class="date-input" @change="fetchStatistics" />
      </div>
      <div class="flex flex-col gap-1.5 flex-1 min-w-[150px]">
        <label class="flex items-center gap-1.5 text-xs text-text-primary font-medium">
          <PhCalendarX :size="16" />
          {{ t('setting.statistic.endDate') }}:
        </label>
        <input v-model="customEndDate" type="date" class="date-input" @change="fetchStatistics" />
      </div>
      <button
        class="px-4 py-2 bg-accent text-white border-none rounded-md font-medium cursor-pointer transition-all hover:opacity-90 hover:-translate-y-px h-9"
        @click="fetchStatistics"
      >
        {{ t('common.form.apply') }}
      </button>
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

.period-btn-active {
  @apply bg-accent text-white border-accent hover:border-accent;
}

.period-btn-inactive {
  @apply border-border bg-bg-secondary text-text-primary hover:border-accent;
}

.date-input {
  @apply px-3 py-2 border border-border rounded-md bg-bg-primary text-text-primary text-xs transition-colors focus:outline-none focus:border-accent;
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
