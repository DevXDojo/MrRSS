import { readonly, shallowRef } from 'vue';

interface TranslationJob {
  url: string;
  body: object;
  isCurrent: () => boolean;
  priority: 'interactive' | 'background';
  resolve: (response: Response) => void;
  reject: (error: unknown) => void;
  controller: AbortController;
  revision: number;
}

const queue: TranslationJob[] = [];
const active = new Set<TranslationJob>();
const state = shallowRef({ pending: 0, retryAt: 0, recovering: false });
export const translationQueue = readonly(state);
let retryAt = 0;
let nextStartAt = 0;
let recoveryDelay = 0;
let needsProbe = false;
let consecutiveLimits = 0;
let revision = 0;
let limitRevision = 0;
let timer: ReturnType<typeof setTimeout> | undefined;
let lastErrorToast = -Infinity;

function updateState(): void {
  state.value = {
    pending: queue.length + active.size,
    retryAt,
    recovering: recoveryDelay > 0,
  };
}

function retryDelay(header: string | null): number {
  const seconds = Number(header);
  const delay = seconds > 0 ? seconds * 1000 : Date.parse(header || '') - Date.now();
  // Honor provider delays; increase the fallback when repeated probes are limited.
  return delay > 0
    ? Math.min(delay, 86400000)
    : Math.min(300000, 60000 * 2 ** Math.min(consecutiveLimits - 1, 3));
}

async function run(job: TranslationJob): Promise<void> {
  const limitRevisionAtStart = limitRevision;
  try {
    const response = await fetch(job.url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(job.body),
      signal: job.controller.signal,
    });
    if (!job.isCurrent() || job.revision !== revision) {
      void response.body?.cancel();
      job.resolve(new Response(null, { status: 204 }));
    } else if (response.status === 429) {
      if (Date.now() >= retryAt) {
        consecutiveLimits++;
        recoveryDelay = Math.min(10000, Math.max(1000, recoveryDelay * 2));
      }
      limitRevision++;
      retryAt = Math.max(retryAt, Date.now() + retryDelay(response.headers.get('Retry-After')));
      needsProbe = true;
      void response.body?.cancel();
      // Keep the same promise pending: callers continue the same title/paragraph.
      queue.push(job);
    } else {
      if (response.ok && limitRevisionAtStart === limitRevision) {
        needsProbe = false;
        consecutiveLimits = 0;
      }
      job.resolve(response);
    }
  } catch (error) {
    if (!job.isCurrent() || job.revision !== revision)
      job.resolve(new Response(null, { status: 204 }));
    else job.reject(error);
  } finally {
    active.delete(job);
    pump();
  }
}

function pump(): void {
  if (timer) clearTimeout(timer);
  timer = undefined;
  for (let i = queue.length - 1; i >= 0; i--) {
    const job = queue[i];
    if (!job.isCurrent() || job.revision !== revision) {
      queue.splice(i, 1);
      job.resolve(new Response(null, { status: 204 }));
    }
  }
  for (const job of active) {
    if (!job.isCurrent() || job.revision !== revision) job.controller.abort();
  }

  const limit = needsProbe ? 1 : 3;
  while (queue.length && active.size < limit && Date.now() >= Math.max(retryAt, nextStartAt)) {
    // Reading and explicit translation take precedence over list prefetches.
    const interactive = queue.findIndex((job) => job.priority === 'interactive');
    const job = queue.splice(interactive < 0 ? 0 : interactive, 1)[0];
    active.add(job);
    nextStartAt = Date.now() + recoveryDelay;
    void run(job);
  }
  updateState();
  // One timer checks cancellation and wakes the queue; no per-request retry loops.
  if (queue.length || active.size) {
    const delay = Math.max(1, Math.max(retryAt, nextStartAt) - Date.now());
    timer = setTimeout(pump, Math.min(1000, delay > 1 ? delay : 1000));
  }
}

// Explicit configuration changes may select a different, healthy provider.
export function resetTranslationCooldown(): void {
  revision++;
  retryAt = 0;
  nextStartAt = 0;
  recoveryDelay = 0;
  needsProbe = false;
  consecutiveLimits = 0;
  lastErrorToast = -Infinity;
  pump();
}

export function requestTranslation(
  url: string,
  body: object,
  isCurrent: () => boolean = () => true,
  priority: TranslationJob['priority'] = 'interactive'
): Promise<Response> {
  return new Promise<Response>((resolve, reject) => {
    queue.push({
      url,
      body,
      isCurrent,
      priority,
      resolve,
      reject,
      controller: new AbortController(),
      revision,
    });
    pump();
  });
}

export function notifyTranslationError(message: string): void {
  if (Date.now() - lastErrorToast < 15000) return;
  lastErrorToast = Date.now();
  window.showToast(message, 'error');
}
