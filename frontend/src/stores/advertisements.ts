import { defineStore } from 'pinia';
import { computed, ref } from 'vue';

export type AdText = Record<string, string>;
export interface Advertisement {
  id: string;
  title: AdText;
  description: AdText;
  url: string;
  coupon_code?: string;
  blocks?: { type: 'text' | 'callout' | 'list'; text?: AdText; items?: AdText[] }[];
}

export function advertisementText(text: AdText | undefined, locale: string): string {
  if (!text) return '';
  return (
    text[locale] ||
    text[locale.split('-')[0]] ||
    text.en ||
    text.zh ||
    Object.values(text).find(Boolean) ||
    ''
  );
}

export const useAdvertisementsStore = defineStore('advertisements', () => {
  const ads = ref<Advertisement[]>([]);
  const snoozedUntil = ref(0);
  const clock = ref(Date.now());
  const popupDismissed = ref(false);
  const saving = ref(false);
  const visibleAds = computed(() => (snoozedUntil.value * 1000 > clock.value ? [] : ads.value));
  let controller: AbortController | null = null;
  let preferenceController: AbortController | null = null;
  let refreshTimer: ReturnType<typeof setInterval> | null = null;
  let clockTimer: ReturnType<typeof setInterval> | null = null;

  async function load() {
    if (saving.value) return;
    controller?.abort();
    const request = new AbortController();
    controller = request;
    try {
      const response = await fetch('/api/ads', { signal: request.signal });
      if (!response.ok) return;
      const data: { ads: Advertisement[]; snoozed_until: number } = await response.json();
      if (request.signal.aborted) return;
      if (!Array.isArray(data.ads) || !Number.isFinite(data.snoozed_until)) return;
      ads.value = data.ads;
      snoozedUntil.value = data.snoozed_until;
      clock.value = Date.now();
    } catch {
      // Promotions are optional; an outage must not interrupt reading.
    } finally {
      if (controller === request) controller = null;
    }
  }

  async function snooze(): Promise<boolean> {
    if (saving.value) return false;
    saving.value = true;
    // A response started before snoozing must not restore the old preference.
    controller?.abort();
    const request = new AbortController();
    preferenceController = request;
    try {
      const response = await fetch('/api/ads/snooze', { method: 'POST', signal: request.signal });
      if (!response.ok) return false;
      const data: { snoozed_until: number } = await response.json();
      if (request.signal.aborted || !Number.isFinite(data.snoozed_until)) return false;
      snoozedUntil.value = data.snoozed_until;
      clock.value = Date.now();
      return true;
    } catch {
      return false;
    } finally {
      if (preferenceController === request) preferenceController = null;
      saving.value = false;
    }
  }

  function start() {
    if (refreshTimer) return;
    void load();
    refreshTimer = setInterval(
      () => {
        if (!saving.value) void load();
      },
      60 * 60 * 1000
    );
    clockTimer = setInterval(() => {
      clock.value = Date.now();
    }, 60 * 1000);
  }

  function stop() {
    controller?.abort();
    preferenceController?.abort();
    if (refreshTimer) clearInterval(refreshTimer);
    if (clockTimer) clearInterval(clockTimer);
    refreshTimer = clockTimer = null;
  }

  return { visibleAds, popupDismissed, saving, load, snooze, start, stop };
});
