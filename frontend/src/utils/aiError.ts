import i18n from '@/i18n';

export interface ParsedAIError {
  code: string;
  message: string;
  payload?: unknown;
}

const knownCodes = new Set([
  'configuration_invalid',
  'usage_limit_reached',
  'rate_limited',
  'authentication_failed',
  'payment_required',
  'model_or_endpoint_not_found',
  'request_too_large',
  'timeout',
  'network_error',
  'provider_unavailable',
  'invalid_response',
  'provider_rejected_request',
  'request_failed',
]);

const legacyCodeAliases: Record<string, string> = {
  AI_CONFIG_FAILED: 'configuration_invalid',
  AI_QUOTA_EXCEEDED: 'rate_limited',
  AI_INVALID_REQUEST: 'provider_rejected_request',
  AI_REQUEST_FAILED: 'request_failed',
};

function asRecord(value: unknown): Record<string, unknown> | null {
  return value !== null && typeof value === 'object' ? (value as Record<string, unknown>) : null;
}

function payloadCode(payload: unknown): string {
  const record = asRecord(payload);
  if (!record) return '';
  const error = asRecord(record.error);
  const candidate = record.error_code ?? error?.code ?? record.code;
  if (typeof candidate !== 'string') return '';
  return legacyCodeAliases[candidate] || candidate.toLowerCase();
}

function payloadText(payload: unknown): string {
  if (typeof payload === 'string') return payload;
  if (payload instanceof Error) return payload.message;
  const record = asRecord(payload);
  if (!record) return '';
  const error = asRecord(record.error);
  const values = [record.error, error?.message, record.message];
  return values.find((value): value is string => typeof value === 'string') || '';
}

function inferCode(payload: unknown, status?: number): string {
  const explicit = payloadCode(payload);
  if (knownCodes.has(explicit)) return explicit;

  switch (status) {
    case 401:
    case 403:
      return 'authentication_failed';
    case 402:
      return 'payment_required';
    case 404:
      return 'model_or_endpoint_not_found';
    case 408:
    case 504:
      return 'timeout';
    case 413:
      return 'request_too_large';
    case 429:
      return 'rate_limited';
    default:
      if (status && status >= 500) return 'provider_unavailable';
  }

  const text = payloadText(payload).toLowerCase();
  if (/\b429\b|rate.?limit|too many requests/.test(text)) return 'rate_limited';
  if (/\b401\b|\b403\b|unauthori[sz]ed|forbidden|invalid api.?key|authentication/.test(text)) {
    return 'authentication_failed';
  }
  if (/\b402\b|insufficient (quota|balance)|payment required|credit/.test(text)) {
    return 'payment_required';
  }
  if (/\b404\b|model.*not found|endpoint.*not found/.test(text)) {
    return 'model_or_endpoint_not_found';
  }
  if (/\b413\b|request.*too large|context length/.test(text)) return 'request_too_large';
  if (/timeout|timed out|deadline exceeded/.test(text)) return 'timeout';
  if (/connection refused|network|no such host|failed to fetch/.test(text)) return 'network_error';
  if (/invalid json|invalid response|empty response|no choices/.test(text))
    return 'invalid_response';
  return 'request_failed';
}

export function getAIErrorMessage(payload: unknown, status?: number): string {
  const code = inferCode(payload, status);
  return String(i18n.global.t(`aiErrors.${code}`));
}

export async function readAIError(response: Response): Promise<ParsedAIError> {
  let raw = '';
  try {
    raw = await response.text();
  } catch {
    // A response body is optional; the HTTP status still provides a stable
    // classification without exposing transport details.
  }
  let payload: unknown;
  try {
    payload = JSON.parse(raw);
  } catch {
    payload = raw;
  }
  const code = inferCode(payload, response.status);
  return { code, message: getAIErrorMessage(payload, response.status), payload };
}
