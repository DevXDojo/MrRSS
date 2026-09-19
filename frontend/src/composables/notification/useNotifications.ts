import { ref, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import type {
  NotificationConfig,
  NotificationDelivery,
  NotificationChannel,
  NotificationRule,
  NotificationPreview,
} from '@/types/notification';

export function useNotifications() {
  const { t, te } = useI18n();
  const config = ref<NotificationConfig | null>(null);
  const history = ref<NotificationDelivery[]>([]);
  const loading = ref(true);
  const saving = ref(false);
  const error = ref('');
  const historyError = ref(false);
  const controller = new AbortController();
  let poll: ReturnType<typeof setInterval> | undefined;
  let polling = false;

  function errorText(code: string) {
    const key = `setting.notifications.errors.${code}`;
    return te(key) ? t(key) : t('setting.notifications.errors.internal');
  }
  async function request<T>(path: string, method = 'GET', body?: unknown): Promise<T> {
    const res = await fetch(`/api/notifications/${path}`, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: controller.signal,
    });
    const data = await res.json();
    if (!res.ok)
      throw new Error(typeof data.error?.message === 'string' ? data.error.message : 'internal');
    return data as T;
  }
  async function load() {
    loading.value = true;
    error.value = '';
    try {
      config.value = await request<NotificationConfig>('config');
    } catch (e) {
      if (!controller.signal.aborted)
        error.value = errorText(e instanceof Error ? e.message : 'internal');
    } finally {
      loading.value = false;
    }
  }
  async function refreshHistory() {
    if (polling) return;
    polling = true;
    try {
      history.value = await request<NotificationDelivery[]>('history');
      historyError.value = false;
    } catch {
      if (!controller.signal.aborted) historyError.value = true;
    } finally {
      polling = false;
    }
  }
  async function save(next: NotificationConfig) {
    if (saving.value) return false;
    saving.value = true;
    error.value = '';
    try {
      config.value = await request<NotificationConfig>('config', 'PUT', next);
      window.showToast(t('setting.notifications.saved'), 'success');
      void refreshHistory();
      return true;
    } catch (e) {
      if (!controller.signal.aborted) {
        error.value = errorText(e instanceof Error ? e.message : 'internal');
        window.showToast(error.value, 'error');
      }
      return false;
    } finally {
      saving.value = false;
    }
  }
  async function test(channel: NotificationChannel) {
    await request('test', 'POST', channel);
  }
  async function preview(rule: NotificationRule) {
    return request<NotificationPreview>('preview', 'POST', rule);
  }
  onMounted(() => {
    void load();
    void refreshHistory();
    poll = setInterval(() => void refreshHistory(), 15000);
  });
  onBeforeUnmount(() => {
    controller.abort();
    if (poll) clearInterval(poll);
  });
  return {
    config,
    history,
    loading,
    saving,
    error,
    historyError,
    load,
    save,
    test,
    preview,
    refreshHistory,
    errorText,
  };
}
