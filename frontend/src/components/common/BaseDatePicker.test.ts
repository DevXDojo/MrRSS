import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { nextTick } from 'vue';

vi.mock('vue-i18n', async () => {
  const { ref } = await import('vue');
  const locale = ref('en-US');
  return {
    useI18n: () => ({ t: (key: string) => key, locale }),
    __test: { locale },
  };
});

vi.mock('@/composables/core/useSettings', async () => {
  const { ref } = await import('vue');
  const settings = ref<Record<string, unknown>>({ date_format: 'locale' });
  return {
    useSettings: () => ({ settings }),
    __test: { settings },
  };
});

import type { Ref } from 'vue';

const { locale } = ((await import('vue-i18n')) as unknown as { __test: { locale: Ref<string> } })
  .__test;
const { settings } = (
  (await import('@/composables/core/useSettings')) as unknown as {
    __test: { settings: Ref<Record<string, unknown>> };
  }
).__test;

const BaseDatePicker = (await import('./BaseDatePicker.vue')).default;

function mountPicker(props: Record<string, unknown> = {}) {
  return mount(BaseDatePicker, {
    props: { modelValue: '2026-10-31', ...props },
  });
}

function inputValue(wrapper: ReturnType<typeof mountPicker>): string {
  return (wrapper.get('input').element as HTMLInputElement).value;
}

// Week start comes from CLDR locale data, which can differ between runtimes
function firstWeekdayLabel(localeTag: string): string {
  const weekInfo = (new Intl.Locale(localeTag) as unknown as { weekInfo?: { firstDay?: number } })
    .weekInfo;
  const mondayOffset = (weekInfo?.firstDay ?? 7) % 7;
  return new Intl.DateTimeFormat(localeTag, { weekday: 'short' }).format(
    new Date(2024, 0, 7 + mondayOffset)
  );
}

describe('BaseDatePicker i18n formatting', () => {
  beforeEach(() => {
    locale.value = 'en-US';
    settings.value = { date_format: 'locale' };
  });

  afterEach(() => {
    document.body.innerHTML = '';
    vi.restoreAllMocks();
  });

  it('renders the date in the active language format', async () => {
    const wrapper = mountPicker();
    expect(inputValue(wrapper)).toBe('10/31/2026');

    locale.value = 'zh-CN';
    await nextTick();
    expect(inputValue(wrapper)).toBe('2026/10/31');
  });

  it('follows the configured date format setting', async () => {
    settings.value = { date_format: 'yyyy-mm-dd' };
    const wrapper = mountPicker();
    expect(inputValue(wrapper)).toBe('2026-10-31');

    settings.value = { date_format: 'dd.mm.yyyy' };
    await nextTick();
    expect(inputValue(wrapper)).toBe('31.10.2026');
  });

  it('parses typed dates using the same field order', async () => {
    const wrapper = mountPicker();
    const input = wrapper.get('input');
    await input.setValue('11/16/2026');
    await input.trigger('blur');

    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['2026-11-16']);
    expect(wrapper.emitted('change')?.[0]).toEqual(['2026-11-16']);
    expect(inputValue(wrapper)).toBe('11/16/2026');
  });

  it('parses typed dates in the Chinese order after switching language', async () => {
    locale.value = 'zh-CN';
    const wrapper = mountPicker();
    const input = wrapper.get('input');
    await input.setValue('2026/11/16');
    await input.trigger('blur');

    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['2026-11-16']);
  });

  it('keeps the previous value when typed text is invalid', async () => {
    const wrapper = mountPicker();
    const input = wrapper.get('input');
    await input.setValue('99/99/9999');
    await input.trigger('blur');

    expect(wrapper.emitted('update:modelValue')).toBeUndefined();
    expect(inputValue(wrapper)).toBe('10/31/2026');
  });

  it('does not accept dates outside the min/max range', async () => {
    const wrapper = mountPicker({ min: '2026-10-01', max: '2026-10-31' });
    const input = wrapper.get('input');
    await input.setValue('11/05/2026');
    await input.trigger('blur');

    expect(wrapper.emitted('update:modelValue')).toBeUndefined();
    expect(inputValue(wrapper)).toBe('10/31/2026');
  });

  it('opens a localized calendar and emits the picked ISO date', async () => {
    const wrapper = mountPicker({ modelValue: '' });
    await wrapper.get('.date-picker-trigger').trigger('click');
    await nextTick();

    const panel = document.body.querySelector('.date-picker-panel');
    expect(panel).not.toBeNull();
    // Month title and weekday labels come from Intl for the active locale
    expect(panel?.querySelector('.date-picker-title')?.textContent).toBe(
      new Intl.DateTimeFormat('en-US', { month: 'long', year: 'numeric' }).format(new Date())
    );
    const weekdays = Array.from(panel?.querySelectorAll('.date-picker-weekdays span') || []);
    expect(weekdays).toHaveLength(7);
    expect(weekdays[0].textContent).toBe(firstWeekdayLabel('en-US'));
    expect(panel?.querySelectorAll('.date-picker-day')).toHaveLength(42);

    const target = Array.from(
      panel?.querySelectorAll<HTMLButtonElement>('.date-picker-day') || []
    ).find((button) => !button.disabled && button.textContent?.trim() === '15');
    expect(target).toBeDefined();
    target?.click();
    await nextTick();

    const iso = wrapper.emitted('update:modelValue')?.[0]?.[0] as string;
    expect(iso).toMatch(/^\d{4}-\d{2}-15$/);
    expect(document.body.querySelector('.date-picker-panel')).toBeNull();
  });

  it('starts the week on the locale first day of week', async () => {
    locale.value = 'zh-CN';
    const wrapper = mountPicker({ modelValue: '' });
    await wrapper.get('.date-picker-trigger').trigger('click');
    await nextTick();

    const weekdays = Array.from(
      document.body.querySelectorAll('.date-picker-weekdays span') || []
    ).map((element) => element.textContent);
    expect(weekdays).toHaveLength(7);
    expect(weekdays[0]).toBe(firstWeekdayLabel('zh-CN'));
  });

  it('clears the value from the panel footer', async () => {
    const wrapper = mountPicker();
    await wrapper.get('.date-picker-trigger').trigger('click');
    await nextTick();

    const actions = Array.from(
      document.body.querySelectorAll<HTMLButtonElement>('.date-picker-action')
    );
    expect(actions.map((action) => action.textContent?.trim())).toEqual([
      'common.datePicker.today',
      'common.datePicker.clear',
    ]);
    actions[1].click();
    await nextTick();

    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['']);
  });
});
