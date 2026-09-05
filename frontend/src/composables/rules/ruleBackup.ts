import type { Condition } from './useRuleOptions';

export interface BackupRule {
  id: number;
  name: string;
  enabled: boolean;
  conditions: Condition[];
  actions: string[];
  position?: number;
}

export function exportRules(rules: BackupRule[]): string {
  return JSON.stringify({ format: 'mrrss-rules', version: 1, rules: [...rules].sort((a, b) => (a.position ?? 0) - (b.position ?? 0)) }, null, 2);
}

export function importRules(text: string, fields: string[], actions: string[], existing: BackupRule[]): BackupRule[] {
  if (new TextEncoder().encode(text).length > 5 * 1024 * 1024) throw new Error('backupTooLarge');
  let data: unknown;
  try { data = JSON.parse(text.replace(/^\uFEFF/, '')); } catch { throw new Error('invalidBackup'); }
  const object = (value: unknown): value is Record<string, unknown> => !!value && typeof value === 'object' && !Array.isArray(value);
  // Also accept the raw MrRSS rules array from older settings backups.
  if (!Array.isArray(data)) {
    if (!object(data) || data.format !== 'mrrss-rules' || data.version !== 1) throw new Error('invalidBackup');
    data = data.rules;
  }
  if (!Array.isArray(data) || data.length === 0 || data.length > 1000) throw new Error('invalidBackup');
  let id = Math.max(Date.now(), ...existing.map((rule) => rule.id).filter(Number.isSafeInteger));
  return data.map((rule: unknown, index) => {
    if (!object(rule) || typeof rule.name !== 'string' || !rule.name.trim() || typeof rule.enabled !== 'boolean' ||
        !Array.isArray(rule.conditions) || !rule.conditions.length || rule.conditions.length > 1000 ||
        !Array.isArray(rule.actions) || !rule.actions.length || !rule.actions.every((action) => typeof action === 'string' && actions.includes(action))) throw new Error('invalidBackup');
    const conditions = rule.conditions.map((condition: unknown, conditionIndex): Condition => {
      if (!object(condition) || typeof condition.field !== 'string' || !fields.includes(condition.field) ||
          typeof condition.negate !== 'boolean' || typeof condition.value !== 'string' ||
          !Array.isArray(condition.values) || !condition.values.every((value) => typeof value === 'string') ||
          (condition.logic != null && condition.logic !== 'and' && condition.logic !== 'or') ||
          (condition.operator != null && !['', 'contains', 'exact', 'regex'].includes(String(condition.operator)))) throw new Error('invalidBackup');
      return { id: conditionIndex + 1, field: condition.field, negate: condition.negate,
        value: condition.value, values: [...condition.values] as string[],
        logic: condition.logic as Condition['logic'], operator: condition.operator as Condition['operator'] };
    });
    if (!Number.isSafeInteger(++id)) throw new Error('invalidBackup');
    return { id, name: rule.name, enabled: rule.enabled, conditions, actions: [...rule.actions] as string[], position: existing.length + index };
  });
}
