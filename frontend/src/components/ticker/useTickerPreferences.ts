import { onMounted, onUnmounted, ref, watch, type Ref, type InjectionKey } from 'vue';
import { Call, Events } from '@wailsio/runtime';
import type { SettingsData } from '@/types/settings';
import { parseSettingsData } from '@/composables/core/useSettings.generated';
import { tickerPayload } from './tickerPreferences';

export const tickerPreferencesKey: InjectionKey<ReturnType<typeof useTickerPreferences>> =
  Symbol('tickerPreferences');

// Each window submits only its changed keys, so a settings panel and a quick
// toolbar action cannot overwrite each other's newer preferences.
export function useTickerPreferences(settings: Ref<SettingsData>) {
  const status = ref<'saved' | 'saving' | 'error'>('saved');
  const pending = new Map<string, string>();
  let observed = tickerPayload(settings.value);
  let inFlight = new Set<string>();
  let saveTask: Promise<boolean> | undefined;
  let applying = false;
  let disposed = false;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let readController: AbortController | undefined;
  const lifetime = new AbortController();
  const disposeEvents: Array<() => void> = [];

  function saveNow(): Promise<boolean> {
    clearTimeout(timer);
    if (disposed) return Promise.resolve(false);
    if (!saveTask)
      saveTask = flushPending().finally(() => {
        saveTask = undefined;
      });
    return saveTask;
  }
  async function flushPending(): Promise<boolean> {
    status.value = 'saving';
    let patch: Record<string, string> = {};
    try {
      do {
        patch = Object.fromEntries(pending);
        pending.clear();
        inFlight = new Set(Object.keys(patch));
        if (inFlight.size) {
          const response = await fetch('/api/settings', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(patch),
            signal: lifetime.signal,
          });
          if (!response.ok) throw new Error('Could not save ticker preferences');
        }
        patch = {};
        await Call.ByName('MrRSS/internal/desktopapi.DesktopCapabilities.ApplyTickerSettings');
        inFlight.clear();
      } while (pending.size && !disposed);
      if (!disposed) status.value = 'saved';
      return true;
    } catch {
      for (const [key, value] of Object.entries(patch)) {
        if (!pending.has(key)) pending.set(key, value);
      }
      if (!disposed) status.value = 'error';
      return false;
    } finally {
      inFlight.clear();
    }
  }

  watch(
    () => tickerPayload(settings.value),
    (values) => {
      if (!applying) {
        for (const [key, value] of Object.entries(values)) {
          if (value !== observed[key]) pending.set(key, value);
        }
        if (pending.size) {
          status.value = 'saving';
          clearTimeout(timer);
          timer = setTimeout(() => {
            void saveNow();
          }, 350);
        }
      }
      observed = values;
    },
    { flush: 'sync' }
  );

  async function reload(): Promise<void> {
    readController?.abort();
    const current = new AbortController();
    readController = current;
    try {
      const response = await fetch('/api/settings', { signal: current.signal });
      if (!response.ok) return;
      const values = parseSettingsData(await response.json());
      if (disposed || current.signal.aborted) return;
      applying = true;
      Object.assign(
        settings.value,
        Object.fromEntries(
          Object.entries(values).filter(
            ([key]) => key.startsWith('ticker_') && !pending.has(key) && !inFlight.has(key)
          )
        )
      );
      observed = tickerPayload(settings.value);
    } catch {
      /* Keep the local preferences on a temporary refresh failure. */
    } finally {
      applying = false;
    }
  }

  onMounted(() => {
    disposeEvents.push(
      Events.On('ticker:settings-changed', () => {
        void reload();
      })
    );
  });
  onUnmounted(() => {
    disposed = true;
    clearTimeout(timer);
    readController?.abort();
    lifetime.abort();
    disposeEvents.forEach((dispose) => dispose());
  });
  return { status, saveNow, reload };
}
