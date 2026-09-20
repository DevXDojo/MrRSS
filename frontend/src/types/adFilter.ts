export interface AdFilterOptions {
  enabled: boolean;
  allowlist: string;
  rules: string;
  ai_enabled: boolean;
  ai_mode: 'review' | 'automatic';
  include_self_promotion: boolean;
}
export interface AdFilterConfig extends AdFilterOptions {
  revision: number;
}
export interface AdFilterReport {
  enabled: boolean;
  removed: number;
  guarded: number;
  reasons: Record<string, number>;
}
export interface ArticleFilterInfo {
  originalContent: string;
  sourceContent?: string;
  report: AdFilterReport;
}
export interface AdDecision {
  id: string;
  category: 'advertisement' | 'affiliate_promotion' | 'self_promotion';
  confidence: number;
  evidence: string;
}
export interface AdAnalysis {
  content: string;
  decisions: AdDecision[];
  scanned: number;
  total: number;
  truncated: boolean;
  cached: boolean;
  guarded: number;
}
export function defaultAdFilterConfig(): AdFilterConfig {
  return {
    enabled: true,
    allowlist: '',
    rules: '',
    ai_enabled: false,
    ai_mode: 'review',
    include_self_promotion: false,
    revision: 0,
  };
}

export function isAdFilterApplicable(options: AdFilterOptions, url: string): boolean {
  if (!options.enabled) return false;
  try {
    const parsed = new URL(url);
    if (!['http:', 'https:'].includes(parsed.protocol)) return false;
    const host = parsed.hostname.toLowerCase().replace(/\.$/, '');
    return !options.allowlist
      .split('\n')
      .map((line) => line.trim().toLowerCase())
      .filter(Boolean)
      .some((domain) => host === domain || host.endsWith('.' + domain));
  } catch {
    return false;
  }
}
