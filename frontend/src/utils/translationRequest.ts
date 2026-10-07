// Share a small request pool and provider cooldown between titles and content.
// Queued work checks the cooldown again before contacting a throttled provider.
let active = 0;
let retryAt = 0;
let lastErrorToast = -Infinity;
const waiting: Array<() => void> = [];

// Explicit configuration changes may select a different, healthy provider.
export function resetTranslationCooldown(): void {
  retryAt = 0;
  lastErrorToast = -Infinity;
}

export async function requestTranslation(
  url: string,
  body: object,
  isCurrent: () => boolean = () => true
): Promise<Response> {
  if (active >= 3) await new Promise<void>((resolve) => waiting.push(resolve));
  else active++;
  try {
    if (!isCurrent()) return new Response(null, { status: 204 });
    if (Date.now() < retryAt) {
      return new Response(null, {
        status: 429,
        headers: { 'Retry-After': String(Math.ceil((retryAt - Date.now()) / 1000)) },
      });
    }
    const res = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (res.status === 429) {
      const header = res.headers.get('Retry-After') || '';
      const seconds = Number(header);
      const delay = seconds > 0 ? seconds * 1000 : Date.parse(header) - Date.now();
      retryAt = Math.max(retryAt, Date.now() + (delay > 0 ? delay : 60000));
    }
    return res;
  } finally {
    const next = waiting.shift();
    if (next) next();
    else active--;
  }
}

export function notifyTranslationError(message: string): void {
  if (Date.now() - lastErrorToast < 15000) return;
  lastErrorToast = Date.now();
  window.showToast(message, 'error');
}
