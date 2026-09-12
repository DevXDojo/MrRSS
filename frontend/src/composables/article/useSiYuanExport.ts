import { onBeforeUnmount, ref } from 'vue';
import { useI18n } from 'vue-i18n';

export function useSiYuanExport() {
  const { t } = useI18n();
  const isExporting = ref(false);
  let controller: AbortController | null = null;

  onBeforeUnmount(() => controller?.abort());

  async function exportToSiYuan(articleID: number): Promise<void> {
    if (isExporting.value) return;
    const request = new AbortController();
    controller = request;
    isExporting.value = true;
    window.showToast(t('setting.plugins.siyuan.exporting'), 'info');
    try {
      const response = await fetch('/api/articles/export/siyuan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ article_id: articleID }),
        signal: request.signal,
      });
      if (!response.ok) {
        const reason =
          response.status === 400
            ? 'configurationError'
            : response.status === 422
              ? 'contentError'
              : 'exportFailed';
        if (!request.signal.aborted)
          window.showToast(t(`setting.plugins.siyuan.${reason}`), 'error');
        return;
      }
      if (!request.signal.aborted)
        window.showToast(t('setting.plugins.siyuan.exported'), 'success');
    } catch {
      if (!request.signal.aborted)
        window.showToast(t('setting.plugins.siyuan.exportFailed'), 'error');
    } finally {
      if (controller === request) {
        controller = null;
        isExporting.value = false;
      }
    }
  }

  return { isExporting, exportToSiYuan };
}
