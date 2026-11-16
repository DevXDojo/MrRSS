<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhCalendar, PhCaretLeft, PhCaretRight } from '@phosphor-icons/vue';
import { useSettings } from '@/composables/core/useSettings';
import { formatCalendarDate } from '@/utils/date';

const props = withDefaults(
  defineProps<{
    modelValue?: string;
    min?: string;
    max?: string;
    disabled?: boolean;
    placeholder?: string;
    ariaLabel?: string;
  }>(),
  {
    modelValue: '',
    min: '',
    max: '',
    disabled: false,
    placeholder: '',
    ariaLabel: '',
  }
);

const emit = defineEmits<{
  'update:modelValue': [value: string];
  change: [value: string];
}>();

const { t, locale } = useI18n();
const { settings } = useSettings();

const PANEL_WIDTH = 272;
const PANEL_HEIGHT = 300;
const PANEL_MARGIN = 6;

const rootRef = ref<HTMLElement>();
const inputRef = ref<HTMLInputElement>();
const panelRef = ref<HTMLElement>();
const isOpen = ref(false);
const isEditing = ref(false);
const panelStyle = ref<Record<string, string>>({
  position: 'fixed',
  top: '0px',
  left: '0px',
  width: `${PANEL_WIDTH}px`,
  zIndex: '70',
});

const localeTag = computed(() => String(locale.value || 'en-US'));
const dateFormat = computed(() => String(settings.value?.date_format || 'locale'));

type DateField = 'year' | 'month' | 'day';

function pad(value: number): string {
  return String(value).padStart(2, '0');
}

function toIso(date: Date): string {
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

function parseIso(value: string): Date | null {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value || '');
  if (!match) return null;
  const date = new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3]));
  return Number.isNaN(date.getTime()) ? null : date;
}

function formatIso(value: string): string {
  const date = parseIso(value);
  return date ? formatCalendarDate(date, localeTag.value, dateFormat.value) : '';
}

const valueDate = computed(() => parseIso(props.modelValue));
const displayValue = computed(() => formatIso(props.modelValue));

function outOfRange(iso: string): boolean {
  if (!iso) return false;
  if (props.min && iso < props.min) return true;
  if (props.max && iso > props.max) return true;
  return false;
}

// ---- Text input (typed dates follow the same i18n format as the display) ----

const text = ref(displayValue.value);

watch(
  displayValue,
  (value) => {
    if (!isEditing.value) text.value = value;
  },
  { immediate: true }
);

// Derive the field order of the active format from a sample date, so typed input
// is parsed with the same order the user sees (e.g. 10/31/2026 vs 2026/10/31).
const formatTokens = computed(() => {
  const sample = formatCalendarDate(new Date(2001, 10, 22), localeTag.value, dateFormat.value);
  const roles: DateField[] = [];
  let pattern = '';
  for (const chunk of sample.match(/\d+|\D+/g) || []) {
    if (/^\d+$/.test(chunk)) {
      if (chunk.length === 4) {
        roles.push('year');
        pattern += '(\\d{4})';
      } else if (chunk === '11') {
        roles.push('month');
        pattern += '(\\d{1,2})';
      } else {
        roles.push('day');
        pattern += '(\\d{1,2})';
      }
    } else {
      pattern += chunk.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    }
  }
  return { pattern, roles };
});

function parseText(value: string): Date | null {
  const { pattern, roles } = formatTokens.value;
  const match = new RegExp(`^\\s*${pattern}\\s*$`).exec(value.trim());
  if (!match) return null;
  const fields: Record<string, number> = {};
  roles.forEach((role, index) => {
    fields[role] = Number(match[index + 1]);
  });
  const date = new Date(fields.year, fields.month - 1, fields.day);
  if (
    date.getFullYear() !== fields.year ||
    date.getMonth() !== fields.month - 1 ||
    date.getDate() !== fields.day
  ) {
    return null;
  }
  return date;
}

function commit(nextIso?: string): void {
  isEditing.value = false;

  let iso: string;
  if (nextIso !== undefined) {
    iso = nextIso;
  } else {
    const raw = text.value.trim();
    if (!raw) {
      iso = '';
    } else {
      const parsed = parseText(raw);
      if (!parsed) {
        // Keep the previous valid value when the typed text cannot be parsed
        text.value = displayValue.value;
        return;
      }
      iso = toIso(parsed);
    }
  }

  if (iso && outOfRange(iso)) {
    text.value = displayValue.value;
    return;
  }

  text.value = formatIso(iso);
  if (iso !== (props.modelValue || '')) {
    emit('update:modelValue', iso);
    emit('change', iso);
  }
}

function onInput(event: Event): void {
  isEditing.value = true;
  text.value = (event.target as HTMLInputElement).value;
}

function onFocus(): void {
  isEditing.value = true;
}

function onBlur(): void {
  commit();
}

function cancelEdit(): void {
  isEditing.value = false;
  text.value = displayValue.value;
  close();
}

// ---- Calendar panel ----

const todayIso = ref(toIso(new Date()));
const viewYear = ref(new Date().getFullYear());
const viewMonth = ref(new Date().getMonth());

function syncView(): void {
  const date = valueDate.value || new Date();
  viewYear.value = date.getFullYear();
  viewMonth.value = date.getMonth();
}

function updatePosition(): void {
  const element = inputRef.value;
  if (!element) return;
  const rect = element.getBoundingClientRect();
  const left = Math.max(8, Math.min(rect.left, window.innerWidth - PANEL_WIDTH - 8));
  let top = rect.bottom + PANEL_MARGIN;
  if (top + PANEL_HEIGHT > window.innerHeight - 8 && rect.top > PANEL_HEIGHT) {
    top = rect.top - PANEL_MARGIN - PANEL_HEIGHT;
  }
  panelStyle.value = {
    position: 'fixed',
    top: `${Math.max(8, top)}px`,
    left: `${left}px`,
    width: `${PANEL_WIDTH}px`,
    zIndex: '70',
  };
}

function addListeners(): void {
  document.addEventListener('pointerdown', onPointerDown, true);
  window.addEventListener('resize', updatePosition);
  window.addEventListener('scroll', updatePosition, true);
  document.addEventListener('keydown', onKeydown);
}

function removeListeners(): void {
  document.removeEventListener('pointerdown', onPointerDown, true);
  window.removeEventListener('resize', updatePosition);
  window.removeEventListener('scroll', updatePosition, true);
  document.removeEventListener('keydown', onKeydown);
}

function open(): void {
  if (props.disabled || isOpen.value) return;
  syncView();
  updatePosition();
  isOpen.value = true;
  addListeners();
}

function close(): void {
  if (!isOpen.value) return;
  isOpen.value = false;
  removeListeners();
}

function toggle(): void {
  if (isOpen.value) {
    // Keep whatever the user typed before closing the panel
    commit();
    close();
    return;
  }
  open();
}

function onPointerDown(event: PointerEvent): void {
  const target = event.target as Node;
  if (rootRef.value?.contains(target) || panelRef.value?.contains(target)) return;
  commit();
  close();
}

function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    cancelEdit();
    return;
  }
  if (event.key === 'ArrowLeft') {
    changeMonth(-1);
  } else if (event.key === 'ArrowRight') {
    changeMonth(1);
  }
}

function changeMonth(delta: number): void {
  const next = new Date(viewYear.value, viewMonth.value + delta, 1);
  viewYear.value = next.getFullYear();
  viewMonth.value = next.getMonth();
}

function selectDate(iso: string): void {
  if (outOfRange(iso)) return;
  commit(iso);
  close();
}

function selectToday(): void {
  selectDate(todayIso.value);
}

function clear(): void {
  commit('');
  close();
}

const weekStart = computed(() => {
  try {
    const info = (
      new Intl.Locale(localeTag.value) as unknown as { weekInfo?: { firstDay?: number } }
    ).weekInfo;
    return typeof info?.firstDay === 'number' ? info.firstDay : 7;
  } catch {
    return 7;
  }
});

const weekdayLabels = computed(() => {
  const formatter = new Intl.DateTimeFormat(localeTag.value, { weekday: 'short' });
  // 2024-01-01 is a Monday, so this walks the local first day of week.
  return Array.from({ length: 7 }, (_, index) =>
    formatter.format(new Date(2024, 0, 1 + ((weekStart.value - 1 + index) % 7)))
  );
});

const monthLabel = computed(() =>
  new Intl.DateTimeFormat(localeTag.value, { month: 'long', year: 'numeric' }).format(
    new Date(viewYear.value, viewMonth.value, 1)
  )
);

const cells = computed(() => {
  const first = new Date(viewYear.value, viewMonth.value, 1);
  const offset = (first.getDay() - (weekStart.value % 7) + 7) % 7;
  const start = new Date(viewYear.value, viewMonth.value, 1 - offset);
  return Array.from({ length: 42 }, (_, index) => {
    const date = new Date(start.getFullYear(), start.getMonth(), start.getDate() + index);
    const iso = toIso(date);
    return {
      iso,
      day: date.getDate(),
      inMonth: date.getMonth() === viewMonth.value,
      isToday: iso === todayIso.value,
      isSelected: iso === (props.modelValue || ''),
      disabled: outOfRange(iso),
    };
  });
});

onBeforeUnmount(removeListeners);
</script>

<template>
  <div ref="rootRef" class="date-picker">
    <input
      ref="inputRef"
      class="date-picker-input"
      type="text"
      autocomplete="off"
      spellcheck="false"
      :value="text"
      :placeholder="placeholder || t('common.datePicker.placeholder')"
      :disabled="disabled"
      :aria-label="ariaLabel || placeholder || t('common.datePicker.placeholder')"
      @input="onInput"
      @click="open"
      @focus="onFocus"
      @blur="onBlur"
      @keydown.enter.prevent="commit()"
      @keydown.esc.prevent="cancelEdit"
    />
    <button
      type="button"
      class="date-picker-trigger"
      :disabled="disabled"
      :aria-label="t('common.datePicker.open')"
      :aria-expanded="isOpen"
      @click.stop="toggle"
    >
      <PhCalendar :size="16" />
    </button>

    <Teleport to="body">
      <div v-if="isOpen" ref="panelRef" class="date-picker-panel" :style="panelStyle">
        <div class="date-picker-header">
          <button
            type="button"
            class="date-picker-nav"
            :aria-label="t('common.datePicker.previousMonth')"
            @click="changeMonth(-1)"
          >
            <PhCaretLeft :size="14" />
          </button>
          <span class="date-picker-title">{{ monthLabel }}</span>
          <button
            type="button"
            class="date-picker-nav"
            :aria-label="t('common.datePicker.nextMonth')"
            @click="changeMonth(1)"
          >
            <PhCaretRight :size="14" />
          </button>
        </div>

        <div class="date-picker-weekdays">
          <span v-for="(label, index) in weekdayLabels" :key="index">{{ label }}</span>
        </div>

        <div class="date-picker-grid">
          <button
            v-for="cell in cells"
            :key="cell.iso"
            type="button"
            class="date-picker-day"
            :class="{
              'is-outside': !cell.inMonth,
              'is-today': cell.isToday,
              'is-selected': cell.isSelected,
            }"
            :disabled="cell.disabled"
            @click="selectDate(cell.iso)"
          >
            {{ cell.day }}
          </button>
        </div>

        <div class="date-picker-footer">
          <button type="button" class="date-picker-action" @click="selectToday">
            {{ t('common.datePicker.today') }}
          </button>
          <button
            type="button"
            class="date-picker-action"
            :disabled="!props.modelValue"
            @click="clear"
          >
            {{ t('common.datePicker.clear') }}
          </button>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
@reference "../../style.css";
.date-picker {
  @apply relative w-full;
}

.date-picker-input {
  @apply w-full px-3 py-2 pr-9 border border-border rounded-md bg-bg-primary text-text-primary text-xs transition-colors;
  @apply focus:outline-none focus:border-accent;
  height: var(--ui-control-height);
}

.date-picker-input::placeholder {
  @apply text-text-tertiary;
}

.date-picker-input:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.date-picker-trigger {
  @apply absolute right-1 top-1/2 -translate-y-1/2 flex items-center justify-center w-7 h-7 rounded border-0 bg-transparent text-text-secondary cursor-pointer transition-colors;
  @apply hover:text-accent;
}

.date-picker-trigger:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.date-picker-panel {
  @apply flex flex-col gap-1.5 p-2 border border-border bg-bg-primary rounded-lg shadow-lg;
}

.date-picker-header {
  @apply flex items-center justify-between gap-1;
}

.date-picker-title {
  @apply text-xs font-semibold text-text-primary;
}

.date-picker-nav {
  @apply flex items-center justify-center w-6 h-6 border border-transparent rounded bg-transparent text-text-secondary cursor-pointer transition-colors;
  @apply hover:bg-bg-tertiary hover:text-text-primary;
}

.date-picker-weekdays {
  @apply grid grid-cols-7 text-[0.6875rem] font-medium text-text-tertiary text-center;
}

.date-picker-grid {
  @apply grid grid-cols-7 gap-0.5;
}

.date-picker-day {
  @apply flex items-center justify-center h-7 border-0 rounded bg-transparent text-xs text-text-primary cursor-pointer transition-colors;
  @apply hover:bg-bg-tertiary;
}

.date-picker-day.is-outside {
  @apply text-text-tertiary;
}

.date-picker-day.is-today {
  color: var(--accent-color);
  font-weight: 600;
}

.date-picker-day.is-selected {
  background-color: var(--accent-color);
  color: #ffffff;
  font-weight: 600;
}

.date-picker-day.is-selected:hover {
  filter: brightness(0.92);
  background-color: var(--accent-color);
}

.date-picker-day:disabled {
  @apply opacity-40 cursor-not-allowed hover:bg-transparent;
}

.date-picker-footer {
  @apply flex items-center justify-between border-t border-border pt-1.5;
}

.date-picker-action {
  @apply px-1 py-0.5 border-0 rounded bg-transparent text-xs font-medium cursor-pointer transition-colors;
  color: var(--accent-color);
}

.date-picker-action:hover {
  background-color: color-mix(in srgb, var(--accent-color) 10%, transparent);
}

.date-picker-action:disabled {
  @apply opacity-40 cursor-not-allowed;
  background-color: transparent;
}
</style>
