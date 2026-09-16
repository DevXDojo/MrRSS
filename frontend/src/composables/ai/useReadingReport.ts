import { ref, onBeforeUnmount } from 'vue';
import { readAIError } from '@/utils/aiError';
import type { ReadingReportResult, ReportSource } from '@/types/readingReport';

// Preview and generation have separate lifetimes. Cancelling or changing a
// selection invalidates late responses without discarding a previous report.
export function useReadingReport() {
  const sources = ref<ReportSource[]>([]);
  const result = ref<ReadingReportResult | null>(null);
  const previewing = ref(false);
  const generating = ref(false);
  const error = ref('');
  let previewController: AbortController | null = null;
  let generationController: AbortController | null = null;

  function cancel() {
    generationController?.abort();
    generationController = null;
    generating.value = false;
  }

  async function preview(articleIDs: number[]) {
    previewController?.abort();
    const controller = new AbortController();
    previewController = controller;
    sources.value = [];
    error.value = '';
    previewing.value = true;
    try {
      const response = await fetch('/api/ai/reading-report/preview', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ article_ids: articleIDs }),
        signal: controller.signal,
      });
      if (!response.ok) throw new Error((await readAIError(response)).message);
      const data = (await response.json()) as { sources: ReportSource[] };
      if (previewController === controller) sources.value = data.sources;
    } catch (e) {
      if (previewController === controller && !controller.signal.aborted) {
        error.value = e instanceof Error ? e.message : '';
      }
    } finally {
      if (previewController === controller) previewing.value = false;
    }
  }

  async function generate(articleIDs: number[], profileID: number, focus: string) {
    if (generating.value || previewing.value || !sources.value.some((s) => s.kind !== 'missing'))
      return;
    const controller = new AbortController();
    generationController = controller;
    generating.value = true;
    error.value = '';
    try {
      const response = await fetch('/api/ai/reading-report', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ article_ids: articleIDs, profile_id: profileID, focus }),
        signal: controller.signal,
      });
      if (!response.ok) throw new Error((await readAIError(response)).message);
      const data = (await response.json()) as ReadingReportResult;
      if (generationController === controller) result.value = data;
    } catch (e) {
      if (generationController === controller && !controller.signal.aborted) {
        error.value = e instanceof Error ? e.message : '';
      }
    } finally {
      if (generationController === controller) {
        generating.value = false;
        generationController = null;
      }
    }
  }

  onBeforeUnmount(() => {
    previewController?.abort();
    previewController = null;
    cancel();
  });
  return { sources, result, previewing, generating, error, preview, generate, cancel };
}
